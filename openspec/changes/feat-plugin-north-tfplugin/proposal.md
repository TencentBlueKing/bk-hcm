# Proposal

## Why

在海垒（HCM）IaC 插件化架构中，上游控制面（`iac-server`、`task-server`、`cloud-server`）统一通过 `hc-service` 暴露的北向 HTTP 协议（`/api/v1/hc/iac_plugins/*`）驱动插件执行。目前三方 HTTP 插件（`protocol = "http"`）的北向到南向执行通道及状态合成组件（`composer.go`）已就绪，本地 Terraform Provider 插件的底层 gRPC 执行器（`tfruntime/executor`）与严格 FIFO 进程池（`tf-plugin/pool`）也已落地，但 `Router` 中针对本地插件（`protocol = "tf_provider"`）的运行态分支仍处于未接入状态（返回 `protocol tf_provider is not supported yet`）。

由于官方/社区 Terraform Provider 插件（基于 `tfprotov5` gRPC 协议）提供的是细粒度底层原语接口，无法直接对齐北向高层语义：单次北向请求需在同一已配置进程租约（`*pool.Lease`）内串行编排多步 gRPC 调用（如 `Execute` 需串行完成配置预校验、漂移检测刷新、目标状态构建、变更计划计算与变更应用），在滚动升级新旧副本并存期间需按本副本启动锁定的二进制版本（`localVersion`）查询状态与 `PluginSchema`，对同名资源与数据源需通过 `config_type` 精确分流校验，且状态挽回与诊断封装需严格对齐 `Terraform Core` 语义。因此亟需建设 `hc-service` 本地插件运行态适配层（`TFPluginAdaptor`）并打通 `Router` 本地插件路由分发（对应需求单 `1069995598138774200`）。

## What Changes

### 本期包含（Scope In）

1. **北向协议 `Validate` 请求与状态/数据源响应契约扩展（对应 F-004 / 规则 R-007）**：
   - **BREAKING**：在 `pkg/api/hc-service/iac-plugin/northbound.go` 的 `ValidateReq` 中新增必填字段 `ConfigType string`（`json:"config_type"`，枚举值 `"resource"` 或 `"data"`），并在 `ValidateReq.Validate()` 中执行非空与合法枚举校验，解决同一插件中同名 `type_name` 同时存在于 `ResourceSchemas` 与 `DataSourceSchemas` 时的校验分流问题；
   - 在 `pkg/api/hc-service/iac-plugin/northbound.go` 的 `StateResult`（`POST /import`、`/refresh`）与 `ReadDataResult`（`POST /read_data`）中新增 `Diagnostics []coreplugin.Diagnostic`（`json:"diagnostics,omitempty"`）字段，用于承载插件告警/错误诊断以及宿主对齐 `Terraform Core` 主动封装的错误诊断。
2. **本地插件运行态适配器 `TFPluginAdaptor` 与协议数据转换（对应 F-001 ~ F-004 / 规则 R-002 ~ R-006）**：
   - 在 `cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor/` 新增 `TFPluginAdaptor`（不持有明文 `Properties`，仅记录 `pluginName`、`instanceID`、`schema` 与进程租约 `lease`），构造时从 `*pool.PluginPool` 借出已完成 `ConfigureProvider` 的 `*pool.Lease`，并通过 `Close()` 幂等调用 `lease.Release()` 归还并发槽位；
   - 实现 `prepareChangeContext`：在同一 `Lease` 上串行完成前置配置静态校验（直接复用 `httpplugin.ValidateConfigAndRecompute`）、`config != nil` 时的 `ValidateResourceTypeConfig` 预校验（含 `Error` 诊断时立即终止，不进入 `ReadResource` 与 `PlanResourceChange`）、`prior_state != nil` 时的 `ReadResource` 漂移检测刷新（处理删除时远端已不存在的短路成功、更新时远端已不存在的自动转重建、远端存在时的基准状态更新与 `drifted` 标记），以及直接复用 `httpplugin.ComposeProposedState`（数组下标采用方括号格式如 `"clouds.names[0].name"`）构建 `proposedNewState`；
   - 实现 `Execute` 与 `Plan` 多步 gRPC 串行编排：支持 `Execute` 在 `PlanResourceChange` 或 `ApplyResourceChange` 失败时的三级状态挽回（优先返回 `ApplyResourceChange` 部分落地的非空 `NewState`，其次返回前置漂移刷新的 `currentState`，否则返回 `nil`），以及 `Plan` 响应中 `codec.Unknown` 递归清理为 `nil` 与 `RequiresReplace`（`[]*tfplugin5.AttributePath`）到路径切片 `ReplacePaths` 的转换；
   - 实现 `Import`（`ImportResourceState` + `ReadResource` 两步补全，当 `ReadResource` 返回 `NewState == nil` 且无 Error 诊断时主动封装 `"cannot import non-existent remote object"` 错误诊断）、`ReadData`（`ValidateDataSourceConfig` + `ReadDataSource` 两步编排，当 `State == nil` 或残留 `codec.Unknown` 时主动封装对应错误诊断）、`Validate`（按 `config_type` 分流调用 `ValidateResourceTypeConfig` 或 `ValidateDataSourceConfig`）及 `Refresh`（单步 `ReadResource` 刷新，远端已删除时返回 `state: nil` 且不报错）；
   - 实现 `Private` 标准 Base64 编解码、`*tfplugin5.Diagnostic` 到 `coreplugin.Diagnostic`（含 `PathStep`）转换、两类错误通道严格隔离（非 `executor.ErrDiagnostic` 的底层进程/通信故障返回外层 `errf.Aborted`，插件业务/配置诊断错误返回外层 `code = 0`）。
