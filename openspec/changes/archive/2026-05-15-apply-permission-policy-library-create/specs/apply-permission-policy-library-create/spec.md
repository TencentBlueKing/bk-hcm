## ADDED Requirements

### Requirement: 应用权限策略库（创建）Resource 接口

系统 SHALL 在 cloud-server 层暴露 `POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply` 接口，将指定权限策略库应用到目标二级账号——为每个账号在云上创建 CAM 策略并在本地创建 `permission_template` 记录。逐个执行，返回每个账号的执行结果。

接口路径参数：`vendor`（枚举，云厂商，如 `tcloud`）、`id`（string，权限策略库 ID）。

请求字段：`account_ids`（body，string array，必填，长度限制 1~100）。

响应字段：`results`（array），每个元素含 `account_id`（string）、`status`（string，枚举：success/failed）、`reason`（string，仅 failed 时返回）。

#### Scenario: 全部成功

- **WHEN** 传入有效的策略库 ID 和 3 个未应用过该策略库的二级账号 ID
- **THEN** 系统 SHALL 逐个为每个账号调用 CAM CreatePolicy 并创建本地 `permission_template`，返回 3 个 `status: "success"` 的结果

#### Scenario: 部分账号已应用

- **WHEN** 传入的 account_ids 中某个账号已应用此策略库（`permission_template` 表中存在 `policy_library_id` 等于该策略库 ID 且 `account_id` 等于该账号的记录）
- **THEN** 该账号的结果 SHALL 为 `status: "failed", reason: "该二级账号已应用此权限策略库"`，其余账号正常处理

#### Scenario: 云 API 调用失败

- **WHEN** 某个账号的 CAM CreatePolicy API 调用失败
- **THEN** 该账号的结果 SHALL 为 `status: "failed", reason: <云API错误信息>`，其余账号正常处理

#### Scenario: account_ids 超过 100 个

- **WHEN** 请求中 account_ids 数组长度超过 100
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: account_ids 为空

- **WHEN** 请求中 account_ids 为空数组或未传
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 策略库不存在

- **WHEN** path 参数 id 对应的策略库不存在
- **THEN** 系统 SHALL 返回错误

#### Scenario: vendor 不支持

- **WHEN** path 参数 vendor 不在支持的枚举范围内
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: IAM 鉴权

