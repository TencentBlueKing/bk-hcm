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

// Package backend 后端存储
package backend

import (
	"hcm/pkg/api/core/task"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// Backend - a common interface for all backends
type Backend interface {
	// ConsumeOnePendingFlow consume one pending flow
	ConsumeOnePendingFlow(kt *kit.Kit) (*task.AsyncFlow, error)
	// GetFlowsByCount get flows by count from backend
	GetFlowsByCount(kt *kit.Kit, flowCount int) ([]task.AsyncFlow, error)
	// AddFlow add flow into backend
	AddFlow(kt *kit.Kit, req *taskserver.AddFlowReq) (string, error)
	// SetFlowChange set flow's change
	SetFlowChange(kt *kit.Kit, flowID string, flowChange *FlowChange) error
	// GetFlowByID get flow by id
	GetFlowByID(kt *kit.Kit, flowID string) (*task.AsyncFlow, error)
	// GetFlows get flows from backend
	GetFlows(kt *kit.Kit, req *taskserver.FlowListReq) ([]*task.AsyncFlow, error)
	// AddTasks add tasks into backend
	AddTasks(kt *kit.Kit, tasks []task.AsyncFlowTask) error
	// GetTasks get tasks from backend
	GetTasks(kt *kit.Kit, taskIDs []string) ([]task.AsyncFlowTask, error)
	// GetTasksByFlowID get tasks by flow id from backend
	GetTasksByFlowID(kt *kit.Kit, flowID string) ([]task.AsyncFlowTask, error)
	// SetTaskChange set task's change
	SetTaskChange(kt *kit.Kit, taskID string, taskChange *TaskChange) error
	// MakeTaskIDs make task ids
	MakeTaskIDs(kt *kit.Kit, num int) ([]string, error)
}

// FlowChange 任务流的变化
type FlowChange struct {
	State     enumor.FlowState `json:"state"`
	Reason    string           `json:"reason"`
	ShareData string           `json:"share_date"`
}

// TaskChange 任务的变化
type TaskChange struct {
	State     enumor.TaskState `json:"state"`
	Reason    string           `json:"reason"`
	ShareData string           `json:"share_date"`
}
