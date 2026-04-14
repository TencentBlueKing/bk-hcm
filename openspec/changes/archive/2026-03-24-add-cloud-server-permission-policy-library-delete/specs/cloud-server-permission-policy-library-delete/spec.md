## ADDED Requirements

### Requirement: 删除权限策略库接口

系统 SHALL 提供 `DELETE /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}` 接口，接收路径参数 `vendor`（枚举值：tcloud）和 `id`（策略库 ID），无请求体。成功时返回 `{ "code": 0, "message": "", "data": null }`。该接口所需权限：资源接入-云资源-云厂商配置。

#### Scenario: 成功删除策略库
- **WHEN** 传入合法的 vendor 和 id，记录存在且 vendor 匹配，用户有 Delete 权限
- **THEN** 系统物理删除该记录，返回 code=0

#### Scenario: id 对应的记录不存在
- **WHEN** 传入的 id 在数据库中不存在
- **THEN** 返回 RecordNotFound 错误

#### Scenario: vendor 不匹配
- **WHEN** 传入的 vendor 与记录实际 vendor 不一致
- **THEN** 返回 InvalidParameter 错误，提示 vendor 不匹配

### Requirement: vendor 参数校验

系统 SHALL 校验路径参数 `vendor` 为合法的云厂商枚举值（当前仅支持 tcloud）。校验失败 SHALL 返回 InvalidParameter 错误。

#### Scenario: vendor 为非法值
- **WHEN** vendor 传入 `"invalid_vendor"`
- **THEN** 返回 InvalidParameter 错误

#### Scenario: id 为空
- **WHEN** id 路径参数为空字符串
- **THEN** 返回 InvalidParameter 错误，提示 id is required

### Requirement: 删除前 vendor 匹配校验

系统 SHALL 在删除前通过 `Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary` 查询记录，校验记录的 vendor 字段与路径参数 vendor 一致。

#### Scenario: 查询并校验 vendor
- **WHEN** 查询到记录且 vendor 匹配
- **THEN** 继续执行后续流程

#### Scenario: 查询到记录但 vendor 不匹配
- **WHEN** 查询到记录但 record.Vendor != 路径参数 vendor
- **THEN** 返回 InvalidParameter 错误

### Requirement: IAM 鉴权

系统 SHALL 在删除前执行 IAM 鉴权，使用 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Delete}`，通过 `svc.authorizer.AuthorizeWithPerm` 校验。

#### Scenario: 无权限
- **WHEN** 用户没有 PermissionPolicyLibrary Delete 权限
- **THEN** 返回权限拒绝错误

### Requirement: 删除操作审计记录

系统 SHALL 在执行实际删除**之前**调用 `svc.audit.ResDeleteAudit(kt, PermissionPolicyLibraryAuditResType, []string{id})` 记录审计。

#### Scenario: 审计记录包含资源 ID
- **WHEN** 执行删除流程
- **THEN** 审计日志中包含被删除资源的 ID

### Requirement: 云权限模板关联检查预留

系统 SHALL 在 vendor 校验之后、审计之前预留云权限模板关联检查位，当前以 TODO 注释标记，后续实现。

#### Scenario: 当前无关联检查
- **WHEN** 云权限模板功能未实现
- **THEN** 检查位直接通过，不阻止删除
