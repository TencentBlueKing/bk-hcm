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
	// PullerRegistry puller registry
	PullerRegistry = make(map[enumor.Vendor]Puller)
)

// GetPuller get puller by vendor
func GetPuller(vendor enumor.Vendor) (Puller, error) {
	puller, ok := PullerRegistry[vendor]
	if !ok {
		return nil, fmt.Errorf("unsupported vendor %s", vendor)
	}
	return puller, nil
}

// Puller puller interface
type Puller interface {
	EnsurePullTask(
		kt *kit.Kit, client *client.ClientSet,
		sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) error
	// GetPullTaskList 返回的map的key表示某个二级账号某月所有应该同步的天数的账单状态
	GetPullTaskList(
		kt *kit.Kit, client *client.ClientSet,
		sd serviced.ServiceDiscover, billSummaryMain *dsbillapi.BillSummaryMainResult) (
		[]*bill.BillDailyPullTaskResult, error)
}
