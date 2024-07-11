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
	"math/rand"
	"time"

	"hcm/cmd/account-server/logics/bill/puller"
	"hcm/cmd/task-server/logics/action/bill/mainsummary"
	"hcm/pkg/api/core"
	dsbillapi "hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/slice"
)

// MainAccountControllerOption option for MainAccountController
type MainAccountControllerOption struct {
	RootAccountID string
	MainAccountID string
	Vendor        enumor.Vendor
	ProductID     int64
	BkBizID       int64
	Client        *client.ClientSet
	Sd            serviced.ServiceDiscover
}

// NewMainAccountController create new main account controller
func NewMainAccountController(opt *MainAccountControllerOption) (*MainAccountController, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be empty")
	}
	if opt.Client == nil {
		return nil, fmt.Errorf("client cannot be empty")
	}
	if opt.Sd == nil {
		return nil, fmt.Errorf("servicediscovery cannot be empty")
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
	splitCtrl, err := NewMainDailySplitController(opt)
	if err != nil {
		return nil, err
	}
	dailySummaryCtrl, err := NewMainSummaryDailyController(opt)
	if err != nil {
		return nil, err
	}
	return &MainAccountController{
		Client:           opt.Client,
		Sd:               opt.Sd,
		RootAccountID:    opt.RootAccountID,
		MainAccountID:    opt.MainAccountID,
		ProductID:        opt.ProductID,
		BkBizID:          opt.BkBizID,
		Vendor:           opt.Vendor,
		splitCtrl:        splitCtrl,
		dailySummaryCtrl: dailySummaryCtrl,
	}, nil
}

// MainAccountController main account controller
type MainAccountController struct {
	Client        *client.ClientSet
	Sd            serviced.ServiceDiscover
	RootAccountID string
	MainAccountID string
	ProductID     int64
	BkBizID       int64
	Vendor        enumor.Vendor

	splitCtrl        *MainDailySplitController
	dailySummaryCtrl *MainSummaryDailyController

	kt         *kit.Kit
	cancelFunc context.CancelFunc
}

// Start run controller
func (mac *MainAccountController) Start() error {
	if mac.kt != nil {
		return fmt.Errorf("controller already start")
	}
	kt := getInternalKit()
	cancelFunc := kt.CtxBackgroundWithCancel()
	mac.kt = kt
	mac.cancelFunc = cancelFunc
	go mac.runBillSummaryLoop(kt)
	go mac.runDailyRawBillLoop(kt)
	go mac.runCalculateBillSummaryLoop(kt)

	// start split controller
	if err := mac.splitCtrl.Start(); err != nil {
		return err
	}
	// start daily summary controller
	if err := mac.dailySummaryCtrl.Start(); err != nil {
		return err
	}
	return nil
}

// do sync
func (mac *MainAccountController) syncBillSummary(kt *kit.Kit) error {
	curBillYear, curBillMonth := getCurrentBillMonth()
	if err := mac.ensureBillSummary(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
		return fmt.Errorf("ensure bill summary for %d %d failed, err %s, rid: %s",
			curBillYear, curBillMonth, err.Error(), kt.Rid)
	}
	lastBillYear, lastBillMonth := getLastBillMonth()
	if err := mac.ensureBillSummary(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
		return fmt.Errorf("ensure bill summary for %d %d failed, err %s, rid: %s",
			lastBillYear, lastBillMonth, err.Error(), kt.Rid)
	}
	return nil
}

