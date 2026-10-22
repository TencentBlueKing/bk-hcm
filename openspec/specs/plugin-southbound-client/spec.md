# plugin-southbound-client Specification

## Purpose
定义 HCM 插件南向协议的 Schema/Capabilities 校验模型、诊断结构（`Diagnostic` / `PathStep`）以及经蓝鲸 API 网关调用全部 11 个南向 HTTP 接口（`Health`、`GetSchema`、`Apply`、`GetResult`、`Stop`、`Plan`、`Validate`、`Import`、`Refresh`、`ReadData`、`List`）的客户端规范。

## Requirements
### Requirement: 南向协议 Schema 结构与元数据校验
系统 SHALL 支持解析南向协议 `GET /schema` 返回的 `PluginSchema`（内嵌 `github.com/hashicorp/terraform-json` 的 `tfjson.ProviderSchema` 并扩展 HCM 顶层字段），并按协议规则校验顶层元数据、顶层 `block` 非空及类型名前缀约束：
- 顶层 `protocol_version`、`plugin.name`、`plugin.version`、`schema_hash`、`provider.block` 必须非空；
- `resource_schemas` 与 `data_source_schemas` 中所有类型名（`type_name`）必须以 `plugin.name + "_"` 为前缀，且各自的 `block` 必须非空（内部属性类型由 `tfjson` / `cty.Type` 在 JSON 反序列化时自动校验）。

#### Scenario: 满足约束的完整 PluginSchema 校验通过
- **WHEN** `PluginSchema` 的 `protocol_version`、`plugin.name`（如 `"bkcmdb"`）、`plugin.version`（如 `"1.2.0"`）、`schema_hash` 非空，`provider`、资源 `bkcmdb_host` 与数据源 `bkcmdb_biz` 的 `block` 非空且类型名满足前缀规则，同时 `capabilities` 满足对应约束
- **THEN** `PluginSchema.Validate()` 返回 `nil`

#### Scenario: 顶层元数据或 block 缺失、类型名前缀不匹配时校验失败
- **WHEN** `protocol_version`、`plugin.name`、`plugin.version`、`schema_hash` 或 `provider.block` 为空，或 `resource_schemas` / `data_source_schemas` 中存在未以 `plugin.name + "_"` 开头的 key（例如插件名为 `"bkcmdb"` 但资源名为 `"other_host"`）
- **THEN** `PluginSchema.Validate()` 返回具体的元数据缺失或类型名前缀错误

### Requirement: 南向协议 Capabilities 声明约束校验
系统 SHALL 校验 `PluginSchema.Capabilities`（`apply`、`plan`、`refresh`、`read_data`、`async_apply`、`stop`、`validate`、`import`、`list` 共 9 个能力位）与 `resource_schemas`、`data_source_schemas` 的一致性：
- 当 `resource_schemas` 非空时，`capabilities.apply`、`capabilities.plan`、`capabilities.refresh` 必须全部为 `true`；
- 当 `resource_schemas` 为空时，所有资源类能力位（`apply`、`plan`、`refresh`、`async_apply`、`stop`、`import`、`list`）均不得为 `true`；
- 当 `data_source_schemas` 非空时，`capabilities.read_data` 必须为 `true`；当 `data_source_schemas` 为空时，`capabilities.read_data` 不得为 `true`；
- 当 `capabilities.stop` 为 `true` 时，`capabilities.async_apply` 必须同时为 `true`。

#### Scenario: 非空 resource_schemas 未声明必需能力位或为空时误声明资源能力位
- **WHEN** `resource_schemas` 非空但 `apply`、`plan`、`refresh` 中任意一项为 `false`；或 `resource_schemas` 为空但声明了 `apply`/`plan`/`refresh`/`async_apply`/`stop`/`import`/`list` 中任意一项为 `true`
- **THEN** `PluginSchema.Validate()` 返回能力位约束错误

#### Scenario: data_source_schemas 与 read_data 能力位不匹配或 stop 未伴随 async_apply
- **WHEN** `data_source_schemas` 非空但 `read_data` 为 `false`，或 `data_source_schemas` 为空但 `read_data` 为 `true`，或 `stop` 为 `true` 但 `async_apply` 为 `false`
- **THEN** `PluginSchema.Validate()` 返回能力位约束错误

