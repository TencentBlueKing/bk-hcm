## ADDED Requirements

### Requirement: 响应模型 PermissionPolicyLibraryAccountIDsResult

系统 SHALL 在 `pkg/api/cloud-server/permission_policy_library.go` 新增 `PermissionPolicyLibraryAccountIDsResult` 结构体，包含 `AccountIDs []string` 字段（JSON tag: `account_ids`），用于返回账号ID列表。

#### Scenario: 响应结构正确
- **WHEN** 调用 `unapplied_account_ids` 接口并成功
- **THEN** 响应 data 字段结构为 `{ "account_ids": ["id1", "id2"] }`，空结果时为 `{ "account_ids": [] }`

---

### Requirement: applier 批量查询已应用账号ID

系统 SHALL 在 `PolicyLibraryApplier` 中新增私有方法 `listAllAppliedAccountIDs(kt, libraryID string) ([]string, error)`，分页扫描 `permission_template` 表（`policy_library_id = libraryID`），汇总并去重所有 `account_id`，全量返回。每页使用 `core.DefaultMaxPageLimit`（500），当返回数量 < Limit 时终止循环。

#### Scenario: 有已应用账号
- **WHEN** 传入有效的 libraryID，permission_template 表中存在关联记录
- **THEN** 返回所有已应用该策略库的 account_id 列表（已去重）

#### Scenario: 无已应用账号
- **WHEN** 传入有效的 libraryID，permission_template 表中无关联记录
- **THEN** 返回空切片，error 为 nil

#### Scenario: 超过单页上限
- **WHEN** 已应用账号数量超过 500
- **THEN** 方法通过多次分页查询完整汇总所有账号ID并返回

---

### Requirement: applier 批量查询范围内账号ID

系统 SHALL 在 `PolicyLibraryApplier` 中新增私有方法 `listAllInScopeAccountIDs(kt *kit.Kit, vendor enumor.Vendor, bizIDs []int64) ([]string, error)`，当 `bizIDs` 为空时直接返回空切片。非空时，使用 `slice.Split(bizIDs, int(core.DefaultMaxPageLimit))` 对 bizIDs 分批，每批用 `tools.ExpressionAnd(tools.RuleEqual("vendor", vendor), tools.RuleIn("bk_biz_id", batch))` 构建过滤条件，内层分页循环扫描账号表（`req.Page.Start` 递增直到返回数量 < Limit）。最终返回去重后的全量账号ID列表。

#### Scenario: bizIDs 为空
- **WHEN** bizIDs 为空切片
- **THEN** 直接返回空切片，不发起任何查询

#### Scenario: 有符合条件的账号
- **WHEN** bizIDs 非空，账号表中存在符合 vendor 和 bk_biz_id 条件的记录
- **THEN** 返回所有匹配账号的 ID 列表（已去重）

#### Scenario: 无符合条件的账号
- **WHEN** bizIDs 非空，但无账号匹配条件
- **THEN** 返回空切片，error 为 nil

---

### Requirement: applier ListUnappliedAccountIDs 入口方法

系统 SHALL 在 `PolicyLibraryApplier` 中新增公开方法 `ListUnappliedAccountIDs(kt *kit.Kit, vendor enumor.Vendor, libraryID string) ([]string, error)`，流程为：
1. 调用 `GetPolicyLibraryDetail` 获取策略库（含 BkBizIDs）
2. 调用 `listAllInScopeAccountIDs(vendor, library.BkBizIDs)` 获取候选账号ID列表
3. 调用 `listAllAppliedAccountIDs(libraryID)` 获取已应用账号ID列表
4. 调用 `slice.NotIn(appliedAccountIDs, inScopeAccountIDs)` 计算差集
5. 返回差集结果

#### Scenario: 部分账号已应用
- **WHEN** 候选账号中部分已有 permission_template 记录
- **THEN** 返回候选账号中未有 permission_template 记录的账号ID列表

#### Scenario: 所有账号均已应用
- **WHEN** 所有候选账号均有 permission_template 记录
- **THEN** 返回空切片

#### Scenario: 策略库不存在
- **WHEN** libraryID 无效，GetPolicyLibraryDetail 返回错误
- **THEN** 方法返回该错误，不继续执行

---

### Requirement: cloud-server 层"查询未应用账号ID"接口

系统 SHALL 实现 `GET /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/unapplied_account_ids` 接口（`ListPermissionPolicyLibraryUnappliedAccountIDs` handler），完成 vendor 校验、id 校验、IAM 鉴权（`meta.Find`），然后调用 `applier.ListUnappliedAccountIDs`，返回 `PermissionPolicyLibraryAccountIDsResult`。路由 SHALL 注册为 GET 方法。

#### Scenario: 正常查询
- **WHEN** vendor 合法、id 有效、用户有 Find 权限
- **THEN** 响应 `{ "code": 0, "data": { "account_ids": [...] } }`

#### Scenario: vendor 不合法
- **WHEN** 路径参数 vendor 不合法
- **THEN** 返回 InvalidParameter 错误

#### Scenario: id 为空
- **WHEN** 路径参数 id 为空字符串
- **THEN** 返回 InvalidParameter 错误

#### Scenario: IAM 鉴权失败
- **WHEN** 当前用户无 Find 权限
- **THEN** 返回 PermissionDenied 错误
