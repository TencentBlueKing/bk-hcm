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

// Package daily 每天账单拉取任务
package daily

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"hcm/cmd/task-server/logics/action/bill/dailypull"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"

	"github.com/shopspring/decimal"
)

const (
	defaultSleepMillisecond = 2000
)

// DailyPuller 执行每天账单拉取任务
type DailyPuller struct {
	RootAccountID string
	MainAccountID string
	// 主账号云id
	BillAccountID string
	ProductID     int64
	BkBizID       int64
	Vendor        enumor.Vendor
	BillYear      int
	BillMonth     int
	Version       int
	// 账单延迟查询时间
	BillDelay int
	Client    *client.ClientSet
	Sd        serviced.ServiceDiscover
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

// EnsurePullTask 检查拉取任务，不存在或失败则新建
func (dp *DailyPuller) EnsurePullTask(kt *kit.Kit) error {
	dayList := getBillDays(dp.BillYear, dp.BillMonth, dp.BillDelay, time.Now())
	if err := dp.ensureDailyPulling(kt, dayList); err != nil {
		return err
	}
	return nil
}

func (dp *DailyPuller) createDailyPullTask(kt *kit.Kit, billDay int) error {
	_, err := dp.Client.DataService().Global.Bill.CreateBillDailyPullTask(kt, &bill.BillDailyPullTaskCreateReq{
		RootAccountID: dp.RootAccountID,
		MainAccountID: dp.MainAccountID,
		Vendor:        dp.Vendor,
		ProductID:     dp.ProductID,
		BkBizID:       dp.BkBizID,
		BillYear:      dp.BillYear,
		BillMonth:     dp.BillMonth,
		BillDay:       billDay,
		VersionID:     dp.Version,
		State:         enumor.MainAccountRawBillPullStatePulling,
		Count:         0,
		Currency:      "",
		Cost:          decimal.NewFromFloat(0),
		FlowID:        "",
	})
	return err
}

func (dp *DailyPuller) updateDailyPullTaskFlowID(kt *kit.Kit, dataID, flowID string) error {
	return dp.Client.DataService().Global.Bill.UpdateBillDailyPullTask(kt, &bill.BillDailyPullTaskUpdateReq{
		ID:     dataID,
		FlowID: flowID,
	})
}

func (dp *DailyPuller) ensureDailyPulling(kt *kit.Kit, dayList []int) error {
	filter := dp.getFilter(0)
	billTaskResult, err := dp.Client.DataService().Global.Bill.ListBillDailyPullTask(
		kt, &bill.BillDailyPullTaskListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: 0,
				Limit: 31,
			},
		})
	if err != nil {
		return fmt.Errorf("get pull task for %v failed, err %s", dp, err.Error())
	}
	billTaskDayMap := make(map[int]struct{})
	for _, billTask := range billTaskResult.Details {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(defaultSleepMillisecond)))
		billTaskDayMap[billTask.BillDay] = struct{}{}
		// 如果没有创建拉取task flow，则创建
		if len(billTask.FlowID) == 0 {
			flowResult, err := dp.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
				Name: enumor.FlowPullRawBill,
				Memo: "pull daily raw bill",
				Tasks: []taskserver.CustomFlowTask{
					dailypull.BuildDailyPullTask(
						dp.RootAccountID,
						dp.MainAccountID,
						dp.BillAccountID,
						dp.Vendor,
						dp.BillYear,
						dp.BillMonth,
						billTask.BillDay,
						dp.Version,
					),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create custom flow, err %s", err.Error())
			}
			logs.Infof("create pull task flow for billTask %+v, rid: %s", billTask, kt.Rid)
			if err := dp.updateDailyPullTaskFlowID(kt, billTask.ID, flowResult.ID); err != nil {
				return fmt.Errorf("update flow id failed, err %s", err.Error())
			}
			continue
		}
		if billTask.State != enumor.MainAccountRawBillPullStatePulling {
			// 跳过非拉取中状态的任务
			continue
		}
		// 如果已经有拉取task flow，则检查拉取任务是否有问题
		flow, err := dp.Client.TaskServer().GetFlow(kt, billTask.FlowID)
		if err != nil {
			if !errf.IsRecordNotFound(err) {
				return fmt.Errorf("failed to get flow by id %s, err %s", billTask.FlowID, err.Error())
			}
			return dp.createNewPullTask(kt, billTask)
		}
		// 如果flow失败了或者flow找不到了，则重新创建一个新的flow
		if flow.State == enumor.FlowFailed || flow.State == enumor.FlowCancel {
			return dp.createNewPullTask(kt, billTask)
		}
	}
	for _, day := range dayList {
		if _, ok := billTaskDayMap[day]; !ok {
			// 如果不存在pull task数据，则创建新的pull task
			if err := dp.createDailyPullTask(kt, day); err != nil {
				logs.Warnf("create dailed pull task %v for day %d failed, err %s, rid: %s",
					dp, day, err.Error(), kt.Rid)
			}
			logs.Infof("create pull task for %v day %d successfully, rid: %s", dp, day, kt.Rid)
		}
	}
	return nil
}

func (dp *DailyPuller) createNewPullTask(kt *kit.Kit, billTask *bill.BillDailyPullTaskResult) error {
	flowResult, err := dp.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name: enumor.FlowPullRawBill,
		Memo: "pull daily raw bill",
		Tasks: []taskserver.CustomFlowTask{
			dailypull.BuildDailyPullTask(
				dp.RootAccountID,
				dp.MainAccountID,
				dp.BillAccountID,
				dp.Vendor,
				dp.BillYear,
				dp.BillMonth,
				billTask.BillDay,
				dp.Version,
			),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create daily raw bill pull flow, err %s", err.Error())
	}
	if err := dp.updateDailyPullTaskFlowID(kt, billTask.ID, flowResult.ID); err != nil {
		return fmt.Errorf("update daily pull flow id failed, err %s", err.Error())
	}

	return nil
}

// GetPullTaskList 获取拉取状态
func (dp *DailyPuller) GetPullTaskList(kit *kit.Kit) ([]*bill.BillDailyPullTaskResult, error) {
	filter := dp.getFilter(0)
	billTaskResult, err := dp.Client.DataService().Global.Bill.ListBillDailyPullTask(
		kit, &bill.BillDailyPullTaskListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: 0,
				Limit: 31,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list pull task failed, err %s", err.Error())
	}
	return billTaskResult.Details, nil
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

func getTaskServerKeyList(sd serviced.ServiceDiscover) ([]string, error) {
	taskServerNameList, err := sd.GetServiceAllNodeKeys(cc.TaskServerName)
	if err != nil {
		logs.Warnf("get task server name list failed, err %s", err.Error())
		return nil, err
	}
	var keyUUIDs []string
	for _, one := range taskServerNameList {
		split := strings.Split(one, "/")
		keyUUIDs = append(keyUUIDs, split[len(split)-1])
	}
	return keyUUIDs, nil
}
