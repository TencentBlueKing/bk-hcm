## ADDED Requirements

### Requirement: Update permission policy library via cloud-server

系统 SHALL 在 cloud-server 层暴露 `PATCH /vendors/{vendor}/permission_policy_libraries/{id}` 接口，支持单条更新权限策略库。Handler 实现 SHALL 完成 vendor 校验、路径参数 id 提取、请求解码校验、IAM 鉴权、审计记录、将单条请求适配为 DS 层批量请求、调用 DS TCloud client 并返回空 data。

接口：`PATCH /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}`

路径参数：`vendor`（string，必填，枚举值：tcloud）、`id`（string，必填，策略库 ID）。

请求字段：`name`（string，可选，max=128）、`policy_document`（string，可选）、`bk_biz_ids`（int64 array，可选）、`memo`（*string，可选，max=255）。

响应：`{ "code": 0, "message": "", "data": null }`。

#### Scenario: 成功更新

- **WHEN** 用户发送 `PATCH /api/v1/cloud/vendors/tcloud/permission_policy_libraries/{id}`，body 包含 `name`="UpdatedPolicy"
- **THEN** 系统 SHALL 返回 `{ "code": 0, "data": null }`，对应记录的 name 被更新

#### Scenario: 更新 policy_document 触发版本递增

- **WHEN** 用户传入与当前不同的 `policy_document`
- **THEN** 系统 SHALL 更新 policy_document，DS 层自动递增 version

#### Scenario: 更新 policy_document（内容相同）

- **WHEN** 用户传入与当前完全相同的 `policy_document`
- **THEN** 系统 SHALL 不递增 version，其他字段正常更新

#### Scenario: vendor 校验失败

- **WHEN** 用户传入不合法的 vendor 值（非 tcloud 等有效枚举值）
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 不支持的 vendor

- **WHEN** vendor 为有效枚举值但非 tcloud（如 aws）
- **THEN** 系统 SHALL 返回 unsupported vendor 错误

#### Scenario: id 为空

- **WHEN** 路径参数 id 为空字符串
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: IAM 鉴权基于资源类型

系统 SHALL 对 Update 操作执行 IAM 鉴权，鉴权方式为调用 `authorizer.AuthorizeWithPerm`，鉴权资源属性为 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Update}`。

#### Scenario: 无权限用户更新

- **WHEN** 用户不具备 PermissionPolicyLibrary Update 权限
- **THEN** 系统 SHALL 返回鉴权失败错误

#### Scenario: 有权限用户更新

- **WHEN** 用户具备 PermissionPolicyLibrary Update 权限，且请求参数合法
- **THEN** 系统 SHALL 正常执行更新流程

### Requirement: 更新操作审计记录

系统 SHALL 在执行 DS BatchUpdate **之前**调用 `svc.audit.ResUpdateAudit(kt, PermissionPolicyLibraryAuditResType, id, updateFields)` 记录审计。`updateFields` 通过 `converter.StructToMap(req)` 从请求体生成。

#### Scenario: 审计记录包含变更字段

- **WHEN** 用户更新 name 和 memo 字段
- **THEN** 审计记录的 `Changed` 字段 SHALL 包含 name 和 memo 的新值

#### Scenario: 审计记录失败

- **WHEN** 审计接口调用失败
- **THEN** 系统 SHALL 返回错误，不执行后续更新操作

### Requirement: 单条→批量请求适配

Handler SHALL 将 CS 层的单条更新请求包装为 DS 层的 `PermissionPolicyLibraryBatchUpdateReq`（数组长度为 1），调用 `client.DataService().TCloud.PermissionPolicyLibrary.BatchUpdate()`。

#### Scenario: 请求正确转发到 DS 层

- **WHEN** CS 层 Handler 收到合法的更新请求，id 从路径参数获取
- **THEN** Handler SHALL 构造 `PermissionPolicyLibraryBatchUpdateReq{PermissionPolicyLibraries: [{id, name, policy_document, bk_biz_ids, memo}]}`，调用 DS TCloud client

### Requirement: CS 层 API 模型定义

系统 SHALL 在 `pkg/api/cloud-server/permission_policy_library.go` 中新增 Update 请求模型：
- `PermissionPolicyLibraryUpdateReq`：`name`（omitempty, max=128）、`policy_document`（omitempty）、`bk_biz_ids`（omitempty）、`memo`（*string, omitempty, max=255）

#### Scenario: 所有字段均可选

- **WHEN** 请求 body 中仅传入 memo 字段
- **THEN** 系统 SHALL 仅更新 memo，其他字段不变

#### Scenario: 空 body

- **WHEN** 请求 body 为 `{}`
- **THEN** 系统 SHALL 校验通过（所有字段可选），执行空更新

### Requirement: 路由注册与 svc struct 扩展

系统 SHALL 在 `cmd/cloud-server/service/permission-policy-library/service.go` 中：
- 注册路由 `PATCH /vendors/{vendor}/permission_policy_libraries/{id}` → `svc.UpdatePermissionPolicyLibrary`
- svc struct 新增 `audit audit.Interface` 字段，从 `c.Audit` 初始化

#### Scenario: 路由可达

- **WHEN** 客户端发送 `PATCH /api/v1/cloud/vendors/tcloud/permission_policy_libraries/xxx`
- **THEN** 请求 SHALL 路由到 `UpdatePermissionPolicyLibrary` handler
