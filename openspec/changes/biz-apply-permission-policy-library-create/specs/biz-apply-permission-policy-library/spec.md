## ADDED Requirements

### Requirement: 业务侧通过 ITSM 审批流程批量为二级账号创建权限策略库应用

系统 SHALL 提供业务接口，允许业务运营人员为指定业务下的二级账号批量提交「应用权限策略库（创建）」的审批申请。
每个账号创建一个独立的 ITSM 审批单，审批通过后系统自动在云上创建 CAM 策略并本地记录 `permission_template`。

#### Scenario: 成功为多个账号创建审批单

- **WHEN** 用户调用 `POST /bizs/{bk_biz_id}/vendors/tcloud/applications/types/apply_permission_policy_library_create`，携带有效的 `policy_library_id` 和两个属于该业务的 `account_ids`
- **THEN** 系统为每个账号各创建一条 ITSM 审批单和 application 记录，返回 `{ids: ["id1", "id2"]}`，状态均为 `pending`

#### Scenario: account_ids 中某账号不属于当前 bk_biz_id

- **WHEN** 用户传入的某个账号的 `bk_biz_id` 与路径中的 `bk_biz_id` 不一致，或账号的管理业务不在权限策略库的 `bk_biz_ids` 范围内
- **THEN** 系统返回 `InvalidParameter` 错误，不创建任何审批单

#### Scenario: account_ids 为空

- **WHEN** 用户传入空的 `account_ids` 数组
- **THEN** 系统返回 `InvalidParameter` 错误

#### Scenario: account_ids 超过100个

- **WHEN** 用户传入超过 100 个账号 ID
- **THEN** 系统返回 `InvalidParameter` 错误

#### Scenario: 用户没有 PermissionPolicyLibrary Apply 权限

- **WHEN** 用户调用该接口但不具备 `meta.PermissionPolicyLibrary + meta.Apply` 鉴权
- **THEN** 系统返回 `PermissionDenied` 错误

#### Scenario: 用户没有业务访问权限

- **WHEN** 用户对路径中的 `bk_biz_id` 没有访问权限
- **THEN** 系统返回 `PermissionDenied` 错误

---

### Requirement: 审批通过后自动执行 CAM 策略创建并记录 permission_template

系统 SHALL 在 ITSM 审批回调通过时，自动为对应账号在腾讯云上创建 CAM 自定义策略，并在本地创建 `permission_template` 记录与该权限策略库关联。

#### Scenario: 审批通过，账号未被应用过该库

- **WHEN** ITSM 回调审批通过，且该账号尚未应用该权限策略库
- **THEN** 系统调用 `PolicyLibraryApplier.ApplyCreate`，创建云上 CAM 策略，创建本地 `permission_template` 记录，记录应用审计日志，单据状态更新为 `completed`，交付详情包含 `policy_library_id` 和 `account_id`

#### Scenario: 审批通过，但审批期间账号已被应用

- **WHEN** ITSM 回调审批通过，但该账号在审批期间已被其他操作应用了同一权限策略库
- **THEN** `ApplyCreate` 返回 `ApplyStatusFailed`，单据状态更新为 `deliver_error`，交付详情记录失败原因

#### Scenario: 审批通过，但权限策略库在审批期间已被删除

- **WHEN** ITSM 回调审批通过，但 `policy_library_id` 对应的策略库已不存在
- **THEN** `ApplyCreate` 返回错误，单据状态更新为 `deliver_error`，交付详情记录错误原因
