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

- [ ] 2.1 新建 `cmd/hc-service/logics/plugin/tf/pool`：错误变量（`ErrPoolClosed`、`ErrTooManyWaiters`）；池直接引用 `tfruntime.ProcessManager` / `Executor` / `Starter`；池包内定义窄接口 `ProviderIF`（仅 8 个业务方法）供 `Lease` 实现
- [ ] 2.2 `PoolConfig` 结构体、默认值与校验（`maxProcesses`、`maxLeasesPerProcess`=1、`minIdleProcesses`、`maxQueueLength`、`maxQueueWait`、`startTimeout`、`configureTimeout`、`maxConfigAge`），构造池时传入，不依赖 `pkg/cc`
- [ ] 2.3 `Config.Key()`：字段排序后的 JSON → SHA-256，口径集中在一个函数；日志只输出 key 前缀

## 3. 队列与数据结构

- [ ] 3.1 `sliceFIFO`（`PushBack / Peek / PopFront / Remove / Len / Drain`），取消延迟移除，失效元素多于有效元素时压缩
- [ ] 3.2 `Pool`、`PluginProcess`、`Binding`（`configValid`）、`request`、`Lease` 数据结构；进程 ID 单调不复用；`Lease` 持有 `tfruntime.Executor` 并实现池包的 `ProviderIF`：逐个转发 8 个业务方法，不暴露 `ConfigureProvider` / `GetProviderSchema`；lease 已归还时转发前直接返回错误
- [ ] 3.3 单测：FIFO 顺序、取消后 `Len`、队首失效元素被跳过、`Drain` 后为空、压缩后顺序不变；`configValid` 的 nil / 配置中 / 已生效 / 已过期

## 4. 调度与交付

- [ ] 4.1 `Acquire`：入队前检查池关闭与队列上限；入队后触发调度；等待结果、`maxQueueWait` 与 `kt.Ctx` 超时
- [ ] 4.2 `pickLocked`：单次遍历进程表，规则 1（匹配可交付）、2（未配置 / 已过期）、3（重配）、4（新建）、5（无）；同类候选按 `inflight` 少者、平局按 ID
- [ ] 4.3 `scheduleLocked`：整轮只取一次 `now`；取一个派一个，选不出结束本轮
- [ ] 4.4 `deliverLocked`、`cancel`（队列中 / 已交付）、`Release`（幂等、进程不存在忽略、归 0 记 `idleSince`、触发调度）
- [ ] 4.5 表驱动单测（假进程表）：规则优先级、过期后降级为规则 2、正在配置不交付、有在途 lease 不重配、队首派不出去后面不越队、`inflight` 不超上限

## 5. 异步创建、配置与写回

- [ ] 5.1 `startCreateLocked` / `startConfigureLocked`：锁内改状态（`starting` / `configuring`、写 `binding`、`pending=[w]`），锁外调用 `Starter` → `Start` / `ConfigureProvider`，使用池 ctx + 超时
- [ ] 5.2 写回：核对进程仍在表中与 `pending` 归属；成功设 `boundAt`、交付 `pending`、转 `running`；失败（阶段一语义）转 `closing`、请求直接失败；末尾触发一次调度
- [ ] 5.3 `recover` 兜底：panic 按传输错误处理且写回一次；`wg` 计数
- [ ] 5.4 假 `ProcessManager` / `Executor`（放在测试文件内）单测：配置成功交付、`Start` 失败（含 `ErrNotImplemented`）、`Configure` 超时、请求取消后进程保留、写回时进程已不在

## 6. 关闭

- [ ] 6.1 `Pool.Close(ctx)`：置关闭标志、`queue.Drain()` 全部返回 `ErrPoolClosed`、空闲进程转 `closing`、有 lease 的进程调 `ProcessManager.Close(ctx)`、`starting` / `configuring` 等写回后关闭、ctx 超时强制结束
- [ ] 6.2 单测：排空队列、创建中关闭、关闭后 `Acquire` 立即失败、`closing` 进程归还不回到空闲
- [ ] 6.3 阶段一收尾：`go test -race` 与 `go vet` 通过；不变式 I1、I2、I3、I6、I7、I10 在测试中以断言函数检查

## 阶段二：并发利用率与自愈

