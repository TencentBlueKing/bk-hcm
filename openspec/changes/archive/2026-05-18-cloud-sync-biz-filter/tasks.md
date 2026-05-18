## 1. 枚举常量定义

- [x] 1.1 在 `pkg/criteria/enumor/global_config.go` 中新增 `GlobalConfigTypeCloudSync GlobalConfigType = "cloud_sync"` 枚举值
- [x] 1.2 在 `pkg/criteria/enumor/global_config.go` 中新增 `GlobalConfigKeyCloudSync` key 类型及 `GlobalConfigKeyCloudSyncBizIDs GlobalConfigKeyCloudSync = "sync_biz_ids"` 常量

## 2. 同步逻辑改造

- [x] 2.1 在 `time_sync.go` 中新增 `fetchSyncBizIDs(kt *kit.Kit, cliSet *client.ClientSet) (map[string][]int64, error)` 函数：调用 `GlobalConfig.List`，按 `config_type=cloud_sync` 且 `config_key=sync_biz_ids` 过滤；配置不存在时返回 nil；解析 `config_value` 为 `map[string][]int64`（key 为租户 ID，value 为该租户允许同步的 biz ID 列表）
- [x] 2.2 在 `CloudResourceSync` 的每轮循环中，在遍历 tenantIDs 之前调用 `fetchSyncBizIDs` 获取 `tenantSyncBizIDs`；若返回 error 则打印日志并 `continue` 跳过本轮；遍历租户时通过 `tenantSyncBizIDs[tenantID]` 取该租户的 biz ID 列表
- [x] 2.3 修改 `allAccountSync` 函数签名，新增 `syncBizIDs []int64` 参数，并更新所有调用处传入 `syncBizIDs`（当前租户对应的 biz ID 列表）
- [x] 2.4 在 `allAccountSync` 中，当 `syncBizIDs` 非空时，新增 `buildBizFilter(syncBizIDs []int64) filter.RuleFactory` 辅助函数：将 `syncBizIDs` 按 500 分批，每批生成 `tools.RuleIn("bk_biz_id", batch)`，超过一批时用 `tools.ExpressionOr` 拼接，否则直接返回单个 AtomRule
- [x] 2.5 在 `allAccountSync` 的 `listReq.Filter` 构造中，当 `syncBizIDs` 非空时将 `buildBizFilter` 的结果追加到 AND 表达式的 Rules 中
