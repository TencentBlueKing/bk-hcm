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
	"context"
	"fmt"
	"time"

	"hcm/cmd/task-server/logics/action/bill/dailysummary"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// NewMainSummaryDailyController create main account daily splitter controller
func NewMainSummaryDailyController(opt *MainAccountControllerOption) (*MainSummaryDailyController, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be empty")
	}
	if opt.Client == nil {
		return nil, fmt.Errorf("client cannot be empty")
	}
	if len(opt.MainAccountID) == 0 {
		return nil, fmt.Errorf("main account id cannot be empty")
	}
	if len(opt.RootAccountID) == 0 {
		return nil, fmt.Errorf("root account id cannot be empty")
	}
	if opt.ProductID == 0 && opt.BkBizID == 0 {
		return nil, fmt.Errorf("product_id or bk_biz_id cannot be empty")
	}
	if len(opt.Vendor) == 0 {
		return nil, fmt.Errorf("vendor cannot be empty")
	}
	return &MainSummaryDailyController{
		Client:        opt.Client,
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		ProductID:     opt.ProductID,
		BkBizID:       opt.BkBizID,
		Vendor:        opt.Vendor,
	}, nil
}

// MainSummaryDailyController main account daily summary controller
type MainSummaryDailyController struct {
	Client        *client.ClientSet
	RootAccountID string
	MainAccountID string
	Version       int
	ProductID     int64
	BkBizID       int64
	Vendor        enumor.Vendor

	kt         *kit.Kit
	cancelFunc context.CancelFunc
}

// Start run controller
func (msdc *MainSummaryDailyController) Start() error {
	if msdc.kt != nil {
		return fmt.Errorf("controller already start")
	}
	kt := getInternalKit()
	cancelFunc := kt.CtxBackgroundWithCancel()
	msdc.kt = kt
	msdc.cancelFunc = cancelFunc
	go msdc.runBillDailySummaryLoop(kt)
	return nil
}

func (msdc *MainSummaryDailyController) runBillDailySummaryLoop(kt *kit.Kit) {
	if err := msdc.doSync(kt); err != nil {
		logs.Warnf("sync daily summary failed, err %s, rid: %s", err.Error(), kt.Rid)
	}
	ticker := time.NewTicker(defaultDailySummaryDuration)
	for {
		select {
		case <-ticker.C:
			if err := msdc.doSync(kt); err != nil {
				logs.Warnf("sync daily summary for account (%s, %s, %s) failed, err %s",
					msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, err.Error())
			}
		case <-kt.Ctx.Done():
			logs.Infof("main account (%s, %s, %s) daily summary controller context done",
				msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor)
			return
		}
	}
}

func (msdc *MainSummaryDailyController) doSync(kt *kit.Kit) error {
	curBillYear, curBillMonth := getCurrentBillMonth()
	if err := msdc.syncDailySummary(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
		return fmt.Errorf("ensure bill summary for %d %d failed, err %s", curBillYear, curBillMonth, err.Error())
	}
	lastBillYear, lastBillMonth := getLastBillMonth()
	if err := msdc.syncDailySummary(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
		return fmt.Errorf("ensure bill summary for %d %d failed, err %s", lastBillYear, lastBillMonth, err.Error())
	}
	return nil
}

func (msdc *MainSummaryDailyController) getBillSummary(
	kt *kit.Kit, billYear, billMonth int) (*bill.BillSummaryMainResult, error) {

	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", msdc.RootAccountID),
		tools.RuleEqual("main_account_id", msdc.MainAccountID),
		tools.RuleEqual("vendor", msdc.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := msdc.Client.DataService().Global.Bill.ListBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("get main account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) != 1 {
		return nil, fmt.Errorf("get invalid main account bill summary resp %v", result)
	}
	return result.Details[0], nil
}

func (msdc *MainSummaryDailyController) syncDailySummary(kt *kit.Kit, billYear, billMonth int) error {

	summary, err := msdc.getBillSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	_, err = msdc.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name: enumor.FlowBillDailySummary,
		Memo: "do daily summary",
		Tasks: []taskserver.CustomFlowTask{
			dailysummary.BuildDailySummaryTask(
				msdc.RootAccountID,
				msdc.MainAccountID,
				msdc.Vendor,
				msdc.ProductID,
				msdc.BkBizID,
				billYear,
				billMonth,
				summary.CurrentVersion,
			),
		},
	})
	if err != nil {
		return fmt.Errorf("create daily summary task flow failed for %s/%s/%s/%d/%d/%d, err %s",
			msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear, billMonth, msdc.Version, err.Error())
	}
	logs.Infof("create daily summary task flow for %s/%s/%s/%d/%d/%d successfully, rid: %s",
		msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear, billMonth, msdc.Version, kt.Rid)
	return nil
}
