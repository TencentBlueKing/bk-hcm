/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package consumer

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/times"
)

// Executor （执行器）: 准备任务执行所需要的超时控制，共享数据等工具，并执行任务。
type Executor interface {
	compctrl.Closer

	// Start 启动执行器。
	Start()
	// SetGetSchedulerFunc 设置获取调度器函数，运行过程中，执行完的节点需要通过调度器获取子节点，且调度器会下发任务到执行器.
	SetGetSchedulerFunc(f func() Scheduler)

	// Push 推送task并执行。
	Push(flow *Flow, task *Task)
	// CancelTasks 关闭指定task_id的任务。
	CancelTasks(taskIDs []string) error
	CancelFlow(kt *kit.Kit, flowID string) error

	GetTaskTypeAvgExecTime(taskType enumor.ActionName) (float64, bool)
	GetFastTaskThresholdSec() float64
}

var _ Executor = new(executor)

// executor 定义任务执行器
type executor struct {
	kt *kit.Kit

	taskExecTimeoutSec  uint
	fastTaskWorkerRatio float64

	cancelMap             sync.Map
	workerWg              sync.WaitGroup
	initWg                sync.WaitGroup
	fastTaskQueue         chan *Task
	slowTaskQueue         chan *Task
	initQueue             *TaskInitQueue
	backend               backend.Backend
	taskTypeTimeWindowMap map[enumor.ActionName]*TimeWindow // 创建完成后这个map无锁，而是timewindow的细粒度锁
	ttTwMapMu             sync.RWMutex                      // 用于保护 taskTypeTimeWindowMap 的并发创建
	timeWindowDurationMin uint                              // 表示回溯多久的历史数据，单位分钟
	timeWindowCapacity    uint                              // TimeWindow 的容量，表示每一类tasktype最多存放的时间数据数量
	fastTaskWorkerNum     uint                              // 快任务worker数量
	sharedWorkerNum       uint                              // 共享worker数量
	fastTaskThresholdSec  float64                           // 快任务的阈值，执行时间小于这个值的任务为快任务，反之为慢，单位秒
	closeCh               chan struct{}
	mc                    *metric
	GetSchedulerFunc      func() Scheduler
}

// SetGetSchedulerFunc 设置获取调度器函数，运行过程中，执行完的节点需要通过调度器获取子节点，且调度器会下发任务到执行器.
func (exec *executor) SetGetSchedulerFunc(f func() Scheduler) {
	exec.GetSchedulerFunc = f
}

// NewExecutor 实例化任务执行器
func NewExecutor(kt *kit.Kit, bd backend.Backend, opt *ExecutorOption, mc *metric) Executor {
	return &executor{
		kt:                    kt,
		backend:               bd,
		workerWg:              sync.WaitGroup{},
		initWg:                sync.WaitGroup{},
		initQueue:             NewTaskInitQueue(opt.InitQueueCapacity, mc),
		taskTypeTimeWindowMap: make(map[enumor.ActionName]*TimeWindow),
		closeCh:               make(chan struct{}, 1),
		taskExecTimeoutSec:    opt.TaskExecTimeoutSec,
		fastTaskWorkerRatio:   opt.FastTaskWorkerRatio,
		fastTaskWorkerNum:     uint(float64(opt.WorkerNumber) * opt.FastTaskWorkerRatio),
		sharedWorkerNum:       opt.WorkerNumber - uint(float64(opt.WorkerNumber)*opt.FastTaskWorkerRatio),
		fastTaskThresholdSec:  opt.FastTaskThresholdSec,
		timeWindowCapacity:    opt.TimeWindowCapacity,
		timeWindowDurationMin: opt.TimeWindowDurationMin,
		fastTaskQueue:         make(chan *Task, opt.FastTaskQueueCapacity),
		slowTaskQueue:         make(chan *Task, opt.SlowTaskQueueCapacity),
		mc:                    mc,
	}
}

