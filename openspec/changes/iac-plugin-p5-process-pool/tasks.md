## 0. 分支准备

- [x] 0.1 `git fetch upstream`，从 `upstream/feat-plugin` 切出 `feat/iac-plugin-process-pool`（本 change 的 openspec 目录为未跟踪文件，切分支时随工作区带过去）；确认基线没有 `pkg/lock`、`pkg/api/core/plugin`、`cmd/hc-service/logics/plugin`
- [x] 0.2 确认基线可用：`pkg/kit`、`pkg/logs`、`pkg/metrics`，`go.mod` 已含 `terraform-json`；`go build ./cmd/hc-service/...` 在基线上通过
- [x] 0.3 对照 P2 分支（`feat/iac-plugin-register-submit`）的 `register/types.go`，核对 `Starter` 签名一致；不一致则以 P2 为准调整契约
- [x] 0.4 阶段零到三的约束：只新增文件，不修改既有文件、不改 `go.mod` / `go.sum`；每个阶段收尾用 `git diff --stat upstream/feat-plugin` 确认（阶段四不受此限）

## 阶段零：对接契约骨架（`logics/tfruntime`）

> 完成标准：`go build ./...` 通过；桩方法一律返回 `ErrNotImplemented`；不新增第三方依赖；无既有文件被修改。只有类型和桩，没有协议逻辑。

## 1. 契约类型

- [x] 1.1 新建 `cmd/hc-service/logics/tfruntime`：`errors.go`（`ErrNotImplemented`、`DiagnosticsError`，`Error()` 不含 Config 内容）
- [x] 1.2 `types.go`：`Severity`、`Diagnostic`、`AttributePath`、`Object`（`map[string]any`，nil 为 null）、`Unknown`、`ProviderSchema`（`pkg/thirdparty/api-gateway/plugin.PluginSchema` 的类型别名）；对照对接方案 6.2 逐个方法生成 `XxxRequest / XxxResponse`（`ConfigureProvider`、`ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`），`Private` 用 `[]byte`
- [x] 1.3 `executor.go`：`Executor` 为完整方法集（`GetProviderSchema`、`ConfigureProvider` + 8 个业务方法），不拆子接口，首参均为 `*kit.Kit`
- [x] 1.4 `process.go`：`ProcessConfig`、`ProcessManager` 接口（`Start / Close(ctx) / Exited / CheckHealth`）、`processManager` 结构体与 `NewProcessManager(cfg)` 工厂、`Starter` 接口与 `defaultStarter`；桩方法按 spec 返回 `ErrNotImplemented`，`Close` 幂等；包级编译期断言
- [x] 1.5 单测只覆盖极小的纯逻辑：桩的返回语义（`Start` 返回 `ErrNotImplemented`、`Close` 幂等、`Exited` 为 false）、`DiagnosticsError` 经 `errors.As` 可区分、Warning 不构成错误
- [ ] 1.6 请对接方评审本阶段产物（design 的 Open Questions 第一条），结论写回 design.md 后再进入阶段一

## 阶段一：最小闭环（调度与交付）

> 完成标准：假 `ProcessManager` 下单测通过；仅新增包，无调用方。

## 2. 池骨架

- [x] 2.1 新建 `cmd/hc-service/logics/plugin/tf/pool`：错误变量（`ErrPoolClosed`、`ErrTooManyWaiters`）；池直接引用 `tfruntime.ProcessManager` / `Executor` / `Starter`；池包内定义窄接口 `ProviderIF`（仅 8 个业务方法）供 `Lease` 实现
- [x] 2.2 `PoolConfig` 结构体、默认值与校验（`maxProcesses`、`maxLeasesPerProcess`=1、`closeTimeout`、`maxQueueLength`、`maxQueueWait`、`startTimeout`、`configureTimeout`、`maxConfigAge`），构造池时传入，不依赖 `pkg/cc`
- [x] 2.3 `Config.Key()`：字段排序后的 JSON → SHA-256，口径集中在一个函数；日志只输出 key 前缀

## 3. 队列与数据结构

