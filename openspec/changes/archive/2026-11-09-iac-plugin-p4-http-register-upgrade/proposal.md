## Why

http 类型插件是三方独立部署的服务，经蓝鲸 API 网关访问，HCM 不管理其进程，因此不能像 tf_provider 那样靠目录自动注册。需要管理员通过接口注册，并在插件方发版后能升级、在后端异常时被标出、配置变化时能修改。没有这组能力，三方插件无法接入。

Schema 同步由对接方负责（`integration.SchemaSyncer`，与 tf_provider 共用同一个实例），本 change 只提供 http 插件的 `SchemaGetter` 并消费同步结果。

技术方案：《插件注册与升级》（iWiki 4047929736）1.1 / 1.2、2.3、四章全部、5.4~5.7；评论：api_path 只传后缀、地址前缀走配置、加密配置位置调整、异常状态带 message、探活失败需标出。

## What Changes

- 编排下沉到 hc-service `cmd/hc-service/logics/iac-plugin/httpplugin`（与 `tfplugin` 同层）：网关配置、探活、`SchemaSyncer` 都在 hc-service，进程内串起"加锁 → 探活 → 写 `init` → `SchemaSyncer.Sync` → 写 `enabled` / `failed`"，不再有 cloud-server ↔ hc-service 的探活 / 同步往返路由
- 探活：请求 `{网关地址前缀}{api_path}/health`（带 HCM 应用身份、固定超时、不跟随重定向、响应体上限），取 `plugin_version`；是 hc-service 进程内能力，不暴露路由
- 新增 `api_path` 地址校验（必须是路径后缀、拼接后落在配置的蓝鲸网关环境内）与加密配置校验（`crypto_type` 在支持列表，`RSA` 公钥可解析且不低于 3072 位）
- 注册接口 `POST /api/v1/cloud/iac_plugins/register`（A-03）、升级接口 `POST /api/v1/cloud/iac_plugins/{id}/upgrade`（A-04）、修改接口 `POST /api/v1/cloud/iac_plugins/{id}/update`（A-05）：cloud-server 只做参数校验与 IAM 鉴权，经 `pkg/client/hc-service` 转发给 hc-service 执行
- 判定直接写在 hc-service 的注册 / 升级流程里（`CompareVersion` + 状态分支，直接返回业务错误），不另立判定枚举；状态继承复用 `pkg/iacplugin.InheritState`（与 tf_provider 共用）
- 定时巡检放 hc-service（`httpplugin` 内，仅 `IsMaster()` 副本执行，与 `tfplugin/watch` 并列）：对所有 http 插件探活并走升级流程；`enabled` 插件连续 3 次探活失败置 `failed`，`disabled` 失败只告警，`failed` 探活成功后重入
- 新增权限（沿用 `iac_plugin` 资源类型：register→create，upgrade / update→update）与审计（P1 DAO 事务内写；巡检操作人为后端 kit 的 `hcm-backend-admin`，与 P2 一致）
- 新增 API 文档（3 个接口）
- **不包含**：`SchemaSyncer` 实现（对接方）、tf 插件相关、启停与查询（P3）、删除接口、回滚、插件的实际运行态调用（Apply / Plan 等）

## Capabilities

### New Capabilities

- `iac-plugin-http-probe`：探活、地址校验、加密配置校验（hc-service 进程内能力）
- `iac-plugin-http-register`：注册接口与修改接口
- `iac-plugin-http-upgrade`：升级接口、巡检、异常状态与恢复

### Modified Capabilities

（无）

## Impact

- **cmd/hc-service/logics/iac-plugin/httpplugin**（新增）：探活、Schema getter、注册 / 升级 / 修改编排、巡检
- **cmd/hc-service/service/iac-plugin**（新增）：hc-service 内部路由 register / upgrade / update，供 cloud-server 转发
- **pkg/api/hc-service/iac-plugin**、**pkg/client/hc-service**：上述内部路由的请求 / 响应与客户端
- **cmd/cloud-server/service/iac-plugin**：新增 register / upgrade / update handler（参数校验 + 鉴权 + 转发）
- **pkg/iacplugin**：地址与加密配置校验、`InheritState`（tf_provider 改用公共实现）
- **pkg/cc**：hc-service 新增 `iacGateway`（直接复用 `ApiGateway`）与 `iacPlugin` 下的探活超时、巡检间隔、失败阈值；cloud-server 无新增配置；helm configmap / values 同步
- **docs/api-docs/web-server/docs/scr**：新增 3 个接口文档
- 使用 P1 的 data-service 客户端、状态机、semver、`pkg/lock`；与 P2 共用 `integration.SchemaSyncer`
- 无表结构变化

## 依赖与后续

- 前置：P1、P2（`SchemaSyncer` 注入点、状态收尾写法）、P3（鉴权；启停与注册共用 `lock.IaCPluginKey` 插件名锁，cloud-server 与 hc-service 连同一个 etcd，锁天然互斥）。
- 对接依赖：Schema 对接方交付 `SchemaSyncer` 前沿用 P2 的占位实现（返回错误，注册 / 升级明确失败）；http 插件取 Schema 的接口路径与响应格式待协议负责人确认。
- 待确认（方案 6.1）：`api_path` 构成规则与地址校验的严格程度（安全边界）。
