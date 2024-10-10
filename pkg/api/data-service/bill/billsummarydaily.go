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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BillSummaryDailyCreateReq create request
type BillSummaryDailyCreateReq struct {
	RootAccountID      string              `json:"root_account_id" validate:"required"`
	MainAccountID      string              `json:"main_account_id" validate:"required"`
	RootAccountCloudID string              `json:"root_account_cloud_id" validate:"required"`
	MainAccountCloudID string              `json:"main_account_cloud_id" validate:"required"`
	Vendor             enumor.Vendor       `json:"vendor" validate:"required"`
	ProductID          int64               `json:"product_id" validate:"omitempty"`
	BkBizID            int64               `json:"bk_biz_id" validate:"omitempty"`
	BillYear           int                 `json:"bill_year" validate:"required"`
	BillMonth          int                 `json:"bill_month" validate:"required"`
	BillDay            int                 `json:"bill_day" validate:"required"`
	VersionID          int                 `json:"version_id" validate:"required"`
	Currency           enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	Cost               decimal.Decimal     `json:"cost" validate:"omitempty"`
	Count              int64               `json:"count" validate:"omitempty"`
}

// Validate ...
func (c *BillSummaryDailyCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillSummaryDailyListReq list request
type BillSummaryDailyListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *BillSummaryDailyListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillSummaryDailyListResult list result
type BillSummaryDailyListResult struct {
	Count   *uint64                 `json:"count,omitempty"`
	Details []billcore.SummaryDaily `json:"details"`
}

// BillSummaryDailyUpdateReq update request
type BillSummaryDailyUpdateReq struct {
	ID       string              `json:"id,omitempty" validate:"required"`
	Currency enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	Cost     *decimal.Decimal    `json:"cost" validate:"omitempty"`
	Count    int64               `json:"count" validate:"omitempty"`
}

// Validate ...
func (req *BillSummaryDailyUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
