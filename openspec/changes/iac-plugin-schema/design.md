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
- 使用 `github.com/hashicorp/terraform-json` 定义标准化的 `PluginSchema` 与 `RawSchema` 模型，支持 JSON 扁平化序列化（`json:",inline"`）与数据库列拆分存储。
- 在 `hc-service` 实现内部接口 `SchemaGetter`、`SchemaManager`（`SyncSchema` + `GetSchema`）与副本内 `SchemaCache`（双 Map 分离、事件驱动清理、`singleflight` 回源合并、异步写缓存、TTL 内轻量 `schema_hash` 比对续期）。

**Non-Goals:**
- 不实现插件注册/升级主流程（状态机流转、分布式锁、子进程拉起与探活巡检）。
- 不在 `hc-service` 单独暴露插件 Schema 的 HTTP 接口（后续对外查询逻辑统一收口在北向协议模块，由北向协议直接调用 `SchemaManager.GetSchema`）。
- 不实现基于Schema的资源实例属性校验引擎（由后续插件实例管理模块负责）。

## Decisions

### 1. 数据模型与存储表拆分设计（`PluginSchema` + `RawSchema` + `iac_plugin_schema`）

- **选择**：
  - 将 `PluginSchema` 拆分为顶层独立字段（`PluginName`、`PluginVersion`、`SchemaHash`）与匿名内嵌的 `RawSchema`（带 `json:",inline"` 标签）：
    ```go
    type PluginSchema struct {
        PluginName    string `json:"plugin_name"`
        PluginVersion string `json:"plugin_version"`
        SchemaHash    string `json:"schema_hash"`
        RawSchema     `json:",inline"`
    }

    type RawSchema struct {
        ProtocolVersion         string                             `json:"protocol_version,omitempty"`
        Capabilities            map[string]bool                    `json:"capabilities"`
        Provider                *tfjson.Schema                     `json:"provider"`
        ResourceSchemas         map[string]*tfjson.Schema          `json:"resource_schemas,omitempty"`
        DataSourceSchemas       map[string]*tfjson.Schema          `json:"data_source_schemas,omitempty"`
        ActionSchemas           map[string]*tfjson.Schema          `json:"action_schemas,omitempty"`
        ResourceIdentitySchemas map[string]*tfjson.IdentitySchema  `json:"resource_identity_schemas,omitempty"`
    }
    ```
  - 数据库表 `iac_plugin_schema` 为系统级表（在 `pkg/dal/table/table.go` 中设置 `EnableTenant: false`），包含独立列 `id`、`plugin_name`、`plugin_version`、`schema_hash` 以及 `schema_detail` JSON 列（使用 `types.JsonField` 存放序列化后的 `RawSchema`），建立唯一键 `idx_uk_plugin_name_plugin_version (plugin_name, plugin_version)`。
- **理由**：
  - 独立列 `plugin_name`、`plugin_version`、`schema_hash` 支持高频唯一索引检索与仅查询 71 字节 `schema_hash` 的轻量级字段投影（`SELECT schema_hash FROM iac_plugin_schema WHERE ...`），避免每次缓存校验加载数 MB 的 `schema_detail` JSON 列。
  - `RawSchema` 使用 `json:",inline"` 内联到 `PluginSchema`，使得插件返回完整 Schema JSON 时可直接一次反序列化为 `PluginSchema`，入库时直接取 `&schema.RawSchema` 序列化写入 `schema_detail` 列，出库时将独立列与反序列化后的 `RawSchema` 组装还原，无需二次裁剪 JSON 树。
  - 引入官方 `github.com/hashicorp/terraform-json`（`tfjson.Schema`、`tfjson.IdentitySchema` 等），与 Terraform Provider `GetProviderSchema` 输出格式天然兼容。
- **备选方案对比**：
  - *方案 B：自定义简化版 Schema Block 结构体* — 容易遗漏 `cty.Type`、嵌套块（`NestingMode`）、`Deprecated`、`Sensitive` 等标准 Terraform Schema 字段，后续演进维护成本高。

### 2. 通用导出校验函数与 `SchemaManager.SyncSchema` 防篡改幂等落库流程

