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
	ProductID     int64
	BkBizID       int64
	Client        *client.ClientSet
}

// MainSummaryDailyController main account daily summary controller
type MainSummaryDailyController struct {
	Client        *client.ClientSet
	RootAccountID string
	MainAccountID string
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
	kt := kit.New()
	cancelFunc := kt.CtxBackgroundWithCancel()
	msdc.kt = kt
	msdc.cancelFunc = cancelFunc
	go msdc.runBillDailySummaryLoop(kt)
	return nil
}

func (msdc *MainSummaryDailyController) runBillDailySummaryLoop(kt *kit.Kit) {
	if err := msdc.syncBillSummary(kt); err != nil {
		logs.Warnf("sync daily summary failed, err %s", err.Error())
	}
	ticker := time.NewTicker(defaultControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if err := msdc.syncBillSummary(kt); err != nil {
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

func (msdc *MainSummaryDailyController) syncBillSummary(kt *kit.Kit) error {
	return nil
}
