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

package types

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletypes "hcm/pkg/dal/table/types"
)

// ListPermissionTemplateDetails list permission template details.
type ListPermissionTemplateDetails struct {
	Count   uint64                               `json:"count,omitempty"`
	Details []tablecloud.PermissionTemplateTable `json:"details,omitempty"`
}

// TCloudPermTmplJoinExt is an alias of corecloud.TCloudPermissionTemplateListExt (shared with
// data-service API extension JSON) for tcloud biz join list filters in the DAO.
type TCloudPermTmplJoinExt = corecloud.TCloudPermissionTemplateListExt

// ListPermTmplJoinOption defines filters for biz-scoped permission_template join sub_account list.
type ListPermTmplJoinOption struct {
	Vendor     enumor.Vendor
	BkBizID    int64
	IDs        []string
	CloudIDs   []string
	Names      []string
	AccountIDs []string
	Creator    string
	Reviser    string
	Page       *core.BasePage
	// Extension is vendor-specific JSON from upper layer; DAO parses for TCloud.
	Extension tabletypes.JsonField
}

// PermissionTmplJoinRow is one row of permission_template with associated sub_account count
// computed via subquery.
type PermissionTmplJoinRow struct {
	tablecloud.PermissionTemplateTable `db:",inline"`
	AssociatedSubAccountCount          int64 `db:"associated_sub_account_count"`
}

// ListPermissionTmplJoinDetails is the join list result.
type ListPermissionTmplJoinDetails struct {
	Count   uint64                  `json:"count,omitempty"`
	Details []PermissionTmplJoinRow `json:"details,omitempty"`
}
