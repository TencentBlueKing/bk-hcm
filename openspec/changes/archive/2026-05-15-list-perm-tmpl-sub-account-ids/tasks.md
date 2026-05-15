## 1. DAO 字段注册

- [x] 1.1 在 `pkg/dal/dao/cloud/sub-account/sub_account.go` 的 `List` 方法 `columnTypes` 中注册 `permission_template_ids` 字段（类型 `enumor.String`），使其支持 `json_overlaps` 操作符过滤
- [x] 1.2 确认 `bk_biz_ids` 是否已在 `columnTypes` 中注册；若未注册，添加 `columnTypes["bk_biz_ids"] = enumor.Numeric`

## 2. API 类型定义

- [x] 2.1 在 `pkg/api/cloud-server/permission_template.go` 新增响应结构体 `PermTmplSubAccountIDsResult`，包含 `SubAccountIDs []string` 字段

## 3. cloud-server 实现

- [x] 3.1 在 `cmd/cloud-server/service/permission-templates/list.go` 新增 `ListPermTmplSubAccountIDs` 方法：
  - 解析路径参数 `bk_biz_id`、`vendor`、`id`
  - 校验 vendor 和 biz_id
  - 业务访问鉴权
  - 构造 `filter.Expression`：`RuleJsonOverlaps("permission_template_ids", []string{templateID})` AND `RuleJsonOverlaps("bk_biz_ids", []int64{bizID})`
  - 构造 `core.ListReq`：filter + Page（全量，Limit 为最大值）+ Fields（仅 `cloud_id`）
  - 调用 data-service `TCloud.SubAccount.ListExt`
  - 提取返回记录的 `CloudID` 组装为 `sub_account_ids`
  - 返回 `PermTmplSubAccountIDsResult`
- [x] 3.2 在 `cmd/cloud-server/service/permission-templates/service.go` 注册新路由：`GET /bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/{id}/sub_account_ids`

## 4. 验证

- [x] 4.1 本地 `go build` 确保编译通过
- [x] 4.2 对照接口文档核对请求路径、参数、响应格式
