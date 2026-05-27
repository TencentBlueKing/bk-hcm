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
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	"hcm/pkg/criteria/enumor"
	tablesubaccountsecret "hcm/pkg/dal/table/cloud/sub-account-secret"
	tabletypes "hcm/pkg/dal/table/types"
)

// ListSubAccountSecretDetails list sub account secret details.
type ListSubAccountSecretDetails struct {
	Count   uint64                        `json:"count,omitempty"`
	Details []tablesubaccountsecret.Table `json:"details,omitempty"`
}

// TCloudSubAccountSecretBizJoinExt is an alias of coresass.TCloudSubAccountSecretListExt (shared with
// data-service API extension JSON) for tcloud biz join list filters in the DAO.
type TCloudSubAccountSecretBizJoinExt = coresass.TCloudSubAccountSecretListExt

// ListSecretJoinAccountOption filters for biz-scoped join list on sub_account_secret.
// Extension holds vendor-specific filter fields; the DAO layer asserts the concrete type
// 即查询三级账号业务为BkBizID，也查询二级账号管理业务为BkBizID下的三级密钥
type ListSecretJoinAccountOption struct {
	Vendor             enumor.Vendor
	BkBizID            int64
	IDs                []string
	Status             []enumor.SubAccountSecretStatus
	AccountIDs         []string
	SubAccountIDs      []string
	AccountManagers    []string
	SubAccountManagers []string
	Page               *core.BasePage
	// Extension is vendor-specific JSON from upper layer; DAO parses by Vendor.
	Extension tabletypes.JsonField
}

// SubAccountSecretBizJoinRow is one row of sub_account_secret joined with sub_account and account.
type SubAccountSecretBizJoinRow struct {
	tablesubaccountsecret.Table `db:",inline"`
	AccountManagers             tabletypes.StringArray `db:"account_managers"`
	AccountName                 string                 `db:"account_name"`
	SubAccountManagers          tabletypes.StringArray `db:"sub_account_managers"`
	SubAccountName              string                 `db:"sub_account_name"`
	SubAccountExtensionJSON     tabletypes.JsonField   `db:"sub_account_extension"`
}

// ListSubAccountSecretBizJoinDetails is the join list result.
type ListSubAccountSecretBizJoinDetails struct {
	Count   uint64                       `json:"count,omitempty"`
	Details []SubAccountSecretBizJoinRow `json:"details,omitempty"`
}
