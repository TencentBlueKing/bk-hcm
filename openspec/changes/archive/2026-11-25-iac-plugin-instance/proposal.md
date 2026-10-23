# Proposal

## Why

在海垒（HCM）IaC 插件化改造架构中，本地插件（`tf_provider`）与三方 HTTP 插件（`http`）的每一次资源或数据源操作均需绑定一份调用身份（插件实例，`plugin_instance_id`），用于承载插件 `provider.block` 所需的启动配置（如云账号密钥 `secret_id` / `secret_key`、跨账号角色扮演 `assume_role`、CMDB `token` 等）。目前平台缺少统一的插件实例持久化表（`iac_plugin_instance`）、对照活跃插件 Schema `provider.block` 的动态属性递归校验与敏感字段同构拆分机制、敏感叶子值加解密存储，以及面向 `cloud-server` / `iac-server` / 执行代理的实例生命周期管理能力（`hc-service` CRUD 逻辑、内部 HTTP 路由与 SDK）。

## What Changes

- **新增数据库表与 SQL 脚本**：
  - 新增 `iac_plugin_instance` 表定义（支持多租户 `tenant_id`，包含 `id`、`plugin_name`、`name`、`attrs`、`secret_attrs`、`extension`、`memo`、`tenant_id` 及审计字段，建立 `idx_plugin_name` 与 `idx_tenant_plugin` 索引，并在 `id_generator` 注册自增主键序列）。
- **新增 Core 模型（`pkg/api/core/plugin`）**：
  - 定义插件实例核心模型（`PluginInstance`）及对外响应形态（含完整明文 `properties` + `extension`）。
- **新增 DAL 数据层（`pkg/dal/table/plugin`、`pkg/dal/dao/plugin`）**：
  - 实现 `IacPluginInstanceTable` 表定义、列描述符、入库/更新校验，并在 `pkg/dal/table/table.go` 注册（开启多租户 `EnableTenant: true`）。
  - 实现 `IacPluginInstance` DAO 接口（`BatchCreateWithTx`、`UpdateWithTx`、`DeleteWithTx`、`List`），并在 `pkg/dal/dao/dao.go` 与 `pkg/dal/dao/types/` 注册。
- **新增 `data-service` 插件实例 CRUD 服务接口与 Global SDK Client（`cmd/data-service/service/plugin`、`pkg/api/data-service/plugin`、`pkg/client/data-service/global`）**：
  - 在 `data-service` 暴露 `iac_plugin_instance` 的批量创建、批量更新、批量删除与列表查询接口（单数路由 `/iac_plugin_instance/...`）；利用现有 `cap.Cipher`，在入库/更新前对 `secret_attrs` 最底层叶子 value 执行 `EncryptToBase64(json.Marshal(leafVal))`，在列表出库后对称解密并 `json.Unmarshal` 还原标量类型；加解密辅助函数放在 `data-service` 服务包内，不放入 `pkg/api/core`。
  - 在 `pkg/client/data-service/global` 注册 `IacPluginInstanceClient` SDK 方法。
- **新增 `hc-service` 实例管理逻辑、内部 HTTP 路由与 HC-Service SDK Client（`cmd/hc-service/logics/plugin/instance`、`cmd/hc-service/service/plugin`、`pkg/api/hc-service/plugin`、`pkg/client/hc-service`）**：
  - 在 `logics/plugin/instance` 以 **concrete struct**（`PluginInstanceManager`）直接实现 `CreatePluginInstance`、`UpdatePluginInstance`、`DeletePluginInstance`、`GetPluginInstance`、`ListPluginInstance`；**不**额外定义 `InstanceManager` / `InstanceStore` 接口。
  - 属性递归校验与同构嵌套 JSON 拆分/深度合并（`splitAndValidateProperties`、`mergeInstanceProperties`，含稀疏数组 `{}` 占位）实现在 `cmd/hc-service/logics/plugin/instance` 内；活跃插件查询与落库读写 **直接调用** `dataCli.Global.IaCPlugin` / `IacPluginInstance`。
  - 创建与更新（传 `properties` 时）校验插件 `state == "enabled"`，通过已有 `*schema.SchemaManager.GetSchema` 取活跃版本 `provider.block` 做校验拆分；更新传入 `properties` 时全量覆盖 `attrs` / `secret_attrs`；`Get`/`List` 将 `attrs` 与已解密 `secret_attrs` Deep Merge 为明文 `properties` 返回（不暴露拆分形态）。
  - 在 `cmd/hc-service/service/plugin` 正式注册内部 HTTP 路由，并在 `pkg/client/hc-service` 提供 `PluginInstance` SDK Client。

## Capabilities

### New Capabilities
- `iac-plugin-instance`: 提供 IaC 插件实例的多租户持久化（`data-service` 表 / DAO / 叶子级 `secret_attrs` 加解密与 Global Client）、`hc-service` 侧基于活跃 `Schema.provider.block` 的动态属性递归校验与同构拆分/深度合并，以及面向内部服务的 HTTP 接口与 SDK Client。

### Modified Capabilities

## Impact

- `scripts/sql/` — 新增 `iac_plugin_instance` 建表与 `id_generator` 初始化 SQL 脚本
- `pkg/api/core/plugin/` — 新增 `PluginInstance` 核心模型（不含拆分/加解密业务逻辑）
- `pkg/dal/table/plugin/` & `pkg/dal/table/table.go` — 新增 `IacPluginInstanceTable` 并注册（`EnableTenant: true`）
- `pkg/dal/dao/plugin/` & `pkg/dal/dao/dao.go` & `pkg/dal/dao/types/` — 新增 `IacPluginInstance` DAO
- `pkg/api/data-service/plugin/` & `cmd/data-service/service/plugin/` & `pkg/client/data-service/global/` — `data-service` 插件实例批量 CRUD、叶子加解密处理器及 Global Client
- `pkg/api/hc-service/plugin/` & `cmd/hc-service/logics/plugin/instance/` & `cmd/hc-service/service/plugin/` & `pkg/client/hc-service/` — 实例管理 logics（含拆分/合并）、HTTP 路由与 SDK Client
