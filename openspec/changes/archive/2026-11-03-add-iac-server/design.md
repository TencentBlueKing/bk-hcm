## Context

在海垒（HCM）插件化改造与 HCL 解析及任务编排整体架构中，平台明确将职责切分为：
- **编排层（`iac-server`）**：承担 HCL 编译器职责（Spec 版本托管、HCL 语法解析、Provider Schema 解码、Plan/Apply DAG 编排）；
- **调度层（`task-server`）**：负责异步 Flow/Task 调度执行；
- **执行层（`hc-service`）**：负责单节点 Terraform Provider 插件调用；
- **存储层（`data-service`）**：唯一允许直连 MySQL 数据库的服务。

当前代码库包含 8 个核心服务（`api-server`、`web-server`、`cloud-server`、`auth-server`、`account-server`、`task-server`、`hc-service`、`data-service`），尚无 `iac-server` 服务载体。本期目标是在不涉及任何具体 IaC 业务接口与数据库表变更的前提下，100% 对齐现有 HCM 微服务脚手架规范，交付可独立编译、注册发现、网关路由、容器化打包与 Helm 部署的 `iac-server`（二进制名 `bk-hcm-iacserver`）基础骨架。

## Goals / Non-Goals

**Goals:**
- 在 `pkg/cc` 与 `pkg/metrics` 中完成 `iac-server` 服务名、最小无状态配置结构体 `IacServerSetting`（`Network`、`Service`、`Log`、`Tenant`）及监控组件常量注册，并提供完备的单测覆盖。
- 搭建 `cmd/iac-server` 标准目录结构（`iac_server.go`、`options`、`app`、`service`、`capability`、`logics/audit`、`etc/iac_server.yaml`、`Makefile`），支持 Etcd 注册、下游四服务发现（`data-service`、`hc-service`、`task-server`、`auth-server`）、`/api/v1/iac` 根容器及 `/healthz`、`/alivez`、`/metrics` 探针与优雅关停。
- 新增 `pkg/client/iac-server` 客户端 SDK 骨架并挂载至 `pkg/client.ClientSet`，同时打通 `cloud-server` 服务发现以及 `api-server`、`web-server` 对 `/api/v1/iac/*` 的反向代理路由。
- 补齐 `cmd/Makefile`、`scripts/install/start_all.sh`、`scripts/install/stop_all.sh`、`docs/support-file/docker/bk-hcm-iacserver/Dockerfile` 以及 `docs/support-file/helm/` 下的 `values.yaml` 与 `templates/iacserver/*`、`templates/bklogconfig.yaml`。

**Non-Goals:**
- 不实现任何具体 IaC 业务接口（HCL 解析、Spec 托管、Plan/Apply 编排等均在后续需求中交付）。
- 不直连 MySQL 数据库，不在 `data-service` 新增任何表或 DAO。
- 不修改前端工程（`front/`）或网关公开文档（`docs/api-docs/`）。

## Decisions

### 1. `IacServerSetting` 配置结构最小化设计（不引入 Database 与第三方网关配置）
- **决策**：`IacServerSetting` 仅包含 `Network`、`Service`、`Log`、`Tenant` 四个字段，实现与 `ApiServerSetting` 对齐的 `trySetFlagBindIP`、`trySetDefault`、`Validate`、`TenantEnable` 方法；本地 `cmd/iac-server/etc/iac_server.yaml` 默认端口设为 `9605`（避开既有服务占用的 `9600`~`9604` 及 `9609`），Helm 容器内端口同其他服务统一设为 `80`。
- **理由**：严格遵守平台微服务分层架构红线「除 `data-service` 外禁止直连 DB」；且本期无 ESB/CMDB/ITSM 直接调用需求，避免预置未使用的第三方配置造成启动校验冗余失败。
- **备选方案**：照搬 `TaskServerSetting`（含 `Database`）或 `AccountServerSetting`（含 `Cmdb`/`Esb`）——因违背无 DB 红线且引入无关配置校验，予以排除。

### 2. `cmd/iac-server` 服务内部结构对标 `account-server` / `cloud-server` 样板
- **决策**：
  - `cmd/iac-server/app/app.go`：发现列表固定为 `[]cc.Name{cc.DataServiceName, cc.HCServiceName, cc.TaskServerName, cc.AuthServerName}`；使用 `ctl.LoadCtl(ctl.WithBasics(sd)...)` 注册通用控制指令；`shutdown.WaitShutdown(20)` 与 HTTP Server `Shutdown` 20s 超时对齐。
  - `cmd/iac-server/service/capability/capability.go`：定义 `Capability` 结构体，持有 `WebService *restful.WebService`、`ApiClient *client.ClientSet`、`Authorizer auth.Authorizer`、`Audit audit.Interface`。
  - `cmd/iac-server/logics/audit/audit.go`：对标 `cmd/account-server/logics/audit/audit.go` 提供本服务独立的轻量 `audit.Interface` 与 `audit.NewAudit(apiClientSet.DataService())` 骨架，避免跨 `cmd/` 包引用其他微服务内部包。
