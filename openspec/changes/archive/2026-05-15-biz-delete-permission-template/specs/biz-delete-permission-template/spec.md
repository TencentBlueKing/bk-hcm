## ADDED Requirements

### Requirement: 删除云权限模板申请单接口

系统 SHALL 提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/delete_permission_template` 接口（v9.9.9+），接收 `id`（权限模板ID）作为请求体，创建 ITSM 申请单。当前仅支持 vendor=tcloud。

接口 SHALL 要求以下权限：`meta.PermissionTemplate` / `meta.Delete`（含 BizID）。

#### Scenario: 创建申请单成功
- **WHEN** 发送 POST 请求，id 指向一个自定义策略且关联三级账号数为 0 的权限模板，业务 ID 归属合法
- **THEN** 系统创建 ITSM 申请单，返回 `{"id": "<application_id>"}`

#### Scenario: 不支持的 vendor
- **WHEN** vendor 不是 tcloud
- **THEN** 系统返回 InvalidParameter 错误

#### Scenario: 缺少 id 字段
- **WHEN** 请求体未提供 id
- **THEN** 系统返回 InvalidParameter 错误

### Requirement: 删除约束校验 — 仅自定义策略

系统 SHALL 在创建申请单时校验目标权限模板的 CloudType，仅允许 `TCloudCustomPolicy`（自定义策略）进行删除，预设策略（`TCloudPresetPolicy`）SHALL 被拒绝。

#### Scenario: 自定义策略可删除
- **WHEN** 目标权限模板的 extension.cloud_type == TCloudCustomPolicy（1）
- **THEN** 校验通过，继续创建申请单

#### Scenario: 预设策略禁止删除
- **WHEN** 目标权限模板的 extension.cloud_type == TCloudPresetPolicy（2）
- **THEN** 系统返回错误："只有自定义策略模板才允许删除"

### Requirement: 删除约束校验 — 关联三级账号数为 0

系统 SHALL 在创建申请单时校验目标权限模板是否被任何三级子账号引用（sub_account.permission_template_ids 包含该模板 ID），若关联账号数大于 0 则 SHALL 拒绝。

#### Scenario: 无关联子账号可删除
- **WHEN** 没有任何三级子账号的 permission_template_ids 包含目标模板 ID
- **THEN** 校验通过，继续创建申请单

#### Scenario: 存在关联子账号禁止删除
- **WHEN** 至少一个三级子账号的 permission_template_ids 包含目标模板 ID
- **THEN** 系统返回错误，提示该模板已被子账号关联，无法删除

### Requirement: 删除约束校验 — 业务归属

系统 SHALL 校验目标权限模板归属的云账号的 bk_biz_id 与请求路径中的 bk_biz_id 一致，否则拒绝操作。

#### Scenario: biz_id 一致可操作
- **WHEN** 模板关联的 account.bk_biz_id 与请求路径 bk_biz_id 相同
- **THEN** 校验通过

#### Scenario: biz_id 不一致拒绝
- **WHEN** 模板关联的 account.bk_biz_id 与请求路径 bk_biz_id 不同
- **THEN** 系统返回权限拒绝错误

### Requirement: ITSM 表单渲染

系统 SHALL 为删除申请单渲染 ITSM 标题和表单内容，包含业务名称、云厂商、云账号名称、权限模板 ID。

#### Scenario: 渲染 ITSM 标题
- **WHEN** 系统渲染申请单标题
- **THEN** 返回格式为 `"申请删除云权限模板(<templateName>)"`

#### Scenario: 渲染 ITSM 表单
- **WHEN** 系统渲染申请单表单内容
- **THEN** 包含"业务"、"云厂商"、"云账号"、"权限模板ID"、"权限模板名称"字段

### Requirement: 审批通过后删除云上 CAM Policy

系统 SHALL 在审批通过（Deliver 阶段）调用 hc-service 删除云上对应的 CAM Policy（通过模板的 cloud_id 和关联账号信息）。

#### Scenario: TCloud 删除成功
- **WHEN** ITSM 审批通过，Deliver 被触发，hc-service DeleteCAMPolicy 调用成功
- **THEN** 进入下一步删除本地记录

#### Scenario: 云端删除失败
- **WHEN** hc-service DeleteCAMPolicy 返回错误
- **THEN** Deliver 返回 DeliverError 状态，记录错误信息，本地记录不删除

### Requirement: 审批通过后删除本地权限模板记录

系统 SHALL 在云端 CAM Policy 删除成功后，先记录删除审计，再调用 data-service 删除本地 permission_template 记录。删除顺序为：① 记录审计（`ResDeleteAudit`）→ ② 调用 data-service `BatchDelete`。

#### Scenario: 本地记录删除成功
- **WHEN** 云端删除成功后，审计记录成功，data-service BatchDelete 调用成功
- **THEN** Deliver 返回 Completed 状态，deliverDetail 包含 `{"id": "<templateID>"}`

#### Scenario: 审计记录失败
- **WHEN** `ResDeleteAudit` 返回错误
- **THEN** Deliver 返回 DeliverError 状态，记录错误信息，本地记录不删除

#### Scenario: 本地删除失败
- **WHEN** data-service BatchDelete 返回错误
- **THEN** Deliver 返回 DeliverError 状态，记录错误信息

### Requirement: hc-service TCloud DeleteCAMPolicy 接口

系统 SHALL 在 hc-service 提供 `DELETE /permission_templates/cam/delete_policy` 接口，接收 `account_id` 和 `policy_ids`（云端策略 ID 列表，至少 1 个），调用腾讯云 CAM SDK 批量删除对应策略。当前在 Deliver 阶段每次只传入单个 policy_id，但接口层支持批量。

#### Scenario: 删除成功
- **WHEN** 传入合法的 account_id 和非空的 policy_ids
- **THEN** 调用 TCloud adaptor `DeletePolicy`，返回 nil error

#### Scenario: policy_ids 为空列表
- **WHEN** policy_ids 为空数组（`[]`）或未提供
- **THEN** 返回 InvalidParameter 错误（validate `required,min=1`）

### Requirement: 枚举扩展 — PermTemplateActionDelete

系统 SHALL 在 `enumor.OperatePermTemplateAction` 中新增 `PermTemplateActionDelete = "delete"` 枚举值，`Validate()` 方法 SHALL 包含该值。

#### Scenario: delete action 校验通过
- **WHEN** 调用 PermTemplateActionDelete.Validate()
- **THEN** 返回 nil
