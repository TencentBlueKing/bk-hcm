## Why

标准运维插件「创建监听器并绑定RS-TCPUDP」分两步各起一个异步 flow，近期第二步（绑 RS）频繁报错 `resID: xxx is processing`。怀疑是主 flow 已对外表现为完成、但负责解锁的 watch flow 尚未被调度执行，导致 `resource_flow_lock` 上的锁未及时释放。

但**当前线上日志无法证实或推翻这个推断**：加锁成功、解锁成功、锁被谁持有、watch flow 与主 flow 的对应关系、flow 在调度链各段的等待时长，全部没有日志。本变更只补充可观测性，不改变任何加解锁时机与业务行为，目标是让下一次线上复现能够被定位。

关联需求：TAPD story 1069995598136675686。

## What Changes

### 1. 加解锁日志（含持锁 owner）

- `checkResFlowRel` 预检测撞锁时，把已经查到但被丢弃的持锁记录打出来：持锁 flowID（`owner`）、加锁时刻。调用方目前用 `_, err :=` 丢掉了这个返回值，改为使用它记录日志。
- 撞锁的两种分支分别打日志并区分：锁表命中（锁还在）vs `resource_flow_rel` 命中（锁已放但关联记录未收尾）。两者根因完全不同。
- 加锁成功（cloud-server 侧 `lockResFlowStatus` 与 data-service 侧 `ResFlowLock`）与解锁成功（`ResFlowUnLock`）补 Info 日志。
- 锁泄漏的识别放在 watch action：`processResFlow` 的终态分支已经按 `owner = flowID` 查过锁表，查不到时当前是静默 `return true, nil`。在这里补一条 Warn（主 flow 已终态但该 flow 名下无锁），即可发现「解锁按 owner 匹配不上」的情况，不需要新增任何查询。

### 2. Watch flow 与主 flow 的关联

- 各 executor 的 `createFlowTask` 创建 watch flow 后丢弃了返回的 watch flowID，导致日志中**无法把 watch flow 与主 flow 关联**。补一条 Info 日志记录二者对应关系。
- `FlowSlaveOperateWatchAction` 补进入/退出日志、解锁前后日志、以及「主 flow 进入终态 → 解锁完成」的耗时。这个耗时就是需求中怀疑的时间窗，是本变更最核心的观测指标。

### 3. 任务与 flow 状态更新成功日志

- `executor.UpdateTask` 补成功日志（taskID / flowID / actionName / 状态迁移 / 耗时）。
- `updateFlowState` / `updateFlowStateAndReason` 补成功日志；CAS 未更新（`RecordNotUpdate`）单独打 Warn，只打期望的源/目标状态，不查库中实际状态。

### 4. 调度链路等待时长

- `dispatcher.Do` 派发 Pending→Scheduled 时记录 flowID、flowName、分配到的 worker。
- `scheduler.runScheduledFlow` 中现有的 `logs.V(3)` 在生产 V 级别下不输出，补 Info 级别的选中日志与 Scheduled→被选中的等待时长。
- `watchInitQueue` 出队时打 initQueue 等待时长（`payload.entryTime` 在该作用域现成可用），`initWorkerTask` 打进入 fast 还是 slow 队列。fast/slow 队列内的等待时长由该条日志与 `workerDo` 已有的 `start execute task` 日志的时间戳差得出——**不给 `Task` 加时间戳字段**。

### 5. rid 修正（前置项）

`executor` 的 `exec.kt` 在 `NewExecutor` 时创建一次，整个 task-server 进程共用一个 rid。`UpdateTask`、`workerDo` 的失败日志都用它，导致日志无法区分任务。改为**同时打印两个 rid**：`exeRid`（进程/执行器实例）与 `rid`（`task.Kit.Rid`，scheduler 轮询 → flow → task 的层级 rid）。代码中 `executor.go` 已有 `exeRid` / `taskRid` 双打的先例，沿用该约定。

**注意**：两个 rid 都不是标准运维那次请求的 rid（`task.Kit` 的根是 scheduler 轮询时新建的 kit）。因此日志的**跨服务关联键必须是 flowID / resID**，rid 只用于串联单进程内的一段调用。所有新增日志必须同时带 flowID 与 resID。

## Capabilities

### New Capabilities

- `clb-res-lock-observability`: CLB 异步任务资源锁（`resource_flow_lock`）的加解锁、持锁归属、watch flow 关联关系与解锁延迟的日志可观测性要求。
- `async-flow-schedule-observability`: 异步任务框架中 flow 派发、调度选中、task 排队与状态更新的日志可观测性要求，含 rid 打印约定。

### Modified Capabilities

无。本变更不修改任何已有 spec 的行为要求。

## Impact

**受影响服务**：cloud-server、data-service、task-server（三层均涉及）

**受影响代码**：

| 文件 | 改动 |
|---|---|
| `cmd/cloud-server/logics/load-balancer/res_lock.go` | 加锁成功日志、撞锁 owner 日志 |
| `cmd/cloud-server/service/load-balancer/async_target_group_add_rs.go` | 同上（`lockResFlowStatus` / `checkResFlowRel` 的另一份实现） |
| `cmd/cloud-server/logics/cvm/res_lock.go` | 同上（CVM 侧同构实现） |
| `cmd/cloud-server/logics/load-balancer/*_executor.go`、`cmd/cloud-server/service/load-balancer/*.go` | `checkResFlowRel` 调用方改为使用返回的持锁记录；`createFlowTask` 记录 watch flowID |
| `cmd/data-service/service/cloud/load-balancer/resource_flow.go` | `ResFlowLock` / `ResFlowUnLock` 成功日志、解锁前查询实际 owner |
| `cmd/task-server/logics/flow/flow_slave_operate_watch.go` | watch 生命周期日志、解锁延迟耗时 |
| `pkg/async/consumer/executor.go` | `UpdateTask` 成功日志、双 rid、排队耗时 |
| `pkg/async/consumer/scheduler.go` | flow 状态更新成功日志、选中日志与等待时长 |
| `pkg/async/consumer/dispatcher.go` | 派发日志 |

**不涉及**：DB schema、对外 API、依赖变更。

**范围红线（纯日志变更）**：

- MUST NOT 修改任何函数、方法、接口的签名，包括 DAO 层。
- MUST NOT 给框架结构体新增字段（`Task`、`Flow`、`InitPayload`、`FlowSlaveOperateWatchOption` 等），也不得为传递日志字段而改变数据在框架内的流转方式。
- MUST NOT 为打日志新增 DB 查询。日志只能使用当前作用域已有的变量；拿不到的字段就不打，改为依靠两条日志的时间戳差或另一条日志补全。
- MUST NOT 修改对外错误消息（如 `resID: %s is processing`），持锁 flowID 只进日志不进错误消息。
- MUST NOT 修改加解锁时机、flow/task 状态机、调度策略与参数。
- 允许的唯一"非纯增量"改动：把丢弃已有返回值的 `_, err :=` 改为接收该返回值用于打日志（`checkResFlowRel` 已返回持锁记录、`CreateTemplateFlow` 已返回 watch flowID）。这不改签名也不改行为，是拿到日志字段的唯一途径。

**风险**：新增日志集中在异步任务高频路径，需控制单条日志体积与打印频率，避免日志量激增；调度链路的高频轮询日志维持在 `logs.V(n)` 或仅在命中关键分支时打印。
