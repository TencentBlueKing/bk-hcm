## ADDED Requirements

### Requirement: ApplicationOperation 枚举常量定义
系统 SHALL 在 `pkg/criteria/enumor/application.go` 中为 `ApplicationOperation` 类型定义全量常量，常量名使用 `Op` 前缀以避免与 `ApplicationType` 常量冲突。对于现有 type 对应的 operation，其字符串值 SHALL 与对应 `ApplicationType` 字符串值相同（保证向后兼容）；新增细粒度操作使用新字符串值。`ApplicationOperation` SHALL 提供 `Validate()` 方法校验合法性。

#### Scenario: 查询现有 type 对应的 operation 常量
- **WHEN** 调用 `enumor.OpAddAccount`
- **THEN** 其值等于字符串 `"add_account"`，与 `string(enumor.AddAccount)` 相同

#### Scenario: 使用新增细粒度 operation 常量
- **WHEN** 调用 `enumor.OpCreateSubAccount`
- **THEN** 其值等于字符串 `"create_sub_account"`，该值不对应任何 `ApplicationType` 常量

#### Scenario: 校验非法 operation 值
- **WHEN** 调用 `enumor.ApplicationOperation("unknown_op").Validate()`
- **THEN** 返回非 nil error

---

### Requirement: ApplicationHandler 接口扩展 GetOperation
`ApplicationHandler` 接口 SHALL 新增 `GetOperation() enumor.ApplicationOperation` 方法。`BaseApplicationHandler` SHALL 提供该方法的默认实现，返回构造时注入的 `operation` 字段值。`NewBaseApplicationHandler` 的签名 SHALL 新增 `operation enumor.ApplicationOperation` 参数。

#### Scenario: 具体 handler 返回正确的 operation
- **WHEN** 调用 `NewApplicationOfAddAccount(opt, authorizer, req)` 构造 handler，并调用 `handler.GetOperation()`
- **THEN** 返回 `enumor.OpAddAccount`

#### Scenario: 细粒度 operation 的 handler 返回新 operation
- **WHEN** 构造一个操作为 `OpCreateSubAccount` 的 handler，并调用 `handler.GetOperation()`
- **THEN** 返回 `enumor.OpCreateSubAccount`，而 `handler.GetType()` 仍返回其所属的粗粒度 `ApplicationType`

---

### Requirement: 创建申请单时写入 operation 字段
系统 SHALL 在调用 `data-service` 创建申请单时，将 `ApplicationCreateReq.Operation` 设置为 `handler.GetOperation()` 的返回值。`data-service` 侧已要求 `operation` 字段非空（required），若 handler 未提供合法 operation，创建请求 SHALL 失败并返回错误。

#### Scenario: 正常创建申请单时 operation 写库
- **WHEN** 前端发起创建 CVM 申请，后端调用对应 handler
- **THEN** 写入 `application` 表的记录中 `operation = "create_cvm"`

#### Scenario: operation 为空时创建申请单失败
- **WHEN** `ApplicationCreateReq.Operation` 为空字符串
- **THEN** data-service 侧 `InsertValidate` 返回错误，申请单不写库

---

### Requirement: bkBizIDs 记录逻辑改用 operation 判断
系统 SHALL 维护一个需要记录业务 ID 的 operation 白名单，`createApplication` 函数 SHALL 改用 `handler.GetOperation()` 的值与该白名单做匹配，而非直接对比 `applicationType`。初始白名单包含：`OpCreateCvm`、`OpCreateDisk`、`OpCreateVpc`、`OpCreateLoadBalancer`、`OpAddAccount`。

#### Scenario: 创建 CVM 申请时记录 bkBizIDs
- **WHEN** 调用 `CreateForCreateCvm` 创建申请单，handler 的 operation 为 `OpCreateCvm`
- **THEN** 写库的 `bk_biz_ids` 字段为 handler 返回的业务 ID 列表（非空）

#### Scenario: 主账号操作不记录 bkBizIDs
- **WHEN** 调用 `CreateForCreateMainAccount`，handler 的 operation 为 `OpCreateMainAccount`
- **THEN** 写库的 `bk_biz_ids` 为空列表

---

### Requirement: 查询接口返回 operation 字段
`ApplicationResp` SHALL 包含 `operation` 字段。`ListBizApplications` 及 `ListApplications` 接口返回的每条申请单记录 SHALL 携带 `operation` 字段。调用方可在请求的 `filter` 中使用 `operation` 字段进行精确过滤，与其他字段过滤方式一致，无需接口层特殊处理。

#### Scenario: 查询申请单列表时返回 operation
- **WHEN** 调用 `POST /api/v1/cloud/bizs/{bk_biz_id}/applications/list`
- **THEN** 响应中每条 `details` 记录包含 `"operation"` 字段且值非空

#### Scenario: 按 operation 字段过滤申请单
- **WHEN** 请求 filter 中包含 `{ "field": "operation", "op": "eq", "value": "create_sub_account" }`
- **THEN** 仅返回 `operation = "create_sub_account"` 的记录

---

### Requirement: 存量数据迁移
系统 SHALL 提供数据迁移 SQL，将已有 `operation` 为空的记录补填为其 `type` 字段的值，保证历史申请单数据在新接口下可正常查询和展示。迁移 SHALL 在新版本服务上线前执行。

#### Scenario: 迁移后存量记录 operation 有值
- **WHEN** 执行迁移 SQL 后，查询 `application` 表
- **THEN** 不存在 `operation = ''` 或 `operation IS NULL` 的记录

#### Scenario: 迁移后存量记录 operation 值与 type 一致
- **WHEN** 查询一条历史申请单（迁移前 operation 为空）
- **THEN** `operation` 字段值等于该记录的 `type` 字段值
