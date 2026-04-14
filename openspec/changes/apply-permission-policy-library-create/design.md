## Context

系统已具备权限策略库（`permission_policy_library`）和权限模板（`permission_template`）的 CRUD 能力，分别在 data-service 和 cloud-server 层实现。现在需要实现"应用权限策略库"操作，将策略库的策略内容通过腾讯云 CAM API 创建为云上策略，同时在本地创建对应的 `permission_template` 记录。

当前代码库中：
- TCloud adaptor（`pkg/adaptor/tcloud/`）已有 CAM Client 封装，但仅实现了 `ListPoliciesGrantingServiceAccess`，未实现 `CreatePolicy`
- hc-service 通过 `CloudAdaptorClient.TCloud(kt, accountID)` 获取带密钥的 TCloud adaptor 实例
- cloud-server 不直接持有 adaptor，需通过 hc-service client 调用云 API
- 审计 `AuditAction` 枚举中无 `Apply`，需新增
- `permission_template` 的 data-service client 已存在（`BatchCreate` 等方法）

后续还需实现"应用更新"（`PUT` 同 URL）和 Biz 层审批接口，核心执行逻辑需设计为可复用。

## Goals / Non-Goals

**Goals:**
- 实现 Resource 层 `POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply` 接口
- 封装 TCloud CAM `CreatePolicy` adaptor 能力
- 在 hc-service 暴露 CAM 策略创建接口
- 新增 `Apply` 审计枚举并集成审计（含关联账号信息）
- 设计公共 applier 逻辑，确保后续 apply 和 biz 层可复用

**Non-Goals:**
- 本次不实现 `PUT /permission_policy_libraries/{id}/apply`（应用更新）
- 本次不实现 Biz 层审批接口（`apply_permission_policy_library_create`）
- 本次不实现 CAM `UpdatePolicy`/`DeletePolicy`
- 不修改同步逻辑

## Decisions

### Decision 1: CAM CreatePolicy 放在 hc-service 层

**选择**：在 hc-service 新增 CAM 策略创建 handler，cloud-server 通过 hc-service client 调用。

**理由**：遵循现有架构模式——所有云 API 调用通过 hc-service 的 `CloudAdaptorClient` 完成，密钥获取（`SecretClient.TCloudSecret`）和 adaptor 初始化均封装在 hc-service 内部。cloud-server 不持有 adaptor 实例。

**替代方案**：在 cloud-server 中直接引入 adaptor。否决原因：破坏分层架构，cloud-server 无法获取账号密钥。

### Decision 2: 公共逻辑通过 applier 结构体封装

**选择**：在 `cmd/cloud-server/service/permission-policy-library/` 中新增 `applier.go`，定义 `PolicyLibraryApplier` 结构体，提供可复用的辅助方法。

**设计**：
- `ApplyCreate(kt, vendor, libraryID, accountIDs)` — 入口方法，按 vendor 分派具体实现
- `GetPolicyLibraryDetail(kt, id)` — 查询策略库详情
- `CheckAccountsBizInScope(kt, allowedBkBizIDs, accountIDs)` — 校验所有目标账号的 bk_biz_id 均在策略库允许范围内
- `CheckAccountApplied(kt, libraryID, accountID)` — 检查账号是否已应用
- `TCloudCreateCAMPolicy(kt, library, accountID)` — 调用 hc-service 创建 CAM 策略
- `TCloudCreateLocalTemplate(kt, library, accountID, cloudPolicyID)` — 创建本地 permission_template 记录
- `RecordApplyAudit(kt, libraryID, accountID)` — 记录审计（策略库 Apply + 关联账号）
- `apply.go` 和后续 `apply_update.go` 各自实现 handler，内部通过 applier 调用

**理由**：Create 和 Update 的校验逻辑互反（未应用 vs 已应用），云 API 不同（CreatePolicy vs UpdatePolicy），DB 操作不同（INSERT vs UPDATE），不适合用 action 参数做 if/switch。提取公共查询/审计方法即可，各操作保持独立。

**替代方案**：统一 `BatchApply(action)` 方法内部按 action 分支。否决原因：分支逻辑会使方法职责不清晰，且 Create/Update 的差异大于共性。

### Decision 3: hc-service 接口设计为"单账号创建 CAM 策略"

**选择**：hc-service 暴露的接口为单账号粒度——传入 `accountID`、策略名、策略内容，返回云上策略 ID（`cloud_id`）。

**理由**：每个二级账号有独立的密钥和 CAM Client 实例，批量操作本质上是循环调用。在 cloud-server 层做循环编排比在 hc-service 层做批量更灵活（可控制并发、收集逐个结果）。

### Decision 4: 审计分两层记录，并关联被应用账号

**选择**：
- 权限策略库的 `Apply` 审计：在 cloud-server 层通过 `audit.Interface.ResOperationAudit` 调用，传入 `AssociatedResType = AccountAuditResType`、`AssociatedResID = accountID`
- 权限模板的 `Create` 审计：由 data-service DAO 内置审计自动记录
- data-service 审计 build 层接收到 Apply 操作时，查询关联账号信息，将账号 ID 和名称写入审计记录的 `Detail.Data`（`AssociatedOperationAudit`）

**理由**：`permission_template` 的 BatchCreate 在 DAO 层已内置创建审计。策略库的"应用"是新的审计动作，需在 cloud-server 层显式调用。通过 `AssociatedResType/AssociatedResID` 将账号信息传递给 data-service，由 data-service 查询账号名称并构建完整的审计详情，与其他关联操作审计（如 EIP 绑定 CVM）保持一致的模式。

### Decision 5: 同步执行 + 逐个结果返回

**选择**：循环逐个调用云接口，每个账号独立成功/失败，不做事务回滚，不做异步任务。

**理由**：技术方案明确要求"同步执行，前端刷新后结果消失，不持久化"。每个账号的操作互相独立，某个账号失败不应影响其他账号。

## Risks / Trade-offs

- **[风险] 批量调用耗时**：100 个账号逐个调用 CAM API 可能耗时较长。→ 缓解：前端已知同步特性，可加 loading 提示。后续如需优化可引入并发控制。
- **[风险] 部分成功部分失败**：无法回滚已成功的账号。→ 缓解：这是设计预期行为，技术方案明确要求逐个执行并返回各自结果。
- **[权衡] hc-service 新增接口增加了调用链路**：cloud-server → hc-service → TCloud。→ 这是遵循现有架构的必要代价，保持职责分离。
