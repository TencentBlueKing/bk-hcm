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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"hcm/cmd/task-server/logics/async/backends/iface"
	"hcm/cmd/task-server/logics/async/flow"
	"hcm/cmd/task-server/logics/async/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
)

// AsyncParser 定义任务流解析器
type AsyncParser struct {
	taskTrees         sync.Map
	workerNumber      int
	normalIntervalSec time.Duration
	workerQueue       chan *flow.TaskNode
	workerWg          sync.WaitGroup

	closeCh chan struct{}
}

// NewParser 实例化任务流解析器
func NewParser(workerNumber int, normalIntervalSec time.Duration) *AsyncParser {
	return &AsyncParser{
		closeCh:           make(chan struct{}),
		workerWg:          sync.WaitGroup{},
		workerQueue:       make(chan *flow.TaskNode, 10),
		workerNumber:      workerNumber,
		normalIntervalSec: normalIntervalSec,
	}
}

// SetWorkerNumber 设置启动的解析器数量
func (ap *AsyncParser) SetWorkerNumber(workerNumber int) {
	ap.workerNumber = workerNumber
}

// SetNormalIntervalSec 设置通用间隔时间
func (ap *AsyncParser) SetNormalIntervalSec(normalIntervalSec time.Duration) {
	ap.normalIntervalSec = normalIntervalSec
}

// Init 初始化解析器并启动执行
func (ap *AsyncParser) Init() {
	// 定期获取等待执行的任务流
	ap.workerWg.Add(1)
	go ap.startWatcher(ap.watchPendingFlow)

	// 启动workerNumber个协程进行任务流解析
	for i := 0; i < ap.workerNumber; i++ {
		ap.workerWg.Add(1)
		go ap.goWorker()
	}
}

// 定期执行do函数体
func (ap *AsyncParser) startWatcher(do func() error) {
	ticker := time.NewTicker(ap.normalIntervalSec)
	defer ticker.Stop()

	closed := false
	for !closed {
		select {
		case <-ap.closeCh:
			closed = true
		case <-ticker.C:
			if err := do(); err != nil {
				logs.Errorf("[async] [module-parser] do watch func error %v", err)
			}
		}
	}

	ap.workerWg.Done()
}

func (ap *AsyncParser) watchPendingFlow() error {
	// 从DB中获取一条待执行的任务流并更新状态为执行中
	flowFromDB, err := iface.GetBackend().ConsumeOnePendingFlow()
	if err != nil {
		if strings.Contains(err.Error(), "flow num is 0") {
			return nil
		}
		logs.Errorf("[async] [module-parser] get flow error %v", err)
		return err
	}
	flowID := flowFromDB.ID

	// 根据任务流ID获取对应的任务集合
	taskResult, err := iface.GetBackend().GetTasksByFlowID(flowID)
	if err != nil {
		if err := ap.changeFlowState(flowID, &iface.FlowChange{
			State:  enumor.FlowFailed,
			Reason: err.Error(),
		}); err != nil {
			logs.Errorf("[async] [module-parser] change flow state error %v", err)
			return err
		}
		logs.Errorf("[async] [module-parser] get tasks by flowid error %v", err)
		return err
	}

	// 任务集合转换
	tasks := task.ConvTaskResultToTask(taskResult)
	if len(tasks) != len(taskResult) {
		return errors.New("conv taskResult to tasks num neq")
	}

	// 构造执行流树
	root, err := flow.BuildTaskRoot(tasks)
	if err != nil {
		if err := ap.changeFlowState(flowID, &iface.FlowChange{
			State:  enumor.FlowFailed,
			Reason: err.Error(),
		}); err != nil {
			logs.Errorf("[async] [module-parser] change flow state %v", err)
			return err
		}
		logs.Errorf("[async] [module-parser] build task tree error %v", err)
		return err
	}

	// 获取可执行的节点
	executableTaskNodes := root.GetExecutableTaskNodes()
	if len(executableTaskNodes) == 0 {
		state := root.ComputeStatus()

		if state == enumor.FlowFailed || state == enumor.FlowSuccess {
			if err := ap.changeFlowState(flowID, &iface.FlowChange{
				State:  state,
				Reason: constant.DefaultJsonValue,
			}); err != nil {
				logs.Errorf("[async] [module-parser] change flow state error %v", err)
				return err
			}
		}

		return nil
	}

	// 存储任务流执行树
	taskTree := flow.NewTaskTree()
	taskTree.Root = root
	ap.taskTrees.Store(flowID, taskTree)

	// 可执行任务推送到执行器
	for _, taskNode := range executableTaskNodes {
		GetExecutor().Push(taskNode)
	}

	return nil
}

