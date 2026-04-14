## 1. 路由注册

- [x] 1.1 在 `cmd/cloud-server/service/application/init.go` 的 `bizService` 中新增 `h.Add("GetBizApplication", "GET", "/applications/{application_id}", svc.GetBizApplication)`

## 2. GetBizApplication Handler 实现

- [x] 2.1 在 `cmd/cloud-server/service/application/get.go` 中新增 `GetBizApplication` handler：解析路径参数 `bk_biz_id` 和 `application_id`
- [x] 2.2 在 handler 中完成业务访问权限鉴权：`authorizer.AuthorizeWithPerm` 使用 `ResourceAttribute{Type: meta.Biz, Action: meta.Access, BizID: bkBizID}`
- [x] 2.3 在 handler 中获取单据详情：调用 `Global.Application.GetApplication`
- [x] 2.4 在 handler 中完成归属校验：检查 `bk_biz_id` 是否在 `application.BkBizIDs` 列表中，使用 `slice.IsItemInSlice`
- [x] 2.5 统一 NotFound 错误策略：权限不足/单据不存在/不归属均返回 `errf.RecordNotFound`
- [x] 2.6 调用公共方法 `buildApplicationGetResp` 构建响应体

## 3. 公共方法抽取

- [x] 3.1 新增 `buildApplicationGetResp` 方法：获取 ITSM 审批链接、构建响应体（包含 Source 字段）、对 Content 进行脱敏处理
- [x] 3.2 修改 `GetApplication` 复用 `buildApplicationGetResp` 方法
