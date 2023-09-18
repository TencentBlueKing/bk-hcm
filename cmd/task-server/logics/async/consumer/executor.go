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

	"hcm/cmd/task-server/logics/async/flow"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AsyncExecutor 定义任务执行器
type AsyncExecutor struct {
	cancelMap         sync.Map
	workerNumber      int
	normalIntervalSec time.Duration
	workerWg          sync.WaitGroup
	initWg            sync.WaitGroup
	workerQueue       chan *flow.TaskNode
	initQueue         chan *flow.TaskNode
}

// NewExecutor 实例化任务执行器
func NewExecutor(workerNumber int, normalIntervalSec time.Duration) *AsyncExecutor {
	return &AsyncExecutor{
		workerWg:          sync.WaitGroup{},
		initWg:            sync.WaitGroup{},
		workerQueue:       make(chan *flow.TaskNode, 10),
		initQueue:         make(chan *flow.TaskNode),
		workerNumber:      workerNumber,
		normalIntervalSec: normalIntervalSec,
	}
}

// Init 初始化执行器并启动执行
func (ae *AsyncExecutor) Init() {
	// 待执行的任务预处理
	ae.initWg.Add(1)
	go ae.watchInitQueue()

	// 启动workerNumber个执行器执行任务
	for i := 0; i < ae.workerNumber; i++ {
		ae.workerWg.Add(1)
		go ae.subWorkerQueue()
	}
}

// 从initQueue队列获取待执行的任务协程
func (ae *AsyncExecutor) watchInitQueue() {
	for p := range ae.initQueue {
		ae.initWorkerTask(p)
	}

	ae.initWg.Done()
}

// 待执行任务的预处理函数
func (ae *AsyncExecutor) initWorkerTask(taskNode *flow.TaskNode) {
	if _, ok := ae.cancelMap.Load(taskNode.Task.ID); ok {
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
	ae.cancelMap.Store(taskNode.Task.ID, cancel)

	// 任务写回workerQueue
	ae.workerQueue <- taskNode
}

// 任务实际执行协程
func (ae *AsyncExecutor) subWorkerQueue() {
	for task := range ae.workerQueue {
		if err := ae.workerDo(task); err != nil {
			logs.Errorf("[async] [module-executor] workerDo func error %v", err)
		}
	}

	ae.workerWg.Done()
}

// 任务执行体
func (ae *AsyncExecutor) workerDo(taskNode *flow.TaskNode) error {
	// 执行任务
	err := taskNode.Task.DoTask()
	if err != nil {
		logs.Errorf("[async] [module-executor] do task action error %v", err)
	}

	// cancelMap清理执行完的任务
	defer ae.cancelMap.Delete(taskNode.Task.ID)

	// 执行完的任务回写到解析器用于获取待执行的任务
	GetParser().EntryTask(taskNode)

	return nil
}

// Push 任务写入到initQueue
func (ae *AsyncExecutor) Push(taskNode *flow.TaskNode) {
	ae.initQueue <- taskNode
}

// CancelTasks 停止指定id的任务
func (ae *AsyncExecutor) CancelTasks(taskIDs []string) error {
	for _, id := range taskIDs {
		if cancel, ok := ae.cancelMap.Load(id); ok {
			ae.cancelMap.Delete(id)
			cancel.(context.CancelFunc)()
		}
	}

	return nil
}

// Close 执行器关闭函数
func (ae *AsyncExecutor) Close() {
	close(ae.initQueue)
	ae.initWg.Wait()
	close(ae.workerQueue)
	ae.workerWg.Wait()
}