### Requirement: Diagnostic 与 PathStep 结构校验与通过判定
系统 SHALL 支持对齐 Terraform `AttributePath.Step` 的 `PathStep` 与 `Diagnostic` 结构校验，以及 `Validate` 响应的通过性判定：
- 每个 `PathStep` 必须且只能设置 `attribute`（非空字符串，表示对象属性或嵌套块名）、`key`（字符串，表示 map 的键）、`index`（`>= 0` 的整数，表示 list 的下标）三者之一；
- 每个 `Diagnostic` 的 `message` 必须非空，`severity` 允许为 `"error"`、`"warning"` 或省略（省略时等效为 `"error"`），`path` 中所有 `PathStep` 必须合法（空数组 `[]` 表示问题与单一字段无关）；
- `ValidateData.IsPassed()` 在 `diagnostics` 为空或仅包含 `severity: "warning"` 时返回 `true`，包含任意有效级别为 `"error"` 的元素时返回 `false`。

#### Scenario: PathStep 严格三选一校验
- **WHEN** `PathStep` 仅设置了非空 `attribute`、任意 `key` 或 `>= 0` 的 `index` 之一
- **THEN** `PathStep.Validate()` 返回 `nil`；若同时设置多个字段、三个字段均未设置、`attribute` 为空白字符串或 `index < 0`，则返回校验错误

#### Scenario: Diagnostics 校验通过与不通过判定
- **WHEN** `ValidateData.Diagnostics` 为空数组，或仅包含 `severity: "warning"` 的元素
- **THEN** `ValidateData.IsPassed()` 返回 `true`；若包含任意 `severity: "error"` 或未填 `severity`（默认 `"error"`）的元素，则返回 `false`

### Requirement: 插件网关客户端初始化与统一响应处理
系统 SHALL 提供 `plugin.NewClient(baseURL, cfg, reg)` 构造南向 HTTP 客户端，将插件在蓝鲸网关上的环境地址 `baseURL`（如 `https://bkapi.example.com/api/bk-hcm-plugin-cmdb/prod`）拆分为 `Scheme://Host` 与网关路径前缀 `basePath`，并复用 `apigateway.ApiGatewayCall` 与 `apigateway.ApiGatewayCallWithoutReq` 处理统一响应 `{ "result": bool, "code": int, "message": string, "data": ... }`。

#### Scenario: 合法网关 baseURL 解析与路由拼接
- **WHEN** 传入包含协议、主机与网关环境路径的 `baseURL` 及非空 `*cc.ApiGateway` 配置
- **THEN** `NewClient` 初始化成功，后续调用接口时向 `{baseURL}/{endpoint}` 发起携带 `X-Bkapi-Authorization` 应用认证头的 HTTP 请求；当 `baseURL` 为空、缺少 `http`/`https` 协议或缺少主机名、或 `cfg` 为 `nil` 时返回错误

#### Scenario: 统一响应外层错误拦截与业务状态透传
- **WHEN** 插件返回 `result: false` 或 `code != 0`（如 `50300` 外部服务不可用、`40400` 资源不存在）
- **THEN** 客户端返回包含 `code` 与 `message` 的 `error`；当插件返回 `result: true, code: 0` 时，客户端将 `data` 反序列化后原样返回给调用方（即使 `ApplyData.Status == "failed"` 或 `ValidateData.Diagnostics` 非空）

### Requirement: Health 与 GetSchema 探活及元数据拉取接口
系统 SHALL 提供 `Health(kt)`（`GET /health`）与 `GetSchema(kt)`（`GET /schema`）客户端方法，用于定期探活、检查 `plugin_version` 变更及拉取插件完整 `PluginSchema`。

#### Scenario: 调用 Health 与 GetSchema 成功
- **WHEN** 调用 `Health(kt)` 或 `GetSchema(kt)` 且插件返回 `result: true, code: 0`
- **THEN** 客户端分别向 `{base_url}/health` 或 `{base_url}/schema` 发起无请求体的 `GET` 请求，并返回解析后的 `*HealthData`（含 `plugin_version`）或 `*PluginSchema`

### Requirement: Apply、GetResult 与 Stop 资源执行及异步生命周期接口
系统 SHALL 提供 `Apply(kt, req)`（`POST /apply`）、`GetResult(kt, req)`（`POST /result`）与 `Stop(kt, req)`（`POST /stop`）客户端方法：
- `ApplyReq` 必须包含非空 `type_name` 与 `launch_config`，且 `prior_state` 与 `proposed_state` 不得同时为 `nil`（`prior_state == nil` 表示创建，`proposed_state == nil` 表示删除，两者均非 `nil` 表示更新）；
- `GetResultReq` 与 `StopReq` 必须包含非空 `type_name`、`launch_config` 与 `apply_id`；
- 三个接口均返回 `*ApplyData`（包含 `status`、`message`、`apply_id`、`state`、`progress`），支持同步成功（`succeeded`）、异步执行中（`running`，携带 `apply_id` 与可选检查点 `state`/`progress`）以及任务失败（`failed`，携带失败原因 `message` 与可选的部分落地实际 `state`）。

