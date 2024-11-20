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
	"hcm/pkg/async/action/run"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"
)

/*
WatchDog （看门狗）:
 1. 处理超时任务
 2. 处理处于Scheduled状态，但执行节点已经挂掉的任务流
 3. 处理处于Running状态，但执行节点正在Shutdown或者已经挂掉的任务流
*/
type WatchDog interface {
	compctrl.Closer
	// Start 启动watch dog，修复异常的异步任务流程。
	Start()
}

// watchDog 任务流、任务纠正策略
type watchDog struct {
	bd backend.Backend
	ld leader.Leader

	taskTimeoutSec      time.Duration
	shutdownWaitTimeSec time.Duration
	watchIntervalSec    time.Duration

	wg      sync.WaitGroup
	closeCh chan struct{}

	runningFlowMap map[string]time.Time
}

// NewWatchDog 创建一个watchdog
func NewWatchDog(bd backend.Backend, ld leader.Leader, opt *WatchDogOption) WatchDog {

	return &watchDog{
		bd:                  bd,
		ld:                  ld,
		taskTimeoutSec:      time.Duration(opt.TaskRunTimeoutSec) * time.Second,
		shutdownWaitTimeSec: time.Duration(opt.ShutdownWaitTimeSec) * time.Second,
		watchIntervalSec:    time.Duration(opt.WatchIntervalSec) * time.Second,
		wg:                  sync.WaitGroup{},
		closeCh:             make(chan struct{}),
		runningFlowMap:      make(map[string]time.Time),
	}
}

// Start 启动定义的WatchDog
func (wd *watchDog) Start() {
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleExpiredTasks)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleScheduledNotExistWorkerFlow)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleRunningNotExistWorkerFlow)
}

// 定期处理异常任务流或任务
func (wd *watchDog) watchWrapper(do func(kt *kit.Kit) error) {
	for {
		select {
		case <-wd.closeCh:
			break
		default:
		}

		kt := NewKit()
		if err := do(kt); err != nil {
			logs.Errorf("%s: watch dog do watch func failed, err: %v, rid: %s", constant.AsyncTaskWarnSign,
				err, kt.Rid)
		}
		time.Sleep(wd.watchIntervalSec)
	}

	wd.wg.Done()
}

// Close 等待当前执行体执行完成后再关闭
func (wd *watchDog) Close() {
	close(wd.closeCh)
	wd.wg.Wait()
}

