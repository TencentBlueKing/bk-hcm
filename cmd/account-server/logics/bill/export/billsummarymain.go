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

// BillSummaryMainTableHeaders 账单调整导出表头
var BillSummaryMainTableHeaders [][]string

var _ table.Table = (*BillSummaryMainTable)(nil)

func init() {
	var err error
	BillSummaryMainTableHeaders, err = BillSummaryMainTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("bill adjustment table header init failed: %v", err)
	}
}

// BillSummaryMainTable 账单调整导出表头结构
type BillSummaryMainTable struct {
	MainAccountID   string `header:"二级账号ID"`
	MainAccountName string `header:"二级账号名称"`
	RootAccountID   string `header:"一级账号ID"`
	RootAccountName string `header:"一级账号名称"`

	BKBizID   string `header:"业务"`
	BKBizName string `header:"业务名称"`

	CurrentMonthRMBCostSynced string `header:"已确认账单人民币（元）"`
	CurrentMonthCostSynced    string `header:"已确认账单美金（美元）"`
	CurrentMonthRMBCost       string `header:"当前账单人民币（元）"`
	CurrentMonthCost          string `header:"当前账单美金（美元）"`
}

// GetValuesByHeader ...
func (b BillSummaryMainTable) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(b)
}

// GetHeaders ...
func (b BillSummaryMainTable) GetHeaders() ([][]string, error) {
	return table.GetHeaders(b)
}
