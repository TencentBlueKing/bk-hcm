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
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BillMonthTaskCreateReq create request
type BillMonthTaskCreateReq struct {
	RootAccountID      string                               `json:"root_account_id" validate:"required"`
	RootAccountCloudID string                               `json:"root_account_cloud_id" validate:"required"`
	Vendor             enumor.Vendor                        `json:"vendor" validate:"required"`
	Type               enumor.MonthTaskType                 `json:"type" validate:"required"`
	BillYear           int                                  `json:"bill_year" validate:"required"`
	BillMonth          int                                  `json:"bill_month" validate:"required"`
	VersionID          int                                  `json:"version_id" validate:"required"`
	State              enumor.RootAccountMonthBillTaskState `json:"state" validate:"required"`
}

// Validate AccountBillConfigBatchCreateReq.
func (c *BillMonthTaskCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillMonthTaskListReq list request
type BillMonthTaskListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *BillMonthTaskListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillMonthTaskListResult list result
type BillMonthTaskListResult struct {
	Count   *uint64           `json:"count,omitempty"`
	Details []*bill.MonthTask `json:"details"`
}

// BillMonthTaskUpdateReq update request
type BillMonthTaskUpdateReq struct {
	ID            string                               `json:"id,omitempty" validate:"required"`
	State         enumor.RootAccountMonthBillTaskState `json:"state,omitempty"`
	Count         uint64                               `json:"count,omitempty"`
	Currency      enumor.CurrencyCode                  `json:"currency,omitempty"`
	Cost          *decimal.Decimal                     `json:"cost,omitempty"`
	PullIndex     uint64                               `json:"pull_index,omitempty"`
	PullFlowID    string                               `json:"pull_flow_id,omitempty"`
	SplitIndex    uint64                               `json:"split_index,omitempty"`
	SplitFlowID   string                               `json:"split_flow_id,omitempty"`
	SummaryFlowID string                               `json:"summary_flow_id,omitempty"`
	// 覆盖更新
	SummaryDetail []bill.MonthTaskSummaryDetailItem `json:"summary_detail,omitempty"`
}

// Validate ...
func (req *BillMonthTaskUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
