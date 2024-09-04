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
	"errors"

	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/runtime/filter"
)

// RootAccountSummaryListReq list request for root account summary
type RootAccountSummaryListReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"required"`
	Page      *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *RootAccountSummaryListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RootAccountSummaryReaccountReq reaccount request for root account summary
type RootAccountSummaryReaccountReq struct {
	BillYear      int    `json:"bill_year" validate:"required"`
	BillMonth     int    `json:"bill_month" validate:"required"`
	RootAccountID string `json:"root_account_id" validate:"required"`
}

// Validate ...
func (req *RootAccountSummaryReaccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RootAccountSummaryConfirmReq confirm request for root account summary
type RootAccountSummaryConfirmReq struct {
	BillYear      int    `json:"bill_year" validate:"required"`
	BillMonth     int    `json:"bill_month" validate:"required"`
	RootAccountID string `json:"root_account_id" validate:"required"`
}

// Validate ...
func (req *RootAccountSummaryConfirmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RootAccountSummarySumReq get request for all account summary
type RootAccountSummarySumReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"omitempty"`
}

// Validate ...
func (req *RootAccountSummarySumReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RootAccountSummarySumResult all root account summary get result
type RootAccountSummarySumResult struct {
	Count   uint64                                             `json:"count"`
	CostMap map[enumor.CurrencyCode]*billcore.CostWithCurrency `json:"cost_map"`
}

// BillSummaryRootResult ...
type BillSummaryRootResult struct {
	*bill.BillSummaryRootResult
	RootAccountName string `json:"root_account_name" `
}

// BillSummaryRootListResult ...
type BillSummaryRootListResult = core.ListResultT[BillSummaryRootResult]

// RootAccountSummaryExportReq ...
type RootAccountSummaryExportReq struct {
	BillYear    int                `json:"bill_year" validate:"required"`
	BillMonth   int                `json:"bill_month" validate:"required"`
	ExportLimit uint64             `json:"export_limit" validate:"required"`
	Filter      *filter.Expression `json:"filter" validate:"omitempty"`
}

// Validate ...
func (r *RootAccountSummaryExportReq) Validate() error {
	if r.ExportLimit > constant.ExcelExportLimit {
		return errors.New("export limit exceed")
	}
	if r.Filter != nil {
		err := r.Filter.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillItemColumns.ColumnTypes())))
		if err != nil {
			return err
		}
	}
	if r.BillYear == 0 {
		return errors.New("year is required")
	}
	if r.BillMonth == 0 {
		return errors.New("month is required")
	}
	if r.BillMonth > 12 || r.BillMonth < 0 {
		return errors.New("month must between 1 and 12")
	}
	return validator.Validate.Struct(r)
}
