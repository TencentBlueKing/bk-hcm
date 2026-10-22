# Proposal

## Why

在海垒（HCM）IaC 插件化改造架构中，本地插件（`tf_provider`）与三方 HTTP 插件（`http`）通过统一的 Schema 描述其支持的 `provider` 启动配置、资源（`resource_schemas`）、数据源（`data_source_schemas`）、动作（`action_schemas`）及资源身份标识（`resource_identity_schemas`）。目前平台缺少统一的插件 Schema 权威存储表（`iac_plugin_schema`）以及位于 `hc-service` 内部的 Schema 同步（`Sync`）、基础校验、进程内副本懒加载缓存（`SchemaCache`）与读取（`GetSchema`）机制，无法支撑插件注册/升级时的 Schema 落库、多副本一致性校验以及北向协议与实例管理对 `SchemaManager.GetSchema` 的按需调用。

## What Changes

- **新增数据库表与 SQL 脚本**：
  - 新增 `iac_plugin_schema` 表定义（系统级表，不含 `tenant_id`，包含 `(plugin_name, plugin_version)` 唯一键 `idx_uk_plugin_name_plugin_version`，独立列存储 `id`、`plugin_name`、`plugin_version`、`schema_hash`，`schema_detail` JSON 列整体存储内联的 `RawSchema` 结构及审计字段）。
- **引入第三方依赖**：
  - 引入 `github.com/hashicorp/terraform-json` 定义插件 Schema 本体结构（`*tfjson.Schema`、`*tfjson.SchemaBlock`、`*tfjson.ResourceIdentitySchema` 等）。
- **新增 Core 模型与通用校验函数、DAL 数据层（`pkg/api/core/plugin`、`pkg/dal/table/plugin`、`pkg/dal/dao/plugin`）**：
  - 定义统一 `PluginSchema` 结构体（顶层独立字段 `PluginName`、`PluginVersion`、`SchemaHash` + `json:",inline"` 内联 `RawSchema`）与 `RawSchema` 结构体（含 `ProtocolVersion`、`Capabilities`、`Provider`、`ResourceSchemas`、`DataSourceSchemas`、`ActionSchemas`、`ResourceIdentitySchemas`）。
  - 提供通用的导出校验与摘要计算函数（`ValidatePluginSchema`、`ValidateCapabilities`、`ComputeRawSchemaHash`），供 `SchemaManager.Sync` 及后续插件注册、升级、南向客户端等多处复用。
  - 实现 `iac_plugin_schema` 的 ORM 表定义、列描述符、入库校验及 DAO 层接口（支持按 `(plugin_name, plugin_version)` 创建、列表查询及指定字段投影查询轻量 `schema_hash`）。
- **新增 `data-service` 服务接口与 SDK Client（`cmd/data-service/service/plugin`、`pkg/api/data-service/plugin`、`pkg/client/data-service/global`）**：
  - 在 `data-service` 暴露 `iac_plugin_schema` 的创建与列表/详情查询接口，并在 `pkg/client/data-service/global` 注册对应的 Global Client 方法。
- **新增 `hc-service` Schema 管理与副本缓存模块（`cmd/hc-service/logics/plugin/schema`）**：
  - 定义内部 Go 接口 `SchemaGetter`（`GetSchema(kt *kit.Kit) (*PluginSchema, error)`）与 `SchemaManager`（`SyncSchema(kt *kit.Kit, getter SchemaGetter) (*SyncSchemaResult, error)`、`GetSchema(kt *kit.Kit, pluginName, pluginVersion string) (*PluginSchema, error)`）。
  - 实现 `SchemaManager.SyncSchema`：调用 `getter.GetSchema(kt)` 拉取 Schema，查询 `iac_plugin` 登记记录并调用通用导出校验函数 `ValidatePluginSchema` 完成身份/版本/南向协议版本校验、`schema_hash` 格式（`sha256:<64位hex>`）及本地重算 `schema_hash` 一致性校验、南向协议 `capabilities` 能力位规则校验及 `provider.block` 存在性校验，执行 `(plugin_name, plugin_version)` 幂等写入与同版本防篡改比对（SyncSchema 阶段不写内存缓存），全流程记录英文日志（携带 `kt.Rid`）。
  - 实现 `SchemaCache`：读取缓存时无论是未命中、发现已过期还是发现 `schema_hash` 与 DB 不一致，均从 DB 查询完整 Schema，并通过启动独立协程异步更新缓存（`schemaCache` + `expireMap`）同时向调用方并发返回 Schema 以提升读取效率；同时支持 `evictCh` 唤醒清理协程回收不再被访问的冷版本过期缓存，并支撑 `SchemaManager.GetSchema`。

## Capabilities

### New Capabilities
- `iac-plugin-schema`: 提供 IaC 插件 Schema 的权威存储（`data-service` `iac_plugin_schema` 表与 DAO/API/Client）、`hc-service` 内部 `SchemaManager` 同步入库与基础校验（含防篡改幂等控制）、进程内副本懒加载缓存（`SchemaCache` 含过期清理与 `schema_hash` 一致性比对）以及按版本读取能力。

### Modified Capabilities

## Impact

- `go.mod` / `go.sum` — 新增 `github.com/hashicorp/terraform-json` 依赖
- `scripts/sql/` — 新增 `iac_plugin_schema` 建表 SQL
- `pkg/api/core/plugin/` — 新增 `PluginSchema`、`RawSchema` 等核心结构定义
- `pkg/api/data-service/plugin/` — 新增 `data-service` 插件 Schema 请求/响应结构定义
- `pkg/dal/table/plugin/` & `pkg/dal/table/table.go` — 新增 `iac_plugin_schema` 表结构及表注册
- `pkg/dal/dao/plugin/` & `pkg/dal/dao/dao.go` & `pkg/dal/dao/types/` — 新增 `iac_plugin_schema` DAO
- `cmd/data-service/service/plugin/` & `cmd/data-service/service/service.go` — 新增 `data-service` 插件 Schema 服务路由与处理器
- `pkg/client/data-service/global/` — 新增插件 Schema Global SDK Client
- `cmd/hc-service/logics/plugin/schema/` — 新增 `SchemaManager`、`SchemaGetter`、`SchemaCache` 实现及单元测试
