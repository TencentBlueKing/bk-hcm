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
	"hcm/pkg/dal/table/types"

	"github.com/shopspring/decimal"
)

// SummaryDaily summary daily
type SummaryDaily struct {
	ID                 string              `json:"id,omitempty"`
	RootAccountID      string              `json:"root_account_id"`
	MainAccountID      string              `json:"main_account_id"`
	RootAccountCloudID string              `json:"root_account_cloud_id"`
	MainAccountCloudID string              `json:"main_account_cloud_id"`
	Vendor             enumor.Vendor       `json:"vendor"`
	ProductID          int64               `json:"product_id"`
	BkBizID            int64               `json:"bk_biz_id"`
	BillYear           int                 `json:"bill_year"`
	BillMonth          int                 `json:"bill_month"`
	BillDay            int                 `json:"bill_day"`
	VersionID          int                 `json:"version_id"`
	Currency           enumor.CurrencyCode `json:"currency"`
	Cost               decimal.Decimal     `json:"cost"`
	Count              int64               `json:"count"`
	core.Revision
}

// Key get key
func (b *SummaryDaily) Key() string {
	return fmt.Sprintf("%s/%s/%s/%d/%d/%d/%d",
		b.RootAccountID, b.MainAccountID, b.Vendor, b.BillYear, b.BillMonth, b.BillDay, b.VersionID)
}

// SummaryRoot result
type SummaryRoot struct {
	ID                        string                      `json:"id,omitempty"`
	RootAccountID             string                      `json:"root_account_id"`
	RootAccountCloudID        string                      `json:"root_account_cloud_id"`
	Vendor                    enumor.Vendor               `json:"vendor"`
	BillYear                  int                         `json:"bill_year"`
	BillMonth                 int                         `json:"bill_month"`
	LastSyncedVersion         int                         `json:"last_synced_version"`
	CurrentVersion            int                         `json:"current_version"`
	Currency                  enumor.CurrencyCode         `json:"currency"`
	LastMonthCostSynced       decimal.Decimal             `json:"last_month_cost_synced"`
	LastMonthRMBCostSynced    decimal.Decimal             `json:"last_month_rmb_cost_synced"`
	CurrentMonthCostSynced    decimal.Decimal             `json:"current_month_cost_synced"`
	CurrentMonthRMBCostSynced decimal.Decimal             `json:"current_month_rmb_cost_synced"`
	MonthOnMonthValue         float64                     `json:"month_on_month_value"`
	CurrentMonthCost          decimal.Decimal             `json:"current_month_cost"`
	CurrentMonthRMBCost       decimal.Decimal             `json:"current_month_rmb_cost"`
	AdjustmentCost            decimal.Decimal             `json:"adjustment_cost"`
	AdjustmentRMBCost         decimal.Decimal             `json:"adjustment_rmb_cost"`
	Rate                      float64                     `json:"rate"`
	BkBizNum                  uint64                      `json:"bk_biz_num"`
	ProductNum                uint64                      `json:"product_num"`
	State                     enumor.RootBillSummaryState `json:"state"`
	CreatedAt                 types.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 types.Time                  `json:"updated_at,omitempty"`
}
