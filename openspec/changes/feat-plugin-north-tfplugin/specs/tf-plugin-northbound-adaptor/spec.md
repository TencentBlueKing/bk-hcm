# Spec Delta

## Purpose

定义 HCM `hc-service` 本地 Terraform Provider（`tf_provider`）插件北向运行态适配器（`cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor`）的进程池租约生命周期管理、前置配置预校验与漂移检测状态机、`Execute`/`Plan`/`Import`/`ReadData`/`Validate`/`Refresh` 六大运行态接口多步 gRPC 串行编排、三级失败状态挽回、`codec.Unknown` 清理、`RequiresReplace` 路径转换、`Private` Base64 编解码及 `Diagnostics` 结构化转换与主动封装规范。

## ADDED Requirements

### Requirement: 进程池 Lease 借出、单次请求内复用与零泄漏幂等归还（规则 R-001）
系统 SHALL 在每次本地插件运行态请求中通过 `NewTFPluginAdaptor` 从目标 `*pool.PluginPool` 调用 `Acquire(kt, properties)` 借出已完成 `ConfigureProvider` 的进程租约 `Lease`，并在单次请求内所有 gRPC 调用中复用同一 `Lease`：
- `TFPluginAdaptor` MUST NOT 保存明文 `properties`，日志与错误中 MUST NOT 出现实例明文凭证；
- 请求结束时 MUST 通过 `Close()` 幂等调用 `lease.Release()` 100% 归还槽位；
- 当池未就绪或 `Acquire` 失败时，系统 MUST 返回外层非 0 错误（`errf.Aborted` 或 `errf.InvalidParameter`）。

