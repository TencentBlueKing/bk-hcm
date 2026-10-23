## Context

方案（iWiki 4050984540）要点：每个插件一个池；调用方 `Acquire` 得到一个已按自己 config 配好的进程名额（lease），持有期内连续调用，归还靠 `Release()`；调度严格 FIFO、只看队首；池里一把锁，锁内只做内存读写，创建 / 配置 / 关闭等慢操作在锁外异步执行后写回。

对接方案（iWiki 4047720169）要点：运行时放在 `cmd/hc-service/logics/tfruntime`，采用 go-plugin 建连（一个连接绑一个进程，连接关闭进程即退出，不存在复用）；进程管理对外 `Start`（拉起、握手、返回执行器）与 `Close`（等在途 RPC → `StopProvider` → `Kill`）；执行器持有 `Provider` 并在同名方法上做 schema 校验与 DynamicValue 编解码；`ConfigureProvider` 内部先 `PrepareProviderConfig` 再 `ConfigureProvider`，凭证只走该字段；退出信号不会传给子进程，必须逐个 `Kill`。

**分支与基线**：进程池相对独立，从远程 `upstream/feat-plugin`（`4fba280d8`）切 `feat/iac-plugin-process-pool` 单独实现，后续再与插件注册、Schema 等分支合并。实测该基线：
- **没有**：`pkg/lock`、`pkg/api/core/plugin`、`pkg/iacplugin`、`cmd/hc-service/logics/plugin`、`IaCPluginConfig`、`service/iac_plugin.go`（即 P1~P4 的成果都不在）；
- **有**：`pkg/kit`、`pkg/logs`、`pkg/metrics`、`pkg/runtime/shutdown`，go.mod 已含 `github.com/hashicorp/terraform-json` 与 grpc，没有 go-plugin / terraform-plugin-go；`pkg/thirdparty/api-gateway/plugin` 里的 `PluginSchema`（内嵌 `tfjson.ProviderSchema`，外加 `ProtocolVersion / Plugin / SchemaHash / Capabilities`），与对接方案 6.2 中 `GetProviderSchema` 的响应形态一致。

项目偏好：不为了单测方便而引入接口 / 适配层；单测只覆盖纯逻辑（判定、状态转换），I/O 编排交给联调。

约束：进程池本身不依赖 go-plugin / gRPC；`Config` 含云凭证，不得进日志与错误信息；provider 是否支持并发 RPC 尚未确认（方案待定 #1）。

## Goals / Non-Goals

**Goals:**
- 把方案落成可分阶段合入、每阶段可独立验证的实现：阶段零先落地对接契约骨架，让池有真实类型可调用；阶段一能跑通"申请 → 配置 → 交付 → 归还"，阶段二补齐并发复用与异常自愈，阶段三补可观测性，阶段四（合并后）接入服务。
- **阶段零到三只新增、不修改**：改动限于 `tfruntime` 与 `pool` 两个新包，不触碰既有文件和 go.mod，使该分支与注册 / Schema 分支合并时不产生文件冲突。阶段四才修改既有文件。
- 调度顺序只由入队顺序决定，不存在饥饿；慢操作不拖住调度。
- 方案中的不变式 I1~I10 可作为单测断言与评审清单。

**Non-Goals:**
- 对接方的进程管理 / 执行器真实逻辑（go-plugin、gRPC、协议适配、DynamicValue 编解码、日志落盘），本 change 只落契约骨架。
- P2 `register` / `watch` 的修改（含让 `register` 改用 `tfruntime` 类型）：阶段四也不做，由对接方真实实现同时满足两处接口。
- Schema 同步；执行代理对 Plan / Apply 的编排、异步 Apply。
- lease 持有超时强制回收、有限越队、升级期池切换、每身份进程上限（均为方案待定或明确不做）。

## Decisions

