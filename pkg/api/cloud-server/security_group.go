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

package cloudserver

import (
	"errors"

	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/runtime/filter"
)

// -------------------------- List --------------------------

// SecurityGroupListReq security group list req.
type SecurityGroupListReq struct {
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

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq security group update request.
type SecurityGroupUpdateReq struct {
	Spec *SecurityGroupSpecUpdate `json:"spec" validate:"required"`
}

// Validate security group update request.
func (req *SecurityGroupUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// SecurityGroupSpecUpdate define security group spec when update.
type SecurityGroupSpecUpdate struct {
	Name string  `json:"name"`
	Memo *string `json:"memo"`
}

// Validate security group spec when update.
func (spec *SecurityGroupSpecUpdate) Validate() error {
	if len(spec.Name) == 0 && spec.Memo == nil {
		return errors.New("name or memo is required")
	}

	if len(spec.Name) != 0 {

	}

	return nil
}
