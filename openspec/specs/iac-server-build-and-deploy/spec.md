# iac-server-build-and-deploy Specification

## Purpose
定义 `iac-server` 的 Makefile 编译打包脚本、二进制启停脚本（`start_all.sh` / `stop_all.sh`）、Dockerfile 镜像构建以及 Helm Chart 部署模版与日志采集配置规范。

## Requirements
### Requirement: Makefile, Start/Stop Scripts, and Dockerfile Integration
系统 SHALL 提供 `cmd/iac-server/Makefile`（定义 `SERVER = bk-hcm-iacserver`、`ARGS = -c ./etc/iac_server.yaml`，支持 `default`、`package`、`docker`、`clean` 目标），并在 `cmd/Makefile`、`scripts/install/start_all.sh`、`scripts/install/stop_all.sh` 中集成 `iac-server`，同时提供 `docs/support-file/docker/bk-hcm-iacserver/Dockerfile`。

#### Scenario: Build binary and generate start/stop scripts without MySQL dependency (AC-007, AC-P02)
- **WHEN** 在无 MySQL 实例的环境下执行 `go build ./...` 与 `make -C cmd/iac-server`
- **THEN** 编译在 30 秒内完成且退出码为 `0`，在构建输出目录中生成 `bk-hcm-iacserver` 二进制文件、`etc/iac_server.yaml` 配置文件及 `iac-server.sh` 启停脚本，且 `scripts/install/start_all.sh` 与 `stop_all.sh` 包含 `iac-server` 启停命令

### Requirement: Helm Chart Templates and Values Synchronization
`docs/support-file/helm` SHALL 与 `cmd/iac-server/etc/iac_server.yaml` 保持配置结构同步，在 `values.yaml` 中新增 `iacserver` 配置段（镜像 `blueking/bk-hcm-iacserver`、容器端口 `80`、副本数、资源限制、日志及 Service 配置），新增 `templates/iacserver/configmap.yaml`、`templates/iacserver/deployment.yaml`、`templates/iacserver/service.yaml`，并在 `templates/bklogconfig.yaml` 的文件日志采集路径列表中追加 `{{ .Values.iacserver.log.logDir }}/*`。

#### Scenario: Helm chart contains complete iacserver manifests and log collection path (AC-007, AC-S01)
- **WHEN** 检查或渲染 `docs/support-file/helm` 下的 `iacserver` 模版与 `values.yaml`
- **THEN** 渲染结果包含 `*-iacserver-config` ConfigMap（含 `network`、`service.etcd`、`log`、`tenant` 四段且无任何明文敏感凭证）、`*-iacserver` Deployment（含等待 `dataservice` 的 `initContainers`、`/alivez` 存活探针、`/healthz` 就绪探针、`POD_IP` 环境变量及外部 Etcd TLS 证书条件挂载）、`*-iacserver` Service 以及包含 `iacserver` 日志路径的 `BkLogConfig`
