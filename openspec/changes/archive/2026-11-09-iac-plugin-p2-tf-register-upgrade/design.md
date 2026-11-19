## Context

方案（iWiki 4047929736 第三章）要点：每个 hc-service 副本启动时各自扫描目录并执行注册判断（滚动发布时 master 可能是旧副本，本地目录里没有新版本二进制，所以不能由 master 统一处理）；同一插件的注册 / 升级按插件名加 etcd 锁串行；注册失败时副本退出，使滚动发布停住、旧副本继续服务。

仓库现状（`cmd/hc-service`）：
- `app.Run` 顺序是 `prepare`（加载配置、建 `serviced`、`NewService`）→ `ListenAndServeRest` → `register`（注册 etcd）；`/healthz` 当前对 etcd 做检查，没有"就绪标志"；
- `service.NewService(sd)` 能拿到服务发现，但 `Capability` 里是否带 `sd` 未确认；
- 无定时任务先例；`pkg/cron` 目前只有 task-server 使用；
- Dockerfile 只有 `COPY bk-hcm-hcservice /data/hcm/`，base 镜像无 `plugins/` 目录。

依赖：P1（data-service 客户端、状态机、semver、`pkg/lock`、`integration` 对接接口）。Schema 同步（`SchemaSyncer`）与 tf 进程实例管理（`ProcessStarter`，含 go-plugin 拉起、握手、gRPC、关闭、日志、退出清理）均由对接方提供，本 change 只通过接口调用。

## Goals / Non-Goals

**Goals:**
- 启动流程串行、可重入：任意阶段崩溃后重启都能收敛到正确状态。
- 失败默认"停在未就绪"，不让没有可用插件的副本接入流量。
- 滚动发布窗口内新旧副本互不影响（旧版本记录不被修改）。

**Non-Goals:**
- http 插件、管理接口、启停响应、进程池。
- Schema 同步实现与校验。
- 新旧版本窗口内"新增资源类型的请求落到旧副本"的校验失败（方案已知限制）。

## Decisions

### D1 注册流程放在 `app.Run` 里，不放 `NewService`
调整为：`prepare` → `ListenAndServeRest`（`/healthz` 返回 503）→ **插件注册** → `register`（etcd）→ 启动版本检查任务。注册失败直接 `return err`，由 `main` 以非 0 退出。就绪标志放在 `Service` 内（`atomic.Bool`），`Healthz` 在未就绪时返回 503 且不检查 etcd；`Alivez` 不变。

备选：在 `NewService` 内同步注册。放弃：此时 HTTP 尚未监听，`/alivez` 无法响应，注册耗时较长时会被 Kubernetes 重启。

### D2 目录检查与注册分包
`tfplugin/discovery` 只做文件系统扫描与 sha256（纯函数，易测）；`tfplugin/register` 做状态流转；`tfplugin/watch` 做运行期版本检查。注册流程对外依赖：data-service 客户端、锁、`SchemaSyncer`、`ProcessStarter`，其中 `SchemaSyncer` 与进程管理为对接接口。判定、版本比较、sha256 校验等纯逻辑用单测覆盖，I/O 编排留给联调。

### D3 注册判定写成纯函数
输入"该插件名下全部记录 + 目录条目"，输出动作枚举（`Exit`、`InsertInit`、`ResetToInit`、`VerifyHash`）。流程图里的全部分支用表驱动单测覆盖，I/O 部分只负责执行动作。这样方案图的每个分支都有可验证的对应。

### D4 sha256 只算本架构
副本启动时只对本架构（`{os}_{arch}`）的可执行文件算 hash，首次注册（持锁）时也只算本架构，不遍历其它架构目录。大文件（数百 MB）用流式 `io.Copy` 到 `sha256`，不一次性读入内存。

代价：`extension.architectures` 只记录实际启动过的架构 —— arm64 的 hash 要等首个 arm64 副本启动才写入；某架构的二进制被替换，只有该架构的副本启动时才能发现。混部架构下可以接受，换来注册路径不做无用的全目录扫描。

### D5 只等锁，不做错误重试
锁键 `iac_plugin/<插件名>`；租约默认 10 分钟，大于注册最长耗时（握手 + `Sync` 超时之和），可配置。

`Lock` 拿不到锁时监听 key 删除后再抢，不轮询。这里的等待是**跨副本串行化的正常排队**，不是错误重试，但必须有上限：持锁副本崩溃后要等租约到期锁才可用（最长 10 分钟），不设上限会让启动无限挂住、只能被 startupProbe 杀掉。上限 `lockWaitTimeout`（默认 2 分钟）到了就按注册失败处理、主动退出。

**不做临时性错误重试**，理由：
- helm 里 hc-service 已有 initContainer 等 dataservice Ready，不存在"data-service 还没起来"的窗口；
- etcd 不可用在 `prepare` 阶段就失败了，进程直接退出，**k8s 的退避重启就是重试**；
- "没有业务错误码即可重试"这种判定不可靠：JSON 解析失败、500、路由不存在都会被归进来；
- 设计要求"失败即退出、让滚动发布停住"，进程内先等几分钟只会推迟故障信号。

### D6 失败时先写库再退出
确定性注册失败：在持锁状态下条件更新（`expect_state=init`）为 `failed` 并写 `message`（截断到 255），释放锁，返回错误。写库本身失败时保持 `init`，同样退出。下一个拿到锁的副本按"`init`/`failed` 一律重入"处理。

