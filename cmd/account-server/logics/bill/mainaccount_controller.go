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
	return &MainAccountController{
		Client:        opt.Client,
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		ProductID:     opt.ProductID,
		BkBizID:       opt.BkBizID,
		Vendor:        opt.Vendor,
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

	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Start run controller
func (mac *MainAccountController) Start() error {
	if mac.ctx != nil {
		return fmt.Errorf("controller already start")
	}
	ctx, cancel := context.WithCancel(context.Background())
	mac.ctx = ctx
	mac.cancelFunc = cancel
	go mac.runBillSummaryLoop(mac.ctx)
	go mac.runDailyRawBillLoop(mac.ctx)
	return nil
}

// Sync do sync
func (mac *MainAccountController) syncBillSummary() error {
	curBillYear, curBillMonth := mac.getCurrentBillMonth()
	if err := mac.ensureBillSummary(curBillYear, curBillMonth); err != nil {
		return err
	}
	lastBillYear, lastBillMonth := mac.getLastBillMonth()
	if err := mac.ensureBillSummary(lastBillYear, lastBillMonth); err != nil {
		return err
	}
	return nil
}

func (mac *MainAccountController) runBillSummaryLoop(ctx context.Context) {
	if err := mac.syncBillSummary(); err != nil {
		logs.Warnf("sync bill summary for account (%s, %s, %s) failed, err %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error())
	}
	ticker := time.NewTicker(defaultControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := mac.syncBillSummary(); err != nil {
				logs.Warnf("sync bill summary for account (%s, %s, %s) failed, err %s",
					mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error())
			}
		case <-ctx.Done():
			logs.Infof("main account (%s, %s, %s) summary controller context done",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor)
			return
		}
	}
}

func (mac *MainAccountController) runDailyRawBillLoop(ctx context.Context) {
	if err := mac.syncDailyRawBill(); err != nil {
		logs.Warnf("sync daily raw bill for account (%s, %s, %s) failed, err %s",
			mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error())
	}
	ticker := time.NewTicker(defaultControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := mac.syncDailyRawBill(); err != nil {
				logs.Warnf("sync daily raw bill for account (%s, %s, %s) failed, err %s",
					mac.RootAccountID, mac.MainAccountID, mac.Vendor, err.Error())
			}
		case <-ctx.Done():
			logs.Infof("main account (%s, %s, %s) raw bill controller context done",
				mac.RootAccountID, mac.MainAccountID, mac.Vendor)
			return
		}
	}
}

func (mac *MainAccountController) syncDailyRawBill() error {
	// 同步拉取任务
	// 上月
	lastBillYear, lastBillMonth := mac.getLastBillMonth()
	lastBillSummaryMain, err := mac.getBillSummary(lastBillYear, lastBillMonth)
	if err != nil {
		return err
	}
	if lastBillSummaryMain.State == constant.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(lastBillSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(getInternalKit(), mac.Client, lastBillSummaryMain); err != nil {
			return err
		}
	}
	// 本月
	curBillYear, curBillMonth := mac.getCurrentBillMonth()
	billSummaryMain, err := mac.getBillSummary(curBillYear, curBillMonth)
	if err != nil {
		return err
	}
	if billSummaryMain.State == constant.MainAccountBillSummaryStateAccounting {
		curPuller, err := puller.GetPuller(billSummaryMain.Vendor)
		if err != nil {
			return err
		}
		if err := curPuller.EnsurePullTask(getInternalKit(), mac.Client, billSummaryMain); err != nil {
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

func (mac *MainAccountController) getBillSummary(billYear, billMonth int) (*dsbillapi.BillSummaryMainResult, error) {
	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", mac.RootAccountID),
		tools.RuleEqual("main_account_id", mac.MainAccountID),
		tools.RuleEqual("vendor", mac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := mac.Client.DataService().Global.Bill.ListBillSummaryMain(
		getInternalKit(), &dsbillapi.BillSummaryMainListReq{
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

func (mac *MainAccountController) createNewBillSummary(billYear, billMonth int) error {
	_, err := mac.Client.DataService().Global.Bill.CreateBillSummaryMain(
		getInternalKit(), &dsbillapi.BillSummaryMainCreateReq{
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
	logs.Infof("main account (%s, %s, %s) in (%04d, %02d) bill summary create successfully",
		mac.RootAccountID, mac.MainAccountID, mac.Vendor, billYear, billMonth)
	return nil
}

func (mac *MainAccountController) ensureBillSummary(billYear, billMonth int) error {
	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("root_account_id", mac.RootAccountID),
		tools.RuleEqual("main_account_id", mac.MainAccountID),
		tools.RuleEqual("vendor", mac.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}...)
	result, err := mac.Client.DataService().Global.Bill.ListBillSummaryMain(
		getInternalKit(), &dsbillapi.BillSummaryMainListReq{
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
		return mac.createNewBillSummary(billYear, billMonth)
	}
	return nil
}

func (mac *MainAccountController) getCurrentBillMonth() (int, int) {
	now := time.Now().UTC()
	return now.Year(), int(now.Month())
}

func (mac *MainAccountController) getLastBillMonth() (int, int) {
	now := time.Now().UTC()
	lastMonthNow := now.AddDate(0, -1, 0)
	return lastMonthNow.Year(), int(lastMonthNow.Month())
}

func (mac *MainAccountController) getKey() string {
	return fmt.Sprintf("%s/%s/%s", mac.RootAccountID, mac.MainAccountID, mac.Vendor)
}
