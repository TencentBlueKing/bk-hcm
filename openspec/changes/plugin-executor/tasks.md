## 1. 依赖与编解码边界

- [x] 1.1 不引入 `terraform-plugin-go`。继续使用已提交的协议 5.9 桩和 `go-plugin` v1.6.3，不另生成 5.3 客户端。
- [x] 1.2 在 `tfruntime` 定义编解码接口：按 `tfplugin5.Schema` 把调用方对象编码为 `tfplugin5.DynamicValue`，以及按同一 Schema 解码。`ProcessManager` 提供设置该接口的方法，不改 `NewProcessManager` 的现有参数。
- [x] 1.3 未注入编解码时，需要编码或解码的方法返回错误且不发 RPC。`GetProviderSchema` 与 `UpgradeResourceState` 的 `RawState` 不走该接口。

## 2. 协议适配

- [x] 2.1 `GRPCClient` 继续接收 `*grpc.ClientConn` 并持有生成的 Provider 客户端。`PluginExecutor` 不再公开实现 `tfplugin5.ProviderClient`。`GRPCServer` 仍返回错误。
- [x] 2.2 本期方法直接填写生成的 Request。请求不设置 `ProviderMeta`，identity、`ClientCapabilities` 和 deferred 留空。响应中的 DynamicValue 直接交给编解码。Diagnostics 和 `RequiresReplace` 直接返回生成的 `tfplugin5` 结构。Schema 从生成的 Response 适配成对外结果。
- [x] 2.3 协商主版本不是 5 时不返回执行器，沿用进程管理的现有判断。

## 3. 关闭门闩

- [x] 3.1 每个会发 RPC 的执行器方法在内部做前置和后置门闩。`Close` 已开始时方法失败，且插件侧该次 RPC 次数为 0。不提供要求调用方使用的 `Call`。
- [x] 3.2 `StopProvider` 仍由 `Close` 在门闩关闭后调用，不进入这道门闩。`Close` 的停止顺序保持不变。

## 4. Schema 与配置

- [x] 4.1 实现 `GetProviderSchema`：请求无业务字段；把 Provider、资源和数据源 Schema 适配为协议 `version`、block、Attribute 和 NestedBlock。`type` 使用协议中的 TF JSON 类型。`protocol_version` 使用 `pkg/criteria/constant` 的 `TFPluginSchemaProtocolVersion`。`schema_hash` 是十六进制 SHA-256，不加前缀。插件名和版本取自进程登记，版本未登记时为空，不使用日志里的 `unknown`。
- [x] 4.2 有资源 Schema 时 `capabilities` 包含 `apply`、`plan`、`validate`、`import`、`refresh`；有数据源 Schema 时包含 `read_data`。未声明的能力不写键。不包含 `async_apply` 和 `list`，也不拷贝 `ServerCapabilities`。
- [x] 4.3 用适配后 provider、资源 Schema 和数据源 Schema 的规范 JSON 计算 SHA-256，作为 `schema_hash`。Schema 的协议 version 算进去。不把指纹自身、diagnostics、插件名和插件版本算进去。执行器保留协议 Schema，供后续按类型名编码。
- [x] 4.4 实现 `ConfigureProvider`：缓存里没有 Schema 时先调用 `GetSchema` 并缓存，再编码 Config，然后 `PrepareProviderConfig`；不含 Error 时用 `PreparedConfig` 调用 `ConfigureProvider`。`TerraformVersion` 填 `pkg/version` 的 `VERSION`，不据此分支。Schema 请求失败或 Prepare 含 Error 时不调用 Configure。凭证只通过 Config 传递。
- [x] 4.5 Configure 成功前，校验、升级、读取、计划、应用、导入和数据源读取都不发 RPC。

## 5. 资源与数据源方法

- [x] 5.1 实现资源校验、状态升级、读取、计划、应用和导入，以及数据源校验和读取。类型名不是对应 Schema 的 key 时失败且不发 RPC。空的配置和状态保持空值，不编码成空对象。`Private` 按字节原样往返，执行器不持久化。
- [x] 5.2 Diagnostics 含 Error 时该调用失败。只有 Warning 时调用成功，并把每条 summary 写入日志，不写 Config 或状态正文。
- [x] 5.3 资源校验或升级含 Error 时只使该次调用失败，不拦住之后的计划。升级成功只把解码后的状态返回给调用方，不改写后续计划的 `PriorState`。`RawState` 的 JSON 与 Flatmap 只填其中之一。
- [x] 5.4 计划只发计划 RPC。新旧状态都空时不发 RPC。只有新状态按创建，两者都有按更新，只有旧状态按删除且返回的 `PlannedState` 为空。`RequiresReplace` 只返回路径。
- [x] 5.5 删除应用返回的 `NewState` 为空。Apply 含 Error 时调用失败，同时把插件返回的 `NewState` 交给调用方。
- [x] 5.6 导入期间不调用 `ReadResource`，也不创建资源。返回项按其自身类型名解码；类型不在 Schema 中时该次调用失败，不改用请求类型的 Schema。数据源校验含 Error 时不拦住之后的读取；读取不调用计划和 Apply。
- [x] 5.7 不向调用方暴露，也不调用 `GetMetadata`、identity、`MoveResourceState`、ephemeral、functions、`async_apply` 和 `list`。

## 6. 验证

- [ ] 6.1 用假 Provider 客户端和假编解码器覆盖：协商非 5 不返回执行器、关闭后门闩、Prepare 含 Error 不 Configure、Configure 前读取次数为 0、编解码拒绝时 RPC 次数为 0、Warning 只记 summary。
- [ ] 6.2 覆盖 capabilities 两种 Schema 组合、`schema_hash` 稳定性和变化、凭证经 Config 到达插件、校验或升级失败后计划仍发出、升级成功后计划仍使用调用方传入的 `PriorState`。
- [ ] 6.3 覆盖计划的创建、更新、删除和双空，`RequiresReplace` 不触发 Apply，删除应用的空状态，Apply 失败仍返回 `NewState`，导入不调用 Read，数据源校验失败不读取。
- [ ] 6.4 确认本变更不修改既有资源 Handler 与 `pkg/adaptor`，不注册北向路由，不新增时延指标。
