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
	RootAccountID   string          `json:"root_account_id" validate:"required"`
	RootAccountName string          `json:"root_account_name" validate:"required"`
	Vendor          enumor.Vendor   `json:"vendor" validate:"required"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	VersionID       string          `json:"version_id" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	Rate            float64         `json:"rate" validate:"required"`
	State           string          `json:"state" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
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
	ID              string          `json:"id,omitempty"`
	RootAccountID   string          `json:"root_account_id" validate:"required"`
	RootAccountName string          `json:"root_account_name" validate:"required"`
	Vendor          enumor.Vendor   `json:"vendor" validate:"required"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	VersionID       string          `json:"version_id" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	Rate            float64         `json:"rate" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
	State           string          `json:"state" validate:"required"`
	CreatedAt       types.Time      `json:"created_at,omitempty"`
	UpdatedAt       types.Time      `json:"updated_at,omitempty"`
}

// BillSummaryRootUpdateReq update request
type BillSummaryRootUpdateReq struct {
	ID              string          `json:"id,omitempty" validate:"required"`
	RootAccountID   string          `json:"root_account_id" validate:"required"`
	RootAccountName string          `json:"root_account_name" validate:"required"`
	Vendor          enumor.Vendor   `json:"vendor" validate:"required"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	VersionID       string          `json:"version_id" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	Rate            float64         `json:"rate" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
	State           string          `json:"state" validate:"required"`
}

// Validate ...
func (req *BillSummaryRootUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
