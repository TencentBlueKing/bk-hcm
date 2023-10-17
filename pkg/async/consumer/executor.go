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
	"sync"

	"hcm/pkg/async/action/run"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/criteria/constant"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
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
}

var _ Executor = new(executor)

// executor 定义任务执行器
type executor struct {
	workerNumber       uint
	taskExecTimeoutSec uint

	cancelMap   sync.Map
	workerWg    sync.WaitGroup
	initWg      sync.WaitGroup
	workerQueue chan *Task
	initQueue   chan *initPayload
	backend     backend.Backend

	closeCh chan struct{}

	GetSchedulerFunc func() Scheduler
}

// SetGetSchedulerFunc 设置获取调度器函数，运行过程中，执行完的节点需要通过调度器获取子节点，且调度器会下发任务到执行器.
func (exec *executor) SetGetSchedulerFunc(f func() Scheduler) {
	exec.GetSchedulerFunc = f
}

// NewExecutor 实例化任务执行器
func NewExecutor(bd backend.Backend, opt *ExecutorOption) Executor {
	return &executor{
		backend:            bd,
		workerWg:           sync.WaitGroup{},
		initWg:             sync.WaitGroup{},
		workerQueue:        make(chan *Task, 10),
		initQueue:          make(chan *initPayload),
		closeCh:            make(chan struct{}, 1),
		workerNumber:       opt.WorkerNumber,
		taskExecTimeoutSec: opt.TaskExecTimeoutSec,
	}
}

// Start 初始化执行器并启动执行
func (exec *executor) Start() {

	logs.Infof("executor start, worker number: %d", exec.workerNumber)

	// 待执行的任务预处理
	exec.initWg.Add(1)
	go exec.watchInitQueue()

	// 启动workerNumber个执行器执行任务
	for i := 0; i < int(exec.workerNumber); i++ {
		exec.workerWg.Add(1)
		go exec.subWorkerQueue()
	}
}

// 从initQueue队列获取待执行的任务协程
func (exec *executor) watchInitQueue() {
	for p := range exec.initQueue {
		exec.initWorkerTask(p.flow, p.task)
	}

	exec.initWg.Done()
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
	task.InitDep(run.NewExecuteContext(task.Kit, flow.ShareData), func(kt *kit.Kit, task *model.Task) error {
		return exec.backend.UpdateTask(kt, task)
	}, flow)

	// cancel存储到cancelMap中
	exec.cancelMap.Store(task.ID, cancel)
	// 任务写回workerQueue
	exec.workerQueue <- task
}

// 任务实际执行协程
func (exec *executor) subWorkerQueue() {
	for task := range exec.workerQueue {
		if err := exec.workerDo(task); err != nil {
			// Task执行失败告警通知
			logs.Errorf("%s: executor sub worker workerDo exec failed, err: %v, rid: %s",
				constant.AsyncTaskWarnSign, err, task.Kit.Rid)
		}
	}

	exec.workerWg.Done()
}

// 任务执行体
func (exec *executor) workerDo(task *Task) (err error) {

	// cancelMap清理执行成功/失败的任务
	defer exec.cancelMap.Delete(task.ID)

	// 执行任务
	if err = task.Run(); err != nil {
		logs.Errorf("task run failed, err: %v, task: %+v, rid: %s", err, task, task.Kit.Rid)

		// 无论任务成功还是失败，都需要交给调度器分析任务流的状态
	}

	// 执行完的任务回写到调度器用于获取待执行的任务
	exec.GetSchedulerFunc().EntryTask(task)

	return err
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

	exec.initQueue <- &initPayload{
		flow: flow,
		task: task,
	}
}

type initPayload struct {
	flow *Flow
	task *Task
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

// Close 执行器关闭函数
func (exec *executor) Close() {

	logs.Infof("executor receive close cmd, start to close")

	defer close(exec.closeCh)
	exec.closeCh <- struct{}{}

	close(exec.initQueue)
	exec.initWg.Wait()
	close(exec.workerQueue)
	exec.workerWg.Wait()

	logs.Infof("executor close success")

}
