# Spec Delta

## MODIFIED Requirements

### Requirement: 插件最高版本启用态校验（规则 R-001）
系统 SHALL 在 `router.Service` 处理北向请求（除指定了 `plugin_version` 的 `GetSchema` 外）时，查询 `data-service` 的 `iac_plugin` 表中 `plugin_name = req.plugin` 的记录并校验启用状态：
- 若该插件为本副本本地 `tf_provider` 插件（命中本副本启动扫描锁定的 `discovery.Entry`）或数据库中插件记录为 `protocol == "tf_provider"`，系统 MUST 按本副本锁定的 `discovery.Entry.PluginVersion`（`localVersion`）在记录中匹配 `(plugin_name, localVersion)`（若本副本无该插件二进制或数据库中对应 `localVersion` 记录不存在/不处于 `state == "enabled"`，系统 MUST 返回 `errf.Aborted` 或 `errf.RecordNotFound`），确保滚动升级新旧副本并存期间始终使用本机 `localVersion`；
- 若该插件为三方 `http` 插件，系统通过 `iacplugin.Latest(records)` 选取最高版本记录；若无任何注册记录 MUST 返回 `errf.RecordNotFound`（`code=2000003`），若最高版本记录的 `state != "enabled"` MUST 返回 `errf.Aborted`（`code=2000006`）。

#### Scenario: 插件最高版本处于 enabled 状态时校验通过
- **WHEN** `iac_plugin` 表中插件 `bkcmdb` 存在版本 `1.0.0`（`disabled`）与 `1.2.0`（`enabled`）两条记录，且调用北向需实例接口或未指定版本的 `GetSchema`
- **THEN** `router.Service` 通过 `iacplugin.Latest` 选中最高版本 `1.2.0`，校验其 `state == "enabled"` 通过并继续后续处理

#### Scenario: 插件未注册或最高版本非 enabled 状态时前置拦截
- **WHEN** 请求的 `plugin` 在 `iac_plugin` 表中不存在，或三方插件最高版本 / 本地插件本副本 `localVersion` 记录处于 `disabled` / `failed` / `init` 状态
- **THEN** `router.Service` 分别返回 `errf.RecordNotFound`（`2000003`）或 `errf.Aborted`（`2000006`），且向南向网关或进程池 `Acquire` 发起的调用次数为 0

#### Scenario: 本地 tf_provider 插件在滚动升级多版本并存时锁定本副本 localVersion
- **WHEN** 本副本扫描锁定本地插件 `tencentcloud@1.81.5`，`iac_plugin` 表中同时存在 `1.81.5`（`enabled`）与新副本注册的 `1.82.0`（`enabled` 或 `init`）
- **THEN** `router.Service` 始终选中本副本 `1.81.5` 记录并查询 `tencentcloud@1.81.5` 的 `PluginSchema`

### Requirement: 活跃 Schema 类型存在性与 Capabilities 能力位门控（规则 R-002）
系统 SHALL 通过 `SchemaManager.GetSchema(kt, PluginInfo{Name: plugin, Version: targetPlugin.PluginVersion})` 获取插件目标启用版本的完整 `PluginSchema`，并对运行态请求的 `type_name` 与 `capabilities` 执行前置门控：
- 对 `ReadData`（`IaCPluginOperationReadData`）接口，`type_name` MUST 存在于 `data_source_schemas` 中；
- 对 `Validate`（`IaCPluginOperationValidate`）接口，当 `req.ConfigType == "resource"` 时 `type_name` MUST 存在于 `resource_schemas` 中，当 `req.ConfigType == "data"` 时 `type_name` MUST 存在于 `data_source_schemas` 中；
- 对其余资源接口（`Execute`、`Plan`、`Import`、`Refresh`、`List`、`GetResult`、`Stop`），`type_name` MUST 存在于 `resource_schemas` 中；
- 当前操作对应的 `capabilities` 能力位 MUST 为 `true`（`Execute` 需 `apply=true`，`GetResult` 需 `async_apply=true`，`Stop` 需 `stop=true && async_apply=true`，`Plan` 需 `plan=true`，`Validate` 需 `validate=true`，`Import` 需 `import=true`，`Refresh` 需 `refresh=true`，`ReadData` 需 `read_data=true`，`List` 需 `list=true`）。

#### Scenario: 资源与数据源类型存在且能力位已声明时校验通过
- **WHEN** `bkcmdb@1.2.0` 的活跃 Schema 包含资源 `bkcmdb_host`、数据源 `bkcmdb_biz`，且 `capabilities` 声明 `apply=true, plan=true, refresh=true, read_data=true, validate=true`，调用方对 `bkcmdb_host` 调用 `Execute`/`Plan`/`Refresh`/`Validate(config_type="resource")` 或对 `bkcmdb_biz` 调用 `ReadData`/`Validate(config_type="data")`
- **THEN** `router.Service` 类型与能力位门控校验通过，允许继续解析实例并路由执行

#### Scenario: 类型不存在、资源/数据源类型错位或能力位未声明时拦截
- **WHEN** 请求的 `type_name="bkcmdb_unknown"` 不在 Schema 中，或对仅在 `resource_schemas` 中的 `bkcmdb_host` 调用 `ReadData` / `Validate(config_type="data")`，或本地插件 `capabilities.async_apply=false` / `capabilities.list=false` 时调用 `GetResult`、`Stop`、`List`
- **THEN** `router.Service` 直接返回 `errf.InvalidParameter`（`2000001`）或 `errf.Aborted`（`2000006`），下游南向插件调用次数与进程池 `Acquire` 调用次数均为 0

