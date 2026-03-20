## ADDED Requirements

### Requirement: 批量创建权限策略库

系统 SHALL 支持通过 data-service 接口批量创建权限策略库记录，单次批量上限为 100 条。每条记录的 `policy_hash`（SHA256）由 data-service 层自动计算，`version` 初始值为 1，`tenant_id` 从请求上下文自动注入。

接口：`POST /vendors/{vendor}/permission_policy_libraries/create`

请求字段：`permission_policy_libraries`（必填，数组）；每条包含 `name`（必填）、`policy_document`（必填）、`bk_biz_ids`（可选）、`memo`（可选）；`vendor` 从 URL 路径参数读取。

#### Scenario: 成功批量创建

- **WHEN** 上层服务传入合法的批量数据（1~100 条）
- **THEN** 系统返回新记录的 id 列表，每条记录的 policy_hash = SHA256(policy_document)，version = 1

#### Scenario: 超过批量上限

- **WHEN** 请求中记录数超过 100 条
- **THEN** 系统返回 InvalidParameter 错误

#### Scenario: 缺少必填字段

- **WHEN** 某条记录缺少 name 或 policy_document
- **THEN** 系统返回 InvalidParameter 错误

---

### Requirement: 批量更新权限策略库（自动维护 hash 和 version）

系统 SHALL 支持批量更新权限策略库的 name、policy_document、bk_biz_ids、memo 字段，单次批量上限为 100 条。每条记录的 `id` 通过请求 Body 传入。当某条记录的 policy_document 发生变化时，data-service 层 SHALL 自动重新计算 policy_hash 并将 version 递增 1；若 policy_document 内容与当前一致，则 policy_hash 和 version 不变。

接口：`PATCH /vendors/{vendor}/permission_policy_libraries/batch/update`

请求字段：`permission_policy_libraries`（必填，数组）；每条包含 `id`（必填）、`name`（可选）、`policy_document`（可选）、`bk_biz_ids`（可选）、`memo`（可选）。

#### Scenario: 批量更新，其中部分记录 policy_document 变化

- **WHEN** 上层服务传入多条记录，其中部分记录的 policy_document 与当前不同
- **THEN** 系统对变化的记录更新 policy_hash = SHA256(new_policy_document)，version = old_version + 1；未变化的记录 policy_hash 和 version 保持不变

#### Scenario: 更新 policy_document（内容相同）

- **WHEN** 某条记录传入与当前完全相同的 policy_document
- **THEN** 系统不更新该记录的 policy_hash 和 version，其他字段正常更新

#### Scenario: 仅更新非 policy_document 字段

- **WHEN** 某条记录的 policy_document 为空，仅包含 name 或 memo 等字段
- **THEN** 系统正常更新对应字段，policy_hash 和 version 保持不变

#### Scenario: 超过批量上限

- **WHEN** 请求中记录数超过 100 条
- **THEN** 系统返回 InvalidParameter 错误

---

### Requirement: 批量删除权限策略库

系统 SHALL 支持通过 filter 表达式物理删除多条权限策略库记录，handler 层直接将 `req.Filter` 传给 DAO，不做额外的预查询。

接口：`DELETE /permission_policy_libraries/batch`

请求字段：`filter`（必填，filter 表达式）

#### Scenario: 成功批量删除

- **WHEN** 上层服务传入有效的 filter 表达式
- **THEN** 系统物理删除匹配记录，返回成功

#### Scenario: filter 为空

- **WHEN** 请求中 filter 未传
- **THEN** 系统返回 InvalidParameter 错误

---

### Requirement: 列表查询权限策略库

系统 SHALL 支持通过标准 filter + page 参数查询权限策略库列表，支持 count 模式和数据模式。查询结果 SHALL 自动按 tenant_id 过滤。请求中仅包含 `filter` 和 `page`，不支持 `fields` 选择。

接口：`POST /permission_policy_libraries/list`

#### Scenario: 按 vendor 过滤查询

- **WHEN** 上层服务传入 filter 包含 vendor 条件
- **THEN** 系统返回该 vendor 下的策略库列表

#### Scenario: 按 bk_biz_ids 过滤查询

- **WHEN** 上层服务传入包含 bk_biz_id 的 filter 条件
- **THEN** 系统返回 bk_biz_ids 包含该业务 ID 的策略库列表

#### Scenario: count 模式

- **WHEN** 请求 page.count = true
- **THEN** 系统返回满足条件的记录总数，不返回数据明细

---

### Requirement: SDK 客户端暴露

系统 SHALL 通过 data-service SDK 对外暴露上述接口：
- TCloud vendor client（`pkg/client/data-service/tcloud/`）通过独立 `PermissionPolicyLibraryClient` 暴露 `BatchCreate`、`BatchUpdate`，方法签名使用 `kt *kit.Kit`
- Global client（`pkg/client/data-service/global/`）通过独立 `PermissionPolicyLibraryClient` 暴露 `ListPermissionPolicyLibrary`、`BatchDelete`，方法签名使用 `kt *kit.Kit`

#### Scenario: 通过 TCloud client 调用 BatchCreate

- **WHEN** 上层服务通过 `dataCli.TCloud.PermissionPolicyLibrary.BatchCreate(kt, req)` 调用
- **THEN** 请求正确路由到 `/vendors/tcloud/permission_policy_libraries/create`

#### Scenario: 通过 Global client 调用 ListPermissionPolicyLibrary

- **WHEN** 上层服务通过 `dataCli.Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(kt, req)` 调用
- **THEN** 请求正确路由到 `/permission_policy_libraries/list`
