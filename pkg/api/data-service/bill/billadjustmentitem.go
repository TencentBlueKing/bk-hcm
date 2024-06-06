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
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BatchBillAdjustmentItemCreateReq batch create request
type BatchBillAdjustmentItemCreateReq []BillAdjustmentItemCreateReq

// BillAdjustmentItemCreateReq create request
type BillAdjustmentItemCreateReq struct {
	FirstAccountID  string          `json:"first_account_id" validate:"required"`
	SecondAccountID string          `json:"second_account_id" validate:"required"`
	ProductID       int64           `json:"product_id" validate:"omitempty"`
	BkBizID         int64           `json:"bk_biz_id" validate:"omitempty"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	BillDay         int             `json:"bill_day" validate:"required"`
	Type            string          `json:"type" validate:"required"`
	Memo            string          `json:"memo"`
	Operator        string          `json:"operator" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
	State           string          `json:"state" validate:"required"`
}

// Validate ...
func (c *BillAdjustmentItemCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillAdjustmentItemListReq list request
type BillAdjustmentItemListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *BillAdjustmentItemListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillAdjustmentItemListResult list result
type BillAdjustmentItemListResult struct {
	Count   *uint64                     `json:"count,omitempty"`
	Details []*BillAdjustmentItemResult `json:"details"`
}

// BillAdjustmentItemResult result
type BillAdjustmentItemResult struct {
	ID              string          `json:"id,omitempty"`
	FirstAccountID  string          `json:"first_account_id" validate:"required"`
	SecondAccountID string          `json:"second_account_id" validate:"required"`
	ProductID       int64           `json:"product_id" validate:"omitempty"`
	BkBizID         int64           `json:"bk_biz_id" validate:"omitempty"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	BillDay         int             `json:"bill_day" validate:"required"`
	Type            string          `json:"type" validate:"required"`
	Memo            string          `json:"memo"`
	Operator        string          `json:"operator" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
	State           string          `json:"state" validate:"required"`
	CreatedAt       types.Time      `json:"created_at,omitempty"`
	UpdatedAt       types.Time      `json:"updated_at,omitempty"`
}

// BillAdjustmentItemUpdateReq update request
type BillAdjustmentItemUpdateReq struct {
	ID              string          `json:"id,omitempty" validate:"required"`
	FirstAccountID  string          `json:"first_account_id" validate:"required"`
	SecondAccountID string          `json:"second_account_id" validate:"required"`
	ProductID       int64           `json:"product_id" validate:"omitempty"`
	BkBizID         int64           `json:"bk_biz_id" validate:"omitempty"`
	BillYear        int             `json:"bill_year" validate:"required"`
	BillMonth       int             `json:"bill_month" validate:"required"`
	BillDay         int             `json:"bill_day" validate:"required"`
	Type            string          `json:"type" validate:"required"`
	Memo            string          `json:"memo"`
	Operator        string          `json:"operator" validate:"required"`
	Currency        string          `json:"currency" validate:"required"`
	Cost            decimal.Decimal `json:"cost" validate:"required"`
	RMBCost         decimal.Decimal `json:"rmb_cost" validate:"required"`
	State           string          `json:"state" validate:"required"`
}

// Validate ...
func (req *BillAdjustmentItemUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
