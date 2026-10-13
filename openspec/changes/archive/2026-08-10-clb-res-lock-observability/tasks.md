## 1. 前置：rid 打印修正

- [x] 1.1 在 `pkg/async/consumer/executor.go` 的 `UpdateTask` 失败日志中，把 `exec.kt.Rid` 改为同时打印 `exeRid`（`exec.kt.Rid`）与 `rid`（`task.Kit.Rid`），`rid` 置于消息末尾；补上 `flowID` 字段
- [x] 1.2 在 `workerDo` 的 `task %s run failed` 日志中同样改为双 rid，并补 `flowID`、`actionName`
- [x] 1.3 核对 `workerDo` 中已有的 `exeRid` / `taskRid` 双打日志，统一字段名为 `exeRid` / `rid`（`rid` 在末尾）
- [x] 1.4 检索 `pkg/async/consumer` 下其余使用 `exec.kt.Rid` 的日志点，逐一改为双 rid 或改用 `task.Kit.Rid`
- [x] 1.5 新增 `taskRid(task *Task)` 判空辅助函数：`CancelFlow` 会构造不带 `Kit` 的临时 `Task`，直接取 `task.Kit.Rid` 会 panic
- [x] 1.6 修复 `CancelFlow` 中 `UpdateTask` 后无条件打 `Errorf` 的既有缺陷（`err == nil` 时也打），改为仅失败时打

## 2. data-service：加解锁日志

> 范围红线：不改任何签名、不新增结构体字段、不新增 DB 查询，只加日志语句。锁泄漏的识别放在 task-server 侧（见 7.3），那里已有现成的查询结果。

- [x] 2.1 `resource_flow.go` 的 `ResFlowLock` 事务成功后补 Info 日志：`resID`、`resType`、`owner`、`taskType`、`status`、rid
- [x] 2.2 `ResFlowUnLock` 事务成功后补 Info 日志：`resID`、`resType`、请求的 `owner` flowID、目标 `status`、rid
- [x] 2.3 `CreateResFlowLock` 成功后补 Info 日志（与 2.1 字段一致）
- [x] 2.4 修复两处既有日志的格式化参数个数不匹配（verb 多于 arg，实际会渲染成 `%!s(MISSING)`）：`CreateResFlowLock` 与 `BatchCreateResFlowRel` 的 `Errorf`；同时把 `err: %v` 调整到消息前部

## 4. cloud-server：加锁成功日志（三份同构实现全覆盖）

- [x] 4.1 `cmd/cloud-server/logics/load-balancer/res_lock.go` 的 `lockResFlowStatus` 在两步（加锁 + flow Init→Pending）均成功后补 Info 日志
- [x] 4.2 `cmd/cloud-server/service/load-balancer/async_target_group_add_rs.go:459` 的 `svc.lockResFlowStatus` 同样处理
- [x] 4.3 `cmd/cloud-server/logics/cvm/res_lock.go:38` 的 `lockResFlowStatus` 同样处理

## 5. cloud-server：撞锁记录持锁 owner

> 范围红线：`checkResFlowRel` 三份实现的**签名与返回的 `errf` 错误消息一律不动**。

> **实施决策（与原计划的偏差，已经用户确认）**：日志集中在 `checkResFlowRel` 内部的三个分支，**不再逐个改造 ~20 个调用点**。原因：中心日志能打出 `resID`/`resType`/`owner`/`lockCreatedAt`/`rid`，比调用点能提供的信息更全；调用点入口可由同一 rid 关联访问日志还原；且避免 20 个文件的 `_, err :=` churn。核查确认 25 个调用点中 17 个本就有错误日志，另 8 个静默 `return err` 的场景已被中心日志覆盖。

- [x] 5.1 ~~逐个改造 `_, err :=` 调用方~~ → 改为中心化实现，见 5.4
- [x] 5.2 ~~同上改造 `service/load-balancer/` 下调用方~~ → 改为中心化实现，见 5.4
- [x] 5.3 已使用 `lockRel` 返回值的调用方：核查 `logics/cvm/power.go:96`、`logics/cvm/reset.go:232` 已用 `PtrToVal(lockRel)` 打出持锁记录，无需改动
- [x] 5.4 `checkResFlowRel` 内部：为区分「锁表命中」与「rel 记录未收尾」，在两个分支各补一条 Error 日志（不改返回值与错误消息）。三份实现均已处理：`logics/load-balancer/res_lock.go`、`service/load-balancer/async_target_group_add_rs.go`、`logics/cvm/res_lock.go`
- [x] 5.5 `service/load-balancer/query.go:434`（`getLoadBalancerLockStatus`）确认为前端轮询锁状态的查询接口，其自身已有 `Errorf`，不额外加日志
- [x] 5.6 **成功路径可见性**：三份 `checkResFlowRel` 的 `return nil, nil` 处补 Info 日志（`resID`、`resType`、rid），使「上一个 flow 已完成解锁、资源空闲」这一时刻可见，用于与解锁日志对比延迟。资源空闲时前端会停止轮询，故该 Info 主要在真实操作入口触发

