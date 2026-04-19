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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// PermissionTemplateExtension permission template extension.
type PermissionTemplateExtension interface {
	TCloudPermissionTemplateExtension
}

// TCloudPermissionTemplateListExt defines TCloud-specific filter fields for biz-scoped permission template
// join list. Used in data-service request extension JSON and in DAO join filter options (same shape).
// CloudMainAccountIDs is resolved to account_ids in cloud-server and NOT forwarded to data-service.
// CloudSubAccountIDs matches sub_account.cloud_id (UIN string) and is forwarded to data-service.
type TCloudPermissionTemplateListExt struct {
	CloudMainAccountIDs []string `json:"cloud_main_account_ids" validate:"omitempty,max=500,dive,lte=255"`
	CloudSubAccountIDs  []string `json:"cloud_sub_account_ids" validate:"omitempty,max=500,dive,lte=255"`
}

// TCloudPermissionTemplateExtension defines tcloud permission template extension.
type TCloudPermissionTemplateExtension struct {
	CloudType enumor.TCloudPolicyType `json:"cloud_type"`
}

// BasePermissionTemplate defines base permission template (without extension).
type BasePermissionTemplate struct {
	ID                    string        `json:"id"`
	CloudID               string        `json:"cloud_id"`
	Name                  string        `json:"name"`
	AccountID             string        `json:"account_id"`
	PolicyLibraryID       *string       `json:"policy_library_id"`
	PolicyLibraryVersion  *int          `json:"policy_library_version"`
	PolicyLibrarySyncTime *string       `json:"policy_library_sync_time"`
	PolicyDocument        string        `json:"policy_document"`
	PolicyHash            string        `json:"policy_hash"`
	Memo                  *string       `json:"memo"`
	Vendor                enumor.Vendor `json:"vendor"`
	*core.Revision        `json:",inline"`
}

// GetID returns the local ID.
func (b BasePermissionTemplate) GetID() string {
	return b.ID
}

// GetCloudID returns the cloud ID.
func (b BasePermissionTemplate) GetCloudID() string {
	return b.CloudID
}

// PermissionTemplate defines permission template with typed extension.
type PermissionTemplate[T PermissionTemplateExtension] struct {
	BasePermissionTemplate `json:",inline"`
	Extension              *T `json:"extension"`
}

// PermissionTmplBasicInfo permission template basic info.
type PermissionTmplBasicInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
