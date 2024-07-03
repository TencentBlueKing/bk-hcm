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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BillSummaryRootCreateReq create request
type BillSummaryRootCreateReq struct {
	RootAccountID             string              `json:"root_account_id" validate:"required"`
	RootAccountName           string              `json:"root_account_name" validate:"omitempty"`
	Vendor                    enumor.Vendor       `json:"vendor" validate:"required"`
	BillYear                  int                 `json:"bill_year" validate:"required"`
	BillMonth                 int                 `json:"bill_month" validate:"required"`
	LastSyncedVersion         int                 `json:"last_synced_version" validate:"omitempty"`
	CurrentVersion            int                 `json:"current_version" validate:"required"`
	Currency                  enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	LastMonthCostSynced       decimal.Decimal     `json:"last_month_cost_synced" validate:"omitempty"`
	LastMonthRMBCostSynced    decimal.Decimal     `json:"last_month_rmb_cost_synced" validate:"omitempty"`
	CurrentMonthCostSynced    decimal.Decimal     `json:"current_month_cost_synced" validate:"omitempty"`
	CurrentMonthRMBCostSynced decimal.Decimal     `json:"current_month_rmb_cost_synced" validate:"omitempty"`
	MonthOnMonthValue         float64             `json:"month_on_month_value" validate:"omitempty"`
	CurrentMonthCost          decimal.Decimal     `json:"current_month_cost" validate:"omitempty"`
	CurrentMonthRMBCost       decimal.Decimal     `json:"current_month_rmb_cost" validate:"omitempty"`
	AjustmentCost             decimal.Decimal     `json:"adjustment_cost" validate:"omitempty"`
	AjustmentRMBCost          decimal.Decimal     `json:"adjustment_rmb_cost" validate:"omitempty"`
	Rate                      float64             `json:"rate" validate:"omitempty"`
	BkBizNum                  uint64              `json:"bk_biz_num" validate:"omitempty"`
	ProductNum                uint64              `json:"product_num" vaidate:"omitempty"`
	State                     string              `json:"state" validate:"omitempty"`
}

// Validate ...
func (c *BillSummaryRootCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillSummaryRootListReq list request
type BillSummaryRootListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *BillSummaryRootListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillSummaryRootListResult list result
type BillSummaryRootListResult struct {
	Count   *uint64                  `json:"count,omitempty"`
	Details []*BillSummaryRootResult `json:"details"`
}

// BillSummaryRootResult result
type BillSummaryRootResult struct {
	ID                        string              `json:"id,omitempty"`
	RootAccountID             string              `json:"root_account_id" validate:"required"`
	RootAccountName           string              `json:"root_account_name" validate:"required"`
	Vendor                    enumor.Vendor       `json:"vendor" validate:"required"`
	BillYear                  int                 `json:"bill_year" validate:"required"`
	BillMonth                 int                 `json:"bill_month" validate:"required"`
	LastSyncedVersion         int                 `json:"last_synced_version" validate:"omitempty"`
	CurrentVersion            int                 `json:"current_version" validate:"required"`
	Currency                  enumor.CurrencyCode `json:"currency" validate:"required"`
	LastMonthCostSynced       decimal.Decimal     `json:"last_month_cost_synced" validate:"omitempty"`
	LastMonthRMBCostSynced    decimal.Decimal     `json:"last_month_rmb_cost_synced" validate:"omitempty"`
	CurrentMonthCostSynced    decimal.Decimal     `json:"current_month_cost_synced" validate:"omitempty"`
	CurrentMonthRMBCostSynced decimal.Decimal     `json:"current_month_rmb_cost_synced" validate:"omitempty"`
	MonthOnMonthValue         float64             `json:"month_on_month_value" validate:"omitempty"`
	CurrentMonthCost          decimal.Decimal     `json:"current_month_cost" validate:"omitempty"`
	CurrentMonthRMBCost       decimal.Decimal     `json:"current_month_rmb_cost" validate:"omitempty"`
	AjustmentCost             decimal.Decimal     `json:"adjustment_cost" validate:"omitempty"`
	AjustmentRMBCost          decimal.Decimal     `json:"adjustment_rmb_cost" validate:"omitempty"`
	Rate                      float64             `json:"rate" validate:"required"`
	BkBizNum                  uint64              `json:"bk_biz_num" validate:"omitempty"`
	ProductNum                uint64              `json:"product_num" vaidate:"omitempty"`
	State                     string              `json:"state" validate:"required"`
	CreatedAt                 types.Time          `json:"created_at,omitempty"`
	UpdatedAt                 types.Time          `json:"updated_at,omitempty"`
}

// BillSummaryRootUpdateReq update request
type BillSummaryRootUpdateReq struct {
	ID                        string              `json:"id,omitempty" validate:"required"`
	RootAccountName           string              `json:"root_account_name" validate:"omitempty"`
	LastSyncedVersion         int                 `json:"last_synced_version" validate:"omitempty"`
	CurrentVersion            int                 `json:"current_version" validate:"omitempty"`
	Currency                  enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	LastMonthCostSynced       *decimal.Decimal    `json:"last_month_cost_synced" validate:"omitempty"`
	LastMonthRMBCostSynced    *decimal.Decimal    `json:"last_month_rmb_cost_synced" validate:"omitempty"`
	CurrentMonthCostSynced    *decimal.Decimal    `json:"current_month_cost_synced" validate:"omitempty"`
	CurrentMonthRMBCostSynced *decimal.Decimal    `json:"current_month_rmb_cost_synced" validate:"omitempty"`
	MonthOnMonthValue         float64             `json:"month_on_month_value" validate:"omitempty"`
	CurrentMonthCost          *decimal.Decimal    `json:"current_month_cost" validate:"omitempty"`
	CurrentMonthRMBCost       *decimal.Decimal    `json:"current_month_rmb_cost" validate:"omitempty"`
	AjustmentCost             *decimal.Decimal    `json:"adjustment_cost" validate:"omitempty"`
	AjustmentRMBCost          *decimal.Decimal    `json:"adjustment_rmb_cost" validate:"omitempty"`
	Rate                      float64             `json:"rate" validate:"omitempty"`
	BkBizNum                  uint64              `json:"bk_biz_num" validate:"omitempty"`
	ProductNum                uint64              `json:"product_num" vaidate:"omitempty"`
	State                     string              `json:"state" validate:"omitempty"`
}

// Validate ...
func (req *BillSummaryRootUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillSummaryBatchSyncReq batch sync request
type BillSummaryBatchSyncReq struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	BillYear  int           `json:"bill_year" validate:"required"`
	BillMonth int           `json:"bill_month" validate:"required"`
}

// Validate ...
func (req *BillSummaryBatchSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}
