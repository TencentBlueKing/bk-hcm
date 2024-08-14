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
	"hcm/cmd/task-server/logics/action/bill/monthtask"
	"hcm/cmd/task-server/logics/action/bill/rootsummary"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
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
	"hcm/pkg/tools/times"
)

// RootAccountControllerOption option for RootAccountController
type RootAccountControllerOption struct {
	RootAccountID      string
	RootAccountCloudID string
	Vendor             enumor.Vendor
	Client             *client.ClientSet
}

// NewRootAccountController create new root account controller
func NewRootAccountController(opt *RootAccountControllerOption) (*RootAccountController, error) {
	if opt == nil {
		return nil, fmt.Errorf("option of root account controller cannot be empty")
	}
	if opt.Client == nil {
		return nil, fmt.Errorf("client of root account controller cannot be empty")
	}
	if len(opt.RootAccountID) == 0 {
		return nil, fmt.Errorf("root account id of root account controller cannot be empty")
	}
	if len(opt.RootAccountCloudID) == 0 {
		return nil, fmt.Errorf("root account cloud id of root account controller cannot be empty")
	}
	if len(opt.Vendor) == 0 {
		return nil, fmt.Errorf("vendor of root account controller cannot be empty")
	}
	return &RootAccountController{
		Client:             opt.Client,
		RootAccountID:      opt.RootAccountID,
		RootAccountCloudID: opt.RootAccountCloudID,
		Vendor:             opt.Vendor,
	}, nil
}

// RootAccountController ...
type RootAccountController struct {
	Client             *client.ClientSet
	Sd                 serviced.ServiceDiscover
	RootAccountID      string
	RootAccountCloudID string
	Vendor             enumor.Vendor

	kt         *kit.Kit
	cancelFunc context.CancelFunc
}

// Start controller
func (rac *RootAccountController) Start() error {
	if rac.kt != nil {
		return fmt.Errorf("controller already start")
	}
	kt := getInternalKit()
	cancelFunc := kt.CtxBackgroundWithCancel()
	rac.kt = kt
	rac.cancelFunc = cancelFunc
	go rac.runBillSummaryLoop(kt)
	go rac.runCalculateBillSummaryLoop(kt)
	go rac.runMonthTaskLoop(kt)

	return nil
}

func (rac *RootAccountController) runBillSummaryLoop(kt *kit.Kit) {
	if err := rac.syncBillSummary(kt.NewSubKit()); err != nil {
		logs.Warnf("sync bill summary for account (%s, %s) failed, err %s, rid: %s",
			rac.RootAccountID, rac.Vendor, err.Error(), kt.Rid)
	}
	ticker := time.NewTicker(*cc.AccountServer().Controller.RootAccountSummarySyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := rac.syncBillSummary(kt.NewSubKit()); err != nil {
				logs.Warnf("sync bill summary for account (%s, %s) failed, err %s, rid: %s",
					rac.RootAccountID, rac.Vendor, err.Error(), kt.Rid)
			}
		case <-kt.Ctx.Done():
			logs.Infof("root account (%s, %s) summary controller context done, rid: %s",
				rac.RootAccountID, rac.Vendor, kt.Rid)
			return
		}
	}
}

func (rac *RootAccountController) runCalculateBillSummaryLoop(kt *kit.Kit) {
	ticker := time.NewTicker(*cc.AccountServer().Controller.RootAccountSummarySyncDuration)
	curMonthFlowID := ""
	lastMonthFlowID := ""
	for {
		select {
		case <-ticker.C:
			subKit := kt.NewSubKit()
			lastBillYear, lastBillMonth := times.GetLastMonthUTC()
			lastMonthFlowID = rac.pollRootSummaryTask(subKit, lastMonthFlowID, lastBillYear, lastBillMonth)
			curBillYear, curBillMonth := times.GetCurrentMonthUTC()
			curMonthFlowID = rac.pollRootSummaryTask(subKit, curMonthFlowID, curBillYear, curBillMonth)

		case <-kt.Ctx.Done():
			logs.Infof("root account (%s, %s) summary controller context done, rid: %s",
				rac.RootAccountID, rac.Vendor, kt.Rid)
			return
		}
	}
}