#### Scenario: 单次 Execute 请求内多次 RPC 复用同一 Lease 并在结束时归还
- **WHEN** 调用本地插件的 `Execute` 接口依次触发 `ValidateResourceTypeConfig`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`
- **THEN** 4 次调用均落在同一次 `Acquire` 借出的 `Lease` 上，且请求返回后 `lease.Release()` 调用次数恰好为 1（重复调用 `Close()` 不重复释放）

#### Scenario: Acquire 失败或 RPC 报错时零凭证泄露且 Lease 零泄漏
- **WHEN** 实例 `properties` 含敏感值 `"SK_TF_SUPER_SECRET_999"`，且发生 `Acquire` 超时/配置失败或后续 RPC 报错提前返回
- **THEN** 已借出的 `Lease` 100% 归还，且全部日志输出与错误返回文本中 `"SK_TF_SUPER_SECRET_999"` 出现次数为 0

### Requirement: Execute 与 Plan 前置配置预校验拦截（规则 R-002）
系统 SHALL 在 `Execute` 与 `Plan` 的 `prepareChangeContext` 阶段，当 `config != nil` 时先调用 `ValidateConfigAndRecompute` 静态校验，再调用 `lease.ValidateResourceTypeConfig(ctx, typeName, config)`：
- 若静态校验不通过或 `prior_private` Base64 解码失败，MUST 直接返回 `errf.InvalidParameter`（`code=2000001`），不调用底层 RPC；
- 若 `ValidateResourceTypeConfig` 返回 `Severity == Error` 的诊断，MUST 立即终止后续流程（`ReadResource`、`PlanResourceChange`、`ApplyResourceChange` 调用次数为 0），返回外层 `code=0` 及 `diagnostics`；
- 若仅含 `Warning` 或无诊断，MUST 保留告警诊断并继续后续流程。

#### Scenario: ValidateResourceTypeConfig 返回 Error 诊断时前置终止不进入后续 RPC
- **WHEN** 调用 `Execute` 或 `Plan` 且 `config != nil`，底层 `ValidateResourceTypeConfig` 返回 `Severity == Error` 的诊断（如枚举值非法）
- **THEN** `ReadResource`、`PlanResourceChange`、`ApplyResourceChange` 调用次数均为 0，`Execute` 返回外层 `code=0, data.status="failed", data.diagnostics=[...]`，`Plan` 返回外层 `code=0, data.state=nil, data.diagnostics=[...]`

### Requirement: 漂移检测刷新、删除短路与更新自动转重建（规则 R-003 & R-004）
系统 SHALL 在 `prepareChangeContext` 中对 `prior_state != nil` 的请求调用 `lease.ReadResource` 刷新最新状态 `NewState`，并调用 `ComposeProposedState` 合成 `proposedNewState`：
- 若 `NewState == nil` 且为删除（`config == nil`），MUST 直接短路返回成功（`state=nil`），不再调用后续 RPC；
- 若 `NewState == nil` 且为更新（`config != nil`），MUST 将 `currentState` 与 `currentPrivate` 置为 `nil`、标记 `drifted = true` 并转为创建；
- 若 `NewState != nil`，MUST 以 `NewState` 与新 `Private` 更新基准状态并记录 `drifted`。

#### Scenario: 删除操作在 ReadResource 发现云端对象已不存在时短路成功
- **WHEN** 调用 `Execute` 或 `Plan` 执行删除（`prior_state != nil, config == nil`），且 `lease.ReadResource` 返回 `NewState == nil`
- **THEN** `PlanResourceChange` 与 `ApplyResourceChange` 调用次数为 0，`Execute` 返回 `code=0, data.status="succeeded", data.state=nil`，`Plan` 返回 `code=0, data.state=nil, data.replace_paths=[]`

#### Scenario: 更新操作在 ReadResource 发现云端对象已不存在时自动转为创建
- **WHEN** 调用 `Execute` 或 `Plan` 执行更新（`prior_state != nil, config != nil`），且 `lease.ReadResource` 返回 `NewState == nil`
- **THEN** 系统将 `currentState` 与 `currentPrivate` 置为 `nil`、标记 `drifted=true`，按创建规则合成 `proposedNewState` 并以 `PriorState=nil` 调用 `PlanResourceChange`

### Requirement: Execute 与 Plan 变更编排、三级失败状态挽回与 Unknown/ReplacePaths 转换（规则 R-005）
系统 SHALL 在 `prepareChangeContext` 后调用 `PlanResourceChange`（并在 `Execute` 中调用 `ApplyResourceChange`）：
- 非 `executor.ErrDiagnostic` 的底层错误 MUST 返回外层 `errf.Aborted`（`code=2000006, data=nil`）；
- `Execute` 因 `Error` 诊断失败时返回外层 `code=0, data.status="failed"`，`data.state` 按三级优先级挽回：① `ApplyResourceChange` 非空 `NewState`；② `drifted == true` 时的 `currentState`；③ `nil`；
- `Plan` 与 `Execute` MUST 将状态中 `codec.Unknown` 递归替换为 `nil`，将 `RequiresReplace` 转为点分路径 `replace_paths`。

#### Scenario: Execute 在 Apply 阶段部分落地失败或前置漂移后失败时按三级优先级挽回状态
- **WHEN** ① `ApplyResourceChange` 返回 `Error` 诊断但同时返回非空 `NewState`；或 ② `ApplyResourceChange` 返回 `Error` 诊断且 `NewState == nil`，但前置 `ReadResource` 检测到 `drifted == true`；或 ③ 首次创建失败且 `NewState == nil`
- **THEN** 外层均返回 `code=0, data.status="failed"`，且 `data.state` 在场景 ① 中为 `applyResp.NewState`、场景 ② 中为刷新后的 `currentState`、场景 ③ 中为 `nil`

#### Scenario: Plan 将 PlannedState 中的 codec.Unknown 转为 nil 并转换 RequiresReplace
- **WHEN** 调用 `Plan` 且 `PlanResourceChange` 返回的 `PlannedState` 中 `id` 与 `cos_bucket_url` 为 `codec.Unknown`，`RequiresReplace` 包含属性路径 `[AttributeName("bucket")]`
- **THEN** 返回外层 `code=0`，`data.state` 中 `id` 与 `cos_bucket_url` 均为 `nil`，`data.replace_paths` 为 `["bucket"]`

### Requirement: Import 两步 RPC 补全与 Refresh 单步刷新适配（规则 R-006）
系统 SHALL 在 `TFPluginAdaptor` 中实现受管资源的 `Import` 与 `Refresh` 接口：
- `Import` 在同一 `Lease` 上串行调用 `ImportResourceState` 与 `ReadResource` 补全属性；当 `len(Imported) == 0` 或 `ReadResource` 返回 `NewState == nil` 且无 Error 诊断时，MUST 主动封装 Error `Diagnostic`（分别为 `"provider returned no imported resource"` 与 `"cannot import non-existent remote object"`）返回（外层 `code=0, data.state=nil`）；
- `Refresh` 解码 `req.Private` 并单步调用 `ReadResource`，若 `NewState == nil` 且无 Error 诊断则正常返回 `code=0, data.state=nil`。

#### Scenario: Import 两步补全成功与云端对象不存在时的 Error Diagnostic 封装
- **WHEN** 调用 `Import` 导入资源：① `ImportResourceState` 返回初始主键状态且 `ReadResource` 返回完整非空 `NewState`；或 ② `ImportResourceState` 成功但 `ReadResource` 返回 `NewState == nil` 且无 Error 诊断
- **THEN** ① 返回外层 `code=0`、补全后的 `data.state` 与 Base64 `data.private`；② 返回外层 `code=0, data.state=nil`，且 `data.diagnostics` 包含 `severity="error", message="cannot import non-existent remote object"`

#### Scenario: Refresh 刷新已存在或云端已删除资源
- **WHEN** 调用 `Refresh` 刷新资源，底层 `ReadResource` 返回最新 `NewState` 或返回 `NewState == nil`（云端已删除且无 Error 诊断）
- **THEN** 返回外层 `code=0` 及最新 `data.state`（已删除时 `data.state=nil` 且 `data.diagnostics` 无 Error 诊断）

### Requirement: ReadData 两步编排与 Validate 按 config_type 分流适配（规则 R-006 & R-007）
系统 SHALL 在 `TFPluginAdaptor` 中实现 `ReadData` 与 `Validate` 接口：
- `ReadData` 在同一 `Lease` 上先调用 `ValidateDataSourceConfig`（含 Error 诊断则立即终止返回），再调用 `ReadDataSource`；若 `State == nil` 且无 Error 诊断或 `State` 含 `codec.Unknown`，MUST 主动封装对应 Error `Diagnostic`（并将 `codec.Unknown` 递归替换为 `nil`）返回；
- `Validate` 按 `req.ConfigType`（`"resource"` | `"data"`）分流调用 `ValidateResourceTypeConfig` 或 `ValidateDataSourceConfig`，返回外层 `code=0` 及 `diagnostics`。

#### Scenario: ReadData 前置校验拦截与空状态或未知值时的 Error Diagnostic 封装
- **WHEN** 调用 `ReadData` 查询数据源：① `ValidateDataSourceConfig` 返回 Error 诊断；或 ② `ReadDataSource` 返回 `State == nil`（无 Error 诊断）；或 ③ `ReadDataSource` 返回的 `State` 含有 `codec.Unknown`
- **THEN** ① `ReadDataSource` 调用次数为 0 并返回校验诊断；② `data.diagnostics` 包含 `"provider produced a null state for data source"`；③ `State` 中的 `codec.Unknown` 被替换为 `nil` 且 `data.diagnostics` 包含 `"provider produced a state value that is not wholly known"`

#### Scenario: Validate 按 config_type 精确分流校验资源或数据源
- **WHEN** 同名 `type_name` 同时存在于 `ResourceSchemas` 与 `DataSourceSchemas` 时，分别以 `config_type="resource"` 和 `config_type="data"` 调用 `Validate`
- **THEN** 分别精确调用 `ValidateResourceTypeConfig` 与 `ValidateDataSourceConfig` 并返回外层 `code=0` 及对应的 `data.diagnostics`