- [x] 3.1 `sliceFIFO`（`PushBack / Peek / PopFront / Remove / Len / Drain`），取消延迟移除，失效元素多于有效元素时压缩
- [x] 3.2 `Pool`、`PluginProcess`、`Binding`（`configValid`）、`request`、`Lease` 数据结构；进程 ID 单调不复用；`Lease` 持有 `tfruntime.Executor` 并实现池包的 `ProviderIF`：逐个转发 8 个业务方法，不暴露 `ConfigureProvider` / `GetProviderSchema`；lease 已归还时转发前直接返回错误
- [x] 3.3 单测：FIFO 顺序、取消后 `Len`、队首失效元素被跳过、`Drain` 后为空、压缩后顺序不变；`configValid` 的 nil / 配置中 / 已生效 / 已过期

## 4. 调度与交付

- [x] 4.1 `Acquire`：入队前检查池关闭与队列上限；入队后触发调度；等待结果、`maxQueueWait` 与 `kt.Ctx` 超时
- [x] 4.2 `pickLocked`：单次遍历进程表，规则 1（匹配可交付）、2（未配置 / 已过期）、3（重配）、4（新建）、5（无）；同类候选按 `inflight` 少者、平局按 ID
- [x] 4.3 `scheduleLocked`：整轮只取一次 `now`；取一个派一个，选不出结束本轮
- [x] 4.4 `deliverLocked`、`cancel`（队列中 / 已交付）、`Release`（幂等、进程不存在忽略、归 0 记 `idleSince`、触发调度）
- [x] 4.5 表驱动单测（假进程表）：规则优先级、过期后降级为规则 2、正在配置不交付、有在途 lease 不重配、队首派不出去后面不越队、`inflight` 不超上限

## 5. 异步创建、配置与写回

- [x] 5.1 `startCreateLocked` / `startConfigureLocked`：锁内改状态（`starting` / `configuring`、写 `binding`、`pending=[w]`），锁外调用 `Starter` → `Start` / `ConfigureProvider`，使用池 ctx + 超时
- [x] 5.2 写回：核对进程仍在表中与 `pending` 归属；成功设 `boundAt`、交付 `pending`、转 `running`；失败（阶段一语义）转 `closing`、请求直接失败；末尾触发一次调度
- [x] 5.3 `recover` 兜底：panic 按传输错误处理且写回一次；`wg` 计数；`Start` / `ConfigureProvider` 加看门狗（不响应 ctx 时超时即写回并关进程）
- [x] 5.4 假 `ProcessManager` / `Executor`（放在测试文件内）单测：配置成功交付、`Start` 失败（含 `ErrNotImplemented`）、`Configure` 超时、请求取消后进程保留、写回时进程已不在

## 6. 关闭

- [x] 6.1 `Pool.Close(ctx)`：置关闭标志、`queue.Drain()` 全部返回 `ErrPoolClosed`、空闲进程转 `closing`、有 lease 的进程调 `ProcessManager.Close(ctx)`、`starting` / `configuring` 等写回后关闭、ctx 超时强制结束
- [x] 6.2 单测：排空队列、创建中关闭、关闭后 `Acquire` 立即失败、`closing` 进程归还不回到空闲
- [x] 6.3 阶段一收尾：`go test -race` 与 `go vet` 通过；不变式 I1、I2、I3、I6、I7、I10 在测试中以断言函数检查

## 阶段二：并发利用率与自愈

> 完成标准：不变式 I1~I10 全部有断言；失败表（4.6）逐行有用例；基准测试跑通。

## 7. 准备期预约

- [x] 7.1 `pickLocked` 增加规则 1b（`starting` / `configuring`、key 一致、`boundAt` 为空、`len(pending) < maxLeasesPerProcess`），优先级位于直接交付之后、空闲配置之前；`dispatchLocked` 增加预约分支
- [x] 7.2 配置成功批量交付：先设 `boundAt`，按 `pending` 顺序交付、跳过已取消，清空 `pending`
- [x] 7.3 单测：同配置突发共享一个进程、预约已满转后续规则、预约不越队、已取消预约被跳过

## 8. 失败分流与重试

