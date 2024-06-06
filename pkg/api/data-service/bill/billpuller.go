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
)

// BillPullerCreateReq create request
type BillPullerCreateReq struct {
	FirstAccountID        string                    `json:"first_account_id" validate:"required"`
	SecondAccountID       string                    `json:"second_account_id" validate:"required"`
	Vendor                enumor.Vendor             `json:"vendor" validate:"required"`
	ProductID             int64                     `json:"product_id" validate:"omitempty"`
	BkBizID               int64                     `json:"bk_biz_id" validate:"omitempty"`
	PullMode              enumor.BillPullMode       `json:"pull_mode" validate:"required"`
	SyncPeriod            enumor.BillSyncPeriodType `json:"sync_period" validate:"required"`
	BillDelay             string                    `json:"bill_delay" validate:"required"`
	FinalBillCalendarDate enumor.BillDayNumber      `json:"final_bill_calendar_date" validate:"required"`
}

// Validate AccountBillConfigBatchCreateReq.
func (c *BillPullerCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillPullerListReq list request
type BillPullerListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *BillPullerListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillPullerListResult list result
type BillPullerListResult struct {
	Count   *uint64             `json:"count,omitempty"`
	Details []*BillPullerResult `json:"details"`
}

// BillPullerResult result
type BillPullerResult struct {
	ID                    string        `json:"id,omitempty"`
	FirstAccountID        string        `json:"first_account_id,omitempty"`
	SecondAccountID       string        `json:"second_account_id,omitempty"`
	Vendor                enumor.Vendor `json:"vendor,omitempty"`
	ProductID             int64         `json:"product_id,omitempty"`
	BkBizID               int64         `json:"bk_biz_id,omitempty"`
	PullMode              string        `json:"pull_mode,omitempty"`
	SyncPeriod            string        `json:"sync_period,omitempty"`
	BillDelay             string        `json:"bill_delay,omitempty"`
	FinalBillCalendarDate int           `json:"final_bill_calendar_date,omitempty"`
	CreatedAt             types.Time    `json:"created_at,omitempty"`
	UpdatedAt             types.Time    `json:"updated_at,omitempty"`
}

// BillPullerUpdateReq update request
type BillPullerUpdateReq struct {
	ID                    string        `json:"id,omitempty" validate:"required"`
	FirstAccountID        string        `json:"first_account_id,omitempty" validate:"required"`
	SecondAccountID       string        `json:"second_account_id,omitempty" validate:"required"`
	Vendor                enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	ProductID             int64         `json:"product_id,omitempty"`
	BkBizID               int64         `json:"bk_biz_id,omitempty"`
	PullMode              string        `json:"pull_mode,omitempty" validate:"required"`
	SyncPeriod            string        `json:"sync_period,omitempty" validate:"required"`
	BillDelay             string        `json:"bill_delay,omitempty" validate:"required"`
	FinalBillCalendarDate int           `json:"final_bill_calendar_date,omitempty"`
}

// Validate ...
func (req *BillPullerUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
