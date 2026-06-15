## Why

当前系统缺少一个根据资源类型（sub_account、sub_account_secret、permission_template）批量查询二级账号元数据信息的接口。用户需要在不同资源维度上查看哪些二级账号有权限被访问，且不同资源类型的权限校验逻辑不同。现有的 `GetAccountInfo` 接口仅支持单个账号查询，已不满足业务需求，需替换为批量查询接口。

## What Changes

- 新增 `ListAccountInfoByType` 接口：`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/accounts/by/type`
- 新增请求结构体 `AccountListByTypeReq`（ids、type 字段）和响应结构体 `AccountListByTypeResp`（details 数组）
- 删除现有 `GetAccountInfo` 接口及相关代码（仅支持单账号查询，已被新批量接口替代）
- 调整现有 `accountResourceAuthChecker` 接口，扩展为支持批量过滤的 `accountTypeAuthChecker`，替换原有单账号维度的校验模式
- 实现基于资源类型的权限校验器策略模式：`accountTypeAuthChecker` 接口，不同资源类型（sub_account、sub_account_secret、permission_template）实现各自的校验逻辑
- 校验流程：先校验业务访问权限，再按资源类型调用对应校验器过滤出有权限的二级账号，最后批量查询账号基本信息并构建扩展字段
- 在 `accountSvc` 上注册新路由

## Capabilities

### New Capabilities
- `account-info-by-type`: 根据资源类型批量查询业务下关联资源的二级账号元数据信息，包含权限校验策略和扩展字段构建

### Modified Capabilities

## Impact

- `cmd/cloud-server/service/account/service.go`：注册新路由，移除 GetAccountInfo 路由
- `cmd/cloud-server/service/account/info.go`：删除 `GetAccountInfo` 方法及 `accountResourceAuthChecker` 接口和 `subAccountChecker` 实现
- `cmd/cloud-server/service/account/`：新增 `info_by_type.go` 文件实现接口逻辑，新增 `info_by_type_checker.go` 实现各类型校验器
- `pkg/api/cloud-server/account/list.go`：新增请求和响应结构体，移除 GetAccountInfo 相关结构体
- `pkg/client/data-service/`：可能需要新增批量查询扩展字段的 client 方法
