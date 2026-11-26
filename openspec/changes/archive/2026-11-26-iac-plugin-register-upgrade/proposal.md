## Why

IaC 插件注册表（`iac_plugin`）与 Schema 存储（`iac_plugin_schema`）就绪后，需要打通 TF Provider 与 HTTP 两类插件的完整注册、升级、运行期校验及管理接口闭环，支撑后续执行态按插件版本安全调度。

## What Changes

- **TF Provider 启动注册与运行期版本检查**：`hc-service` 启动时扫描本地插件目录、校验本架构可执行文件 SHA256、持插件名锁串行注册/升级并同步 Schema；就绪前 `/healthz` 返回 503 且不注册 etcd；运行期按周期检查本副本版本是否落后，连续 3 周期落后则退出。
- **HTTP 插件注册、升级、修改与巡检**：`cloud-server` 提供注册、升级、修改接口并转发至 `hc-service`；`hc-service` 经蓝鲸 API 网关完成地址与 RSA-3072 公钥校验、探活、Schema 同步与收尾；master 节点周期巡检自动发现新版本或标记连续失败。
- **插件管理查询与启停**：`cloud-server` 提供 `list` 与 `switch` 接口，按插件名下最高版本判定启停规则，持同一把插件名锁与注册/升级串行；`disabled → enabled` 前校验当前版本 Schema 已存在。
- **配置与部署**：补齐 `hc-service` 插件目录、etcd 锁、巡检及网关配置，更新 Dockerfile、Helm chart 与 5 篇接口文档。

## Capabilities

### New Capabilities

- `iac-plugin-register-upgrade`：涵盖 TF Provider 自动发现/注册/版本监控、HTTP 插件探活/注册/升级/修改/巡检，以及插件列表与启停管理。

## Impact

- `cmd/hc-service`：插件扫描、注册、版本监控、HTTP 编排与巡检、启动就绪门控与内部路由。
- `cmd/cloud-server`、`cmd/auth-server`、`pkg/iam/meta`：管理与 HTTP 转发接口、权限映射（`iac_plugin` → `global_configuration`）。
- `pkg/iacplugin`、`pkg/lock`、`pkg/cc`、`pkg/api`、`pkg/client`：semver 版本比较与状态继承、路径/公钥校验、etcd 插件锁、配置与协议定义。
- `docs/api-docs/web-server/docs/scr/`、`docs/support-file/`：接口文档与部署配置。
