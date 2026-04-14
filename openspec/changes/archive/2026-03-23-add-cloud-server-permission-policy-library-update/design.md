## Context

cloud-server 层的 permission policy library 当前仅实现 Create 和 List 接口。data-service 层已有完整的 BatchUpdate 链路（handler → DAO → DB），TCloud client 的 `BatchUpdate` 方法也已可用。cloud-server 层需要补齐 Update handler，并遵循项目中其他资源（Account、VPC、SecurityGroup、Subnet 等）的统一审计模式。

当前 svc struct 只包含 `client` 和 `authorizer`，不含 `audit`，需要扩展。

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 层实现单条更新接口 `PATCH /vendors/{vendor}/permission_policy_libraries/{id}`
- 遵循项目既有的审计模式：`converter.StructToMap(req)` → `svc.audit.ResUpdateAudit()`，在执行更新**之前**记录审计
- 复用已有的 data-service BatchUpdate 链路（单条→批量适配）
- 保持与 Create/List handler 一致的代码风格和鉴权模式

**Non-Goals:**
- 不修改 data-service 层的 BatchUpdate 逻辑
- 不处理 BkBizIDs 空数组清空语义问题（保持现有行为）
- 不为 Create 操作补充 cloud-server 层审计（Create 审计已在 DAO 层实现）
- 不实现批量更新（cloud-server 层接口仅支持单条更新）

## Decisions

### 1. svc struct 扩展 audit 字段

**决策**: 在 `svc` struct 中新增 `audit audit.Interface` 字段，从 `c.Audit` 初始化。

**理由**: 这是项目中所有需要审计的 service（Account、VPC、SecurityGroup 等）的统一做法，保持一致性。

**替代方案**: 在 handler 内部直接创建 audit client → 违背现有模式，增加耦合。

### 2. 审计时序：先审计后更新

**决策**: 在调用 DS BatchUpdate 之前记录审计。

**理由**: `permissionPolicyLibraryUpdateAuditBuild`（data-service 审计构建函数）会查询原始数据存入 `Detail.Data`，配合 `Detail.Changed`（变更字段）形成完整的变更前后对比。如果先更新再审计，原始数据就已被覆盖。这与 Account/VPC/SecurityGroup 的 Update handler 一致。

### 3. 单条→批量适配模式

**决策**: cloud-server 接收路径参数 `{id}`，将请求体的字段连同 ID 包装为 `PermissionPolicyLibraryBatchUpdateReq`（数组长度为 1），调用 DS TCloud client 的 `BatchUpdate`。

**理由**: 与 Create 接口的适配模式一致（单条请求 → DS 层批量接口），复用已有链路。

### 4. 鉴权方式：AuthorizeWithPerm

**决策**: 使用 `authorizer.AuthorizeWithPerm` 进行 IAM 鉴权，资源属性为 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Update}`。

**理由**: 与 Create handler 的鉴权模式一致。permission policy library 不绑定业务（无 bk_biz_id 归属关系），采用资源类型级别鉴权而非实例级别鉴权。

**替代方案**: 使用 `ValidWithAuthHandler` + `GetResBasicInfo` 进行实例级鉴权 → 不适用，因为 permission_policy_library 不在 `cloud_basic_info` 资源体系中。

## Risks / Trade-offs

- **[审计失败导致更新被阻断]** → 审计调用失败时 handler 直接返回错误，不执行更新。这是项目既有行为，保证审计完整性优先于操作可用性。
- **[Memo 字段零值语义]** → `PermissionPolicyLibraryUpdate` 中 `Memo` 为 `*string` 类型，cloud-server 层的 `UpdateReq` 中 `Memo` 也应使用 `*string` 以区分"不传"和"传空字符串"。
