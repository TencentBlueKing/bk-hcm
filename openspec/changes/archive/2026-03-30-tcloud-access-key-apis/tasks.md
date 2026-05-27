## 1. Adaptor 类型定义

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud.go` 中新增 `CreateAccessKeyOption`、`CreateAccessKeyResult`、`DeleteAccessKeyOption`、`UpdateAccessKeyOption` 类型及其 Validate 方法

## 2. Adaptor 接口声明与实现

- [x] 2.1 在 `pkg/adaptor/tcloud/interface.go` 的 TCloud 接口中声明 `CreateAccessKey`、`DeleteAccessKey`、`UpdateAccessKey` 三个方法
- [x] 2.2 在 `pkg/adaptor/tcloud/account.go` 中实现三个方法，调用腾讯云 CAM SDK 对应 API

## 3. hc-service API 类型定义

- [x] 3.1 在 `pkg/api/hc-service/sub-account/tcloud.go` 中新增 `TCloudCreateAccessKeyReq`/`Resp`、`TCloudDeleteAccessKeyReq`、`TCloudUpdateAccessKeyReq` 类型及 Validate 方法

## 4. hc-service Handler 实现

- [x] 4.1 在 `cmd/hc-service/service/account/secret.go` 中实现 `TCloudCreateAccessKey`、`TCloudDeleteAccessKey`、`TCloudUpdateAccessKey` 三个 handler 方法

## 5. 路由注册

- [x] 5.1 在 `cmd/hc-service/service/account/service.go` 中注册三个新路由：`/vendors/tcloud/sub_accounts/secrets/create`、`/vendors/tcloud/sub_accounts/secrets/delete`、`/vendors/tcloud/sub_accounts/secrets/update`

## 6. 验证

- [x] 6.1 确认编译通过（`go build ./cmd/hc-service/...`）
