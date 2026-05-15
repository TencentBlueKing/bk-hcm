## 1. 同步接口定义

- [x] 1.1 在 `cmd/hc-service/logics/res-sync/tcloud/client.go` 的 `Interface` 中新增 `SubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption) (*SyncResult, error)` 方法签名

## 2. 核心同步逻辑实现

- [x] 2.1 新建 `cmd/hc-service/logics/res-sync/tcloud/sub_account_secret.go`，实现主入口函数 `SubAccountSecret`：校验入参、调用从云端/DB 拉取函数、执行 diff、分别调用增删改封装函数
- [x] 2.2 实现 `listSubAccountSecretFromCloud`：从 DB 查询该 AccountID 下所有子账号（含 extension/UIN），遍历子账号调用 `cli.cloudCli.ListAccessKeys`，跳过 UIN 为 0 的子账号并记录 warning，聚合返回全量云端密钥列表
- [x] 2.3 实现 `fetchLastUsedTime`：将全量云端密钥的 AccessKeyID 按 10 个一批调用 `cli.cloudCli.GetSecurityLastUsed`，填充 `LastUsedTime` 到对应密钥
- [x] 2.4 实现 `listSubAccountSecretFromDB`：使用 `cli.dbCli.TCloud.SubAccountSecret.ListSubAccountSecretWithExtension` 分页拉取该 AccountID 下所有 DB 密钥，返回 `[]coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension]`
- [x] 2.5 定义云端密钥包装类型（实现 `GetCloudID() string` 返回 AccessKeyID），使其可传入 `common.Diff` 泛型函数
- [x] 2.6 实现 `isSubAccountSecretChange` 函数：比较 `Status`、`CloudCreatedAt`、`LastUsedTime`，返回是否需要更新（`DisabledTime` 不参与比较）
- [x] 2.7 实现 `createSubAccountSecret`：将新增密钥列表构造为 `SubAccountSecretBatchCreateReq`，调用 `cli.dbCli.TCloud.SubAccountSecret.BatchCreateSubAccountSecret` 批量写入
- [x] 2.8 实现 `updateSubAccountSecret`：将有变更的密钥构造为 `SubAccountSecretBatchUpdateReq`（仅含 `Status`、`CloudCreatedAt`、`LastUsedTime`、`Extension`，`DisabledTime` 置 nil），调用 `BatchUpdateSubAccountSecret` 批量更新
- [x] 2.9 实现 `deleteSubAccountSecret`：按本地 ID 批量删除 DB 中云端已不存在的密钥，使用 `filter.ContainersExpression("id", ids)` 构造过滤条件

## 3. HTTP Handler 调用串联

- [x] 3.1 在 `cmd/hc-service/service/sync/tcloud/sub_account.go` 的 `SyncSubAccount` handler 中，`SubAccount` 同步成功后追加调用 `syncCli.SubAccountSecret`，并在失败时记录错误日志返回错误

## 4. 验证

- [x] 4.1 确认编译通过（`var _ Interface = new(client)` 静态断言不报错）
- [x] 4.2 手动触发 `POST /vendors/tcloud/sub_accounts/sync`，确认 SubAccount 和 SubAccountSecret 均正常同步，日志输出符合预期
