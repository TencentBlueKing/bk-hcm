## RENAMED Requirements

### Requirement: GPU 模版字段名统一
FROM: `Schema` (json: `"schema"`)
TO: `TplSchema` (json: `"tpl_schema"`)

涉及 API 请求/响应体及 table 层的 JSON 序列化字段名，与 DB 列名 `tpl_schema` 保持一致。

#### Scenario: 创建 GPU 模版使用新字段名
- **GIVEN** 调用方提交创建 GPU 模版请求
- **WHEN** 请求体中使用 `"tpl_schema"` 字段传递模版内容
- **THEN** 系统正确解析并创建模版记录

#### Scenario: 更新 GPU 模版使用新字段名
- **GIVEN** 调用方提交更新 GPU 模版请求
- **WHEN** 请求体中使用 `"tpl_schema"` 字段传递模版内容
- **THEN** 系统正确解析并更新模版记录

#### Scenario: 查询 GPU 模版返回新字段名
- **GIVEN** 已有 GPU 模版数据
- **WHEN** 查询模版列表或详情
- **THEN** 响应体中模版内容字段名为 `"tpl_schema"`