// changeFlowState 带有重试机制的修改任务流状态
func (ap *AsyncParser) changeFlowState(flowID string, flowChange *iface.FlowChange) error {
	// 获取任务流
	flow, err := iface.GetBackend().GetFlowByID(flowID)
	if err != nil {
		return err
	}

	// 状态校验
	if err := flowChange.State.ValidateBeforeState(flow.State); err != nil {
		return err
	}

	// 操作机制更新任务流状态
	maxRetryCount := uint32(3)
	r := retry.NewRetryPolicy(uint(maxRetryCount), [2]uint{1000, 15000})
	var lastError error
	lastError = nil
	for r.RetryCount() < maxRetryCount {
		if err := iface.GetBackend().SetFlowChange(flowID, flowChange); err != nil {
			lastError = err
			r.Sleep()
		}

		if lastError == nil {
			break
		}
	}

	return lastError
}

// 任务流解析协程
func (ap *AsyncParser) goWorker() {
	for task := range ap.workerQueue {
		if err := ap.executeNext(task); err != nil {
			logs.Errorf("[async] [module-parser] run executeNext func error %v", err)
		}
	}

	ap.workerWg.Done()
}

// 任务流解析函数体，根据任务获取下次可执行的任务集合
func (ap *AsyncParser) executeNext(taskNode *flow.TaskNode) error {
	tree, ok := ap.getTaskTree(taskNode.Task.FlowID)
	if !ok {
		return fmt.Errorf("flow %s can not found task tree", taskNode.Task.FlowID)
	}

	// 获取下次执行的任务
	executableTaskNodes, find := taskNode.GetNextTaskNodes()
	if !find {
		return fmt.Errorf("task %s can not found next task node", taskNode.Task.ID)
	}

	if len(executableTaskNodes) == 0 {
		state := tree.Root.ComputeStatus()

		if state == enumor.FlowFailed || state == enumor.FlowSuccess {
			if err := ap.changeFlowState(taskNode.Task.FlowID, &iface.FlowChange{
				State:  state,
				Reason: constant.DefaultJsonValue,
			}); err != nil {
				logs.Errorf("[async] [module-parser] change flow state error %v", err)
				return err
			}

			ap.taskTrees.Delete(taskNode.Task.FlowID)
		}

		return nil
	}

	// 可执行任务推送到执行器
	for _, taskNode := range executableTaskNodes {
		GetExecutor().Push(taskNode)
	}

	return nil
}

// 获取存储的任务流树
func (ap *AsyncParser) getTaskTree(flowID string) (*flow.TaskTree, bool) {
	tasks, ok := ap.taskTrees.Load(flowID)
	if !ok {
		return nil, false
	}

	return tasks.(*flow.TaskTree), true
}

// EntryTaskIns 任务写回到执行器用于获取下一批可执行的任务
func (ap *AsyncParser) EntryTask(taskNode *flow.TaskNode) {
	ap.workerQueue <- taskNode
}

// Close 解析器关闭函数
func (ap *AsyncParser) Close() {
	select {
	case <-ap.closeCh:
		logs.V(3).Infof("[async] [module-parser] parser has already closed")
		return
	default:
	}

	close(ap.closeCh)
	close(ap.workerQueue)

	ap.workerWg.Wait()
}
