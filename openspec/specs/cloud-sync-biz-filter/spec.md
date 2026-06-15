# cloud-sync-biz-filter Specification

## Purpose
TBD - created by archiving change cloud-sync-biz-filter. Update Purpose after archive.
## Requirements
### Requirement: 支持通过 global_config 配置同步业务白名单
系统 SHALL 在 `global_config` 表中读取 `config_type="cloud_sync"`、`config_key="sync_biz_ids"` 的配置项，其值为允许同步的业务 ID（`bk_biz_id`）白名单，**按租户维度组织**，格式为 JSON 对象：`{"tenantID": [bizID1, bizID2], ...}`，key 为租户 ID，value 为该租户允许同步的 biz ID 列表（元素类型为 int64）。

#### Scenario: 配置存在且非空（某租户有白名单）
- **WHEN** `global_config` 中存在 `cloud_sync/sync_biz_ids` 配置，且解析出的 map 中该租户对应的 biz ID 列表非空
- **THEN** 系统 SHALL 以该租户的 biz ID 列表作为同步业务白名单，仅同步 `bk_biz_id` 在列表中的账号

#### Scenario: 配置不存在，或某租户不在 map 中，或其 biz ID 列表为空
- **WHEN** `global_config` 中不存在 `cloud_sync/sync_biz_ids` 配置，或解析后的 map 为空对象，或该租户的 biz ID 列表为空
- **THEN** 系统 SHALL 对该租户的所有账号执行全量同步（与原有行为一致）

#### Scenario: 读取配置失败
- **WHEN** 调用 `GlobalConfig.List` 接口返回错误
- **THEN** 系统 SHALL 跳过本轮同步周期（`continue`），记录错误日志，等待下一轮重试

---

### Requirement: 按业务白名单过滤账号列表
当业务白名单非空时，系统 SHALL 在 Account.List 的查询条件中追加业务 ID 过滤，仅返回 `bk_biz_id` 在白名单内的账号。

#### Scenario: 白名单业务 ID 数量不超过 500
- **WHEN** 业务白名单包含 ≤500 个 ID
- **THEN** 系统 SHALL 构造单个 `AtomRule{bk_biz_id IN [全部ID]}`，追加至 AND filter 中

#### Scenario: 白名单业务 ID 数量超过 500
- **WHEN** 业务白名单包含 >500 个 ID
- **THEN** 系统 SHALL 将 ID 列表按 500 分批，每批生成一个 `AtomRule{bk_biz_id IN [batch]}`，所有批次通过 OR 表达式拼接，整体嵌套在 AND filter 中，形成：`AND[vendor=X, type=resource, OR[bk_biz_id IN b1, bk_biz_id IN b2, ...]]`

#### Scenario: 账号 bk_biz_id 不在白名单中
- **WHEN** 某账号的 `bk_biz_id` 不在业务白名单内
- **THEN** 该账号 SHALL NOT 出现在 Account.List 结果中，不触发同步

---

### Requirement: 公共资源仅同步一次
系统 SHALL 保持现有逻辑：在每个租户+vendor 的同步周期内，公共资源（`syncPublicResource`）仅随第一个被处理的账号触发一次同步，业务过滤不影响此行为。

#### Scenario: 白名单过滤后存在可同步账号
- **WHEN** 业务白名单过滤后，存在至少一个账号
- **THEN** 公共资源随第一个账号同步一次，后续账号的 `syncPublicResource` 为 false

#### Scenario: 白名单过滤后无可同步账号
- **WHEN** 业务白名单过滤后，该租户+vendor 下无匹配账号
- **THEN** 公共资源 SHALL NOT 被同步（无账号触发入口）

---

### Requirement: global_config 枚举值定义
系统 SHALL 在 `pkg/criteria/enumor/global_config.go` 中新增以下枚举值：
- `GlobalConfigTypeCloudSync GlobalConfigType = "cloud_sync"`（config_type）
- `GlobalConfigKeyCloudSyncBizIDs GlobalConfigKeyCloudSync = "sync_biz_ids"`（config_key，config_value 为 JSON 对象 `{"tenantID": [bizID1, bizID2], ...}`）

#### Scenario: 代码引用配置 key/type
- **WHEN** 同步逻辑读取 global_config 时
- **THEN** SHALL 使用上述枚举常量（而非硬编码字符串）构造查询条件