系统 SHALL 对"应用权限策略库"操作执行 IAM 鉴权，鉴权资源属性为 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Apply, ResourceID: id}`。

#### Scenario: 无权限用户

- **WHEN** 用户不具备 PermissionPolicyLibrary Apply 权限
- **THEN** 系统 SHALL 返回鉴权失败错误

#### Scenario: 有权限用户

- **WHEN** 用户具备权限且请求参数合法
- **THEN** 系统 SHALL 正常执行应用流程

### Requirement: 业务范围校验

系统 SHALL 在执行应用操作前，校验所有目标账号的管理业务（`bk_biz_id`）均在策略库允许的 `bk_biz_ids` 范围内。任意账号的 `bk_biz_id` 不在范围内，整个请求 SHALL 返回 InvalidParameter 错误，不执行任何账号的应用操作。

#### Scenario: 账号业务不在范围内

- **WHEN** 某个 account_id 对应账号的 `bk_biz_id` 不在策略库的 `bk_biz_ids` 列表中
- **THEN** 系统 SHALL 返回 InvalidParameter 错误，错误信息包含越界账号的 ID 和 bk_biz_id

#### Scenario: 账号不存在

- **WHEN** account_ids 中某个 ID 在数据库中不存在，导致查询结果数量与请求数量不符
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 所有账号业务均在范围内

- **WHEN** 所有 account_id 对应账号的 `bk_biz_id` 均在策略库的 `bk_biz_ids` 中
- **THEN** 系统 SHALL 继续执行后续逐账号应用流程

### Requirement: 应用执行流程（单账号）

对于每个 account_id，系统 SHALL 按以下顺序执行：
1. 查询 `permission_template` 表，检查该账号是否已应用此策略库（`policy_library_id = 策略库ID` 且 `account_id = 当前账号`）。若已存在，标记为失败
2. 通过 hc-service client 调用 CAM CreatePolicy，传入 `account_id`、策略名（策略库 `name`）、策略内容（策略库 `policy_document`）、描述（`"Created from policy library {name} (v{version})"`）。若失败，标记为失败
3. 通过 data-service client 创建 `permission_template` 记录，字段包含：`cloud_id`（CAM 返回的策略 ID，uint64 转 string）、`name`（策略库 name）、`account_id`、`policy_library_id`（策略库 ID）、`policy_library_version`（策略库当前 version）、`policy_library_sync_time`（当前 UTC 时间，RFC3339 格式）、`policy_document`（策略库 policy_document）、`memo`（`"Applied from policy library: {name}"`）、`extension.cloud_type`（`TCloudCustomPolicy`）
4. 记录审计（策略库 Apply 审计 + 关联账号信息）

#### Scenario: 检查已应用时短路返回

- **WHEN** `permission_template` 表中已存在匹配记录
- **THEN** 系统 SHALL 直接标记该账号为失败，不调用云 API

#### Scenario: CAM 创建成功但本地创建失败

- **WHEN** CAM CreatePolicy 成功但 data-service 创建 `permission_template` 失败
- **THEN** 该账号的结果 SHALL 为 failed，reason 格式为 `"云策略已创建(id={cloudPolicyID}), 但本地模板创建失败: {err}"`

### Requirement: 审计记录

系统 SHALL 对每个成功应用的账号记录两类审计：
1. 权限策略库审计：`ResType = PermissionPolicyLibraryAuditResType`，`Action = Apply`，关联资源 `AssociatedResType = AccountAuditResType`，`AssociatedResID = accountID`
2. 权限模板创建审计：由 data-service DAO 内置审计自动记录（`Action = Create`）

#### Scenario: 应用成功时审计

- **WHEN** 某账号应用成功（CAM 创建成功 + 本地创建成功）
- **THEN** 系统 SHALL 记录权限策略库的 Apply 审计，审计 Detail 中通过 `AssociatedOperationAudit` 记录被应用账号的 ID 和名称

#### Scenario: 应用失败时不记录审计

- **WHEN** 某账号应用失败（任意步骤失败）
- **THEN** 系统 SHALL 不记录该账号的权限策略库 Apply 审计

#### Scenario: 审计错误不影响主流程

- **WHEN** 审计记录调用失败
- **THEN** 系统 SHALL 仅记录错误日志，不返回错误，该账号结果仍为 success

### Requirement: API Model 定义

系统 SHALL 在 `pkg/api/cloud-server/permission_policy_library.go` 中定义：
- `ApplyPermissionPolicyLibraryReq`：`AccountIDs`（string array，required，min=1，max=100）
- `ApplyPermissionPolicyLibraryResult`：`Results`（`[]ApplyAccountResult`）
- `ApplyAccountResult`：`AccountID`（string）、`Status`（string，枚举：success/failed）、`Reason`（string，omitempty）
- 常量 `ApplyStatusSuccess = "success"`、`ApplyStatusFailed = "failed"`

#### Scenario: 请求校验

- **WHEN** 调用 `Validate()` 方法
- **THEN** 系统 SHALL 校验 AccountIDs 非空且长度在 1~100 之间

### Requirement: 公共 applier 逻辑

系统 SHALL 在 `cmd/cloud-server/service/permission-policy-library/applier.go` 中定义 `PolicyLibraryApplier` 结构体，封装完整的应用逻辑及可复用辅助方法：
- `ApplyCreate(kt, vendor, libraryID, accountIDs)` — 入口方法，按 vendor 分派具体实现
- `GetPolicyLibraryDetail(kt, id)` — 查询策略库详情
- `CheckAccountsBizInScope(kt, allowedBkBizIDs, accountIDs)` — 校验账号业务范围
- `CheckAccountApplied(kt, libraryID, accountID)` — 检查账号是否已应用
- `TCloudCreateCAMPolicy(kt, library, accountID)` — 调用 hc-service 创建 CAM 策略
- `TCloudCreateLocalTemplate(kt, library, accountID, cloudPolicyID)` — 创建本地 permission_template 记录
- `RecordApplyAudit(kt, libraryID, accountID)` — 记录策略库 Apply 审计（含关联账号）

这些方法 SHALL 可被后续的 apply_update handler 和 biz 层 ApplicationHandler 复用。

#### Scenario: applier 被 apply handler 使用

- **WHEN** `apply.go` 中的 handler 调用 `ApplyCreate`
- **THEN** handler SHALL 通过 `NewPolicyLibraryApplier(cli, audit)` 创建实例并调用

#### Scenario: RecordApplyAudit 传递关联账号

- **WHEN** 调用 `RecordApplyAudit(kt, libraryID, accountID)`
- **THEN** 系统 SHALL 以 `AssociatedResType = AccountAuditResType`、`AssociatedResID = accountID` 调用 `audit.ResOperationAudit`
