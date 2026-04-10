## Context

项目已有成熟的 data-service 分层架构（table → DAO → handler → SDK client）。`permission_policy_library` 是一张新的业务表，需要遵循此架构接入。表中 `policy_hash`（SHA256）和 `version` 需要由 data-service 层自动维护，对上层服务透明。

当前参考实现：`cert`（cloud 目录下的典型 CRUD 资源）、`main_account`（有 vendor + 租户隔离的写操作）。

## Goals / Non-Goals

**Goals:**
- 在 `cloud/` 目录下实现权限策略库的完整 CRUD 数据层接口
- DataService 层自动计算并维护 `policy_hash` 和 `version`（内容不变则不递增 version）
- 通过 TCloud vendor SDK 暴露写操作（Create/Update），通过 Global SDK 暴露读/删操作
- 支持多租户隔离（`InjectTenantIDOpt`）

**Non-Goals:**
- 策略内容的语法校验（由上层调用方负责）
- 审计日志（本次不做）
- 除 TCloud 外的其他 vendor SDK 客户端（后续按需扩展）

## Decisions

### 1. 目录放在 `cloud/` 下

与 `cert`、`argument-template` 保持一致，放于 `cloud/` 目录下。`permission_policy_library` 是与云厂商相关的通用资源定义，不属于账号体系（`account-set/`）。

### 2. version 只在 policy_hash 变化时递增

Update 流程：
```
1. 上层传入 policy_document（可能为空）
2. 若 policy_document 非空：
   a. 计算 new_hash = SHA256(policy_document)
   b. 通过 List 查询当前记录获取 existing_hash 和 existing_version
   c. 若 new_hash != existing_hash → 更新 hash，version = existing_version + 1
   d. 若 new_hash == existing_hash → 不更新 hash/version（内容未变化）
3. 其他字段（name、bk_biz_ids、memo）按正常 update 处理
```

**替代方案**：用 SQL `IF(policy_hash != 'new_hash', version+1, version)` — 拒绝，因为 `RearrangeSQLDataWithOption` 不支持条件表达式，维护成本高。

### 3. vendor 放在写操作 URL 中

- `POST /vendors/{vendor}/permission_policy_libraries/create`
- `PATCH /vendors/{vendor}/permission_policy_libraries/{id}`
- `DELETE /permission_policy_libraries/batch`（无 vendor）
- `POST /permission_policy_libraries/list`（无 vendor）
- `GET /permission_policy_libraries/{id}`（无 vendor）

与 `cert`、`main_account` 路由惯例一致。

### 4. SDK 客户端分布

- `pkg/client/data-service/tcloud/permission_policy_library.go`：Create、Update（方法直接挂在 `restClient` 上，无需改 tcloud/client.go 的 struct）
- `pkg/client/data-service/global/permission_policy_library.go`：List、Get、BatchDelete（独立 `PermissionPolicyLibraryClient` struct，需在 global/client.go 注册）

### 5. 租户隔离

表含 `tenant_id` 字段，所有 SQL 操作使用 `orm.NewInjectTenantIDOpt(kt.TenantID)` 自动注入，与 `main_account` 一致。表注册时设置 `EnableTenant: true`。

### 6. 无 Audit，无加密

策略内容为明文 JSON，不涉及密钥，无需加密。本次不做审计记录，DAO struct 只需 `Orm` + `IDGen`。

## Risks / Trade-offs

- **Update 时多一次查询** → 为了获取 existing_hash 进行对比，需要先 List 一次。低频操作可接受，若有性能问题后续可缓存。
- **bk_biz_ids 的 JSON 过滤** → 作为 JSON 数组字段，filter 系统的 JSON 查询能力依赖底层 MySQL `json_contains`，需确认 filter 框架已支持。
