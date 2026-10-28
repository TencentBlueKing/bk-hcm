# Design

## Context

动机与背景见 `proposal.md - Why`。当前 `hc-service` 已具备以下相关模块：

1. **北向 HTTP 入口与路由器（**`pkg/api/hc-service/iac-plugin/northbound.go`**、**`cmd/hc-service/logics/iac-plugin/router/`**）**：已定义 11 个北向接口路由并实现三方 `http` 插件的路由分发，但 `router.Service` 的 `tf_provider` 分支仍是返回 `errf.Aborted` 的占位实现；同时 `ValidateReq` 尚缺 `config_type` 字段，`StateResult` 与 `ReadDataResult` 尚缺 `diagnostics` 字段。
2. **状态合成器（**`cmd/hc-service/logics/iac-plugin/http-plugin/composer.go`**）**：已实现对照 `*tfjson.SchemaBlock` 的 `ValidateConfigAndRecompute` 与 `ComposeProposedState`，支持 4 类属性规则及顶层、`NestedBlocks`、`AttributeNestedType` 递归合成（列表嵌套块 `recompute` 路径支持方括号下标 `"clouds.names[0].name"` 与无下标通配 `"clouds.names.name"`），可由 `TFPluginAdaptor` 直接无修改复用。
3. **本地插件目录扫描、严格 FIFO 进程池与底层执行器**：
   - `cmd/hc-service/logics/iac-plugin/tf-plugin/discovery`：启动时扫描本地插件目录产出只读 `[]discovery.Entry`（含锁定的 `PluginName` 与 `PluginVersion`）；
   - `cmd/hc-service/logics/iac-plugin/tf-plugin/pool`：`PluginRegistry.GetPool(pluginName)` 返回 `*PluginPool`，`PluginPool.Acquire(kt, config)` 返回已完成 `ConfigureProvider` 的 `*Lease`（实现 `pool.ProviderIF` 接口），`Lease.Release()` 幂等归还进程并发槽位；
   - `cmd/hc-service/logics/tfruntime/executor 与 codec：当插件 RPC 返回的 Diagnostics 含 Severity == ERROR 时，业务方法同时返回结果结构体与包装了 executor.ErrDiagnostic 的 error；解码状态中尚未计算的未知值以 codec.Unknown 哨兵值表示。`

## Goals / Non-Goals

**Goals:**

- 在不改变底层 `tfruntime/executor`、`tf-plugin/pool` 与 `http-plugin/composer.go` 内部实现的前提下，构建 `cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor` 运行态适配包，完成单次北向请求内同一 `Lease` 的多步 gRPC 串行编排、状态挽回与诊断转换。
- 在 `router.Service` 中打通本地 `tf_provider` 插件的本副本 `localVersion` 锁定寻址、`Health` 探活、按 `config_type` 精确门控 `Validate` 以及 6 个运行态接口分发。
- 直接复用 `httpplugin.ValidateConfigAndRecompute` 与 `httpplugin.ComposeProposedState` 完成配置静态校验与 `ProposedNewState` 合成（数组元素 `recompute` 路径统一采用 `"clouds.names[0].name"` 格式）。
- 满足性能与安全约束：Mock `ProviderIF` + `SchemaCache` 命中下单次适配层编排 P99 < 5ms（1,000 次 Benchmark）、50 协程 `-race` 零竞争、`Lease` 100% 归还、日志与错误消息中实例明文凭证零泄露。

**Non-Goals:**

- 不在 `TFPluginAdaptor` 中维护异步任务表或实现 `GetResult`、`Stop`、`List`（本地 TF 插件不支持异步 Apply 与 List）。
- 不传递或解析 `ResourceIdentity`（底层基于 TF 协议 5.3，北向请求结构体保留 `PriorIdentity`/`Identity` 字段兼容报文，运行态固定忽略入参并返回 `Identity: nil`）。

## Architecture & Module Layout

