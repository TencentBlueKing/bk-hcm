# tcloud-account-info

## Requirements

### Requirement: GetAccountInfoBySecret 返回 AppId
`GetAccountInfoBySecret` 方法调用 CAM `GetUserAppId` 接口时，MUST 同时返回 `AppId` 字段（整数类型）。`TCloudInfoBySecret` 结构体 MUST 新增 `AppId int64` 字段。

#### Scenario: 成功获取包含 AppId 的账号信息
- **WHEN** 调用 `GetAccountInfoBySecret` 且 CAM 接口正常返回
- **THEN** 返回的 `TCloudInfoBySecret` 包含 `CloudMainAccountID`（OwnerUin）、`CloudSubAccountID`（Uin）以及 `AppId`（如 `1250000000`）

#### Scenario: CAM 接口返回的 AppId 为空
- **WHEN** 调用 CAM 接口成功但 `resp.Response.AppId` 为 nil
- **THEN** 系统返回错误，提示 AppId 为空
