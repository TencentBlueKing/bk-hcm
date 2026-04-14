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
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// PermissionPolicyLibraryBatchCreateReq permission policy library batch create request.
type PermissionPolicyLibraryBatchCreateReq struct {
	PermissionPolicyLibraries []PermissionPolicyLibraryCreate `json:"permission_policy_libraries" validate:"required,min=1"`
}

// Validate permission policy library batch create request.
func (req *PermissionPolicyLibraryBatchCreateReq) Validate() error {
	if len(req.PermissionPolicyLibraries) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("permission_policy_libraries count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// PermissionPolicyLibraryCreate defines a single item in batch create request.
type PermissionPolicyLibraryCreate struct {
	Name           string  `json:"name" validate:"required,max=128"`
	PolicyDocument string  `json:"policy_document" validate:"required"`
	BkBizIDs       []int64 `json:"bk_biz_ids" validate:"required"`
	Memo           *string `json:"memo" validate:"omitempty,max=255"`
}

// -------------------------- Update --------------------------

// PermissionPolicyLibraryBatchUpdateReq permission policy library batch update request.
type PermissionPolicyLibraryBatchUpdateReq struct {
	PermissionPolicyLibraries []PermissionPolicyLibraryUpdate `json:"permission_policy_libraries" validate:"required,min=1"`
}

// Validate permission policy library batch update request.
func (req *PermissionPolicyLibraryBatchUpdateReq) Validate() error {
	if len(req.PermissionPolicyLibraries) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("permission_policy_libraries count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// PermissionPolicyLibraryUpdate defines a single item in batch update request.
type PermissionPolicyLibraryUpdate struct {
	ID             string  `json:"id" validate:"required"`
	Name           string  `json:"name" validate:"omitempty,max=128"`
	PolicyDocument string  `json:"policy_document" validate:"omitempty"`
	BkBizIDs       []int64 `json:"bk_biz_ids" validate:"omitempty"`
	Memo           *string `json:"memo" validate:"omitempty,max=255"`
}

// -------------------------- Delete --------------------------

// PermissionPolicyLibraryBatchDeleteReq permission policy library batch delete request.
type PermissionPolicyLibraryBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate permission policy library batch delete request.
func (req *PermissionPolicyLibraryBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// PermissionPolicyLibraryListReq permission policy library list request.
type PermissionPolicyLibraryListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate permission policy library list request.
func (req *PermissionPolicyLibraryListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionPolicyLibraryListResult permission policy library list result.
type PermissionPolicyLibraryListResult struct {
	Count   uint64                                  `json:"count"`
	Details []corecloud.BasePermissionPolicyLibrary `json:"details"`
}

// PermissionPolicyLibraryListResp permission policy library list response.
type PermissionPolicyLibraryListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *PermissionPolicyLibraryListResult `json:"data"`
}