```
Northbound HTTP Handler (cmd/hc-service/service/iac-plugin/northbound.go)
   │  Decode JSON + req.Validate() (incl. ValidateReq.ConfigType in {"resource","data"})
   ▼
Plugin Router (cmd/hc-service/logics/iac-plugin/router/router.go)
   ├── Health (tf_provider)    ──► check localVersion in discovery.Entry + iac_plugin(name, localVersion).State == enabled
   ├── GetSchema               ──► SchemaManager.GetSchema (if version=="": localVersion for tf_provider, Latest for http)
   ├── GetResult / Stop / List ──► intercepted by capabilities gate (async_apply=false, list=false → errf.Aborted)
   └── Execute / Plan / Validate / Import / Refresh / ReadData:
         1. Resolve target plugin record:
              - tf_provider: match (plugin, localVersion) in store.List, verify state == "enabled"
              - http:        iacplugin.Latest(records), verify state == "enabled"
         2. SchemaManager.GetSchema(plugin, targetVersion) + checkTypeInSchema (Validate uses ConfigType) + checkCapability
         3. instanceMgr.GetPluginInstance(instanceID) → verify inst.PluginName == plugin
         4. Dispatch tf_provider:
              ├── poolRegistry.GetPool(plugin)
              ├── adaptor.NewTFPluginAdaptor(kt, pluginPool, schema, instanceID, properties) → pool.Acquire(kt, properties)
              ├── defer tfAdaptor.Close() → lease.Release()
              └── tfAdaptor.<Execute|Plan|Validate|Import|Refresh|ReadData>(kt, req)
```

### 包路径与文件划分

| 模块路径 | 文件 | 核心职责 |
|---|---|---|
| `pkg/api/hc-service/iac-plugin/` | `northbound.go`, `northbound_test.go` | 定义 `ConfigTypeResource = "resource"`、`ConfigTypeDataSource = "data"`，`ValidateReq` 新增必填 `ConfigType`，`StateResult` 与 `ReadDataResult` 新增 `Diagnostics` |
| `cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor/` | `tf_adaptor.go`, `convert.go`, `tf_adaptor_test.go`, `benchmark_test.go` | `TFPluginAdaptor` 构造/释放、`prepareChangeContext`（复用 `httpplugin.ComposeProposedState`）、6 个运行态适配方法、Base64/Diagnostic/AttributePath/`codec.Unknown` 转换及单测与 Benchmark |
| `cmd/hc-service/logics/iac-plugin/router/` | `router.go`, `router_test.go` | 本地插件 `localVersion` 寻址、`Validate` 按 `ConfigType` 校验 Schema、本地插件 `Health` 探活、构造 `TFPluginAdaptor` 分发 6 个运行态接口 |
| `cmd/hc-service/service/` | `iac_plugin.go` | 先初始化 `s.poolRegistry`，再将 `s.poolRegistry` 与 `s.plugins` 注入 `router.New` |

## Detailed Interface & Algorithm Design

### 1. 北向契约扩展（`pkg/api/hc-service/iac-plugin/northbound.go`）

- 新增常量：
  ```go
  const (
      ConfigTypeResource   = "resource"
      ConfigTypeDataSource = "data"
  )
  ```
- `ValidateReq` 增加 `ConfigType string`（`json:"config_type"`），`Validate()` 校验 `req.ConfigType == ConfigTypeResource || req.ConfigType == ConfigTypeDataSource`，不满足时返回错误（Handler 映射为 `errf.InvalidParameter`）。
- `StateResult` 与 `ReadDataResult` 增加 `Diagnostics []coreplugin.Diagnostic`（`json:"diagnostics,omitempty"`）。

### 2. `TFPluginAdaptor` 与协议转换设计（`cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor/`）

#### 2.1 `TFPluginAdaptor` 结构体与构造/释放

在 `cmd/hc-service/logics/iac-plugin/tf-plugin/pool/lease.go` 的 `ProviderIF` 接口中补齐 `Release()`（`*pool.Lease` 已实现 `Release()`），`TFPluginAdaptor` 直接以 `lease pool.ProviderIF` 作为结构体成员，不定义额外的 `LeaseIF` 或测试专用构造函数：

```go
type TFPluginAdaptor struct {
	pluginName string
	instanceID string
	schema     *coreplugin.PluginSchema
	lease      pool.ProviderIF
	closeOnce  sync.Once
}
```

- `NewTFPluginAdaptor(kt *kit.Kit, p *pool.PluginPool, schema *coreplugin.PluginSchema, instanceID string, properties map[string]any) (*TFPluginAdaptor, error)`：
  - 校验 `p`、`schema` 非空；
  - 调用 `p.Acquire(kt, properties)` 获取 `*pool.Lease` 并直接赋给 `lease` 字段；
  - 若 `Acquire` 返回错误：当 `errors.Is(err, executor.ErrInvalidConfig)` 时返回 `errf.Newf(errf.InvalidParameter, "configure provider for plugin %s instance %s failed", p.Name(), instanceID)`；其余错误（如 `pool.ErrPoolClosed`、`pool.ErrTooManyWaiters`、`pool.ErrAcquireTimeout` 等）返回 `errf.Newf(errf.Aborted, "acquire plugin %s process lease for instance %s failed: %v", p.Name(), instanceID, err)`（日志与错误中绝不打印 `properties`）。
