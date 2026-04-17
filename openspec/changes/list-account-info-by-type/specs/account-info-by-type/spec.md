## ADDED Requirements

### Requirement: ListAccountInfoByType 接口
系统 SHALL 提供根据资源类型批量查询二级账号元数据信息的接口，仅返回对应用户有权限的二级账号信息。

#### Scenario: 成功查询 sub_account 类型
- **WHEN** 用户传入 ids=["id1","id2"]、type="sub_account"、bk_biz_id=1、vendor="tcloud"
- **THEN** 系统先校验用户是否有 biz_id=1 的业务访问权限，再查询 sub_account 表，筛选条件为 account_id IN ids AND bk_biz_ids JSON_CONTAINS 1 AND vendor="tcloud"，通过 countPage 方式查询 data-service，返回有权限的账号信息列表

#### Scenario: 成功查询 sub_account_secret 类型
- **WHEN** 用户传入 ids=["id1"]、type="sub_account_secret"、bk_biz_id=1、vendor="tcloud"
- **THEN** 系统先校验用户是否有 biz_id=1 的业务访问权限，再查询 sub_account 表筛选满足 account_id IN ids AND bk_biz_ids JSON_CONTAINS 1 AND vendor="tcloud" 的记录，然后查询 sub_account_secret 表确认这些 sub_account 记录下是否存在密钥，仅返回同时存在 sub_account 且有密钥的账号信息列表

#### Scenario: 成功查询 permission_template 类型
- **WHEN** 用户传入 ids=["id1"]、type="permission_template"、bk_biz_id=1
- **THEN** 系统先校验用户是否有 biz_id=1 的业务访问权限，再查询 account 表，筛选条件为 id IN ids AND usage_biz_ids JSON_CONTAINS 1，校验当前业务是否属于该账号的使用业务，返回有权限的账号信息列表

#### Scenario: 无业务访问权限
- **WHEN** 用户对 bk_biz_id=1 没有业务访问权限
- **THEN** 系统返回权限拒绝错误

#### Scenario: ids 超过限制
- **WHEN** 用户传入超过 100 个 id
- **THEN** 系统返回参数校验错误

#### Scenario: 不支持的资源类型
- **WHEN** 用户传入 type="unsupported_type"
- **THEN** 系统返回参数校验错误

### Requirement: accountTypeAuthChecker 策略模式
系统 SHALL 使用可扩展的策略模式，根据资源类型调用不同的权限校验器。

#### Scenario: 新增资源类型校验器
- **WHEN** 需要支持新的资源类型
- **THEN** 只需实现 accountTypeAuthChecker 接口并在初始化时注册到 typeCheckerMap，无需修改主流程代码

### Requirement: 响应结构包含云厂商扩展字段
系统 SHALL 根据云厂商返回对应的扩展字段，扩展字段中包含 cloud_main_account_id。

#### Scenario: tcloud 扩展字段
- **WHEN** vendor="tcloud"
- **THEN** extension 字段包含 cloud_main_account_id

#### Scenario: aws 扩展字段
- **WHEN** vendor="aws"
- **THEN** extension 字段包含 cloud_account_id