- **选择**：
  - **抽离通用导出校验函数（供全局多处复用）**：
    - 在 `pkg/api/core/plugin`（并在 `cmd/hc-service/logics/plugin/schema` 暴露便捷调用）提供通用导出函数：
      - `ComputeRawSchemaHash(raw *RawSchema) (string, error)`：使用 `pkg/tools/json.Marshal`（已启用 `SortMapKeys: true`）对 `RawSchema` 序列化并计算 `"sha256:" + hex`；
      - `ValidateCapabilities(raw *RawSchema) error`：按南向协议 §4.1 规则校验 `Capabilities`；
      - `ValidatePluginSchema(schema *PluginSchema, expectedPluginName, expectedPluginVersion, protocol string) error`：执行完整的 Schema 基础校验（身份版本匹配、`protocol_version`、`schema_hash` 正则格式 + 本地重算一致性、`Capabilities` 协议规则、`Provider.Block` 非空）。
  - `SyncSchema(kt *kit.Kit, getter SchemaGetter) (*SyncSchemaResult, error)` 按固定四步执行，并在关键节点通过 `hcm/pkg/logs` 记录英文日志（附带 `kt.Rid`）：
    1. **获取 Schema**：调用 `getter.GetSchema(kt)` 获取 `*PluginSchema`。
    2. **基础校验**：
       - 校验 `PluginName`、`PluginVersion` 非空，并通过 `data-service` 查询 `iac_plugin` 表中 `(plugin_name, plugin_version)` 登记记录是否存在且一致；
       - 调用通用导出函数 `ValidatePluginSchema(schema, pluginRecord.PluginName, pluginRecord.PluginVersion, pluginRecord.Protocol)` 完成：
         - 根据 `iac_plugin.protocol` 校验 `ProtocolVersion`：当 `protocol == "http"` 时 `ProtocolVersion` 必须非空（如 `"1.0"`）；当 `protocol == "tf_provider"` 时允许为空字符串 `""`；
         - 校验 `SchemaHash`：不仅校验格式符合正则 `^sha256:[0-9a-f]{64}$`（长度 71），还调用 `ComputeRawSchemaHash(&schema.RawSchema)` 计算本地 SHA-256 摘要，若与远端返回的 `SchemaHash` 不一致则记录错误日志并返回 `errf.InvalidParameter`；
         - 校验 `Capabilities`（与海垒插件南向协议 §4.1 能力位规则对齐）：
           - `len(Capabilities) > 0`，且所有 key 必须属于合法能力位集合（`apply`、`plan`、`refresh`、`read_data`、`async_apply`、`stop`、`validate`、`import`、`list`）；
           - `resource_schemas` 非空（`len(ResourceSchemas) > 0`）时，`apply`、`plan`、`refresh` 必须全部为 `true`；
           - `data_source_schemas` 非空（`len(DataSourceSchemas) > 0`）时，`read_data` 必须为 `true`；为空（`len(DataSourceSchemas) == 0`）时，不得声明 `read_data`（即 `Capabilities` 中不能包含 `read_data` 键）；
           - 声明 `stop`（`Capabilities["stop"] == true`）时，必须同时声明 `async_apply`（`Capabilities["async_apply"] == true`）；
           - 任一能力位规则不满足即拒绝同步，返回指明具体能力位与原因的 `errf.InvalidParameter`；
         - 校验 `Provider != nil && Provider.Block != nil`。
    3. **比对与写库**：
       - 通过 `data-service` 查询 `iac_plugin_schema` 中 `(plugin_name, plugin_version)` 的 `schema_hash`；
       - 若记录不存在：调用 `data-service` 创建记录，打印 Info 日志并返回 `sync_result = "success"`、`sync_message` 为空（若并发写入触发唯一键冲突，重新查询数据库记录并走哈希比对分支）；
       - 若记录已存在且 `db.SchemaHash == incoming.SchemaHash`：幂等跳过写库，打印 Info 日志并返回 `sync_result = "success"`、`sync_message` 为空；
       - 若记录已存在且 `db.SchemaHash != incoming.SchemaHash`，或拉取、基础校验失败：打印 Error 日志，拒绝覆盖，返回 `sync_result = "failed"`、`sync_message` 为原因、`schema` 为 nil。
    4. **懒加载原则**：`Sync` 阶段不写入本地 `SchemaCache`。
- **理由**：
  - 将 `ValidatePluginSchema`、`ValidateCapabilities`、`ComputeRawSchemaHash` 设计为通用导出函数，后续插件注册/升级、南向客户端或测试工具均可直接调用，避免重复实现校验规则。
  - 同名同版本插件的 Schema 一旦发布即不可变（Immutable per version），若哈希不一致说明二进制或远端契约发生了未升版篡改，必须阻断并告警。

