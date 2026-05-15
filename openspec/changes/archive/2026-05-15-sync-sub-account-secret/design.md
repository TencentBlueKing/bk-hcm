## Context

HCM 的 hc-service 提供了针对腾讯云各类资源的同步能力，通过 `res-sync/tcloud` 包下的 `Interface` 统一管理。目前三级账号（SubAccount）已有完整的同步流程（`sub_account.go`），利用 `common.Diff` 函数完成云端/DB 数据 diff 后分批执行增删改。

三级账号密钥（SubAccountSecret）存储在本地 `sub_account_secret` 表，扩展字段 `TCloudSubAccountSecretExtension` 记录了 `CloudSecretID`（即腾讯云的 AccessKeyID）、`CloudMainAccountID`、`CloudSubAccountID`。云端密钥通过 CAM API `ListAccessKeys`（按子账号 UIN 查询）获取，密钥最近使用时间通过 `GetSecurityLastUsed` 获取。

本次变更在现有 SubAccount 同步框架基础上，增加 SubAccountSecret 同步能力，逻辑结构与 SubAccount 高度一致。

## Goals / Non-Goals

**Goals:**
- 新增 `SubAccountSecret(kt, opt) (*SyncResult, error)` 到 TCloud 同步 `Interface`
- 实现完整的三级账号密钥同步逻辑：从云拉取 → DB 查询 → Diff → 批量新增/更新/删除
- 以云端数据为准，保持本地 DB 与云端一致
- 在现有 `SyncSubAccount` HTTP handler（`POST /vendors/tcloud/sub_accounts/sync`）中，完成 `SubAccount` 同步后紧接着调用 `SubAccountSecret` 同步，无需新增路由

**Non-Goals:**
- 不为密钥同步单独新增 HTTP 路由
- 不同步 `DisabledTime` 字段（该字段由本地业务逻辑管理）
- 不支持主动向云端创建或删除密钥（仅单向同步云→DB）
- 不涉及其他云厂商（仅 TCloud）

## Decisions

### 1. 同步粒度：以 AccountID 为入口，遍历所有三级账号

**决策**：`SyncSubAccountOption` 只含 `AccountID`（二级账号 ID），与 `SubAccount` 复用同一 option 结构。密钥同步在 `SyncSubAccount` handler 中 `SubAccount` 调用完成后串行触发，使用同一个 `syncCli`。

**理由**：密钥同步依赖三级账号数据（需要 UIN），将两者绑定在同一个触发入口，保证先同步账号再同步密钥的执行顺序，同时避免新增路由增加维护成本。

**替代方案**：为密钥同步单独新增 HTTP 路由——增加了调用复杂度，且调用方必须按序分两次触发，不如串联在 SubAccount 同步内更内聚。

### 2. 云端数据聚合方式

**决策**：先获取该 AccountID 下所有子账号（从 DB），再遍历每个子账号调用 `ListAccessKeys`，将结果聚合为一个平铺列表（`cloudSubAccountSecretKey` = `CloudSecretID` = AccessKeyID）。

**理由**：`ListAccessKeys` 是按 UIN 查询的，必须先知道所有子账号的 UIN。UIN 存在子账号的 extension 扩展字段中。

**LastUsedTime 获取**：`GetSecurityLastUsed` 支持批量查询（最多 10 个），在聚合所有云端密钥后，按 10 个一批并发查询并填充到密钥数据中。

### 3. Diff 键设计

**决策**：以 `CloudSecretID`（AccessKeyID）作为唯一标识进行 diff 匹配：
- 云端：每个 `AccessKeyInfo.AccessKeyID` 作为云端唯一 ID
- DB 端：`SubAccountSecret.Extension.CloudSecretID` 作为 DB 唯一 ID（通过 `GetCloudID()` 返回）

**理由**：AccessKeyID 在腾讯云全局唯一，跨子账号不会重复，适合作为 diff 键。使用已有的 `common.Diff` 泛型函数，需要对云端数据封装一个实现 `GetCloudID()` 接口的包装类型。

### 4. 更新字段范围

**决策**：更新时只同步以下字段：
- `Status`（Active/Inactive）
- `CloudCreatedAt`（密钥创建时间）
- `LastUsedTime`（最近使用时间，来自 `GetSecurityLastUsed`）
- Extension 中的 `CloudSecretID`、`CloudMainAccountID`、`CloudSubAccountID`

**不更新**：`DisabledTime`（本地业务管理字段，禁用时间由本地操作写入，不从云端覆盖）

### 5. 删除策略

**决策**：云端不存在的密钥从 DB 删除，使用 `filter.ContainersExpression("id", ids)` 按本地 ID 批量删除（与 SubAccount 删除不同，密钥没有"主账号构造"数据的特殊情况）。

**理由**：密钥是纯云端资源，云端删除后本地即可同步删除，不需要二次验证。

## Risks / Trade-offs

- **子账号 UIN 缺失风险**：若某个子账号在 DB 中没有 UIN（extension 为空或 UIN 为 0），则跳过该子账号的密钥拉取并记录警告日志，避免整体同步失败。Mitigation: 在遍历子账号时添加 UIN 为 0 的防御判断。
- **`GetSecurityLastUsed` API 限流**：官方限制 20 次/秒。该接口每批最多传 10 个密钥 ID（`GetSecurityLastUsedMaxKeys = 10`），当密钥数量较多时需要多批次调用。Mitigation：① 使用 `slice.Split` 按 10 个分批调用，减少单次调用的 key 数量；② 项目统一通过腾讯云 SDK 的 `profile.RateLimitExceededMaxRetries`（最多 6 次）+ `profile.RateLimitExceededRetryDuration`（随机 600~1000ms）处理限流重试，该机制在 `cloud_client.go` 中当 `kt.RequestSource == enumor.AsynchronousTasks` 时自动开启（`SetRateLimitRetryWithRandomInterval(true)`），同步任务场景下无需额外处理。
- **`ListAccessKeys` 调用频次**：每个子账号调用一次，频次与子账号数量正相关。同上，依赖 SDK 的限流重试机制兜底。
- **密钥数量较大时内存占用**：全量拉取后在内存中 diff。Mitigation: 当前已有 SubAccount 同步采用相同模式，密钥数量一般远小于其他资源，可接受。

## Migration Plan

1. 部署新版本 hc-service
2. 通过已有接口 `POST /vendors/tcloud/sub_accounts/sync` 触发存量同步，密钥同步将自动跟随执行
3. 回滚：回退 hc-service 版本，DB 数据无影响（仅为同步写入，不破坏现有数据）
