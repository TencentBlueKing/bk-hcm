## ADDED Requirements

### Requirement: SubAccountSecret 同步接口存在于 TCloud 同步 Interface

TCloud 同步 `Interface`（`cmd/hc-service/logics/res-sync/tcloud/client.go`）中 SHALL 新增 `SubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption) (*SyncResult, error)` 方法签名，并由 `client` 结构体实现。

#### Scenario: Interface 包含 SubAccountSecret 方法

- **WHEN** 编译 `cmd/hc-service/logics/res-sync/tcloud` 包
- **THEN** `Interface` 中存在 `SubAccountSecret` 方法，`client` 通过 `var _ Interface = new(client)` 静态断言满足该接口

### Requirement: SubAccountSecret 同步在 SyncSubAccount Handler 中串行触发

`SyncSubAccount` HTTP handler SHALL 在 `SubAccount` 同步完成后，紧接着调用 `SubAccountSecret` 同步；两者使用同一 `syncCli` 和同一 `AccountID`。

#### Scenario: SyncSubAccount 成功后触发密钥同步

- **WHEN** 调用 `POST /vendors/tcloud/sub_accounts/sync` 且 `SubAccount` 同步成功
- **THEN** 系统自动调用 `SubAccountSecret` 同步，并在日志中记录密钥同步的开始与结果

#### Scenario: SubAccount 同步失败时不触发密钥同步

- **WHEN** 调用 `POST /vendors/tcloud/sub_accounts/sync` 且 `SubAccount` 同步返回错误
- **THEN** 接口立即返回错误，不执行密钥同步

### Requirement: 从云端拉取所有子账号密钥

系统 SHALL 先从 DB 查询该 AccountID 下所有三级账号（含 UIN 扩展字段），再逐一调用 `ListAccessKeys` 获取各子账号的密钥列表；不存在 UIN（UIN 为 0）的子账号 SHALL 跳过并记录 warning 日志。

#### Scenario: 正常获取云端密钥列表

- **WHEN** AccountID 下存在若干子账号，每个子账号均有有效 UIN
- **THEN** 系统对每个子账号调用一次 `ListAccessKeys`，将结果聚合为全量云端密钥列表

#### Scenario: 子账号 UIN 为 0 时跳过

- **WHEN** 某子账号的 extension 中 UIN 为 0 或 extension 为 nil
- **THEN** 系统跳过该子账号的密钥拉取，记录 warning 日志，继续处理其他子账号

### Requirement: 获取密钥最近使用时间

系统 SHALL 在聚合所有云端密钥后，按每批最多 10 个 AccessKeyID 调用 `GetSecurityLastUsed`，将返回的 `LastUsedDate` 填充到对应密钥的 `LastUsedTime` 字段。

#### Scenario: 批量获取 LastUsedTime

- **WHEN** 云端密钥总数超过 10 个
- **THEN** 系统分多批（每批 ≤ 10）调用 `GetSecurityLastUsed`，所有密钥的 `LastUsedTime` 均被正确填充

#### Scenario: GetSecurityLastUsed 调用失败时返回错误

- **WHEN** `GetSecurityLastUsed` API 返回错误（非限流重试耗尽后的最终错误）
- **THEN** 同步函数返回该错误，终止本次同步

### Requirement: Diff 并执行增删改

系统 SHALL 以 `CloudSecretID`（AccessKeyID）为唯一键，对云端密钥列表与 DB 密钥列表执行 diff，得出：
- **新增列表**（云端有、DB 无）→ 批量创建
- **更新列表**（云端有、DB 有且字段有变化）→ 批量更新
- **删除列表**（云端无、DB 有）→ 批量删除

#### Scenario: 云端新增密钥

- **WHEN** 云端存在某 AccessKeyID，DB 中不存在对应的 `CloudSecretID`
- **THEN** 系统调用 data-service 的 `BatchCreateSubAccountSecret`，写入该密钥的 `AccountID`、`SubAccountID`、`Status`、`CloudCreatedAt`、`LastUsedTime` 及 Extension

#### Scenario: 云端密钥字段发生变化

- **WHEN** 云端某密钥的 `Status` 或 `CloudCreatedAt` 或 `LastUsedTime` 与 DB 中不一致
- **THEN** 系统调用 `BatchUpdateSubAccountSecret` 更新对应字段；`DisabledTime` 字段不被覆盖

#### Scenario: 云端密钥被删除

- **WHEN** DB 中存在某密钥，但云端对应 AccessKeyID 已不存在
- **THEN** 系统调用 data-service 删除接口（按本地 ID 批量删除）从 DB 中移除该密钥

#### Scenario: 云端与 DB 完全一致

- **WHEN** diff 结果为空（无新增、无更新、无删除）
- **THEN** 系统正常返回空 `SyncResult`，不执行任何写操作

### Requirement: DisabledTime 不参与同步

`DisabledTime` 字段 SHALL NOT 在同步流程中被写入或覆盖；该字段仅由本地业务操作管理。

#### Scenario: 更新操作不含 DisabledTime

- **WHEN** 系统执行密钥更新操作
- **THEN** 构造的 `SubAccountSecretUpdate` 结构体中 `DisabledTime` 字段为 nil，不发送给 data-service

### Requirement: 限流由 SDK 内置机制处理

同步过程中调用的所有腾讯云 CAM API（`ListAccessKeys`、`GetSecurityLastUsed`）遇到限流错误时，SHALL 由 TCloud SDK 的 `RateLimitExceededMaxRetries`（6 次）+ 随机间隔（600~1000ms）机制自动重试，业务代码无需额外 sleep。

#### Scenario: 异步任务触发时限流重试生效

- **WHEN** `kt.RequestSource == enumor.AsynchronousTasks` 且 CAM API 返回限流错误
- **THEN** SDK 自动重试最多 6 次，每次随机等待 600~1000ms 后重发请求
