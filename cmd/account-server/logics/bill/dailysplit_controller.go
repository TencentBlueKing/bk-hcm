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
	"strconv"
	"time"

	"hcm/cmd/account-server/logics/bill/puller"
	"hcm/cmd/task-server/logics/action/bill/dailysplit"
	"hcm/pkg/api/core"
	dsbillapi "hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"
)

// NewMainDailySplitController create main account daily splitter controller
func NewMainDailySplitController(opt *MainAccountControllerOption) (*MainDailySplitController, error) {
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
	return &MainDailySplitController{
		Client:        opt.Client,
		TenantID:      opt.TenantID,
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		ProductID:     opt.ProductID,
		BkBizID:       opt.BkBizID,
		Vendor:        opt.Vendor,

		RootAccountCloudID: opt.RootAccountCloudID,
		MainAccountCloudID: opt.MainAccountCloudID,
	}, nil
}

// MainDailySplitController main account daily summary controller
type MainDailySplitController struct {
	Client        *client.ClientSet
	TenantID      string
	RootAccountID string
	MainAccountID string
	ProductID     int64
	BkBizID       int64
	Vendor        enumor.Vendor
	ext           map[string]string
	kt            *kit.Kit

	RootAccountCloudID string
	MainAccountCloudID string

	cancelFunc context.CancelFunc
}

// Start run controller
func (msdc *MainDailySplitController) Start() error {
	if msdc.kt != nil {
		return fmt.Errorf("controller already start")
	}
	kt := getInternalKit()
	kt.SetTenant(msdc.TenantID)
	cancelFunc := kt.CtxBackgroundWithCancel()
	msdc.kt = kt
	msdc.cancelFunc = cancelFunc

	if msdc.Vendor == enumor.Aws {
		if err := msdc.setAwsExtension(kt); err != nil {
			return err
		}
	}
	go msdc.runBillDailySplitLoop(kt)
	return nil
}

func (msdc *MainDailySplitController) setAwsExtension(kt *kit.Kit) error {

	billAllocation := cc.AccountServer().BillAllocation
	// matching saving plan allocation option
	for _, spOpt := range billAllocation.AwsSavingsPlans {
		if spOpt.RootAccountCloudID != msdc.RootAccountCloudID {
			continue
		}
		logs.Infof("setting aws savings plans config for %s daily split, arn: %s, rid: %s",
			msdc.MainAccountCloudID, spOpt.SpArnPrefix, kt.Rid)
		msdc.ext = dailysplit.BuildAwsDailySplitOptionExt(spOpt.SpArnPrefix)
	}
	return nil
}

func (msdc *MainDailySplitController) runBillDailySplitLoop(kt *kit.Kit) {
	if err := msdc.doSync(kt); err != nil {
		logs.Warnf("sync daily split failed, err %s, rid: %s", err.Error(), kt.Rid)
	}
	ticker := time.NewTicker(*cc.AccountServer().Controller.ControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := msdc.doSync(kt); err != nil {
				logs.Warnf("sync daily split for account (%s, %s, %s) failed, err %s, rid: %s",
					msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, err.Error(), kt.Rid)
			}
		case <-kt.Ctx.Done():
			logs.Infof("main account (%s, %s, %s) daily split controller context done, rid: %s",
				msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, kt.Rid)
			return
		}
	}
}

func (msdc *MainDailySplitController) doSync(kt *kit.Kit) error {
	curBillYear, curBillMonth := times.GetCurrentMonthUTC()
	if err := msdc.syncDailySplit(kt.NewSubKit(), curBillYear, curBillMonth); err != nil {
		return fmt.Errorf("ensure daily split for %d %d failed, err %s", curBillYear, curBillMonth, err.Error())
	}
	lastBillYear, lastBillMonth := times.GetLastMonthUTC()
	if err := msdc.syncDailySplit(kt.NewSubKit(), lastBillYear, lastBillMonth); err != nil {
		return fmt.Errorf("ensure daily split for %d %d failed, err %s", lastBillYear, lastBillMonth, err.Error())
	}
	return nil
}

func (msdc *MainDailySplitController) getBillSummary(
	kt *kit.Kit, billYear, billMonth int) (*dsbillapi.BillSummaryMain, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", msdc.RootAccountID),
		tools.RuleEqual("main_account_id", msdc.MainAccountID),
		tools.RuleEqual("vendor", msdc.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
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
		return nil, fmt.Errorf("get invalid main account bill summary resp %d", result.Count)
	}
	return result.Details[0], nil
}

