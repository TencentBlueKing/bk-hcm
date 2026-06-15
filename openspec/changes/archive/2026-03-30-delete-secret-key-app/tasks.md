## 1. 请求类型定义

- [x] 1.1 新建 `pkg/api/cloud-server/application/delete_sub_account_secret.go`，定义批量删除请求类型

## 2. hc-service 客户端补充

- [x] 2.1 在 `pkg/client/hc-service/tcloud/account.go` 新增 `DeleteAccessKey` 方法

## 3. Handler 包实现

- [x] 3.1 新建 `handlers/sub-account/delete-secret-key/init.go`：Handler 结构体、构造函数、init 注册
- [x] 3.2 新建 `handlers/sub-account/delete-secret-key/check.go`：参数校验、密钥存在性、三级账号存在性
- [x] 3.3 新建 `handlers/sub-account/delete-secret-key/prepare.go`：内容结构体、GenerateApplicationContent、GetItsmApprover
- [x] 3.4 新建 `handlers/sub-account/delete-secret-key/create_itsm_ticket.go`：RenderItsmTitle、RenderItsmForm
- [x] 3.5 新建 `handlers/sub-account/delete-secret-key/deliver.go`：Deliver 交付逻辑（云上删除 + 本地删除）

## 4. 路由与创建函数

- [x] 4.1 在 `cmd/cloud-server/service/application/init.go` 新增路由
- [x] 4.2 在 `cmd/cloud-server/service/application/create.go` 新增 import 和创建处理函数

## 5. 验证

- [x] 5.1 编译验证 `cmd/cloud-server`
- [x] 5.2 Linter 检查
