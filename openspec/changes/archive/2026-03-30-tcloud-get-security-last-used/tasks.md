## 1. Adaptor 类型定义

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud.go` 中新增 `GetSecurityLastUsedOption`、`SecretIdLastUsed` 类型

## 2. Adaptor 接口声明与实现

- [x] 2.1 在 `pkg/adaptor/tcloud/interface.go` 中声明 `GetSecurityLastUsed` 方法
- [x] 2.2 在 `pkg/adaptor/tcloud/account.go` 中实现该方法

## 3. hc-service API 类型与 Handler

- [x] 3.1 在 `pkg/api/hc-service/sub-account/tcloud.go` 中新增 `TCloudGetSecurityLastUsedReq` 类型
- [x] 3.2 在 `cmd/hc-service/service/account/secret.go` 中实现 handler
- [x] 3.3 在 `cmd/hc-service/service/account/service.go` 中注册路由

## 4. 验证

- [x] 4.1 确认编译通过