- [x] 8.1 `failPendingLocked` / `retryOrFailLocked`：取走 `pending` 并置空、保持原相对顺序回队尾、`attempts > maxRetries` 或池关闭则失败
- [x] 8.2 写回按错误类型分流：`errors.As(*tfruntime.DiagnosticsError)` 为业务错误 → `running` + `binding=nil` + 整批失败；`Start` 失败 / 超时、`Configure` 超时 / 传输错误、准备期进程退出 → `closing` 或删记录 + 重试
- [x] 8.3 `Start` 失败后确认无残留（`Exited()` 为真，或 `Close(ctx)` 强制结束后确认），确认不了转 `closing`
- [x] 8.4 取消补全：请求在 `pending` 中只置 `done`，交付与回收时跳过；取消触发调度
- [x] 8.5 单测：失败表逐行（业务错误整批失败、传输错误回队尾保序、重试耗尽、重复回收不重复入队）、日志与错误不含 config

## 9. 定时巡检

- [x] 9.1 `idleCountLocked`：按统一口径统计空闲数（`running && inflight==0`，不论是否已配置 / 过期；加上未被认领的预热 `starting`）
- [x] 9.2 池内 ticker（`poolCheckInterval`）：`idle < min` 补足未配置进程（容量有余才补，只在巡检时补建，池满不挤占）；`idle > min` 才回收，只回收空闲时长超过 `idleReclaimTimeout` 的进程，空闲最久优先，一轮最多回收 `idle - min` 个
- [x] 9.3 `Exited()` 检查（删记录、释放容量、回收 `pending`、触发调度）与 `CheckHealth()` 检查（仅 `running`，报错转 `closing`）；每轮末尾无条件触发调度；`Close` 停止巡检
- [x] 9.4 单测：已配置空闲进程计入空闲数、`idle == min` 不回收、`idle=5,min=2` 全部超时只回收 3 个且空闲最久优先、部分未超时只回收超时者、回收后被占用由预热补回、池满不预热、退出进程释放容量并回收 `pending`、`CheckHealth` 失败转 `closing`、兜底调度能唤醒队首
- [x] 9.5 基准测试：`maxProcesses × maxLeasesPerProcess` 取典型值与较大值，度量 `scheduleLocked` 单轮耗时与排队 1000 / 10000 个请求时的 `Peek` 开销
- [x] 9.6 阶段二收尾：`go test -race`、`go vet` 通过；不变式 I4、I5、I8、I9 补断言

## 阶段三：可观测性

> 完成标准：告警与指标可用，日志与指标不含 config / 凭证；全部检查通过。

## 10. 参数校验补全

- [x] 10.1 `PoolConfig` 补齐 `maxRetries`、`idleReclaimTimeout`、`poolCheckInterval` 的默认值与校验（`maxProcesses ≥ 1`、`maxLeasesPerProcess ≥ 1`、`minIdleProcesses ≤ maxProcesses`、时长大于 0）；构造池时校验失败返回错误
- [x] 10.2 单测：缺省值通过校验、非法组合被拒绝

## 11. 指标

- [x] 11.1 ~~lease 持有超时告警~~：已移除（第一版不做，留作备选）
- [x] 11.2 指标（11 个原子指标，5 个事件记录点 + 1 个抓取时 Collector）：`acquire_seconds{result}`、`lease_hold_seconds`、`start_seconds{result}`、`configure_seconds{result}`、`retries_total`；抓取时读取 `queue_length`、`queue_oldest_wait_seconds`、`processes{state}`、`inflight_leases`、`max_processes`、`max_leases_per_process`
- [x] 11.3 复核全部日志、错误与指标标签不含 config / 凭证，只出现 key 前缀

## 12. 收尾

- [x] 12.1 `go build ./...`、`go vet`、新增包 `go test -race` 通过
- [x] 12.2 `git diff --stat upstream/feat-plugin` 确认只有新增文件，`go.mod` / `go.sum` 无变化
- [x] 12.3 阶段三完成。**此处停止**：阶段四必须等分支与注册 / Schema 分支（P1~P4）合并后才能开工

## 阶段四：接入与运维（合并后）

> 前置：阶段三完成，且 `feat/iac-plugin-process-pool` 已与 P1~P4 所在分支合并。完成标准：联调环境启停插件可建池 / 关池，服务退出无残留 provider 进程；开关默认关闭，关闭时行为与接入前一致。

