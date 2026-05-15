## 1. 模型与关联确认

- [x] 1.1 阅读 `permission_template`、`sub_account` 表及中间关联（若有），确认 `ListJoinSubAccount` 的 JOIN 条件与 `associated_sub_account_count` 的计算方式
- [x] 1.2 阅读 `account` 表及 `type` 枚举，确认 `cloud_id` + `resource` 查询与 biz/vendor 约束字段

## 2. API 与类型定义

- [x] 2.1 在 `pkg/api/data-service/cloud` 新增 `PermissionTemplateFilters` 及 `ListPermissionTmplJoinExt` 请求/响应类型（含分页、校验 tag）
- [x] 2.2 在 `pkg/api/cloud-server`（或现有 permission template 包）新增与 `list_permission_template.md` 一致的列表请求/响应类型及 `Validate()`
- [x] 2.3 在 `pkg/client/data-service` 增加调用 `ListPermissionTmplJoinExt` 的客户端方法

## 3. DAO 与 data-service

- [x] 3.1 在 `pkg/dal/dao/cloud/permission-template`（或现有 dao 包）实现 `ListJoinSubAccount`：联表 `permission_template` + `sub_account`，支持 filters、`count` 与分页
- [x] 3.2 在 `cmd/data-service/service/cloud` 注册并实现 `ListPermissionTmplJoinExt` handler，组装 kit、校验、调用 DAO
- [ ] 3.3 为 DAO/service 补充必要单测或沿用项目列表测试模式（若仓库要求）

## 4. cloud-server permission_templates 服务

- [x] 4.1 新建 `cmd/cloud-server/service/permission_templates/service.go`：注册 `POST .../permission_templates/list` 至 `ListPermissionTemplate`（或等价命名）
- [x] 4.2 实现 `list.go`：解析路径参数、decode、validate、biz 访问鉴权（对齐 `subaccount-secret/list.go`）；**仅负责编排**，复杂拼装委托下层函数
- [x] 4.3 在 `cmd/cloud-server/logics/account`（或经评估后的共用包）**抽取/复用**：按 `cloud_id` + `type=resource` 解析 `account_id` 的函数；以及按 `account_id` **批量**拉取 account 并返回 `id -> 展示字段` map 的函数（与 sub_account_secret 等列表对齐则合并为公共实现）
- [x] 4.4 实现 account **前置过滤**查询：当存在 `cloud_account_ids`（及与 account 相关的 extension 映射）时，解析 `account_id`；为空则短路返回空结果；若本请求后续仍需拼装，**尽量复用**已加载的 account map，避免重复 RPC
- [x] 4.5 在联表列表返回后：对明细 `account_id` 去重，批量补全 account（与 4.3 的 map 合并），**组装** API 明细（含 `cloud_account_id` 等）；`page.count === true` 时跳过拼装
- [x] 4.6 将 vendor、biz、`account_id` 列表及其余字段映射为 `PermissionTemplateFilters`，调用 data-service 客户端
- [x] 4.7 在主服务路由中挂载 `permission_templates` 服务

## 5. 文档与校验

- [x] 5.1 对照 `docs/api-docs/web-server/docs/biz/permission-template/list_permission_template.md` 核对字段名、分页语义与示例
- [x] 5.2 运行 `openspec validate permission-templates-list`（或项目约定校验）确保 change 工件有效；本地 `go build` 相关模块
- [ ] 5.3 对 logic 层拼装函数补充单测（或在仓库约定下以最小用例覆盖去重批量加载与 count 模式跳过行为）
