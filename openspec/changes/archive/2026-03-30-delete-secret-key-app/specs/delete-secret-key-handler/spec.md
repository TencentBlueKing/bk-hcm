## ADDED Requirements

### Requirement: 批量创建删除密钥申请
系统 SHALL 提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/delete_sub_account_secret` 端点，接受密钥 ID 列表，每个密钥创建独立的审批申请单。

#### Scenario: 成功创建申请单
- **WHEN** 用户提交有效的密钥 ID 列表
- **THEN** 系统为每个密钥创建独立的审批申请单，返回申请单 ID 列表

#### Scenario: 密钥不存在
- **WHEN** 请求中包含不存在的密钥 ID
- **THEN** 系统返回错误，提示密钥不存在

#### Scenario: 三级账号不存在
- **WHEN** 密钥对应的三级账号不存在
- **THEN** 系统返回错误，提示三级账号不存在

### Requirement: 参数校验
系统 SHALL 校验请求参数的有效性，包括密钥 ID 列表非空、长度不超过 100。

#### Scenario: 密钥列表为空
- **WHEN** 请求中 `ids` 为空数组
- **THEN** 系统返回参数校验错误

### Requirement: 权限校验
系统 SHALL 校验当前用户对目标业务的访问权限。

#### Scenario: 无权限用户
- **WHEN** 用户无目标业务的访问权限
- **THEN** 系统返回权限拒绝错误

### Requirement: ITSM 工单渲染
系统 SHALL 生成包含云厂商、密钥 ID 等信息的 ITSM 审批表单。

#### Scenario: 渲染审批标题
- **WHEN** 系统生成 ITSM 工单标题
- **THEN** 标题包含云厂商名称和"删除密钥"操作描述

#### Scenario: 渲染审批表单
- **WHEN** 系统生成 ITSM 工单表单
- **THEN** 表单包含云厂商、所属二级账号、三级账号名称、密钥云上 ID

### Requirement: 审批通过后交付
系统 SHALL 在审批通过后先删除云上密钥，再删除本地 DB 密钥记录。

#### Scenario: TCloud 交付成功
- **WHEN** 审批通过，云上删除成功，本地删除成功
- **THEN** 申请单状态变更为 `completed`

#### Scenario: 云上删除成功但本地删除失败
- **WHEN** 审批通过，云上删除成功，但本地 DB 删除失败
- **THEN** 申请单状态变更为 `deliver_error`，记录错误日志

#### Scenario: 云上删除失败
- **WHEN** 审批通过，但云上密钥删除失败
- **THEN** 申请单状态变更为 `deliver_error`

### Requirement: hc-service 客户端补充
cloud-server SHALL 通过 hc-service client 的 `DeleteAccessKey` 方法调用 hc-service 删除云上密钥。

#### Scenario: 调用 DeleteAccessKey
- **WHEN** cloud-server 需要删除云上密钥
- **THEN** 通过 `client.HCService().TCloud.Account.DeleteAccessKey` 发送请求
