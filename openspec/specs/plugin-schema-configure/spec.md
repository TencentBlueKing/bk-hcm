# plugin-schema-configure Specification

## Purpose
定义 Terraform 插件执行器读取并适配统一 Provider Schema（含 `schema_hash` 与 `capabilities` 推导）、按 `PrepareProviderConfig` → `ConfigureProvider` 顺序写入配置及配置完成前的门控规范。

## Requirements

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