### D7 `integration` 是"待对接接口"包，不是长期架构中心
`pkg/iacplugin/integration` 当前只承载**尚未对接**的接口定义（`SchemaGetter`、`SchemaSyncer`、`ProcessStarter`、`ProcessManager`、`Executor`）与占位 `PluginSchema`。它是过渡形态：对接方交付后，真实实现落在各自的包 —— 进程实例管理归 hc-service 运行时包，Schema 同步归 Schema 对接方自己的包，http 调用归 cloud-server 的 httpplugin 包；届时接口可按 Go 惯例挪到消费方，本包逐步收窄甚至删除。**本包不放任何实现。**

注册流程只依赖 `ProcessStarter`：`NewProcessManager` 返回 `ProcessManager`，其 `Start` 返回 `Executor`（`GetProviderSchema` 返回装配好的统一 Schema），`Close` 在 `ProcessManager` 上。进程实例管理与 tf 的 `GetSchema`（含转写、摘要）均由对接方提供，本 change 不引入 go-plugin / tfprotov5 等依赖。联调前用假实现。注册流程的责任只有一条：`Start` 成功后，无论后续成败都必须 `Close`（`defer`）。

### D8 运行期版本检查：每个副本各自的 ticker
启动时扫描只能解决"本次启动的二进制是什么版本"，解决不了"运行中被别的副本超车"：二进制部署没有滚动替换语义，副本 A 被重启并注册了新版本后，仍用旧二进制运行的副本 B 不会自己发现，会一直用旧版本处理请求，注册表与实际不一致。

因此加一个副本本地检查：ticker 周期（默认 60 秒）读取注册表中每个插件已完成（`enabled` / `disabled`）的最高版本 M，与本副本目录里的版本 V 比较；V < M 连续 3 个周期则主动退出，交给编排重建。要点：

- 必须是每个副本各自执行，不能只挑 master —— 比较的是"本副本的本地二进制"，每个副本的结论不同；http 插件的探活与巡检是另一回事，在 P4 由 master 执行。
- 只读注册表，不加锁、不写库、不写审计，不能干扰正在进行的注册。
- 连续 3 次而非 1 次，避开"新版本刚注册、旧副本仍在正常服务"的短暂窗口。
- 读注册表失败只记日志、跳过本轮、不计数，避免把 data-service 抖动误判成版本落后；版本恢复一致则清零。
- 计数放内存，进程重启或与注册流程的 `init` 重入都会自然收敛。

hc-service 没有定时任务框架先例，且这是副本本地行为，不引入 `pkg/cron`；用 `time.Ticker` + `ctx`（跟随 shutdown 通知）实现。

### D9 审计由 data-service 审计接口落库
hc-service 以 `system` 为操作人调用 data-service 的审计创建能力。需先确认 hc-service 已有可用的审计客户端；没有则通过 data-service 现有审计路由补一个调用封装（见 Open Questions）。

### D10 镜像与 chart 只加目录，不加业务逻辑
Dockerfile 增加 `COPY plugins /data/hcm/plugins/`（目录可为空，文件需可执行权限）；base 镜像不改。helm values 增加插件挂载项（卷与挂载路径 `/data/hcm/plugins/<插件名>`），键名开发时确定。

## Risks / Trade-offs

- [注册失败的副本反复被 Kubernetes 重启] → 已接受：确定性错误会一直失败到被修复，重启间隔由 k8s 自身退避，滚动发布超时后判失败，旧副本继续服务。
- [租约小于实际注册耗时，锁提前释放，并发保护失效] → 租约默认值显著大于超时之和，条件更新作为第二道保护。
- [`Sync` 对接方迟迟不交付] → 联调前用临时桩让注册明确失败；真实 `Sync` 交付后仅替换一个依赖。
- [大文件 sha256 拖慢启动] → 只算本架构的单个可执行文件，流式计算，不读入内存；已注册版本启动时同样只算一次。
- [热更新插件目录但不重启 hc-service] → 该副本自己的注册流程不会发现（它不重扫目录），但只要有任一副本重启并注册了新版本，它在 3 个检查周期内会因版本落后退出。
- [`/healthz` 改造影响其他逻辑] → 仅在"未就绪"分支新增 503，其余逻辑不变。

## Migration Plan

1. 先合入镜像与配置（目录为空时流程为空操作，且"注册表无记录"不触发缺失退出），验证不影响现网启动。
2. 再放入插件并发布，观察首次注册。
3. 回滚：回退 hc-service 版本；注册表记录保留，不影响旧版本。

## Open Questions

- hc-service 现有审计调用能力：已有则直接用，没有则补封装（任务 7.1 先确认）。
- 锁租约、锁等待、版本检查周期的默认值：先取 10 分钟 / 2 分钟 / 60 秒，联调后按实际耗时调整。
- `ProcessStarter` 的启动超时、`Close` 的阻塞时长由对接方定义：评审 P1 对接接口时对齐，注册流程的总时限需据此校验。
- hc-service 退出时对注册流程中尚未 `Close` 的进程如何兜底（进程退出信号不会传给子进程）：由对接方的进程登记 / 清理负责，评审时确认。
