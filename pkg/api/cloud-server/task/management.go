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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// ManagementListStateReq list task management state request.
type ManagementListStateReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate ManagementListStateReq.
func (req ManagementListStateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ManagementListStateResult defines list task management state result.
type ManagementListStateResult struct {
	Details []ManagementState `json:"details"`
}

// ManagementState defines task management state.
type ManagementState struct {
	ID    string                     `json:"id"`
	State enumor.TaskManagementState `json:"state"`
}

// ManagementListByCondReq list task management request.
type ManagementListByCondReq struct {
	Resource   enumor.TaskManagementResource `json:"resource" validate:"required"`
	StartTime  string                        `json:"start_time" validate:"required"`
	EndTime    string                        `json:"end_time" validate:"required"`
	Page       *core.BasePage                `json:"page" validate:"required"`
	AccountIDs []string                      `json:"account_ids" validate:"omitempty,max=10"`
	Operations []string                      `json:"operations" validate:"omitempty,max=10"`
	Source     enumor.TaskManagementSource   `json:"source" validate:"omitempty"`
	State      enumor.TaskManagementState    `json:"state" validate:"omitempty"`
	Creator    string                        `json:"creator" validate:"omitempty"`
}

// Validate ManagementListByCondReq.
func (r ManagementListByCondReq) Validate() error {
	if err := r.Resource.Validate(); err != nil {
		return err
	}

	if r.Page == nil {
		return errf.Newf(errf.InvalidParameter, "page can't be nil")
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	if r.Source != "" {
		if err := r.Source.Validate(); err != nil {
			return err
		}
	}

	if r.State != "" {
		if err := r.State.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(r)
}
