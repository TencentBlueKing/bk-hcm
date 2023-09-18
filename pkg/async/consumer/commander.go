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

// Commander 外部命令执行体
type Commander interface {
	RunFlows(flowIDs []string) error
	RetryFlows(flowIDs []string) error
	RetryTasks(taskIDs []string) error
	CancelTasks(taskIDs []string) error
}

// NewCommander new commander.
func NewCommander() Commander {
	return &commander{}
}

// commander ...
type commander struct {
}

// RunFlows 控制执行多个任务流
func (ac *commander) RunFlows(flowIDs []string) error {
	return nil
}

// RetryFlows 控制重试多个任务流
func (ac *commander) RetryFlows(flowIDs []string) error {
	return nil
}

// RetryTasks 控制重试多个任务
func (ac *commander) RetryTasks(taskIDs []string) error {
	return nil
}

// CancelTasks 控制取消多个任务
func (ac *commander) CancelTasks(taskIDs []string) error {
	return nil
}
