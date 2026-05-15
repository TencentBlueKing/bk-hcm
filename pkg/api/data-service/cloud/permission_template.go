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
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// PermissionTemplateBatchCreateReq permission template batch create request.
type PermissionTemplateBatchCreateReq[T corecloud.PermissionTemplateExtension] struct {
	PermissionTemplates []PermissionTemplateCreate[T] `json:"permission_templates" validate:"required,min=1"`
}

// Validate permission template batch create request.
func (req *PermissionTemplateBatchCreateReq[T]) Validate() error {
	if len(req.PermissionTemplates) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("permission_templates count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// PermissionTemplateCreate defines a single item in batch create request.
type PermissionTemplateCreate[T corecloud.PermissionTemplateExtension] struct {
	CloudID               string  `json:"cloud_id" validate:"required,max=64"`
	Name                  string  `json:"name" validate:"required,max=128"`
	AccountID             string  `json:"account_id" validate:"required,max=64"`
	PolicyLibraryID       *string `json:"policy_library_id" validate:"omitempty,max=64"`
	PolicyLibraryVersion  *int    `json:"policy_library_version" validate:"omitempty"`
	PolicyLibrarySyncTime *string `json:"policy_library_sync_time" validate:"omitempty"`
	PolicyDocument        string  `json:"policy_document" validate:"required"`
	Memo                  *string `json:"memo" validate:"omitempty,max=255"`
	Extension             *T      `json:"extension" validate:"required"`
}

// -------------------------- Update --------------------------

// PermissionTemplateBatchUpdateReq permission template batch update request.
type PermissionTemplateBatchUpdateReq[T corecloud.PermissionTemplateExtension] struct {
	PermissionTemplates []PermissionTemplateUpdate[T] `json:"permission_templates" validate:"required,min=1,max=100"`
}

// Validate permission template batch update request.
func (req *PermissionTemplateBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionTemplateUpdate defines a single item in batch update request.
type PermissionTemplateUpdate[T corecloud.PermissionTemplateExtension] struct {
	ID                    string  `json:"id" validate:"required"`
	Name                  string  `json:"name" validate:"omitempty,max=128"`
	PolicyLibraryID       *string `json:"policy_library_id" validate:"omitempty,max=64"`
	PolicyLibraryVersion  *int    `json:"policy_library_version" validate:"omitempty"`
	PolicyLibrarySyncTime *string `json:"policy_library_sync_time" validate:"omitempty"`
	PolicyDocument        string  `json:"policy_document" validate:"omitempty"`
	Memo                  *string `json:"memo" validate:"omitempty,max=255"`
	Extension             *T      `json:"extension" validate:"omitempty"`
}

// -------------------------- Delete --------------------------

// PermissionTemplateBatchDeleteReq permission template batch delete request.
type PermissionTemplateBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate permission template batch delete request.
func (req *PermissionTemplateBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List (without extension) --------------------------

// PermissionTemplateListReq permission template list request.
type PermissionTemplateListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate permission template list request.
func (req *PermissionTemplateListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionTemplateListResult permission template list result (without extension).
type PermissionTemplateListResult struct {
	Count   uint64                             `json:"count"`
	Details []corecloud.BasePermissionTemplate `json:"details"`
}

// PermissionTemplateListResp permission template list response.
type PermissionTemplateListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *PermissionTemplateListResult `json:"data"`
}

// -------------------------- List with extension --------------------------

// PermissionTemplateExtListReq permission template ext list request.
type PermissionTemplateExtListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate permission template ext list request.
func (req *PermissionTemplateExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PermissionTemplateExtListResult permission template list result with extension.
type PermissionTemplateExtListResult[T corecloud.PermissionTemplateExtension] struct {
	Count   uint64                            `json:"count,omitempty"`
	Details []corecloud.PermissionTemplate[T] `json:"details,omitempty"`
}

// PermissionTemplateExtListResp permission template ext list response.
type PermissionTemplateExtListResp[T corecloud.PermissionTemplateExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *PermissionTemplateExtListResult[T] `json:"data"`
}

// -------------------------- Join List (permission_template + sub_account) --------------------------

// PermissionTemplateFilters defines biz-scoped list filters; style aligned with SubAccountSecretFilters.
type PermissionTemplateFilters struct {
	IDs        []string             `json:"ids" validate:"omitempty,max=500"`
	CloudIDs   []string             `json:"cloud_ids" validate:"omitempty,max=500"`
	Names      []string             `json:"names" validate:"omitempty,max=500"`
	AccountIDs []string             `json:"account_ids" validate:"omitempty,max=500,dive,lte=64"`
	Creator    string               `json:"creator" validate:"omitempty,lte=64"`
	Reviser    string               `json:"reviser" validate:"omitempty,lte=64"`
	Extension  tabletypes.JsonField `json:"extension,omitempty"`
	// PolicyLibraryIDIsNull filters by policy_library_id nullability:
	//   nil   → no filter    true  → IS NULL   false → IS NOT NULL
	PolicyLibraryIDIsNull *bool `json:"policy_library_id_is_null,omitempty"`
}

// Validate validates PermissionTemplateFilters.
func (f *PermissionTemplateFilters) Validate() error {
	return validator.Validate.Struct(f)
}

// PermissionTmplJoinExtListReq defines permission template join-list request for data-service
type PermissionTmplJoinExtListReq struct {
	PermissionTemplateFilters `json:",inline"`
	BkBizID                   int64          `json:"bk_biz_id" validate:"required"`
	Page                      *core.BasePage `json:"page" validate:"required"`
}

// Validate validates PermissionTmplJoinExtListReq.
func (req *PermissionTmplJoinExtListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.PermissionTemplateFilters.Validate(); err != nil {
		return err
	}

	return req.Page.Validate(core.NewDefaultPageOption())
}

// PermissionTmplJoinExtDetail is one row in permission_template join list response.
type PermissionTmplJoinExtDetail struct {
	corecloud.BasePermissionTemplate `json:",inline"`
	AssociatedSubAccountCount        int64                                        `json:"associated_sub_account_count"`
	Extension                        *corecloud.TCloudPermissionTemplateExtension `json:"extension"`
}

// PermissionTmplJoinExtListResult defines permission template join list response.
type PermissionTmplJoinExtListResult struct {
	Count   uint64                        `json:"count"`
	Details []PermissionTmplJoinExtDetail `json:"details"`
}

// PermissionTmplJoinListResp defines permission template join list HTTP response.
type PermissionTmplJoinListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *PermissionTmplJoinExtListResult `json:"data"`
}
