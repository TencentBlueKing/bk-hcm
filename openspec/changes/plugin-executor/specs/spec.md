# plugin-executor-client

## ADDED Requirements

### Requirement: 执行器只绑定协议主版本 5 的连接
系统 MUST 只在 `NegotiatedVersion()` 为 5 的连接上构造可使用的执行器。协商结果不是 5 时，系统 MUST 结束该进程并返回错误，调用方 MUST NOT 拿到执行器。本期 MUST 只使用已提交的协议 5.9 桩，MUST NOT 另生成一套协议 5.3 客户端。

#### Scenario: 协商结果为主版本 5
- **WHEN** 进程管理完成握手且 `NegotiatedVersion()` 为 5
- **THEN** 调用方拿到绑定该连接的执行器

#### Scenario: 协商结果不是 5
- **WHEN** 插件完成握手但 `NegotiatedVersion()` 不是 5
- **THEN** 系统结束该进程并返回错误，不返回执行器

### Requirement: 执行器直接填写生成的协议请求
`GRPCClient` MUST 接收已建立的 `*grpc.ClientConn`，并持有 `tfplugin5` 生成的 Provider 客户端。执行器 MUST 直接填写生成的 Request。`ProviderMeta`、identity、`ClientCapabilities` 和 deferred MUST 留空。响应里的 `DynamicValue` MUST 直接交给编解码。对外结果仍要使用的 Schema MUST 从生成的 Response 转换。`RequiresReplace` 与 Diagnostics MUST 直接返回生成的 `tfplugin5` 结构。`GRPCServer` MUST 返回错误。hc-service MUST NOT 作为该协议的 gRPC 服务端。

#### Scenario: 建立客户端
- **WHEN** `go-plugin` 以已连接的 `*grpc.ClientConn` 调用 `GRPCClient`
- **THEN** 返回的执行器可以通过生成的客户端发送 Provider 请求并解析响应

#### Scenario: 拒绝作为服务端
- **WHEN** 调用该 `GRPCPlugin` 的 `GRPCServer`
- **THEN** 方法返回错误，不注册 Provider 服务

### Requirement: 关闭开始后不再向插件发出业务 RPC
每个会向插件发 RPC 的执行器方法 MUST 在方法内部做前置和后置门闩。前置检查发现 `Close` 已开始时，方法 MUST 失败，且插件侧该次 RPC 次数为 0。后置在方法返回时离开门闩。系统 MUST NOT 要求调用方经由单独的 `Call` 才能访问这些方法。`StopProvider` 由 `Close` 在门闩关闭后调用，MUST NOT 再进入这道门闩。

#### Scenario: 关闭之后调用读取
- **WHEN** `Close` 已经开始，调用方再调用 `ReadResource`
- **THEN** 调用失败，插件侧 `ReadResource` 次数为 0

#### Scenario: 关闭时存在在途方法
- **WHEN** 一个执行器方法尚未返回时调用方调用 `Close`
- **THEN** 系统不再接受新的方法进入，等待该在途方法返回后才调用 `StopProvider`

### Requirement: 编解码拒绝入参时不发 RPC
执行器 MUST 在把配置或状态发给插件之前调用编解码。编解码返回错误时，执行器 MUST 使该次调用失败，且插件侧该次 RPC 次数为 0。执行器 MUST NOT 在本变更中实现 `DynamicValue` 的类型构造、msgpack 编码或按 Schema 解码。

#### Scenario: 计划入参被编解码拒绝
- **WHEN** 编解码拒绝 `PlanResourceChange` 的入参
- **THEN** 该次计划失败，插件侧 RPC 次数为 0

### Requirement: 不传 ProviderMeta，Private 按字节原样往返
执行器发出的协议请求 MUST NOT 设置 `ProviderMeta`。`Private`、`PriorPrivate` 和 `PlannedPrivate` MUST 按字节原样传入和返回。执行器 MUST NOT 解码这些字节，也 MUST NOT 把它们持久化到执行器之外。

#### Scenario: 读取时带上上次的 Private
- **WHEN** 调用方以一组私有字节调用 `ReadResource`，且插件在响应中返回另一组私有字节
- **THEN** 插件收到的字节与调用方传入的相同，调用方收到的字节与插件返回的相同，执行器不保存这两段字节

### Requirement: 本变更不改既有资源路径
本变更的代码差异 MUST NOT 包含既有资源 Handler 目录和 `pkg/adaptor` 下的文件。系统 MUST NOT 新增北向 HTTP API，也 MUST NOT 实现进程池调度。

#### Scenario: 统计既有目录的变更
- **WHEN** 统计本变更里既有资源 Handler 与 `pkg/adaptor` 的变更文件
- **THEN** 这些目录的变更文件数为 0

