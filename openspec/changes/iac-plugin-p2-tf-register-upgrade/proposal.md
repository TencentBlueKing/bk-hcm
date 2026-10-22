## Why

tf_provider 插件随 hc-service 镜像或挂载目录交付，需要在 hc-service 启动时自动扫描、注册，并在启动时发现更高版本二进制后自动升级；同时要保证滚动发布期间新旧副本并存可用、注册失败时不让坏副本接入流量。没有这套流程，插件注册表里就没有可用记录，后续的管理接口与运行时都无从谈起。

技术方案：《插件注册与升级》（iWiki 4047929736）第一章 1.1 / 1.2、第二章 2.2、第三章 3.1~3.8、第六章；评论：二进制部署的旧副本退出检查、审计要记。

## What Changes

- 新增目录检查：按 `{插件根目录}/{插件名}/{版本}/{架构}/` 约定扫描，每个插件取最高版本（semver）与本副本架构的唯一可执行文件，计算 sha256，结果作为只读的副本内存信息
- 新增启动注册流程：目录检查 → 启动 HTTP 服务（暂不注册 etcd，`/healthz` 返回 503）→ 逐个插件持 etcd 锁处理 → 检查"注册表有记录、目录里没有"的插件 → 注册 etcd、`/healthz` 返回 200
- 新增注册 / 升级判定：V 低于已完成最高版本 M 则退出；V 记录不存在则插入 `init`；`init` / `failed` 重入；`enabled` / `disabled` 只校验 sha256
- 注册时通过对接接口 `ProcessStarter` 取得 `ProcessManager`，`Start` 拿到 `Executor`，把 `GetProviderSchema` 适配成 `SchemaGetter` 传给 `SchemaSyncer.Sync`（对接方），无论成败都由 `ProcessManager.Close`；成功后继承上一版本状态，失败置 `failed` 并退出
- 新增 sha256 校验：只算本架构 —— 首次注册写入本架构 sha256，同版本启动时比对本架构，不一致则退出
- 失败一律退出：目录检查失败不启动服务；注册失败置 `failed` 后退出；与已有记录冲突（版本回退、sha256 不一致）不修改任何记录并退出。进程内不做错误重试，交给编排重启
- 新增运行期版本检查（每个副本各自执行）：本副本二进制版本低于注册表已完成最高版本，连续 3 个检查周期后退出
- 明确部署契约：变更插件目录后必须重启 hc-service，新版本才会被该副本注册
- 新增注册 / 升级 / 状态变化的审计（操作人 `system`）
- 改造 hc-service `/healthz`：注册完成前返回 503，`/alivez` 保持 200
- **不包含**：http 插件注册与升级（P4）、管理接口与启停（P3）、**tf 进程实例管理（go-plugin 拉起 / 握手 / gRPC / 关闭 / 日志 / 退出清理，对接方）**、各副本对 `disabled` 的进程池响应（运行态）、Schema 同步实现（对接方）、进程池

## Capabilities

### New Capabilities

- `iac-plugin-tf-discovery`：目录约定与目录检查、sha256 计算、副本内存信息
- `iac-plugin-tf-register`：启动注册与升级流程、状态继承、失败处理、启动时序与探针、审计
- `iac-plugin-tf-version-watch`：运行期版本检查，低版本副本退出

### Modified Capabilities

（无）

## Impact

- **cmd/hc-service/logics/iac-plugin/tfplugin**（新增）：`discovery`、`register`、`watch`
- **cmd/hc-service/app/app.go**、**cmd/hc-service/service/service.go**：启动时序改造（就绪标志、`/healthz` 503、注册 etcd 延后）
- **pkg/cc 与 cmd/hc-service/etc**：新增插件根目录（默认 `/data/hcm/plugins`）、锁租约、锁等待时限、版本检查周期（默认 60 秒）配置项；helm configmap / values 同步
- **docs/support-file/docker**：hc-service 镜像支持放入插件目录；helm 增加插件目录挂载项（键名开发时确定）
- 使用 P1（data-service 客户端、状态机、semver、etcd 锁、`integration` 对接接口）
- 无表结构变化

## 依赖与后续

- 前置：P1。
- 对接依赖：对接方提供 `SchemaSyncer` 与 `ProcessStarter`/`ProcessManager`/`Executor` 的真实实现。
- 退出清理：hc-service 退出时关闭存活插件进程由对接方的进程管理负责；本 change 只保证注册流程自己拉起的进程一定 `Close`。
- 被依赖：P3（启停对 tf 的约束）。