func (rac *RootAccountController) runMonthTaskLoop(kt *kit.Kit) {
	ticker := time.NewTicker(*cc.AccountServer().Controller.RootAccountSummarySyncDuration)
	for {
		select {
		case <-ticker.C:
			lastBillYear, lastBillMonth := times.GetLastMonthUTC()
			if err := rac.ensureMonthTask(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
				logs.Warnf("ensure last month task for (%s, %s) failed, err %s, rid: %s",
					rac.RootAccountID, rac.Vendor, err.Error(), kt.Rid)
			}
			curBillYear, curBillMonth := times.GetCurrentMonthUTC()
			if err := rac.ensureMonthTask(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
				logs.Warnf("ensure current month task for (%s, %s) failed, err %s, rid: %s",
					rac.RootAccountID, rac.Vendor, err.Error(), kt.Rid)
			}

		case <-kt.Ctx.Done():
			logs.Infof("root account (%s, %s) summary controller context done, rid: %s",
				rac.RootAccountID, rac.Vendor, kt.Rid)
			return
		}
	}
}

func (rac *RootAccountController) pollRootSummaryTask(subKit *kit.Kit, flowID string, billYear, billMonth int) string {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(defaultSleepMillisecond)))

	if len(flowID) == 0 {
		result, err := rac.createRootSummaryTask(subKit, billYear, billMonth)
		if err != nil {
			logs.Warnf("create new root summary task for %s/%s %d-%d failed, err %s, rid: %s",
				rac.RootAccountID, rac.Vendor, billYear, billMonth, err.Error(), subKit.Rid)
			return flowID
		}

		logs.Infof("create root summary task for %s/%s %d-%d successfully, flow id %s, rid: %s",
			rac.RootAccountID, rac.Vendor, billYear, billMonth, flowID, subKit.Rid)
		return result.ID
	}
	flow, err := rac.Client.TaskServer().GetFlow(subKit, flowID)
	if err != nil {
		logs.Warnf("get flow by id %s failed, err %s, rid: %s", flowID, err.Error(), subKit.Rid)
		return flowID
	}
	if flow.State == enumor.FlowSuccess || flow.State == enumor.FlowFailed || flow.State == enumor.FlowCancel {

		result, err := rac.createRootSummaryTask(subKit, billYear, billMonth)
		if err != nil {
			logs.Warnf("create new root summary task for %s/%s %d-%d failed, err %s, rid: %s",
				rac.RootAccountID, rac.Vendor, billYear, billMonth, err.Error(), subKit.Rid)
			return flowID
		}

		logs.Infof("create main summary task for %s/%s %d-%d successfully, flow id %s, rid: %s",
			rac.RootAccountID, rac.Vendor, billYear, billMonth, flowID, subKit.Rid)
		return result.ID
	}
	return flowID
}

func (rac *RootAccountController) createRootSummaryTask(
	kt *kit.Kit, billYear, billMonth int) (*core.CreateResult, error) {

	memo := fmt.Sprintf("[%s] root %s(%.16s) %d-%d", rac.Vendor,
		rac.RootAccountID, rac.RootAccountCloudID, billYear, billMonth)

	return rac.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name: enumor.FlowBillRootAccountSummary,
		Memo: memo,
		Tasks: []taskserver.CustomFlowTask{
			rootsummary.BuildRootSummaryTask(
				rac.RootAccountID, rac.Vendor, billYear, billMonth),
		},
	})
}