### D1 分五阶段，阶段零到三不接入流量
| 阶段 | 内容 | 验收 |
| --- | --- | --- |
| 零：对接契约骨架 | `logics/tfruntime` 的接口、结构体、`ProcessManager` 桩与工厂 | `go build` 通过；桩方法一律返回 `ErrNotImplemented`；无既有文件被修改 |
| 一：最小闭环 | 队列、结构、`Acquire / cancel / Release`、规则 ①a ② ③ ④ ⑤、异步创建 / 配置 / 写回、配置过期、`Close`；失败 = 关进程 + 请求直接失败 | 假 `ProcessManager` 下单测通过不变式 I1~I3、I6、I7、I10；顺序与容量行为符合 spec |
| 二：并发与自愈 | 预约 ①b、失败分流与重试、取消三态、panic / 超时、定时任务（预热、空闲回收、`Exited`、`CheckHealth`、兜底调度） | 加入 I4、I5、I8、I9；表驱动覆盖 4.6 的失败表；基准测试 |
| 三：可观测性 | lease 泄漏告警、池指标与日志、`PoolConfig` 校验补全 | 指标与日志不含 config / 凭证；`go test -race` 通过 |
| 四：接入与运维（合并后） | `Registry`、hc-service 接入与开关、`pkg/cc` 配置、yaml / helm | 联调：启停插件建池 / 关池、服务退出无残留进程；开关关闭时行为与接入前一致 |

理由：阶段零只有类型，没有行为，风险最低，却是池的所有后续阶段能"对着真实类型编译"的前提；阶段一已含全部"状态机骨架"（写回、容量、取消），后两阶段是在骨架上加分支。池在合并前没有任何调用方，因此阶段零到三都不影响现网。阶段四依赖插件表读取、`IaCPluginConfig` 和 `service/iac_plugin.go`，这些在 `upstream/feat-plugin` 基线上不存在，必须等分支合并后才能开工。`maxLeasesPerProcess` 在待定 #1 未确认前默认 1（一个进程一个请求），模型自然退化，不影响阶段划分。

### D2 包位置与对外形态
池放在 `cmd/hc-service/logics/plugin/tf/pool`，与将来合并进来的 `discovery / register / watch` 并列；契约放在 `cmd/hc-service/logics/tfruntime`（对接方案 4.1 约定位置）。对外：`Pool.Acquire(kt, config) (*Lease, error)`、`Pool.Close(ctx)`、`Lease.Release()`。方法入参沿用项目的 `*kit.Kit`：排队等待受 `kt.Ctx` 与 `maxQueueWait` 约束；进程创建 / 配置用池自己的 `ctx`，不跟随请求取消（方案 6.10）。池通过构造参数拿到 `tfruntime.Starter`，不自行创建。

### D3 池只依赖 `tfruntime` 的契约类型
`tfruntime.Executor` 是对接方案 6.2 的**完整**执行器方法集（含 `GetProviderSchema`、`ConfigureProvider` 与 8 个业务方法），契约层不做"哪些给业务方"的区分。池直接引用 `tfruntime.Executor / ProcessManager / Starter`。

"交付给调用方的是哪一部分"属于池的职责：池包自己定义窄接口 `ProviderIF`（仅 8 个业务方法，池包内类型），`Lease` 持有 `tfruntime.Executor` 并实现 `ProviderIF`——方法逐个转发给执行器，不暴露 `ConfigureProvider` 与 `GetProviderSchema`，从类型上保证调用方无法自行改配置。封装层还承担池的职责：lease 已归还或已失效（`Release` 之后、进程被关闭）时，转发前直接返回错误而不是打到可能已被重配的进程上。这一层随阶段一的 `Lease` 一起实现。

`Starter` 的签名与 P2 分支（`feat/iac-plugin-register-submit`）里的 `PluginProcessStarter` 保持一致（`NewProcessManager(kt, pluginName, pluginVersion, execPath)`），合并时 P2 可以直接替换为 `tfruntime.Starter`。`Close` 与 P2 的差别是带 `ctx`（对接方案 6.3 的 `Close` 阻塞，池关闭又需要超时控制），该差异在合并计划里处理，本分支不碰 P2。

### D4 队列用具体类型，不抽 `FIFO` 接口
方案给了 `FIFO` 接口并保留链表备选。按"只在确有多实现需求时才抽象"，本版直接实现 `sliceFIFO`（`PushBack / Peek / PopFront / Remove / Len / Drain`），取消靠延迟移除。需要换链表时再提接口，调度代码只依赖这六个方法，改动面很小。

