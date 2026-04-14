## ADDED Requirements

### Requirement: 业务侧更新云权限模板接口

系统 SHALL 提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/update_permission_template` 接口，供业务侧用户对自定义权限模板提交更新申请单，经 ITSM 审批后执行变更。

接口 SHALL 要求以下权限：`meta.PermissionTemplate` / `meta.Update`（含 BizID）。

请求参数：
- `id`（string，必填）：目标权限模板 ID
- `policy_library_id`（string，必填）：要绑定的权限策略库 ID
- `memo`（string，可选）：更新后的模板备注

响应返回审批单据 ID：`{ "id": "<application_id>" }`

#### Scenario: 自定义模板更新申请提交成功

- **WHEN** 用户提交合法请求，模板为自定义模板（`policy_library_id=nil`, TCloud `cloud_type=1`），目标策略库存在且在 biz scope 内
- **THEN** 系统创建 ITSM 审批单，返回审批单 ID

#### Scenario: 非自定义模板且非同一策略库被拒绝

- **WHEN** 用户提交请求，目标模板 `policy_library_id` 不为 nil 且不等于目标 `policy_library_id`
- **THEN** 系统返回 `InvalidParameter` 错误，拒绝申请

#### Scenario: 模板不属于当前业务

- **WHEN** 用户提交请求，模板关联账号的 `bk_biz_id` 与路径参数 `bk_biz_id` 不匹配
- **THEN** 系统返回权限错误

#### Scenario: 策略库不在 biz scope 内

- **WHEN** 用户提交请求，目标策略库的 `bk_biz_ids` 不包含当前 `bk_biz_id`
- **THEN** 系统返回 `InvalidParameter` 错误

---

### Requirement: 更新云权限模板 Deliver 执行

审批通过后，系统 SHALL 执行以下操作（通过 `ApplyUpdateWithTmplInfo` 方法）：
1. 调用 `CheckPermTmplUpdatability` 校验模板是否可更新（自定义模板 或 已绑同一策略库）
2. 按 `policy_library_id` 获取策略库详情，取得新的 `policy_document`
3. 按 `templateID` 查询本地权限模板记录，获取 `account_id` 和 `cloud_id`（云端策略 ID）
4. 调用云端 API 更新 CAM Policy（更新 `policy_document` 和 `description`），其中 `description` 使用用户传入的 `memo`
5. 更新本地权限模板记录：`policy_document`、`policy_library_id`（从 nil 变为有值）、`policy_library_version`、`policy_library_sync_time`、`memo`
6. 写入 Update 审计记录

#### Scenario: Deliver 执行成功

- **WHEN** ITSM 审批通过，触发 Deliver
- **THEN** 云端 CAM Policy 内容更新，本地模板记录 `policy_library_id`、`policy_document`、`memo` 等字段均已更新，审计写入

#### Scenario: 云端更新失败

- **WHEN** 调用腾讯云 API 更新 CAM Policy 失败
- **THEN** Deliver 返回 `DeliverError`，本地记录不变

#### Scenario: 本地更新失败

- **WHEN** 云端 CAM Policy 已更新成功，但本地 DB 写入失败
- **THEN** Deliver 返回 `DeliverError`，并在错误信息中注明「云策略已更新，但本地模板更新失败」

---

### Requirement: ITSM 审批单渲染

系统 SHALL 为更新操作生成可读的审批单标题和表单内容。

标题格式：`申请更新云权限模板(<template_id>)`

表单 SHALL 包含：业务名称、云厂商、云账号名称、权限模版 ID、权限策略库名称、策略库 ID、策略内容。

#### Scenario: 标题渲染

- **WHEN** 创建 ITSM 审批单
- **THEN** 标题包含权限模板 ID

#### Scenario: 表单渲染

- **WHEN** 创建 ITSM 审批单
- **THEN** 表单包含业务名称、云厂商、云账号名称、权限模版 ID、权限策略库名称、策略库 ID、策略内容，方便审批人判断