# plugin-schema-configure

## ADDED Requirements

### Requirement: GetProviderSchema 返回适配后的统一 Schema
`GetProviderSchema` 的请求 MUST 没有业务字段。响应 MUST 包含 `protocol_version`、`plugin.name`、`plugin.version`、`schema_hash`、`capabilities`、`provider`、`resource_schemas`、`data_source_schemas` 和 `diagnostics`。`protocol_version` MUST 固定为 `"1.0"`。`plugin.name` 与 `plugin.version` MUST 来自进程登记；版本未登记时 `plugin.version` MUST 为空字符串。系统 MUST NOT 把日志里的 `unknown` 填进该字段。`capabilities` MUST 由执行器按本要求填写，MUST NOT 从协议响应的 `ServerCapabilities` 拷贝。

协议中的 Provider、资源、数据源 Schema MUST 适配为统一结构：顶层是协议里的 `version` 和 `block`；`block.attributes` 是属性名到 Attribute 的映射；`block.block_types` 是块名到 NestedBlock 的映射，没有时为空对象。Attribute 的 `type` MUST 是协议类型字节对应的 TF JSON 类型。`required`、`optional`、`computed` MUST 保留协议中的布尔值。`sensitive` 与 `description` 在协议提供时保留。`schema_hash` MUST 是这段适配结果的十六进制 SHA-256，MUST NOT 加 `sha256:` 前缀。

#### Scenario: 读取 Schema
- **WHEN** 调用方调用 `GetProviderSchema`，且进程登记了插件名、未登记插件版本
- **THEN** 响应的 `protocol_version` 为 `"1.0"`，`plugin.name` 为登记的插件名，`plugin.version` 为空，资源与数据源 Schema 使用适配后的 Attribute 与 NestedBlock

#### Scenario: Schema 内容不变时指纹相同
- **WHEN** 两次 `GetProviderSchema` 得到的适配后 Schema 内容相同
- **THEN** 两次响应的 `schema_hash` 相同

#### Scenario: Schema 内容变化时指纹变化
- **WHEN** 适配后的资源或数据源 Schema 与上一次相比有字段变化
- **THEN** `schema_hash` 与上一次不同

### Requirement: capabilities 只声明本期实现的能力
`resource_schemas` 至少有一个条目时，`capabilities` MUST 包含值为 true 的 `apply`、`plan`、`validate`、`import`、`refresh`。`resource_schemas` 没有条目时，`capabilities` MUST NOT 包含这些键。`data_source_schemas` 至少有一个条目时，`capabilities` MUST 包含值为 true 的 `read_data`。`data_source_schemas` 没有条目时，`capabilities` MUST NOT 包含 `read_data`。`capabilities` MUST NOT 包含 `async_apply` 和 `list`。未声明的能力 MUST 是键不存在，而不是值为 false。

#### Scenario: 只有资源 Schema
- **WHEN** 资源 Schema 非空且数据源 Schema 为空时 `GetProviderSchema` 返回
- **THEN** `capabilities` 中 `apply`、`plan`、`validate`、`import`、`refresh` 均为 true，且不包含 `read_data`、`async_apply`、`list`

#### Scenario: 只有数据源 Schema
- **WHEN** 资源 Schema 为空且数据源 Schema 非空时 `GetProviderSchema` 返回
- **THEN** `capabilities` 不包含 `apply`、`plan`、`validate`、`import`、`refresh`，且 `read_data` 为 true

### Requirement: ConfigureProvider 先 Prepare 再 Configure
调用方 MUST 只调用 `ConfigureProvider`。执行器尚未缓存 Schema 时，MUST 先调用 `GetProviderSchema` 对应的协议请求并缓存生成的 Schema，再把 Config 交给编解码，然后调用 `PrepareProviderConfig`。Schema 请求失败时 MUST NOT 调用 Prepare 或 Configure。该响应的 Diagnostics 不含 Error 时，执行器 MUST 用 `PreparedConfig` 调用 `ConfigureProvider`。`PrepareProviderConfig` 的响应 MUST NOT 返回给调用方。`TerraformVersion` MUST 填 hc-service 的版本字符串，执行器 MUST NOT 根据该字符串选择不同调用分支。凭证 MUST 只通过 Config 到达插件，执行器 MUST NOT 把凭证写入进程环境变量。

#### Scenario: Prepare 含 Error
- **WHEN** `PrepareProviderConfig` 的 Diagnostics 含 severity 为 Error
- **THEN** 执行器不再调用 `ConfigureProvider`，插件侧 `ConfigureProvider` 次数为 0，该次配置失败