> 完成标准：不变式 I1~I10 全部有断言；失败表（4.6）逐行有用例；基准测试跑通。

## 7. 准备期预约

- [ ] 7.1 `pickLocked` 增加规则 1b（`starting` / `configuring`、key 一致、`boundAt` 为空、`len(pending) < maxLeasesPerProcess`），优先级位于直接交付之后、空闲配置之前；`dispatchLocked` 增加预约分支
- [ ] 7.2 配置成功批量交付：先设 `boundAt`，按 `pending` 顺序交付、跳过已取消，清空 `pending`
- [ ] 7.3 单测：同配置突发共享一个进程、预约已满转后续规则、预约不越队、已取消预约被跳过

## 8. 失败分流与重试

- [ ] 8.1 `failPendingLocked` / `retryOrFailLocked`：取走 `pending` 并置空、保持原相对顺序回队尾、`attempts > maxRetries` 或池关闭则失败
- [ ] 8.2 写回按错误类型分流：`errors.As(*tfruntime.DiagnosticsError)` 为业务错误 → `running` + `binding=nil` + 整批失败；`Start` 失败 / 超时、`Configure` 超时 / 传输错误、准备期进程退出 → `closing` 或删记录 + 重试
- [ ] 8.3 `Start` 失败后确认无残留（`Exited()` 为真，或 `Close(ctx)` 强制结束后确认），确认不了转 `closing`
- [ ] 8.4 取消补全：请求在 `pending` 中只置 `done`，交付与回收时跳过；取消触发调度
- [ ] 8.5 单测：失败表逐行（业务错误整批失败、传输错误回队尾保序、重试耗尽、重复回收不重复入队）、日志与错误不含 config

## 9. 定时巡检

- [ ] 9.1 `idleCountLocked`：按统一口径统计空闲数（`running && inflight==0`，不论是否已配置 / 过期；加上未被认领的预热 `starting`）
- [ ] 9.2 池内 ticker（`poolCheckInterval`）：`idle < min` 补足未配置进程（容量有余才补，失败退避，池满不挤占）；`idle > min` 才回收，只回收空闲时长超过 `idleReclaimTimeout` 的进程，空闲最久优先，一轮最多回收 `idle - min` 个
- [ ] 9.3 `Exited()` 检查（删记录、释放容量、回收 `pending`、触发调度）与 `CheckHealth()` 检查（仅 `running`，报错转 `closing`）；每轮末尾无条件触发调度；`Close` 停止巡检
- [ ] 9.4 单测：已配置空闲进程计入空闲数、`idle == min` 不回收、`idle=5,min=2` 全部超时只回收 3 个且空闲最久优先、部分未超时只回收超时者、回收后被占用由预热补回、池满不预热、退出进程释放容量并回收 `pending`、`CheckHealth` 失败转 `closing`、兜底调度能唤醒队首
- [ ] 9.5 基准测试：`maxProcesses × maxLeasesPerProcess` 取典型值与较大值，度量 `scheduleLocked` 单轮耗时与排队 1000 / 10000 个请求时的 `Peek` 开销
- [ ] 9.6 阶段二收尾：`go test -race`、`go vet` 通过；不变式 I4、I5、I8、I9 补断言

## 阶段三：可观测性

> 完成标准：告警与指标可用，日志与指标不含 config / 凭证；全部检查通过。

## 10. 参数校验补全

- [ ] 10.1 `PoolConfig` 补齐 `maxRetries`、`idleReclaimTimeout`、`poolCheckInterval`、`leaseHoldWarn` 的默认值与校验（`maxProcesses ≥ 1`、`maxLeasesPerProcess ≥ 1`、`minIdleProcesses ≤ maxProcesses`、时长大于 0）；构造池时校验失败返回错误
- [ ] 10.2 单测：缺省值通过校验、非法组合被拒绝

## 11. lease 泄漏告警与指标

- [ ] 11.1 `Lease` 记录 `acquiredAt`；巡检中对持有超过 `leaseHoldWarn` 的 lease 输出一次告警日志并增加指标，不强制回收；归还后不再告警
- [ ] 11.2 指标：排队数、排队时长、各状态进程数、空闲进程数、`inflight` 总数、重试数、失败数（业务 / 传输 / 启动）、队列满拒绝与排队超时数；沿用 `pkg/metrics` 既有习惯，以插件名区分
- [ ] 11.3 复核全部日志、错误与指标标签不含 config / 凭证，只出现 key 前缀
- [ ] 11.4 单测：同一超时 lease 只告警一次、归还后不再告警（用可注入的时钟，不抽象接口）

