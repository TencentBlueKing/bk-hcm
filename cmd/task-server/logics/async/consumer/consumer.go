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
	"sync"
	"time"

	"hcm/cmd/task-server/logics/async/backends/iface"
	"hcm/cmd/task-server/logics/async/flow"
	"hcm/cmd/task-server/logics/async/task"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/retry"
)

// Consumer ...
type Consumer struct {
	sd      serviced.ServiceDiscover
	backend iface.Backend
}

// NewConsumer new consumer
func NewConsumer(sd serviced.ServiceDiscover, backend iface.Backend) *Consumer {
	return &Consumer{
		sd:      sd,
		backend: backend,
	}
}

// RunConsumer run consumer
func (c *Consumer) RunConsumer() {

	limiter := make(chan struct{}, cc.TaskServer().Async.ConsumerNum)

	for {
		time.Sleep(time.Duration(cc.TaskServer().Async.ConsumerIntervalSecond) * time.Second)

		// etcd判主, 非主不执行
		if !c.sd.IsMaster() {
			continue
		}

		limiter <- struct{}{}
		kt := kit.New()
		kt.User = constant.AsyncUserKey
		kt.AppCode = constant.AsyncAppCodeKey
		go c.consumer(kt, limiter)
	}
}

func (c *Consumer) consumer(kt *kit.Kit, limiter chan struct{}) error {
	defer func(rid string) {
		logs.V(3).Infof("[async] consumer flow end with rid %s", rid)
		<-limiter
	}(kt.Rid)

	logs.V(3).Infof("[async] consumer flow start with rid %s", kt.Rid)

	// 从DB中获取1条待执行的任务流并更新状态为执行中
	flowFromDB, err := c.backend.ConsumeOnePendingFlow()
	if err != nil {
		return err
	}
	flowID := flowFromDB.ID

	// 根据任务流ID获取对应的任务集合
	taskResult, err := c.backend.GetTasksByFlowID(flowID)
	if err != nil {
		if err := c.ConsumerChangeFlowState(flowID, enumor.FlowFailed, err.Error()); err != nil {
			logs.Errorf("[async] change flow state err: %v, with rid %s", err, kt.Rid)
			return err
		}
		logs.Errorf("[async] get tasks by flow id err: %v, with rid %s", err, kt.Rid)
		return err
	}

	// 任务集合转换
	tasks := task.ConvTaskResultToTask(taskResult)

	// 构造执行流树
	root, err := flow.BuildTaskRoot(tasks)
	if err != nil {
		if err := c.ConsumerChangeFlowState(flowID, enumor.FlowFailed, err.Error()); err != nil {
			logs.Errorf("[async] change flow state err: %v, with rid %s", err, kt.Rid)
			return err
		}
		logs.Errorf("[async] build task root err: %v, flow id is %s with rid %s", err, flowID, kt.Rid)
		return err
	}

	taskTree := flow.NewTaskTree()
	taskTree.Root = root
	taskTree.Reason = constant.AsyncDefaultJson
	taskTree.FlowState = enumor.FlowPending
	taskTree.RunTaskNodes[root.Task.ID] = root
	for {
		// 检查任务流的状态
		taskTree.FlowState = root.ComputeStatus()
		logs.V(3).Infof("[async] consumer flow state is %s with rid %s", taskTree.FlowState, kt.Rid)

		if taskTree.FlowState == enumor.FlowFailed ||
			taskTree.FlowState == enumor.FlowSuccess ||
			taskTree.Reason != constant.AsyncDefaultJson {
			break
		}

		// 执行action
		logs.V(3).Infof("[async] consumer flow run tasks %v with rid %s", c.getRunningTasks(taskTree.RunTaskNodes), kt.Rid)
		nextNodes := make([]*flow.TaskNode, 0)
		wg := &sync.WaitGroup{}
		for _, taskNode := range taskTree.RunTaskNodes {
			wg.Add(1)
			go func(taskNode *flow.TaskNode, wg *sync.WaitGroup) {
				defer wg.Done()
				if err := taskNode.Task.DoTask(kt, c.backend); err != nil {
					logs.Errorf("[async] do task err: %v, flow id is %s with rid %s", err, flowID, kt.Rid)
					taskTree.Reason = err.Error()
					taskNode.Task.State = enumor.TaskFailed
					taskTree.FlowState = enumor.FlowFailed
					return
				}

				nodes, find := taskNode.GetNextTaskNodes()
				if !find {
					logs.Errorf("[async] can not find next task nodes, flow id is %s with rid %s", flowID, kt.Rid)
					taskTree.Reason = err.Error()
					taskNode.Task.State = enumor.TaskFailed
					taskTree.FlowState = enumor.FlowFailed
					return
				}

				taskTree.RunTaskNodesLock.Lock()
				delete(taskTree.RunTaskNodes, taskNode.Task.ID)
				nextNodes = append(nextNodes, nodes...)
				defer taskTree.RunTaskNodesLock.Unlock()
			}(taskNode, wg)
		}
		wg.Wait()
		// 设置下次执行的节点
		for _, node := range nextNodes {
			taskTree.RunTaskNodes[node.Task.ID] = node
		}
	}

	if err := c.ConsumerChangeFlowState(flowID, taskTree.FlowState, taskTree.Reason); err != nil {
		logs.Errorf("[async] change flow state err: %v, with rid %s", err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] consumer flow run %s with rid %s", taskTree.FlowState, kt.Rid)

	return nil
}

// ConsumerChangeFlowState consumer change flow state
func (c *Consumer) ConsumerChangeFlowState(flowID string, state enumor.FlowState,
	reason string) error {
	// 获取任务流
	flow, err := c.backend.GetFlowByID(flowID)
	if err != nil {
		return err
	}

	// 状态校验
	if err := state.ValidateBeforeState(flow.State); err != nil {
		return err
	}

	// 更新任务流状态操作重试
	maxRetryCount := uint32(3)
	r := retry.NewRetryPolicy(uint(maxRetryCount), [2]uint{1000, 15000})
	var lastError error
	lastError = nil
	for r.RetryCount() < maxRetryCount {
		if err := c.backend.SetFlowStateWithReason(flowID, state, reason); err != nil {
			lastError = err
			r.Sleep()
		}

		if lastError == nil {
			break
		}
	}

	return lastError
}

func (c *Consumer) getRunningTasks(m map[string]*flow.TaskNode) []string {
	ret := make([]string, 0, len(m))

	for _, one := range m {
		ret = append(ret, one.Task.ID+"|"+one.Task.ActionName)
	}

	return ret
}
