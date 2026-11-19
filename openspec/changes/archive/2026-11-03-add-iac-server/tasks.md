## 1. 配置中心与监控组件标识扩展（`iac-server-config`）

- [x] 1.1 在 `pkg/cc/service.go` 中新增服务名常量 `IacServerName Name = "iac-server"` 与 `IacServerSetting` 结构体（仅含 `Network`、`Service`、`Log`、`Tenant` 四项配置，不含 `Database`），实现 `trySetFlagBindIP`、`trySetDefault`、`Validate`、`TenantEnable` 方法
- [x] 1.2 在 `pkg/cc/cc.go` 中新增 `IacServer() IacServerSetting` 全局配置获取函数，并在 `pkg/cc/load.go` 的 `loadFromFile` 中新增 `case IacServerName: s = new(IacServerSetting)` 分支
- [x] 1.3 在 `pkg/metrics/http_request.go` 中新增组件标签常量 `ComponentIacServer = string(cc.IacServerName)`
- [x] 1.4 新增默认配置文件 `cmd/iac-server/etc/iac_server.yaml`（默认本地监听端口 `9605`，Etcd 地址 `127.0.0.1:2379`，包含 `network`、`service`、`log` 基础配置且无硬编码敏感凭证）
- [x] 1.5 新增 `pkg/cc/iac_server_test.go` 单元测试，覆盖合法最小配置加载与默认值填充（AC-001）、缺失配置文件/参数冲突/非法 IP/单边 TLS 证书拦截（AC-002、AC-S01）以及 `IacServerSetting` 不含 `Database` 字段断言（AC-008）

## 2. `cmd/iac-server` 微服务骨架实现（`iac-server-bootstrap`）

- [x] 2.1 新增 `cmd/iac-server/options/options.go` 与 `cmd/iac-server/iac_server.go`，完成 `cc.InitService(cc.IacServerName)`、命令行参数解析、`--version` 检查及 `app.Run(opts)` 入口调用
- [x] 2.2 新增 `cmd/iac-server/logics/audit/audit.go` 与 `cmd/iac-server/service/capability/capability.go`，定义 `Capability` 结构体（包含 `WebService`、`ApiClient`、`Authorizer`、`Audit`）
- [x] 2.3 新增 `cmd/iac-server/service/service.go`，实现 `NewService`、`ListenAndServeRest`、`apiSet`（根路径 `/api/v1/iac`，Produces `application/json`）、`Healthz`（`/healthz`）、`Alivez`（`/alivez`）及 `handler.SetCommonHandler(root)`
- [x] 2.4 新增 `cmd/iac-server/app/app.go`，串联 `cc.LoadSettings`、`logs.InitLogger`、`metrics.InitMetrics`、`serviced.NewServiceD`（下游发现列表为 `DataServiceName`、`HCServiceName`、`TaskServerName`、`AuthServerName`）、`svc.ListenAndServeRest`、`sd.Register`、`ctl.LoadCtl` 及 20 秒优雅关停注销
- [x] 2.5 新增 `cmd/iac-server/service/service_test.go` 单元测试，使用 `httptest` 验证 `/alivez`、`/healthz`（正常态与关停态 503 `errf.UnHealthy`，覆盖 AC-005、AC-006）以及访问 `/api/v1/iac/non-existent-path` 安全返回 404（覆盖 AC-008）

## 3. 内部客户端 SDK 与上下游服务发现/网关反向代理集成（`iac-server-client-and-proxy`）

- [x] 3.1 新增 `pkg/client/iac-server/client.go`（定义 `Client` 与 `NewClient`，基础路径 `/api/%s/iac`），在 `pkg/client/api.go` 的 `ClientSet` 中新增 `IacServer() *iacserver.Client` 方法，并补充客户端与未授权服务发现拦截单测（覆盖 AC-003、AC-004）
- [x] 3.2 在 `cmd/cloud-server/app/app.go` 的 `DiscoveryOption.Services` 中追加 `cc.IacServerName`
- [x] 3.3 在 `cmd/api-server/app/app.go` 的 `DiscoveryOption.Services` 与 `cmd/api-server/service/proxy.go` 的 `discoverServices` 中追加 `cc.IacServerName`，并在 `prepareRequest` 的 `switch servicePath` 中新增 `case "iac": service = cc.IacServerName`
- [x] 3.4 在 `cmd/web-server/app/app.go` 的 `DiscoveryOption.Services` 与 `cmd/web-server/service/proxy.go` 的 `discoverServices` 中追加 `cc.IacServerName`，并在 `proxyRequest` 的 `switch servicePath` 中新增 `case "iac": service = cc.IacServerName`

## 4. 构建脚本、启停脚本、Dockerfile 与 Helm 部署模版集成（`iac-server-build-and-deploy`）

- [x] 4.1 新增 `cmd/iac-server/Makefile`（配置 `SERVER = bk-hcm-iacserver`、`ARGS = -c ./etc/iac_server.yaml`，实现 `default`、`package`、`docker`、`clean` 目标），并在 `cmd/Makefile` 的 `all`、`package`、`docker`、`clean` 目标中追加 `iac-server`
- [x] 4.2 在 `scripts/install/start_all.sh` 与 `scripts/install/stop_all.sh` 中追加 `iac-server` 启停命令，并新增 `docs/support-file/docker/bk-hcm-iacserver/Dockerfile`
- [x] 4.3 更新 `docs/support-file/helm/values.yaml`（新增 `iacserver` 配置段）、`docs/support-file/helm/templates/bklogconfig.yaml`（追加 `iacserver` 日志采集路径），并新增 `docs/support-file/helm/templates/iacserver/configmap.yaml`、`deployment.yaml`、`service.yaml`

## 5. 自动化编译、单测与配置交付物自检（代码实现阶段自动化验证）

- [x] 5.1 执行 `go build ./...` 与 `make -C cmd/iac-server`，验证无 MySQL 依赖下全量编译通过、单服务编译耗时 < 30s，且成功生成 `bk-hcm-iacserver` 二进制与启停脚本（AC-007、AC-P02）
- [x] 5.2 执行 `go test -v ./pkg/cc/... ./cmd/iac-server/... ./pkg/client/... ./cmd/api-server/... ./cmd/web-server/...`，确认新增与存量单元测试 100% 通过（覆盖 AC-001 ~ AC-006、AC-008）
- [x] 5.3 检查 `cmd/iac-server/etc/iac_server.yaml` 与 `docs/support-file/helm/` 配置结构一致性及敏感信息红线（若环境已安装 `helm` CLI 则执行 `helm template docs/support-file/helm` 校验 YAML 渲染语法，覆盖 AC-007、AC-S01）

> **说明（需外部真实环境/人工参与的联调验证项，归属阶段 6「发布到环境」或集成测试环境执行，不阻塞阶段 4 本地编码勾选）**：
> - 在已部署真实 Etcd 实例的环境中启动 `bk-hcm-iacserver`，验证 Etcd `/hcm/services/iac-server/` 注册键值生成、冷启动耗时 < 5s（AC-P02）及发送 `SIGTERM` 后键值注销（AC-003）。
> - 在单实例（200m CPU / 512Mi Memory）容器或测试环境中使用 `wrk -t4 -c100 -d30s` 压测 `GET /alivez` 与 `GET /healthz`，验证 P99 < 50ms（AC-P01）。