func (msdc *MainDailySplitController) syncDailySplit(kt *kit.Kit, billYear, billMonth int) error {

	logs.Infof("[%s] start daily split sync, period: %d-%d, main %s(%s), root %s(%s), rid: %s",
		msdc.Vendor, billYear, billMonth, msdc.MainAccountCloudID, msdc.MainAccountID,
		msdc.RootAccountCloudID, msdc.RootAccountID, kt.Rid)

	summary, err := msdc.getBillSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	curPuller, err := puller.GetDailyPuller(summary.Vendor)
	if err != nil {
		return err
	}
	pullTaskList, err := curPuller.GetPullTaskList(kt, msdc.Client, summary)
	if err != nil {
		return err
	}

	for _, task := range pullTaskList {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(defaultSleepMillisecond)))
		if task.State != enumor.MainAccountRawBillPullStatePulled {
			continue
		}
		if len(task.SplitFlowID) == 0 {
			logs.Infof("split task of day %d main account %v bill should be create", task.BillDay, summary)
			flowID, err := msdc.createDailySplitFlow(kt, summary, billYear, billMonth, task.BillDay)
			if err != nil {
				logs.Errorf("create daily split task for %v, %d/%d/%d failed, err %s, rid: %s",
					summary, billYear, billMonth, task.BillDay, err.Error(), kt.Rid)
				continue
			}
			logs.Infof("create daily split task flow %s for %s/%s/%s/%d/%d/%d/%d successfully, rid: %s",
				task.ID, msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear,
				billMonth, task.BillDay, summary.CurrentVersion, kt.Rid)
			if err := msdc.updateDailyPullTaskFlowID(kt, task.ID, flowID); err != nil {
				logs.Warnf("set pull task %s split flow id to %s failed, err %s, rid: %s",
					task.ID, flowID, err.Error(), kt.Rid)
				continue
			}
			continue
		}
		// 如果已经有拉取task flow，则检查拉取任务是否有问题
		flow, err := msdc.Client.TaskServer().GetFlow(kt, task.SplitFlowID)
		if err != nil {
			return fmt.Errorf("failed to get flow by id %s, err %s", task.SplitFlowID, err.Error())
		}
		if flow.State == enumor.FlowFailed || flow.State == enumor.FlowCancel {

			flowID, err := msdc.createDailySplitFlow(kt, summary, billYear, billMonth, task.BillDay)
			if err != nil {
				logs.Errorf("create daily split task for %v, %d/%d/%d failed, err %s, rid: %s",
					summary, billYear, billMonth, task.BillDay, err.Error(), kt.Rid)
				continue
			}
			logs.Infof("create new daily split task flow %s for %s/%s/%s/%d/%d/%d/%d successfully, rid: %s",
				task.ID, msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear,
				billMonth, task.BillDay, summary.CurrentVersion, kt.Rid)
			if err := msdc.updateDailyPullTaskFlowID(kt, task.ID, flowID); err != nil {
				logs.Warnf("update pull task %s split flow id to %s failed, err %s, rid: %s",
					task.ID, flowID, err.Error(), kt.Rid)
				continue
			}
		}
	}
	return nil
}

func (msdc *MainDailySplitController) createDailySplitFlow(kt *kit.Kit, summary *dsbillapi.BillSummaryMain,
	billYear, billMonth, billDay int) (string, error) {

	memo := fmt.Sprintf("[%s] main %s(%.16s)v%d %4d-%02d-%02d",
		summary.Vendor, summary.MainAccountID, summary.MainAccountCloudID, summary.CurrentVersion,
		summary.BillYear, summary.BillMonth, billDay)

	params := map[string]string{
		"root_account_id": summary.RootAccountID,
		"main_account_id": summary.MainAccountID,
		"vendor":          string(summary.Vendor),
		"bill_year":       strconv.Itoa(summary.BillYear),
		"bill_month":      strconv.Itoa(summary.BillMonth),
		"bill_day":        strconv.Itoa(billDay),
		"version":         strconv.Itoa(summary.CurrentVersion),
	}

	result, err := msdc.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
		Name:      enumor.FlowSplitBill,
		Memo:      memo,
		ShareData: tableasync.NewShareData(params),
		Tasks: []taskserver.CustomFlowTask{
			dailysplit.BuildDailySplitTask(
				msdc.RootAccountID,
				msdc.MainAccountID,
				msdc.Vendor,
				billYear,
				billMonth,
				billDay,
				summary.CurrentVersion,
				msdc.ext,
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("create daily split task flow failed for %s/%s/%s/%d/%d/%d/%d, err %s",
			msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear,
			billMonth, billDay, summary.CurrentVersion, err.Error())
	}
	return result.ID, nil
}

func (msdc *MainDailySplitController) updateDailyPullTaskFlowID(kt *kit.Kit, dataID, flowID string) error {
	return msdc.Client.DataService().Global.Bill.UpdateBillDailyPullTask(kt, &dsbillapi.BillDailyPullTaskUpdateReq{
		ID:          dataID,
		SplitFlowID: flowID,
	})
}
