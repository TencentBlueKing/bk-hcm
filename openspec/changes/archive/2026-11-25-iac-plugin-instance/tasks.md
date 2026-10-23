# Tasks

## 1. SQL Scripts and Core Model

- [x] 1.1 在 `scripts/sql/` 下新增 `iac_plugin_instance` 建表与 `id_generator` 初始化 SQL 脚本（含 `id`、`plugin_name`、`name`、`attrs`、`secret_attrs`、`extension`、`memo`、`tenant_id` 及审计字段，建立 `idx_plugin_name` 与 `idx_tenant_plugin` 索引），并核对 DDL 与设计规范一致
- [x] 1.2 在 `pkg/api/core/plugin/instance.go` 中定义 `PluginInstance` 核心模型（`id`、`plugin_name`、`name`、`properties`、`extension`、`memo`、审计字段等对外形态），并通过单元测试或编译验证 JSON tag 与序列化行为正确（执行 `go test ./pkg/api/core/plugin/...` 通过）

## 2. DAL Table and DAO Layer (`pkg/dal`)

- [x] 2.1 在 `pkg/dal/table/plugin/instance.go` 中实现 `IacPluginInstanceTable` 表模型、列描述符与 `InsertValidate` / `UpdateValidate`（禁止更新 `id`、`plugin_name`、`creator`），在 `pkg/dal/table/table.go` 注册（`EnableTenant: true`），并在 `pkg/dal/table/plugin/instance_test.go` 补充表校验单测（执行 `go test ./pkg/dal/table/plugin/...` 通过）
- [x] 2.2 在 `pkg/dal/dao/plugin/instance.go` 中实现 `IacPluginInstance` DAO（`BatchCreateWithTx`、`UpdateWithTx`、`DeleteWithTx`、`List`），在 `pkg/dal/dao/types/plugin.go`、`pkg/dal/dao/plugin.go` 与 `pkg/dal/dao/dao.go` 注册，并验证 `go build ./pkg/dal/...` 编译通过

## 3. Data-Service REST API and Global SDK Client

- [x] 3.1 在 `pkg/api/data-service/plugin/instance.go` 中定义 BatchCreate/BatchUpdate/BatchDelete/List 请求与响应结构，并在 `instance_test.go` 验证请求校验（执行 `go test ./pkg/api/data-service/plugin/...` 通过）
- [x] 3.2 在 `cmd/data-service/service/plugin/instance.go`（及 `service.go` 路由注册）中实现 CRUD 处理器：入库/更新前对 `secret_attrs` 叶子执行 `EncryptToBase64(string(json.Marshal(leafVal)))`，查询出库后对称解密 + `json.Unmarshal`；加解密辅助函数放在本服务包内。通过单测与 `go build ./cmd/data-service/...` 验证
- [x] 3.3 在 `pkg/client/data-service/global/plugin_instance.go` 实现 `IacPluginInstanceClient` 并挂到 `client.go`，验证 `go build ./pkg/client/...` 通过

## 4. HC-Service Instance Logics, HTTP Routes, and SDK Client

- [x] 4.1 在 `cmd/hc-service/logics/plugin/instance/` 以 **concrete struct**（`PluginInstanceManager`，持有 `*dataservice.Client` 与 `*schema.SchemaManager`）实现 `CreatePluginInstance` / `UpdatePluginInstance` / `DeletePluginInstance` / `GetPluginInstance` / `ListPluginInstance`；同包实现 `splitAndValidateProperties` / `mergeInstanceProperties`（同构拆分合并、含稀疏 `{}` 占位）；活跃插件与实例读写 **直接调用** `dataCli.Global.IaCPlugin` 与 `dataCli.Global.IacPluginInstance`。**不**新增 `InstanceManager` / `InstanceStore` 接口。单测覆盖拆分合并、未知/只读拒绝、停用插件拒绝、更新全量覆盖与元数据局部更新（执行 `go test ./cmd/hc-service/logics/plugin/instance/...` 通过）
- [x] 4.2 在 `pkg/api/hc-service/plugin/instance.go` 定义 HTTP 协议结构；在 `cmd/hc-service/service/plugin/instance.go` 注册路由（`/plugin_instances/create`、`PATCH|DELETE|GET /plugin_instances/{id}`、`POST /plugin_instances/list`）；在 `pkg/client/hc-service` 挂载 `PluginInstance` SDK Client。验证相关包编译与测试通过

## Workflow follow-up

- 评审与实现验证完成后，使用 `/opsx-archive` 归档本变更。
