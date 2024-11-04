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

// CurrencyCode 货币代码
type CurrencyCode string

const (
	// CurrencyUSD usd currency
	CurrencyUSD CurrencyCode = "USD"
	// CurrencyCNY rmb currency
	CurrencyCNY CurrencyCode = "CNY"
	// CurrencyRMB rmb currency
	CurrencyRMB = CurrencyCNY
)

// BillAdjustmentType 调账类型
type BillAdjustmentType string

const (
	// BillAdjustmentIncrease 增加
	BillAdjustmentIncrease BillAdjustmentType = "increase"
	// BillAdjustmentDecrease 减少
	BillAdjustmentDecrease BillAdjustmentType = "decrease"
)

// BillAdjustmentState 调账明细状态
type BillAdjustmentState string

const (
	// BillAdjustmentStateConfirmed 已确认
	BillAdjustmentStateConfirmed BillAdjustmentState = "confirmed"
	// BillAdjustmentStateUnconfirmed 未确认
	BillAdjustmentStateUnconfirmed BillAdjustmentState = "unconfirmed"
)

// RootBillSummaryState 一级账号账单汇总状态枚举
type RootBillSummaryState string

const (
	// RootAccountBillSummaryStateAccounting 核算中
	RootAccountBillSummaryStateAccounting RootBillSummaryState = "accounting"
	// RootAccountBillSummaryStateAccounted 已核算
	RootAccountBillSummaryStateAccounted RootBillSummaryState = "accounted"
	// RootAccountBillSummaryStateConfirmed 已确认
	RootAccountBillSummaryStateConfirmed RootBillSummaryState = "confirmed"
	// RootAccountBillSummaryStateSyncing 同步中
	RootAccountBillSummaryStateSyncing RootBillSummaryState = "syncing"
	// RootAccountBillSummaryStateSynced 已同步
	RootAccountBillSummaryStateSynced RootBillSummaryState = "synced"
	// RootAccountBillSummaryStateStop 已停止
	RootAccountBillSummaryStateStop RootBillSummaryState = "stopped"
)

// MainBillSummaryState  二级账号账单汇总状态
type MainBillSummaryState string

const (
	// MainAccountBillSummaryStateAccounting 核算中
	MainAccountBillSummaryStateAccounting MainBillSummaryState = "accounting"

	// MainAccountBillSummaryStateWaitMonthTask 等待月度分账
	MainAccountBillSummaryStateWaitMonthTask MainBillSummaryState = "waiting_month_task"

	// MainAccountBillSummaryStateAccounted 已核算
	MainAccountBillSummaryStateAccounted MainBillSummaryState = "accounted"

	// MainAccountBillSummaryStateSyncing 同步中
	MainAccountBillSummaryStateSyncing MainBillSummaryState = "syncing"

	// MainAccountBillSummaryStateSynced 已同步
	MainAccountBillSummaryStateSynced MainBillSummaryState = "synced"

	// MainAccountBillSummaryStateStop 停止中
	MainAccountBillSummaryStateStop MainBillSummaryState = "stopped"
)

// MainRawBillPullState 二级账号账单拉取状态
type MainRawBillPullState string

const (
	// MainAccountRawBillPullStatePulling 拉取中
	MainAccountRawBillPullStatePulling MainRawBillPullState = "pulling"

	// MainAccountRawBillPullStatePulled 已拉取
	MainAccountRawBillPullStatePulled MainRawBillPullState = "pulled"

	// MainAccountRawBillPullStateSplit 已分账
	MainAccountRawBillPullStateSplit MainRawBillPullState = "split"

	// MainAccountRawBillPullStateAccounted 已核算
	MainAccountRawBillPullStateAccounted MainRawBillPullState = "accounted"

	// MainAccountRawBillPullStateStop 停止中
	MainAccountRawBillPullStateStop MainRawBillPullState = "stopped"
)

// BillSyncState 云账单同步状态
type BillSyncState string