// handleExpiredTasks 将超时任务和所属的任务流，设置为失败状态，失败原因：ErrTaskExecTimeout
func (wd *watchDog) handleExpiredTasks(kt *kit.Kit) error {

	input := &backend.ListInput{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.In.Factory(),
					Value: []enumor.TaskState{enumor.TaskRunning, enumor.TaskRollback},
				},
				&filter.AtomRule{
					Field: "updated_at",
					Op:    filter.LessThan.Factory(),
					Value: times.ConvStdTimeFormat(times.ConvStdTimeNow().Add(-wd.taskTimeoutSec)),
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: listExpiredTasksLimit,
		},
	}
	tasks, err := wd.bd.ListTask(kt, input)
	if err != nil {
		logs.Errorf("list expired tasks failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(tasks) == 0 {
		logs.V(3).Infof("handleExpiredTasks not found task, skip, rid: %s", kt.Rid)
		return nil
	}

	ids := make([]string, 0, len(tasks))
	for _, one := range tasks {
		// 检查任务的重试策略，是否已超时
		isExpired := wd.checkIsExpireTask(kt, one)
		if !isExpired {
			continue
		}

		ids = append(ids, one.ID)
		if err = wd.updateTimeoutTask(kt, one.ID); err != nil {
			return err
		}

		flows := []model.Flow{
			{
				ID:    one.FlowID,
				State: enumor.FlowFailed,
				Reason: &tableasync.Reason{
					Message: ErrTaskExecTimeout,
				},
			},
		}
		if err = wd.bd.BatchUpdateFlow(kt, flows); err != nil {
			logs.Errorf("update flow to failed state failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.V(5).Infof("handleExpiredTasks success, count: %d, ids: %v, rid: %s", len(ids), ids, kt.Rid)

	return nil
}

func (wd *watchDog) updateTimeoutTask(kt *kit.Kit, id string) error {
	task := &model.Task{
		ID:    id,
		State: enumor.TaskFailed,
		Reason: &tableasync.Reason{
			Message: ErrTaskExecTimeout,
		},
	}
	if err := wd.bd.UpdateTask(kt, task); err != nil {
		logs.Errorf("update task to failed state failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// handleScheduledNotExistWorkerFlow 将处于等待调度【Scheduled】且分配的节点已经下线的任务流重新设置为Pending.
func (wd *watchDog) handleScheduledNotExistWorkerFlow(kt *kit.Kit) error {

	flows, err := wd.queryNotExistNodesFlowByState(kt, enumor.FlowScheduled)
	if err != nil {
		return err
	}

	if len(flows) == 0 {
		return nil
	}

	mds := make([]model.Flow, 0, len(flows))
	ids := make([]string, 0, len(flows))
	for _, one := range flows {
		ids = append(ids, one.ID)
		mds = append(mds, model.Flow{
			ID:     one.ID,
			State:  enumor.FlowPending,
			Worker: converter.ValToPtr(""),
		})
	}
	if err = wd.bd.BatchUpdateFlow(kt, mds); err != nil {
		logs.Errorf("update flows failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	logs.Infof("handleScheduledNotExistWorkerFlow success, count: %d, ids: %v, rid: %s", len(ids), ids, kt.Rid)

	return nil
}

func (wd *watchDog) queryNotExistNodesFlowByState(kt *kit.Kit, state enumor.FlowState) ([]model.Flow, error) {
	nodes, err := wd.ld.AliveNodes()
	if err != nil {
		logs.Errorf("query alive nodes failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(nodes) == 0 {
		//  can not get node list sometimes, skip
		return nil, nil
	}

	input := &backend.ListInput{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.Equal.Factory(),
					Value: state,
				},
				&filter.AtomRule{
					Field: "worker",
					Op:    filter.NotIn.Factory(),
					Value: nodes,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	flows, err := wd.bd.ListFlow(kt, input)
	if err != nil {
		logs.Errorf("list flows failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return flows, nil
}

// handleRunningNotExistWorkerFlow 处理处于Running状态且处理Worker已经下线的Flow。
func (wd *watchDog) handleRunningNotExistWorkerFlow(kt *kit.Kit) error {

	flows, err := wd.queryNotExistNodesFlowByState(kt, enumor.FlowRunning)
	if err != nil {
		return err
	}

	if len(flows) == 0 {
		logs.V(3).Infof("handleRunningNotExistWorkerFlow not found flow, skip, rid: %s", kt.Rid)
		return nil
	}

	ids := make([]string, 0, len(flows))
	for _, flow := range flows {
		// 捕获到Running的节点有可能正在还未彻底结束的节点上运行，所以，
		// 需要等待上一个节点Shutdown结束后再处理，否则会有两个节点处理同一个Flow。
		firstWatchTime, exist := wd.runningFlowMap[flow.ID]
		if !exist {
			wd.runningFlowMap[flow.ID] = times.ConvStdTimeNow()
			continue
		}

		if !firstWatchTime.Before(times.ConvStdTimeNow().Add(-wd.shutdownWaitTimeSec)) {
			continue
		}

		// 如果上一个节点已经Shutdown，表示可以处理Running的Flow
		if err = wd.handleRunningFlow(kt, flow); err != nil {
			logs.Errorf("handle running flow in not exist worker failed, id: %s, rid: %s", flow.ID, kt.Rid)
			return err
		}

		ids = append(ids, flow.ID)
		delete(wd.runningFlowMap, flow.ID)
	}

	if len(ids) != 0 {
		logs.Infof("handleRunningNotExistWorkerFlow, count: %d, ids: %v, rid: %s", len(ids), ids, kt.Rid)
	}

	return nil
}

func (wd *watchDog) handleRunningFlow(kt *kit.Kit, flow model.Flow) error {
	// 根据任务流ID获取对应的任务集合
	taskModels, err := listTaskByFlowID(kt, wd.bd, flow.ID)
	if err != nil {
		return err
	}

	// 构造执行流树
	root, err := BuildTaskRoot(taskModels)
	if err != nil {
		return err
	}

	// 如果树已经处于结束状态，则直接更新
	state := root.TreeState()
	if state == enumor.FlowSuccess || state == enumor.FlowFailed {
		if err = updateFlowState(kt, wd.bd, flow.ID, enumor.FlowRunning, state); err != nil {
			logs.Errorf("update flow state to %s failed, err: %v, rid: %s", state, err, kt.Rid)
			return err
		}

		return nil
	}

	ids := root.GetExecStateTasks()
	// 如果没有处于执行中的节点，将Flow置于Pending状态，等待重新被调度
	if len(ids) == 0 {
		mds := []model.Flow{
			{
				ID:     flow.ID,
				State:  enumor.FlowPending,
				Worker: converter.ValToPtr(""),
			},
		}
		if err = wd.bd.BatchUpdateFlow(kt, mds); err != nil {
			logs.Errorf("update flows failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		return nil
	}

	// 否则，找出所有处于执行状态的节点，判断它的执行节点是否已经退出，如果退出将Task回滚或者置于失败状态。
	if err = wd.handleRunningTasks(kt, flow, ids); err != nil {
		return err
	}

	return nil
}

// handleRunningTasks 找出所有处于执行状态的节点，判断它的执行节点是否已经退出，如果退出将Task回滚或者置于失败状态。
func (wd *watchDog) handleRunningTasks(kt *kit.Kit, flow model.Flow, ids []string) error {
	tasks, err := listTaskByIDs(kt, wd.bd, ids)
	if err != nil {
		logs.Errorf("list task by ids failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return err
	}

	for _, task := range tasks {
		// 检查任务的重试策略，是否已超时
		isExpired := wd.checkIsExpireTask(kt, task.Task)
		if isExpired {
			// 如果任务已经超时，更新为失败状态，失败原因超时
			if err = wd.updateTimeoutTask(kt, task.ID); err != nil {
				return err
			}
			continue
		}

		// 如果任务不能重试，将任务置于失败状态
		if !task.Retry.IsEnable() {
			md := &model.Task{
				ID:    task.ID,
				State: enumor.TaskFailed,
				Reason: &tableasync.Reason{
					Message: ErrTaskNodeShutdown,
				},
			}
			if err = wd.bd.UpdateTask(kt, md); err != nil {
				logs.Errorf("update task to failed state failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}

		taskExecKit := run.NewExecuteContext(task.Kit, flow.ShareData)
		task.InitDep(taskExecKit, func(taskKit *kit.Kit, task *model.Task) error {
			return wd.bd.UpdateTask(kt, task)
		}, &Flow{Flow: flow})

		// 如果任务可以重试，更新任务状态为RollBack ，等待回滚
		if err = task.Rollback(); err != nil {
			logs.Errorf("rollback not exist node task failed, err: %v, rid: %s", err, kt.Rid)

			md := &model.Task{
				ID:    task.ID,
				State: enumor.TaskFailed,
				Reason: &tableasync.Reason{
					Message: fmt.Sprintf("rollback not exist node task failed, err: %v", err),
				},
			}
			if patchErr := wd.bd.UpdateTask(kt, md); patchErr != nil {
				logs.Errorf("update task to failed state failed, err: %v, rid: %s", patchErr, kt.Rid)
				return err
			}
		}
	}
	return nil
}

// checkIsExpireTask 检查任务是否超时
func (wd *watchDog) checkIsExpireTask(kt *kit.Kit, task model.Task) bool {
	if task.Retry == nil || !task.Retry.IsEnable() || task.Retry.Policy == nil {
		return task.UpdatedAt < times.ConvStdTimeFormat(times.ConvStdTimeNow().Add(-wd.taskTimeoutSec))
	}

	// 检查任务的重试策略，是否已超时
	retryMillSec := task.Retry.Policy.Count * task.Retry.Policy.SleepRangeMS[0]
	updateDate, err := time.Parse(constant.TimeStdFormat, task.UpdatedAt)
	if err == nil {
		expireTime := updateDate.Add(time.Duration(retryMillSec) * time.Millisecond)
		if expireTime.UnixNano() > time.Now().UnixMicro() {
			logs.V(5).Infof(
				"check task is not expired, taskID: %s, flowID: %s, updateAt: %s, expireTime: %s, rid: %s",
				task.ID, task.FlowID, task.UpdatedAt, expireTime.Format(constant.DateTimeLayout), kt.Rid)
			return false
		}
	}

	return true
}