### D5 单锁 + "锁内改状态 → 锁外慢操作 → 锁内写回"
- 只有入口函数取锁（`Acquire / cancel / Release / 任务写回 / 定时任务 / Close`），内部函数以 `Locked` 结尾。
- 每个进程同一时刻至多一个异步任务（I4），任务必定写回且仅写回一次（I5）；任务带超时；panic 经 `recover` 按传输错误处理并仍写回。
- 写回顺序：核对进程仍在表中且 `pending` 属于本次任务 → 按返回原因写进程与请求 → 调一次 `scheduleLocked()`；任何分支不得直接 return 而丢下请求。
- 进程独占靠状态（`starting / configuring`）表达，不设进程级锁。
- 任务序号：每个进程记录当前任务序号 `task`，写回时 `endTaskLocked` 核对"记录仍在表中且序号一致"才写，避免过期写回改到别的任务。
- 外部调用加看门狗：`Start` 与 `ConfigureProvider` 在独立协程中执行，超时即按失败写回、不再等待（实现未必响应 ctx）；随后关闭进程，由 `ProcessManager.Close` 让阻塞调用返回，迟到结果丢弃。为此 `NewProcessManager` 同步调用、只构造对象，保证 `Start` 超时时池已持有 manager、能关掉进程。
- 外部调用的 panic 经 `safeCall` 转成错误，只包裹外部调用本身，写回代码不在 recover 范围内，从结构上保证写回恰好一次。
- 关闭未能确认进程退出时（`Close` 报错且 `Exited()` 为假），记录保留为 closing 继续占用容量，不删除，避免实际进程数超过上限；由阶段二的退出检查再次确认后删除。

### D6 调度：取一个派一个，选进程单次遍历进程表
`scheduleLocked` 整轮只取一次 `now`；循环 `Peek → pickLocked → PopFront → dispatchLocked`，选不出进程即结束本轮（不往后找）。`pickLocked` 只遍历进程表（O(进程数)），优先级：匹配可交付 > 可预约（阶段二）> 空闲未配置 / 已过期 > 空闲且配了别的 config > 新建 > 无。同类候选 `inflight` 少者优先，平局按 ID。阶段一的 `pickLocked` 不含"可预约"分支，阶段二只增加该分支。

### D7 `binding` 同时承载"目标配置"与"生效时间"
`binding == nil` 未配置；`boundAt == nil` 正在配置（只能被预约，不能交付）；`boundAt != nil` 已生效，是否过期看 `maxConfigAge`。配置开始时就覆盖 `binding`（旧配置从此不可信），失败清空。规则 ② 的条件是"未配置或已过期"，而不是"binding 为空"。

### D8 `Config` 与 key
匹配只比 `binding.key`（config 规范化后的 SHA-256）。规范化口径属方案待定 #2，阶段一先用"字段排序后的 JSON"并集中在一个函数里，口径定稿时只改此处。`Config` 不实现 `String()`，日志只记 key 前缀。

### D9 失败语义分两档
- 阶段一：任何 `Start / Configure` 失败 → 进程转 `closing` 并关闭（业务错误可保留进程，见阶段二），触发的请求返回错误。
- 阶段二：按方案 4.6 的表：业务错误 → 进程回 `running`、`binding=nil`、`pending` 整批失败；传输 / 超时 / 进程退出 → 关进程、`pending` 按原相对顺序回队尾，`attempts > maxRetries` 则失败。重试不改 config。

业务错误与传输错误的判别见 D13：`ConfigureProvider` 返回 `*tfruntime.DiagnosticsError` 为业务错误，其余为传输错误。

### D10 定时任务用本池 `time.Ticker`；预热与回收共用一个"空闲数"口径
池内巡检（默认 30 秒）：预热 / 空闲回收、`Exited()`、`CheckHealth()`（仅 `running`），末尾无条件 `scheduleLocked()`。慢操作在锁外，用 `ctx` 控制退出，不引入 `pkg/cron`。池的 `Close` 会停掉自己的巡检。注册表（阶段四）另有一个周期更短的巡检（默认 3 秒），对照本副本已启用的 tf 插件建池 / 关池，同样用 `ticker + ctx`，随服务退出（`shutdown` 通知）停止。

**空闲进程**统一定义为 `state == running && inflight == 0`，**不论是否已配置、配置是否过期**（已配置的空闲进程同样可被规则 ② ③ 复用）。此外，尚未被任何请求认领的预热进程（`starting` 且 `pending` 为空）按"将成为空闲"计入，避免重复补建。记 `idle` 为该数量，`min = minIdleProcesses`：

| 条件 | 动作 |
| --- | --- |
| `idle < min` 且池内进程数 `< maxProcesses` | 补建 `min - idle` 个未配置进程（受容量限制，失败退避；池满不挤占服务中的进程） |
| `idle > min` | 在空闲时长超过 `idleReclaimTimeout` 的空闲进程中，按空闲最久优先回收，**最多回收 `idle - min` 个**，使回收后恰好不低于 `min` |
| `idle <= min` | 不回收 |