// Start 初始化执行器并启动执行
func (exec *executor) Start() {
	logs.Infof("executor start, total worker number: %d (fast task worker: %d, shared worker: %d)",
		exec.fastTaskWorkerNum+exec.sharedWorkerNum, exec.fastTaskWorkerNum, exec.sharedWorkerNum)

	// 待执行的任务预处理
	exec.initWg.Add(1)
	go exec.watchInitQueue()

	// 启动fast task worker，只从快任务队列取任务
	for i := 0; i < int(exec.fastTaskWorkerNum); i++ {
		exec.workerWg.Add(1)
		go exec.fastTaskWorker()
	}

	// 启动shared worker，优先从慢任务队列取任务，没有才从快任务队列取
	for i := 0; i < int(exec.sharedWorkerNum); i++ {
		exec.workerWg.Add(1)
		go exec.sharedTaskWorker()
	}
}

// fastTaskWorker 只从快任务队列取任务
func (exec *executor) fastTaskWorker() {
	for task := range exec.fastTaskQueue {
		if err := exec.workerDo(task); err != nil {
			// Task执行失败告警通知
			logs.Errorf("%s: executor fast task worker exec failed, err: %v, taskID: %s, action: %s, rid: %s",
				constant.AsyncTaskWarnSign, err, task.ID, task.ActionName, task.Kit.Rid)
		}
	}
	exec.workerWg.Done()
}

// sharedTaskWorker 共享任务工作器，实现优先级调度策略
// 设计思路：
// 1. 优先处理慢任务队列中的任务，确保耗时较长的任务能够及时得到处理
// 2. 当慢任务队列为空时，才处理快任务队列中的任务
// 3. 采用两阶段select策略，避免快任务饥饿慢任务的情况
//
// 调度策略：
// - 第一阶段：非阻塞检查慢任务队列，如果有任务立即处理
// - 第二阶段：如果慢任务队列为空，则阻塞等待任意队列的任务到达
func (exec *executor) sharedTaskWorker() {
	defer exec.workerWg.Done()
	for {
		// 第一阶段：非阻塞优先检查慢任务队列
		// 使用非阻塞select确保慢任务优先级，避免快任务抢占慢任务的执行机会
		select {
		case <-exec.closeCh:
			return
		case task := <-exec.slowTaskQueue:
			// 慢任务队列有任务，立即处理
			if err := exec.workerDo(task); err != nil {
				logs.Errorf("%s: executor shared task worker workerDo exec failed, err: %v, taskID: %s, "+
					"action: %s, rid: %s", constant.AsyncTaskWarnSign, err, task.ID, task.ActionName, task.Kit.Rid)
			}
			// 处理完成后继续下一轮循环，再次优先检查慢任务队列
			continue
		default:
			// 慢任务队列为空，进入第二阶段
		}

		// 第二阶段：阻塞等待任意队列的任务到达
		// 当慢任务队列为空时，公平地监听两个队列，哪个队列先来任务就先处理，处理完重新回到第一阶段
		select {
		case <-exec.closeCh:
			return
		case task := <-exec.slowTaskQueue:
			if err := exec.workerDo(task); err != nil {
				logs.Errorf("%s: executor shared task worker workerDo exec failed, err: %v, taskID: %s, "+
					"action: %s, rid: %s", constant.AsyncTaskWarnSign, err, task.ID, task.ActionName, task.Kit.Rid)
			}
		case task := <-exec.fastTaskQueue:
			if err := exec.workerDo(task); err != nil {
				logs.Errorf("%s: executor shared task worker workerDo exec failed, err: %v, taskID: %s, "+
					"action: %s, rid: %s", constant.AsyncTaskWarnSign, err, task.ID, task.ActionName, task.Kit.Rid)
			}
		}
	}
}

// 从initQueue优先队列获取待执行的任务协程
func (exec *executor) watchInitQueue() {
	for {
		// 阻塞等待任务
		payload, ok := exec.initQueue.Pop()
		// initQueue关闭且为空，直接退出
		if !ok {
			exec.initWg.Done()
			return
		}
		exec.initWorkerTask(payload.flow, payload.task)
	}
}

