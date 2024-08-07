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
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// MainAccountSummaryListReq list request for main account summary
type MainAccountSummaryListReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"omitempty"`
	Page      *core.BasePage     `json:"page" validate:"omitempty"`
}

// Validate ...
func (req *MainAccountSummaryListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// MainAccountSummarySumReq get request for all account summary
type MainAccountSummarySumReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"omitempty"`
}

// Validate ...
func (req *MainAccountSummarySumReq) Validate() error {
	return validator.Validate.Struct(req)
}

// MainAccountSummarySumResult all root account summary get result
type MainAccountSummarySumResult struct {
	Count   uint64                                             `json:"count"`
	CostMap map[enumor.CurrencyCode]*billcore.CostWithCurrency `json:"cost_map"`
}

// MainAccountSummaryListResult main account summary list result
type MainAccountSummaryListResult struct {
	Count   uint64                      `json:"count,omitempty"`
	Details []*MainAccountSummaryResult `json:"details"`
}

// MainAccountSummaryResult main account summary get result
type MainAccountSummaryResult struct {
	bill.BillSummaryMainResult
	MainAccountCloudID   string `json:"main_account_cloud_id" validate:"required"`
	MainAccountCloudName string `json:"main_account_cloud_name" validate:"required"`
}
