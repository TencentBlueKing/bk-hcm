## ADDED Requirements

### Requirement: 批量创建密钥状态变更申请
系统 SHALL 提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/update_sub_account_secret_status` 端点，接受批量密钥状态变更请求，每个密钥创建独立的审批申请单。

#### Scenario: 成功创建申请单
- **WHEN** 用户提交有效的批量密钥状态变更请求（包含密钥 ID 和目标状态）
- **THEN** 系统为每个密钥创建独立的审批申请单，返回申请单 ID 列表

#### Scenario: 密钥不存在
- **WHEN** 请求中包含不存在的密钥 ID
- **THEN** 系统返回错误，提示密钥不存在

#### Scenario: 三级账号不存在
- **WHEN** 密钥对应的三级账号不存在
- **THEN** 系统返回错误，提示三级账号不存在

### Requirement: 参数校验
系统 SHALL 校验请求参数的有效性，包括密钥 ID 必填、状态值必须为 `enabled` 或 `disabled`。

#### Scenario: 无效状态值
- **WHEN** 请求中 status 既不是 `enabled` 也不是 `disabled`
- **THEN** 系统返回参数校验错误

#### Scenario: 密钥列表为空
- **WHEN** 请求中 `sub_account_secrets` 为空数组
- **THEN** 系统返回参数校验错误

### Requirement: 权限校验
系统 SHALL 校验当前用户对目标业务的访问权限。

#### Scenario: 无权限用户
- **WHEN** 用户无目标业务的访问权限
- **THEN** 系统返回权限拒绝错误

### Requirement: ITSM 工单渲染
系统 SHALL 生成包含云厂商、密钥 ID、目标状态等信息的 ITSM 审批表单。

#### Scenario: 渲染审批标题
- **WHEN** 系统生成 ITSM 工单标题
- **THEN** 标题包含云厂商名称和操作类型（启用/禁用密钥）

#### Scenario: 渲染审批表单
- **WHEN** 系统生成 ITSM 工单表单
- **THEN** 表单包含云厂商、所属二级账号、三级账号名称、密钥 ID、目标状态

### Requirement: 审批通过后交付
系统 SHALL 在审批通过后先更新云上密钥状态，再更新本地 DB 数据。

#### Scenario: TCloud 交付成功
- **WHEN** 审批通过，云上更新成功，本地更新成功
- **THEN** 申请单状态变更为 `completed`

#### Scenario: 云上更新成功但本地更新失败
- **WHEN** 审批通过，云上更新成功，但本地 DB 更新失败
- **THEN** 申请单状态变更为 `deliver_error`，记录错误日志

#### Scenario: 云上更新失败
- **WHEN** 审批通过，但云上密钥状态更新失败
- **THEN** 申请单状态变更为 `deliver_error`

### Requirement: hc-service 客户端补充
cloud-server SHALL 通过 hc-service client 的 `UpdateAccessKey` 方法调用 hc-service 更新云上密钥状态。

#### Scenario: 调用 UpdateAccessKey
- **WHEN** cloud-server 需要更新云上密钥状态
- **THEN** 通过 `client.HCService().TCloud.Account.UpdateAccessKey` 发送请求
