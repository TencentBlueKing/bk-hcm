## Why

进程管理已经能拉起协议主版本 5 的插件进程，但 `PluginExecutor` 上除 `Stop` 以外的 Provider RPC 仍返回未实现。进程池需要一个执行器来完成 Schema 读取、配置写入、资源操作和数据源读取，否则调度层只能看到连接，无法调用官方插件。

本次变更对应 TAPD 需求「执行器实现」（story 139221716），依据 iWiki《技术方案》第 4 节与第 6.2 节。这些方法的差别是调用顺序和 Diagnostics，作为同一个执行器交付。

## What Changes

- 在已协商主版本为 5 的连接上补齐调用：`GRPCClient` 持有 `pkg/plugin/tfplugin5` 生成的 Provider 客户端，执行器直接填写生成的 Request。`GRPCServer` 继续返回错误。协商结果不是 5 时仍不向调用方返回执行器，沿用进程管理里的判断。
- `PluginExecutor` 对外改为方案第 6.2 节的方法：`GetProviderSchema`、`ConfigureProvider`、`ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`。这些方法和 `Close` 组成执行器包里的 `Executor` 接口，出入参放在 `cmd/hc-service/types`。生成桩的收发放到未导出调用中，并补上关闭门闩。`Start` 返回这个接口。
- `GetProviderSchema` 把协议 Schema 适配成统一结构，并填写 `protocol_version`、插件名、插件版本、`schema_hash` 和 `capabilities`。`ConfigureProvider` 对上只有一个方法，内部先 `PrepareProviderConfig`，没有 Error 再调用 `ConfigureProvider`。
- 资源路径提供校验、升级、读取、计划、应用和导入。数据源路径提供校验和读取。`Error` 只使该次调用失败，不拦住之后另一次调用；`Warning` 只记日志。`Private` 按字节原样往返。
- 配置和状态在发给插件前交给编解码，响应在返回调用方前解码。编解码拒绝入参时不发出该次 RPC。本变更不实现 msgpack 和按 Schema 解码。
- 不改现有 `service` → `logics` → `pkg/adaptor` 路径，不改已有资源 Handler，不实现北向入口和进程池。

## Capabilities

### New Capabilities

- `plugin-executor-client`: 把协议主版本 5 的连接适配成可调用的 Provider，执行器方法在关闭开始后不再向插件发 RPC。
- `plugin-schema-configure`: 读取并适配 Provider Schema，按 Prepare 然后 Configure 的顺序写入配置，成功前不调用资源和数据源方法。
- `plugin-resource-data-source`: 执行资源与数据源方法，按 Diagnostics 和空状态决定是否发出后续 RPC。

### Modified Capabilities

无。`openspec/specs/` 下没有执行器、Schema 适配或资源协议方法的正式能力。进程管理变更里「本变更不封装执行器 RPC」只约束那次变更，不改写。

## Impact

- 代码位于 `cmd/hc-service/logics/tfruntime` 和 `cmd/hc-service/types`。具体执行器不再实现 `tfplugin5.ProviderClient` 的公开方法集；`ProcessManager.Start` 返回 `executor.Executor`，仓库内还没有其他调用方。
- 不引入 `terraform-plugin-go`。执行器内部使用已提交的协议 5.9 `tfplugin5` 结构，不另生成 5.3，也不再把生成结果转成 `tfprotov5`。
- 编解码由依赖需求「DynamicValue 编解码」（story 139221653）提供。执行器只定义调用边界：编码失败则 RPC 次数为 0。
- 不新增北向 HTTP API、表结构或进程环境变量。凭证只出现在 `ConfigureProvider` 的 Config 中。
- 不影响现有云资源同步、账单和账号等 adaptor 调用。
