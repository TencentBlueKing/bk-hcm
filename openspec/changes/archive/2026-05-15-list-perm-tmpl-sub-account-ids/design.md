## Context

- 接口文档要求：`GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/{id}/sub_account_ids`，全量返回关联的三级账号 ID 列表，不分页。
- 表定义：`sub_account` 表中 `permission_template_ids types.StringArray`（JSON 数组）存储三级账号关联的权限模板 ID，`bk_biz_ids types.Int64Array` 存储业务 ID 列表。
- data-service 现有接口：`POST /sub_accounts/list`（`ListSubAccount`）和 `POST /vendors/{vendor}/sub_accounts/list`（`ListSubAccountExt`）均接受 `core.ListReq`（含 `filter.Expression` + `Page` + `Fields`），已支持通用过滤查询。
- DAO 层现状：`pkg/dal/dao/cloud/sub-account/sub_account.go` 的 `List` 方法中 `columnTypes` 注册了 `extension.uin`、`extension.console_login`、`extension.cloud_main_account_id`，**未注册** `permission_template_ids`，因此当前无法通过 `filter.Expression` 的 `json_overlaps` 操作符查询该字段。
- 过滤工具：`pkg/dal/dao/tools/filter.go` 中 `RuleJsonOverlaps[T](fieldName, values)` 可生成 `JSON_OVERLAPS` 的 `AtomRule`，已在 `load-balancer`、`task` 等模块使用。
- 权限验证：业务访问权限，与现有 `ListBizPermissionTemplate` 一致。

## Goals / Non-Goals

**Goals:**

- 实现 `GET .../permission_templates/{id}/sub_account_ids` 接口，返回指定权限模板关联的三级账号 ID 列表。
- 仅返回业务 ID 匹配的三级账号（`bk_biz_ids` 包含 `bk_biz_id`）。
- 全量返回，不分页。
- 复用 data-service 现有的 `ListExt` 接口，通过 `filter.Expression` 查询，不在 data-service 新增接口。
- 在 sub_account DAO 的 `columnTypes` 中注册 `permission_template_ids`，使 `json_overlaps` 过滤生效。

**Non-Goals:**

- 不在本次变更中修改权限模板的其他接口。
- 不在 data-service 新增接口或路由。
- 不支持分页（需求明确全量返回）。

## Decisions

1. **复用 data-service `ListSubAccountExt` 接口**
   - **做法**：cloud-server 构造 `core.ListReq`，使用 `filter.Expression`（`json_overlaps` 操作符）过滤 `permission_template_ids` 包含指定模板 ID 的记录，同时过滤 `bk_biz_ids` 包含指定业务 ID 的记录。
   - **理由**：data-service 已有通用列表接口，支持 `filter.Expression`；无需新增接口。
   - **备选**：新增 data-service 专用接口 — **否决**，过度设计。

2. **注册 `permission_template_ids` 到 DAO `columnTypes`**
   - **做法**：在 `pkg/dal/dao/cloud/sub-account/sub_account.go` 的 `List` 方法中，添加 `columnTypes["permission_template_ids"] = enumor.String`。
   - **理由**：`permission_template_ids` 是 `types.StringArray`（JSON 字符串数组），`json_overlaps` 操作符需要字段在 `columnTypes` 中注册才能通过校验。注册为 `enumor.String` 与字段内元素类型一致。
   - **备选**：在 cloud-server 层手写 SQL — **否决**，不符合项目架构分层。

3. **返回 `cloud_id` 作为 `sub_account_ids`**
   - **做法**：接口响应中的 `sub_account_ids` 使用三级账号的 `cloud_id`（云上 ID），而非本地 `id`。
   - **理由**：接口文档明确响应示例为 `["00000001", "00000002"]`，且字段描述为"三级账号ID"，根据业务语义应为云侧 ID。
   - **备选**：返回本地 `id` — 需与文档/前端确认，当前按文档实现。

4. **全量不分页**
   - **做法**：设置 `Page.Limit` 为 `constant.MaxBatchLimit`（或其他项目约定最大值），`Page.Start = 0`，一次性取回所有匹配记录。
   - **理由**：需求明确全量返回；一个权限模板关联的三级账号数量有限，不会造成性能问题。

5. **接口路径与鉴权**
   - **做法**：`GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/{id}/sub_account_ids`；业务访问权限。
   - **理由**：与现有 `ListBizPermissionTemplate` 保持一致。

## Risks / Trade-offs

- **[Risk] `permission_template_ids` 注册类型不当** → **Mitigation**：字段为 `types.StringArray`（JSON 字符串数组），注册为 `enumor.String` 与 `RuleJsonOverlaps` 传入 `[]string` 一致，`JSONOverlapsOp.SQLExprAndValue` 生成 `JSON_OVERLAPS(permission_template_ids, JSON_ARRAY(:placeholder))` 语法正确。
- **[Risk] 全量返回数量过多** → **Mitigation**：一个权限模板关联的三级账号数量有限；若后续需要分页可扩展。
- **[Risk] `bk_biz_ids` 字段也需要注册 `columnTypes`** → **Mitigation**：检查现有 `columnTypes` 是否已注册 `bk_biz_ids`；若未注册需一并添加。

## Migration Plan

- 纯新增 API + DAO 字段注册：无数据迁移。
- 部署顺序：先 **data-service**（DAO `columnTypes` 补充），再 **cloud-server**（新路由），或同步发布；兼容旧客户端（仅新增路径）。
