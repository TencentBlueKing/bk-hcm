## ADDED Requirements

### Requirement: TCloud CAM UpdatePolicy adaptor 封装

系统 SHALL 在 `pkg/adaptor/types/account/` 提供 `TCloudUpdatePolicyOption`（含 `Region string`、`PolicyID uint64`（required）、`PolicyDocument *string`（可选）、`Description *string`（可选））类型定义，并在 `pkg/adaptor/tcloud/cam_policy.go` 实现 `UpdatePolicy(kt, opt) error` 方法，调用 TCloud CAM SDK 的 `UpdatePolicyWithContext`。`TCloud` interface SHALL 新增 `UpdatePolicy` 方法声明。`Region` 为空时默认使用 `constant.TCloudDefaultRegion`。

#### Scenario: 更新策略成功
- **WHEN** 调用 `UpdatePolicy`，传入合法的 `PolicyID`，并提供 `PolicyDocument` 或 `Description` 中至少一个
- **THEN** 云上 CAM 策略内容被更新，方法返回 nil error

#### Scenario: PolicyID 为 0
- **WHEN** 调用 `UpdatePolicy`，`PolicyID` 为 0
- **THEN** 方法返回 InvalidParameter 错误（struct validate 检查）

---

### Requirement: hc-service 层 CAM 策略更新接口

系统 SHALL 在 hc-service 暴露 `PATCH /vendors/tcloud/permission_templates/cam/update_policy` 接口，接收 `UpdateCAMPolicyReq`（含 `AccountID string`（required）、`PolicyID uint64`（required）、`PolicyDocument *string`（可选）、`Description *string`（可选）），通过 `CloudAdaptorClient.TCloud(kt, accountID)` 获取 TCloud adaptor，调用 `UpdatePolicy`，成功时返回 nil data。hc-service client SHALL 新增 `UpdateCAMPolicy(kt, req) error` 方法。`Validate` 方法 SHALL 先进行 struct 校验，再检查 `PolicyDocument` 与 `Description` 至少有一个非 nil，否则返回错误。

#### Scenario: 更新成功
- **WHEN** 发送 PATCH 请求，accountID 有效，PolicyID 存在，PolicyDocument 或 Description 至少一个不为 nil
- **THEN** 云上策略内容被更新，响应 `{ "code": 0, "data": null }`

#### Scenario: accountID 缺失
- **WHEN** 请求中 account_id 为空
- **THEN** 返回 InvalidParameter 错误

#### Scenario: PolicyID 为 0
- **WHEN** 请求中 policy_id 为 0
- **THEN** 返回 InvalidParameter 错误

#### Scenario: PolicyDocument 和 Description 均为 nil
- **WHEN** 请求中 policy_document 和 description 均未提供
- **THEN** 返回 InvalidParameter 错误（至少提供一个更新字段）

---

### Requirement: cloud-server 层 applier 扩展

系统 SHALL 在 `PolicyLibraryApplier` 中新增以下方法：

1. `GetAccountTemplate(kt, libraryID, accountID)` — 查询指定账号在策略库下的已有模板记录，返回 `*corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension]`（nil 表示未应用）
2. `TCloudUpdateCAMPolicy(kt, library, accountID, cloudPolicyID uint64)` — 调用 hc-service 更新 CAM 策略，传入 `PolicyDocument`（取 `library.PolicyDocument`）和 `Description`（取 `library.Memo`）
3. `TCloudUpdateLocalTemplate(kt, library, templateID string)` — 调用 data-service `BatchUpdate` 更新模板的 `PolicyDocument`、`PolicyLibraryVersion`、`PolicyLibrarySyncTime`、`Memo`
4. `ApplyUpdate(kt, vendor, libraryID, accountIDs)` — 入口方法，调用 `GetPolicyLibraryDetail`、`CheckAccountsBizInScope`，按 vendor 分派
5. `tcloudApplyUpdateForAccount` — 逐账号执行：`GetAccountTemplate`（nil → failed）→ 解析 CloudID → `TCloudUpdateCAMPolicy` → `TCloudUpdateLocalTemplate` → `RecordApplyAudit`

#### Scenario: 账号未应用
- **WHEN** `GetAccountTemplate` 返回 nil
- **THEN** 该账号结果为 `failed: "该二级账号未应用此权限策略库"`

#### Scenario: CloudID 解析失败
- **WHEN** 模板的 cloud_id 无法解析为 uint64
- **THEN** 该账号结果为 `failed: <解析错误原因>`

#### Scenario: 云 API 更新失败
- **WHEN** `TCloudUpdateCAMPolicy` 返回错误
- **THEN** 该账号结果为 `failed: <错误原因>`，不影响其他账号

#### Scenario: 本地模板更新失败
- **WHEN** `TCloudUpdateLocalTemplate` 返回错误
- **THEN** 该账号结果为 `failed: <错误原因>`（注意：云上策略已更新但本地未同步）

#### Scenario: 更新成功
- **WHEN** 所有步骤均成功
- **THEN** 该账号结果为 `success`，本地模板的 `policy_document`、`policy_library_version`、`policy_library_sync_time`、`memo` 均已更新为策略库当前值

---

### Requirement: cloud-server 层"应用权限策略库（更新）"接口

系统 SHALL 实现 `PUT /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply` 接口（`ApplyPermissionPolicyLibraryUpdate` handler），复用 `ApplyPermissionPolicyLibraryReq`（`account_ids` 字符串数组，长度限制 100）和 `ApplyPermissionPolicyLibraryResult`（`results` 数组含 `account_id`、`status`、`reason`）API 模型。Handler SHALL 完成 vendor 校验、id 校验、请求解码、IAM 鉴权（`meta.Apply`），然后调用 `applier.ApplyUpdate`。接口 SHALL 同步执行，逐账号返回结果。

#### Scenario: 全部账号更新成功
- **WHEN** PUT 请求中所有 account_id 均已应用该策略库
- **THEN** 响应 results 中每条记录 status 均为 "success"

#### Scenario: 部分账号未应用
- **WHEN** 部分 account_id 未应用该策略库
- **THEN** 未应用的账号 status 为 "failed"，reason 为 "该二级账号未应用此权限策略库"，已应用的账号正常更新

#### Scenario: account_ids 为空
- **WHEN** 请求中 account_ids 为空数组
- **THEN** 返回 InvalidParameter 错误

#### Scenario: account_ids 超过 100
- **WHEN** 请求中 account_ids 长度超过 100
- **THEN** 返回 InvalidParameter 错误

#### Scenario: vendor 不合法
- **WHEN** 路径参数 vendor 不合法
- **THEN** 返回 InvalidParameter 错误

#### Scenario: IAM 鉴权失败
- **WHEN** 当前用户无 Apply 权限
- **THEN** 返回 403 鉴权错误
