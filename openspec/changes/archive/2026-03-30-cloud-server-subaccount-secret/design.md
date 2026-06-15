## Context

cloud-server 是面向前端的 API 网关层，负责参数/权限/业务逻辑校验后，调用 hc-service（云 API）和 data-service（持久化）。已有 `account-secret` 模块可作为参考模式。路由中使用 `{vendor}` 路径参数，通过 `switch vendor` 分发到厂商特定实现，便于未来扩展其他公有云。

## Goals / Non-Goals

**Goals:**
- 实现创建子账号密钥的完整业务流程，支持多云厂商扩展
- 函数拆分清晰，单个函数不超过 80 行
- 每行不超过 120 列

**Non-Goals:**
- 暂不实现删除、更新、列表接口（后续扩展）
- 暂只实现 TCloud，其他厂商后续按同一模式扩展

## Decisions

### 多云扩展设计

参照 `account-secret` 和 `cos` 模块的模式：
1. 路由含 `{vendor}` 参数：`POST /bizs/{bk_biz_id}/vendors/{vendor}/subaccount_secrets/create`
2. 入口 handler 负责通用校验（参数、权限、子账号归属）
3. `switch vendor` 分发到 `createTCloudSubAccountSecret` 等厂商方法
4. 厂商方法各自负责调用对应的 hc-service 和 data-service client

### 请求流程

1. 解析请求体（子账号 HCM ID）
2. 从 URL path 提取 `bk_biz_id` 和 `vendor`，校验 vendor 合法性
3. IAM 权限校验（`meta.SubAccountSecret` + `meta.Create`）
4. 通过 data-service 查询子账号信息（id, cloud_id, account_id, bk_biz_ids, vendor）
5. 校验子账号存在、vendor 匹配、子账号属于该业务
6. `switch vendor` 分发到厂商实现
7. [TCloud] 将 cloud_id 转为 TargetUin，调用 hc-service `CreateAccessKey`
8. [TCloud] 调用 data-service `BatchCreateSubAccountSecret` 持久化
9. 返回 DB 记录 ID + 密钥扩展信息

### 服务结构

```
cmd/cloud-server/service/subaccount-secret/
  service.go   - InitService, service struct, route registration
  create.go    - CreateBizSubAccountSecret handler + vendor dispatch + TCloud impl
```

### 依赖补充

- `pkg/client/hc-service/tcloud/account.go` 需补充 `CreateAccessKey` 方法
- `pkg/api/cloud-server/sub-account-secret/` 需新增请求/响应类型
