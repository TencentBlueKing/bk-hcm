# Design

## Context

背景与动机见 `proposal.md - Why`，行为契约见 `specs/iac-plugin-instance/spec.md`。

在海垒（HCM）现有的微服务分层架构与已落地的 `iac-plugin`、`iac-plugin-schema` 变更基础上：
- `data-service` 负责数据库表（MySQL）的统一访问控制与原子 CRUD，并且已在 `cmd/data-service/service/capability` 中内置初始化了 `Cipher`（`pkg/cryptography.Crypto`，AES-GCM，已用于云账号密钥加解密）；`hc-service` 当前未配置 `Crypto`。
- `hc-service` 已实现 Schema 查询能力（`cmd/hc-service/logics/plugin/schema`），可按 `coreplugin.PluginInfo`（`name`, `version`）读取 `*coreplugin.PluginSchema`（底层 `*tfjson.SchemaBlock`）。
- 插件登记表 `iac_plugin` 已提供 `dataCli.Global.IaCPlugin.List` 查询能力。
- 项目内 `hc-service/logics` 的常见写法是：**concrete struct 持有 `*dataservice.Client`，业务方法内直接调用 `dataCli.Global.Xxx`**（如 `CvmRelManger`、subnet logics），而不是再包一层 `XxxStore` 接口。

## Goals / Non-Goals

**Goals:**
- 在 `data-service` 建立支持多租户（`EnableTenant: true`）的 `iac_plugin_instance` 表及其完整 DAL/DAO/Service/Global Client 链路；服务层用 `cap.Cipher` 完成 `secret_attrs` 叶子值透明加解密。
- 在 `hc-service/logics/plugin/instance` 以 concrete 类型实现实例 CRUD：含 `provider.block` 递归校验、同构拆分/Deep Merge；直接调用 data-service Client。
- 正式暴露 `hc-service` 内部 HTTP 路由与 `pkg/client/hc-service` SDK。

**Non-Goals:**
- 不实现进程池、`request_identity` / `revision`、实例失效态（`state` / `invalid_reason`）。
- 不提供独立 `GetPluginInstanceSecret`（`Get`/`List` 直接返回合并后的明文 `properties`）。
- 不实现三方插件 `launch_config` 二次 RSA 加密。
- 不在本变更实现 `cloud-server` 账号同步触发与 `iac-server` 删除前 State 引用检查。
- **不**为单实现再抽 `InstanceManager` / `InstanceStore` 接口（对齐项目主流 logics 写法）。

## Decisions

### 1. 属性递归校验与同构嵌套 JSON 拆分/合并放在 `logics/plugin/instance`

- **选择**：
  - 在 `cmd/hc-service/logics/plugin/instance` 内实现 `splitAndValidateProperties` / `mergeInstanceProperties`：
    1. 自顶向下对照 `*tfjson.SchemaBlock`（`Attributes` + `NestedBlocks`）校验未知字段、纯 `computed`、必填项、值类型。
    2. 路径上第一级 `Sensitive == true` 的子树整支进入 `secret_attrs`（保留嵌套 key / 数组结构）；非敏感复合结构继续向下拆。
    3. 列表/集合稀疏敏感元素用 `{}` 占位对齐下标。
    4. 读取时同构树 Deep Merge 还原 `properties`。
  - `PluginInstance` 结构体定义可放在 `pkg/api/core/plugin`（纯模型）；**拆分/合并业务逻辑不放 `pkg/api/core`**。
- **理由**：
  - 与仓库主流一致：api/core 放模型与轻量模型校验；领域拆分逻辑归属 logics。
  - 调用方仅有 `hc-service`，无需为跨服务复用抬到 `pkg/api`。

### 2. `secret_attrs` 叶子加解密：保留 `json.Marshal`，收口在 `data-service`

- **选择**：
  - 加解密只在 `data-service` 插件实例 Service 层（复用 `cap.Cipher`）。
  - 叶子加密：`cipher.EncryptToBase64(string(json.Marshal(leafVal)))`；解密：`DecryptFromBase64` + `json.Unmarshal`，保证 `string` / `number` / `bool` 类型保真。
  - 辅助函数放在 `cmd/data-service/service/plugin` 包内，**不**放进 `pkg/api/core/plugin`。
- **理由**：
  - 对齐方案「data-service 入库前加密」；`hc-service` 无需配置 Crypto。
  - 保留 Marshal：统一处理非 string 敏感叶子的类型还原；实例敏感叶子数量很少，开销可忽略。

### 3. Concrete 实例服务 + 直接调用 data-service Client（不设 InstanceManager / InstanceStore）

- **选择**：
  - 使用 concrete 类型：
    ```go
    type PluginInstanceManager struct {
        dataCli   *dataservice.Client
        schemaMgr *schema.SchemaManager
    }
    ```
  - 方法：`CreatePluginInstance` / `UpdatePluginInstance` / `DeletePluginInstance` / `GetPluginInstance` / `ListPluginInstance`。
  - 活跃插件：`dataCli.Global.IaCPlugin.List(kt, filter: plugin_name + state=enabled)`。
  - 实例读写：`dataCli.Global.IacPluginInstance.BatchCreate/BatchUpdate/BatchDelete/List`。
  - Schema：调用已有 `schemaMgr.GetSchema` 拿到活跃版本 `provider.block`。
  - 更新：`properties != nil` 时全量覆盖 `attrs`/`secret_attrs`；为 nil 时只改 `name`/`extension`/`memo`。
- **理由**：
  - 对齐 `CvmRelManger` 等 logics：持有 `dataCli`，方法内直调 Client。
  - 单实现无需 Manager/Store 接口；单测可测纯函数拆分/合并，或对 Client 做有限 mock，不必为测而加抽象层。

### 4. `hc-service` HTTP 路由与 SDK

- **选择**：
  - 路由（`/api/v1/hc` 下）：
    - `POST /plugin_instances/create`
    - `PATCH /plugin_instances/{id}`
    - `DELETE /plugin_instances/{id}`
    - `GET /plugin_instances/{id}`
    - `POST /plugin_instances/list`
  - `pkg/client/hc-service` 挂载全局 `PluginInstance` 客户端（非 `/vendors/{vendor}` 路径）。
- **理由**：插件实例按 `plugin_name` 区分，不绑定 `enumor.Vendor`。

## Risks / Trade-offs

- **[Risk] 稀疏数组拆分丢下标导致合并错位** → **Mitigation**: 强制 `{}` 占位，并单测覆盖稀疏敏感元素与回环合并。
- **[Risk] 插件升级修改 `sensitive` 标记后存量读取** → **Mitigation**: `attrs`/`secret_attrs` 分列；读路径不依赖当前 Schema，只解密 + Deep Merge；下次 Update 再按新 Schema 全量重拆。

## Migration Plan

1. 部署前执行 `iac_plugin_instance` DDL 与 `id_generator` 初始化。
2. 先发 `data-service`，再发 `hc-service`。
3. 回滚按逆序回退镜像即可；新增表与接口不影响存量云资源链路。