## 12. 收尾

- [ ] 12.1 `go build ./...`、`go vet`、新增包 `go test -race` 通过
- [ ] 12.2 `git diff --stat upstream/feat-plugin` 确认只有新增文件，`go.mod` / `go.sum` 无变化
- [ ] 12.3 阶段三完成。**此处停止**：阶段四必须等分支与注册 / Schema 分支（P1~P4）合并后才能开工

## 阶段四：接入与运维（合并后）

> 前置：阶段三完成，且 `feat/iac-plugin-process-pool` 已与 P1~P4 所在分支合并。完成标准：联调环境启停插件可建池 / 关池，服务退出无残留 provider 进程；开关默认关闭，关闭时行为与接入前一致。

## 13. 合并适配

- [ ] 13.1 核对前置：分支已合并；`IaCPluginConfig`、`service/iac_plugin.go`、读取插件注册表的 data-service 客户端均存在；`go build ./...` 与两边既有单测通过。任一不满足不得继续
- [ ] 13.2 `tfruntime.ProviderSchema` 别名改指 `pkg/api/core/plugin.PluginSchema`（或加转换），以实际类型为准
- [ ] 13.3 核对 `tfruntime.Starter` 与 P2 `PluginProcessStarter` 签名一致；不一致则调整 `tfruntime`，不修改 P2 `register`

## 14. 配置项

- [ ] 14.1 `pkg/cc` 的 `IaCPluginConfig` 增加池参数与接入开关（默认关闭）、`registryCheckInterval`；`trySetDefault` / `Validate` 补默认值与校验，规则与 `PoolConfig.Validate` 一致（可直接复用其校验，避免两处各一套）
- [ ] 14.2 `cmd/hc-service/etc/hc_service.yaml`、helm `hcservice` 的 configmap 与 values 同步，注释写明含义与取值约束（`maxConfigAge` 须短于云凭证有效期；`minIdleProcesses` 是空闲数下限，回收只处理超出部分）
- [ ] 14.3 单测：缺省值通过校验、非法组合被拒绝；`helm template` 默认 values 渲染结果含池配置项且开关为关

## 15. 注册表与服务接入

- [ ] 15.1 `pool.Registry`：`Get(pluginName)`（未建池返回 `ErrPluginNotEnabled`）、按本副本已启用的 tf 插件建池并预热、读取状态失败跳过本轮、停用时 `Close(ctx)` 后移除、同一插件至多一个池
- [ ] 15.2 `registryCheckInterval` ticker + `ctx`，随服务退出停止
- [ ] 15.3 `cmd/hc-service/service/iac_plugin.go`：注册完成并就绪后建立注册表（开关为开时），`StopPluginTasks` 中先停巡检再关全部池并等待；只新增调用，不改 P2 注册 / 版本检查逻辑；注册期读 Schema 的临时进程不进池
- [ ] 15.4 开关关闭时不创建注册表、不预热；对接方真实 `ProcessManager` 就绪前保持关闭（桩的 `Start` 返回 `ErrNotImplemented`）
- [ ] 15.5 单测只覆盖纯逻辑：注册表的"应建 / 应关"差集判定

## 16. 联调与验证

- [ ] 16.1 与对接方真实 `ProcessManager` 联调：首次申请创建并配置、同配置复用、换配置重配、配置过期重配、业务错误整批失败、进程被 kill 后自愈、空闲回收与预热补回
- [ ] 16.2 并发验证：同配置突发共享一个进程（`maxLeasesPerProcess>1`，前提确认 provider 支持并发 RPC）、不同配置在容量满时排队保序
- [ ] 16.3 停用插件与服务退出：排队请求收到池已关闭错误、无残留 provider 进程
- [ ] 16.4 开关关闭回归：与接入前对比启动时序、`/healthz`、进程列表一致
- [ ] 16.5 `go build ./...`、`go vet`、受影响包 `go test -race` 通过
