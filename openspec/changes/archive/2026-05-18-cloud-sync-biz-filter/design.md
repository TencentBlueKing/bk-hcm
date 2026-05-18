## Context

云资源定时同步（`CloudResourceSync`）当前按租户维度遍历所有账号并触发同步。在多租户迁移场景中，迁移是按业务（`bk_biz_id`）维度分批推进的，已迁移和未迁移的业务账号混合在同一系统中，需要一种轻量的运行时控制机制，仅对已迁移业务的账号发起同步。

账号表（`account`）上有 `bk_biz_id` 字段，直接记录账号所属业务 ID，无需跨表 JOIN。`global_config` 表提供 key-value 形式的全局配置能力，已有成熟的读写接口。

## Goals / Non-Goals

**Goals:**
- 通过 `global_config` 配置业务 ID 白名单，在同步时按 `bk_biz_id` 过滤账号
- 配置为空时保持全量同步的原有行为（向后兼容）
- 对 IN 子句长度做防御性处理（≤500 个 ID 一批，超出部分用 OR 拼接）

**Non-Goals:**
- 不修改 API 接口或数据库结构
- 不引入新的外部依赖或新的数据服务客户端
- 不对 `account_biz_rel`（多对多业务关联表）做处理，只使用账号主表的 `bk_biz_id`

## Decisions

### D1：使用 global_config 存储白名单，而非配置文件

**选择**：在 `global_config` 表中存储业务 ID 列表，`config_type = "cloud_sync"`，`config_key = "sync_biz_ids"`，`config_value` 为 JSON 对象 `{"tenantID": [bizID1, bizID2], ...}`，按租户维度分别配置允许同步的 biz ID 列表。

**理由**：运营期间可动态调整（无需重启服务），与现有配置管理体系一致，无需引入新机制；按租户维度组织便于多租户迁移场景下对不同租户独立控制。

**备选方案**：yaml/env 配置文件 — 需重启服务生效，运维成本较高，放弃。

---

### D2：使用账号主表 bk_biz_id 字段，而非 account_biz_rel 关联表

**选择**：在 Account.List 的 filter 中直接追加 `bk_biz_id IN [...]`。

**理由**：账号主表已有 `bk_biz_id` 字段，单表查询，无需额外的 JOIN 或二次查询。`account_biz_rel` 是多对多关系，引入会使同步逻辑复杂化。

**备选方案**：查 `account_biz_rel` 获取 account_id 集合再过滤 — 需要额外客户端，逻辑复杂，放弃。

---

### D3：IN 子句超 500 时使用 OR 拼接多个 AtomRule

**选择**：将 `syncBizIDs` 按 500 分批，生成 `bk_biz_id IN [batch1] OR bk_biz_id IN [batch2] OR ...` 嵌套在 AND 表达式内，作为单次查询的复合 filter 条件。

**理由**：单次 API 调用，不改变分页逻辑；`tools.RuleIn` + `ExpressionOr` 可直接组合，无额外实现成本。

**备选方案**：多次 API 调用（每批一次）— 会引发 `syncPublicResource` 多批管理复杂性，放弃。

---

### D4：每轮同步读取一次配置，不缓存

**选择**：在 `CloudResourceSync` 的每次轮询循环中，重新读取 `global_config` 中的业务白名单。

**理由**：运营人员可能随时调整白名单，每次读取确保及时生效，读取开销极小（一次 List 查询）。

## Risks / Trade-offs

- **[风险] bk_biz_id=0 的账号在白名单模式下不会被同步** → 迁移期间不影响（这类账号不属于任何迁移业务）；迁移结束后清空白名单即可恢复全量同步。
- **[风险] global_config 读取失败** → 同步循环以 `continue` 跳过本轮，并打印错误日志；下一轮自动重试，不影响服务稳定性。
- **[Trade-off] OR 拼接多批次会使 SQL WHERE 子句变长** → 业务 ID 通常不超过数千个，生成的 SQL 远低于 MySQL 默认 `max_allowed_packet`（4MB），风险可控。

## Migration Plan

1. 部署新版本服务（行为与当前一致，白名单为空时全量同步）
2. 通过 `global_config` 写入迁移批次的业务 ID 列表（`config_type=cloud_sync, config_key=sync_biz_ids`）
3. 验证同步日志，确认仅对指定业务账号触发同步
4. 迁移完成后，清空或删除 `sync_biz_ids` 配置，恢复全量同步

**回滚**：删除 `global_config` 中的 `sync_biz_ids` 记录即可，服务无需重启。

## Open Questions

（无）
