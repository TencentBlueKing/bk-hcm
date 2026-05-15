## ADDED Requirements

### Requirement: Adaptor 支持获取子用户绑定的策略列表

系统 SHALL 通过 TCloud CAM adaptor 提供 `ListAttachedUserAllPolicies` 方法，获取指定子用户绑定的所有策略列表。

#### Scenario: 成功获取子用户绑定的策略列表
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，传入有效的子用户名（`TargetUin`）
- **THEN** 系统返回该子用户绑定的所有策略列表，包括策略 ID、策略名称、策略类型（预设/自定义）、附加时间等信息

#### Scenario: 子用户未绑定任何策略
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，传入未绑定任何策略的子用户名
- **THEN** 系统返回空列表，不报错

### Requirement: 支持分页拉取策略列表

系统 SHALL 支持通过分页参数（`Page`、`Rp`）分批拉取子用户绑定的策略列表，避免单次请求数据量过大。

#### Scenario: 使用默认分页参数
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，不指定分页参数
- **THEN** 系统使用默认分页参数（Page=1, Rp=20）返回结果

#### Scenario: 自定义分页参数拉取大量数据
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，指定 `Page=1, Rp=100`
- **THEN** 系统返回指定分页大小的策略列表

### Requirement: 支持限流重试机制

系统 SHALL 继承 adaptor 已有的限流重试机制，当 CAM API 返回限流错误时，自动进行指数退避重试。

#### Scenario: API 限流时自动重试
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，CAM API 返回限流错误（如 `RequestLimitExceeded`）
- **THEN** 系统自动进行重试，使用随机间隔避免惊群效应，最大重试次数与 adaptor 全局配置一致

#### Scenario: 重试次数耗尽后返回错误
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法，CAM API 持续返回限流错误，重试次数耗尽
- **THEN** 系统返回错误，包含原始 API 错误信息
