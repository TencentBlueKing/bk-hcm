/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"

	"github.com/shopspring/decimal"
)

// MonthTaskSummaryDetailItem detail item of month task summary
type MonthTaskSummaryDetailItem struct {
	MainAccountID string              `json:"main_account_id"`
	IsFinished    bool                `json:"is_finished"`
	Currency      enumor.CurrencyCode `json:"currency"`
	Cost          decimal.Decimal     `json:"cost"`
	Count         uint64              `json:"count"`
}

// MonthTask result
type MonthTask struct {
	ID                 string                               `json:"id"`
	RootAccountID      string                               `json:"root_account_id,omitempty"`
	RootAccountCloudID string                               `json:"root_account_cloud_id,omitempty"`
	Vendor             enumor.Vendor                        `json:"vendor,omitempty"`
	Type               enumor.MonthTaskType                 `json:"type"`
	BillYear           int                                  `json:"bill_year,omitempty"`
	BillMonth          int                                  `json:"bill_month,omitempty"`
	VersionID          int                                  `json:"version_id,omitempty"`
	State              enumor.RootAccountMonthBillTaskState `json:"state,omitempty"`
	Count              uint64                               `json:"count,omitempty"`
	Currency           enumor.CurrencyCode                  `json:"currency,omitempty"`
	Cost               decimal.Decimal                      `json:"cost,omitempty"`
	PullIndex          uint64                               `json:"pull_index,omitempty"`
	PullFlowID         string                               `json:"pull_flow_id,omitempty"`
	SplitIndex         uint64                               `json:"split_index,omitempty"`
	SplitFlowID        string                               `json:"split_flow_id,omitempty"`
	SummaryFlowID      string                               `json:"summary_flow_id,omitempty"`
	SummaryDetail      []MonthTaskSummaryDetailItem         `json:"summary_detail,omitempty"`
	Creator            string                               `json:"creator,omitempty"`
	Reviser            string                               `json:"reviser,omitempty"`
	CreatedAt          types.Time                           `json:"created_at,omitempty"`
	UpdatedAt          types.Time                           `json:"updated_at,omitempty"`
}

// String ...
func (t MonthTask) String() string {
	return fmt.Sprintf("[%s:%s]%s(%s) %d-%02dv%d:%s", t.Vendor, t.Type, t.RootAccountCloudID, t.RootAccountID,
		t.BillYear, t.BillMonth, t.VersionID, t.State)
}
