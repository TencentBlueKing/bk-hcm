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

package bill

import (
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"

	"github.com/shopspring/decimal"
)

// -------------------------- List --------------------------

// AwsBillListResult define aws bill list result.
type AwsBillListResult struct {
	Count   int64 `json:"count"`
	Details any   `json:"details"`
}

// AwsBillListResp define aws bill list resp.
type AwsBillListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AwsBillListResult `json:"data"`
}

// AwsRootSpUsageTotalReq ...
type AwsRootSpUsageTotalReq struct {
	// 根账号id
	RootAccountID string `json:"root_account_id" validate:"required"`
	// 筛选使用账号云id，为空则不筛选
	SpUsageAccountCloudIds []string `json:"sp_usage_account_cloud_ids" `
	SpArnPrefix            string   `json:"sp_arn_prefix" validate:"omitempty"`

	Year  uint `json:"year" validate:"required"`
	Month uint `json:"month" validate:"required,min=1,max=12"`
	// 起始日
	StartDay uint `json:"start_day" validate:"required,min=1,max=31"`
	// 截止日
	EndDay uint `json:"end_day" validate:"required,min=1,max=31"`
}

// Validate ...
func (r *AwsRootSpUsageTotalReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AwsSpUsageTotalResult ...
type AwsSpUsageTotalResult struct {
	SPCost        *decimal.Decimal `json:"sp_cost"`
	UnblendedCost *decimal.Decimal `json:"unblended_cost"`
	SPNetCost     *decimal.Decimal `json:"sp_net_cost"`
	AccountCount  uint64           `json:"account_count"`
}