#### Scenario: 凭证经 Config 到达插件
- **WHEN** 调用方 Config 的 `secret_id` 为 `ak-test-value`，且 Prepare 与 Configure 都不含 Error
- **THEN** 插件侧 `ConfigureProvider` 收到的 Config 中 `secret_id` 为 `ak-test-value`，插件环境变量中没有该值

### Requirement: Configure 成功之前不调用资源和数据源方法
在 `ConfigureProvider` 返回且 Diagnostics 不含 Error 之前，执行器 MUST NOT 向插件发出资源方法或数据源方法。这些方法包括校验、升级、读取、计划、应用、导入和数据源读取。`ConfigureProvider` 成功之后，参数合法的资源方法 MUST 可以发往插件。

#### Scenario: 尚未配置成功就读取
- **WHEN** 进程已拉起，且 `ConfigureProvider` 尚未返回不含 Error 的结果，调用方调用 `ReadResource`
- **THEN** 插件侧 `ReadResource` 次数为 0

#### Scenario: 配置成功后读取
- **WHEN** `ConfigureProvider` 已返回且 Diagnostics 不含 Error，调用方发起参数合法的 `ReadResource`
- **THEN** 插件侧 `ReadResource` 次数为 1

# plugin-resource-data-source

## ADDED Requirements

### Requirement: Diagnostics 的 Error 使调用失败，Warning 只记日志
响应 Diagnostics 中任一条目的 severity 为 Error 时，该调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次独立调用。severity 只有 Warning 时，调用结果 MUST NOT 标记为失败，执行器 MUST 把每条 Warning 的 summary 写入日志。空的属性路径表示问题不指向单个字段。

#### Scenario: 响应只有 Warning
- **WHEN** 插件响应的 Diagnostics 只含 severity 为 Warning
- **THEN** 调用结果不标记为失败，且该 Warning 的 summary 写入日志

#### Scenario: 响应含 Error
- **WHEN** 插件响应的 Diagnostics 含 severity 为 Error
- **THEN** 该调用失败，调用方仍能看到这些 Diagnostics

### Requirement: 类型名必须是对应 Schema 的 key
资源方法的 `TypeName` MUST 是 `resource_schemas` 的 key。数据源方法的 `TypeName` MUST 是 `data_source_schemas` 的 key。不满足时执行器 MUST 使该次调用失败，且插件侧该次 RPC 次数为 0。

#### Scenario: 资源类型名不在 Schema 中
- **WHEN** 调用方对 `ReadResource` 传入的类型名不是资源 Schema 的 key
- **THEN** 调用失败，插件侧 `ReadResource` 次数为 0

### Requirement: 资源校验或升级失败不拦住后续计划
`ValidateResourceTypeConfig` 或 `UpgradeResourceState` 的 Diagnostics 含 Error 时，该次调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次 `PlanResourceChange`。升级成功时，解码后的状态 MUST 只出现在该次返回值里。执行器 MUST NOT 用它替换之后 `PlanResourceChange` 的 `PriorState`。请求中的升级前状态是 `RawState`，`JSON` 与 `Flatmap` 只填其中之一，执行器 MUST NOT 把它当作 `DynamicValue` 编码。

#### Scenario: 校验失败后请求计划
- **WHEN** `ValidateResourceTypeConfig` 的 Diagnostics 含 Error，调用方继续请求该类型的计划
- **THEN** 插件侧 `PlanResourceChange` 次数为 1

#### Scenario: 升级失败后请求计划
- **WHEN** 升级响应的 Diagnostics 含 Error，调用方请求该类型的计划
- **THEN** 插件侧 `PlanResourceChange` 次数为 1，且 `PriorState` 仍是调用方传入的状态

#### Scenario: 升级成功后计划仍使用调用方状态
- **WHEN** 升级响应不含 Error，调用方随后以自己的状态请求该类型的计划
- **THEN** 发往插件的 `PriorState` 是调用方传入的状态，升级结果只在升级返回值中

### Requirement: 计划只按空状态区分创建、更新和删除
`PlanResourceChange` MUST 只发出计划 RPC，MUST NOT 发出 `ApplyResourceChange`。`PriorState` 与 `ProposedNewState` 都为空时，执行器 MUST NOT 发出计划 RPC。`PriorState` 为空且 `ProposedNewState` 有值时，计划 MUST 按创建处理。两者都有值时，计划 MUST 按更新处理。`PriorState` 有值且 `ProposedNewState` 为空时，计划 MUST 按删除处理，返回的 `PlannedState` MUST 为空。`RequiresReplace` MUST 原样返回生成的属性路径，执行器 MUST NOT 把它改写成另一套结构，也 MUST NOT 在本次调用中删除或创建资源。空的状态 MUST 以空值交给插件，MUST NOT 编码成空对象。