- `Close()`：通过 `closeOnce` 幂等调用 `a.lease.Release()`。

#### 2.2 错误分类辅助函数（`classifyRPCError`）

底层 `PluginExecutor` 在 RPC 正常返回但 `Diagnostics` 含 `Severity == ERROR` 时，会返回 `result != nil` 且 `errors.Is(err, executor.ErrDiagnostic)`：

- 若 `err != nil && !errors.Is(err, executor.ErrDiagnostic)`：属于进程崩溃、gRPC 断开或编解码错误，记录包含 `rid`、`pluginName`、`instanceID`、`typeName` 的错误日志，向上返回 `errf.Newf(errf.Aborted, ...)`（外层 HTTP `code != 0, data: null`）；
- 若 `err == nil || errors.Is(err, executor.ErrDiagnostic)`：属于正常业务返回或插件诊断错误，外层 Go 函数返回 `err == nil`（外层 HTTP `code = 0`），通过 `hasErrorDiagnostic(rawDiags)`（或 `errors.Is(err, executor.ErrDiagnostic)`）判断是否阻断后续步骤并填充 `data.diagnostics`。

#### 2.3 `prepareChangeContext` 流程（`Execute` 与 `Plan` 共用）

1. 校验 `req.PriorState` 与 `req.Config` 不同时为 `nil`，校验 `a.schema.ResourceSchemas[req.TypeName]` 及 `Block` 非空，解码 `req.PriorPrivate` 为 `priorPrivateBytes`；调用 `httpplugin.ValidateConfigAndRecompute(resSchema.Block, req.Config, req.Recompute)` 静态校验配置与 `recompute`。
2. **当** `req.Config != nil`**（创建或更新）**：
   - 调用 `a.lease.ValidateResourceTypeConfig(kt.Ctx, req.TypeName, req.Config)`；若含 Error 诊断，置 `validateFailed = true` 立即返回（不调用 `ReadResource` 与 `PlanResourceChange`）；
   - 若 `req.PriorState != nil`（更新），调用 `a.lease.ReadResource(kt.Ctx, req.TypeName, req.PriorState, priorPrivateBytes)`：
     - 若含 Error 诊断，置 `validateFailed = true` 立即返回；
     - 若 `readResp.NewState == nil`（云端已删），置 `currentState = nil, currentPrivate = nil, drifted = true`（转重建）；
     - 若 `readResp.NewState != nil`，置 `currentState = readResp.NewState, currentPrivate = readResp.Private`，当 `!reflect.DeepEqual(req.PriorState, readResp.NewState)` 时标记 `drifted = true`；
   - 调用 `httpplugin.ComposeProposedState(resSchema.Block, currentState, req.Config, req.Recompute)` 生成 `proposedNewState`。
3. **当** `req.Config == nil`**（删除）**：
   - 调用 `a.lease.ReadResource(kt.Ctx, req.TypeName, req.PriorState, priorPrivateBytes)`：
     - 若含 Error 诊断，置 `validateFailed = true` 立即返回；
     - 若 `readResp.NewState == nil`，置 `shortCircuitDelete = true` 直接返回；
     - 若 `readResp.NewState != nil`，置 `currentState = readResp.NewState, currentPrivate = readResp.Private, proposedNewState = nil`，并对比记录 `drifted`。

#### 2.4 协议转换细节（`convert.go`）

1. `**Private` Base64 编解码**：
   - `decodePrivate(b64 *string) ([]byte, error)`：`nil` 或空串返回 `nil, nil`，否则 `base64.StdEncoding.DecodeString`；
   - `encodePrivate(raw []byte) *string`：`len(raw) == 0` 返回 `nil`，否则返回 `cvt.ValToPtr(base64.StdEncoding.EncodeToString(raw))`。
2. `**Diagnostic` 转换（`convertDiagnostics`）**：
   - 遍历 `[]*tfplugin5.Diagnostic`：`Diagnostic_ERROR` 映射为 `tfjson.DiagnosticSeverityError`（`"error"`），`Diagnostic_WARNING` 映射为 `tfjson.DiagnosticSeverityWarning`（`"warning"`）；
   - `Summary` 与 `Detail` 原样填入内嵌 `tfjson.Diagnostic`；`Message` 组合规则：若 `Summary != ""` 且 `Detail != ""` 则为 `Summary + ": " + Detail`，否则取非空者；
   - `Path` 转换：遍历 `diag.GetAttribute().GetSteps()`，`AttributeName` 转为 `coreplugin.NewAttributeStep(name)`，`ElementKeyString` 转为 `coreplugin.NewKeyStep(key)`，`ElementKeyInt` 转为 `coreplugin.NewIndexStep(int(idx))`（无步骤时为空切片 `[]coreplugin.PathStep{}`）。
