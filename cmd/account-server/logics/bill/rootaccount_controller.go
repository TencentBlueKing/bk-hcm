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

	monthtask2 "hcm/cmd/account-server/logics/bill/monthtask"
	"hcm/cmd/task-server/logics/action/bill/rootsummary"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
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
	// ext extension option for root account
	ext map[string]string

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
		runner := monthtask2.NewDefaultMonthTaskRunner(kt,
			rac.Vendor, rac.RootAccountID, rac.RootAccountCloudID, rac.Client)
		select {
		case <-ticker.C:
			lastBillYear, lastBillMonth := times.GetLastMonthUTC()
			if err := runner.EnsureMonthTask(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
				logs.Errorf("ensure last month task for (%s, %s) failed, err: %v, rid: %s",
					rac.RootAccountID, rac.Vendor, err, kt.Rid)
			}
			curBillYear, curBillMonth := times.GetCurrentMonthUTC()
			if err := runner.EnsureMonthTask(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
				logs.Errorf("ensure current month task for (%s, %s) failed, err: %v, rid: %s",
					rac.RootAccountID, rac.Vendor, err, kt.Rid)
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
	// TODO: 改为入参指定月份范围
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

func (rac *RootAccountController) getBillSummary(kt *kit.Kit, billYear, billMonth int) (*billcore.SummaryRoot, error) {

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

// Stop controller
func (rac *RootAccountController) Stop() {
	if rac.cancelFunc != nil {
		rac.cancelFunc()
	}
}
