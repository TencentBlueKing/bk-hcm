/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package azure bill puller Option
package azure

import (
	"fmt"

	"hcm/cmd/account-server/logics/bill/puller"
	"hcm/cmd/account-server/logics/bill/puller/daily"
	"hcm/pkg/api/data-service/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
)

const (
	defaultAzureDelay = 1
)

func init() {
	puller.DailyPullerRegistry[enumor.Azure] = &AzurePuller{
		BillDelay: defaultAzureDelay,
	}
	puller.MonthPullerRegistry[enumor.Azure] = &AzurePuller{
		BillDelay: defaultAzureDelay,
	}
}

// AzurePuller huawei puller
type AzurePuller struct {
	BillDelay int
}

// EnsurePullTask ...
func (hp *AzurePuller) EnsurePullTask(kt *kit.Kit, client *client.ClientSet,
	sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) error {

	azureMainAccount, err := client.DataService().Azure.MainAccount.Get(kt, billSummaryMain.MainAccountID)
	if err != nil {
		return fmt.Errorf("get azure main account failed, err %s", err.Error())
	}

	dp := &daily.DailyPuller{
		RootAccountID: billSummaryMain.RootAccountID,
		MainAccountID: billSummaryMain.MainAccountID,
		BillAccountID: azureMainAccount.CloudID,
		ProductID:     billSummaryMain.ProductID,
		BkBizID:       billSummaryMain.BkBizID,
		Vendor:        billSummaryMain.Vendor,
		BillYear:      billSummaryMain.BillYear,
		BillMonth:     billSummaryMain.BillMonth,
		Version:       billSummaryMain.CurrentVersion,
		BillDelay:     hp.BillDelay,
		Client:        client,
		Sd:            sd,
	}
	return dp.EnsurePullTask(kt)
}

// GetPullTaskList ...
func (hp *AzurePuller) GetPullTaskList(kt *kit.Kit, client *client.ClientSet,
	sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) (
	[]*bill.BillDailyPullTaskResult, error) {

	dp := &daily.DailyPuller{
		RootAccountID: billSummaryMain.RootAccountID,
		MainAccountID: billSummaryMain.MainAccountID,
		ProductID:     billSummaryMain.ProductID,
		BkBizID:       billSummaryMain.BkBizID,
		Vendor:        billSummaryMain.Vendor,
		BillYear:      billSummaryMain.BillYear,
		BillMonth:     billSummaryMain.BillMonth,
		Version:       billSummaryMain.CurrentVersion,
		BillDelay:     hp.BillDelay,
		Client:        client,
		Sd:            sd,
	}
	return dp.GetPullTaskList(kt)
}

// HasMonthPullTask return true if it has month pull task
func (hp *AzurePuller) HasMonthPullTask() bool {
	return false
}
