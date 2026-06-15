## Why

业务侧需要查询某个云权限模板关联的三级账号ID列表（全量返回，不分页），用于前端展示"该权限模板应用了哪些三级账号"。当前 `sub_account` 表已有 `permission_template_ids` 字段存储关联关系，且 data-service 已有 `ListSubAccount`/`ListExt` 通用列表接口支持 `filter.Expression` 过滤，无需新增 data-service 接口。

## What Changes

- 在 **cloud-server** `cmd/cloud-server/service/permission-templates/list.go` 新增 `ListPermTmplSubAccountIDs` 方法，实现 `GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/{id}/sub_account_ids` 接口。
- 在 **cloud-server** `service.go` 注册新路由（GET 方法）。
- 在 **cloud-server** 方法中增加业务访问权限校验（与现有 `ListBizPermissionTemplate` 一致），校验用户对指定 `bk_biz_id` 的访问权限，无权限则拒绝请求。
- 在 **cloud-server API** `pkg/api/cloud-server/permission_template.go` 新增响应结构体 `PermTmplSubAccountIDsResult`。
- 在 **sub_account DAO** `pkg/dal/dao/cloud/sub-account/sub_account.go` 的 `List` 方法 `columnTypes` 中注册 `permission_template_ids` 字段（类型 `enumor.String`），以支持 `json_overlaps` 操作符过滤。
- cloud-server 调用 data-service 的 `TCloud.SubAccount.ListExt` 接口，使用 `RuleJsonOverlaps("permission_template_ids", []string{templateID})` + `RuleJsonOverlaps("bk_biz_ids", []int64{bizID})` 构造 filter，查询 `permission_template_ids` 包含指定模板 ID 且 `bk_biz_ids` 包含指定业务 ID 的三级账号。
- 仅返回 `sub_account_ids`（即三级账号的 `cloud_id` 字段），全量不分页。

## Capabilities

### New Capabilities

- `biz-permission-template-sub-account-ids`：业务作用域下查询云权限模板关联的三级账号ID列表；全量返回，不分页；仅包含业务 ID 匹配的三级账号。

### Modified Capabilities

- （无）

## Impact

- **cloud-server**：`cmd/cloud-server/service/permission-templates/`（新增方法 + 路由注册）；`pkg/api/cloud-server/permission_template.go`（新增响应类型）。
- **data-service DAO**：`pkg/dal/dao/cloud/sub-account/sub_account.go`（`columnTypes` 补充 `permission_template_ids` 注册）。
- **共享类型**：无新增。
- **数据表**：`sub_account`（查询 `permission_template_ids` 和 `bk_biz_ids` 字段）。
