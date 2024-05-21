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
	"fmt"
	"sync"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

/*
Scheduler （调度器）: TODO: 换为 捕获器、消费器，添加假死任务销毁逻辑
 1. 获取分配给当前节点的处于Scheduled状态的任务流，构建任务流树，将待执行任务推送到执行器执行。
 2. 分析执行器执行完的任务，判断任务流树状态，如果任务流处理完，更新状态，否则将子节点推送到执行器执行。
*/
type Scheduler interface {
	compctrl.Closer

	// Start 启动调度器。
	Start()

	// EntryTask 分析执行完的任务，并解析出当前任务的子任务去执行。
	EntryTask(task *Task)
	// DeleteFlowTaskTree 清空任务树，阻止继续调度
	DeleteFlowTaskTree(flowID string)
}

// scheduler 定义任务流调度器
type scheduler struct {
	workerNumber     uint
	watchIntervalSec time.Duration

	taskTrees   sync.Map
	workerQueue chan *Task
	workerWg    sync.WaitGroup

	backend  backend.Backend
	executor Executor
	leader   leader.Leader

	closeCh chan struct{}
}

// NewScheduler 实例化任务流调度器
func NewScheduler(bd backend.Backend, exec Executor, ld leader.Leader, opt *SchedulerOption) Scheduler {

	return &scheduler{
		closeCh:          make(chan struct{}),
		workerWg:         sync.WaitGroup{},
		workerQueue:      make(chan *Task, 10),
		workerNumber:     opt.WorkerNumber,
		watchIntervalSec: time.Duration(opt.WatchIntervalSec) * time.Second,
		backend:          bd,
		executor:         exec,
		leader:           ld,
	}
}

// Start 初始化调度器并启动执行
func (sch *scheduler) Start() {

	logs.Infof("scheduler start, worker number: %d, interval: %v", sch.workerNumber, sch.watchIntervalSec)

	// 定期获取等待执行的任务流
	sch.workerWg.Add(2)
	go sch.scheduledFlowWatcher()
	go sch.canceledFlowWatcher()

	// 启动workerNumber个协程进行任务流解析
	for i := 0; i < int(sch.workerNumber); i++ {
		sch.workerWg.Add(1)
		go sch.goWorker()
	}
}

// flowWatcher 定期查询调度到该节点的flow
func (sch *scheduler) scheduledFlowWatcher() {
	for {
		select {
		case <-sch.closeCh:
			break
		default:
		}
		// Kit: Kit initiate, 每次执行创建新kit
		kt := NewKit()
		if err := sch.runScheduledFlow(kt); err != nil {
			logs.Errorf("%s: scheduler watch scheduled flow  failed, err: %v, rid: %s",
				constant.AsyncTaskWarnSign, err, kt.Rid)
		}

		time.Sleep(sch.watchIntervalSec)
	}

	sch.workerWg.Done()
}

