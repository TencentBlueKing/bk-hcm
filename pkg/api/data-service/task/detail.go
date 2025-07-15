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

package task

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/task"
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CreateDetailReq define create task detail request.
type CreateDetailReq struct {
	Items []CreateDetailField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate CreateDetailReq.
func (req CreateDetailReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateDetailField define task detail create field.
type CreateDetailField struct {
	BkBizID          int64                   `json:"bk_biz_id" validate:"required"`
	TaskManagementID string                  `json:"task_management_id" validate:"required"`
	FlowID           string                  `json:"flow_id"`
	TaskActionIDs    []string                `json:"task_action_ids"`
	Operation        enumor.TaskOperation    `json:"operation" validate:"required"`
	Param            interface{}             `json:"param" validate:"required"`
	Result           interface{}             `json:"result"`
	State            enumor.TaskDetailState  `json:"state"`
	Reason           string                  `json:"reason"`
	Extension        *coretask.ManagementExt `json:"extension"`
}

// Validate CreateDetailField.
func (req CreateDetailField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateDetailReq define update task detail request.
type UpdateDetailReq struct {
	Items []UpdateTaskDetailField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate UpdateDetailReq.
func (req UpdateDetailReq) Validate() error {
	return validator.Validate.Struct(req)
}

// UpdateTaskDetailField define task detail update field.
type UpdateTaskDetailField struct {
	ID string `json:"id" validate:"required"`

	BkBizID          int64                  `json:"bk_biz_id"`
	TaskManagementID string                 `json:"task_management_id"`
	FlowID           string                 `json:"flow_id"`
	TaskActionIDs    []string               `json:"task_action_ids"`
	Operation        enumor.TaskOperation   `json:"operation"`
	Param            interface{}            `json:"param,omitempty"`
	Result           interface{}            `json:"result,omitempty"`
	State            enumor.TaskDetailState `json:"state"`
	Reason           string                 `json:"reason"`
	Extension        *coretask.DetailExt    `json:"extension"`
}

// Validate UpdateTaskDetailField.
func (req UpdateTaskDetailField) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchUpdateTaskDetailReq ...
type BatchUpdateTaskDetailReq struct {
	IDs           []string               `json:"ids" validate:"required,min=1"`
	Reason        string                 `json:"reason"`
	State         enumor.TaskDetailState `json:"state"`
	FlowID        string                 `json:"flow_id"`
	TaskActionIDs []string               `json:"task_action_ids,omitempty"`
}

// Validate ...
func (req BatchUpdateTaskDetailReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListDetailResult defines list result.
type ListDetailResult = core.ListResultT[task.Detail]

// -------------------------- Delete --------------------------

// DeleteDetailReq task detail delete request.
type DeleteDetailReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate task detail delete request.
func (req *DeleteDetailReq) Validate() error {
	return validator.Validate.Struct(req)
}