func (mac *MainAccountController) runBillSummaryLoop(kt *kit.Kit) {
	if err := mac.syncBillSummary(kt.NewSubKit()); err != nil {
		logs.Warnf("sync bill summary for account (%s, %s, %s) failed, err %s, rid: %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error(), kt.Rid)
	}
	ticker := time.NewTicker(*cc.AccountServer().Controller.MainAccountSummarySyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := mac.syncBillSummary(kt.NewSubKit()); err != nil {
				logs.Warnf("sync bill summary for account (%s, %s, %s) failed, err %s, rid: %s",
					mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error(), kt.Rid)
			}
		case <-kt.Ctx.Done():
			logs.Infof("main account (%s, %s, %s) summary controller context done, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, kt.Rid)
			return
		}
	}
}

func (mac *MainAccountController) runCalculateBillSummaryLoop(kt *kit.Kit) {
	ticker := time.NewTicker(*cc.AccountServer().Controller.MainAccountSummarySyncDuration)
	curMonthflowID := ""
	lastMonthflowID := ""
	for {
		select {
		case <-ticker.C:
			subKit := kt.NewSubKit()
			lastBillYear, lastBillMonth := getLastBillMonth()
			lastMonthflowID = mac.pollMainSummaryTask(subKit, lastMonthflowID, lastBillYear, lastBillMonth)
			curBillYear, curBillMonth := getCurrentBillMonth()
			curMonthflowID = mac.pollMainSummaryTask(subKit, curMonthflowID, curBillYear, curBillMonth)

		case <-kt.Ctx.Done():
			logs.Infof("main account (%s, %s, %s) summary controller context done, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, kt.Rid)
			return
		}
	}
}

func (mac *MainAccountController) pollMainSummaryTask(subKit *kit.Kit, flowID string, billYear, billMonth int) string {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(defaultSleepMillisecond)))
	taskServerNameList, err := getTaskServerKeyList(mac.Sd)
	if err != nil {
		logs.Warnf("get task server name list failed, err %s", err.Error())
		return flowID
	}
	if len(flowID) == 0 {
		result, err := mac.createMainSummaryTask(subKit, billYear, billMonth)
		if err != nil {
			logs.Warnf("create new main summary task for %s/%s/%s %d-%d failed, err %s, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor,
				billYear, billMonth, err.Error(), subKit.Rid)
			return flowID
		}
		logs.Infof("create main summary task for %s/%s/%s %d-%d successfully, flow id %s, rid: %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor,
			billYear, billMonth, result.ID, subKit.Rid)
		return result.ID

	}
	flow, err := mac.Client.TaskServer().GetFlow(subKit, flowID)
	if err != nil {
		logs.Warnf("get flow by id %s failed, err %s, rid: %s", flowID, err.Error(), subKit.Rid)
		return flowID
	}
	// task server异常重启之后，已经处于scheduled的状态的flow永远不会执行
	// 此处需要进行判断，如果flow的worker不在当前task 列表中，并且处于scheduled状态，则需要重新创建flow
	if flow.State == enumor.FlowSuccess ||
		flow.State == enumor.FlowFailed ||
		flow.State == enumor.FlowCancel ||
		(flow.State == enumor.FlowScheduled &&
			flow.Worker != nil &&
			!slice.IsItemInSlice[string](taskServerNameList, *flow.Worker)) {

		if flow.State == enumor.FlowScheduled {
			if err := mac.Client.TaskServer().CancelFlow(subKit, flow.ID); err != nil {
				logs.Warnf("cancel flow %v failed, err %s, rid: %s", flow, err.Error(), subKit.Rid)
				return flowID
			}
		}
		result, err := mac.createMainSummaryTask(subKit, billYear, billMonth)
		if err != nil {
			logs.Warnf("create new main summary task for %s/%s/%s %d-%d failed, err %s, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor,
				billYear, billMonth, err.Error(), subKit.Rid)
			return flowID
		}

		logs.Infof("create main summary task for %s/%s/%s %d-%d successfully, flow id %s, rid: %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor,
			billYear, billMonth, flowID, subKit.Rid)
		return result.ID
	}
	return flowID
}

func (mac *MainAccountController) createMainSummaryTask(
	kt *kit.Kit, billYear, billMonth int) (*core.CreateResult, error) {

	return mac.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name: enumor.FlowBillMainAccountSummary,
		Memo: "calculate main account bill summary",
		Tasks: []taskserver.CustomFlowTask{
			mainsummary.BuildMainSummaryTask(
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth),
		},
	})
}

