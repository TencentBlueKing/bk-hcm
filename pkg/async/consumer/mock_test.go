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

	coretask "hcm/pkg/api/core/task"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

type mockParser struct{}

// Close ...
func (psr *mockParser) Close() {
	fmt.Println("mock parser close")
}

func (psr *mockParser) Start() {
	fmt.Println("mock parser start")
}

func (psr *mockParser) EntryTask(task *task.Task) {
	fmt.Printf("mocker parser receive task: %+v\n", task)
}

type mockBackend struct{}

// SetBackendKit ...
func (bd *mockBackend) SetBackendKit(kt *kit.Kit) {
	// TODO implement me
	panic("implement me")
}

// ConsumeOnePendingFlow ...
func (bd *mockBackend) ConsumeOnePendingFlow() (*coretask.AsyncFlow, error) {
	// TODO implement me
	panic("implement me")
}

// GetFlowsByCount ...
func (bd *mockBackend) GetFlowsByCount(flowCount int) ([]coretask.AsyncFlow, error) {
	// TODO implement me
	panic("implement me")
}

// AddFlow ...
func (bd *mockBackend) AddFlow(req *taskserver.AddFlowReq) (string, error) {
	// TODO implement me
	panic("implement me")
}

// SetFlowChange ...
func (bd *mockBackend) SetFlowChange(flowID string, flowChange *backend.FlowChange) error {
	// TODO implement me
	panic("implement me")
}

// GetFlowByID ...
func (bd *mockBackend) GetFlowByID(flowID string) (*coretask.AsyncFlow, error) {
	// TODO implement me
	panic("implement me")
}

// GetFlows ...
func (bd *mockBackend) GetFlows(req *taskserver.FlowListReq) ([]*coretask.AsyncFlow, error) {
	// TODO implement me
	panic("implement me")
}

// AddTasks ...
func (bd *mockBackend) AddTasks(tasks []coretask.AsyncFlowTask) error {
	// TODO implement me
	panic("implement me")
}

// GetTasks ...
func (bd *mockBackend) GetTasks(taskIDs []string) ([]coretask.AsyncFlowTask, error) {
	// TODO implement me
	panic("implement me")
}

// GetTasksByFlowID ...
func (bd *mockBackend) GetTasksByFlowID(flowID string) ([]coretask.AsyncFlowTask, error) {
	// TODO implement me
	panic("implement me")
}

// SetTaskChange ...
func (bd *mockBackend) SetTaskChange(taskID string, taskChange *backend.TaskChange) error {
	fmt.Printf("backend SetTaskChange, taskID: %s, change: %+v\n", taskID, taskChange)
	return nil
}

// MakeTaskIDs ...
func (bd *mockBackend) MakeTaskIDs(num int) ([]string, error) {
	// TODO implement me
	panic("implement me")
}

// printTask ...
type printTask struct {
	ShareData map[string]string
}

// NewPrintTask ...
func NewPrintTask() *printTask {
	return &printTask{
		ShareData: make(map[string]string),
	}
}

// Name ...
func (task *printTask) Name() string {
	return string(enumor.TestPrintTask)
}

// NewParameter ...
func (task *printTask) NewParameter(parameter interface{}) interface{} {
	fmt.Println("print task exec NewParameter")
	return nil
}

// GetShareData ...
func (task *printTask) GetShareData() map[string]string {
	fmt.Println("print task exec GetShareData")
	return task.ShareData
}

// RunBefore ...
func (task *printTask) RunBefore(kt *kit.Kit, ctxWithTimeOut context.Context, params interface{}) error {
	fmt.Println("print task exec RunBefore")
	return nil
}

// Run ...
func (task *printTask) Run(kt *kit.Kit, ctxWithTimeOut context.Context, params interface{}) error {
	fmt.Println("print task exec Run")
	return nil
}

// RunBeforeSuccess ...
func (task *printTask) RunBeforeSuccess(kt *kit.Kit, ctxWithTimeOut context.Context, params interface{}) error {
	fmt.Println("print task exec RunBeforeSuccess")
	return nil
}

// RunBeforeFailed ...
func (task *printTask) RunBeforeFailed(kt *kit.Kit, ctxWithTimeOut context.Context, params interface{}) error {
	fmt.Println("print task exec RunBeforeFailed")
	return nil
}

// RetryBefore ...
func (task *printTask) RetryBefore(kt *kit.Kit, ctxWithTimeOut context.Context, params interface{}) error {
	fmt.Println("print task exec RetryBefore")
	return nil
}

type mockExecutor struct{}

// Close ...
func (exec *mockExecutor) Close() {
	// TODO implement me
	panic("implement me")
}

// Start ...
func (exec *mockExecutor) Start() {
	fmt.Println("mock executor exec Start")
}

// SetGetParserFunc ...
func (exec *mockExecutor) SetGetParserFunc(f func() Parser) {
	fmt.Println("mock executor exec SetGetParserFunc")
}

// Push ...
func (exec *mockExecutor) Push(task *task.Task) {
	fmt.Printf("mock executor exec Push, task: %+v\n", task)
}

// CancelTasks ...
func (exec *mockExecutor) CancelTasks(taskIDs []string) error {
	fmt.Printf("mock executor exec CancelTasks, taskIDs: %v\n", taskIDs)
	return nil
}

var _ Executor = new(mockExecutor)