3. `**RequiresReplace` 转换（`convertReplacePaths`）**：
   - 遍历 `[]*tfplugin5.AttributePath`，将每个 `AttributePath` 的 `Steps` 按顺序拼接：`AttributeName` 以 `.` 连接，`ElementKeyString` 以 `.` 连接，`ElementKeyInt` 格式化为 `[i]`，生成 `[]string`（非 `nil`，无重建路径时返回 `[]string{}`）。
4. `**codec.Unknown` 递归清理与检测（`sanitizeUnknown`）**：
   - 递归遍历 `map[string]any` 与 `[]any`：若某元素等于 `codec.Unknown`，将其替换为 `nil` 并返回 `hadUnknown = true`：
     - 在 `Plan` 与 `Execute` 中，将 `PlannedState` / `NewState` 中的 `codec.Unknown` 清理为 `nil`；
     - 在 `ReadData` 中，若 `hadUnknown == true`，除清理为 `nil` 外，还主动追加一条 `message="provider produced a state value that is not wholly known"` 的 Error `Diagnostic`。

### 3. `Router` 本地插件接入设计（`cmd/hc-service/logics/iac-plugin/router/router.go`）

- `router.Service` 增加字段：
  - `poolRegistry *pool.PluginRegistry`
  - `plugins map[string]discovery.Entry`（直接复用 `iacPluginManager` 启动扫描到的只读本地插件索引表）
- **版本锁定寻址（`getTargetPlugin` / `getEnabledTargetPlugin`）**：
  - 查询 `store.List(kt, s.dataCli, pluginName)`；
  - 若 `localEntry, isLocal := s.plugins[pluginName]` 命中，或数据库记录中存在 `Protocol == enumor.IaCPluginProtocolTFProvider`：
    - 若 `!isLocal`（数据库中是 `tf_provider` 但本副本未扫描到该插件目录）：返回 `errf.Newf(errf.Aborted, "tf_provider plugin %s binary not found on this replica", pluginName)`；
    - 在 `records` 中精确查找 `rec.PluginVersion == localEntry.PluginVersion`：若未找到返回 `errf.RecordNotFound`；`getEnabledTargetPlugin` 进一步校验 `rec.State == enumor.IaCPluginStateEnabled`，否则返回 `errf.Aborted`；
  - 否则（三方 `http` 插件）：保持 `iacplugin.Latest(records)` 取最高版本并校验 `state == "enabled"`。
- `**Validate` 类型门控（`ResolveValidate`）**：
  - 对 `Validate` 请求，根据 `req.ConfigType` 精确检查：`ConfigTypeResource` 要求 `schema.ResourceSchemas[typeName] != nil`，`ConfigTypeDataSource` 要求 `schema.DataSourceSchemas[typeName] != nil`。

## Decisions

### 1. 直接复用 `http-plugin/composer.go` 完成配置校验与 `ProposedNewState` 合成

- **选择**：在 `TFPluginAdaptor` 中直接调用 `httpplugin.ValidateConfigAndRecompute` 与 `httpplugin.ComposeProposedState`，不额外修改 `composer.go`，也不在 `adaptor` 包重写 `BuildProposedNewState`。
- **原因**：`composer.go` 已完整实现顶层属性、SDKv2 `NestedBlocks`（Single/Group/List/Set/Map）与 Plugin Framework `AttributeNestedType` 的递归状态合成与 `recompute` 校验（数组下标路径统一采用 `"clouds.names[0].name"` 方括号格式），与本地 `tf_provider` 插件所需的目标状态构建规则 100% 一致。这是与三方 HTTP 插件之间**已经落地的公共逻辑**，实现时必须调用现有导出函数，禁止复制粘贴一套。
- **备选方案**：在 `adaptor/proposed_state.go` 另写一套 `BuildProposedNewState` + `applyRecomputePaths` → 否决，存在两套重复的 Schema 遍历逻辑。

### 2. 漂移检测：复用分支决策语义，不硬抽与 HTTP 共用的整段编排

