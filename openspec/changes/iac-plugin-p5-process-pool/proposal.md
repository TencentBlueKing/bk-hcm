## Why

P2 的注册流程只会为"读 Schema"拉起一次性进程、用完即关；运行期执行代理（Plan / Apply 等业务调用）需要反复使用 tf_provider 进程，而 `ConfigureProvider` 要带云凭证、进程启动与配置都是慢操作，不能一个请求拉一个进程。需要一个进程池：按"凭证配置"复用已配好的进程、控制进程数与并发、保证请求公平不饿死，并在进程异常时自愈。

进程池本身相对独立：它只依赖"进程管理 / 执行器"这一组对接对象，不依赖插件注册、Schema 管理、插件表和配置。因此本 change 分两段推进：阶段零到三**单独开分支实现、不修改任何既有文件**；等池稳定并与注册 / Schema 等分支合并之后，再做**阶段四**（池注册表、hc-service 接入、`pkg/cc` 配置项、helm），阶段四依赖插件表与注册能力，在合并前不得开工。

为了让池有真实类型可调用、可编译，本 change 越过边界，先落地对接方的**契约骨架**：接口、请求 / 响应结构体、进程管理结构体及其工厂，方法体不实现，返回"未实现"错误，由对接方后续填充。

技术方案：《进程池》（iWiki 4050984540）；进程管理与执行器的对接约定见《插件对接方案》（iWiki 4047720169）。

## 分支策略

- 基线：远程 `upstream/feat-plugin`（当前 `4fba280d8`，含南向协议与加密，**不含** P2 注册 / P3 管理接口 / P4 http 注册）。
- 实现分支：`feat/iac-plugin-process-pool`，从 `upstream/feat-plugin` 切出。
- 在该基线上不存在、因此本 change **零依赖**的内容：`pkg/lock`、`pkg/api/core/plugin`、`pkg/iacplugin`、`cmd/hc-service/logics/plugin`、`IaCPluginConfig`、`service/iac_plugin.go`。
- 在该基线上存在、本 change **会用到**的内容：`pkg/kit`、`pkg/logs`、`pkg/metrics`、`github.com/hashicorp/terraform-json`（已在 go.mod）、`pkg/thirdparty/api-gateway/plugin` 的 `PluginSchema`。
- 阶段零到三只新增文件，不改动任何既有文件，不改动 `go.mod`。
- 合并：阶段三完成后，将该分支与注册 / Schema 分支（P1~P4）合并；合并步骤见 design 的"合并计划"。
- 阶段四：在合并后的分支上实现，会修改既有文件（`service/iac_plugin.go`、`pkg/cc`、`hc_service.yaml`、helm），只在此阶段允许。

## What Changes

按风险递增分五个阶段交付。阶段零到三只新增 `tfruntime` 与 `pool` 两个包、不接入流量；阶段四在合并后接入服务。

- **阶段零：对接契约骨架（`logics/tfruntime`，只有类型，没有逻辑）**
  - 新增 `Executor` 接口（对接方案 6.2 的完整方法集，不拆分），覆盖全部方法：`GetProviderSchema`、`ConfigureProvider`、`ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`
  - 新增各方法的请求 / 响应结构体与公共结构 `Diagnostic`
  - `GetProviderSchema` 返回 `ProviderSchema`，它是对 `pkg/thirdparty/api-gateway/plugin.PluginSchema` 的类型别名（字段与对接方案 6.2 的响应一致），合并时只改这一处别名
  - 新增 `ProcessManager` 接口（`Start`、`Close(ctx)`，以及池需要的 `Exited`、`CheckHealth`）、`ProcessConfig` 结构体、`processManager` 结构体及 `NewProcessManager` 工厂、`Starter` 接口；所有方法体返回 `ErrNotImplemented`
  - 不引入 go-plugin / tfprotov5 / tftypes 依赖，不写握手、gRPC、编解码
- **阶段一：最小闭环（调度与交付）**
  - 新增每插件一个进程池，对外 `Acquire(kt, config) → Lease`、`Lease.Release()`、`Close(ctx)`；池只交付 lease，业务方法由调用方通过 lease 调用（`Lease` 是池对 `tfruntime.Executor` 的一层封装，只暴露业务方法，不暴露配置与取 Schema），配置在交付前已完成
  - 新增单条严格 FIFO 等待队列（切片 + 延迟移除取消）、队列长度上限、单请求排队时限
  - 新增调度规则：只取队首；已生效且匹配且有空位直接交付 → 空闲进程（未配置 / 已过期 / 配置了别的）配置 → 未达上限新建 → 否则队首等待；一把池锁，锁内只做内存读写，创建与配置在锁外异步执行并写回
  - 新增进程状态 `starting / configuring / running / closing` 与 `inflight` 计数、配置有效期 `maxConfigAge`（调度时判断，不设定时解绑）
  - 阶段一的失败处理保持最简：进程侧失败则关闭该进程、对应请求直接失败，不重试、不预约
