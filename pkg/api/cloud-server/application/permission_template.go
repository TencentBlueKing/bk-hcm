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

// BizCreatePermissionTemplateReq is the request to create an application for creating a
// permission template from a policy library for a single account.
type BizCreatePermissionTemplateReq struct {
	// AccountID is the target secondary account ID.
	AccountID string `json:"account_id" validate:"required"`
	// PolicyLibraryID is the ID of the permission policy library to use.
	PolicyLibraryID string `json:"policy_library_id" validate:"required"`
	// Name is the name of the cloud permission template to create.
	Name string `json:"name" validate:"required"`
	// Memo is an optional description.
	Memo *string `json:"memo"`
}

// Validate validates the request.
func (req *BizCreatePermissionTemplateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	return validator.ValidatePermTmplName(req.Name)
}

// BizUpdatePermissionTemplateReq is the request to create an application for updating a
// custom permission template to use a new policy library.
type BizUpdatePermissionTemplateReq struct {
	// ID is the ID of the existing permission template to update.
	ID string `json:"id" validate:"required"`
	// PolicyLibraryID is the new permission policy library ID to bind.
	PolicyLibraryID string `json:"policy_library_id" validate:"required"`
	// Memo is an optional description.
	Memo *string `json:"memo"`
}

// Validate validates the request.
func (req *BizUpdatePermissionTemplateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BizDeletePermissionTemplateReq is the request to create an application for deleting a
// custom permission template.
type BizDeletePermissionTemplateReq struct {
	// ID is the ID of the permission template to delete.
	ID string `json:"id" validate:"required"`
}

// Validate validates the request.
func (req *BizDeletePermissionTemplateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BasePermTemplateContent is the common header embedded in all permission template application content structs.
// Each action's content struct embeds this base and adds action-specific fields.
type BasePermTemplateContent struct {
	// Action distinguishes create/update/delete operations.
	Action enumor.OperatePermTemplateAction `json:"action"`
	// Vendor is the cloud vendor.
	Vendor enumor.Vendor `json:"vendor"`
	// BkBizID is the business ID from the request path.
	BkBizID int64 `json:"bk_biz_id"`
}
