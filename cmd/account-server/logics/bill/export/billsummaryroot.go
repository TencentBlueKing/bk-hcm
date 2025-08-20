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

package export

import (
	"hcm/pkg/logs"
	"hcm/pkg/table"
)

// BillSummaryRootTableHeaders 账单调整导出表头
var BillSummaryRootTableHeaders [][]string

var _ table.Table = (*BillSummaryRootTable)(nil)

func init() {
	var err error
	BillSummaryRootTableHeaders, err = BillSummaryRootTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("bill adjustment table header init failed: %v", err)
	}
}

// BillSummaryRootTable 账单调整导出表头结构
type BillSummaryRootTable struct {
	RootAccountID             string `header:"一级账号ID"`
	RootAccountName           string `header:"一级账号名称"`
	State                     string `header:"账号状态"`
	CurrentMonthRMBCostSynced string `header:"账单同步（人民币-元）当月"`
	LastMonthRMBCostSynced    string `header:"账单同步（人民币-元）上月"`
	CurrentMonthCostSynced    string `header:"账单同步（美金-美元）当月"`
	LastMonthCostSynced       string `header:"账单同步（美金-美元）上月"`
	MonthOnMonthValue         string `header:"账单同步环比"`
	CurrentMonthRMB           string `header:"当前账单人民币（元）"`
	CurrentMonthCost          string `header:"当前账单美金（美元）"`
	AdjustRMBCost             string `header:"调账人民币（元）"`
	AdjustCost                string `header:"调账美金（美元）"`
}

// GetValuesByHeader ...
func (b BillSummaryRootTable) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(b)
}

// GetHeaders ...
func (b BillSummaryRootTable) GetHeaders() ([][]string, error) {
	return table.GetHeaders(b)
}