这样"保持空闲"与"空闲回收"不再冲突：`min` 是空闲进程数的下限，回收只处理超出下限的那部分；回收后若有请求把空闲进程占走，下一轮 `idle < min` 会由预热补回。

### D11 Lease 只存进程 ID
`Lease` 保存 `procID`、`pool`、业务方法入口与 `acquiredAt`，不保存进程记录指针；`Release` 用 `sync.Once`，进程不在表中则忽略，ID 单调不复用。阶段三仅对持有时长超阈值输出日志与指标，不强制回收。

### D12 配置：阶段零到三用 `PoolConfig`，阶段四再映射到 `pkg/cc`
`maxProcesses`、`maxLeasesPerProcess`（默认 1）、`minIdleProcesses`（1）、`startTimeout`（30s）、`configureTimeout`（30s）、`closeTimeout`（30s，池内部关闭进程的时限；`Pool.Close` 用调用方 ctx）、`maxRetries`（2）、`maxQueueWait`、`maxQueueLength`、`idleReclaimTimeout`（10m）、`maxConfigAge`（30m）、`poolCheckInterval`（30s）、`leaseHoldWarn`。全部由 `PoolConfig` 结构体承载，带 `trySetDefault / Validate`，构造池时传入。校验：`maxProcesses ≥ 1`、`maxLeasesPerProcess ≥ 1`、`minIdleProcesses ≤ maxProcesses`、各时长大于 0。阶段四把这些参数加入 `IaCPluginConfig`（yaml `iacPlugin` 下，沿用 `trySetDefault / Validate` 约定），另加接入开关与 `registryCheckInterval`（默认 3 秒），由 hc-service 在构造池时映射成 `PoolConfig`；`Validate` 与 `PoolConfig.Validate` 的规则保持一致，避免两处各一套。yaml 与 helm 的 hcservice configmap / values 同步新增并写明含义与取值约束（`maxConfigAge` 须短于云凭证有效期；`minIdleProcesses` 是空闲数下限，回收只处理超出部分）。开关默认关闭。

### D13 对接契约骨架（阶段零）：只落类型，桩方法返回 `ErrNotImplemented`
包路径 `cmd/hc-service/logics/tfruntime`。越界范围严格限定为"结构体与接口 + 能让池编译和被调用的最小桩"，不写任何协议逻辑。

| 文件 | 内容 |
| --- | --- |
| `errors.go` | `ErrNotImplemented`；`DiagnosticsError`（持有 `[]Diagnostic`，用于把 Error 级 Diagnostics 转成 `error`，`Error()` 不含 Config 内容） |
| `types.go` | `Diagnostic`（`Severity / Summary / Detail / Attribute`）、`Severity` 常量（`Error / Warning`）、`Object`（`map[string]any`，nil 表示 null）、`Unknown`（未知值哨兵）、`AttributePath`；`ProviderSchema`（对 `pkg/thirdparty/api-gateway/plugin.PluginSchema` 的类型别名）；各方法的 `XxxRequest / XxxResponse`，字段与对接方案 6.2 一一对应（`TypeName`、`Config`、`PriorState`、`ProposedNewState`、`PlannedState`、`PlannedPrivate`、`Private`、`RequiresReplace`、`ImportedResource` 等，`Private` 为 `[]byte`） |
| `executor.go` | `Executor`（完整方法集：`GetProviderSchema`、`ConfigureProvider` + 8 个业务方法）；所有方法首参 `*kit.Kit` |
| `process.go` | `ProcessConfig`（`PluginName / PluginVersion / ExecPath / Args / WorkDir / Env / LogDir`）；`ProcessManager` 接口；结构体 `processManager` 及 `NewProcessManager(cfg ProcessConfig) (ProcessManager, error)`；`Starter` 接口，签名与 P2 的 `PluginProcessStarter` 一致，默认实现 `defaultStarter` 内部组装 `ProcessConfig` 后调用上面的工厂 |

`ProcessManager` 接口：

```go
type ProcessManager interface {
    Start(kt *kit.Kit) (Executor, error) // 拉起、握手，返回执行器
    Close(ctx context.Context) error     // 阻塞：等在途 RPC → StopProvider → Kill；可重复调用
    Exited() bool                        // 进程是否已退出，非阻塞
    CheckHealth() error                  // 连接是否可用，非阻塞
}
```