- **现状对照**（三方 `httpplugin.prepareStateChange` vs 本地 `TFPluginAdaptor.prepareChangeContext`）：
  | 步骤 | 三方 HTTP | 本地 `tf_provider` |
  |---|---|---|
  | 静态校验 `config`/`recompute` | `ValidateConfigAndRecompute` | **同左（直接复用）** |
  | 刷新前额外动作 | RSA 加密 `launch_config` + 建南向 Client | `ValidateResourceTypeConfig`（插件 RPC，仅 `config != nil`） |
  | 漂移刷新调用 | 南向 `cli.Refresh`（仅 state；受 `capabilities.refresh` 门控） | `lease.ReadResource`（返回 `NewState` + `Private` + `Diagnostics`） |
  | 刷新后分支 | 删除+远端空 → 短路；更新+远端空 → `baselinePrior=nil` 转创建；远端非空 → 用刷新态覆盖 | **同一套分支语义**（另记 `drifted`、携带 `currentPrivate`） |
  | 合成目标状态 | `ComposeProposedState` | **同左（直接复用）** |
- **选择**：
  1. `**ProposedNewState` 合成与静态校验**：必须调用 `httpplugin` 既有导出函数，零拷贝复用。
  2. **漂移刷新后的分支决策**（短路 / 转重建 / 覆盖基准态）：语义与 HTTP 一致，可在实现时抽一个**无 I/O 的纯函数**（例如输入 `priorState`、`config`、`refreshedState`，输出 `shortCircuit`、`baselinePrior`、`drifted`），供 `TFPluginAdaptor` 使用；若抽取成本低且签名干净，可顺手让 `httpplugin.prepareStateChange` 也改用该纯函数，但**本变更不强制重构 HTTP 通道**。
  3. **不要**把整段 `prepareStateChange` / `prepareChangeContext` 抽成跨协议公共编排：两侧刷新前的准备工作、刷新 API、错误通道（HTTP 网关 `Aborted` vs TF `ErrDiagnostic`）、以及是否携带 `Private` 均不同，硬抽只会引入回调/接口噪音。
- **实现注意**：写 `prepareChangeContext` 时先对齐 HTTP 的分支表，再补上 TF 独有的 `ValidateResourceTypeConfig`、`Private` 编解码与 `Diagnostics` Error 拦截；合成阶段只调 `ComposeProposedState`，不要再手写字段合并。

### 3. `TFPluginAdaptor` 直接持有 `lease pool.ProviderIF` 成员，且不保存明文 `Properties`

- **选择**：在 `pool.ProviderIF` 中补齐 `Release()` 方法，`TFPluginAdaptor` 直接声明 `lease pool.ProviderIF` 成员，不定义额外的 `LeaseIF` 或测试专用构造函数；在 `NewTFPluginAdaptor` 内部完成 `pool.Acquire(kt, properties)` 后仅将返回的 `*pool.Lease` 存入 `lease`，结构体不设 `properties` 字段。
- **原因**：① 彻底杜绝后续多步 RPC 日志或结构体打印泄露云账号密钥；② `*pool.Lease` 天然实现 `pool.ProviderIF`，生产代码无任何冗余抽象，同包单测与 `BenchmarkTFPluginAdaptorExecute` 可直接构造 `&TFPluginAdaptor{lease: mock}`。

### 4. `Router` 对本地插件严格锁定本副本 `discovery.Entry.PluginVersion`

- **选择**：`Router` 从 `discovery.Entry` 构建只读 `localVersions` 表，对本地 `tf_provider` 插件始终按 `(plugin, localVersion)` 查询 `iac_plugin` 状态与 `PluginSchema`。
- **原因**：滚动升级期间新副本可能已向 `iac_plugin` 注册更高版本 `enabled` 记录，而旧副本本机二进制仍是旧版本；若按 `iacplugin.Latest` 查 Schema 会导致旧副本用新 Schema 操作旧二进制产生编解码错配。

## Risks / Trade-offs

- **[Risk] 多步 RPC 执行中途发生错误提前返回导致进程池** `Lease` **未归还** → **Mitigation**：`Router` 在 `NewTFPluginAdaptor` 成功后立即执行 `defer tfAdaptor.Close()`，且 `Close()` 使用 `sync.Once` 包装 `lease.Release()`，确保任何返回路径下 100% 归还且幂等。
- **[Risk]** `ValidateReq` **新增必填** `config_type` **对既有调用方与测试的影响** → **Mitigation**：同步更新 `pkg/api/hc-service/iac-plugin/northbound_test.go`、`pkg/client/hc-service/iac-plugin/client_test.go`、`cmd/hc-service/logics/iac-plugin/router/router_test.go`、`cmd/hc-service/logics/iac-plugin/http-plugin/executor_test.go` 与 `cmd/hc-service/service/iac-plugin/northbound_test.go` 中所有构造 `ValidateReq` 的调用点。
