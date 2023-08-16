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

package iface

import (
	"hcm/pkg/api/core/task"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
)

// Backend - a common interface for all backends
type Backend interface {
	// ConsumeOnePendingFlow consume one pending flow
	ConsumeOnePendingFlow() (*task.AsyncFlow, error)
	// GetFlowsByCount get flows by count from backend
	GetFlowsByCount(flowCount int) ([]task.AsyncFlow, error)
	// AddFlow add flow into backend
	AddFlow(req *taskserver.AddFlowReq) (string, error)
	// SetFlowStateWithReason set flow state with reason
	SetFlowStateWithReason(flowID string, state enumor.FlowState, reason string) error
	// GetFlowByID get flow by id
	GetFlowByID(flowID string) (*task.AsyncFlow, error)
	// GetFlows get flows from backend
	GetFlows(req *taskserver.FlowListReq) ([]*task.AsyncFlow, error)
	// AddTasks add tasks into backend
	AddTasks(tasks []task.AsyncFlowTask) error
	// GetTasks get tasks from backend
	GetTasks(taskIDs []string) ([]task.AsyncFlowTask, error)
	// GetTasksByFlowID get tasks by flow id from backend
	GetTasksByFlowID(flowID string) ([]task.AsyncFlowTask, error)
	// SetTaskStateWithReason set task state with reason
	SetTaskStateWithReason(taskID string, state enumor.TaskState, reason string) error
	// MakeTaskIDs make task ids
	MakeTaskIDs(num int) ([]string, error)
}
