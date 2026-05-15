## ADDED Requirements

### Requirement: ListBizAccountByResType 接口
系统 SHALL 提供根据资源类型批量查询二级账号元数据信息的业务级接口，仅返回对应用户有权限的二级账号信息。

#### 接口定义
- **路径**: `POST /bizs/{bk_biz_id}/vendors/{vendor}/accounts/list/by/res_type`
- **函数名**: `ListBizAccountByResType`

#### 请求参数
- **bk_biz_id**: 路径参数，业务ID (int64)
- **vendor**: 路径参数，云厂商 (tcloud/aws/gcp/azure/huawei)
- **ids**: 请求体，账号ID列表 (string数组，必填，1-100个)
- **res_type**: 请求体，资源类型 (meta.ResourceType，必填)

#### 支持的资源类型
- `sub_account`: 三级账号，需校验 sub_account 表的 account_id 和 bk_biz_ids
- `sub_account_secret`: 三级账号密钥，需校验 sub_account 存在且 sub_account_secret 存在密钥
- `permission_template`: 权限模板，需校验 account 表的 usage_biz_ids 包含当前业务

#### Scenario: 成功查询 sub_account 类型
- **WHEN** 用户传入 ids=["id1","id2"]、res_type="sub_account"、bk_biz_id=1、vendor="tcloud"
- **THEN** 系统先校验用户是否有 biz_id=1 的业务访问权限，再查询 sub_account 表，筛选条件为 account_id IN ids AND bk_biz_ids JSON_CONTAINS 1 AND vendor="tcloud"，通过分页查询 data-service，返回去重后的有权限的账号ID列表，再批量查询账号详情返回

#### Scenario: 成功查询 sub_account_secret 类型
- **WHEN** 用户传入 ids=["id1"]、res_type="sub_account_secret"、bk_biz_id=1、vendor="tcloud"
- **THEN** 系统先校验用户业务访问权限，然后先查询 sub_account 表筛选满足 account_id IN ids AND bk_biz_ids JSON_CONTAINS 1 AND vendor="tcloud" 的记录，再查询 sub_account_secret 表确认这些 sub_account 记录下是否存在密钥，仅返回同时存在 sub_account 且有密钥的账号信息列表

#### Scenario: 成功查询 permission_template 类型
- **WHEN** 用户传入 ids=["id1"]、res_type="permission_template"、bk_biz_id=1、vendor="tcloud"
- **THEN** 系统先校验用户业务访问权限，再查询 account 表，筛选条件为 id IN ids AND vendor="tcloud"，然后校验当前业务是否属于该账号的使用业务(usage_biz_ids JSON_CONTAINS)，返回匹配的账号信息列表

#### Scenario: 无业务访问权限
- **WHEN** 用户对 bk_biz_id=1 没有业务访问权限
- **THEN** 系统返回 `errf.PermissionDenied` 权限拒绝错误

#### Scenario: ids 超过限制
- **WHEN** 用户传入超过 100 个 id
- **THEN** 请求参数校验器返回校验错误

#### Scenario: 不支持的资源类型
- **WHEN** 用户传入不支持的 res_type
- **THEN** 系统返回参数校验错误：`the checker not support res_type: {res_type}`

#### Scenario: 无权限的账号返回空列表
- **WHEN** 用户传入的账号ID全部没有权限
- **THEN** 系统返回空的 details 数组

### Requirement: accountResTypeAuthChecker 策略模式接口
系统 SHALL 使用可扩展的策略模式，根据资源类型调用不同的权限校验器。

#### 接口定义
```go
type accountResTypeAuthChecker interface {
    filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64, vendor enumor.Vendor) ([]string, error)
}
```

#### Scenario: 创建校验器
- **WHEN** 调用 `newAuthChecker(client, resType)`
- **THEN** 根据 res_type 返回对应的校验器实现
- **res_type=sub_account**: 返回 `subAccountAuthChecker`
- **res_type=sub_account_secret**: 返回 `subAccountSecretAuthChecker`
- **res_type=permission_template**: 返回 `permissionTemplateAuthChecker`

#### Scenario: 新增资源类型校验器
- **WHEN** 需要支持新的资源类型
- **THEN** 只需实现 `accountResTypeAuthChecker` 接口并在 `newAuthChecker` 中注册，无需修改主流程代码

### Requirement: 响应结构包含云厂商扩展字段
系统 SHALL 根据云厂商返回对应的扩展字段，扩展字段中包含云厂商特定的账号信息。

#### 响应结构
```go
type AccountInfoByTypeDetail struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    BkBizID     int64                  `json:"bk_biz_id"`
    UsageBizIDs []int64                `json:"usage_biz_ids"`
    Managers    []string               `json:"managers"`
    Extension   map[string]interface{} `json:"extension"`
}
```

#### Scenario: tcloud 扩展字段
- **WHEN** vendor="tcloud"
- **THEN** extension 字段包含 cloud_main_account_id、cloud_sub_account_id、root_account_id 等腾讯云特定字段

#### Scenario: aws 扩展字段
- **WHEN** vendor="aws"
- **THEN** extension 字段包含 cloud_account_id、cloud_iam_username、iam_user_type 等AWS特定字段