约定：
- `GetProviderSchema` 返回 `*ProviderSchema`。该别名当前指向基线上已有的南向 `PluginSchema`，其形态与对接方案 6.2 的响应一致；P2 分支使用的是 `pkg/api/core/plugin.PluginSchema`，合并时只需调整这一处别名（或在合并 change 里加一次转换）。池本身不调用 `GetProviderSchema`。
- 请求 / 响应里的 `Config / State` 一律是 `Object`；DynamicValue 编解码、按 schema 校验由对接方在 `Executor` 实现里完成，契约不暴露 `tftypes`。未知值用 `Unknown{}` 表达，是否足够由对接方在实现时确认。
- 桩实现 `processManager` 持有 `ProcessConfig` 与关闭标志位等必要字段，方法体直接 `return nil, ErrNotImplemented`（`Exited()` 返回 `false`、`CheckHealth()` 返回 `ErrNotImplemented`、`Close` 返回 `nil` 并幂等）。不实现 `Executor`：`Start` 在桩里返回 `ErrNotImplemented`，不提供假执行器；池的测试用例在测试文件内自带假实现。
- 包级编译期断言 `var _ ProcessManager = (*processManager)(nil)`。
- 不引入 go-plugin / tfprotov5 / tftypes 依赖。对接方实现时再引入，届时可能需要为 `processManager` 增加字段，但接口与结构体签名保持不变。
- 要求对接方实现时遵守的行为（写入 spec）：`Close` 开始后已返回的执行器不再接受调用；`Close` 可重复调用；`Exited / CheckHealth` 不阻塞；`ConfigureProvider` 内部先 `PrepareProviderConfig` 再 `ConfigureProvider`，凭证不进进程环境变量。

业务错误与传输错误的区分：契约约定 `ConfigureProvider` 的失败有两类——响应含 `Error` 级 `Diagnostics` 时返回 `*DiagnosticsError`（业务错误）；其余（超时、连接断开、进程退出）返回普通 `error`（传输错误）。池用 `errors.As` 判别。

### D14 池注册表与服务接入（阶段四，合并后）
- **`Registry`**：`map[pluginName]*Pool`，一把锁保护；对外 `Get(pluginName)`（未建池返回 `ErrPluginNotEnabled`）、`Run(ctx)`（周期巡检）、`Close(ctx)`（关全部池）。巡检读取本副本已启用的 tf 插件（数据来自注册表 / data-service，副本本地只关心有本地二进制的插件），读到已启用且无池则建池并预热；读到停用则对池 `Close(ctx)` 后移除。读取失败只记日志、跳过本轮，不误关池。同一插件至多一个池。
- **Starter 来源**：`Registry` 构造时接收 `tfruntime.Starter`，与 P2 注册流程使用同一个真实实现，但两处各自持有自己声明的接口类型；P2 `register` 现有的注入点保持不变（本 change 不修改 `register`）。
- **接入时机**：沿用 P2 的启动时序——目录检查 → 注册 → 就绪（`SetPluginReady`）之后再建立注册表；服务退出时 `StopPluginTasks` 先停注册表巡检，再对全部池 `Close(ctx)` 并等待。注册期读 Schema 的临时进程不进池。
- **开关**：`iacPlugin` 下新增接入开关，默认关闭；关闭时不创建注册表、不预热、不拉起任何进程，行为与接入前一致。真实 `ProcessManager` 就绪前开关保持关闭，否则桩的 `Start` 返回 `ErrNotImplemented` 会让所有 `Acquire` 失败。
- **配置与 helm**：见 D12。

## Risks / Trade-offs

