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

package application

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// BizApplyPermissionPolicyLibraryCreateReq is the request to create an application for applying a permission
// policy library (create action) to multiple accounts.
type BizApplyPermissionPolicyLibraryCreateReq struct {
	// PolicyLibraryID is the ID of the permission policy library to apply.
	PolicyLibraryID string `json:"policy_library_id" validate:"required"`
	// AccountIDs is the list of account IDs to apply the library to.
	AccountIDs []string `json:"account_ids" validate:"required,min=1,max=100"`
}

// Validate validates the request.
func (req *BizApplyPermissionPolicyLibraryCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BizApplyPermissionPolicyLibraryUpdateReq is the request to create an application for applying a
// permission policy library (update action) to multiple permission templates.
type BizApplyPermissionPolicyLibraryUpdateReq struct {
	// PolicyLibraryID is the ID of the permission policy library to apply.
	PolicyLibraryID string `json:"policy_library_id" validate:"required"`
	// PermissionTemplateIDs is the list of permission template IDs to update.
	PermissionTemplateIDs []string `json:"permission_template_ids" validate:"required,min=1,max=100"`
}

// Validate validates the request.
func (req *BizApplyPermissionPolicyLibraryUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApplyPermPolicyLibBaseContent is the common header embedded in all apply_permission_policy_library
// application content structs. Used for action dispatch in NewHandlerFromApplication.
type ApplyPermPolicyLibBaseContent struct {
	// Action distinguishes create from update operations.
	Action    enumor.PermPolicyLibAction  `json:"action"`
	Operation enumor.ApplicationOperation `json:"operation"`
	// Vendor is the cloud vendor.
	Vendor enumor.Vendor `json:"vendor"`
	// BkBizID is the business ID from the request path.
	BkBizID int64 `json:"bk_biz_id"`
	// PolicyLibraryID is the permission policy library ID.
	PolicyLibraryID string `json:"policy_library_id"`
}

// ApplyPermPolicyLibCreateContent is the content stored in the application record for
// the apply_permission_policy_library (create action) type.
type ApplyPermPolicyLibCreateContent struct {
	ApplyPermPolicyLibBaseContent
	// AccountID is the single account ID for this application record.
	AccountID string `json:"account_id"`
}

// ApplyPermPolicyLibUpdateContent is the content stored in the application record for
// the apply_permission_policy_library (update action) type.
type ApplyPermPolicyLibUpdateContent struct {
	ApplyPermPolicyLibBaseContent
	// PermissionTemplateID is the single permission template ID for this application record.
	PermissionTemplateID string `json:"permission_template_id"`
}