// 待执行任务的预处理函数
func (exec *executor) initWorkerTask(flow *Flow, task *Task) {
	if _, ok := exec.cancelMap.Load(task.ID); ok {
		logs.Warnf("%s: executor task %s is already running, rid: %s", constant.AsyncTaskWarnSign,
			task.ID, task.Kit.Rid)
		return
	}

	// 设置超时控制
	cancel := task.Kit.CtxWithTimeoutMS(int(exec.taskExecTimeoutSec) * 1000)

	// 设置共享数据更新函数
	flow.ShareData.Save = func(kt *kit.Kit, data *tableasync.ShareData) error {
		return exec.backend.BatchUpdateFlow(kt, []model.Flow{{ID: flow.ID, ShareData: data}})
	}

	// 设置task执行所需要的 kit，更新Task函数，所属流
	task.InitDep(run.NewExecuteContext(task.Kit, flow.ShareData), func(taskKit *kit.Kit, task *model.Task) error {
		return exec.backend.UpdateTask(exec.kt, task)
	}, flow)

	// cancel存储到cancelMap中
	exec.cancelMap.Store(task.ID, cancel)

	// 根据task所属tasktype平均执行时间将任务推送到对应的队列，第一次执行的tasktype先放慢任务队列
	if task.ExecTime >= exec.fastTaskThresholdSec {
		exec.slowTaskQueue <- task
		return
	}
	exec.fastTaskQueue <- task
}

// 任务执行体
func (exec *executor) workerDo(task *Task) (err error) {
	// cancelMap清理执行成功/失败的任务
	defer exec.cancelMap.Delete(task.ID)
	// 无论任务成功还是失败，都需要交给scheduler分析任务流的状态
	// 执行完的任务回写到scheduler用于获取待执行的任务
	defer exec.GetSchedulerFunc().EntryTask(task)
	var runErr error
	var failedRet any

	// 执行任务
	act, exist := action.GetAction(task.ActionName)
	if !exist {
		return fmt.Errorf("action: %s not found", task.ActionName)
	}

	if err := task.ValidateBeforeExec(act); err != nil {
		return err
	}
	logs.Infof("start execute task %s, actionID:%s, action: %s, flow: %s, rid: %s",
		task.ID, task.ActionID, task.ActionName, task.FlowID, task.Kit.Rid)
	defer func() {
		if fatalErr := recover(); fatalErr != nil {
			logs.Errorf("[hcm server panic], taskID: %s, flowID: %s, err: %v, rid: %s, debug strace: %s",
				task.ID, task.Flow.ID, err, task.Kit.Rid, debug.Stack())
			if fErr, ok := fatalErr.(error); ok {
				runErr = fErr
			} else {
				runErr = fmt.Errorf("panic: %v", fatalErr)
			}
		}
		if runErr == nil {
			return
		}
		err = runErr
		logs.Errorf("task %s run failed, err: %v, task: %+v, result: %+v, rid: %s",
			task.ID, runErr, task, failedRet, exec.kt.Rid)
		if errf.IsContextCanceled(runErr) {
			task.State = enumor.TaskCancel
			return
		}
		nextState := enumor.TaskFailed
		if patchErr := exec.UpdateTask(task, nextState, runErr.Error(), failedRet); patchErr != nil {
			logs.Errorf("task %s set %s state failed after run failed, err: %v, patchErr: %v, exeRid: %s, taskRid: %s",
				task.ID, nextState, runErr, patchErr, exec.kt.Rid, task.Kit.Rid)
			err = fmt.Errorf("task %s set %s state failed, after run failed, err: %v, patchErr: %v",
				task.ID, nextState, runErr, patchErr)
			return
		}
		return
	}()

	if !task.Retry.IsEnable() {
		_, failedRet, runErr = exec.runTaskOnce(task, act)
		return
	}
	if task.State == enumor.TaskRollback && task.Reason.RollbackCount >= task.Retry.Policy.Count {
		// 超过指定重试次数，置为失败
		runErr = fmt.Errorf("too many retries: %w", errors.New(task.Reason.Message))
		return
	}
	// 减去已经执行的count
	task.Retry.Policy.Count -= task.Reason.RollbackCount
	failedRet, runErr = task.Retry.Run(func() (stop bool, failRet any, err error) {
		needRetry, failRet, err := exec.runTaskOnce(task, act)
		if err == nil {
			return false, nil, nil
		}

		if !needRetry {
			return true, failRet, err
		}
		// 允许重试，将Task状态由 running -> rollback，进行回滚
		if patchErr := exec.UpdateTask(task, enumor.TaskRollback, err.Error(), failRet); patchErr != nil {
			e := fmt.Errorf("task set rollback state failed, after runAction failed, err: %v, patchErr: %v",
				err, patchErr)
			return false, failRet, e
		}
		return false, nil, nil
	})
	return nil
}

