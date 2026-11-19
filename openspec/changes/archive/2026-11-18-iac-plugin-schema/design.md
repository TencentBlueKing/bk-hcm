# Design

## Context

背景与动机见 `proposal.md - Why`，行为契约见 `specs/iac-plugin-schema/spec.md`。

在海垒（HCM）现有的微服务分层架构中：
- `data-service` 负责数据库表（MySQL）的统一访问控制与原子 CRUD，所有表需要在 `pkg/dal/table/table.go` 注册描述符，并通过 `pkg/dal/dao`、`cmd/data-service/service` 和 `pkg/client/data-service` 暴露给上游服务。
- `hc-service` 是插件进程/网关交互与核心逻辑承载层，多副本无状态部署。单个插件版本的完整 Schema JSON 体积可达数 MB，若每次读取均从 `data-service` 拉取完整 JSON 列会造成较大的网络与反序列化开销；若完全依赖本地内存缓存而不做哈希校验，又可能因人工运维排障订正数据库记录而产生多副本不一致。
- 此外，`SchemaManager.Sync` 的基础校验依赖插件登记表 `iac_plugin`（校验 `plugin_name`、`plugin_version` 存在性及 `protocol` 类型），而插件注册模块在另一分支并行开发；同时为了在本分支独立完成闭环自测，需要提供一套用于自测的 `SchemaGetter` 实现。

## Goals / Non-Goals

**Goals:**
- 在 `data-service` 建立系统级权威存储表 `iac_plugin_schema` 及其完整 DAL/DAO/Service/Global Client 链路，支持完整记录读写与 `schema_hash` 单字段轻量投影查询。
- 使用 `github.com/hashicorp/terraform-json` 定义标准化的 `PluginSchema` 与 `SchemaDetail` 模型，支持 JSON 扁平化序列化（`json:",inline"`）与数据库列拆分存储。
- 在 `hc-service` 实现内部接口 `SchemaGetter`、`SchemaManager`（`SyncSchema` + `GetSchema` + `Close`）与副本内 `SchemaCache`（双 Map 分离、事件驱动清理、`singleflight` 回源合并、同步写缓存、TTL 内轻量 `schema_hash` 比对续期）。

**Non-Goals:**
- 不实现插件注册/升级主流程（状态机流转、分布式锁、子进程拉起与探活巡检）。
- 不在 `hc-service` 单独暴露插件 Schema 的 HTTP 接口（后续对外查询逻辑统一收口在北向协议模块，由北向协议直接调用 `SchemaManager.GetSchema`）。
- 不实现基于Schema的资源实例属性校验引擎（由后续插件实例管理模块负责）。

## Decisions

### 1. 数据模型与存储表拆分设计（`PluginSchema` + `SchemaDetail` + `iac_plugin_schema`）

- **选择**：
  - 将 `PluginSchema` 拆分为顶层插件标识结构体 `Plugin`（`PluginInfo`，包含 `Name`、`Version`）、`SchemaHash` 与匿名内嵌的 `SchemaDetail`（带 `json:",inline"` 标签，内部内嵌 `tfjson.ProviderSchema`）：
    ```go
    type PluginInfo struct {
        Name    string `json:"name"`
        Version string `json:"version"`
    }

    type PluginSchema struct {
        Plugin       PluginInfo `json:"plugin"`
        SchemaHash   string     `json:"schema_hash"`
        SchemaDetail `json:",inline"`
    }

    type SchemaDetail struct {
        tfjson.ProviderSchema `json:",inline"`

        ProtocolVersion string       `json:"protocol_version,omitempty"`
        Capabilities    Capabilities `json:"capabilities"`
    }
    ```
  - 数据库表 `iac_plugin_schema` 为系统级表（在 `pkg/dal/table/table.go` 中设置 `EnableTenant: false`），包含独立列 `id`、`plugin_name`、`plugin_version`、`schema_hash`（`varchar(160)`、`validate:"lte=160"`，覆盖 `sha256`/`sha384`/`sha512`）以及 `schema_detail` JSON 列（使用 `types.JsonField` 存放序列化后的 `SchemaDetail`），建立唯一键 `idx_uk_plugin_name_plugin_version (plugin_name, plugin_version)`。
