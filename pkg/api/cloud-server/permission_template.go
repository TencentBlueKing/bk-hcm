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
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	tabletypes "hcm/pkg/dal/table/types"
)

// ListBizPermissionTemplateReq defines the cloud-server request for listing business permission templates.
type ListBizPermissionTemplateReq struct {
	IDs                    []string                       `json:"ids" validate:"omitempty,max=500"`
	CloudIDs               []string                       `json:"cloud_ids" validate:"omitempty,max=500"`
	PermissionTemplateType *enumor.PermissionTemplateType `json:"permission_template_type" validate:"omitempty"`
	Names                  []string                       `json:"names" validate:"omitempty,max=500"`
	Extension              tabletypes.JsonField           `json:"extension,omitempty"`
	Creator                string                         `json:"creator" validate:"omitempty,lte=64"`
	Reviser                string                         `json:"reviser" validate:"omitempty,lte=64"`
	Page                   *core.BasePage                 `json:"page" validate:"required"`
}

// Validate validates ListBizPermissionTemplateReq.
func (req *ListBizPermissionTemplateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return req.Page.Validate(core.NewDefaultPageOption())
}

// BizPermissionTemplateDetail defines one row in the biz-scoped permission template list response.
type BizPermissionTemplateDetail struct {
	corecloud.BasePermissionTemplate `json:",inline"`
	CloudAccountID                   string                                       `json:"cloud_account_id"`
	PolicyLibraryName                string                                       `json:"policy_library_name"`
	AssociatedSubAccountCount        int64                                        `json:"associated_sub_account_count"`
	Extension                        *corecloud.TCloudPermissionTemplateExtension `json:"extension"`
}

// BizPermissionTemplateListResult defines the biz-scoped permission template list response.
type BizPermissionTemplateListResult struct {
	Count   uint64                        `json:"count"`
	Details []BizPermissionTemplateDetail `json:"details"`
}

// PermTmplSubAccountIDsResult defines the response for listing sub account ids associated with a permission template.
type PermTmplSubAccountIDsResult struct {
	SubAccounts []PermRelateSubAccountInfo `json:"sub_accounts"`
}

// PermRelateSubAccountInfo defines the sub account info associated with a permission template.
type PermRelateSubAccountInfo struct {
	ID      string `json:"id"`
	CloudID string `json:"cloud_id"`
}