- **理由**：与仓库现有微服务分层（`main -> options -> app -> service -> capability/logics`）保持 100% 一致，后续研发挂载业务 Service 时无需重构 `Service` 或 `Capability` 签名。

### 3. 上下游服务发现与网关路由转发点位
- **决策**：
  - 下游客户端：新增 `pkg/client/iac-server/client.go`（包名 `iacserver`，基础路径 `/api/%s/iac`），在 `pkg/client/api.go` 的 `ClientSet` 中新增 `IacServer() *iacserver.Client`。
  - 上游调用方：在 `cmd/cloud-server/app/app.go` 的 `DiscoveryOption.Services` 追加 `cc.IacServerName`。
  - 网关层：在 `cmd/api-server/app/app.go`、`cmd/api-server/service/proxy.go`、`cmd/web-server/app/app.go`、`cmd/web-server/service/proxy.go` 的服务发现列表追加 `cc.IacServerName`，并在 `proxy.go` 的 `switch servicePath` 中新增 `case "iac": service = cc.IacServerName`。
- **理由**：`api-server` 与 `web-server` 均通过 URL 第 3 段（`/api/v1/<servicePath>/...`）静态映射目标 `cc.Name` 并从本地 `discovery` map 查找实例；必须同时更新 `app.go` 的 `DiscoveryOption.Services` 与 `proxy.go` 的 `discoverServices` + `switch` 分支，否则网关会报 `Service Not Supported` 或 `DeniedServers`。

### 4. Helm 部署模版对齐 `taskserver` / `accountserver` 标准形态
- **决策**：
  - `docs/support-file/helm/templates/iacserver/deployment.yaml` 的 `initContainers` 包含等待 `dataservice`、`authserver`、`hcservice` 的初始化容器（对标 `accountserver/deployment.yaml`），确保核心依赖就绪后再拉起 `iacserver`。
  - `docs/support-file/helm/templates/iacserver/configmap.yaml` 仅渲染 `network`、`service.etcd`、`log`、`tenant` 四段，与 `IacServerSetting` 严格对齐。
- **理由**：遵循用户规则「当涉及配置的调整时，需要变更服务 etc 目录下的 yaml 文件和 docs/support-file/helm」，保证本地与 K8s 容器环境配置结构完全一致。

## Risks / Trade-offs

- **[Risk] `pkg/cc` 全局单例 `initOnce` / `runtimeOnce` 在单元测试中多测试函数共享状态** → **Mitigation**：在 `pkg/cc/iac_server_test.go` 中参考 `pkg/cc/async_flow_cleanup_test.go` 的模式，直接针对 `IacServerSetting` 实例测试 `yaml.Unmarshal`、`trySetFlagBindIP`、`trySetDefault`、`Validate`、`TenantEnable`，并在测试 `LoadSettings` / `loadFromFile` / `IacServer()` 时妥善设置 `serviceName = IacServerName`，避免受 `sync.Once` 顺序影响。
- **[Risk] `api-server` / `web-server` 的 `DiscoveryOption` 与 `proxy.go` 服务列表漏改一处导致运行时代理 404** → **Mitigation**：在 `tasks.md` 中明确列出 `app.go` 与 `proxy.go` 共 4 处修改点并逐一核对。

## Migration Plan

1. **合入与构建**：合入代码后，CI 流水线通过 `cmd/Makefile` 自动编译 `bk-hcm-iacserver` 并构建 `blueking/bk-hcm-iacserver` 镜像。
2. **部署上线**：Helm Upgrade 时自动创建 `*-iacserver-config` ConfigMap、`*-iacserver` Service 与 Deployment，并滚动更新 `apiserver`、`webserver`、`cloudserver`。
3. **回滚策略**：由于本期不涉及任何数据库 DDL 变更或存量接口协议修改，若需回滚，直接回退 Helm Chart 版本或将 `iacserver.replicas` 缩容为 `0` 即可，对存量服务无副作用。

## Open Questions

无（所有端口、配置段、服务发现范围与网关路由点位均已在需求澄清阶段由用户确认闭环）。