## 6. watch flow 与主 flow 的关联日志

- [x] 6.1 改造以 `_, err = ...CreateTemplateFlow` 丢弃 watch flowID 的位置，接收返回值并打 Info 日志（watch flowID、主 flowID、`resID`、`resType`、`taskType`）：`logics/load-balancer/create_layer4_listener_executor.go:269`、`create_layer7_listener_executor.go:461`、`create_url_rule_executor.go:277`、`layer4_listener_bind_rs_executor.go:294`、`layer7_listener_bind_rs_executor.go:367`、`delete_listener_executor.go:506`、`modify_listener_rs_weight_executor.go:508`、`unbind_listener_rs_executor.go:452`
- [x] 6.2 同上改造 `service/load-balancer/flow.go:69` 与 `service/load-balancer/async.go:177`（后者额外打出 `clonedFromFlowID`，因为它是 CloneFlow 重试路径，主 flowID 与请求中的 flowID 不同）
- [x] 6.3 同上改造 `logics/cvm/reset.go:352`

## 7. task-server：watch action 生命周期与解锁延迟

- [x] 7.1 `cmd/task-server/logics/flow/flow_slave_operate_watch.go` 的 `Run` 入口补 Info 日志：主 flowID、`resID`、`resType`、`taskType`、rid
- [x] 7.2 `processResFlow` 观察到主 flow 终态时补 Info 日志：终态值、`flowInfo.UpdatedAt` 到当前时刻的秒数（`terminalObservedDelaySec`）。新增 `elapsedSecSince(timeStr string)` 辅助函数解析 RFC3339 时间字段，解析失败返回 -1 表示不可用
- [x] 7.3 `processResFlow` 终态分支中 `queryResFlowLock` 返回空、当前静默 `return true, nil` 的位置补 Warn 日志：主 flow 已终态但该 flow 名下无锁记录，含主 flowID、`resID`、`resType`、终态值、rid。该查询已按 `owner = opt.FlowID` 过滤，结果现成可用，不新增任何查询
- [x] 7.4 解锁完成后补 Info 日志（`terminalToUnlockSec`、目标 `resStatus`）。因 `processUnlockResFlow` 签名不可动，日志放在 `processResFlow` 内调用返回处——那里 `flowInfo` 在作用域内；同时把原先 `return true, err` 的静默失败补上 Error 日志
- [x] 7.5 `Run` 的轮询循环内不新增 Info 日志
- [ ] 7.6 ~~`FlowInit` 分支「锁未创建、继续等待」的现有 Warn 改为一次 watch 执行内只打一次~~ → **有意未做**：该 Warn 位于 `processResFlow` 内，而去重标记只能放在 `Run` 的循环里，传递它必须改 `processResFlow` 签名，违反本变更红线。保持原样的风险可接受：该分支仅在「先建 watch 后加锁」的瞬时窗口触发，正常为 0~1 次；若锁始终未创建则属真实故障，此时 5 分钟上限内最多 600 行，且 7.7 的超时 Error 已给出完整上下文
- [x] 7.7 `Run` 超时分支的 `wait timeout` 改为先打 Error 日志（含主 flowID、`resID`、`resType`、已等待时长、rid）再返回错误
- [x] 7.8 `processResFlow` 中 `updateFlowStateByCAS` 的失败日志补 `resID`、`resType` 与 rid（原日志**完全没有 rid**）；`updateFlowStateByCAS` 自身补 CAS 成功日志

## 8. 异步框架：状态更新与调度链路日志

> 范围红线：`Task`、`Flow`、`InitPayload` 等框架结构体**一律不加字段**；不为打日志新增 DB 查询。只能用当前作用域已有的变量。