- [队首缺资源时后面的请求一起等] → 已接受（严格 FIFO 的代价，只影响延迟）；用 `maxQueueLength` 与 `maxQueueWait` 兜底；越队放宽作为后续备选。
- [provider 不支持并发 RPC 时 `maxLeasesPerProcess>1` 会出错] → 默认 1；改大前须先确认待定 #1。
- [lease 泄漏导致进程少一个名额] → 调用方统一 `defer Release()`；阶段三告警；极端情况重启可恢复；强制回收留作备选。
- [卡住的进程池发现不了：连接正常但某操作一直卡] → 第一版靠 lease 持有告警发现，已在方案中接受。
- [配置过期只挡新请求，长 Apply 可能跨过凭证有效期] → 方案待定 #3，`maxConfigAge` 取值须短于凭证有效期并留余量。
- [越界落契约骨架，可能与对接方后续实现不一致] → 骨架严格对照对接方案 6.2 / 6.3 逐字段生成，只放类型和桩；评审阶段零的产物时请对接方确认，不一致在阶段一之前修正；对接方只需填充方法体，不需要改调用方。
- [阶段四误在合并前开工] → 阶段四的前置条件写进 tasks（任务 13.1 先核对分支已合并且依赖齐备）；未满足不得开始。
- [在较早的基线上开发，合并时可能有接口差异] → 本分支只新增文件，与 P2~P4 无文件重叠，git 层面不会冲突；语义层面的差异（`Close(ctx)`、`ProviderSchema` 别名）已在合并计划中列出。
- [`Starter` 签名与 P2 分支的 `PluginProcessStarter` 假设一致] → 开工前对照 P2 分支 `register/types.go` 再核一次；若 P2 已变动，以 P2 为准调整契约。
- [空闲数把已配置进程也算空闲，可能让凭证长期留在内存中] → 受 `maxConfigAge` 约束：过期后不能直接交付，只能被重配；空闲回收按空闲最久优先，超出 `min` 的部分会被回收。
- [`Exited / CheckHealth` 超出对接方案 6.3 列出的 `Start / Close`] → 已在契约中补齐，两者均不阻塞且 go-plugin 原生支持（`Client.Exited()`、对连接做轻量探活），列入 Open Questions 请对接方确认。
- [阶段一缺少退出检测与重试] → 池在合并前无调用方；阶段二完成后才具备接入条件。
- [未压测的性能估算] → 临界区为微秒级内存操作，一轮派发量取决于 `maxProcesses × maxLeasesPerProcess`；阶段二结束前补一次基准测试。

## Migration Plan

### 阶段零到三（本分支）
1. 从 `upstream/feat-plugin` 切出 `feat/iac-plugin-process-pool`。
2. 阶段零：新增 `tfruntime` 包。
3. 阶段一、二、三：新增 `pool` 包。
全程只新增文件，不影响现网；分支内无调用方。

### 合并（阶段三完成后，阶段四之前）
1. 将 `feat/iac-plugin-process-pool` 与注册 / Schema 所在分支（P1~P4）合并；因阶段零到三只新增文件，预期无文件冲突。
2. `tfruntime.ProviderSchema` 别名改指 `pkg/api/core/plugin.PluginSchema`（或加一次转换），以合并时的实际类型为准。
3. 核对 `tfruntime.Starter` 与 P2 `PluginProcessStarter` 签名一致（合并时发现不一致，以 P2 为准调整 `tfruntime`，不改 P2）。
4. `go build ./...` 与两边既有单测通过后，才进入阶段四。

### 阶段四（合并后）
1. 先合 `pkg/cc` 配置项与默认值、yaml / helm 同步（开关默认关闭，对现网为空操作）。
2. 再合 `Registry` 与 `service/iac_plugin.go` 的接入代码，仍保持开关关闭。
3. 对接方真实 `ProcessManager` 就绪后，在联调环境打开开关做联调与验证。
4. 回滚：关闭开关或回退 hc-service 版本；池无持久化状态，进程随服务退出由 `Close` 清理，不影响注册表记录。

### 回滚（阶段零到三）
本分支内无调用方，不合并或回退分支即可，无任何运行时影响。

## Open Questions

- 请对接方评审阶段零的契约骨架：`Exited()`、非阻塞 `CheckHealth()` 是否可行；`Close(ctx)` 可重复调用且 ctx 超时后强制退出；`Start` 能否取消；业务错误以 `*DiagnosticsError` 返回、传输错误返回普通 `error` 的约定；`Object` / `Unknown` 是否足够表达 null 与未知值；是否需要子进程 PID（方案待定 #5）。阶段一开工前确认。
- provider 是否支持并发 RPC、Apply 是否必须独占（待定 #1）。
- `Config` 规范化口径与脱敏（待定 #2）；`maxConfigAge` 与凭证有效期的取值（待定 #3）。
- 池关闭时处于 Plan 与 Apply 之间的 lease，其后续调用返回错误（待定 #8，已定），执行代理需据此处理。
- 合并时 `ProviderSchema` 到底指向哪个 `PluginSchema`：以合并时 P2 分支的实际类型为准。
- 阶段四：插件"已启用"状态的读取方式（data-service 的列表接口还是复用 P2 `watch` 的读取路径）与开关键名，在合并后开工时确定。
