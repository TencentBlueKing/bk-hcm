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
	"fmt"
	"regexp"

	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
)

var permissionPolicyLibraryNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// PermissionPolicyLibraryCreateReq defines create permission policy library request.
type PermissionPolicyLibraryCreateReq struct {
	Name           string  `json:"name" validate:"required,max=128"`
	PolicyDocument string  `json:"policy_document" validate:"required"`
	BkBizIDs       []int64 `json:"bk_biz_ids" validate:"required"`
	Memo           string  `json:"memo" validate:"required,max=255"`
}

// Validate PermissionPolicyLibraryCreateReq.
func (req *PermissionPolicyLibraryCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	if !permissionPolicyLibraryNameRegexp.MatchString(req.Name) {
		return fmt.Errorf("invalid name: %s, only allows english letters, numbers, underscore (_) and hyphen (-)",
			req.Name)
	}
	return nil
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
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	if len(req.Name) > 0 && !permissionPolicyLibraryNameRegexp.MatchString(req.Name) {
		return fmt.Errorf("invalid name: %s, only allows english letters, numbers, underscore (_) and hyphen (-)",
			req.Name)
	}
	return nil
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

// ApplyPermissionPolicyLibraryCreateReq defines request for applying a permission policy library (create action).
type ApplyPermissionPolicyLibraryCreateReq struct {
	AccountIDs []string `json:"account_ids" validate:"required,min=1,max=100"`
}

// Validate ApplyPermissionPolicyLibraryCreateReq.
func (req *ApplyPermissionPolicyLibraryCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApplyPermissionPolicyLibraryUpdateReq defines request for applying a permission policy library (update action).
type ApplyPermissionPolicyLibraryUpdateReq struct {
	PermissionTemplateIDs []string `json:"permission_template_ids" validate:"required,min=1,max=100"`
}

// Validate ApplyPermissionPolicyLibraryUpdateReq.
func (req *ApplyPermissionPolicyLibraryUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApplyAccountResult defines the apply result for a single account (used in create action).
type ApplyAccountResult struct {
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

// ApplyTemplateResult defines the apply result for a single permission template (used in update action).
type ApplyTemplateResult struct {
	PermissionTemplateID string `json:"permission_template_id"`
	Status               string `json:"status"`
	Reason               string `json:"reason,omitempty"`
}

const (
	// ApplyStatusSuccess indicates the apply operation succeeded.
	ApplyStatusSuccess = "success"
	// ApplyStatusFailed indicates the apply operation failed.
	ApplyStatusFailed = "failed"
)

// ApplyPermissionPolicyLibraryResult defines the result of applying a permission policy library (create action).
type ApplyPermissionPolicyLibraryResult struct {
	Results []ApplyAccountResult `json:"results"`
}

// ApplyPermissionPolicyLibraryUpdateResult defines the result of applying a permission policy library (update action).
type ApplyPermissionPolicyLibraryUpdateResult struct {
	Results []ApplyTemplateResult `json:"results"`
}

// PermissionPolicyLibraryAccountIDsResult defines the result for account IDs query.
type PermissionPolicyLibraryAccountIDsResult struct {
	AccountIDs []string `json:"account_ids"`
}

// PermissionPolicyLibraryPermTmplResult defines the result for listing permission templates under a policy library.
type PermissionPolicyLibraryPermTmplResult struct {
	Details any `json:"details"`
}
