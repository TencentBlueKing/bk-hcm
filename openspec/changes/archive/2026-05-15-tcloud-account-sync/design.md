## Context

hc-service 的 `res-sync/tcloud` 模块负责将腾讯云资源同步到本地 DB。目前已有三级账号（子账号 SubAccount）的同步逻辑（`SubAccount` 方法），但二级账号（资源账号 Account）的部分字段——邮箱、云上创建时间、LoginFlag、ActionFlag 以及密钥状态——未被同步，导致 DB 数据与云上现实存在差异。

当前二级账号数据来源：创建或编辑时由用户手工填写，之后不会自动从云上更新。Account 表的 extension 字段包含 `cloud_secret_id`，account_secret 表存储密钥扩展信息，两者均需要基于云上实际状态进行同步。

## Goals / Non-Goals

**Goals:**

- 新增 `Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error)` 方法到 Interface。
- 调用 TCloud CAM API 拉取二级账号的邮箱、云上创建时间、LoginFlag、ActionFlag。
- 调用 TCloud CAM `ListAccessKeys` 拉取密钥状态，按 `cloud_secret_id` 匹配，同步到 account_secret 表。
- 仅当字段确实发生变化时才触发更新（减少无效写操作）。

**Non-Goals:**

- 不负责账号的创建或删除（只做更新）。
- 不同步 SubAccount（三级账号），已有独立逻辑。
- 不修改 cloud_secret_key 等加密字段（安全边界，由用户维护）。
- 不支持批量多账号触发（调用方按账号逐个调用）。

## Decisions

### D1：使用 DescribeSubAccounts 而非 ListAccount 获取账号详情

`ListAccount` 返回主账号下所有子账号列表，但需要遍历匹配 UIN；`DescribeSubAccounts` 可以按指定 UIN 精确查询，且返回 `CreateTime`。选择 `DescribeSubAccounts`，一次调用即可获取邮箱和创建时间（`ListAccount` 也返回 Email，两者均可，最终选用 `DescribeSubAccounts` 以支持精确查询）。

### D2：LoginFlag / ActionFlag 通过 DescribeSafeAuthFlag 获取

`DescribeSafeAuthFlag` 是专用于获取子账号 MFA 设置的 API，直接按 UIN 查询，精确且无副作用。返回结果映射到 `enumor.AccountProtectionFlag` 类型，再写入 account extension。

### D3：密钥状态通过 ListAccessKeys 匹配 CloudSecretID

account_secret 表的 extension 中存有 `cloud_secret_id`，使用该值从 `ListAccessKeys` 返回列表中找到对应的密钥记录，取其 `Status`（`Active`/`Inactive`）映射到枚举值 `enumor.AccountSecretStatusNormal`/`enumor.AccountSecretStatusInvalid`。

### D4：Account extension 字段扩展方案与 UpdateMerge 问题

**问题根因**：data-service `updateAccount` 使用 `json.UpdateMerge`（底层 `gjson @join`），其规则是：source JSON 中**出现**的字段（包括 `null`）覆盖 destination；**缺席**的字段保留 destination 值。因此：

- 若给 `LoginFlag`/`ActionFlag` 加 `omitempty`：nil 指针序列化时字段缺席 → DB 旧值被保留 → 无法清空
- 若不加 `omitempty`：cloud-server 构建 `shouldUpdatedExtension` 时不设置这两个字段，也会带 `null` → 每次用户更新凭证就意外清空已同步值

**不能全量替换 extension（SyncExtension 方案被否）**：extension 中 `cloud_secret_id`、`cloud_secret_key`（加密存储）等字段是用户设置的，sync 不感知，全量替换有数据丢失风险。

**解决方案：在 `AccountUpdateReq` 中新增 `SyncExtensionPatch *json.RawMessage` 字段**，复用现有 `json.UpdateMerge` 基础设施，做精准字段 patch：

```go
// data-service AccountUpdateReq 新增字段
SyncExtensionPatch *json.RawMessage `json:"sync_extension_patch,omitempty"`
```

data-service update 处理逻辑（新增分支，与现有 `Extension` 分支互斥）：

```go
if req.SyncExtensionPatch != nil {
    dbAccount, err := getAccountFromTable(accountID, svc, cts)
    // 复用现有 UpdateMerge：patch JSON 中出现的字段（含 null）覆盖 DB，其余字段保持不变
    updatedExt, err := json.UpdateMerge(req.SyncExtensionPatch, string(dbAccount.Extension))
    account.Extension = tabletype.JsonField(updatedExt)
}
```

hc-service sync 侧在本地定义一个**不带 `omitempty`** 的 patch struct，nil 序列化为 JSON `null`，从而让 @join 正确覆盖 DB 旧值：

```go
type accountAuthFlagPatch struct {
    LoginFlag  *enumor.AccountProtectionFlag `json:"login_flag"`   // no omitempty: nil → "null"
    ActionFlag *enumor.AccountProtectionFlag `json:"action_flag"`  // no omitempty: nil → "null"
}
patchJSON, _ := json.Marshal(&accountAuthFlagPatch{LoginFlag: cloudLogin, ActionFlag: cloudAction})
rawPatch := json.RawMessage(patchJSON)
// 传入 SyncExtensionPatch，data-service 只更新 login_flag/action_flag，其余字段完全不动
```

此方案：只 patch 指定字段，`cloud_secret_key` 等其他字段安全无虞；nil 可以正确写 null；现有调用方行为完全不变。

### D5：仅对有 CloudSubAccountID 的账号执行同步

资源账号（ResourceAccount）的 extension 中有 `cloud_sub_account_id`，登记账号可能为空。若为空则跳过同步，返回空 SyncResult。

## Risks / Trade-offs

- [风险] `DescribeSafeAuthFlag` 对主账号本身的 UIN 调用可能报错 → 若 CloudSubAccountID 与 CloudMainAccountID 相同，跳过 LoginFlag/ActionFlag 的同步，仅同步邮箱和创建时间。
- [风险] `ListAccessKeys` 调用需要主账号密钥权限；若权限不足会报错 → 记录错误日志后跳过密钥状态同步，不中断整体流程。
- [已解决] `LoginFlag`/`ActionFlag` 为 nil 代表云上账号未设置保护，本地必须清空。通过新增 `AccountUpdateReq.SyncExtensionPatch` 字段，复用 UpdateMerge 做精准字段 patch，nil 序列化为 JSON null 后 @join 正确覆盖 DB，其他 extension 字段（如加密的 cloud_secret_key）完全不受影响，详见 D4。

## Migration Plan

- 不需要数据迁移，仅添加新代码文件和修改 Interface/更新结构体。
- 部署时直接上线新版 hc-service，现有账号在下次同步周期触发时自动补全数据。
- 回滚：回退 hc-service 到旧版本，新字段值会保留在 DB 但不再被自动更新（无破坏性）。

## Open Questions

- `DescribeSubAccounts` 返回的 `CreateTime` 格式是否与现有 `CloudCreatedAt` 存储格式一致？需在实现时验证。
- 同步触发频率由调用方（定时任务）决定，本次设计不限定，留给 hc-service 调度层配置。
