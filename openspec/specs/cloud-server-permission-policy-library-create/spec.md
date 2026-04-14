## ADDED Requirements

### Requirement: Create permission policy library via cloud-server

系统 SHALL 在 cloud-server 层暴露 `POST /vendors/{vendor}/permission_policy_libraries/create` 接口，支持单个创建权限策略库。Handler 实现 SHALL 完成 vendor 校验、请求解码校验、IAM 鉴权、将单个请求适配为 DS 层批量请求、调用 DS TCloud client 并返回单个 ID。

接口：`POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/create`

请求字段：`name`（string，必填）、`policy_document`（string，必填）、`bk_biz_ids`（int64 array，必填）、`memo`（string，必填）。

响应字段：`id`（string，新创建的策略库 ID）。

#### Scenario: 成功创建

- **WHEN** 用户发送 `POST /api/v1/cloud/vendors/tcloud/permission_policy_libraries/create`，body 包含 `name`="ReadOnlyPolicy"、`policy_document`=有效 JSON 字符串、`bk_biz_ids`=[2, 3]
- **THEN** 系统 SHALL 返回 `{ "code": 0, "data": { "id": "<new_id>" } }`

#### Scenario: vendor 校验失败

- **WHEN** 用户传入不合法的 vendor 值（非 tcloud 等有效枚举值）
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 缺少必填字段 name

- **WHEN** 请求 body 中未传 name 或 name 为空字符串
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 缺少必填字段 policy_document

- **WHEN** 请求 body 中未传 policy_document 或为空字符串
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 缺少必填字段 bk_biz_ids

- **WHEN** 请求 body 中未传 bk_biz_ids
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: bk_biz_ids 为空数组

- **WHEN** 请求 body 中 bk_biz_ids 传入空数组 `[]`
- **THEN** 系统 SHALL 正常创建，返回新策略库 ID

### Requirement: IAM 鉴权基于资源类型

系统 SHALL 对 Create 操作执行 IAM 鉴权，鉴权方式为直接调用 `authorizer.AuthorizeWithPerm`，鉴权资源属性为 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Create}`，不依赖 accountID。

#### Scenario: 无权限用户创建

- **WHEN** 用户不具备 PermissionPolicyLibrary Create 权限
- **THEN** 系统 SHALL 返回鉴权失败错误

#### Scenario: 有权限用户创建

- **WHEN** 用户具备 PermissionPolicyLibrary Create 权限，且请求参数合法
- **THEN** 系统 SHALL 正常执行创建流程

### Requirement: 单个→批量请求适配

Handler SHALL 将 CS 层的单个创建请求包装为 DS 层的 `PermissionPolicyLibraryBatchCreateReq`（数组长度为 1），调用 `client.DataService().TCloud.PermissionPolicyLibrary.BatchCreate()`，从返回的 `BatchCreateResult.IDs` 中取第一个元素作为响应的 `id` 字段。

#### Scenario: 请求正确转发到 DS 层

- **WHEN** CS 层 Handler 收到合法的创建请求
- **THEN** Handler SHALL 构造 `PermissionPolicyLibraryBatchCreateReq{PermissionPolicyLibraries: [{name, policy_document, bk_biz_ids, memo}]}`，调用 DS TCloud client

#### Scenario: 仅支持 tcloud vendor

- **WHEN** vendor 为 tcloud
- **THEN** 系统 SHALL 调用 TCloud client 完成创建

- **WHEN** vendor 为其他有效但未实现的枚举值
- **THEN** 系统 SHALL 返回 unsupported vendor 错误

### Requirement: CS 层 API 模型定义

系统 SHALL 在 `pkg/api/cloud-server/permission_policy_library.go` 中新增 Create 请求和响应模型：
- `PermissionPolicyLibraryCreateReq`：`name`（required, max=128）、`policy_document`（required）、`bk_biz_ids`（required）、`memo`（required, max=255）
- `PermissionPolicyLibraryCreateResult`：`id`（string）

#### Scenario: 请求模型校验

- **WHEN** 请求解码成功后调用 Validate()
- **THEN** 系统 SHALL 校验 name 非空且不超过 128 字符、policy_document 非空、bk_biz_ids 非 nil、memo 非空且不超过 255 字符