### Requirement: 插件实例归属强校验与同构树深度合并复用（规则 R-001）
系统 SHALL 在 `router.Service` 中通过 `plugininst.PluginInstanceManager.GetPluginInstance(kt, instanceID)` 读取 `plugin_instance_id` 对应的插件实例记录并完成 `attrs` / `secret_attrs` 同构树深度合并与归属强校验：
- 若实例不存在，系统 MUST 返回 `errf.RecordNotFound`（`2000003`）；
- `PluginInstanceManager.GetPluginInstance` 内部将普通属性树 `attrs` 与已由 `data-service` 解密的敏感属性树 `secret_attrs` 递归深度合并（嵌套 `map[string]any` 同名键递归合并子节点，`[]any` 按下标逐元素递归合并子节点，标量叶子以 `secret_attrs` 覆盖或填入）还原为完整明文 `inst.Properties`；
- `router.Service` MUST 强校验 `inst.PluginName == req.Plugin`，若不一致则直接返回 `errf.InvalidParameter`（`2000001`），严防跨插件混用实例套取凭证，且严禁在任何级别日志或错误消息中打印明文属性值；
- 当插件 `protocol == "http"` 时，组装 `*RouteContext` 分发至 `*httpplugin.Service` 对应方法；当插件 `protocol == "tf_provider"` 时，通过 `*pool.PluginRegistry` 获取插件进程池并构造 `*adaptor.TFPluginAdaptor`（注册 `defer adaptor.Close()`），分发 `Execute`、`Plan`、`Validate`、`Import`、`Refresh`、`ReadData` 六个运行态方法。

#### Scenario: 实例归属匹配且同构树深度合并成功分发
- **WHEN** 插件处于 `enabled` 状态，实例 `"00000002"` 的 `plugin_name` 与请求 `plugin` 一致，且协议为 `http` 或 `tf_provider`
- **THEN** `router.Service` 通过 `PluginInstanceManager.GetPluginInstance` 获取合并后的明文 `Properties`，校验归属通过并按协议分发至 `*httpplugin.Service` 或 `*adaptor.TFPluginAdaptor`（请求结束后自动归还 `Lease`）

#### Scenario: 跨插件混用实例或实例不存在时拒绝访问
- **WHEN** 请求 `plugin="bkcmdb"` 但传入的 `plugin_instance_id` 归属 `tencentcloud`（`inst.PluginName != "bkcmdb"`），或 `plugin_instance_id` 在数据库中不存在
- **THEN** `router.Service` 分别返回 `errf.InvalidParameter`（`2000001`）或 `errf.RecordNotFound`（`2000003`），不向网关或进程池发起任何请求

### Requirement: 北向无实例接口 GetSchema（零南向调用）与 Health 探活（规则 R-007）
系统 SHALL 提供无需 `plugin_instance_id` 的北向 `GetSchema` 与 `Health` 接口：
- `GetSchema`（`GET /api/v1/hc/iac_plugins/schema`）：若传入非空 `plugin_version` 则直接调用 `SchemaManager.GetSchema(kt, PluginInfo{Name: plugin, Version: plugin_version})`；若未传 `plugin_version` 则先查询 `iac_plugin` 获取该 `plugin` 的启用版本（本地插件取本副本 `localVersion`，三方插件取最高版本）并校验 `state == "enabled"`，再调用 `SchemaManager.GetSchema` 返回完整 `*coreplugin.PluginSchema`；全程 MUST 只读进程内 `SchemaCache` 与 `data-service` `iac_plugin_schema` 表，**触发调用三方南向 HTTP `/schema` 或拉起本地插件子进程的次数 MUST 为 0**；
- `Health`（`GET /api/v1/hc/iac_plugins/health`）：当目标插件为三方 `http` 插件时查询最高版本记录并调用 `httpPlugin.Health(kt, latest)` 探活返回 `{ "plugin_version": health.PluginVersion }`；当目标插件为本地 `tf_provider` 插件时，校验本副本存在该插件 `discovery.Entry` 且 `iac_plugin` 表中 `(plugin, localVersion)` 状态为 `enabled`，返回 `{ "plugin_version": localVersion }`；不满足条件或探活失败时返回 `errf.Aborted`（`code=2000006, data=nil`）。

#### Scenario: GetSchema 读取启用版本或指定历史版本且零南向 HTTP 调用
- **WHEN** `iac_plugin_schema` 已存有 `bkcmdb@1.2.0` 的 Schema，调用方请求 `GET /api/v1/hc/iac_plugins/schema?plugin=bkcmdb` 或 `GET /api/v1/hc/iac_plugins/schema?plugin=bkcmdb&plugin_version=1.2.0`
- **THEN** 返回完整 `PluginSchema`，且对三方南向 HTTP `/schema` 的调用次数为 0

#### Scenario: 本地 tf_provider 插件 Health 探活返回本副本 localVersion
- **WHEN** 本副本扫描到本地插件 `tencentcloud@1.81.5` 且 `iac_plugin` 表中 `tencentcloud@1.81.5` 状态为 `enabled`，调用 `GET /api/v1/hc/iac_plugins/health?plugin=tencentcloud`
- **THEN** 返回外层 `code=0` 及 `data.plugin_version="1.81.5"`；若 `tencentcloud@1.81.5` 状态为 `disabled` 或 `failed` 则返回 `errf.Aborted`（`code=2000006`）

#### Scenario: GetSchema 查询不存在的插件/版本或 Health 探活三方网关失败
- **WHEN** ① 调用 `GetSchema` 查询不存在的插件 `not_exist` 或不存在的版本 `9.9.9`；或 ② 调用 `Health` 时三方插件网关不可达或返回非 0 错误
- **THEN** ① 返回 `errf.RecordNotFound`（`code=2000003, data=nil`）；② 返回 `errf.Aborted`（`code=2000006, data=nil`）
