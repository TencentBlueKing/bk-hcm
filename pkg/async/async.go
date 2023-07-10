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

package async

import (
	"context"
	"fmt"

	"hcm/pkg/cc"
	"hcm/pkg/logs"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	mtasks "github.com/RichardKnop/machinery/v1/tasks"
	"github.com/google/uuid"
)

// AsyncServer interface
type AsyncServer interface {
	// 初始化生产者
	InitProducer(cnf cc.Async) error
	// 初始化消费者
	InitConsumer(cnf cc.Async) error
	// 发送单任务
	SendSingleTask(tasks *mtasks.Signature) (string, error)
	// 发送组任务
	SendGroupTasks(tasks ...*mtasks.Signature) (string, error)
	// 发送带回调的组任务
	SendChordTasks(callback *mtasks.Signature, tasks ...*mtasks.Signature) (string, error)
	// 发送链任务
	SendChainTasks(tasks ...*mtasks.Signature) (string, error)
	// 获取单任务状态
	GetSingleTaskState(taskID string) (*mtasks.TaskState, error)
	// 获取组任务状态
	GetGroupTasksState(groupID string, groupTaskCount int) ([]*mtasks.TaskState, error)
	// 获取带回调的组任务状态
	GetChordTasksState(groupID string, groupTaskCount int,
		callbackID string) ([]*mtasks.TaskState, *mtasks.TaskState, error)
	// 获取链任务状态
	GetChainTasksState(taskIDs ...string) ([]*mtasks.TaskState, error)
}

// TaskServer task server
type TaskServer struct {
	cxt         context.Context
	cancel      context.CancelFunc
	server      *machinery.Server
	worker      *machinery.Worker
	tasks       *TaskManager
	managerType string
}

var _ AsyncServer = new(TaskServer)

// NewTaskServer new task server
func NewTaskServer(taskManagerType string) *TaskServer {
	cxt, cancel := context.WithCancel(context.Background())
	return &TaskServer{
		cxt:         cxt,
		cancel:      cancel,
		tasks:       GetTaskManager(),
		managerType: taskManagerType,
	}
}

