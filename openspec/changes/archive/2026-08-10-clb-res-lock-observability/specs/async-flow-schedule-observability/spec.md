## ADDED Requirements

### Requirement: 异步框架的改动限定为纯日志

本能力涉及 `pkg/async` 下的异步任务框架，其实现 MUST 只增加日志语句。

- MUST NOT 修改任何函数、方法、接口的签名。
- MUST NOT 给 `Task`、`Flow`、`InitPayload`、`InitPayloadWithScore` 等框架结构体新增字段。
- MUST NOT 为打日志新增 DB 查询或改变数据在框架内的流转方式。
- MUST NOT 调整 `listScheduledFlowLimit`、worker 数量、队列容量、flow 打分算法等调度行为。

日志只能使用当前作用域已有的变量。跨作用域的耗时 MUST 由两条日志的时间戳差还原。

#### Scenario: 框架 diff 不含结构性改动

- **WHEN** 审查 `pkg/async` 下的代码 diff
- **THEN** 改动只包含 `logs.*` 调用及其所需的局部变量
- **AND** 不存在结构体字段新增、签名变更、调度行为变更

### Requirement: 异步任务日志的 rid 打印约定

异步任务框架中涉及 task 的日志 MUST 同时打印执行器实例 rid 与 task rid，二者语义不同且都不可省略。

- `exec.kt.Rid`：在 `NewExecutor` 时创建一次，标识 task-server 进程/执行器实例，用于与 flow 表的 `worker` 字段互相印证。日志字段名 MUST 为 `exeRid`。
- `task.Kit.Rid`：由 scheduler 轮询 kit 经 `NewSubKit` 逐层派生，形如 `<轮询rid>/<flow后缀>/<task后缀>`，用于串联单进程内的 flow → task 调用链。日志字段名 MUST 为 `rid` 且位于消息末尾。

由于两者均不是调用方原始请求的 rid，涉及 task 的日志 MUST 同时包含 `taskID` 与 `flowID` 作为跨服务关联键。

#### Scenario: task 相关日志同时含两个 rid

- **WHEN** 在 executor 中打印任一条与具体 task 相关的日志
- **THEN** 该日志含 `exeRid: <executor rid>` 与 `rid: <task rid>`
- **AND** `rid` 位于消息末尾，`exeRid` 位于其之前
- **AND** 该日志含 `taskID` 与 `flowID`

#### Scenario: 现有失败日志的 rid 修正

- **WHEN** `executor.UpdateTask` 或 `workerDo` 打印失败日志
- **THEN** 不再仅使用 `exec.kt.Rid`，而是同时给出 `exeRid` 与 task rid

### Requirement: task 状态更新成功记录日志

task 状态更新成功后 MUST 打印 Info 级别日志，使任务状态迁移的时序可被还原。

#### Scenario: task 状态更新成功

- **WHEN** `executor.UpdateTask` 成功把 task 状态更新到目标状态
- **THEN** 打印 Info 日志，含 `taskID`、`flowID`、`actionName`、源状态与目标状态、`exeRid`、rid

#### Scenario: task action 执行结束

- **WHEN** 一次 task action 的 `Run` 返回
- **THEN** 打印 Info 日志，含 `taskID`、`flowID`、`actionName`、执行结果状态、执行耗时、`exeRid`、rid

### Requirement: flow 状态更新成功与 CAS 冲突记录日志

flow 状态更新 MUST 区分成功、CAS 未命中两种结果并分别打印。

`updateFlowState` 与 `updateFlowStateAndReason` 采用 CAS 更新，源状态不匹配时 DAO 返回 `errf.RecordNotUpdate`。该情况表示有其他执行路径抢先修改了状态，是重要的竞争线索，MUST 单独记录而非混在通用错误里。

#### Scenario: flow 状态更新成功

- **WHEN** flow 状态经 CAS 成功更新
- **THEN** 打印 Info 日志，含 `flowID`、`flowName`、源状态与目标状态、rid

#### Scenario: CAS 未命中

- **WHEN** flow 状态 CAS 更新影响行数为 0，返回 `errf.RecordNotUpdate`
- **THEN** 打印 Warn 日志，含 `flowID`、期望的源状态、目标状态、rid
- **AND** 不新增查询去获取库中实际状态；实际状态由同一 flowID 的其他状态更新日志还原

### Requirement: flow 派发与调度选中可观测

flow 从 Pending 到实际开始执行经过派发、调度选中、worker 排队三段等待，每段 MUST 可从日志得到等待时长。

`scheduler.runScheduledFlow` 中现有的 `logs.V(3)` 在生产 V 级别下不输出，MUST 补充 Info 级别日志。

#### Scenario: flow 被派发到执行节点

- **WHEN** dispatcher 把 flow 状态由 Pending 更新为 Scheduled 并分配 worker
- **THEN** 打印 Info 日志，含 `flowID`、`flowName`、分配到的 worker 节点、rid

#### Scenario: flow 被调度选中执行

- **WHEN** scheduler 从候选 flow 中选中某 flow 并将其状态更新为 Running
- **THEN** 打印 Info 日志，含 `flowID`、`flowName`、从 Scheduled 到被选中的等待秒数、本轮候选总数与选中数、rid

#### Scenario: 候选 flow 列表明细

- **WHEN** scheduler 完成一轮候选 flow 的打分与筛选
- **THEN** 候选明细以 `logs.V(3)` 打印，MUST NOT 使用 Info 级别

#### Scenario: task 在 initQueue 中的排队时长

- **WHEN** `watchInitQueue` 从 initQueue 取出一个 payload
- **THEN** 打印 Info 日志，含 `taskID`、`flowID`、`actionName`、由该作用域现成的 `payload.entryTime` 算出的 initQueue 等待秒数、rid

#### Scenario: task 进入 fast 或 slow 队列

- **WHEN** `initWorkerTask` 依据已有的 `task.ExecTime` 与 `fastTaskThresholdSec` 决定把 task 推入 fast 或 slow 队列
- **THEN** 打印 Info 日志，含 `taskID`、`flowID`、目标队列名、rid
- **AND** fast/slow 队列内的等待时长由本日志与 `workerDo` 已有的 `start execute task` 日志的时间戳差还原
- **AND** MUST NOT 为传递入队时刻而给 `Task` 结构体新增字段
