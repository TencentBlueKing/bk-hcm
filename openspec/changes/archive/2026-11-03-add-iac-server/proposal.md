## Why

在海垒（HCM）插件化改造与 HCL 解析及任务编排整体架构中，平台将职责划分为「编排在 `iac-server`，调度在 `task-server`，单节点插件执行在 `hc-service`，数据存储在 `data-service`」。目前 HCM 尚缺少独立承担 HCL 编译器（Spec 版本托管、HCL 语法解析、Plan/Apply DAG 编排）职责的 `iac-server` 微服务载体及全链路路由发现支持，亟需先行建立标准化的 `iac-server`（`bk-hcm-iacserver`）服务骨架、上下游服务发现通路与构建部署交付模版，为后续 M2/M3 迭代并行挂载具体 IaC 业务模块提供基础底座。

## What Changes

### 本期包含（In Scope）

- **配置中心与监控标识扩展**：
  - 在 `pkg/cc` 中新增服务名常量 `IacServerName = "iac-server"`、`IacServerSetting` 配置结构体（仅含 `Network`、`Service`、`Log`、`Tenant` 四项基础配置，不直连数据库）、`cc.IacServer()` 全局读取函数及 `loadFromFile` 加载分支，并补充配置加载与校验单元测试。
  - 在 `pkg/metrics/http_request.go` 中新增组件标签常量 `ComponentIacServer = string(cc.IacServerName)`。
  - 新增本地默认配置文件 `cmd/iac-server/etc/iac_server.yaml`（默认本地监听端口 `9605`，Etcd 默认地址 `127.0.0.1:2379`）。
- **`cmd/iac-server` 微服务骨架**：
  - 新增 `cmd/iac-server` 进程入口與生命周期管理（`iac_server.go`、`options/options.go`、`app/app.go`、`service/service.go`、`service/capability/capability.go`、`logics/audit/audit.go`）。
  - 支持连接 Etcd 注册 `iac-server` 服务实例，发现下游 `data-service`、`hc-service`、`task-server`、`auth-server` 四个服务，并在收到 `SIGTERM`/`SIGINT` 时先注销 Etcd 节点再优雅关闭 HTTP Server（20s 超时）。
  - 挂载 `/api/v1/iac` 根 `restful.WebService` 容器（本期不注册具体业务子路由，预留挂载入口）以及 `/healthz`、`/alivez`、`/metrics`、`/Version`、`/ctl` 等标准运维与探针端点。
- **内部客户端 SDK 骨架与上下游服务发现/网关反向代理集成**：
  - 新增 `pkg/client/iac-server/client.go` 客户端骨架（基础路径 `/api/%s/iac`），并在 `pkg/client/api.go` 的 `ClientSet` 中新增 `IacServer() *iacserver.Client` 方法。
  - 在 `cmd/cloud-server/app/app.go` 的 `DiscoveryOption.Services` 中注册 `cc.IacServerName`。
  - 在 `cmd/api-server`（`app/app.go` 与 `service/proxy.go`）及 `cmd/web-server`（`app/app.go` 与 `service/proxy.go`）中新增 `cc.IacServerName` 服务发现与 `/api/v1/iac`（`case "iac"`）反向代理转发。
- **构建脚本、启停脚本、容器化与 Helm 部署配置集成**：
  - 新增 `cmd/iac-server/Makefile`（产出 `bk-hcm-iacserver`），并在 `cmd/Makefile` 的 `all`、`package`、`docker`、`clean` 目标中集成 `iac-server`。
  - 在 `scripts/install/start_all.sh` 与 `scripts/install/stop_all.sh` 中追加 `iac-server` 启停命令。
  - 新增 `docs/support-file/docker/bk-hcm-iacserver/Dockerfile`。
  - 同步更新 `docs/support-file/helm/`（`values.yaml` 新增 `iacserver` 配置段，新增 `templates/iacserver/configmap.yaml`、`deployment.yaml`、`service.yaml`，并在 `templates/bklogconfig.yaml` 中追加 `iacserver` 日志采集路径）。

### 本期不包含（Out of Scope）

- **不包含** `iac-server` 的具体业务功能实现（如 HCL Spec 托管与版本管理、HCL 语法解析、Provider Schema 解码生成 ExpectConfig、Plan/Apply DAG 编排、State 管理、版本对比与执行预览等）。
- **不包含** `data-service` 中 Spec / State / 执行记录等数据库表结构新增或 DAO 接口实现。
- **不包含** `task-server`、`hc-service`、`cloud-server` 的编排调度或插件执行业务逻辑改造。
- **不包含** 前端页面（`front/`）或 OpenAPI 网关文档（`docs/api-docs/`）的新增与改造。

## Capabilities

### New Capabilities

- `iac-server-config`: 配置中心（`pkg/cc`）与监控组件标识（`pkg/metrics`）扩展支持 `iac-server`，含 `IacServerName`、`IacServerSetting`（无直连 DB 配置）、`cc.IacServer()`、YAML 加载与强校验及本地配置文件 `cmd/iac-server/etc/iac_server.yaml`。
- `iac-server-bootstrap`: `cmd/iac-server` 服务启动生命周期、Etcd 服务注册与下游服务发现、HTTP Server 根路由容器（`/api/v1/iac`）、运维探针（`/healthz`、`/alivez`、`/metrics`）及优雅关停骨架。
- `iac-server-client-and-proxy`: 内部客户端 SDK 骨架（`pkg/client/iac-server` 与 `ClientSet.IacServer()`）及上下游服务发现与网关代理集成（`cloud-server` 服务发现、`api-server` 与 `web-server` 的 `/api/v1/iac` 反向代理转发）。
- `iac-server-build-and-deploy`: `iac-server` 构建脚本（`Makefile`）、启停脚本（`start_all.sh` / `stop_all.sh`）、容器镜像（`Dockerfile`）与 Helm Chart 部署模版（`values.yaml`、`configmap.yaml`、`deployment.yaml`、`service.yaml`、`bklogconfig.yaml`）集成。

### Modified Capabilities

无（本期不改变既有 capability 的业务规范要求）。

## Impact

- **配置与基础库**：`pkg/cc/service.go`、`pkg/cc/cc.go`、`pkg/cc/load.go`、`pkg/cc/iac_server_test.go`（新增）、`pkg/metrics/http_request.go`。
- **新增微服务目录**：`cmd/iac-server/`（`iac_server.go`、`Makefile`、`etc/iac_server.yaml`、`options/options.go`、`app/app.go`、`service/service.go`、`service/capability/capability.go`、`logics/audit/audit.go`）。
- **客户端 SDK 与存量服务发现/代理**：`pkg/client/iac-server/client.go`（新增）、`pkg/client/api.go`、`cmd/cloud-server/app/app.go`、`cmd/api-server/app/app.go`、`cmd/api-server/service/proxy.go`、`cmd/web-server/app/app.go`、`cmd/web-server/service/proxy.go`。
- **构建与部署交付文件**：`cmd/Makefile`、`scripts/install/start_all.sh`、`scripts/install/stop_all.sh`、`docs/support-file/docker/bk-hcm-iacserver/Dockerfile`、`docs/support-file/helm/values.yaml`、`docs/support-file/helm/templates/iacserver/{configmap,deployment,service}.yaml`、`docs/support-file/helm/templates/bklogconfig.yaml`。
- **外部系统与数据库影响**：仅依赖 Etcd 进行服务注册与发现；遵循架构红线不直连 MySQL，无数据库 DDL 变更，对存量业务接口零破坏性影响。
