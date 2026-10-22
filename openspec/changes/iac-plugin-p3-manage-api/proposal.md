## Why

插件注册表建好后，管理员需要查看插件状态，并能启用 / 停用插件（下线插件、临时停用异常插件）。启停还要保证与注册、升级串行，且不能把"Schema 不存在"的停用插件直接启用。没有这组接口，只能直连数据库改状态。

Schema 是否存在的判断依赖对接方，本 change 通过一个方法调用对接方的查询能力，不自行查询 Schema 存储；交付前该方法默认放行。

技术方案：《插件注册与升级》（iWiki 4047929736）5.1~5.3、5.7；评论：列表与详情合并为一个接口、启用 / 停用合并为一个接口。

## What Changes

- cloud-server 新增 `POST /api/v1/cloud/iac_plugins/list`（A-01）：沿用 `filter` + `page`，可按 `id`、`plugin_name`、`protocol`、`state` 过滤，列表与详情共用
- cloud-server 新增 `POST /api/v1/cloud/iac_plugins/switch`（A-02）：请求体 `{"plugin_name":"...","state":"enabled|disabled"}`，按插件"当前版本"（版本最高）判断，只改注册表 `state`，立即返回
- 启停对该插件名下所有处于 `enabled` / `failed`（目标 disabled）或 `disabled`（目标 enabled）的版本批量生效，`init` 版本不动；持插件名 etcd 锁，与注册、升级串行
- 启用 `disabled` 插件前，检查当前版本 Schema 已存在（对接接口），否则拒绝并提示先恢复
- 新增权限：`meta.ResourceType` 增加 `iac_plugin`，find / update / create 映射 IAM `global_configuration`（平台级，无关联资源实例）
- 新增审计：状态实际变化时写 update 审计，已是目标状态的幂等返回不写
- 新增请求入口可复用的"插件可用性判定"函数（按状态返回区分的错误信息），供运行态在后续阶段使用
- 新增 API 文档（`docs/api-docs/web-server/docs/scr/`，版本占位 `v9.9.9`）
- **不包含**：http 插件的注册 / 升级 / 修改接口（P4）、各 hc-service 副本对停用的进程池响应与巡检（运行态）、tf 插件的注册 / 升级接口（tf 随目录自动注册）、删除接口

## Capabilities

### New Capabilities

- `iac-plugin-manage-query`：插件查询接口与权限
- `iac-plugin-manage-switch`：启用 / 停用接口、状态规则、串行保证、审计、可用性判定函数

### Modified Capabilities

（无）

## Impact

- **cmd/cloud-server/service/iac-plugin**（新增）：`InitService`、list / switch handler，并在 `service.go` 注册
- **pkg/api/cloud-server**：新增请求 / 响应结构体
- **pkg/iam/meta/resource.go**、**cmd/auth-server/service/auth/adaptor.go**、**gen_id.go**：新增资源类型及映射
- **cmd/cloud-server/service/iac-plugin**：`isIaCPluginSchemaExists`（Schema 是否存在），交付前默认返回 `true` + warn
- **docs/api-docs/web-server/docs/scr**：新增 2 个接口文档
- 使用 P1 的 data-service 客户端、状态机、`pkg/lock`
- 无表结构变化；cloud-server 需要 etcd 客户端配置（`cc.CloudServer().Service.Etcd` 已存在）

## 依赖与后续

- 前置：P1。
- 对接依赖：Schema 对接方提供存在性查询接口；交付前默认放行，交付后替换 `isIaCPluginSchemaExists`。
- 被依赖：P4 复用同包的串行加锁与权限。
