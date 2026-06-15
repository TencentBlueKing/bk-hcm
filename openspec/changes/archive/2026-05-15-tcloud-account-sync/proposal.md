## Why

TCloud 二级账号（资源账号）的部分关键字段——邮箱（Email）、云上创建时间（CloudCreatedAt）、登录保护设置（LoginFlag）、敏感操作保护设置（ActionFlag）以及对应密钥状态——目前未在同步流程中更新，导致 DB 中的数据与云上实际状态不一致。需要在 hc-service 的 res-sync 模块中新增 `Account` 同步方法，对齐二级账号的云上实际信息。

## What Changes

- 在 `cmd/hc-service/logics/res-sync/tcloud` 的 `Interface` 中新增 `Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error)` 方法。
- 新增 `account.go` 实现文件，包含从云上拉取账号信息、比对、更新的完整逻辑：
  - 调用 TCloud `DescribeSubAccounts` API 获取子账号的邮箱和云上创建时间。
  - 调用 TCloud `DescribeSafeAuthFlag` API 获取登录保护（LoginFlag）和敏感操作保护（ActionFlag）。
  - 调用 TCloud `ListAccessKeys` API，根据 extension 中的 `cloud_secret_id` 匹配，获取密钥启用/禁用状态。
- 将以上数据写入 `account` 表的 `email`、`cloud_created_at` 字段，以及 extension 中的 `login_flag`、`action_flag` 字段。
- 将密钥状态同步到 `account_secret` 表的 `status` 字段（`normal`/`invalid`）。

## Capabilities

### New Capabilities

- `tcloud-account-info-sync`: 同步 TCloud 二级账号的云上信息（邮箱、创建时间、LoginFlag、ActionFlag、密钥状态）到本地 DB，覆盖 account 表和 account_secret 表的更新逻辑。

### Modified Capabilities

## Impact

- `cmd/hc-service/logics/res-sync/tcloud/client.go`：Interface 新增 `Account` 方法。
- `cmd/hc-service/logics/res-sync/tcloud/account.go`：新增文件，实现同步逻辑。
- `pkg/adaptor/tcloud/interface.go`（可能）：若 `DescribeSubAccounts`、`DescribeSafeAuthFlag`、`ListAccessKeys` 尚未在 Interface 中声明则需补充，但实际这些方法已存在于 TCloudImpl。
- 调用方（定时任务或手动触发 API）可直接复用现有的 `SyncBaseParams` 模式。
- 依赖 data-service 的 `TCloud.Account.Update` 及 `TCloud.AccountSecret.BatchUpdateAccountSecret` 接口，无需新增。