- **理由**：
  - 独立列 `plugin_name`、`plugin_version`、`schema_hash` 支持高频唯一索引检索与仅查询 `schema_hash`（最长 135 字节）的轻量级字段投影（`SELECT schema_hash FROM iac_plugin_schema WHERE ...`），避免每次缓存校验加载数 MB 的 `schema_detail` JSON 列。
  - `SchemaDetail` 使用 `json:",inline"` 内联到 `PluginSchema`，使得插件返回完整 Schema JSON 时可直接一次反序列化为 `PluginSchema`，入库时直接取 `schema.SchemaDetail` 序列化写入 `schema_detail` 列，出库时将独立列与反序列化后的 `SchemaDetail` 组装还原，无需二次裁剪 JSON 树。
  - 引入官方 `github.com/hashicorp/terraform-json`（`tfjson.ProviderSchema`），与 Terraform Provider `GetProviderSchema` 输出格式天然兼容，且与 `pkg/thirdparty/api-gateway/plugin` 共享同一套结构体与校验逻辑。
- **备选方案对比**：
  - *方案 B：自定义简化版 Schema Block 结构体* — 容易遗漏 `cty.Type`、嵌套块（`NestingMode`）、`Deprecated`、`Sensitive` 等标准 Terraform Schema 字段，后续演进维护成本高。

### 2. 通用导出校验函数与 `SchemaManager.SyncSchema` 防篡改幂等落库流程

- **选择**：
  - **抽离通用导出校验函数（供全局多处复用）**：
    - 在 `pkg/api/core/plugin` 提供通用导出能力：
      - `ComputeSchemaHashByAlgo(algo enumor.SchemaHashAlgo, plugin PluginInfo, detail *SchemaDetail) (string, error)`：使用 `pkg/tools/json.Marshal`（已启用 `SortMapKeys: true`）对包含 `plugin`（`name`、`version`）与 `SchemaDetail` 的载荷序列化并计算 `"<algo>:<lowercase hex>"`（支持 `sha256`、`sha384`、`sha512`）；
      - `(detail *SchemaDetail) ValidateCapabilities() error`：按南向协议能力位规则校验 `Capabilities` 正反向约束；`Capabilities` 使用强类型结构体定义，反序列化时忽略未知能力位字段以保证插件协议的前向兼容演进（不做未知 key 白名单拒绝）；
      - `ValidatePluginSchema(schema *PluginSchema, protocol enumor.IaCPluginProtocol) error`：执行单一权威的 Schema 校验入口（`plugin` 非空、`protocol_version`、`schema_hash` 算法格式解析 + 本地重算一致性、`ValidateCapabilities()`、`ConfigSchema.Block` 及类型名 `<plugin.name>_` 前缀与 `Block` 非空）。
  - `SyncSchema(kt *kit.Kit, getter SchemaGetter) (*SyncSchemaResult, error)` 按固定四步执行，并在关键节点通过 `hcm/pkg/logs` 记录英文日志（附带 `kt.Rid`）：
    1. **获取 Schema**：调用 `getter.GetSchema(kt)` 获取 `*PluginSchema`。
    2. **基础校验**：
       - 通过 `data-service` 查询 `iac_plugin` 表中 `(plugin_name, plugin_version)` 登记记录是否存在；
       - 调用通用导出函数 `ValidatePluginSchema(schema, pluginRecord.Protocol)` 完成：
         - 根据 `iac_plugin.protocol` 校验 `ProtocolVersion`：当 `protocol == "http"` 时 `ProtocolVersion` 必须非空（如 `"1.0"`）；当 `protocol == "tf_provider"` 时允许为空字符串 `""`；
         - 校验 `SchemaHash`：通过 `ParseSchemaHash` 校验算法白名单（`sha256`/`sha384`/`sha512`）与对应长度的小写 hex，并调用 `ComputeSchemaHashByAlgo` 计算本地摘要，若与远端返回的 `SchemaHash` 不一致则记录错误日志并返回 `errf.InvalidParameter`；
         - 调用 `schema.ValidateCapabilities()` 校验能力位正反向约束；
         - 校验 `ConfigSchema != nil && ConfigSchema.Block != nil` 以及各类型名 `<plugin.name>_` 前缀与 `Block` 非空。
    3. **比对与写库**：
       - 通过 `data-service` 查询 `iac_plugin_schema` 中 `(plugin_name, plugin_version)` 的 `schema_hash`；
       - 若记录已存在且 `db.SchemaHash == incoming.SchemaHash`：幂等跳过写库，打印 Info 日志并返回 `SyncResult = "success"`、`SyncMessage` 为空；
       - 若记录已存在且 `db.SchemaHash != incoming.SchemaHash`，或拉取、基础校验、写库失败：打印 Error 日志，拒绝覆盖，返回 `SyncResult = "failed"`、`SyncMessage` 为原因、`Schema` 为 nil；
       - 若记录不存在：调用 `data-service` 创建记录，打印 Info 日志并返回 `SyncResult = "success"`、`SyncMessage` 为空。
    4. **懒加载原则**：`SyncSchema` 阶段不写入本地 `SchemaCache`。