func (mac *MainAccountController) runDailyRawBillLoop(kt *kit.Kit) {
	if err := mac.syncDailyRawBill(kt); err != nil {
		logs.Warnf("sync daily raw bill for account (%s, %s, %s) failed, err %s, rid: %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error(), kt.Rid)
	}
	ticker := time.NewTicker(defaultControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := mac.syncDailyRawBill(kt); err != nil {
				logs.Warnf("sync daily raw bill for account (%s, %s, %s) failed, err %s, rid: %s",
					mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error(), kt.Rid)
			}
		case <-kt.Ctx.Done():
			logs.Infof("main account (%s, %s, %s) raw bill controller context done, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, kt.Rid)
			return
		}
	}
}

func (mac *MainAccountController) syncDailyRawBill(kt *kit.Kit) error {
	// 同步拉取任务
	// 上月
	lastBillYear, lastBillMonth := getLastBillMonth()
	lastBillSummaryMain, err := mac.getMainBillSummary(kt, lastBillYear, lastBillMonth)
	if err != nil {
		return err
	}
	if lastBillSummaryMain.State == enumor.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(lastBillSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(kt, mac.Client, mac.Sd, lastBillSummaryMain); err != nil {
			return err
		}
	}
	// 本月
	curBillYear, curBillMonth := getCurrentBillMonth()
	billSummaryMain, err := mac.getMainBillSummary(kt, curBillYear, curBillMonth)
	if err != nil {
		return err
	}
	if billSummaryMain.State == enumor.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(billSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(kt, mac.Client, mac.Sd, billSummaryMain); err != nil {
			return err
		}
	}
	return nil
}

// Stop stop controller
func (mac *MainAccountController) Stop() {
	if mac.cancelFunc != nil {
		mac.cancelFunc()
	}
}

func (mac *MainAccountController) getRootBillSummary(
	kt *kit.Kit, billYear, billMonth int) (*dsbillapi.BillSummaryRootResult, error) {

	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", mac.RootAccountID),
		tools.RuleEqual("vendor", mac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := mac.Client.DataService().Global.Bill.ListBillSummaryRoot(
		kt, &dsbillapi.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		logs.Warnf("get root account bill summary of %s/%s %d-%d failed, err %s",
			mac.RootAccountID, mac.Vendor, billYear, billMonth, err.Error())
		return nil, fmt.Errorf("get root account bill summary of %s/%s %d-%d failed, err %s",
			mac.RootAccountID, mac.Vendor, billYear, billMonth, err.Error())
	}
	if len(result.Details) == 0 {
		return nil, nil
	}
	if len(result.Details) != 1 {
		return nil, fmt.Errorf("invalid length of root account bill summary resp, %v", result.Details)
	}
	return result.Details[0], nil
}

func (mac *MainAccountController) getMainBillSummary(
	kt *kit.Kit, billYear, billMonth int) (*dsbillapi.BillSummaryMainResult, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", mac.RootAccountID),
		tools.RuleEqual("main_account_id", mac.MainAccountID),
		tools.RuleEqual("vendor", mac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	result, err := mac.Client.DataService().Global.Bill.ListBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get main account bill summary, err %s", err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("main account bill summary for %s, %d-%d not found", mac.getKey(), billYear, billMonth)
	}
	return result.Details[0], nil
}

func (mac *MainAccountController) createNewBillSummary(
	kt *kit.Kit, billYear, billMonth int, billSummary *dsbillapi.BillSummaryRootResult) error {
	_, err := mac.Client.DataService().Global.Bill.CreateBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainCreateReq{
			RootAccountID:     mac.RootAccountID,
			MainAccountID:     mac.MainAccountID,
			BkBizID:           mac.BkBizID,
			ProductID:         mac.ProductID,
			Vendor:            mac.Vendor,
			BillYear:          billYear,
			BillMonth:         billMonth,
			LastSyncedVersion: billSummary.LastSyncedVersion,
			CurrentVersion:    billSummary.CurrentVersion,
			State:             enumor.MainAccountBillSummaryStateAccounting,
		})
	if err != nil {
		return fmt.Errorf("failed to create bill summary for main account (%s, %s, %s) in in (%d, %02d), err %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, err.Error())
	}
	logs.Infof("main account (%s, %s, %s) in (%d, %02d) bill summary create successfully, rid: %s",
		mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, kt.Rid)
	return nil
}

func (mac *MainAccountController) ensureBillSummary(kt *kit.Kit, billYear, billMonth int) error {
	rootSummary, err := mac.getRootBillSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	if rootSummary == nil {
		return fmt.Errorf("root account bill summary of %s/%s %d-%d not found, wait",
			mac.RootAccountID, mac.Vendor, billYear, billMonth)
	}

	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", mac.RootAccountID),
		tools.RuleEqual("main_account_id", mac.MainAccountID),
		tools.RuleEqual("vendor", mac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := mac.Client.DataService().Global.Bill.ListBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return fmt.Errorf("ensure main account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) == 0 {
		return mac.createNewBillSummary(kt, billYear, billMonth, rootSummary)
	}
	if len(result.Details) != 1 {
		return fmt.Errorf("main account bill summary for %s, %d-%d not found", mac.getKey(), billYear, billMonth)
	}
	mainSummary := result.Details[0]
	if rootSummary.CurrentVersion != mainSummary.CurrentVersion {
		req := &dsbillapi.BillSummaryMainUpdateReq{
			ID:             mainSummary.ID,
			CurrentVersion: rootSummary.CurrentVersion,
			State:          enumor.MainAccountBillSummaryStateAccounting,
		}
		if err = mac.Client.DataService().Global.Bill.UpdateBillSummaryMain(
			kt, req); err != nil {
			logs.Warnf("failed to update bill summary for main account (%s, %s, %s) in in (%d, %02d), err %s, rid: %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, err.Error(), kt.Rid)
			return fmt.Errorf("failed to update bill summary for main account (%s, %s, %s) in in (%d, %02d), err %s",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, err.Error())
		}
		logs.Infof("successfully update main account bill summary for main account "+
			"(%s, %s, %s) in in (%d, %02d) to version %d, last synced version %d, rid: %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth,
			rootSummary.CurrentVersion, rootSummary.LastSyncedVersion, kt.Rid)
	}
	return nil
}

func (mac *MainAccountController) getKey() string {
	return fmt.Sprintf("%s/%s/%s", mac.RootAccountID, mac.MainAccountID, mac.Vendor)
}
