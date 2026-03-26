## 1. Proto 层：新增请求结构体

- [x] 1.1 在 `pkg/api/cloud-server/application/create.go` 中新增 `SysCreateCommonReq` 结构体，内嵌 `CreateCommonReq`，新增 `Applicant string` 字段（`json:"applicant" validate:"required,min=1"`），并实现 `Validate()` 方法

## 2. 重构：create 链路支持显式 applicant 参数

- [x] 2.1 为 `create()`、`createItsmTicket()`、`createApplication()` 增加 `applicant string` 参数，分别用于 ITSM 工单的 `Creator` 和申请单的 `Applicant`
- [x] 2.2 更新所有已有调用方（`CreateForAddAccount`、`CreateForCreateCvm`、`CreateForCreateVpc`、`CreateForCreateDisk`、`CreateForCreateLB`、`CreateForCreateMainAccount`、`CreateForUpdateMainAccount`），传递 `cts.Kit.User` 作为 applicant 参数

## 3. Handler 层：新增系统提单方法

- [x] 3.1 在 `cmd/cloud-server/service/application/create.go` 中新增 `SysCreateForCreateLB` 方法（紧跟 `CreateForCreateLB` 之后），实现逻辑：解析 vendor → 解码 `SysCreateCommonReq` 并校验 → 权限检查（使用 Kit.User 即系统调用方身份）→ 按 vendor 分发到 TCloud/TCloudZiyan Handler → 调用 `create()` 并传递 `req.Applicant`

## 4. 路由注册

- [x] 4.1 在 `cmd/cloud-server/service/application/init.go` 中注册新路由：`h.Add("SysCreateForCreateLB", "POST", "/vendors/{vendor}/system/applications/types/create_load_balancer", svc.SysCreateForCreateLB)`，放在现有 `CreateForCreateLB` 路由之后

## 5. 接口文档

- [x] 5.1 参考 `docs/api-docs/web-server/docs/resource/load-balancer/create_application_for_create_tcloud_load_balancer.md` 格式，在同目录下新增系统提单接口文档 `sys_create_application_for_create_tcloud_load_balancer.md`，说明路径、请求参数（含 `applicant`）、响应格式，版本使用 `v9.9.9`
- [x] 5.2 参考 `docs/api-docs/web-server/docs/resource/load-balancer/create_application_for_create_tcloud_ziyan_load_balancer.md` 格式，在同目录下新增 `sys_create_application_for_create_tcloud_ziyan_load_balancer.md`
