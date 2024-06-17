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

package enumor

import "fmt"

// BillSyncPeriodType 账单同步周期类型
type BillSyncPeriodType string

// Validate the BillSyncPeriodType is valid or not
func (b BillSyncPeriodType) Validate() error {
	switch b {
	case Daily, Weekly, Monthly:
	default:
		return fmt.Errorf("unsupported bill sync period type: %s", b)
	}
	return nil
}

const (
	// Daily 每天拉取
	Daily BillSyncPeriodType = "daily"
	// Weekly 每周拉取
	Weekly BillSyncPeriodType = "weekly"
	// Monthly 每月拉取
	Monthly BillSyncPeriodType = "monthly"
)

// BillPullMode is bill pull mode
type BillPullMode string

// Validate the BillPullMode is valid or not
func (b BillPullMode) Validate() error {
	switch b {
	case AutoPull, ManualPull:
	default:
		return fmt.Errorf("unsupported bill pull mode: %s", b)
	}
	return nil
}

const (
	// AutoPull 自动拉取
	AutoPull BillPullMode = "auto"
	// ManualPull 手动拉取
	ManualPull BillPullMode = "manual"
)

// BillDayNumber is bill date type
type BillDayNumber int

// Validate the BillDayNumber is valid or not
func (b BillDayNumber) Validate() error {
	if b < 1 || b > 31 {
		return fmt.Errorf("unsupported bill day number %d", b)
	}
	return nil
}

const (
	// ActionPullDailyBill action for pull daily bill
	ActionPullDailyBill = "pulldailybill"
	// ActionBillSummary action for calculate bill summary
	ActionBillSummary = "billsummary"
	// ActionDailySummary action for calculate daily summary
	ActionDailySummary = "dailysummary"
)

const (
	// CurrencyUSD usd currency
	CurrencyUSD = "USD"
	// CurrencyRMB rmb currency
	CurrencyRMB = "RMB"
)
