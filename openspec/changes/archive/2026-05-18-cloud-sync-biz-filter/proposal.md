## Why

多租户迁移是按业务维度逐步推进的，当前云资源定时同步逻辑会对所有账号发起同步，导致尚未迁移业务的账号也会被同步，造成干扰。需要在同步过程中按业务 ID 进行过滤，使迁移前后的业务互不影响。

## What Changes

- 在 `global_config` 中新增配置项，存储"按租户维度的业务 ID 白名单"（`sync_biz_ids`），`config_value` 为 JSON 对象 `{"tenantID": [bizID1, bizID2], ...}`
- 云资源定时同步逻辑读取该配置：
  - 若配置不存在或解析后为空对象，行为不变（各租户全量同步所有账号）
  - 若某租户的 biz ID 列表非空，仅同步该租户下 `bk_biz_id` 在白名单内的账号；不在 map 中的租户执行全量同步
- 账号过滤通过在 Account.List 的 filter 中追加 `bk_biz_id IN [...]` 条件实现
- 业务 ID 超过 500 条时，拆分为多个 `AtomRule` 通过 OR 拼接，防止超出 IN 子句长度限制

## Capabilities

### New Capabilities

- `cloud-sync-biz-filter`: 云资源定时同步的业务白名单过滤能力。通过 global_config 配置业务 ID 列表，同步时仅处理属于指定业务的账号，支持多租户迁移场景下的按业务粒度同步控制。

### Modified Capabilities

（无现有 spec 级行为变更）

## Impact

- `pkg/criteria/enumor/global_config.go`：新增 `GlobalConfigTypeCloudSync` 枚举类型及 `GlobalConfigKeyCloudSyncBizIDs` key 常量
- `cmd/cloud-server/service/sync/time_sync.go`：
  - `CloudResourceSync`：每轮同步开始时调用 `fetchSyncBizIDs` 获取 `tenantSyncBizIDs map[string][]int64`，遍历租户时按租户 ID 取对应的 biz ID 列表
  - `allAccountSync`：新增 `syncBizIDs []int64` 参数（由调用方传入当前租户的 biz ID 列表），按需构造业务 ID 过滤条件
  - `fetchSyncBizIDs`：返回 `map[string][]int64`（按租户维度），config_value 反序列化为 `map[string][]int64`
- 依赖 `cliSet.DataService().Global.GlobalConfig.List` 接口（已有）
- 不影响 API 接口，不影响数据库结构，仅为同步行为的运行时过滤