// queryCurrNodeFlow 查询主节点分配给当前节点处于 Scheduled 状态的任务流。
func (sch *scheduler) queryCurrNodeFlow(kt *kit.Kit, state enumor.FlowState, limit int32) (
	[]model.Flow, error) {

	if limit > int32(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("limit should <= %d", core.DefaultMaxPageLimit)
	}
	input := &backend.ListInput{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("state", state),
			tools.RuleEqual("worker", sch.leader.CurrNode())),
		Page: &core.BasePage{
			Start: 0,
			Limit: uint(limit),
		},
	}
	result, err := sch.backend.ListFlow(kt, input)
	if err != nil {
		logs.Errorf("list flows failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return result, nil
}

// canceledFlowWatcher 查询当前节点上被取消的flow并执行task取消操作
func (sch *scheduler) canceledFlowWatcher() {
	for {
		select {
		case <-sch.closeCh:
			break
		default:
		}
		// Kit: Kit initiate, 每次执行创建新kit
		kt := NewKit()
		if err := sch.handleCanceledFlow(kt); err != nil {
			logs.Errorf("%s: scheduler watch canceled failed, err: %v, rid: %s",
				constant.AsyncTaskWarnSign, err, kt.Rid)
		}

		time.Sleep(sch.watchIntervalSec)
	}

	sch.workerWg.Done()
}

func (sch *scheduler) handleCanceledFlow(kt *kit.Kit) error {

	dbFlows, err := sch.queryCurrNodeFlow(kt, enumor.FlowCancel, listScheduledFlowLimit)
	if err != nil {
		logs.Errorf("fail to list canceled flow, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if len(dbFlows) == 0 {
		return nil
	}
	for _, flow := range dbFlows {
		logs.Infof("canceling flow: %s", flow.ID)
		// 清空任务树，阻止继续调度
		sch.DeleteFlowTaskTree(flow.ID)
		err = updateFlowToCancel(kt, sch.backend, flow.ID, cvt.PtrToVal(flow.Worker), enumor.FlowCancel)
		if err != nil {
			logs.Errorf("fail to update flow clear worker id, err: %v, flow id: %s rid: %s",
				err, flow.ID, kt.Rid)
			// keep canceling other flow
			continue
		}

		if err := sch.executor.CancelFlow(kt, flow.ID); err != nil {
			logs.Errorf("fail to handle flow canceling, err: %v, flow id: %s, rid: %s", err, flow.ID, kt.Rid)
			// keep canceling other flow
		}
	}
	return nil
}

// listTaskByFlowID 查询当前FlowID全部的任务节点
func listTaskByFlowID(kt *kit.Kit, bd backend.Backend, flowID string) ([]*Task, error) {

	input := &backend.ListInput{
		Filter: tools.EqualExpression("flow_id", flowID),
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	tasks := make([]*Task, 0)
	for {
		result, err := bd.ListTask(kt, input)
		if err != nil {
			logs.Errorf("list task failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result {
			tasks = append(tasks, &Task{
				Task: one,
				// Note: second sub kit, flow -> task
				Kit: kt.NewSubKit(),
			})
		}

		if len(result) < int(core.DefaultMaxPageLimit) {
			break
		}

		input.Page.Start += uint32(input.Page.Limit)
	}

	return tasks, nil
}

// listTaskByFlowID 查询当前FlowID全部的任务节点
func listTaskByIDs(kt *kit.Kit, bd backend.Backend, ids []string) ([]*Task, error) {

	split := slice.Split(ids, int(core.DefaultMaxPageLimit))
	input := &backend.ListInput{
		Page: core.NewDefaultBasePage(),
	}
	tasks := make([]*Task, 0, len(ids))
	for _, partIDs := range split {
		input.Filter = tools.ContainersExpression("id", partIDs)
		result, err := bd.ListTask(kt, input)
		if err != nil {
			logs.Errorf("list task failed, err: %v, ids: %+v, rid: %s", err, partIDs, kt.Rid)
			return nil, err
		}

		for _, one := range result {
			tasks = append(tasks, &Task{
				Task: one,
				Kit:  kt.NewSubKit(),
			})
		}
	}

	return tasks, nil
}

// parseAndPushFlow 解析Flow并推送Flow下一批待执行的节点到执行器。
func (sch *scheduler) parseFlowAndPushTask(kt *kit.Kit, flow *Flow) error {
	// 根据任务流ID获取对应的任务集合
	tasks, err := listTaskByFlowID(kt, sch.backend, flow.ID)
	if err != nil {
		if stateErr := updateFlowStateAndReason(kt, sch.backend, flow.ID, enumor.FlowRunning, enumor.FlowFailed,
			err.Error()); stateErr != nil {

			logs.Errorf("update flow state and reason failed, after list task by flow id failed, err: %v, rid: %s",
				stateErr, kt.Rid)
		}

		logs.Errorf("list task by flow id failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 构造执行流树
	root, err := BuildTaskRoot(tasks)
	if err != nil {
		if stateErr := updateFlowStateAndReason(kt, sch.backend, flow.ID, enumor.FlowRunning, enumor.FlowFailed,
			err.Error()); stateErr != nil {

			logs.Errorf("update flow state and reason failed, after build task root failed, err: %v, rid: %s",
				stateErr, kt.Rid)
		}

		logs.Errorf("build task root failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	taskTree := &TaskTree{
		Flow: flow,
		Root: root,
	}

	// 获取可执行的节点
	executableTaskNodes := taskTree.Root.GetExecutableTasks()
	if len(executableTaskNodes) == 0 {
		state := taskTree.Root.ComputeState()

		if state == enumor.FlowSuccess {
			if err = updateFlowState(kt, sch.backend, flow.ID, enumor.FlowRunning, state); err != nil {
				logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}
		}

		if state == enumor.FlowFailed {
			if err = updateFlowStateAndReason(kt, sch.backend, flow.ID, enumor.FlowRunning, state,
				ErrSomeTaskExecFailed); err != nil {

				logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}
		}

		return nil
	}

	// 存储任务流执行树
	sch.taskTrees.Store(flow.ID, taskTree)

	taskIDMap := make(map[string]*Task, len(tasks))
	for _, one := range tasks {
		taskIDMap[one.ID] = one
	}

	// 所有可执行任务推送到执行器
	flow.State = enumor.FlowRunning
	for _, taskID := range executableTaskNodes {
		sch.executor.Push(flow, taskIDMap[taskID])
	}

	return nil
}

// updateFlowState 更新Flow状态，采用CAS加三次重试。source原状态，dest目标状态。
func updateFlowState(kt *kit.Kit, bd backend.Backend, flowID string, source,
	dest enumor.FlowState) error {

	return updateFlowStateAndReason(kt, bd, flowID, source, dest, "")
}

// updateFlowState 更新Flow状态和原因，采用CAS加三次重试。source原状态，dest目标状态。
func updateFlowStateAndReason(kt *kit.Kit, bd backend.Backend, flowID string, source, dest enumor.FlowState,
	reason string) error {

	info := backend.UpdateFlowInfo{
		ID:     flowID,
		Source: source,
		Target: dest,
	}
	if len(reason) != 0 {
		info.Reason = &tableasync.Reason{
			PreState: string(source),
			Message:  reason,
		}
	}

	rty := retry.NewRetryPolicy(DefRetryCount, DefRetryRangeMS)
	err := rty.BaseExec(kt, func() error {
		return bd.BatchUpdateFlowStateByCAS(kt, []backend.UpdateFlowInfo{info})
	})
	if err != nil {
		return err
	}

	return nil
}

// updateFlowToCancel 状态改为取消，清空 worker字段,
func updateFlowToCancel(kt *kit.Kit, bd backend.Backend, flowId, oldWorkerID string, source enumor.FlowState) error {

	info := backend.UpdateFlowInfo{
		ID:     flowId,
		Source: source,
		Target: enumor.FlowCancel,
		Reason: &tableasync.Reason{
			Message:  "canceled from " + oldWorkerID,
			PreState: string(source),
		},
		Worker: cvt.ValToPtr(""),
	}

	rty := retry.NewRetryPolicy(DefRetryCount, DefRetryRangeMS)
	err := rty.BaseExec(kt, func() error {
		return bd.BatchUpdateFlowStateByCAS(kt, []backend.UpdateFlowInfo{info})
	})
	if err != nil {
		logs.Errorf("fail to update flow state to cancel, err: %v, flow id: %s, rid: %s", err, flowId, kt.Rid)
		return err
	}

	return nil
}

// watchScheduledFlow 查询主节点分配给当前节点的的flow，并执行
func (sch *scheduler) runScheduledFlow(kt *kit.Kit) error {

	// 从DB中获取一条待执行的任务流并更新状态为执行中
	dbFlows, err := sch.queryCurrNodeFlow(kt, enumor.FlowScheduled, listScheduledFlowLimit)
	if err != nil {
		logs.Errorf("")
		return err
	}

	flows := slice.Map(dbFlows, func(one model.Flow) *Flow {
		// Note: first sub kit, scheduler.watcher -> flow
		return &Flow{Flow: one, Kit: kt.NewSubKit()}
	})

	if len(flows) == 0 {
		logs.V(3).Infof("current node: %s not found scheduled flow to handleRunningFlow, rid: %s",
			sch.leader.CurrNode(), kt.Rid)
		return nil
	}

	for _, flow := range flows {
		if err = updateFlowState(flow.Kit, sch.backend, flow.ID, enumor.FlowScheduled, enumor.FlowRunning); err != nil {
			logs.Errorf("update flow state failed, err: %v, rid: %s", err, flow.Kit.Rid)
			return err
		}

		if err = sch.parseFlowAndPushTask(flow.Kit, flow); err != nil {
			logs.Errorf("parse flow and push task failed, err: %v, rid: %s", err, flow.Kit.Rid)
			return err
		}
	}

	return nil
}

// 任务流解析协程
func (sch *scheduler) goWorker() {
	for task := range sch.workerQueue {
		if err := sch.executeNext(task.Flow.Kit, task); err != nil {
			logs.Errorf("%s: scheduler exec executeNext failed, err: %v, rid: %s", constant.AsyncTaskWarnSign,
				err, task.Kit.Rid)
		}
	}

	sch.workerWg.Done()
}

// 任务流解析函数体，根据任务获取下次可执行的任务集合
func (sch *scheduler) executeNext(kt *kit.Kit, task *Task) error {
	if task.State == enumor.TaskCancel {
		// skip canceled task
		return nil
	}
	tree, ok := sch.getTaskTree(task.FlowID)
	if !ok {
		logs.Errorf("execute next get task tree failed, flowID: %s, rid: %s", task.FlowID, kt.Rid)
		return fmt.Errorf("flow: %s not found", task.FlowID)
	}

	// 获取下次执行的任务
	executableIds := tree.Root.GetNextExecutableTaskNodes(task)
	if len(executableIds) == 0 {
		// 没有可执行的节点了，计算整棵树的执行状态，更新flow结果
		state := tree.Root.ComputeState()
		switch state {
		case enumor.FlowSuccess:
			if err := updateFlowState(kt, sch.backend, task.FlowID, enumor.FlowRunning, state); err != nil {
				logs.Errorf("update flow state to `%s` failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}

			sch.DeleteFlowTaskTree(task.FlowID)
		case enumor.FlowFailed:
			if err := updateFlowStateAndReason(kt, sch.backend, task.FlowID, enumor.FlowRunning, state,
				ErrSomeTaskExecFailed); err != nil {

				logs.Errorf("update flow state to `%s` failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}

			sch.DeleteFlowTaskTree(task.FlowID)
		case enumor.FlowCancel:
			// task cancel 一般是由flow状态触发，因此这里不处理
		}

		return nil
	}

	return sch.pushTasks(kt, tree.Flow, executableIds)
}

func (sch *scheduler) pushTasks(kt *kit.Kit, flow *Flow, ids []string) error {

	tasks, err := listTaskByIDs(kt, sch.backend, ids)
	if err != nil {
		logs.Errorf("list task by ids failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return err
	}

	// 可执行任务推送到执行器
	for _, one := range tasks {
		sch.executor.Push(flow, one)
	}

	return nil
}

// 获取存储的任务流树
func (sch *scheduler) getTaskTree(flowID string) (*TaskTree, bool) {
	tasks, ok := sch.taskTrees.Load(flowID)
	if !ok {
		return nil, false
	}

	return tasks.(*TaskTree), true
}

// EntryTask 任务写回到执行器用于获取下一批可执行的任务
func (sch *scheduler) EntryTask(taskNode *Task) {
	sch.workerQueue <- taskNode
}

// Close 调度器关闭函数
func (sch *scheduler) Close() {

	logs.Infof("scheduler receive close cmd, start to close")

	select {
	case <-sch.closeCh:
		logs.V(3).Infof("scheduler already closed")
		return
	default:
	}

	close(sch.closeCh)
	close(sch.workerQueue)

	sch.workerWg.Wait()

	logs.Infof("scheduler receive close cmd, start to close")

}

// DeleteFlowTaskTree 清空任务树，阻止继续调度
func (sch *scheduler) DeleteFlowTaskTree(flowID string) {

	sch.taskTrees.Delete(flowID)
}
