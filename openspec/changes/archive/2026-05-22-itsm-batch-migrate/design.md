## Context

当前 `itsm.SystemMigrate` 硬编码使用单个流程模板，模板之间是增量关系（每个新模板只包含新增的流程定义），
且 ITSM `/system/migrate/` 接口不幂等，重复调用会报错。

现有基础设施：
- `global_config` 表已有成熟的 CRUD 接口（List/BatchCreate/BatchUpdate），`cloud_sync` 场景已验证可行
- `logics_admin` 已持有 `client.ClientSet`，可直接读写 `global_config`
- 不需要考虑并发场景

## Goals / Non-Goals

**Goals:**
- 支持有序、增量地注册多个 ITSM 流程模板
- 通过 `global_config` 追踪每个租户的注册进度，失败后可从断点恢复
- 新增模板时只需追加模板常量和列表条目，不改动流程逻辑

**Non-Goals:**
- 不处理并发场景
- 不提供回滚/反注册能力
- 不解决 ITSM API 调用成功但 global_config 更新失败的极端窗口期问题（概率极低，可手动修复）

## Decisions

### 1. 进度管理放在 logics_admin 而非 itsm 包

**选择**：`logics_admin.InitItsmProcess` 负责进度读写和循环注册，`itsm` 包只做纯粹的 API 调用。

**理由**：`itsm` 包是第三方 API 客户端封装，不应引入 DataService 依赖。`logics_admin` 已有 `client.ClientSet`，
天然适合做业务编排。

**替代方案**：在 `itsm` 包中注入 `GlobalConfigClient`。被否决，因为破坏了包的单一职责。

### 2. global_config 存储方式：每租户一条记录

**选择**：
- `config_type = "itsm"`
- `config_key = "itsm_migrate_version_{tenantID}"`
- `config_value` 存储最后成功注册的模板名（JSON 字符串）

**理由**：每租户独立，无竞争。虽然记录数随租户增长，但租户数量有限，不是问题。

**替代方案**：单条记录 + JSON 对象 `{"tenantA": "v2", "tenantB": "v1"}`。被否决，虽然与 `cloud_sync` 一致，
但多租户并发更新同一条记录时需要额外处理（尽管当前不需要考虑并发）。

### 3. 进度记录策略：执行后记录（思路 B）

**选择**：先调用 ITSM API，成功后再更新 global_config。

**理由**：实现简单。窗口期风险（ITSM 成功但 config 更新失败）概率极低（本地 DB 操作），
万一发生可手动修改 global_config 恢复。

### 4. itsm.Client 接口改造

**选择**：将 `SystemMigrate(kt, systemID)` 改为 `SystemMigrate(kt, systemID, templateContent string)`，
由调用方指定要注册的模板内容。同时暴露 `MigrateTemplates` 有序列表供调用方遍历。

**理由**：保持 itsm 包的通用性，模板选择逻辑由业务层控制。

## Risks / Trade-offs

- **[ITSM 调用成功但 config 更新失败]** → 手动修改 global_config 记录恢复。发生概率极低。
- **[模板顺序错误]** → `MigrateTemplates` slice 的追加顺序即执行顺序，代码 review 时需关注。
- **[已有环境首次升级]** → 需手动在 global_config 中插入已有进度记录，或在代码中兼容"无记录时假设 v1 已完成"的逻辑。
