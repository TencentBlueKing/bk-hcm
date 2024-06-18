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

package daily

import (
	"fmt"
	"time"

	"hcm/cmd/task-server/logics/action/bill/dailypull"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// DailyPuller 执行每天账单拉取任务
type DailyPuller struct {
	RootAccountID string
	MainAccountID string
	ProductID     int64
	BkBizID       int64
	Vendor        enumor.Vendor
	BillYear      int
	BillMonth     int
	Version       int
	// 账单延迟查询时间
	BillDelay int
	Client    *client.ClientSet
}

func (dp *DailyPuller) getFilter(billDay int) *filter.Expression {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", dp.RootAccountID),
		tools.RuleEqual("main_account_id", dp.MainAccountID),
		tools.RuleEqual("vendor", dp.Vendor),
		tools.RuleEqual("version_id", dp.Version),
		tools.RuleEqual("bill_year", dp.BillYear),
		tools.RuleEqual("bill_month", dp.BillMonth),
	}
	if billDay > 0 {
		expressions = append(expressions, tools.RuleEqual("bill_day", billDay))
	}
	return tools.ExpressionAnd(expressions...)
}

func (dp *DailyPuller) EnsurePullTask(kit *kit.Kit) error {
	dayList := getBillDays(dp.BillYear, dp.BillMonth, dp.BillDelay, time.Now())
	for _, day := range dayList {
		if err := dp.ensureDailyPulling(kit, day); err != nil {
			return err
		}
	}
	return nil
}

func (dp *DailyPuller) createDailyPullTask(kit *kit.Kit, billDay int) error {
	_, err := dp.Client.DataService().Global.Bill.CreateBillDailyPullTask(kit, &bill.BillDailyPullTaskCreateReq{
		RootAccountID: dp.RootAccountID,
		MainAccountID: dp.MainAccountID,
		Vendor:        dp.Vendor,
		ProductID:     dp.ProductID,
		BkBizID:       dp.BkBizID,
		BillYear:      dp.BillYear,
		BillMonth:     dp.BillMonth,
		BillDay:       billDay,
		VersionID:     dp.Version,
		State:         constant.MainAccountRawBillPullStatePulling,
		Count:         0,
		Currency:      "",
		Cost:          decimal.NewFromFloat(0),
		FlowID:        "",
	})
	return err
}

func (dp *DailyPuller) updateDailyPullTaskFlowID(kit *kit.Kit, dataID, flowID string) error {
	return dp.Client.DataService().Global.Bill.UpdateBillDailyPullTask(kit, &bill.BillDailyPullTaskUpdateReq{
		ID:     dataID,
		FlowID: flowID,
	})
}

func (dp *DailyPuller) ensureDailyPulling(kt *kit.Kit, billDay int) error {
	filter := dp.getFilter(billDay)
	billTaskResult, err := dp.Client.DataService().Global.Bill.ListBillDailyPullTask(
		kt, &bill.BillDailyPullTaskListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return fmt.Errorf("get pull task for %d failed, err %s", billDay, err.Error())
	}
	// 如果不存在pull task数据，则创建新的pull task
	if len(billTaskResult.Details) == 0 {
		return dp.createDailyPullTask(kt, billDay)
	}
	if len(billTaskResult.Details) != 1 {
		return fmt.Errorf("more than 1 pull task found, details %v", billTaskResult.Details)
	}
	billTask := billTaskResult.Details[0]

	// 如果没有创建拉取task flow，则创建
	if len(billTask.FlowID) == 0 {
		flowResult, err := dp.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
			Name: enumor.FlowPullRawBill,
			Memo: "pull daily raw bill",
			Tasks: []taskserver.CustomFlowTask{
				dailypull.BuildDailyPullTask(
					dp.RootAccountID,
					dp.MainAccountID,
					dp.Vendor,
					dp.BillYear,
					dp.BillMonth,
					billDay,
					dp.Version,
				),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create custom flow, err %s", err.Error())
		}
		logs.Infof("create pull task flow for billTask %v, rid: %s", billTask, kt.Rid)
		if err := dp.updateDailyPullTaskFlowID(kt, billTask.ID, flowResult.ID); err != nil {
			return fmt.Errorf("update flow id failed, err %s", err.Error())
		}
		return nil
	}

	// 如果已经有拉取task flow，则检查拉取任务是否有问题
	flow, err := dp.Client.TaskServer().GetFlow(kt, billTask.FlowID)
	if err != nil {
		return fmt.Errorf("failed to get flow by id %s", billTask.FlowID)
	}
	// 如果flow失败了，则重新创建一个新的flow
	if flow.State == enumor.FlowFailed {
		flowResult, err := dp.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
			Name: enumor.FlowPullRawBill,
			Memo: "pull daily raw bill",
			Tasks: []taskserver.CustomFlowTask{
				dailypull.BuildDailyPullTask(
					dp.RootAccountID,
					dp.MainAccountID,
					dp.Vendor,
					dp.BillYear,
					dp.BillMonth,
					billDay,
					dp.Version,
				),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create custom flow, err %s", err.Error())
		}
		if err := dp.updateDailyPullTaskFlowID(kt, billTask.ID, flowResult.ID); err != nil {
			return fmt.Errorf("update flow id failed, err %s", err.Error())
		}
		return nil
	}
	return nil
}

// GetPullState 获取拉取状态
func (dp *DailyPuller) GetPullState(kit *kit.Kit) (string, error) {
	filter := dp.getFilter(0)
	billTaskResult, err := dp.Client.DataService().Global.Bill.ListBillDailyPullTask(
		kit, &bill.BillDailyPullTaskListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return "", fmt.Errorf("list pull task failed, err %s", err.Error())
	}

	days := daysInMonth(dp.BillYear, time.Month(dp.BillMonth))
	if len(billTaskResult.Details) != days {
		return constant.MainAccountRawBillPullStatePulling, nil
	}
	for _, pullTask := range billTaskResult.Details {
		if len(pullTask.FlowID) == 0 {
			return constant.MainAccountRawBillPullStatePulling, nil
		}
		// 如果已经有拉取task flow，则检查拉取任务是否有问题
		flow, err := dp.Client.TaskServer().GetFlow(kit, pullTask.FlowID)
		if err != nil {
			return "", fmt.Errorf("failed to get flow by id %s", pullTask.FlowID)
		}
		if flow.State != enumor.FlowSuccess {
			return constant.MainAccountRawBillPullStatePulling, nil
		}
	}

	return constant.MainAccountRawBillPullStatePulled, nil
}

// daysInMonth 返回给定年份和月份的天数
func daysInMonth(year int, month time.Month) int {
	// 获取下个月的第一天
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	// 获取本月的最后一天
	lastOfThisMonth := firstOfNextMonth.AddDate(0, 0, -1)

	return lastOfThisMonth.Day()
}

func getBillDays(billYear, billMonth, billDelay int, now time.Time) []int {
	timeBill := now.AddDate(0, 0, -billDelay)
	var retList []int
	for t := time.Date(
		billYear, time.Month(billMonth), 1, 0, 0, 0, 0, now.Location()); t.Before(timeBill); t = t.AddDate(0, 0, 1) {
		if t.After(timeBill) {
			break
		}
		retList = append(retList, t.Day())
	}
	return retList
}
