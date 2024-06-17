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
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// MainSummaryDailyControllerOption option for MainSummaryDailyController
type MainSummaryDailyControllerOption struct {
	RootAccountID string
	MainAccountID string
	Vendor        enumor.Vendor
	Version       int
	ProductID     int64
	BkBizID       int64
	Client        *client.ClientSet
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
		logs.Warnf("sync daily summary failed, err %s", err.Error())
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

func (msdc *MainSummaryDailyController) syncDailySummary(kt *kit.Kit, billYear, billMonth int) error {
	_, err := msdc.Client.TaskServer().CreateCustomFlow(kt, &taskserver.AddCustomFlowReq{
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
				msdc.Version,
			),
		},
	})
	if err != nil {
		return fmt.Errorf("create daily summary task flow failed for %s/%s/%s/%d/%d/%d, err %s",
			msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear, billMonth, msdc.Version, err.Error())
	}
	logs.Infof("create daily summary task flow for %s/%s/%s/%d/%d/%d successfully",
		msdc.RootAccountID, msdc.MainAccountID, msdc.Vendor, billYear, billMonth, msdc.Version)
	return nil
}