// TODO: 支持复杂任务编排
// TODO: 配置扩展支持多种backend和broker插件
// TODO: 如何支持任务优先级（多队列）
// TODO: 优雅退出失效
// InitProducer 初始化生产者
func (ts *TaskServer) InitProducer(cnf cc.Async) error {
	mCnf := &config.Config{
		DefaultQueue:    cnf.Queue,
		ResultsExpireIn: 3600,
		Broker:          cnf.Broker,
		ResultBackend:   cnf.Backend,
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	var err error
	ts.server, err = machinery.NewServer(mCnf)
	if err != nil {
		logs.Errorf("async task server new machinery failed, %v", err)
		return err
	}

	allTasks := ts.tasks.GetAllTasksByManagerType(ts.managerType)
	if err := ts.server.RegisterTasks(allTasks); err != nil {
		logs.Errorf("async task server register tasks failed, %v", err)
		return err
	}

	return nil
}

// InitConsumer 初始化消费者
func (ts *TaskServer) InitConsumer(cnf cc.Async) error {
	ts.worker = ts.server.NewWorker(cnf.Queue, 10)

	errorHandler := func(err error) {
		logs.Errorf("async task error handler: %v", err)
	}
	preTaskHandler := func(signature *mtasks.Signature) {
		logs.Infof("async start task handler for: %s", signature.Name)
	}
	postTaskHandler := func(signature *mtasks.Signature) {
		logs.Infof("async end task handler for: %s", signature.Name)
	}
	ts.worker.SetPostTaskHandler(postTaskHandler)
	ts.worker.SetErrorHandler(errorHandler)
	ts.worker.SetPreTaskHandler(preTaskHandler)

	go func() {
		if err := ts.worker.Launch(); err != nil {
			logs.Errorf("async task server worker launch failed, %v", err)
			return
		}
	}()

	return nil
}

/*
	***SingleTask***
	{
		"task_type":"task",
		"steps":[
			{
				"task_name": "tcloud_create_cvm",
				"task_priority": 3
			}
		]
	}

	***SingleTaskWithOnError***
	{
		"task_type":"task",
		"steps":[
			{
				"task_name": "create_clb",
				"task_priority": 3,
				"args":[
					{
						"type": "string",
						"value": "test"
					}
				],
				"on_error":[
					{
						"task_name": "undo_create_clb",
						"task_priority": 3
					}
				]
			}
		]
	}

	***ChainTask***
	{
		"task_type":"chain",
		"steps":[
			{
				"task_name": "before_check_clb",
				"task_priority": 3,
				"args":[
					{
						"type": "string",
						"value": "test"
					}
				]
			},
			{
				"task_name": "check_clb",
				"task_priority": 3
			}
		]
	}
*/

// SendSingleTask 发送单任务
func (ts *TaskServer) SendSingleTask(task *mtasks.Signature) (string, error) {
	if task.UUID == "" {
		task.UUID = fmt.Sprintf("task_%v", uuid.New().String())
	}

	err := rebuildOnFunc(task)
	if err != nil {
		return "", fmt.Errorf("async rebuild OnFunc failed: %s", err.Error())
	}

	result, err := ts.server.SendTaskWithContext(context.Background(), task)
	if err != nil {
		return "", fmt.Errorf("async could not send task: %s", err.Error())
	}

	return result.Signature.UUID, nil
}

// SendGroupTasks 发送组任务
func (ts *TaskServer) SendGroupTasks(tasks ...*mtasks.Signature) (string, error) {
	group, err := mtasks.NewGroup(tasks...)
	if err != nil {
		return "", fmt.Errorf("async creating group error: %s", err.Error())
	}

	err = rebuildOnFunc(tasks...)
	if err != nil {
		return "", fmt.Errorf("async rebuild OnFunc failed: %s", err.Error())
	}

	_, err = ts.server.SendGroupWithContext(context.Background(), group, 10)
	if err != nil {
		return "", fmt.Errorf("async could not send group: %s", err.Error())
	}

	return group.GroupUUID, nil
}

// SendChordTasks 发送带回调的组任务
func (ts *TaskServer) SendChordTasks(callback *mtasks.Signature, tasks ...*mtasks.Signature) (string, error) {
	group, err := mtasks.NewGroup(tasks...)
	if err != nil {
		return "", fmt.Errorf("async creating group error: %s", err.Error())
	}

	err = rebuildOnFunc(tasks...)
	if err != nil {
		return "", fmt.Errorf("async rebuild OnFunc failed: %s", err.Error())
	}

	chord, err := mtasks.NewChord(group, callback)
	if err != nil {
		return "", fmt.Errorf("async creating chord error: %s", err)
	}

	err = rebuildOnFunc(callback)
	if err != nil {
		return "", fmt.Errorf("async rebuild OnFunc failed: %s", err.Error())
	}

	_, err = ts.server.SendChordWithContext(context.Background(), chord, 10)
	if err != nil {
		return "", fmt.Errorf("async could not send chord: %s", err.Error())
	}

	return group.GroupUUID, nil
}

// SendChainTasks 发送链任务
func (ts *TaskServer) SendChainTasks(tasks ...*mtasks.Signature) (string, error) {
	chain, err := mtasks.NewChain(tasks...)
	if err != nil {
		return "", fmt.Errorf("async creating chain error: %s", err)
	}

	err = rebuildOnFunc(tasks...)
	if err != nil {
		return "", fmt.Errorf("async rebuild OnFunc failed: %s", err.Error())
	}

	_, err = ts.server.SendChainWithContext(context.Background(), chain)
	if err != nil {
		return "", fmt.Errorf("async could not send chain: %s", err.Error())
	}

	return chain.Tasks[0].UUID, nil
}

// GetSingleTaskState 获取单任务状态
func (ts *TaskServer) GetSingleTaskState(taskID string) (*mtasks.TaskState, error) {
	taskState, err := ts.server.GetBackend().GetState(taskID)
	if err != nil {
		return nil, fmt.Errorf("async get single task state error: %s", err)
	}

	return taskState, nil
}

// GetGroupTasksState 获取组任务状态
func (ts *TaskServer) GetGroupTasksState(groupID string, groupTaskCount int) ([]*mtasks.TaskState, error) {
	taskStates, err := ts.server.GetBackend().GroupTaskStates(groupID, groupTaskCount)
	if err != nil {
		return nil, fmt.Errorf("async get group task state error: %s", err)
	}

	return taskStates, nil
}

// GetChordTasksState 获取带回调的组任务状态
func (ts *TaskServer) GetChordTasksState(groupID string, groupTaskCount int,
	callbackID string) ([]*mtasks.TaskState, *mtasks.TaskState, error) {

	taskStates, err := ts.server.GetBackend().GroupTaskStates(groupID, groupTaskCount)
	if err != nil {
		return nil, nil, fmt.Errorf("async get chord group task state error: %s", err)
	}

	taskState, err := ts.server.GetBackend().GetState(callbackID)
	if err != nil {
		return taskStates, nil, fmt.Errorf("async get single task state error: %s", err)
	}

	return taskStates, taskState, nil
}

// GetChainTasksState 获取链任务状态
func (ts *TaskServer) GetChainTasksState(taskIDs ...string) ([]*mtasks.TaskState, error) {
	taskStates := make([]*mtasks.TaskState, 0, len(taskIDs))

	for _, id := range taskIDs {
		taskState, err := ts.server.GetBackend().GetState(id)
		if err != nil {
			taskStates = append(taskStates, &mtasks.TaskState{
				TaskUUID: id,
			})
		}
		taskStates = append(taskStates, taskState)
	}

	return taskStates, nil
}

func rebuildOnFunc(tasks ...*mtasks.Signature) error {
	for _, task := range tasks {
		if len(task.OnError) > 0 {
			for _, one := range task.OnError {
				one.UUID = fmt.Sprintf("task_on_error_%v", uuid.New().String())
				one.RoutingKey = task.RoutingKey
			}
		}

		if len(task.OnSuccess) > 0 {
			for _, one := range task.OnSuccess {
				one.UUID = fmt.Sprintf("task_on_success_%v", uuid.New().String())
				one.RoutingKey = task.RoutingKey
			}
		}
	}

	return nil
}
