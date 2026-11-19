## ADDED Requirements

### Requirement: Service Process Entry, Etcd Registration, Downstream Discovery, and Graceful Shutdown
`cmd/iac-server` SHALL 提供进程入口（`iac_server.go`、`options/options.go`、`app/app.go`），在启动时初始化服务标识 `cc.IacServerName`、加载配置、初始化日志与 Prometheus 指标，通过 `serviced.NewServiceD` 建立 Etcd 服务发现（下游发现列表 MUST 包含 `cc.DataServiceName`、`cc.HCServiceName`、`cc.TaskServerName`、`cc.AuthServerName`），在启动 HTTP Server 监听后将自身注册至 Etcd，并在接收到退出信号（`SIGTERM`/`SIGINT`）时先执行 `sd.Deregister()` 注销实例再在 20 秒超时窗口内优雅关闭 HTTP Server。

#### Scenario: Normal startup, Etcd registration, and graceful shutdown on signal (AC-003, AC-P02)
- **WHEN** Etcd 服务可用且执行 `bk-hcm-iacserver -c ./cmd/iac-server/etc/iac_server.yaml` 启动服务，随后向进程发送 `SIGTERM` 信号
- **THEN** 服务在 5 秒内完成启动并在 Etcd `/hcm/services/iac-server/` 下注册实例键值；收到 `SIGTERM` 后先从 Etcd 注销该实例键值，并在 20 秒内优雅关闭 HTTP Server，进程退出码为 `0`

#### Scenario: Fast failure when Etcd is unreachable at startup (AC-004)
- **WHEN** 配置文件中 `service.etcd.endpoints` 指向不可达地址且 `dialTimeoutMS: 200`，启动 `bk-hcm-iacserver`
- **THEN** `serviced.NewServiceD` 在连接超时后快速返回错误，进程向 `stderr` 输出 `"new service discovery failed"` 相关错误信息、关闭日志并以退出码 `1` 终止

### Requirement: Root RESTful WebService Container and Capability Assembly
`cmd/iac-server/service` SHALL 构建 `http.ServeMux` 并挂载根 `restful.WebService` 容器（根路径 MUST 为 `/api/v1/iac`，Produces MUST 为 `application/json`），同时组装 `capability.Capability`（持有 `WebService`、`ApiClient`、`Authorizer`、`Audit`），本期不注册具体业务子路由。

#### Scenario: Unregistered business route under /api/v1/iac safely returns 404 (AC-008)
- **WHEN** `iac-server` 正常运行，客户端发起 `GET /api/v1/iac/non-existent-path` 请求
- **THEN** 服务返回 HTTP `404 Not Found` 响应，且进程不发生 panic

### Requirement: Operational Health, Liveness, and Metrics Endpoints
`cmd/iac-server/service` SHALL 暴露 `/healthz`、`/alivez` 探针端点并通过 `handler.SetCommonHandler` 挂载 `/metrics`、`/Version`、`/ctl` 通用运维端点。

#### Scenario: Probes and metrics respond normally during healthy runtime (AC-005, AC-P01)
- **WHEN** `iac-server` 正常运行且 Etcd 连接健康，客户端分别请求 `GET /alivez`、`GET /healthz` 与 `GET /metrics`
- **THEN** `GET /alivez` 返回 HTTP 200 且响应体为 `{"code":0,"message":"alive"}`；`GET /healthz` 返回 HTTP 200 且响应体为 `{"code":0,"message":"healthy"}`；`GET /metrics` 返回 HTTP 200 且包含 Prometheus 指标文本；100 并发下探针 P99 响应时间 < 50ms

#### Scenario: Probes return unhealthy status during shutdown or Etcd failure (AC-006)
- **WHEN** `iac-server` 处于关停状态（`shutdown.IsShuttingDown() == true`）或运行期 `serviced.Healthz` 检测 Etcd 返回错误时，客户端请求 `GET /alivez` 与 `GET /healthz`
- **THEN** 关停状态下 `/alivez` 与 `/healthz` 均返回 HTTP 503 且 JSON 响应 `code` 为 `2000009`（`errf.UnHealthy`）；运行期 Etcd 探活失败时 `/healthz` 返回 JSON 响应 `code` 为 `2000009` 且 `message` 包含 `"serviced healthz error"`
