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
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BillDailyPullTaskCreateReq BillDailyPullTask create request
type BillDailyPullTaskCreateReq struct {
	RootAccountID      string          `json:"root_account_id" validate:"required"`
	MainAccountID      string          `json:"main_account_id" validate:"required"`
	Vendor             enumor.Vendor   `json:"vendor" validate:"required"`
	ProductID          int64           `json:"product_id" validate:"omitempty"`
	BkBizID            int64           `json:"bk_biz_id" validate:"omitempty"`
	BillYear           int             `json:"bill_year" validate:"required"`
	BillMonth          int             `json:"bill_month" validate:"required"`
	BillDay            int             `json:"bill_day" validate:"required"`
	VersionID          int             `json:"version_id" validate:"required"`
	State              string          `json:"state" vaildate:"required"`
	Count              int64           `json:"count" validate:"omitempty"`
	Currency           string          `json:"currency" validate:"omitempty"`
	Cost               decimal.Decimal `json:"cost" validate:"omitempty"`
	FlowID             string          `json:"flow_id" validate:"omitempty"`
	SplitFlowID        string          `json:"split_flow_id" validate:"omitempty"`
	DailySummaryFlowID string          `json:"daily_summary_flow_id" validate:"omitempty"`
}

// Validate validates BillDailyPullTaskCreateReq
func (c *BillDailyPullTaskCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillDailyPullTaskListReq BillDailyPullTask list request
type BillDailyPullTaskListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate validates BillDailyPullTaskListReq
func (req *BillDailyPullTaskListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillDailyPullTaskListResult BillDailyPullTask list response
type BillDailyPullTaskListResult struct {
	Count   *uint64                    `json:"count,omitempty"`
	Details []*BillDailyPullTaskResult `json:"details"`
}

// BillDailyPullTaskResult BillDailyPullTask result
type BillDailyPullTaskResult struct {
	ID                 string           `json:"id,omitempty"`
	RootAccountID      string           `json:"root_account_id" validate:"required"`
	MainAccountID      string           `json:"main_account_id" validate:"required"`
	Vendor             enumor.Vendor    `json:"vendor" validate:"required"`
	ProductID          int64            `json:"product_id" validate:"omitempty"`
	BkBizID            int64            `json:"bk_biz_id" validate:"omitempty"`
	BillYear           int              `json:"bill_year" validate:"required"`
	BillMonth          int              `json:"bill_month" validate:"required"`
	BillDay            int              `json:"bill_day" validate:"required"`
	VersionID          int              `json:"version_id" validate:"required"`
	State              string           `json:"state" vaildate:"required"`
	Count              int64            `json:"count" validate:"omitempty"`
	Currency           string           `json:"currency" validate:"omitempty"`
	Cost               *decimal.Decimal `json:"cost" validate:"omitempty"`
	FlowID             string           `json:"flow_id" validate:"omitempty"`
	SplitFlowID        string           `json:"split_flow_id" validate:"omitempty"`
	DailySummaryFlowID string           `json:"daily_summary_flow_id" validate:"omitempty"`
	CreatedAt          types.Time       `json:"created_at,omitempty"`
	UpdatedAt          types.Time       `json:"updated_at,omitempty"`
}

// Key get key
func (b *BillDailyPullTaskResult) Key() string {
	return fmt.Sprintf("%s/%s/%s/%d/%d/%d/%d",
		b.RootAccountID, b.MainAccountID, b.Vendor, b.BillYear, b.BillMonth, b.BillDay, b.VersionID)
}

// BillDailyPullTaskUpdateReq ...
type BillDailyPullTaskUpdateReq struct {
	ID                 string              `json:"id,omitempty" validate:"required"`
	RootAccountID      string              `json:"root_account_id" validate:"omitempty"`
	MainAccountID      string              `json:"main_account_id" validate:"omitempty"`
	Vendor             enumor.Vendor       `json:"vendor" validate:"omitempty"`
	ProductID          int64               `json:"product_id" validate:"omitempty"`
	BkBizID            int64               `json:"bk_biz_id" validate:"omitempty"`
	BillYear           int                 `json:"bill_year" validate:"omitempty"`
	BillMonth          int                 `json:"bill_month" validate:"omitempty"`
	BillDay            int                 `json:"bill_day" validate:"omitempty"`
	VersionID          int                 `json:"version_id" validate:"omitempty"`
	State              string              `json:"state" vaildate:"omitempty"`
	Count              int64               `json:"count" validate:"omitempty"`
	Currency           enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	Cost               decimal.Decimal     `json:"cost" validate:"omitempty"`
	FlowID             string              `json:"flow_id" validate:"omitempty"`
	SplitFlowID        string              `json:"split_flow_id" validate:"omitempty"`
	DailySummaryFlowID string              `json:"daily_summary_flow_id" validate:"omitempty"`
}

// Validate validates BillDailyPullTaskUpdateReq
func (req *BillDailyPullTaskUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