## 13. 合并适配

- [x] 13.1 核对前置：集成分支从 `feat/iac-plugin-register-submit` 切出，已覆盖阶段零到三的代码；`IaCPluginConfig`、`service/iac_plugin.go`、插件注册表的 data-service 客户端均存在；构建与相关包测试通过
- [x] 13.2 `tfruntime.ProviderSchema` 别名改指 `pkg/api/core/plugin.PluginSchema`
- [x] 13.3 `register` 删除自带的三个对接接口，改用 `tfruntime.Starter` / `ProcessManager` / `Executor`；`Close` 带 30 秒超时的 ctx；注册、版本检查逻辑不动

## 14. 配置项

- [x] 14.1 `pkg/cc` 的 `IaCPluginConfig` 增加 `Pool`（`IaCPluginPoolConfig`，含 `enabled` 开关，默认关闭、`registryCheckIntervalSeconds` 默认 5）；`trySetDefault` / `Validate` 补默认值与负数校验，取值范围由 `pool.NewPluginRegistry` 复用 `PoolConfig` 校验，不重复一套
- [x] 14.2 `hc_service.yaml`、helm `values.yaml` 同步（configmap 对 `iacPlugin` 整体透传，无需改模板），注释写明含义与取值约束；两个文件已验证可解析为同一份配置
- [x] 14.3 配置解析验证已做；`helm template` 渲染未验证（本机没有 helm），渲染由 configmap 的 `toYaml` 透传，风险低，部署前建议补一次

## 15. 注册表与服务接入

- [x] 15.1 `pool.PluginRegistry`：`GetPool(pluginName)`（未建池返回 `ErrPluginNotEnabled`）、按本副本已启用的 tf 插件建池、读取失败或版本号不合法时跳过本轮、停用时同步关池后再允许重建、同一插件至多一个池
- [x] 15.2 后台同步：`StartSync` 立即同步一次再按周期同步；`Close` 取消后台同步、关闭全部池并等待协程退出，可重复调用；读插件表只查本副本插件并翻页读全；建池后立即预热；停用池的关闭不受后台同步 context 影响
- [x] 15.3 `service/iac_plugin.go`：`initPluginPools`（`initIaCPlugin` 中创建并校验）、`StartPluginPools`（`SetPluginReady` 之后启动）、`ClosePluginPools`（finalizer 中 `defer`，摘除服务发现之后关闭，总时限 10 秒）；注册、版本检查逻辑不动
- [x] 15.4 开关关闭或没有本地 tf 插件时不创建注册表，`StartPluginPools` / `ClosePluginPools` 为空操作；对接方真实 `ProcessManager` 就绪前保持关闭（桩的 `Start` 返回 `ErrNotImplemented`）
- [x] 15.5 单测（生产代码不为测试留钩子，所以依赖 data-service 与后台协程的部分不做单测，靠联调验证：读取失败 / 版本号非法时跳过本轮不关池、翻页与 `IN` 过滤、`Start` 幂等、`Close` 等协程退出）：覆盖的是不依赖外部的部分：`enabledPlugins`（已完成最高版本判定，8 个场景加非法版本号）、`applyEnabled`（建池、停用摘除、本副本没有的插件不建池、关闭后新建的池交还调用方）与 `Close` 的生命周期、建池即预热、`NewPluginRegistry` 对非法参数的拒绝

## 16. 联调与验证

- [ ] 16.1 与对接方真实 `ProcessManager` 联调：首次申请创建并配置、同配置复用、换配置重配、配置过期重配、业务错误整批失败、进程被 kill 后自愈、空闲回收与预热补回
- [ ] 16.2 并发验证：同配置突发共享一个进程（`maxLeasesPerProcess>1`，前提确认 provider 支持并发 RPC）、不同配置在容量满时排队保序
- [ ] 16.3 停用插件与服务退出：排队请求收到池已关闭错误、无残留 provider 进程
- [ ] 16.4 开关关闭回归：与接入前对比启动时序、`/healthz`、进程列表一致
- [x] 16.5 `go build ./...` 通过；变更涉及的包 `go vet` 与 `go test -race` 通过（`service/account` 的 vet 报错是基线上就有的，与本 change 无关）