func (rac *RootAccountController) syncBillSummary(kt *kit.Kit) error {
	curBillYear, curBillMonth := times.GetCurrentMonthUTC()
	if err := rac.ensureBillSummary(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
		return fmt.Errorf("ensure root account bill summary for %d %d failed, err %s, rid: %s",
			curBillYear, curBillMonth, err.Error(), kt.Rid)
	}
	lastBillYear, lastBillMonth := times.GetLastMonthUTC()
	if err := rac.ensureBillSummary(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
		return fmt.Errorf("ensure root account bill summary for %d %d failed, err %s, rid: %s",
			lastBillYear, lastBillMonth, err.Error(), kt.Rid)
	}
	return nil
}

func (rac *RootAccountController) getBillSummary(kt *kit.Kit, billYear, billMonth int) (
	*dsbillapi.BillSummaryRootResult, error) {

	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rac.RootAccountID),
		tools.RuleEqual("vendor", rac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := rac.Client.DataService().Global.Bill.ListBillSummaryRoot(
		kt, &dsbillapi.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("get root account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("root account bill summary not found")
	}
	return result.Details[0], nil
}

func (rac *RootAccountController) ensureBillSummary(kt *kit.Kit, billYear, billMonth int) error {
	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rac.RootAccountID),
		tools.RuleEqual("vendor", rac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := rac.Client.DataService().Global.Bill.ListBillSummaryRoot(
		kt, &dsbillapi.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return fmt.Errorf("ensure root account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) == 0 {
		return rac.createNewBillSummary(kt, billYear, billMonth)
	}
	return nil
}

func (rac *RootAccountController) createNewBillSummary(kt *kit.Kit, billYear, billMonth int) error {
	_, err := rac.Client.DataService().Global.Bill.CreateBillSummaryRoot(
		kt, &dsbillapi.BillSummaryRootCreateReq{
			RootAccountID:      rac.RootAccountID,
			RootAccountCloudID: rac.RootAccountCloudID,
			Vendor:             rac.Vendor,
			BillYear:           billYear,
			BillMonth:          billMonth,
			LastSyncedVersion:  -1,
			CurrentVersion:     1,
			State:              enumor.RootAccountBillSummaryStateAccounting,
		})
	if err != nil {
		return fmt.Errorf("failed to create bill summary for root account (%s, %s) in in (%d, %02d), err %s",
			rac.RootAccountID, rac.Vendor, billYear, billMonth, err.Error())
	}
	logs.Infof("root account (%s, %s) in (%d, %02d) bill summary create successfully, rid: %s",
		rac.RootAccountID, rac.Vendor, billYear, billMonth, kt.Rid)
	return nil
}

func (rac *RootAccountController) ensureMonthTask(kt *kit.Kit, billYear, billMonth int) error {
	rootSummary, err := rac.getBillSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	monthPuller, err := puller.GetMonthPuller(rac.Vendor)
	if err != nil {
		return err
	}
	if !monthPuller.HasMonthPullTask() {
		logs.Infof("no month pull task for root account (%s, %s), skip, rid: %s", rac.RootAccountID, rac.Vendor, kt.Rid)
		return nil
	}

	mainSummaryList, err := rac.listAllMainSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	isAllAccounted := calculateAccountingState(mainSummaryList, rootSummary)
	if !isAllAccounted {
		logs.Infof("not all main account bill summary for root account (%s, %s, %d-%02d) were accounted, wait, rid: %s",
			rac.RootAccountID, rac.Vendor, billYear, billMonth, kt.Rid)
		return nil
	}

	monthTask, err := rac.getMonthPullTask(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	if monthTask == nil {
		if err := rac.createMonthPullTaskStub(kt, rootSummary); err != nil {
			logs.Errorf("fail to create month pull task, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		return nil
	}
	// 判断versionID是否一致，不一致，则重新创建month pull task
	if monthTask.VersionID != rootSummary.CurrentVersion {
		if err := rac.deleteMonthPullTask(kt, billYear, billMonth); err != nil {
			return err
		}
		return rac.createMonthPullTaskStub(kt, rootSummary)
	}
	switch monthTask.State {
	case enumor.RootAccountMonthBillTaskStatePulling:
		if err := rac.ensureMonthTaskPullStage(kt, monthTask); err != nil {
			return err
		}
	case enumor.RootAccountMonthBillTaskStatePulled:
		if err := rac.ensureMonthTaskSplitStage(kt, monthTask); err != nil {
			return err
		}
	case enumor.RootAccountMonthBillTaskStateSplit:
		if err := rac.ensureMonthTaskAccountStage(kt, monthTask); err != nil {
			return err
		}
	case enumor.RootAccountMonthBillTaskStateAccounted:
		// nothing
	}
	return nil
}

func calculateAccountingState(mainSummaryList []*dsbillapi.BillSummaryMainResult,
	rootSummary *dsbillapi.BillSummaryRootResult) (isAllAccounted bool) {

	isAllAccounted = true

	for _, mainSummary := range mainSummaryList {
		if mainSummary.CurrentVersion != rootSummary.CurrentVersion {
			isAllAccounted = false
			return
		}
		if mainSummary.State != enumor.MainAccountBillSummaryStateAccounted &&
			mainSummary.State != enumor.MainAccountBillSummaryStateWaitMonthTask {
			isAllAccounted = false
			return
		}
	}
	return isAllAccounted
}

func (rac *RootAccountController) listAllMainSummary(
	kt *kit.Kit, billYear, billMonth int) ([]*dsbillapi.BillSummaryMainResult, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rac.RootAccountID),
		tools.RuleEqual("vendor", rac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	result, err := rac.Client.DataService().Global.Bill.ListBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Count: true,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list main account bill summary of %+v failed, err %s", rac, err.Error())
	}
	if result.Count == 0 {
		return nil, fmt.Errorf("empty count in result %+v", result)
	}
	logs.Infof("found %d main account summary for %+v, rid: %s", result.Count, rac, kt.Rid)
	var mainSummaryList []*dsbillapi.BillSummaryMainResult
	for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		result, err = rac.Client.DataService().Global.Bill.ListBillSummaryMain(
			kt, &dsbillapi.BillSummaryMainListReq{
				Filter: tools.ExpressionAnd(expressions...),
				Page: &core.BasePage{
					Start: 0,
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, fmt.Errorf("list main account bill summary of %+v failed, err %s", rac, err.Error())
		}
		mainSummaryList = append(mainSummaryList, result.Details...)
	}
	return mainSummaryList, nil
}

func (rac *RootAccountController) ensureMonthTaskPullStage(kt *kit.Kit, task *dsbillapi.BillMonthTaskResult) error {
	if len(task.PullFlowID) == 0 {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypePull)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:         task.ID,
			PullFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task pull flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully create month pull task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	flow, err := rac.Client.TaskServer().GetFlow(kt, task.PullFlowID)
	if err != nil {
		logs.Warnf("get flow by id %s failed, err %s, rid: %s", task.PullFlowID, err.Error(), kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypePull)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:         task.ID,
			PullFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task pull flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month pull task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (rac *RootAccountController) ensureMonthTaskSplitStage(kt *kit.Kit, task *dsbillapi.BillMonthTaskResult) error {
	if len(task.SplitFlowID) == 0 {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypeSplit)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:          task.ID,
			SplitFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task split flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully create month split task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	flow, err := rac.Client.TaskServer().GetFlow(kt, task.SplitFlowID)
	if err != nil {
		logs.Warnf("get flow by id %s failed, err %s, rid: %s", task.SplitFlowID, err.Error(), kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypeSplit)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:          task.ID,
			SplitFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task split flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month split task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (rac *RootAccountController) ensureMonthTaskAccountStage(kt *kit.Kit, task *dsbillapi.BillMonthTaskResult) error {
	if len(task.SummaryFlowID) == 0 {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypeSummary)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:            task.ID,
			SummaryFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task summary flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully create month summary task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	flow, err := rac.Client.TaskServer().GetFlow(kt, task.SummaryFlowID)
	if err != nil {
		logs.Warnf("get flow by id %s failed, err %s, rid: %s", task.SummaryFlowID, err.Error(), kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := rac.createMonthFlow(
			kt, rac.RootAccountID, task.BillYear, task.BillMonth, enumor.MonthTaskTypeSummary)
		if err != nil {
			logs.Warnf("failed to create month task, err %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := rac.Client.DataService().Global.Bill.UpdateBillMonthPullTask(kt, &dsbillapi.BillMonthTaskUpdateReq{
			ID:            task.ID,
			SummaryFlowID: result.ID,
		}); err != nil {
			logs.Warnf("failed to update month pull task summary flow id %s, err: %s, rid: %s",
				result.ID, err.Error(), kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month summary task, flow id: %s, rid: %s", result.ID, kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (rac *RootAccountController) createMonthFlow(kt *kit.Kit, rootAccountID string, billYear, billMonth int,
	t enumor.MonthTaskType) (*core.CreateResult, error) {

	memo := fmt.Sprintf("[%s] root %s %s(%s) %d-%02d ", rac.Vendor, t, rootAccountID, rac.RootAccountCloudID,
		billYear, billMonth)

	return rac.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name: enumor.FlowBillMonthTask,
		Memo: memo,
		Tasks: []taskserver.CustomFlowTask{
			monthtask.BuildMonthTask(
				t, rootAccountID, rac.Vendor, billYear, billMonth),
		},
	})
}

func (rac *RootAccountController) createMonthPullTaskStub(kt *kit.Kit,
	rootSummary *dsbillapi.BillSummaryRootResult) error {

	createReq := &dsbillapi.BillMonthTaskCreateReq{
		RootAccountID:      rac.RootAccountID,
		RootAccountCloudID: rac.RootAccountCloudID,
		Vendor:             rac.Vendor,
		BillYear:           rootSummary.BillYear,
		BillMonth:          rootSummary.BillMonth,
		VersionID:          rootSummary.CurrentVersion,
		State:              enumor.RootAccountMonthBillTaskStatePulling,
	}
	taskResult, err := rac.Client.DataService().Global.Bill.CreateBillMonthPullTask(kt, createReq)
	if err != nil {
		logs.Infof("create month pull task failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return err
	}
	logs.Infof("create month pull task stub success, taskID: %s, rid: %s", taskResult.ID, kt.Rid)
	return nil
}

func (rac *RootAccountController) getMonthPullTask(
	kt *kit.Kit, billYear, billMonth int) (*dsbillapi.BillMonthTaskResult, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rac.RootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	result, err := rac.Client.DataService().Global.Bill.ListBillMonthPullTask(kt, &dsbillapi.BillMonthTaskListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		logs.Warnf("get month pull task failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return nil, fmt.Errorf("get month pull task failed, err: %s", err.Error())
	}
	if len(result.Details) == 0 {
		return nil, nil
	}
	if len(result.Details) != 1 {
		logs.Warnf("get invalid length month pull task, resp: %v, rid: %s", result, kt.Rid)
		return nil, fmt.Errorf("get invalid length month pull task, resp: %v", result)
	}
	return result.Details[0], nil
}

func (rac *RootAccountController) deleteMonthPullTask(kt *kit.Kit, billYear, billMonth int) error {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rac.RootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	return rac.Client.DataService().Global.Bill.DeleteBillMonthPullTask(kt, &dataservice.BatchDeleteReq{
		Filter: tools.ExpressionAnd(expressions...),
	})
}

// Stop controller
func (rac *RootAccountController) Stop() {
	if rac.cancelFunc != nil {
		rac.cancelFunc()
	}
}
