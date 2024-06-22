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

	"hcm/cmd/account-server/logics/bill/puller"
	"hcm/pkg/api/core"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// MainAccountControllerOption option for MainAccountController
type MainAccountControllerOption struct {
	RootAccountID string
	MainAccountID string
	Vendor        enumor.Vendor
	ProductID     int64
	BkBizID       int64
	Client        *client.ClientSet
}

// NewMainAccountController create new main account controller
func NewMainAccountController(opt *MainAccountControllerOption) (*MainAccountController, error) {
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

// Sync do sync
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
	ticker := time.NewTicker(defaultControllerSyncDuration)
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
	lastBillSummaryMain, err := mac.getBillSummary(kt, lastBillYear, lastBillMonth)
	if err != nil {
		return err
	}
	if lastBillSummaryMain.State == constant.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(lastBillSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(kt, mac.Client, lastBillSummaryMain); err != nil {
			return err
		}
	}
	// 本月
	curBillYear, curBillMonth := getCurrentBillMonth()
	billSummaryMain, err := mac.getBillSummary(kt, curBillYear, curBillMonth)
	if err != nil {
		return err
	}
	if billSummaryMain.State == constant.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(billSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(kt, mac.Client, billSummaryMain); err != nil {
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

func (mac *MainAccountController) getBillSummary(kt *kit.Kit, billYear, billMonth int) (*dsbillapi.BillSummaryMainResult, error) {
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

func (mac *MainAccountController) createNewBillSummary(kt *kit.Kit, billYear, billMonth int) error {
	_, err := mac.Client.DataService().Global.Bill.CreateBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainCreateReq{
			RootAccountID:     mac.RootAccountID,
			MainAccountID:     mac.MainAccountID,
			BkBizID:           mac.BkBizID,
			ProductID:         mac.ProductID,
			Vendor:            mac.Vendor,
			BillYear:          billYear,
			BillMonth:         billMonth,
			LastSyncedVersion: -1,
			CurrentVersion:    1,
			State:             constant.MainAccountBillSummaryStateAccounting,
		})
	if err != nil {
		return fmt.Errorf("failed to create bill summary for main account (%s, %s, %s) in in (%04d, %02d), err %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, err.Error())
	}
	logs.Infof("main account (%s, %s, %s) in (%04d, %02d) bill summary create successfully, rid: %s",
		mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth, kt.Rid)
	return nil
}

func (mac *MainAccountController) ensureBillSummary(kt *kit.Kit, billYear, billMonth int) error {
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
		return mac.createNewBillSummary(kt, billYear, billMonth)
	}
	return nil
}

func (mac *MainAccountController) getKey() string {
	return fmt.Sprintf("%s/%s/%s", mac.RootAccountID, mac.MainAccountID, mac.Vendor)
}
