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
	"hcm/pkg/api/core"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/validator"
	typesasync "hcm/pkg/dal/dao/types/async"
	"hcm/pkg/kit"
)

// Backend - a common interface for all backends
type Backend interface {
	/*
		Flow 相关接口
	*/
	// CreateFlow 创建任务流
	CreateFlow(kt *kit.Kit, flow *model.Flow) (string, error)
	// BatchUpdateFlow 批量更新任务流
	BatchUpdateFlow(kt *kit.Kit, flows []model.Flow) error
	// ListFlow 查询任务流
	ListFlow(kt *kit.Kit, input *ListInput) ([]model.Flow, error)
	// BatchUpdateFlowStateByCAS CAS批量更新Flow状态
	BatchUpdateFlowStateByCAS(kt *kit.Kit, infos []UpdateFlowInfo) error

	/*
		Task 相关接口
	*/
	// BatchCreateTask 批量创建任务
	BatchCreateTask(kt *kit.Kit, tasks []model.Task) ([]string, error)
	// UpdateTask 更新任务
	UpdateTask(kt *kit.Kit, task *model.Task) error
	// UpdateTaskStateByCAS CAS更新任务状态
	UpdateTaskStateByCAS(kt *kit.Kit, info *UpdateTaskInfo) error
	// ListTask 查询任务
	ListTask(kt *kit.Kit, input *ListInput) ([]model.Task, error)

	// RetryTask 重试任务 将flow置为running, task 置为pending
	RetryTask(kt *kit.Kit, flowID, taskID string) error
}

// ListInput 查询输入参数
type ListInput core.ListReq

// UpdateFlowInfo define update flow info.
type UpdateFlowInfo typesasync.UpdateFlowInfo

// Validate UpdateFlowInfo
func (info *UpdateFlowInfo) Validate() error {
	return validator.Validate.Struct(info)
}

// UpdateTaskInfo define update task info.
type UpdateTaskInfo typesasync.UpdateTaskInfo

// Validate UpdateTaskInfo
func (info *UpdateTaskInfo) Validate() error {
	return validator.Validate.Struct(info)
}
