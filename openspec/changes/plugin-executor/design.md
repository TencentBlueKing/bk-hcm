## Context

`terraform-plugin/20261021-plugin-process-manager` 已经用 `go-plugin` 拉起插件进程。`ProcessManager.Start` 在协商主版本为 5 时返回 `PluginExecutor`。该类型目前断言实现 `tfplugin5.ProviderClient`：`Stop` 会发到生成的客户端，其余 RPC 返回 `errRPCNotImplemented`，门闩记在 TODO 里。桩代码是 `pkg/plugin/tfplugin5` 中协议头为 5.9 的生成结果。仓库还没有 `terraform-plugin-go`，也没有 DynamicValue 编解码。

技术方案把调用链定成：进程池 → 进程管理 → 执行器 → Provider。执行器持有生成的 Provider 客户端，在公开方法上做校验、编解码和 Diagnostics。gRPC 收发走本仓库生成的客户端，执行器直接填写生成的 Request。

编解码是另一张子需求。执行器在发 RPC 前调用它，不在本变更里写 msgpack。既有云 API 路径是 `service` → `logics` → `pkg/adaptor`，本执行器与它并行。

## Goals / Non-Goals

**Goals:**

- 在协议主版本 5 的连接上提供方案第 6.2 节的执行器方法。
- 执行器直接使用生成的 `tfplugin5` 结构，`GRPCServer` 继续返回错误。
- `GetProviderSchema` 返回适配后的统一 Schema，并填写协议文档版本、插件标识、指纹和能力位。
- `ConfigureProvider` 内部按 Prepare、Configure 的顺序执行；成功前不发出资源和数据源 RPC。
- 按空状态决定本次计划是否发出。一次调用的 Error 不拦住之后另一次调用。升级结果只返回给调用方。`Private` 按字节原样往返。
- 业务方法补上关闭门闩，与现有 `Close` 顺序衔接。

**Non-Goals:**

- 不改进程拉起、握手、日志文件、环境变量和 `Close` 的停止顺序。
- 不实现 DynamicValue 的类型构造、msgpack 编码和按 Schema 解码。
- 不实现北向 `service/plugin`、进程池 `logics/plugin`、`ProviderMeta`、`async_apply`、`list`。
- 不调用 identity、ephemeral、functions、`MoveResourceState`、`GetMetadata`。
- 不替换 `pkg/adaptor`，不修改已有资源 Handler。
- 不新增时延或吞吐指标。

## Decisions

### 1. 继续使用现有 PluginExecutor，公开方法改为第 6.2 节

`Start` 返回 `executor.Executor`。这个接口放在执行器包，包含第 6.2 节方法和 `Close`。具体类型仍是 `executor.PluginExecutor`，`Attach` 和 `SetCodec` 只给进程管理接线，不进入接口。公开方法的出入参放在 `cmd/hc-service/types`。该具体类型不再实现 `tfplugin5.ProviderClient`。生成桩保存在未导出字段里。同名方法不能同时使用生成的 Request 和 schema 对象，所以公开签名采用后者。

备选是新增第二个执行器类型，再改 `Start` 的返回值。调用方还没有出现，两种做法都能编译，但第二个类型会把关闭、门闩和进程句柄再包一层。进程管理已经把句柄命名为执行器，直接补齐方法即可。

`Close` 与 `Stop` 保持现有顺序：拒绝新方法、等待在途方法、`StopProvider`、`Kill`。`StopProvider` 不进入业务门闩。

### 2. 协议适配放在执行器内部，只使用生成的 tfplugin5

不引入 `terraform-plugin-go`。执行器直接填写 `tfplugin5` 的 Request，并缓存生成的 Schema。编解码按 `tfplugin5.Schema` 产出 `tfplugin5.DynamicValue`，执行器直接放进请求，响应里的 `DynamicValue` 直接交回解码。Diagnostics 和 `RequiresReplace` 直接返回生成的 `tfplugin5` 结构。`ProviderMeta`、identity、`ClientCapabilities` 和 deferred 留空，不参与分支。

调用方看到的是适配后的 Schema、对象形式的配置和状态，以及 Diagnostics。空的配置或状态保持空值，不编码成空对象，这样创建和删除仍由「是否为空」区分。

备选是让公开方法直接使用生成的 Request。那样调用方要理解 DynamicValue 和 proto 字段，和第 6.2 节的对象入参不一致。

### 3. 编解码通过接口注入，本变更提供假实现仅供测试

在 `tfruntime` 定义执行器依赖的编解码接口：按一份 `tfplugin5.Schema` 把调用方对象编码成 `DynamicValue`，以及按同一份 Schema 解码。`ProcessManager` 增加设置该接口的方法，不改 `NewProcessManager` 的现有参数。未设置时，需要编码或解码的方法返回错误且不发 RPC。`GetProviderSchema` 只做 Schema 形状转换，不调用该接口。`UpgradeResourceState` 的 `RawState` 也不调用它。

