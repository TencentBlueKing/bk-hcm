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

// Package puller ...
package puller

import (
	"fmt"

	"hcm/pkg/api/data-service/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
)

var (
	// DailyPullerRegistry puller registry
	DailyPullerRegistry = make(map[enumor.Vendor]DailyPuller)
	// MonthPullerRegistry month puller registry
	MonthPullerRegistry = make(map[enumor.Vendor]MonthPuller)
)

// GetDailyPuller get puller by vendor
func GetDailyPuller(vendor enumor.Vendor) (DailyPuller, error) {
	puller, ok := DailyPullerRegistry[vendor]
	if !ok {
		return nil, fmt.Errorf("unsupported vendor %s for daily puller", vendor)
	}
	return puller, nil
}

// GetMonthPuller get puller by vendor
func GetMonthPuller(vendor enumor.Vendor) (MonthPuller, error) {
	puller, ok := MonthPullerRegistry[vendor]
	if !ok {
		return nil, fmt.Errorf("unsupported vendor %s for month puller", vendor)
	}
	return puller, nil
}

// DailyPuller puller interface
type DailyPuller interface {
	EnsurePullTask(
		kt *kit.Kit, client *client.ClientSet,
		sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) error
	// GetPullTaskList 返回的map的key表示某个二级账号某月所有应该同步的天数的账单状态
	GetPullTaskList(
		kt *kit.Kit, client *client.ClientSet,
		sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) (
		[]*bill.BillDailyPullTaskResult, error)
}

// MonthPuller month puller interface
type MonthPuller interface {
	// HasMonthPullTask return true if it needs month pull task
	HasMonthPullTask() bool
}