- **阶段二：并发利用率与自愈**
  - 新增准备期预约（规则 ①b）：同配置请求加入正在启动 / 配置的进程的 `pending`，配置成功后一起交付
  - 新增失败分流与重试：业务错误整批失败并复用进程；`Start` 失败、`Configure` 超时 / 传输错误、准备期进程退出则关进程，请求回队尾重试至多 `maxRetries` 次
  - 新增取消三态处理（队列中 / `pending` 中 / 已交付）、任务 panic 恢复与超时
  - 新增池内定时任务：**空闲进程数统一口径（`running && inflight == 0`，不论是否已配置）**；空闲数低于 `minIdleProcesses` 时预热补足，**仅当空闲数高于 `minIdleProcesses` 时才回收超时空闲进程，且最多回收到恰好剩 `minIdleProcesses` 个**；另有 `Exited()` 检查、`CheckHealth()` 检查，并无条件触发一次调度作为兜底
- **阶段三：可观测性**
  - 新增 lease 持有超时告警（只告警，不强制回收）
  - 新增池指标与日志（排队时长、队列长度、各状态进程数、空闲数、重试数、失败分类）；日志与指标不含 config / 凭证
- **阶段四：接入与运维（前置：阶段三完成且分支已与注册 / Schema 分支合并）**
  - 新增池注册表：按本副本已启用的 tf 插件建池 / 预热，运行期按周期读取插件状态，停用时 `Close` 后移除，同一插件至多一个池
  - hc-service 接入：注册完成并就绪后建立注册表，服务退出时先停巡检再关全部池；接入由配置开关控制，默认关闭
  - `pkg/cc` 的 `IaCPluginConfig` 增加池参数与开关，映射到 `PoolConfig`，补默认值与 Validate
  - `cmd/hc-service/etc/hc_service.yaml`、helm 的 configmap / values 同步新增配置项并写明含义与取值约束
  - 联调：与对接方真实 `ProcessManager` 联调、并发验证、停用插件与服务退出验证
- **不包含**：
  - 对接方真实实现（go-plugin 拉起 / 握手 / gRPC / 编解码 / 日志落盘 / `StopProvider`），本 change 只留骨架
  - **P2 `register` / `watch` 的修改**：包括让 `register` 改用 `tfruntime` 类型，阶段四也不做；对接方真实实现同时满足两处的接口即可，`register` 现有的注入点保持不变
  - Schema 同步、执行代理对 Plan / Apply 的编排与异步 Apply、lease 持有超时强制回收（方案待定 #10）、有限越队（待定 #7）、插件升级时池的切换（待定 #4）、每身份进程上限（方案明确不设）

## Capabilities

### New Capabilities

- `iac-plugin-tfruntime-contract`：对接方契约骨架——执行器 / 进程管理接口、请求响应结构体、Diagnostics、未实现语义（阶段零）
- `iac-plugin-pool-scheduling`：lease 交付模型、FIFO 队列、调度规则、配置有效期、容量与队列限制（阶段一，预约规则在阶段二补充）
- `iac-plugin-pool-process-lifecycle`：进程状态与容量占用、异步创建 / 配置及写回、归还与关闭（阶段一）
- `iac-plugin-pool-resilience`：准备期预约、失败分流与重试、取消处理、定时巡检与预热回收（阶段二）
- `iac-plugin-pool-observability`：池参数校验、lease 泄漏告警、指标与脱敏（阶段三）
- `iac-plugin-pool-integration`：池注册表与启停、hc-service 接入与开关、配置项与 helm 同步（阶段四）

### Modified Capabilities

（无。）

## Impact

- **cmd/hc-service/logics/tfruntime**（新增）：契约骨架，仅类型与桩方法
- **cmd/hc-service/logics/plugin/tf/pool**（新增）：`Pool`、`PluginProcess`、`Lease`、`request`、队列、调度、定时任务；直接使用 `tfruntime` 的接口
- 阶段零到三：无既有文件被修改；无新增第三方依赖（`terraform-json` 已在 go.mod；go-plugin 等由对接方真实实现时引入）
- 阶段四（合并后）：
  - **cmd/hc-service/logics/plugin/tf/pool**：新增 `Registry`
  - **cmd/hc-service/service/iac_plugin.go**、**cmd/hc-service/app**：建立 / 关闭注册表（新增调用，不改 P2 注册逻辑）
  - **pkg/cc**：`IaCPluginConfig` 增加池参数与开关
  - **cmd/hc-service/etc/hc_service.yaml**、**docs/support-file/helm** 的 hcservice configmap 与 values
- 无表结构变化；无新增对外 API

## 依赖与后续

- 前置：`upstream/feat-plugin`（仅需其提供 `pkg/kit`、`pkg/logs`、`pkg/metrics`、`PluginSchema`）。
- 对接方后续工作：在阶段零的骨架上填充 `processManager` 与 `Executor` 的真实实现。联调前池用测试内的假实现。
- 阶段四依赖：与 P1~P4 所在分支合并完成（需要读取插件注册表的 data-service 客户端、`IaCPluginConfig`、`service/iac_plugin.go` 的就绪与退出钩子）。
- 被依赖：执行代理通过注册表的 `Get(pluginName)` 取得池，再 `Acquire` 获取 lease。
- 后续备选：lease 持有超时强制回收、有限越队、升级期池切换，待线上数据或对接方结论后另起 change。
