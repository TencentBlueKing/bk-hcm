## ADDED Requirements

### Requirement: 契约包与依赖边界
系统 SHALL 在 `cmd/hc-service/logics/tfruntime` 提供对接方（进程管理与执行器）的契约类型，作为进程池依赖的唯一一份定义。该包在本 change 内 MUST 只包含接口、结构体、常量、错误变量和返回"未实现"的桩方法，MUST NOT 包含握手、gRPC、go-plugin、协议适配、DynamicValue 编解码、日志落盘的实现，MUST NOT 新增第三方依赖，MUST NOT 修改 `go.mod`。阶段零到三 MUST 只新增文件：不得修改任何既有文件，尤其不得修改 P2 注册相关代码。阶段四（接入）可修改的既有文件限于 `service/iac_plugin.go`（只新增调用）、`pkg/cc`、`hc_service.yaml` 与 helm，仍 MUST NOT 修改 P2 `register` / `watch` 的既有逻辑。

#### Scenario: 包可编译且无新依赖
- **WHEN** 阶段零合入后执行构建
- **THEN** 构建通过，`go.mod` 与 `go.sum` 无变化

#### Scenario: 不修改既有文件
- **WHEN** 阶段三完成时，对比实现分支与其基线 `upstream/feat-plugin` 的差异
- **THEN** 差异中只有新增文件（`tfruntime`、`pool`、本 change 的 openspec 文档），没有被修改或删除的既有文件

### Requirement: 执行器契约
`Executor` 接口 SHALL 是对接方案第 6.2 节执行器的完整方法集，不拆分子接口：`GetProviderSchema`、`ConfigureProvider`，以及业务方法 `ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`。每个方法 MUST 以 `*kit.Kit` 为首参，并有独立的请求与响应结构体，字段与对接方案第 6.2 节一一对应（例如 `PlanResourceChange` 的请求含 `TypeName`、`Config`、`PriorState`、`ProposedNewState`、`PriorPrivate`，响应含 `PlannedState`、`RequiresReplace`、`PlannedPrivate`、`Diagnostics`）。`GetProviderSchema` MUST 返回 `*ProviderSchema`，`ProviderSchema` MUST 是对基线已有的 `pkg/thirdparty/api-gateway/plugin.PluginSchema` 的类型别名，MUST NOT 另定义一套 Schema 结构；合并其他分支时只调整该别名。契约层 MUST NOT 区分"对业务方开放"与"仅内部使用"的方法，该区分由上游（进程池）封装。`Config`、`State` 类 MUST 使用 `Object`（`map[string]any`，nil 表示 null）；未知值 MUST 以 `Unknown` 类型表达；`Private` MUST 为 `[]byte`。契约 MUST NOT 暴露 `tftypes` 或 `tfprotov5` 类型。

#### Scenario: 执行器方法集完整
- **WHEN** 检查 `Executor` 接口
- **THEN** 同时包含 `GetProviderSchema`、`ConfigureProvider` 与 8 个业务方法，且不存在 `ProviderIF` 之类的子接口

#### Scenario: 删除计划的表达
- **WHEN** 构造一个删除操作的 `PlanResourceChangeRequest`
- **THEN** `PriorState` 有值而 `ProposedNewState` 为 nil，结构体能表达该状态而不需要额外字段

#### Scenario: Private 原样往返
- **WHEN** 响应中带 `Private`
- **THEN** 其类型为字节切片，契约不要求也不提供解码

### Requirement: Diagnostics 与错误分类
契约 SHALL 定义 `Diagnostic`（`Severity`、`Summary`、`Detail`、`Attribute`）与 `Severity` 取值（`Error`、`Warning`）。当执行器某次调用的响应含 `Error` 级 Diagnostic 时，实现 MUST 返回 `*DiagnosticsError`（业务错误）；超时、连接断开、进程退出等 MUST 返回普通 `error`（传输错误），二者 MUST 可通过 `errors.As` 区分。`Warning` 级 Diagnostic MUST NOT 使调用失败。`DiagnosticsError` 的 `Error()` 文本 MUST NOT 包含请求中的 `Config` 内容。

#### Scenario: 业务错误可识别
- **WHEN** `ConfigureProvider` 因缺少密钥被插件以 Error 级 Diagnostic 拒绝
- **THEN** 返回的错误 `errors.As(*DiagnosticsError)` 为真

#### Scenario: 传输错误可识别
- **WHEN** `ConfigureProvider` 因超时失败
- **THEN** 返回的错误 `errors.As(*DiagnosticsError)` 为假

#### Scenario: Warning 不失败
- **WHEN** 响应只含 Warning 级 Diagnostic
- **THEN** 调用视为成功

### Requirement: 进程管理契约
`ProcessManager` 接口 SHALL 提供 `Start(kt) (Executor, error)`、`Close(ctx) error`、`Exited() bool`、`CheckHealth() error`。`Start` MUST 返回执行器接口，进程管理 MUST NOT 代理业务方法。`Close` MUST 为阻塞调用并可重复调用，成功表示进程已退出，ctx 超时后 MUST 强制结束进程；`Close` 开始后已返回的执行器 MUST NOT 再接受调用。`Exited` 与 `CheckHealth` MUST NOT 阻塞。契约 SHALL 提供 `ProcessConfig`（插件名、版本、可执行路径、参数、工作目录、环境变量、日志目录）、`NewProcessManager(cfg)` 工厂，以及 `Starter` 接口（签名 `NewProcessManager(kt, pluginName, pluginVersion, execPath)`，与 P2 分支的 `PluginProcessStarter` 保持一致，便于合并时直接替换）。凭证 MUST NOT 通过 `ProcessConfig.Env` 传递，只能通过 `ConfigureProvider`。

#### Scenario: 创建管理对象
- **WHEN** 调用 `NewProcessManager` 传入合法 `ProcessConfig`
- **THEN** 返回满足 `ProcessManager` 接口的对象，尚未拉起进程

#### Scenario: 重复关闭
- **WHEN** 对同一个 `ProcessManager` 先后调用两次 `Close`
- **THEN** 第二次不报错、不重复结束进程

#### Scenario: 凭证不进环境变量
- **WHEN** 审查 `ProcessConfig` 与 `Starter` 的字段
- **THEN** 不存在用于传递云账号凭证的字段

### Requirement: 桩实现的未实现语义
本 change 提供的 `processManager` 桩 SHALL 对未实现的方法一律返回 `ErrNotImplemented`：`Start` 返回 `(nil, ErrNotImplemented)`，`CheckHealth` 返回 `ErrNotImplemented`；`Exited` 返回 `false`；`Close` 幂等并返回 `nil`。桩 MUST NOT 提供假的 `Executor`，调用方 MUST 能通过 `errors.Is(err, ErrNotImplemented)` 判断。进程池的单元测试 MUST 使用测试文件内的假实现，而不是依赖本桩。

#### Scenario: 桩的 Start
- **WHEN** 对桩调用 `Start`
- **THEN** 返回 `ErrNotImplemented`，不拉起任何进程

#### Scenario: 池遇到桩
- **WHEN** 池通过默认 `Starter` 创建进程并调用 `Start`
- **THEN** 触发的请求收到可通过 `errors.Is(err, ErrNotImplemented)` 识别的错误，不 panic，进程被关闭、容量被释放