// runTaskOnce 只有执行Action运行逻辑失败才会允许重试，更改状态失败不进行重试。
// 如果执行成功直接写入状态和结果，失败时才将状态和结果返回到上层
func (exec *executor) runTaskOnce(task *Task, act action.Action) (needRetry bool, failedResult any, err error) {
	params, err := task.prepareParams(act)
	if err != nil {
		return false, nil, err
	}
	if task.State == enumor.TaskRollback {
		rollbackAct, ok := act.(action.RollbackAction)
		if !ok {
			return false, nil, fmt.Errorf("action: %s has no RollbackAction", act.Name())
		}

		if err = rollbackAct.Rollback(task.ExecuteKit, params); err != nil {
			return true, nil, fmt.Errorf("rollback failed, err: %v", err)
		}

		if err = exec.UpdateTaskState(task, enumor.TaskPending); err != nil {
			return false, nil, err
		}
	}

	if task.State == enumor.TaskPending {
		if err = exec.UpdateTaskState(task, enumor.TaskRunning); err != nil {
			return false, nil, err
		}
		taskStartTime := time.Now()
		result, err := act.Run(task.ExecuteKit, params)
		if err != nil {
			if errf.IsContextCanceled(err) {
				// 被取消不需要重试
				return false, result, err
			}
			return true, result, fmt.Errorf("run failed, err: %v, time: %s",
				err, times.ConvStdTimeNow())
		}

		// 只记录task执行成功的执行时间
		exec.getTimeWindow(task.ActionName).Push(time.Since(taskStartTime).Seconds())

		// 如果执行成功，返回 result 属于成功结果，设置成功状态时，同时设置成功结果。如果执行失败，
		// 结果属于失败结果，交与上层更新失败或回滚等操作，更新失败结果。
		if err = exec.UpdateTaskStateResult(task, enumor.TaskSuccess, result); err != nil {
			return false, result, err
		}
	}

	return false, nil, nil
}

// getTimeWindow 并发安全地获取或创建指定任务类型的 TimeWindow
func (exec *executor) getTimeWindow(taskType enumor.ActionName) *TimeWindow {
	// 只读
	exec.ttTwMapMu.RLock()
	tw, ok := exec.taskTypeTimeWindowMap[taskType]
	exec.ttTwMapMu.RUnlock()
	if ok {
		return tw
	}

	// 创建新 TimeWindow（只有第一次创建会走到这里）
	exec.ttTwMapMu.Lock()
	defer exec.ttTwMapMu.Unlock()
	// 再次检查，防止并发创建
	if tw, ok = exec.taskTypeTimeWindowMap[taskType]; ok {
		return tw
	}
	// 创建并写入
	tw = NewTimeWindow(exec.timeWindowCapacity, exec.timeWindowDurationMin)
	exec.taskTypeTimeWindowMap[taskType] = tw
	return tw
}

// Push 任务写入到initQueue
func (exec *executor) Push(flow *Flow, task *Task) {

	// try to exit the sender goroutine as early as possible.
	// try-receive and try-send select blocks are specially optimized by the standard Go compiler,
	// so they are very efficient.
	select {
	case <-exec.closeCh:
		logs.Infof("scheduler has already closed, so will not execute next task instances")
		return
	default:
	}

	err := exec.initQueue.Push(&InitPayload{
		flow:      flow,
		task:      task,
		entryTime: time.Now(),
	})
	if err != nil {
		logs.Errorf("fail to push InitPayload to task init queue, err: %v", err)
		return
	}
}

