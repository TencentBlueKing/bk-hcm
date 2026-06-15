## ADDED Requirements

### Requirement: 业务视角单据明细查看
系统 SHALL 支持通过业务访问权限查看该业务下的单据明细。用户需具备目标业务的访问权限，且单据必须归属于该业务。

#### Scenario: 成功查看单据明细
- **GIVEN** 用户拥有业务 `bk_biz_id=123` 的访问权限
- **AND** 单据 `application_id=abc` 的 `bk_biz_ids` 包含 `123`
- **WHEN** 向 `GET /api/v1/cloud/bizs/123/applications/abc` 发送请求
- **THEN** 系统返回单据明细（`ApplicationGetResp` 结构），敏感字段已脱敏

#### Scenario: ITSM来源单据包含审批链接
- **GIVEN** 用户拥有业务访问权限，单据归属该业务
- **AND** 单据来源为 `enumor.Itsm`
- **WHEN** 向业务视角单据接口发送请求
- **THEN** 返回体包含 `TicketUrl` 字段（ITSM审批链接）

#### Scenario: 无业务访问权限
- **GIVEN** 用户没有业务 `bk_biz_id=123` 的访问权限
- **WHEN** 向 `GET /api/v1/cloud/bizs/123/applications/abc` 发送请求
- **THEN** 系统返回 `NotFound` 错误，内部日志记录鉴权失败原因

#### Scenario: 单据不归属该业务
- **GIVEN** 用户拥有业务 `bk_biz_id=123` 的访问权限
- **AND** 单据 `application_id=abc` 的 `bk_biz_ids` 为 `[456, 789]`（不包含 `123`）
- **WHEN** 向业务视角单据接口发送请求
- **THEN** 系统返回 `NotFound` 错误，内部日志记录归属不匹配原因

#### Scenario: 单据不存在
- **GIVEN** 用户拥有业务访问权限
- **AND** 单据 `application_id=abc` 不存在
- **WHEN** 向业务视角单据接口发送请求
- **THEN** 系统返回 `NotFound` 错误
