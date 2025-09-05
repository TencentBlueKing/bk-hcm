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
	"sort"
	"strconv"
	"sync"
	"time"

	adaptormock "hcm/pkg/adaptor/mock"
	"hcm/pkg/api/core"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

// FlowTypeMinPriority flowtype的最小优先级（值越大优先级越低）
const FlowTypeMinPriority = 10
const DefaultFlowTypePriority = FlowTypeMinPriority / 2

/*
Scheduler （调度器）: TODO: 换为 捕获器、消费器，添加假死任务销毁逻辑

	1.获取分配给当前节点的处于Scheduled状态的任务流，构建任务流树，将待执行任务推送到执行器执行。
	2.分析执行器执行完的任务，判断任务流树状态，如果任务流处理完，更新状态，否则将子节点推送到执行器执行。
*/
type Scheduler interface {
	compctrl.Closer

	// Start 启动调度器。
	Start()

	// EntryTask 分析执行完的任务，并解析出当前任务的子任务去执行。
	EntryTask(task *Task) error
	// DeleteFlowTaskTree 清空任务树，阻止继续调度
	DeleteFlowTaskTree(flowID string)
	// SetFlowTypePriority 设置flowtype优先级
	SetFlowTypePriority(flowType enumor.FlowName, priority int)
}

// scheduler 定义任务流调度器
type scheduler struct {
	workerNumber uint

	taskTrees   sync.Map
	workerQueue *concurrence.UnboundedBlockingLinkedList[*Task]
	workerWg    sync.WaitGroup

	backend  backend.Backend
	executor Executor
	leader   leader.Leader

	sp      SleepPolicy
	closeCh chan struct{}

	scheduledFlowFetcherConcurrency uint
	canceledFlowFetcherConcurrency  uint
	// flowtype理论执行时间（关键路径），单位秒
	flowTypeExecTimeMap   *adaptormock.Store[enumor.FlowName, float64]
	flowTypeRunningNumMap *concurrence.ConcurrentMapCounter
	flowTypePriorityMap   sync.Map
	flowTypeMinPriority   int
	// flowtype实际执行时间，单位秒
	flowtypeActualTime sync.Map
	// 记录每个flow从被选中开始的时间
	flowEntryTimeMap sync.Map
	mc               *metric
}

// NewScheduler 实例化任务流调度器
func NewScheduler(bd backend.Backend, exec Executor, ld leader.Leader, opt *SchedulerOption, globalCfgCli *global.GlobalConfigsClient, mc *metric) Scheduler {
	sch := &scheduler{
		closeCh:                         make(chan struct{}),
		workerWg:                        sync.WaitGroup{},
		workerQueue:                     concurrence.NewUnboundedBlockingLinkedList[*Task](),
		workerNumber:                    opt.WorkerNumber,
		sp:                              SleepPolicy{baseInterval: time.Duration(opt.WatchIntervalSec) * time.Second},
		backend:                         bd,
		executor:                        exec,
		leader:                          ld,
		scheduledFlowFetcherConcurrency: opt.ScheduledFlowFetcherConcurrency,
		canceledFlowFetcherConcurrency:  opt.CanceledFlowFetcherConcurrency,
		flowTypeExecTimeMap:             &adaptormock.Store[enumor.FlowName, float64]{},
		flowTypeRunningNumMap:           concurrence.NewConcurrentMapCounter(),
		flowTypePriorityMap:             sync.Map{},
		flowTypeMinPriority:             FlowTypeMinPriority,
		flowtypeActualTime:              sync.Map{},
		flowEntryTimeMap:                sync.Map{},
		mc:                              mc,
	}
	sch.flowTypeExecTimeMap.Init()

	kt := NewKit()
	req := &datagconf.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.FlowTypePriority),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := globalCfgCli.List(kt, req)
	if err != nil {
		logs.Errorf("failed to get flow type priority from global config, err: %v", err)
		return sch
	}
	for _, one := range result.Details {
		priority, err := strconv.Atoi(string(one.ConfigValue))
		if err != nil {
			logs.Errorf("failed to parse flow type priority, err: %v, config_value: %v", err, one.ConfigValue)
			return sch
		}
		sch.flowTypePriorityMap.Store(one.ConfigKey, priority)
	}
	return sch
}

// Start 初始化调度器并启动执行
func (sch *scheduler) Start() {

	logs.Infof("scheduler start, worker number: %d, default loop interval: %s", sch.workerNumber,
		sch.sp.baseInterval.String())

	// 定期获取等待执行的任务流
	go sch.scheduledFlowWatcher()
	go sch.canceledFlowWatcher()

	// 启动workerNumber个协程进行后续可执行任务流的解析（第一批可执行节点之后的）
	for i := 0; i < int(sch.workerNumber); i++ {
		go sch.goWorker()
	}
}

