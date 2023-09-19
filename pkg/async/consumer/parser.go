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

	"hcm/pkg/async/backend"
	"hcm/pkg/async/closer"
	"hcm/pkg/async/flow"
	"hcm/pkg/async/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
)

// Parser 任务流解析器
type Parser interface {
	closer.Closer

	// Start 启动解析器。
	Start()

	// EntryTask 分析执行完的任务，并解析出当前任务的子任务去执行。
	EntryTask(task *flow.TaskNode)
}

// parser 定义任务流解析器
type parser struct {
	taskTrees         sync.Map
	workerNumber      int
	normalIntervalSec time.Duration
	workerQueue       chan *flow.TaskNode
	workerWg          sync.WaitGroup
	backend           backend.Backend
	executor          Executor

	closeCh chan struct{}
}

// NewParser 实例化任务流解析器
func NewParser(bd backend.Backend, exec Executor, workerNumber int, normalIntervalSec time.Duration) Parser {
	return &parser{
		closeCh:           make(chan struct{}),
		workerWg:          sync.WaitGroup{},
		workerQueue:       make(chan *flow.TaskNode, 10),
		workerNumber:      workerNumber,
		normalIntervalSec: normalIntervalSec,
		backend:           bd,
		executor:          exec,
	}
}

// Start 初始化解析器并启动执行
func (psr *parser) Start() {
	kt := NewAsyncKit()

	// 定期获取等待执行的任务流
	psr.workerWg.Add(1)
	go psr.startWatcher(kt, psr.watchPendingFlow)

	// 启动workerNumber个协程进行任务流解析
	for i := 0; i < psr.workerNumber; i++ {
		psr.workerWg.Add(1)
		go psr.goWorker(kt)
	}
}

// startWatcher 定期执行do函数体
func (psr *parser) startWatcher(kt *kit.Kit, do func(kt *kit.Kit) error) {
	ticker := time.NewTicker(psr.normalIntervalSec)
	defer ticker.Stop()

	closed := false
	for !closed {
		select {
		case <-psr.closeCh:
			closed = true
		case <-ticker.C:
			if err := do(kt); err != nil {
				logs.Errorf("[async] [module-parser] do watch func error %v", err)
			}
		}
	}

	psr.workerWg.Done()
}

func (psr *parser) watchPendingFlow(kt *kit.Kit) error {
	// 从DB中获取一条待执行的任务流并更新状态为执行中
	flowFromDB, err := psr.backend.ConsumeOnePendingFlow(kt)
	if err != nil {
		if strings.Contains(err.Error(), "flow num is 0") {
			return nil
		}
		logs.Errorf("[async] [module-parser] get flow error %v", err)
		return err
	}
	flowID := flowFromDB.ID

	// 根据任务流ID获取对应的任务集合
	taskResult, err := psr.backend.GetTasksByFlowID(kt, flowID)
	if err != nil {
		if err := psr.changeFlowState(kt, flowID, &backend.FlowChange{
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
		if err := psr.changeFlowState(kt, flowID, &backend.FlowChange{
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
			if err := psr.changeFlowState(kt, flowID, &backend.FlowChange{
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
	psr.taskTrees.Store(flowID, taskTree)

	// 可执行任务推送到执行器
	for _, taskNode := range executableTaskNodes {
		psr.executor.Push(taskNode)
	}

	return nil
}

// changeFlowState 带有重试机制的修改任务流状态
func (psr *parser) changeFlowState(kt *kit.Kit, flowID string, flowChange *backend.FlowChange) error {
	// 获取任务流
	flow, err := psr.backend.GetFlowByID(kt, flowID)
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
		if err := psr.backend.SetFlowChange(kt, flowID, flowChange); err != nil {
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
func (psr *parser) goWorker(kt *kit.Kit) {
	for task := range psr.workerQueue {
		if err := psr.executeNext(kt, task); err != nil {
			logs.Errorf("[async] [module-parser] run executeNext func error %v", err)
		}
	}

	psr.workerWg.Done()
}

// 任务流解析函数体，根据任务获取下次可执行的任务集合
func (psr *parser) executeNext(kt *kit.Kit, taskNode *flow.TaskNode) error {
	tree, ok := psr.getTaskTree(taskNode.Task.FlowID)
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
			if err := psr.changeFlowState(kt, taskNode.Task.FlowID, &backend.FlowChange{
				State:  state,
				Reason: constant.DefaultJsonValue,
			}); err != nil {
				logs.Errorf("[async] [module-parser] change flow state error %v", err)
				return err
			}

			psr.taskTrees.Delete(taskNode.Task.FlowID)
		}

		return nil
	}

	// 可执行任务推送到执行器
	for _, taskNode := range executableTaskNodes {
		psr.executor.Push(taskNode)
	}

	return nil
}

// 获取存储的任务流树
func (psr *parser) getTaskTree(flowID string) (*flow.TaskTree, bool) {
	tasks, ok := psr.taskTrees.Load(flowID)
	if !ok {
		return nil, false
	}

	return tasks.(*flow.TaskTree), true
}

// EntryTask 任务写回到执行器用于获取下一批可执行的任务
func (psr *parser) EntryTask(taskNode *flow.TaskNode) {
	psr.workerQueue <- taskNode
}

// Close 解析器关闭函数
func (psr *parser) Close() {
	select {
	case <-psr.closeCh:
		logs.V(3).Infof("[async] [module-parser] parser has already closed")
		return
	default:
	}

	close(psr.closeCh)
	close(psr.workerQueue)

	psr.workerWg.Wait()
}
