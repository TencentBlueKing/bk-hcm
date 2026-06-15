## 1. Proto 层结构体定义

- [x] 1.1 在 `pkg/api/cloud-server/account/list.go` 中新增 `AccountListByTypeReq` 请求结构体（IDs、Type 字段），实现 Validate 方法
- [x] 1.2 在 `pkg/api/cloud-server/account/list.go` 中新增 `AccountListByTypeResp` 响应结构体（Details 数组）和 `AccountInfoByTypeDetail` 详情结构体
- [x] 1.3 删除 `pkg/api/cloud-server/account/list.go` 中 `GetAccountInfo` 相关的请求和响应结构体

## 2. 权限校验器实现

- [x] 2.1 在 `cmd/cloud-server/service/account/info_by_type_checker.go` 中定义 `accountTypeAuthChecker` 接口，包含 `filterAuthorizedIDs` 方法，替代原有 `accountResourceAuthChecker`
- [x] 2.2 实现 `subAccountTypeChecker`：查询 sub_account 表，条件为 account_id IN ids AND bk_biz_ids JSON_CONTAINS bizID AND vendor=vendor，通过 countPage 查询 data-service，返回匹配的 account_id 列表
- [x] 2.3 实现 `subAccountSecretTypeChecker`：先查询 sub_account 表获取满足条件的 account_id 列表（同 sub_account 校验），再查询 sub_account_secret 表确认这些 sub_account 记录下是否存在密钥，仅返回有密钥的 account_id
- [x] 2.4 实现 `permissionTemplateTypeChecker`：查询 account 表，条件为 id IN ids AND usage_biz_ids JSON_CONTAINS bizID，校验当前业务是否属于该账号的使用业务，返回匹配的 account id 列表
- [x] 2.5 在 `accountSvc` 中添加 `typeCheckerMap` 初始化逻辑，注册三种校验器

## 3. 接口主逻辑实现

- [x] 3.1 在 `cmd/cloud-server/service/account/info_by_type.go` 中实现 `ListAccountInfoByType` 入口函数：参数校验、业务访问权限校验
- [x] 3.2 实现 `filterAuthorizedAccountIDs`：根据 type 从 typeCheckerMap 获取校验器，调用 filterAuthorizedIDs 过滤
- [x] 3.3 实现 `batchGetAccountBaseInfo`：批量查询账号基本信息（name、bk_biz_id、managers 等）
- [x] 3.4 实现 `buildAccountInfoByTypeDetails`：将账号基本信息和扩展字段组装为响应结构体

## 4. 清理旧接口

- [x] 4.1 删除 `cmd/cloud-server/service/account/info.go` 中的 `GetAccountInfo` 方法、`accountResourceAuthChecker` 接口和 `subAccountChecker` 实现
- [x] 4.2 在 `cmd/cloud-server/service/account/service.go` 中移除 `GetAccountInfo` 路由，注册 `ListAccountInfoByType` 路由

## 5. 验证

- [x] 5.1 编译通过，确认无 lint 错误
