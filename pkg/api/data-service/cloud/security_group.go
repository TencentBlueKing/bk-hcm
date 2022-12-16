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

package cloud

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// SecurityGroupCreateReq security group create request.
type SecurityGroupCreateReq[Extension cloud.SecurityGroupExtension] struct {
	Spec      *cloud.SecurityGroupSpec `json:"spec" validate:"required"`
	Extension *Extension               `json:"extension" validate:"required"`
}

// Validate security group create request.
func (req *SecurityGroupCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq security group update request.
type SecurityGroupUpdateReq[Extension cloud.SecurityGroupExtension] struct {
	Spec      *SecurityGroupSpecUpdate `json:"spec" validate:"omitempty"`
	Extension *Extension               `json:"extension" validate:"omitempty"`
}

// Validate security group update request.
func (req *SecurityGroupUpdateReq[T]) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Spec == nil && req.Extension == nil {
		return errf.New(errf.InvalidParameter, "spec and extension require at least one for update")
	}

	return nil
}

// SecurityGroupSpecUpdate define security group spec when update.
type SecurityGroupSpecUpdate struct {
	Name     string  `json:"name" validate:"omitempty"`
	Assigned bool    `json:"assigned" validate:"omitempty"`
	Memo     *string `json:"memo" validate:"omitempty"`
}

// -------------------------- List --------------------------

// SecurityGroupListReq security group list req.
type SecurityGroupListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate security group list request.
func (req *SecurityGroupListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SecurityGroupListResult define security group list result.
type SecurityGroupListResult struct {
	Count   uint64                     `json:"count,omitempty"`
	Details []*cloud.BaseSecurityGroup `json:"details,omitempty"`
}

// SecurityGroupListResp define security group list resp.
type SecurityGroupListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SecurityGroupListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// SecurityGroupDeleteReq security group delete request.
type SecurityGroupDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate security group delete request.
func (req *SecurityGroupDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Get --------------------------

// SecurityGroupGetResp define security group get resp.
type SecurityGroupGetResp[T cloud.SecurityGroupExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *cloud.SecurityGroup[T] `json:"data"`
}
