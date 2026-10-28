# Spec Delta

## MODIFIED Requirements

### Requirement: GetResult、Stop、Validate、Import、Refresh、ReadData 与 List 接口参数校验
系统 SHALL 对其余 7 个北向 POST 接口执行严格的前置参数校验与响应字段承载：
- `POST /api/v1/hc/iac_plugins/result` 与 `POST /api/v1/hc/iac_plugins/stop`：`apply_id` MUST 为非空白字符串；
- `POST /api/v1/hc/iac_plugins/validate`：`config_type` MUST 为必填字段且取值 MUST 为 `"resource"` 或 `"data"`（常量 `ConfigTypeResource` / `ConfigTypeDataSource`），`config` MUST 为非 `nil` 的 `map[string]any`；
- `POST /api/v1/hc/iac_plugins/read_data`：`config` MUST 为非 `nil` 的 `map[string]any`，响应 `ReadDataResult` 除 `state` 外 SHALL 包含可选 `diagnostics`（`[]coreplugin.Diagnostic`）字段；
- `POST /api/v1/hc/iac_plugins/import`：`import_id` MUST 为非空白字符串，响应 `StateResult` 除 `state`、`private`、`identity` 外 SHALL 包含可选 `diagnostics`（`[]coreplugin.Diagnostic`）字段；
- `POST /api/v1/hc/iac_plugins/refresh`：`state` MUST 为非空 `map[string]any`（`len(state) > 0`），可选 `private` 非空时 MUST 为合法标准 Base64 字符串，响应复用 `StateResult`（含可选 `diagnostics` 字段）；
- `POST /api/v1/hc/iac_plugins/list`：`limit` MUST 满足 `coreplugin.MinListLimit(1) <= limit <= coreplugin.MaxListLimit(500)`，支持可选 `config` 过滤条件与 `page_token` 游标（首页为 `""`）。

#### Scenario: List 分页 limit 边界校验与其余接口必填项校验拦截
- **WHEN** 调用 `List` 传入 `limit=0` 或 `limit=501`，或调用 `GetResult`/`Stop` 传入空 `apply_id`，或调用 `Import` 传入空 `import_id`，或调用 `Refresh` 传入空 `state`，或调用 `Validate`/`ReadData` 传入 `config=nil`，或调用 `Validate` 未传 `config_type` 或传入非 `"resource"`/`"data"` 的非法值
- **THEN** Handler 在前置校验阶段 100% 拦截并返回 `errf.InvalidParameter`（`code=2000001`）；若请求体 JSON 语法错误则返回 `errf.DecodeRequestFailed`（`code=2000004`）

#### Scenario: Validate 传入合法 config_type ("resource" 或 "data") 校验通过
- **WHEN** 调用 `ValidateReq.Validate()` 并传入 `config_type="resource"` 或 `config_type="data"` 且 `config != nil`
- **THEN** `Validate()` 返回 `nil`