#### Scenario: 计划不修改外部系统
- **WHEN** `PriorState` 有值且 `ProposedNewState` 有值，调用方只调用 `PlanResourceChange`
- **THEN** 插件侧只有计划 RPC，没有 `ApplyResourceChange`

#### Scenario: 新旧状态都为空
- **WHEN** `PriorState` 为空且 `ProposedNewState` 为空，调用方调用 `PlanResourceChange`
- **THEN** 插件侧计划 RPC 次数为 0

#### Scenario: 创建计划
- **WHEN** `PriorState` 为空且 `ProposedNewState` 有值，调用方调用 `PlanResourceChange`
- **THEN** 计划按创建处理

#### Scenario: 删除计划
- **WHEN** `PriorState` 有值且 `ProposedNewState` 为空，调用方调用 `PlanResourceChange`
- **THEN** 计划按删除处理，且返回的 `PlannedState` 为空

#### Scenario: RequiresReplace 只返回路径
- **WHEN** 计划响应的 `RequiresReplace` 包含字段 `name`
- **THEN** 返回路径包含 `name`，且本次调用的 `ApplyResourceChange` 次数为 0

### Requirement: 应用把插件返回的状态交给调用方
删除计划完成后，`ApplyResourceChange` 返回的 `NewState` MUST 为空。插件 Apply 返回的 `NewState` 非空且 Diagnostics 含 Error 时，调用 MUST 失败，调用方收到的 `NewState` MUST 与插件返回的字段一致。执行器 MUST NOT 在 Apply 失败时丢弃已返回的 `NewState`。

#### Scenario: 删除应用的状态为空
- **WHEN** 删除计划已完成且 `ApplyResourceChange` 返回
- **THEN** 调用方收到的 `NewState` 为空

#### Scenario: Apply 失败仍返回 NewState
- **WHEN** 插件 Apply 返回的 `NewState` 非空，且 Diagnostics 含 Error
- **THEN** 调用方收到失败，且收到的 `NewState` 与插件返回字段一致

### Requirement: 导入不创建资源也不再读取
`ImportResourceState` MUST NOT 在本方法期间调用 `ReadResource`，也 MUST NOT 创建外部资源。每一项返回资源 MUST 按其自身类型名对应的 Schema 解码。该类型名不在资源 Schema 中时，该次调用 MUST 失败，执行器 MUST NOT 改用请求类型的 Schema。外部标识指向已存在资源时，返回项 MUST 包含 `State`。外部标识在插件侧不存在时，Diagnostics MUST 含 Error，且外部系统新增资源数为 0。

#### Scenario: 导入已存在的资源
- **WHEN** 外部标识指向已存在资源并调用 `ImportResourceState`
- **THEN** 返回项包含 `State`，本方法期间 `ReadResource` 次数为 0，外部系统新增资源数为 0

#### Scenario: 返回类型不在 Schema 中
- **WHEN** 导入响应中某项的类型名不是资源 Schema 的 key
- **THEN** 该次调用失败，且该项状态没有按请求类型的 Schema 解码

#### Scenario: 导入不存在的标识
- **WHEN** 外部标识在插件侧不存在并调用 `ImportResourceState`
- **THEN** Diagnostics 含 Error，且外部系统新增资源数为 0

### Requirement: 数据源只经过校验和读取
`ValidateDataSourceConfig` 的 Diagnostics 含 Error 时，该次调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次 `ReadDataSource`。`ReadDataSource` MUST 返回服从该数据源 Schema 的字段，且 MUST NOT 调用 `PlanResourceChange` 或 `ApplyResourceChange`。数据源路径 MUST NOT 经过计划和应用。

#### Scenario: 数据源校验失败后读取
- **WHEN** 数据源校验的 Diagnostics 含 Error，调用方请求读取
- **THEN** 插件侧 `ReadDataSource` 次数为 1

#### Scenario: 校验通过后读取数据源
- **WHEN** 查询条件通过校验并调用 `ReadDataSource`
- **THEN** 返回字段都存在于该数据源 Schema，且 `PlanResourceChange` 与 `ApplyResourceChange` 次数为 0

### Requirement: 不实现本期范围外的协议方法
执行器 MUST NOT 调用或向调用方暴露 `GetMetadata`、资源 identity、`MoveResourceState`、ephemeral 资源、`GetFunctions`、`CallFunction`、`async_apply` 和 `list`。

#### Scenario: 调用方只使用本期方法
- **WHEN** 调用方完成 Schema、配置、资源计划和数据源读取
- **THEN** 插件侧不出现 `GetMetadata`、`MoveResourceState`、ephemeral、function 或 list 的 RPC