// flowWatcher 定期查询调度到该节点的flow
func (sch *scheduler) scheduledFlowWatcher() {
	// 初始化协程池
	pool := newTenantWorkerPool(sch.scheduledFlowFetcherConcurrency,
		func(tenantID string) {
			kt := NewKit()
			kt.TenantID = tenantID
			working, err := sch.runScheduledFlow(kt)
			if err != nil {
				logs.Errorf("%s: scheduler watch scheduled flow failed for tenant %s, err: %v, rid: %s",
					constant.AsyncTaskWarnSign, tenantID, err, kt.Rid)
				sch.sp.ExceptionSleep()
				return
			}
			if working {
				sch.sp.ShortSleep()
			}
		})

	// 主任务分发循环
	for {
		select {
		case <-sch.closeCh:
			pool.shutdownPoolGracefully()
			sch.workerWg.Done()
			logs.Infof("received stop signal, stop watch scheduled flow job success.")
			return
		default:
		}

		err := pool.executeWithTenant()
		if err != nil {
			logs.Errorf("scheduledFlowWatcher failed to executeWithTenant, err: %v", err)
			sch.sp.ExceptionSleep()
			continue
		}

		sch.sp.NormalSleep()
	}
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

// queryCurrNodeAllFlow 查询主节点分配给当前节点所有处于 Scheduled 状态的任务流，没有limit上限。
func (sch *scheduler) queryCurrNodeAllFlow(kt *kit.Kit, state enumor.FlowState) ([]model.Flow, error) {
	flows := make([]model.Flow, 0)
	page := core.NewDefaultBasePage()
	for {
		result, err := sch.backend.ListFlow(kt, &backend.ListInput{
			Page: page,
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("state", state),
				tools.RuleEqual("worker", sch.leader.CurrNode())),
		})
		if err != nil {
			logs.Errorf("list flows failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		flows = append(flows, result...)
		// 如果当前页数据不足一页，说明后面没有更多数据了
		if uint(len(result)) < page.Limit {
			break
		}
		page.Start += uint32(page.Limit)
	}
	return flows, nil
}

// canceledFlowWatcher 查询当前节点上被取消的flow并执行task取消操作
func (sch *scheduler) canceledFlowWatcher() {
	// 初始化协程池
	pool := newTenantWorkerPool(sch.canceledFlowFetcherConcurrency,
		func(tenantID string) {
			kt := NewKit()
			kt.TenantID = tenantID
			working, err := sch.handleCanceledFlow(kt)
			if err != nil {
				logs.Errorf("%s: scheduler watch canceled failed for tenant %s, err: %v, rid: %s",
					constant.AsyncTaskWarnSign, tenantID, err, kt.Rid)
				sch.sp.ExceptionSleep()
				return
			}
			if working {
				sch.sp.ShortSleep()
			}
		})

	for {
		select {
		case <-sch.closeCh:
			pool.shutdownPoolGracefully()
			sch.workerWg.Done()
			logs.Infof("received stop signal, stop watch canceled flow job success.")
			return
		default:
		}

		err := pool.executeWithTenant()
		if err != nil {
			logs.Errorf("canceledFlowWatcher failed to executeWithTenant, err: %v", err)
			sch.sp.NormalSleep()
			continue
		}

		sch.sp.NormalSleep()
	}
}

func (sch *scheduler) handleCanceledFlow(kt *kit.Kit) (working bool, err error) {

	dbFlows, err := sch.queryCurrNodeFlow(kt, enumor.FlowCancel, listScheduledFlowLimit)
	if err != nil {
		logs.Errorf("fail to list canceled flow, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	if len(dbFlows) == 0 {
		return false, nil
	}

	for _, flow := range dbFlows {
		logs.Infof("canceling flow: %s, rid: %s", flow.ID, kt.Rid)

		// 清空任务树，阻止继续调度
		sch.DeleteFlowTaskTree(flow.ID)
		sch.flowTypeRunningNumMap.Inc(string(flow.Name), -1)
		sch.flowEntryTimeMap.Delete(flow.ID)
		err = updateFlowToCancel(kt, sch.backend, flow.ID, cvt.PtrToVal(flow.Worker), enumor.FlowCancel)
		if err != nil {
			logs.Errorf("fail to update flow clear worker id, err: %v, flow id: %s rid: %s",
				err, flow.ID, kt.Rid)
			// keep canceling other flow
			continue
		}

		err := sch.executor.CancelFlow(kt, flow.ID)
		if err != nil {
			logs.Errorf("fail to handle flow canceling, err: %v, flow id: %s, rid: %s", err, flow.ID, kt.Rid)
			// keep canceling other flow
			continue
		}

		logs.Infof("cancel flow: %s success, rid: %s", flow.ID, kt.Rid)
	}

	return true, nil
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

// parseAndPushFlow 解析Flow并返回Flow第一批可执行的节点。
func (sch *scheduler) parseFlow(kt *kit.Kit, flow *Flow) ([]*Task, error) {
	// 根据任务流ID获取对应的任务集合
	tasks, err := listTaskByFlowID(kt, sch.backend, flow.ID)
	if err != nil {
		if stateErr := updateFlowStateAndReason(kt, sch.backend, flow.ID, enumor.FlowRunning, enumor.FlowFailed,
			err.Error()); stateErr != nil {

			logs.Errorf("update flow state and reason failed, after list task by flow id failed, err: %v, rid: %s",
				stateErr, kt.Rid)
		}

		logs.Errorf("list task by flow id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
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
		return nil, err
	}

	taskTree := &TaskTree{
		Flow: flow,
		Root: root,
	}

	// 获取可执行的节点
	executableTaskNodes := taskTree.Root.GetExecutableTasks()
	if len(executableTaskNodes) == 0 {
		state := taskTree.Root.TreeState()

		if state == enumor.FlowSuccess {
			if err = updateFlowState(kt, sch.backend, flow.ID, enumor.FlowRunning, state); err != nil {
				logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
				return nil, err
			}
		}

		if state == enumor.FlowFailed {
			if err = updateFlowStateAndReason(kt, sch.backend, flow.ID, enumor.FlowRunning, state,
				ErrSomeTaskExecFailed); err != nil {

				logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
				return nil, err
			}
		}

		return nil, nil
	}

	// 存储任务流执行树
	sch.taskTrees.Store(flow.ID, taskTree)
	flow.State = enumor.FlowRunning
	executableSet := make(map[string]struct{}, len(executableTaskNodes))
	for _, taskID := range executableTaskNodes {
		executableSet[taskID] = struct{}{}
	}

	// 选出可执行的节点的task结构体并填充预估的执行时间
	execTasks := make([]*Task, 0, len(executableTaskNodes))
	for _, task := range tasks {
		task.Flow = flow
		if _, exists := executableSet[task.ID]; exists {
			avgExecTime, neverExec := sch.executor.GetTaskTypeAvgExecTime(task.ActionName)
			task.ExecTime = avgExecTime
			// 如果该任务类型从未执行过，则当做慢任务处理，避免如果真的是慢任务会一直占着快任务worker
			if neverExec {
				task.ExecTime = sch.executor.GetFastTaskThresholdSec()
			}
			execTasks = append(execTasks, task)
		}
	}
	return execTasks, nil
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
func (sch *scheduler) runScheduledFlow(kt *kit.Kit) (working bool, err error) {
	for flowType, count := range sch.flowTypeRunningNumMap.Snapshot() {
		sch.mc.flowTypeRunningNum.WithLabelValues(flowType).Set(float64(count))
	}

	// 从DB中获取所有"待执行"的任务流并更新状态为"执行中"，开始对flow中的任务进行执行
	allFlows, err := sch.queryCurrNodeAllFlow(kt, enumor.FlowScheduled)
	if err != nil {
		logs.Errorf("list flows failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	if len(allFlows) == 0 {
		logs.V(3).Infof("current node: %s not found scheduled flow to handleRunningFlow, rid: %s",
			sch.leader.CurrNode(), kt.Rid)
		return false, nil
	}

	topkFlows, _ := sch.getTopKFlows(allFlows, 0.75*listScheduledFlowLimit, listScheduledFlowLimit)

	ids := make([]string, 0)
	flows := slice.Map(topkFlows, func(one model.Flow) *Flow {
		ids = append(ids, one.ID)
		// Note: first sub kit, scheduler.watcher -> flow
		return &Flow{Flow: one, Kit: kt.NewSubKit()}
	})

	logs.V(3).Infof("list %d flows, try to run them, ids: %v, rid: %s", len(topkFlows), ids, kt.Rid)

	allTasks := make([]*Task, 0)
	for _, flow := range flows {
		if err = updateFlowState(flow.Kit, sch.backend, flow.ID, enumor.FlowScheduled, enumor.FlowRunning); err != nil {
			logs.Errorf("update flow state failed, err: %v, rid: %s", err, flow.Kit.Rid)
			return false, err
		}

		tasks, err := sch.parseFlow(flow.Kit, flow)
		if err != nil {
			logs.Errorf("parse flow and push task failed, err: %v, rid: %s", err, flow.Kit.Rid)
			return false, err
		}
		allTasks = append(allTasks, tasks...)
		// 更新flow type的running数量
		sch.flowTypeRunningNumMap.Inc(string(flow.Name), 1)
		// 更新flow的entry time
		sch.flowEntryTimeMap.Store(flow.ID, time.Now())
	}

	for _, task := range allTasks {
		sch.executor.Push(task.Flow, task)
	}

	logs.V(3).Infof("update all the flows to running state success. rid: %s", kt.Rid)

	return true, nil
}

// getTopKFlows eachFlowTypeNum是每种flow最多可以取的数量，k是最终返回的flow总数上限
// 如果每种flow各取eachFlowTypeNum个后加起来仍不足k个，则按照先来后到补足
// 比如5个A类flow紧接着2个B类flow，k=5，eachFlowTypeNum=2，则返回A类3个，B类2个
func (sch *scheduler) getTopKFlows(allFlows []model.Flow, eachFlowTypeNum int, k int) ([]model.Flow, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	// 如果k大于等于所有flows数量，直接返回所有flows
	if k >= len(allFlows) {
		result := make([]model.Flow, len(allFlows))
		copy(result, allFlows)
		return result, nil
	}

	// 使用map记录每种flow已收集的数量
	counts := make(map[enumor.FlowName]int)
	selected := make(map[string]bool, len(allFlows))
	flows := make([]model.Flow, 0, min(len(allFlows), k))

	// 每种flow各取eachFlowTypeNum个
	for _, flow := range allFlows {
		if counts[flow.Name] < eachFlowTypeNum {
			flows = append(flows, flow)
			counts[flow.Name]++
			selected[flow.ID] = true
		}
	}

	// 如果每种flow各取eachFlowTypeNum个后加起来仍不足k个，则按照先来后到补足，allFlows中靠前的flow就是先到的
	// 即某种flow其实有可能超过eachFlowTypeNum个
	if len(flows) < k {
		for _, flow := range allFlows {
			if !selected[flow.ID] && len(flows) < k {
				flows = append(flows, flow)
				selected[flow.ID] = true
			}
		}
		return flows, nil
	}
	result, err := sch.rankTopKFlows(flows, k)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (sch *scheduler) rankTopKFlows(flows []model.Flow, k int) ([]model.Flow, error) {
	// 创建带分数的结构体用于排序
	type scoredFlow struct {
		flow  model.Flow
		score float64
	}

	scoredFlows := make([]scoredFlow, 0, len(flows))

	// 计算每个flow的分数
	for _, flow := range flows {
		scoredFlows = append(scoredFlows, scoredFlow{
			flow:  flow,
			score: sch.caculateFlowTypeScore(flow),
		})
	}

	// 按分数降序排序
	sort.Slice(scoredFlows, func(i, j int) bool {
		return scoredFlows[i].score > scoredFlows[j].score
	})

	// 取前k个
	result := make([]model.Flow, 0, min(len(scoredFlows), k))
	for i := 0; i < len(scoredFlows) && i < k; i++ {
		result = append(result, scoredFlows[i].flow)
	}

	return result, nil
}

// caculateFlowTypeScore 计算flow的分数，等待时间越长、执行时间越短、框架内同类flow执行数量越少、优先级越高，则分数越高
func (sch *scheduler) caculateFlowTypeScore(flow model.Flow) float64 {
	priority, _ := sch.flowTypePriorityMap.LoadOrStore(flow.Name, DefaultFlowTypePriority)
	execTime, _ := sch.flowTypeExecTimeMap.Get(flow.Name)
	runningNum, runningNumTotal := sch.flowTypeRunningNumMap.GetValueAndSum(string(flow.Name))
	// UpdatedAt是flow变成scheduled态那一刻，
	updatedTime, err := time.Parse(time.RFC3339, flow.UpdatedAt)
	if err != nil {
		logs.Infof("parse time %q failed: %v", flow.UpdatedAt, err)
	}
	// now-updatedTime=等待时间，单位秒
	waitTime := time.Now().Unix() - updatedTime.Unix()

	norPriority := 1 - float64(priority.(int))/float64(sch.flowTypeMinPriority)
	norExecTime := 1 / (1 + execTime)
	norRunningNum := 1 - float64(runningNum)/float64(runningNumTotal)
	norWaitTime := float64(waitTime) / (float64(waitTime) + 1)

	score := 2*norPriority + norExecTime + norRunningNum + norWaitTime
	return score
}

// 任务流解析协程
func (sch *scheduler) goWorker() {
	sch.workerWg.Add(1)

	for {
		task, ok := sch.workerQueue.Pop()
		if !ok {
			break
		}

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
		state := tree.Root.TreeState()
		switch state {
		case enumor.FlowSuccess:
			if err := updateFlowState(kt, sch.backend, task.FlowID, enumor.FlowRunning, state); err != nil {
				logs.Errorf("update flow state to `%s` failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}
			// 只对执行成功的flow求关键路径执行时间以及实际执行时间
			criticalPathExecTime, neverExec := sch.computeFlowTypeCriticalPath(task.Flow, tree)
			if !neverExec {
				sch.flowTypeExecTimeMap.Set(task.Flow.Name, criticalPathExecTime)
			}
			entryTime, exists := sch.flowEntryTimeMap.Load(task.FlowID)
			if exists {
				t, ok := entryTime.(time.Time)
				if !ok {
					logs.Errorf("entry time is not time, flowID: %s, rid: %s", task.FlowID, kt.Rid)
					return fmt.Errorf("entry time is not time, flowID: %s", task.FlowID)
				}
				sch.mc.flowTypeExecTime.WithLabelValues(string(task.Flow.Name)).Observe(time.Since(t).
					Seconds())
			}

			sch.DeleteFlowTaskTree(task.FlowID)
			sch.flowTypeRunningNumMap.Inc(string(task.Flow.Name), -1)
			sch.flowEntryTimeMap.Delete(task.FlowID)
		case enumor.FlowFailed:
			if err := updateFlowStateAndReason(kt, sch.backend, task.FlowID, enumor.FlowRunning, state,
				ErrSomeTaskExecFailed); err != nil {

				logs.Errorf("update flow state to `%s` failed, err: %v, rid: %s", state, err, kt.Rid)
				return err
			}

			sch.DeleteFlowTaskTree(task.FlowID)
			sch.flowTypeRunningNumMap.Inc(string(task.Flow.Name), -1)
			sch.flowEntryTimeMap.Delete(task.FlowID)
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

	for _, task := range tasks {
		avgExecTime, neverExec := sch.executor.GetTaskTypeAvgExecTime(task.ActionName)
		task.ExecTime = avgExecTime
		if neverExec {
			task.ExecTime = sch.executor.GetFastTaskThresholdSec()
		}
	}
	// 根据任务的ExecTime属性进行排序，快任务在前面
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ExecTime < tasks[j].ExecTime
	})

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
func (sch *scheduler) EntryTask(taskNode *Task) error {
	err := sch.workerQueue.Push(taskNode)
	if err != nil {
		return err
	}
	return nil
}

// 计算某个flowtype的关键路径执行时间。neverExec为true表示该flowtype在整个服务生命周期内从未被执行，无法计算其关键路径执行时间
func (sch *scheduler) computeFlowTypeCriticalPath(flow *Flow, taskTree *TaskTree) (criticalPathExecTime float64, neverExec bool) {
	taskTypeAvgExecTimeMap := make(map[enumor.ActionName]float64)
	// 遍历flow中的所有任务，获取每个任务类型的平均执行时间
	for _, t := range flow.Tasks {
		_, exists := taskTypeAvgExecTimeMap[t.ActionName]
		if exists {
			continue
		}
		avgExecTime, neverExec := sch.executor.GetTaskTypeAvgExecTime(t.ActionName)
		if neverExec {
			return 0, true
		}
		taskTypeAvgExecTimeMap[t.ActionName] = avgExecTime
	}

	// 直接获取最大执行时间
	maxTime := taskTree.Root.GetAllPathsMaxTime(taskTypeAvgExecTimeMap)

	return maxTime, false
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
	sch.workerQueue.Close()

	sch.workerWg.Wait()

	logs.Infof("scheduler receive close cmd, close success!")

}

// DeleteFlowTaskTree 清空任务树，阻止继续调度
func (sch *scheduler) DeleteFlowTaskTree(flowID string) {

	sch.taskTrees.Delete(flowID)
}

// SetFlowTypePriorityMap 设置flowtype的优先级map
func (sch *scheduler) SetFlowTypePriority(flowType enumor.FlowName, priority int) {
	sch.flowTypePriorityMap.Store(flowType, priority)
}