const (

	// BillSyncRecordStateNew 新增同步记录
	BillSyncRecordStateNew BillSyncState = "new"

	// BillSyncRecordStateSyncingBillItem 同步账单中
	BillSyncRecordStateSyncingBillItem BillSyncState = "syncing_bill_item"

	// BillSyncRecordStateSyncingAdjustment 同步调账中
	BillSyncRecordStateSyncingAdjustment BillSyncState = "syncing_adjustment_item"

	// BillSyncRecordStateWaitNotifying 同步等待通知
	BillSyncRecordStateWaitNotifying BillSyncState = "wait_notifying"

	// BillSyncRecordStateSynced 已同步
	BillSyncRecordStateSynced BillSyncState = "synced"

	// BillSyncRecordStateFailed 同步失败
	BillSyncRecordStateFailed BillSyncState = "failed"
)

// RootAccountMonthBillTaskState 一级账号月度账单（除去每日账单）状态
type RootAccountMonthBillTaskState string

const (
	// RootAccountMonthBillTaskStatePulling 拉取中
	RootAccountMonthBillTaskStatePulling = "pulling"

	// RootAccountMonthBillTaskStatePulled 已拉取
	RootAccountMonthBillTaskStatePulled = "pulled"

	// RootAccountMonthBillTaskStateSplit 已分账
	RootAccountMonthBillTaskStateSplit = "split"

	// RootAccountMonthBillTaskStateAccounted 已核算
	RootAccountMonthBillTaskStateAccounted = "accounted"

	// RootAccountMonthBillTaskStateStop 停止中
	RootAccountMonthBillTaskStateStop = "stopped"
)

// MonthTaskType 月度任务类型
type MonthTaskType string

const (
	// AwsOutsideBillMonthTask usage start date outside current bill month
	AwsOutsideBillMonthTask MonthTaskType = "outside_month_bill"
	// AwsSavingsPlansMonthTask aws savings plans month task
	AwsSavingsPlansMonthTask MonthTaskType = "savings_plans"
	// AwsSupportMonthTask aws support month task
	AwsSupportMonthTask MonthTaskType = "support"

	// GcpCreditsMonthTask gcp credits month task
	GcpCreditsMonthTask MonthTaskType = "credits"
	// GcpSupportMonthTask gcp support month task
	GcpSupportMonthTask MonthTaskType = "support"
)

// MonthTaskStep 月度任务步骤
type MonthTaskStep string

const (
	// MonthTaskStepPull 拉取类型
	MonthTaskStepPull MonthTaskStep = "pull"
	// MonthTaskStepSplit 分账类型
	MonthTaskStepSplit MonthTaskStep = "split"
	// MonthTaskStepSummary 汇总类型
	MonthTaskStepSummary MonthTaskStep = "summary"
)

const (
	// MonthRawBillPathName 拉取原始账单保存路径
	MonthRawBillPathName = "monthbill"
	// MonthRawBillSpecialDatePathName 特殊日期原始账单保存路径
	MonthRawBillSpecialDatePathName = "00"
)

// MonthTaskSpecialBillDay special bill day 0 to represent the whole month
const MonthTaskSpecialBillDay = 0

var (
	// BillAdjustmentStateNameMap is the map of bill adjustment state name
	BillAdjustmentStateNameMap = map[BillAdjustmentState]string{
		BillAdjustmentStateConfirmed:   "已确认",
		BillAdjustmentStateUnconfirmed: "未确认",
	}

	// BillAdjustmentTypeNameMap is the map of bill adjustment type name
	BillAdjustmentTypeNameMap = map[BillAdjustmentType]string{
		BillAdjustmentIncrease: "增加",
		BillAdjustmentDecrease: "减少",
	}

	// RootAccountBillSummaryStateMap 一级账号账单汇总状态中文名
	RootAccountBillSummaryStateMap = map[RootBillSummaryState]string{
		RootAccountBillSummaryStateAccounting: "核算中",
		RootAccountBillSummaryStateAccounted:  "已核算",
		RootAccountBillSummaryStateConfirmed:  "已确认",
		RootAccountBillSummaryStateSyncing:    "同步中",
		RootAccountBillSummaryStateSynced:     "已同步",
		RootAccountBillSummaryStateStop:       "停止中",
	}
)