- **理由**：
  - 收敛为唯一的 `ValidatePluginSchema` 校验入口，后续插件注册/升级、南向客户端均可直接调用，消除规则分叉。
  - 同名同版本插件的 Schema 一旦发布即不可变（Immutable per version），即使库中已存在同版本记录（如异常中断重入或 HTTP 插件更新 `api_path` 同版本校验），仍拉取最新 Schema 与已落库 `schema_hash` 比对，防止同版本契约被篡改。

### 3. `SchemaCache` 读取流程、同步回填与冷版本过期清理机制

- **选择**：
  - 在 `cmd/hc-service/logics/plugin/schema/cache.go` 中实现 `SchemaCache`，全流程补充 `logs` 英文日志，并由 `SchemaManager.Close()` 透传调用 `SchemaCache.Close()` 释放后台协程：
    - `schemaCache map[coreplugin.PluginInfo]*coreplugin.PluginSchema`：存放完整 Schema 共享只读指针；
    - `expireMap map[coreplugin.PluginInfo]time.Time`：存放各键的过期时间点；
    - `evictCh chan struct{}` / `stopCh chan struct{}`：非阻塞唤醒后台清理协程 `evictLoop()` 回收**已过期且不再被访问的冷版本缓存**；
    - `sfGroup singleflight.Group`：以 `pluginName + ":" + pluginVersion` 为 key 合并并发回源请求。
  - **统一高效的缓存读取与同步回填流程（`Get(kt, plugin)`）**：
    1. **命中且未过期**：向 `data-service` 发起仅包含轻量字段投影的 `schema_hash` 查询：
       - 若 `cached.SchemaHash == dbHash`（有效命中）：刷新 `expireMap[plugin] = time.Now().Add(c.ttl)` 续期，直接返回内存中的共享只读 `*PluginSchema`；
       - 若 `cached.SchemaHash != dbHash`（发现不一致）：走下方第 2 步回源查全量记录并更新缓存。
    2. **未命中 / 发现已过期 / 发现 `schema_hash` 不一致**：
       - 若是发现已过期，顺带非阻塞触发一次 `triggerEvict()`（顺便清理缓存中其他可能已过期的历史冷版本）；
       - 统一通过 `singleflight` 从 `data-service` 查询完整 Schema 记录（若 DB 中记录不存在，清理当前 key 缓存并返回错误）；
       - 查到完整 Schema 后，在 `singleflight` 回调内同步调用 `updateCache` 更新 `schemaCache` 与 `expireMap` 再返回给调用方，保证后续读取立即可见。
- **理由**：
  - 无论是缓存未命中、过期还是 `schema_hash` 不一致，统一收敛到「`singleflight` 查 DB 完整记录 → 同步更新缓存 → 返回共享只读 Schema」，避免并发回源重复打满 DB。
  - **为什么清理协程需要遍历 `expireMap`**：对于当前正在被读取的 key，即使过期或不一致，也会由 `updateCache` 直接覆盖更新；但插件升级后遗留的**历史旧版本**在过期后不会再触发针对旧版本的 `Get`，如果不遍历 `expireMap` 清理已过期的冷版本 key，旧版本的数 MB 对象会驻留在副本内存中无法释放。

## Risks / Trade-offs

- **[Risk] 每次缓存命中仍有一次轻量 DB `schema_hash` 查询的网络 RTT** → **Mitigation**: 查询指定轻量投影字段且命中 `(plugin_name, plugin_version)` 唯一索引，单次传输仅百字节级，规避了数 MB JSON 列的 IO 与反序列化开销。
- **[Risk] 缓存直接返回共享 `*PluginSchema` 指针存在被调用方误改的风险** → **Mitigation**: 在 `SchemaManager.GetSchema` 与 `SchemaCache.Get` 文档注释中明确约束返回对象为进程内共享只读实例，禁止原地修改。

## Migration Plan

1. **数据库迁移**：在部署新版本服务前执行 SQL 脚本创建 `iac_plugin_schema` 表。
2. **服务发布**：先发布 `data-service` 以暴露 `iac_plugin_schema` 接口，再发布 `hc-service`。
3. **回滚策略**：新增表与内部逻辑不影响存量业务链路，回滚时直接回退服务镜像即可，`iac_plugin_schema` 表可安全保留或按需清理。