- [x] 8.1 `pkg/async/consumer/executor.go` 的 `UpdateTask` 成功后补 Info 日志：`taskID`、`flowID`、`actionName`、源状态→目标状态、双 rid
- [x] 8.2 `runTaskOnce` 中 action `Run` 返回后补 Info 日志：结果状态、执行耗时（已有 `cost` 局部变量）、`taskID`、`flowID`、`actionName`、双 rid；失败分支同样补上（原先仅打指标不打日志）
- [x] 8.3 `watchInitQueue` 从 `initQueue.Pop()` 取到 payload 后补 Info 日志：`taskID`、`flowID`、`actionName`、`time.Since(payload.entryTime)` 即 initQueue 等待秒数、rid。`entryTime` 在该作用域内现成可用，**不得**为把它传到 `workerDo` 而给 `Task` 加字段
- [x] 8.4 `initWorkerTask` 在推入 fast/slow 队列前补 Info 日志：`taskID`、`flowID`、目标队列（由已有的 `task.ExecTime >= exec.fastTaskThresholdSec` 判定）、rid。fast/slow 队列内的等待时长由本条日志与 `workerDo` 已有的 `start execute task` 日志的时间戳差得出，不额外埋点
- [x] 8.5 `pkg/async/consumer/scheduler.go` 的 `updateFlowStateAndReason` 成功后补 Info 日志：`flowID`、源→目标、rid（该函数原先**成功失败都无任何日志**，所有 flow 状态流转都是静默的）
- [x] 8.6 `updateFlowStateAndReason` 遇 `errf.RecordNotUpdate` 时打 Warn，含 `flowID`、期望的源状态、目标状态、rid。**不新增查询**去捞库中实际状态——CAS 失败本身即说明源状态已被他人改写，实际状态由同一 flowID 的其他状态更新日志（8.5）还原
- [x] 8.7 `runScheduledFlow` 中 flow 被选中并更新为 Running 后补 Info 日志：`flowID`、`flowName`、`scheduledWaitSec`（由已有的 `flow.UpdatedAt` 按 RFC3339 解析算出，解析失败取 -1）、本轮候选总数与选中数；同时给该处已有的失败日志补上 `flowID`
- [x] 8.8 `runScheduledFlow` 中候选明细相关的现有日志保持 `logs.V(3)`，不提级
- [x] 8.9 `pkg/async/consumer/dispatcher.go` 的 `Do` 在 `BatchUpdateFlowStateByCAS` 成功后补 Info 日志：逐个 `flowID`、`flowName`、分配到的 worker、rid

## 9. 校验

- [x] 9.1 `gofmt -l` 检查改动文件通过（无输出），导入分组符合规范，无超 120 列
- [x] 9.2 `go build ./...` 全量编译通过（首次遗漏 `scheduler.go` 的 `errf` 导入，已补）
- [x] 9.3 `go vet ./cmd/... ./pkg/async/...`：告警全部落在未改动文件（woa-server 测试、mongo 未键化字段、hc-service 非常量格式串、`leader_change_handler.go` unreachable code），属既有基线，改动文件零告警
- [x] 9.4 **范围核对**：`git diff` 确认 19 个文件 +236/-32，无签名变更、无结构体字段新增、无为打日志而加的 DB 查询、无对外错误消息变更、无同构实现合并。除日志语句外的改动仅为「把 `_, err :=` 改为接收已有返回值」以及两个新增的判空/解析辅助函数（`taskRid`、`elapsedSecSince`）
- [x] 9.5 逐条比对：每条新增日志的格式化 verb 数与实参数一致（`go vet` 不把 `logs.Errorf` 识别为 printf wrapper，故人工核对），`rid: %s` 均在消息末尾、`err: %v` 均在消息前部
- [x] 9.6 日志增量：单次 CLB 操作新增 Info 约 6~8 条（加锁 1、预检查通过 1、watch flow 关联 1、watch 启动 1、观察到终态 1、解锁成功 1，加上按 task 数量计的状态更新与队列日志）。轮询路径未新增 Info；唯一保留的高频 Warn 见 7.6 说明
- [x] 9.7 `go test ./pkg/async/... ./cmd/task-server/logics/flow/... ./cmd/cloud-server/logics/...` 全部通过
- [ ] 9.8 **待人工执行**：测试环境跑一次「建监听器 + 绑 RS」两步流程，核对日志能否还原出持锁 owner、watch flowID 与主 flowID 的对应、主 flow 终态到解锁的延迟、flow 三段等待时长
