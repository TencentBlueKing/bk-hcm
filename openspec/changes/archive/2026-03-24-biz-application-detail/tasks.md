## 1. 路由注册

- [x] 1.1 在 `cmd/cloud-server/service/application/init.go` 中新增路由 `GET /bizs/{bk_biz_id}/applications/{application_id}`，Handler 为 `svc.GetBizApplication`

## 2. 接口实现

- [x] 2.1 在 `cmd/cloud-server/service/application/get.go` 中新增 `GetBizApplication` 函数，解析路径参数 `bk_biz_id` 和 `application_id`
- [x] 2.2 实现业务访问权限鉴权（`meta.Biz` + `meta.Access`），鉴权失败返回 `NotFound`
- [x] 2.3 调用 Data Service 获取单据详情，复用 `a.client.DataService().Global.Application.GetApplication()`
- [x] 2.4 实现归属校验：检查 `bk_biz_id` 是否在单据的 `bk_biz_ids` 列表中，不匹配返回 `NotFound`
- [x] 2.5 构建返回体：复用 `ApplicationGetResp`，调用 `RemoveSenseField()` 脱敏，ITSM来源获取 `TicketUrl`
- [x] 2.6 添加内部日志：鉴权失败和归属不匹配记录 WARN 日志，区分原因

## 3. 文档

- [x] 3.1 新增接口文档 `docs/api-docs/web-server/docs/biz/get_biz_application.md`，说明接口路径、参数、权限要求、返回结构和错误码
