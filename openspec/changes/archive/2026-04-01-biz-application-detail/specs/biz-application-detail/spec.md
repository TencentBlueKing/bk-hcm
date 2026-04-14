## ADDED Requirements

### Requirement: 业务视角下查看单据明细接口

系统 SHALL 提供 `GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}` 接口，接收路径参数 `bk_biz_id`（业务 ID）和 `application_id`（单据 ID），无请求体。成功时返回单据详情，包含 `id`、`source`、`sn`、`type`、`status`、`applicant`、`content`（脱敏后）、`delivery_detail`、`memo`、`revision`、`ticket_url` 字段。该接口所需权限：业务访问权限。

#### Scenario: 成功查看单据明细
- **WHEN** 传入合法的 bk_biz_id 和 application_id，用户有业务访问权限，单据存在且归属该业务
- **THEN** 系统返回单据详情，Content 字段已脱敏

#### Scenario: 单据不存在
- **WHEN** 传入的 application_id 在数据库中不存在
- **THEN** 返回 RecordNotFound 错误，提示 "application not found"

#### Scenario: 单据不归属当前业务
- **WHEN** 单据存在但 bk_biz_id 不在单据的 bk_biz_ids 列表中
- **THEN** 返回 RecordNotFound 错误，提示 "application not found"

### Requirement: 业务访问权限鉴权

系统 SHALL 使用 `meta.Biz` + `meta.Access` 组合进行权限校验，通过 `svc.authorizer.AuthorizeWithPerm` 校验用户是否有访问该业务的权限。

#### Scenario: 无业务访问权限
- **WHEN** 用户没有访问 bk_biz_id 对应业务的权限
- **THEN** 返回 RecordNotFound 错误（统一错误策略，不泄露信息）

### Requirement: 归属校验

系统 SHALL 在获取单据详情后，校验请求的 `bk_biz_id` 是否在单据的 `bk_biz_ids` 列表中，使用 `slice.IsItemInSlice` 进行检查。

#### Scenario: bk_biz_id 在 bk_biz_ids 列表中
- **WHEN** 请求的 bk_biz_id 存在于单据的 bk_biz_ids 列表中
- **THEN** 继续执行后续流程，返回单据详情

#### Scenario: bk_biz_id 不在 bk_biz_ids 列表中
- **WHEN** 请求的 bk_biz_id 不存在于单据的 bk_biz_ids 列表中
- **THEN** 返回 RecordNotFound 错误

### Requirement: 统一 NotFound 错误策略

系统 SHALL 对以下情况统一返回 `errf.RecordNotFound` 错误，提示 "application not found"，避免泄露敏感信息：
- 用户无业务访问权限
- 单据不存在
- 单据不归属当前业务

#### Scenario: 权限不足时返回 NotFound
- **WHEN** 用户无权限访问该业务
- **THEN** 返回 RecordNotFound 而非 PermissionDenied

### Requirement: ITSM 审批链接获取

系统 SHALL 调用 `itsmCli.GetTicketResult(cts.Kit, application.SN)` 获取 ITSM 审批链接，并在响应体中返回 `ticket_url` 字段。

#### Scenario: 成功获取审批链接
- **WHEN** ITSM 接口调用成功
- **THEN** 响应体中 ticket_url 字段包含审批链接

### Requirement: Content 字段脱敏处理

系统 SHALL 对响应体中的 `content` 字段调用 `RemoveSenseField` 进行脱敏处理，移除敏感信息。

#### Scenario: Content 脱敏
- **WHEN** 返回单据详情
- **THEN** content 字段已移除敏感信息

### Requirement: 公共方法 buildApplicationGetResp

系统 SHALL 抽取 `buildApplicationGetResp` 公共方法，供 `GetApplication` 和 `GetBizApplication` 共同使用，减少代码重复。该方法负责：获取 ITSM 审批链接、构建响应体（包含 Source 字段）、对 Content 进行脱敏处理。

#### Scenario: GetApplication 复用公共方法
- **WHEN** 调用 GetApplication 接口
- **THEN** 内部调用 buildApplicationGetResp 构建响应体

#### Scenario: GetBizApplication 复用公共方法
- **WHEN** 调用 GetBizApplication 接口
- **THEN** 内部调用 buildApplicationGetResp 构建响应体