#### Scenario: Apply 同步执行创建、更新、删除与异步转 running
- **WHEN** 调用 `Apply` 传入合法的 `type_name`、`launch_config` 及 `prior_state`/`proposed_state` 组合
- **THEN** 客户端向 `{base_url}/apply` 发起 `POST` 请求并返回 `*ApplyData`；若 `prior_state` 与 `proposed_state` 同时为 `nil`，客户端在发起 HTTP 请求前直接返回参数校验错误

#### Scenario: Apply、GetResult 或 Stop 返回任务失败（status=failed）及部分落地状态
- **WHEN** 插件执行任务异常中断或撤单成功，返回 `result: true, code: 0` 且 `data.status == "failed"`
- **THEN** 客户端不视为 HTTP 调用失败，正常返回 `*ApplyData`，调用方从 `ApplyData.Message` 获取失败原因，并在外部系统已发生部分变更时从 `ApplyData.State` 获取实际已落地状态

### Requirement: Plan 与 Validate 变更预览及配置校验接口
系统 SHALL 提供 `Plan(kt, req)`（`POST /plan`）与 `Validate(kt, req)`（`POST /validate`）客户端方法：
- `PlanReq` 校验规则与 `ApplyReq` 一致（`type_name`、`launch_config` 非空，`prior_state` 与 `proposed_state` 不同时为 `nil`），返回 `*PlanData`（包含预计执行后的 `state` 与需先删后建的字段名列表 `replace_paths`）；
- `ValidateReq` 必须包含非空 `type_name`、`launch_config` 及非 `nil` 的 `config`，返回 `*ValidateData`（包含 `diagnostics` 列表）。

#### Scenario: Plan 返回预计状态与需要先删后建的 replace_paths
- **WHEN** 调用 `Plan` 预览资源变更（如修改不可原地更新的字段）且插件返回 `result: true, code: 0`
- **THEN** 客户端向 `{base_url}/plan` 发起 `POST` 请求并返回 `*PlanData`，其中包含预计 `State`（尚无法计算的 computed 字段为 `null`）与 `ReplacePaths`（如 `["bk_host_innerip"]`）

#### Scenario: Validate 返回业务规则校验诊断列表
- **WHEN** 调用 `Validate` 传入用户配置 `config` 且插件返回 `result: true, code: 0`
- **THEN** 客户端向 `{base_url}/validate` 发起 `POST` 请求并返回 `*ValidateData`，调用方可通过 `ValidateData.Diagnostics` 获取全部字段级或跨字段问题，并通过 `ValidateData.IsPassed()` 判定是否通过

### Requirement: Import、Refresh、ReadData 与 List 状态读取及资源枚举接口
系统 SHALL 提供 `Import(kt, req)`（`POST /import`）、`Refresh(kt, req)`（`POST /refresh`）、`ReadData(kt, req)`（`POST /read_data`）与 `List(kt, req)`（`POST /list`）客户端方法：
- `ImportReq` 必须包含非空 `type_name`、`launch_config` 与 `import_id`，返回 `*StateData`；
- `RefreshReq` 必须包含非空 `type_name`、`launch_config` 与非空 `state`，返回 `*StateData`（当外部系统中该对象已不存在时，`result: true, code: 0` 且 `StateData.State` 为 `nil`）；
- `ReadDataReq` 必须包含非空 `type_name`、`launch_config` 与非 `nil` 的 `config`，返回 `*StateData`；
- `ListReq` 必须包含非空 `type_name`、`launch_config`，且 `limit` 必须在 `1~500` 闭区间内，支持可选 `config` 过滤条件与 `page_token` 游标（首页传 `""`），返回 `*ListData`（包含当前页 `items[].import_id`、`items[].state` 及下一页游标 `next_page_token`）。

#### Scenario: Import 纳管、Refresh 漂移刷新（含外部对象已删除）与 ReadData 数据源查询
- **WHEN** 分别调用 `Import`、`Refresh` 或 `ReadData` 且请求参数合法
- **THEN** 客户端向对应路径发起 `POST` 请求并返回 `*StateData`；其中 `Refresh` 在外部对象已被删除时返回 `StateData{State: nil}` 且不报错

#### Scenario: List 分页拉取资源列表与 limit 边界校验
- **WHEN** 调用 `List` 且 `limit` 在 `1~500` 范围内
- **THEN** 客户端向 `{base_url}/list` 发起 `POST` 请求并返回 `*ListData`（含 `Items` 与 `NextPageToken`）；当 `limit < 1` 或 `limit > 500` 时，客户端在发起 HTTP 请求前直接返回参数校验错误