3. **`Router` 本地 `tf_provider` 插件 `localVersion` 寻址、门控与分发接入（对应 F-001 / 规则 R-001）**：
   - 扩展 `cmd/hc-service/logics/iac-plugin/router` 与 `cmd/hc-service/service/iac_plugin.go`，向 `router.Service` 注入本副本扫描的 `[]discovery.Entry` 与 `*pool.PluginRegistry`；
   - 当请求的插件为本副本本地插件（命中本副本 `discovery.Entry`）或数据库中协议为 `tf_provider` 时，严格按本副本锁定的 `localVersion` 查询 `iac_plugin` 表校验 `state == "enabled"`，并按 `(plugin, localVersion)` 调用 `SchemaManager.GetSchema` 获取与本机二进制严格一致的 `PluginSchema`；
   - 在 `Validate` 路由门控中按 `req.ConfigType`（`"resource"` | `"data"`）精确校验 `type_name` 存在于 `ResourceSchemas` 或 `DataSourceSchemas`；
   - 实现本地插件 `Health` 探活（校验本副本 `discovery.Entry` 存在且 `(plugin, localVersion)` 状态为 `enabled`，返回 `{"plugin_version": localVersion}`）、能力位门控拦截 `GetResult` / `Stop` / `List`（不借出 `Lease`），以及 6 个运行态接口构造 `TFPluginAdaptor`（配合 `defer adaptor.Close()`）执行分发。

### 本期不包含（Scope Out）

- 三方 HTTP 插件南向执行通道与 `composer.go` 状态合成器本身改动（已在兄弟需求 `1069995598139285207` 中实现，本变更直接复用 `composer.go`）；
- 本地插件进程池内部 FIFO 调度算法与底层 `tfruntime/executor`、`tfruntime/codec` gRPC 协议编解码实现（已在兄弟需求 `1069995598139391605` 与 `1069995598138774196` 中实现）；
- `UpgradeResourceState` 跨 Schema 版本状态迁移；
- 本地插件异步任务表、`GetResult`、`Stop`、`List` 接口实现（TF 协议 5.3 为同步执行且不支持 List RPC）。

## Capabilities

### New Capabilities
- `tf-plugin-northbound-adaptor`: 本地 Terraform Provider（`tf_provider`）插件北向运行态适配器（`cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor`），封装单次请求内进程池 `Lease` 的借出与幂等归还、`prepareChangeContext`（前置配置预校验、漂移检测刷新、删除短路/更新转重建与复用 `ComposeProposedState` 合成目标状态）、`Execute`/`Plan`/`Import`/`ReadData`/`Validate`/`Refresh` 六大运行态接口多步 gRPC 串行编排、三级失败状态挽回、`codec.Unknown` 清理、`RequiresReplace` 路径转换、`Private` Base64 编解码及 `Diagnostics` 转换与主动封装。

### Modified Capabilities
- `plugin-northbound-contract-sdk`: 扩展北向 `ValidateReq` 增加必填字段 `config_type`（`"resource"` | `"data"`）及校验规则；扩展 `StateResult` 与 `ReadDataResult` 增加 `diagnostics` 字段。
- `plugin-instance-router`: 扩展 `router.Service` 支持本地 `tf_provider` 插件按本副本 `localVersion` 锁定寻址与状态/Schema 校验、按 `ValidateReq.ConfigType` 精确校验资源或数据源 Schema、本地插件 `Health` 探活、不支持接口前置拦截以及通过 `TFPluginAdaptor` 分发 6 个运行态接口。

## Impact

- `pkg/api/hc-service/iac-plugin/` — `northbound.go`、`northbound_test.go`（`ValidateReq` 新增 `ConfigType` 及常量 `ConfigTypeResource`/`ConfigTypeDataSource`，`StateResult` 与 `ReadDataResult` 新增 `Diagnostics`）
- `pkg/client/hc-service/iac-plugin/` — `client_test.go`（同步更新 `ValidateReq` 测试用例中的 `ConfigType`）
- `cmd/hc-service/logics/iac-plugin/http-plugin/` — `executor_test.go`（同步更新 `ValidateReq` 测试用例中的 `ConfigType`）
- `cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor/` — 新增 `tf_adaptor.go`、`convert.go`、`tf_adaptor_test.go`、`benchmark_test.go`（实现 `TFPluginAdaptor` 六大运行态接口编排、转换辅助函数、全分支单测、50 协程 `-race` 并发单测、零敏感日志泄露单测与 P99 < 5ms Benchmark）
- `cmd/hc-service/logics/iac-plugin/router/` — `router.go`、`router_test.go`（支持本地插件 `localVersion` 锁定寻址、`Validate` 按 `ConfigType` 分流校验、本地插件 `Health` 探活与 6 个运行态接口 `TFPluginAdaptor` 分发）
- `cmd/hc-service/service/` & `cmd/hc-service/service/iac-plugin/` — `iac_plugin.go`、`northbound_test.go`（调整 `poolRegistry` 初始化顺序并将 `poolRegistry` 与 `s.plugins` 注入 `router.New`，更新 Handler 测试用例）