### 3. `SchemaCache` 读取流程、协程并发更新缓存与冷版本过期清理机制

- **选择**：
  - 在 `cmd/hc-service/logics/plugin/schema/cache.go` 中实现 `SchemaCache`，全流程补充 `logs` 英文日志：
    - `schemaCache map[cacheKey]*PluginSchema`：存放完整 Schema 指针；
    - `expireMap map[cacheKey]time.Time`：存放各 `cacheKey` 的过期时间点；
    - `evictCh chan struct{}`：容量为 1 的非阻塞通道，用于唤醒后台清理协程 `StartEvictWorker(ctx)` 回收**已过期且不再被访问的冷版本缓存**；
    - `sfGroup singleflight.Group`：以 `pluginName + ":" + pluginVersion` 为 key 合并并发回源请求；
    - `asyncUpdateCache(key, dbSchema)`：启动独立 goroutine 获取写锁更新 `schemaCache[key]` 与 `expireMap[key]`。
  - **统一高效的缓存读取与并发回填流程（`Get(kt, key)`）**：
    1. **命中且未过期**：向 `data-service` 发起仅包含 `Fields: []string{"schema_hash"}` 的轻量查询：
       - 若 `cached.SchemaHash == dbHash`（有效命中）：刷新 `expireMap[key] = time.Now().Add(c.ttl)` 续期，直接返回内存中的 `*PluginSchema`；
       - 若 `cached.SchemaHash != dbHash`（发现不一致）：走下方第 2 步查 DB 完整记录 + 起协程更新缓存 + 并发返回 Schema。
    2. **未命中 / 发现已过期 / 发现 `schema_hash` 不一致**：
       - 若是发现已过期，顺带非阻塞触发一次 `triggerEvict()`（顺便清理缓存中其他可能已过期的历史冷版本）；
       - 统一通过 `singleflight` 从 `data-service` 查询完整 Schema 记录（若 DB 中记录不存在，清理当前 key 缓存并返回错误）；
       - 查到完整 Schema 后，**启动独立协程（`asyncUpdateCache`）异步更新 `schemaCache[key]` 与 `expireMap[key]`，同时当前函数立即向调用方并发返回该 Schema**，避免写锁等待阻塞当前请求，提升读取效率。
- **理由**：
  - 无论是缓存未命中、过期还是 `schema_hash` 不一致，统一收敛到「`singleflight` 查 DB 完整记录 → 独立协程更新缓存 + 并发返回 Schema」，既保证了数据最新，又让调用方无需等待缓存写锁。
  - **为什么清理协程需要遍历 `expireMap`**：对于当前正在被读取的 key，即使过期或不一致，也会由上面的 `asyncUpdateCache` 直接覆盖更新；但插件升级后遗留的**历史旧版本**（例如从 `1.81.4` 升级到 `1.81.5` 后，再也不会有请求去读 `1.81.4`）在过期后永远不会再触发针对旧版本的 `Get`，如果不遍历 `expireMap` 找出所有 `now.After(expireAt)` 的冷版本 key 并删除，旧版本的数 MB 对象会永久驻留在副本内存中无法释放。

## Risks / Trade-offs

- **[Risk] 每次缓存命中仍有一次轻量 DB `schema_hash` 查询的网络 RTT** → **Mitigation**: 查询指定 `Fields: []string{"schema_hash"}` 且命中 `(plugin_name, plugin_version)` 唯一索引（覆盖索引扫描），单次传输仅百字节级，完全规避了数 MB JSON 列的 IO 与反序列化开销，符合设计方案 §3.2.2 的强一致性权衡。
- **[Risk] 多副本并发首次执行 `Sync` 时的唯一键冲突** → **Mitigation**: `Sync` 写库遇到唯一键冲突时自动回查 `iac_plugin_schema` 并比对 `schema_hash`，若一致则按幂等成功（`Created: false`）返回，保证并发安全。

## Migration Plan

1. **数据库迁移**：在部署新版本服务前执行 SQL 脚本创建 `iac_plugin_schema` 表。
2. **服务发布**：先发布 `data-service` 以暴露 `iac_plugin_schema` 接口，再发布 `hc-service`。
3. **回滚策略**：新增表与内部逻辑不影响存量业务链路，回滚时直接回退服务镜像即可，`iac_plugin_schema` 表可安全保留或按需清理。
