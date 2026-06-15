## ADDED Requirements

### Requirement: Resource 接口查询策略库关联账号 ID

系统 SHALL 提供 `GET /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/account_ids` 接口，根据策略库 ID 查询 `permission_template` 表中关联的全部二级账号 ID，去重后全量返回，不分页。

访问权限要求：资源接入-云资源-云厂商配置（`meta.PermissionPolicyLibrary + meta.Find`）。

返回结构：
```json
{
  "code": 0,
  "message": "",
  "data": {
    "account_ids": ["00000001", "00000002"]
  }
}
```

#### Scenario: 正常查询策略库已关联账号

- **WHEN** 调用方携带合法 vendor 和 id，且具备查询权限
- **THEN** 系统返回该策略库在 `permission_template` 表中关联的全部去重账号 ID 列表

#### Scenario: 策略库无已关联账号时返回空列表

- **WHEN** 策略库存在但尚未有任何账号应用
- **THEN** 系统返回 `account_ids: []`

#### Scenario: 无权限时拒绝访问

- **WHEN** 调用方不具备 `meta.PermissionPolicyLibrary + meta.Find` 权限
- **THEN** 系统返回 403 权限拒绝错误

#### Scenario: vendor 参数非法时返回错误

- **WHEN** 路径中 vendor 不是合法枚举值
- **THEN** 系统返回 400 参数错误

---

### Requirement: Biz 接口查询业务下策略库关联账号 ID

系统 SHALL 提供 `GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_policy_libraries/{id}/account_ids` 接口，查询 `permission_template` 表中关联的二级账号 ID，**仅返回管理业务（bk_biz_id）等于路径参数 bk_biz_id 的账号**，去重后全量返回，不分页。

访问权限要求：业务访问（`meta.Biz + meta.Access`）。

返回结构同 Resource 接口。

#### Scenario: 正常查询业务下策略库已关联账号

- **WHEN** 调用方携带合法 bk_biz_id、vendor 和 id，且具备业务访问权限
- **THEN** 系统返回该策略库已关联且管理业务为指定 bk_biz_id 的全部去重账号 ID

#### Scenario: bk_biz_id 不在策略库关联业务列表中时返回错误

- **WHEN** 路径中 bk_biz_id 不在该策略库的 BkBizIDs 列表中
- **THEN** 系统返回 400 参数错误，提示该业务与策略库无关联

#### Scenario: 无符合业务条件的账号时返回空列表

- **WHEN** 策略库存在关联账号，但均不属于指定 bk_biz_id 管理
- **THEN** 系统返回 `account_ids: []`

#### Scenario: bk_biz_id 无业务访问权限时拒绝

- **WHEN** 调用方不具备 bk_biz_id 的业务访问权限
- **THEN** 系统返回 403 权限拒绝错误

#### Scenario: bk_biz_id 参数非法时返回错误

- **WHEN** 路径中 bk_biz_id 无法解析为 int64
- **THEN** 系统返回 400 参数错误
