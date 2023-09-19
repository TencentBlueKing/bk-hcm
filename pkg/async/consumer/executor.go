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
	"fmt"
	"sync"
	"time"

	"hcm/pkg/async/closer"
	"hcm/pkg/async/flow"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Executor 任务执行器
type Executor interface {
	closer.Closer

	// Start 启动执行器。
	Start()
	// SetGetParserFunc 设置获取解析器函数，运行过程中，执行完的节点需要通过解析器获取子节点，且解析器会下发任务到执行器.
	SetGetParserFunc(f func() Parser)

	// Push 推送task并执行。
	Push(task *flow.TaskNode)
	// CancelTasks 关闭指定task_id的任务。
	CancelTasks(taskIDs []string) error
}

var _ Executor = new(executor)

// executor 定义任务执行器
type executor struct {
	cancelMap         sync.Map
	workerNumber      int
	normalIntervalSec time.Duration
	workerWg          sync.WaitGroup
	initWg            sync.WaitGroup
	workerQueue       chan *flow.TaskNode
	initQueue         chan *flow.TaskNode

	GetParserFunc func() Parser
}

// SetGetParserFunc 设置获取解析器函数，运行过程中，执行完的节点需要通过解析器获取子节点，且解析器会下发任务到执行器.
func (exec *executor) SetGetParserFunc(f func() Parser) {
	exec.GetParserFunc = f
}

// NewExecutor 实例化任务执行器
func NewExecutor(workerNumber int, normalIntervalSec time.Duration) Executor {
	return &executor{
		workerWg:          sync.WaitGroup{},
		initWg:            sync.WaitGroup{},
		workerQueue:       make(chan *flow.TaskNode, 10),
		initQueue:         make(chan *flow.TaskNode),
		workerNumber:      workerNumber,
		normalIntervalSec: normalIntervalSec,
	}
}

// Start 初始化执行器并启动执行
func (exec *executor) Start() {
	// 待执行的任务预处理
	exec.initWg.Add(1)
	go exec.watchInitQueue()

	// 启动workerNumber个执行器执行任务
	for i := 0; i < exec.workerNumber; i++ {
		exec.workerWg.Add(1)
		go exec.subWorkerQueue()
	}
}

// 从initQueue队列获取待执行的任务协程
func (exec *executor) watchInitQueue() {
	for p := range exec.initQueue {
		exec.initWorkerTask(p)
	}

	exec.initWg.Done()
}

// 待执行任务的预处理函数
func (exec *executor) initWorkerTask(taskNode *flow.TaskNode) {
	if _, ok := exec.cancelMap.Load(taskNode.Task.ID); ok {
		logs.V(3).Infof("[async] [module-executor] task %s is already running", taskNode.Task.ID)
		return
	}

	// 设置超时控制
	c, cancel := context.WithTimeout(context.TODO(), time.Duration(taskNode.Task.TimeoutSecs)*time.Second)
	taskNode.Task.SetCtxWithTimeOut(c)

	// 设置kit
	kt := kit.NewAsyncKit()
	newRid := fmt.Sprintf("%s-%s-%s", taskNode.Task.FlowID, taskNode.Task.ID, kt.Rid)
	kt.Rid = newRid
	taskNode.Task.SetKit(kt)

	// cancel存储到cancelMap中
	exec.cancelMap.Store(taskNode.Task.ID, cancel)

	// 任务写回workerQueue
	exec.workerQueue <- taskNode
}

// 任务实际执行协程
func (exec *executor) subWorkerQueue() {
	for task := range exec.workerQueue {
		if err := exec.workerDo(task); err != nil {
			logs.Errorf("[async] [module-executor] workerDo func error %v", err)
		}
	}

	exec.workerWg.Done()
}

// 任务执行体
func (exec *executor) workerDo(taskNode *flow.TaskNode) error {
	// 执行任务
	err := taskNode.Task.DoTask()
	if err != nil {
		logs.Errorf("[async] [module-executor] do task action error %v", err)
	}

	// cancelMap清理执行完的任务
	defer exec.cancelMap.Delete(taskNode.Task.ID)

	// 执行完的任务回写到解析器用于获取待执行的任务
	exec.GetParserFunc().EntryTask(taskNode)

	return nil
}

// Push 任务写入到initQueue
func (exec *executor) Push(taskNode *flow.TaskNode) {
	exec.initQueue <- taskNode
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
	close(exec.initQueue)
	exec.initWg.Wait()
	close(exec.workerQueue)
	exec.workerWg.Wait()
}