真实的 msgpack 构造由编解码子需求实现，再注入这个接口。本变更的测试使用计数用的假编解码器，用来证明拒绝入参时 RPC 次数为 0，以及成功路径会把编解码结果放进协议字段。

备选是在本变更里直接依赖一个尚未存在的编解码包。那样执行器无法单独编译，也会把 msgpack 写进错误的需求。

### 4. 执行器只记住是否已成功 Configure

执行器保存这一项状态，并用已有的互斥锁保护：

| 状态 | 作用 |
|---|---|
| 是否已成功 Configure | 为假时，资源和数据源方法不发 RPC |

一次调用的 Error 只使该次调用失败，不拒绝之后另一次计划或读取。升级结果只放进该次返回值，计划的 `PriorState` 使用调用方传入的值。`GetProviderSchema` 把生成的 Schema 留在执行器里，供后续按类型名取出编码用的 Schema。`ConfigureProvider` 发现缓存为空时自己加载并缓存，不要求调用方先取 Schema。类型名不在对应映射中时，方法在编码前失败。

计划的操作类别只看空值：都空则不发 RPC；只有新状态是创建；都有值是更新；只有旧状态是删除，且返回的 `PlannedState` 为空。计划方法不调用 Apply。删除应用把返回的 `NewState` 置空。Apply 含 Error 时仍然把解码后的 `NewState` 放进结果，同时让调用失败。

### 5. Schema 指纹对适配结果做规范 JSON 的 SHA-256

`schema_hash` 是适配后文档的十六进制 SHA-256，不加 `sha256:` 前缀。参与计算的内容是 provider、资源 Schema 和数据源 Schema，键按字典序，包含每份 Schema 的协议 `version`。不包含 `schema_hash` 自身，也不包含 `diagnostics`、插件名和插件版本。因此 Schema 正文不变时指纹不变，任一字段变化时指纹变化。有资源 Schema 时 `capabilities` 包含 `apply`、`plan`、`validate`、`import`、`refresh`；有数据源 Schema 时包含 `read_data`。能力位不单独参与指纹。

属性 `type` 使用协议里 TF JSON 类型字节解码后的结构，原样放进适配结果。NestedBlock 以块名为键。协议没有的可选描述不补空字符串以外的默认文案。

### 6. Warning 写 summary，凭证不进环境变量

Diagnostics 里只要有 Error，该调用就失败。只有 Warning 时调用成功，每条 summary 用现有日志写入，并带上插件名。不把 Config 或状态正文写入这条日志。`TerraformVersion` 使用 `pkg/version` 的 `VERSION`，只写入 `ConfigureProvider` 的协议字段。

插件版本沿用进程登记。登记为空时，Schema 里的 `plugin.version` 为空；日志行继续使用现有的 `unknown`，两处不相混。本变更不改日志文件规则。

## Risks / Trade-offs

- [Risk] 公开方法签名从生成的 Provider 客户端改成第 6.2 节对象，已有编译期断言会去掉。→ Mitigation：仓库内没有外部调用方。`Start` 返回 `executor.Executor`，`Close` 顺序不变。
- [Risk] 编解码尚未合入，配置和状态方法在未注入时不能真正调用插件。→ Mitigation：接口和失败行为先固定；测试用假实现覆盖「拒绝则 RPC 次数为 0」。真实编码留在依赖需求。
- [Risk] 生成请求漏填或误填协议 5.9 字段。→ Mitigation：本期只填写第 6.2 节用到的字段；identity、functions、deferred 明确留空。
- [Risk] 删除计划强制 `PlannedState` 为空，可能盖住插件返回的状态。→ Mitigation：按需求，删除计划的返回状态就是空。私有字节仍原样返回。
- [Risk] Schema 指纹算法一旦发布就要保持稳定。→ Mitigation：本变更把算法定成规范 JSON 的 SHA-256，并写进测试：相同正文同指纹，改一个属性则指纹变化。

## Migration Plan

1. 桩代码和 `go-plugin` 版本不动，不引入 `terraform-plugin-go`。
2. 给 `PluginExecutor` 补上协议适配、第 6.2 节方法和门闩。此时没有北向入口，不调用 `Start` 的部署行为不变。
3. 编解码需求合入后，由进程池在 `Start` 前注入编解码实现。未注入时相关方法失败，不发 RPC。
4. 回滚时移除执行器方法并恢复未实现的 RPC 即可。没有表结构、配置项和北向 API。进程拉起与日志不受影响。

## Open Questions

无。协议小版本沿用已落地的 5.9；指纹算法、编解码接口和调用顺序状态已在上面定下。