// InitPayload 任务初始化信息
type InitPayload struct {
	flow      *Flow
	task      *Task
	entryTime time.Time
}

// CancelTasks 停止指定id的任务
func (exec *executor) CancelTasks(taskIDs []string) error {
	for _, id := range taskIDs {
		if cancel, ok := exec.cancelMap.Load(id); ok {
			exec.cancelMap.Delete(id)
			cancel.(context.CancelFunc)()
		}
	}

	return nil
}

// CancelFlow 取消指定id flow
func (exec *executor) CancelFlow(kt *kit.Kit, flowId string) error {

	taskList, err := exec.backend.ListTask(kt, &backend.ListInput{
		Filter: tools.EqualExpression("flow_id", flowId),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return err
	}
	cancelIDs := make([]string, 0)
	for _, task := range taskList {
		switch task.State {

		case enumor.TaskPending, enumor.TaskInit, enumor.TaskRollback, enumor.TaskFailed, enumor.TaskRunning:
			// 	更新数据库状态
			err := exec.UpdateTask(&Task{Task: task}, enumor.TaskCancel, string(task.State), nil)
			logs.Errorf("fail to update task(%s) state for cancel, err: %v, rid: %s", task.ID, err, kt.Rid)
			cancelIDs = append(cancelIDs, task.ID)
		case enumor.TaskSuccess, enumor.TaskCancel:
			// 	跳过
		}
	}
	if len(cancelIDs) == 0 {
		return nil
	}
	// cancel 后在executor 中写回task canceled状态
	if err := exec.CancelTasks(cancelIDs); err != nil {
		logs.Errorf("fail to cancel task, err: %v, flow id: %s, task ids %v, rid: %s",
			err, flowId, cancelIDs, kt.Rid)
		return err
	}

	return nil
}

// Close 执行器关闭函数
func (exec *executor) Close() {

	logs.Infof("executor receive close cmd, start to close")

	defer close(exec.closeCh)
	exec.closeCh <- struct{}{}

	exec.initQueue.Close()
	exec.initWg.Wait()
	close(exec.fastTaskQueue)
	close(exec.slowTaskQueue)
	exec.workerWg.Wait()

	logs.Infof("executor close success")

}

// UpdateTaskState update task state.
func (exec *executor) UpdateTaskState(task *Task, state enumor.TaskState) error {
	return exec.UpdateTask(task, state, "", nil)
}

// UpdateTaskStateResult update task state and result.
func (exec *executor) UpdateTaskStateResult(task *Task, state enumor.TaskState, result interface{}) error {
	return exec.UpdateTask(task, state, "", result)
}

// UpdateTask update task, record state of cancel state
func (exec *executor) UpdateTask(task *Task, state enumor.TaskState, reason string, result interface{}) error {
	md, err := task.buildTaskUpdateModel(exec.kt, state, reason, result)
	if err != nil {
		return err
	}

	rty := retry.NewRetryPolicy(DefRetryCount, DefRetryRangeMS)
	err = rty.BaseExec(exec.kt, func() error {
		return exec.backend.UpdateTask(exec.kt, md)
	})
	if err != nil {
		logs.Errorf("task update state failed, err: %v, retryCount: %d, id: %s, state: %s, reason: %s, rid: %s",
			err, DefRetryCount, task.ID, state, reason, exec.kt.Rid)
		return err
	}

	task.State = state

	return nil
}

// GetTaskTypeAvgExecTime get task type avg exec time by corresponding timewindow
func (exec *executor) GetTaskTypeAvgExecTime(taskType enumor.ActionName) (avgExecTime float64, neverExec bool) {
	return exec.getTimeWindow(taskType).GetAvg()
}

// GetFastTaskThresholdSec get fast task threshold.
func (exec *executor) GetFastTaskThresholdSec() float64 {
	return exec.fastTaskThresholdSec
}
