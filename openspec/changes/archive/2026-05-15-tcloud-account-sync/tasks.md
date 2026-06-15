## 1. data-service 层改造

- [x] 1.1 在 `pkg/api/data-service/cloud/account.go` 的 `AccountUpdateReq` 中新增 `SyncExtensionPatch *json.RawMessage` 字段（`omitempty`），用于 sync 场景精准 patch extension 中的指定字段
- [x] 1.2 在 `cmd/data-service/service/cloud/account/update.go` 的 `updateAccount` 泛型函数中，新增对 `req.SyncExtensionPatch != nil` 的处理分支：先读 DB 当前 extension，再调用 `json.UpdateMerge(req.SyncExtensionPatch, dbExtension)` 合并，结果赋给 `account.Extension`（与现有 `Extension` 分支互斥）

## 2. 核心同步逻辑实现

- [x] 2.1 新建 `cmd/hc-service/logics/res-sync/tcloud/account.go`，定义 `SyncAccountOption`（含 `AccountID string`）及其 `Validate` 方法
- [x] 2.2 实现 `(cli *client) Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error)` 主函数：从 DB 读取账号信息，校验 `CloudSubAccountID` 非空，调用子函数完成同步
- [x] 2.3 实现 `syncAccountBaseInfo`：调用 `DescribeSubAccounts` 获取 email 和 create_time，与 DB 对比，有变化时更新 `account.email` 和 `account.cloud_created_at`
- [x] 2.4 实现 `syncAccountAuthFlag`：调用 `DescribeSafeAuthFlag` 获取 `LoginFlag`/`ActionFlag`；若 CloudSubAccountID == CloudMainAccountID 则跳过；构造更新 SyncExtension 所需的完整 extension JSON（读 DB → 反序列化 → 更新 LoginFlag/ActionFlag → 重新序列化）
- [x] 2.5 实现 `syncAccountSecretStatus`：从 DB 读取该账号的 account_secret 列表（调用 `TCloud.AccountSecret.ListAccountSecretWithExtension`），调用 `ListAccessKeys` 拉取云上密钥状态，按 `cloud_secret_id` 匹配并更新 secret status（`Active` → `normal`，`Inactive` → `invalid`）；ListAccessKeys 失败时只记录日志，不中断整体同步

## 3. Interface 注册

- [x] 3.1 在 `cmd/hc-service/logics/res-sync/tcloud/client.go` 的 `Interface` 接口中新增 `Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error)` 方法声明
- [x] 3.2 确认 `client` 结构体实现了新增的 `Account` 方法（Go 编译期校验接口满足）

## 4. 变更检测辅助函数

- [x] 4.1 实现 `isAccountBaseInfoChange` 函数：比对 cloud DescribeSubAccounts 返回值与 DB account 的 email、cloud_created_at，返回是否需要更新
- [x] 4.2 实现 `isAccountAuthFlagChange` 函数：比对云上 LoginFlag/ActionFlag（含 nil）与 DB extension 中的值，返回是否需要更新

## 5. 调用入口接入

- [x] 5.1 在 `cmd/hc-service/service/account/` 目录下新建或更新 account 同步 handler，参考 `sub_account.go` 的 `SyncSubAccount` 实现，新增 `SyncAccount` handler（HTTP 路由 + 调用 `synccli.Account`）
- [x] 5.2 在 `cmd/hc-service/service/account/service.go` 中注册新增的路由
