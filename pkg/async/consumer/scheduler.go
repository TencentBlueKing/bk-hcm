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
	"hcm/pkg/async/closer"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

/*
Scheduler （调度器）:
	1. 获取分配给当前节点的处于Scheduled状态的任务流，构建任务流树，将待执行任务推送到执行器执行。
	2. 分析执行器执行完的任务，判断任务流树状态，如果任务流处理完，更新状态，否则将子节点推送到执行器执行。
*/
type Scheduler interface {
	closer.Closer

	// Start 启动调度器。
	Start()

	// EntryTask 分析执行完的任务，并解析出当前任务的子任务去执行。
	EntryTask(task *Task)
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
	sch.workerWg.Add(1)
	go sch.startWatcher(sch.watchScheduledFlow)

	// 启动workerNumber个协程进行任务流解析
	for i := 0; i < int(sch.workerNumber); i++ {
		sch.workerWg.Add(1)
		go sch.goWorker()
	}
}

// startWatcher 定期执行do函数体
func (sch *scheduler) startWatcher(do func(kt *kit.Kit) error) {
	ticker := time.NewTicker(sch.watchIntervalSec)
	defer ticker.Stop()

	for {
		select {
		case <-sch.closeCh:
			break
		case <-ticker.C:
			kt := NewKit()
			if err := do(kt); err != nil {
				logs.Errorf("%s: scheduler watcher do failed, err: %v, rid: %s", constant.AsyncTaskWarnSign, err, kt.Rid)
			}
		}
	}

	sch.workerWg.Done()
}

// queryCurrNodeFlow 查询主节点分配给当前节点处于 Scheduled 状态的任务流。
func (sch *scheduler) queryCurrNodeFlow(kt *kit.Kit, limit int32) ([]*Flow, error) {

	if limit > int32(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("limit should <= %d", core.DefaultMaxPageLimit)
	}

	input := &backend.ListInput{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.Equal.Factory(),
					Value: enumor.FlowScheduled,
				},
				&filter.AtomRule{
					Field: "worker",
					Op:    filter.Equal.Factory(),
					Value: sch.leader.CurrNode(),
				},
			},
		},
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

	flows := make([]*Flow, 0)
	for _, one := range result {
		flows = append(flows, &Flow{
			Flow: one,
			Kit:  kt.NewSubKit(),
		})
	}

	return flows, nil
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
				Kit:  kt.NewSubKit(),
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

		if state == enumor.FlowSuccess || state == enumor.FlowFailed {
			if err = updateFlowState(kt, sch.backend, flow.ID, enumor.FlowRunning, state); err != nil {
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

	// 可执行任务推送到执行器
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
			Message: reason,
		}
	}

	rty := retry.NewRetryPolicy(defRetryCount, defRetryRangeMS)
	err := rty.BaseExec(kt, func() error {
		return bd.BatchUpdateFlowStateByCAS(kt, []backend.UpdateFlowInfo{info})
	})
	if err != nil {
		return err
	}

	return nil
}

// watchScheduledFlow 查询主节点分配给当前节点的
func (sch *scheduler) watchScheduledFlow(kt *kit.Kit) error {

	// 从DB中获取一条待执行的任务流并更新状态为执行中
	flows, err := sch.queryCurrNodeFlow(kt, listScheduledFlowLimit)
	if err != nil {
		logs.Errorf("")
		return err
	}

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
	tree, ok := sch.getTaskTree(task.FlowID)
	if !ok {
		logs.Errorf("execute next get task tree failed, flowID: %s, rid: %s", task.FlowID, kt.Rid)
		return fmt.Errorf("flow: %s not found", task.FlowID)
	}

	// 获取下次执行的任务
	ids := tree.Root.GetNextTaskNodes(task)
	if len(ids) == 0 {
		state := tree.Root.ComputeState()

		if state == enumor.FlowSuccess || state == enumor.FlowFailed {
			if err := updateFlowState(kt, sch.backend, task.FlowID, enumor.FlowRunning, state); err != nil {
				logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}

			sch.taskTrees.Delete(task.FlowID)
		}

		return nil
	}

	return sch.pushTasks(kt, tree.Flow, ids)
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
