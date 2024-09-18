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
	"hcm/pkg/api/core/task"
	core "hcm/pkg/api/core/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CreateDetailReq define create task detail request.
type CreateDetailReq struct {
	Items []CreateDetailField `json:"items" validate:"required,min=1"`
}

// Validate CreateDetailReq.
func (req CreateDetailReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CreateDetailField define task detail create field.
type CreateDetailField struct {
	BkBizID          int64                  `json:"bk_biz_id" validate:"required"`
	TaskManagementID string                 `json:"task_management_id" validate:"required"`
	FlowID           string                 `json:"flow_id"`
	TaskActionIDs    []string               `json:"task_action_ids"`
	Operation        enumor.TaskOperation   `json:"operation" validate:"required"`
	Param            interface{}            `json:"param" validate:"required"`
	State            enumor.TaskDetailState `json:"state"`
	Reason           string                 `json:"reason"`
	Extension        *core.ManagementExt    `json:"extension"`
}

// Validate CreateDetailField.
func (req CreateDetailField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateDetailReq define update task detail request.
type UpdateDetailReq struct {
	Items []UpdateTaskDetailField `json:"items" validate:"required,min=1"`
}

// Validate UpdateDetailReq.
func (req UpdateDetailReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
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
	State            enumor.TaskDetailState `json:"state"`
	Reason           string                 `json:"reason"`
	Extension        *core.DetailExt        `json:"extension"`
}

// Validate UpdateTaskDetailField.
func (req UpdateTaskDetailField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListDetailResult defines list result.
type ListDetailResult struct {
	Count   uint64        `json:"count"`
	Details []task.Detail `json:"details"`
}

// DetailListResp defines list task detail response.
type DetailListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListDetailResult `json:"data"`
}

// -------------------------- Delete --------------------------

// DeleteDetailReq task detail delete request.
type DeleteDetailReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate task detail delete request.
func (req *DeleteDetailReq) Validate() error {
	return validator.Validate.Struct(req)
}
