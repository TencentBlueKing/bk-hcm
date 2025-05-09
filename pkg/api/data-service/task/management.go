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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/task"
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CreateManagementReq define create task management request.
type CreateManagementReq struct {
	Items []CreateManagementField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate CreateManagementReq.
func (req CreateManagementReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateManagementField define task management create field.
type CreateManagementField struct {
	BkBizID    int64                         `json:"bk_biz_id" validate:"required"`
	Source     enumor.TaskManagementSource   `json:"source" validate:"required"`
	Vendors    []enumor.Vendor               `json:"vendors" validate:"required"`
	State      enumor.TaskManagementState    `json:"state"`
	AccountIDs []string                      `json:"account_ids" validate:"required"`
	Resource   enumor.TaskManagementResource `json:"resource" validate:"required"`
	Operations []enumor.TaskOperation        `json:"operations" validate:"required"`
	FlowIDs    []string                      `json:"flow_ids"`
	Extension  *coretask.ManagementExt       `json:"extension"`
}

// Validate CreateManagementField.
func (req CreateManagementField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateManagementReq define update task management request.
type UpdateManagementReq struct {
	Items []UpdateTaskManagementField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate UpdateManagementReq.
func (req UpdateManagementReq) Validate() error {
	return validator.Validate.Struct(req)
}

// UpdateTaskManagementField define task management update field.
type UpdateTaskManagementField struct {
	ID string `json:"id" validate:"required"`

	BkBizID    int64                         `json:"bk_biz_id"`
	Source     enumor.TaskManagementSource   `json:"source"`
	Vendors    []enumor.Vendor               `json:"vendors"`
	State      enumor.TaskManagementState    `json:"state"`
	AccountIDs []string                      `json:"account_ids"`
	Resource   enumor.TaskManagementResource `json:"resource"`
	Operations []enumor.TaskOperation        `json:"operations"`
	FlowIDs    []string                      `json:"flow_ids"`
	Extension  *coretask.ManagementExt       `json:"extension,omitempty"`
}

// Validate UpdateTaskDetailField.
func (req UpdateTaskManagementField) Validate() error {
	return validator.Validate.Struct(req)
}

// CancelReq define cancel request.
type CancelReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate CancelReq.
func (req CancelReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListManagementResult defines list result.
type ListManagementResult = core.ListResultT[task.Management]

// ManagementListResp defines list task management response.
type ManagementListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListManagementResult `json:"data"`
}

// -------------------------- Delete --------------------------

// DeleteManagementReq task management delete request.
type DeleteManagementReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate task management delete request.
func (req *DeleteManagementReq) Validate() error {
	return validator.Validate.Struct(req)
}
