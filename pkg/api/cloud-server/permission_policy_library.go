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
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
)

// PermissionPolicyLibraryCreateReq defines create permission policy library request.
type PermissionPolicyLibraryCreateReq struct {
	Name           string  `json:"name" validate:"required,max=128"`
	PolicyDocument string  `json:"policy_document" validate:"required"`
	BkBizIDs       []int64 `json:"bk_biz_ids" validate:"required"`
	Memo           string  `json:"memo" validate:"required,max=255"`
}

// Validate PermissionPolicyLibraryCreateReq.
func (req *PermissionPolicyLibraryCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionPolicyLibraryUpdateReq defines update permission policy library request.
type PermissionPolicyLibraryUpdateReq struct {
	Name           string  `json:"name" validate:"omitempty,max=128"`
	PolicyDocument string  `json:"policy_document" validate:"omitempty"`
	BkBizIDs       []int64 `json:"bk_biz_ids" validate:"omitempty"`
	Memo           *string `json:"memo" validate:"omitempty,max=255"`
}

// Validate PermissionPolicyLibraryUpdateReq.
func (req *PermissionPolicyLibraryUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionPolicyLibraryCreateResult defines create permission policy library result.
type PermissionPolicyLibraryCreateResult struct {
	ID string `json:"id"`
}

// PermissionPolicyLibraryResult defines permission policy library result with extra computed fields.
type PermissionPolicyLibraryResult struct {
	corecloud.BasePermissionPolicyLibrary `json:",inline"`
	AssociatedAccountCount                int `json:"associated_account_count"`
}

// PermissionPolicyLibraryListResult defines list permission policy library result.
type PermissionPolicyLibraryListResult struct {
	Count   uint64                          `json:"count"`
	Details []PermissionPolicyLibraryResult `json:"details"`
}
