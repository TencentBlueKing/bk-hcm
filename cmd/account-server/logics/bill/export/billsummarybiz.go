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

import "hcm/pkg/logs"

// BillSummaryBizTableHeader 账单调整导出表头
var BillSummaryBizTableHeader []string

var _ Table = (*BillSummaryBizTable)(nil)

func init() {
	var err error
	BillSummaryBizTableHeader, err = BillSummaryBizTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("bill adjustment table header init failed: %v", err)
	}
}

// BillSummaryBizTable 账单调整导出表头结构
type BillSummaryBizTable struct {
	BkBizID                   string `header:"业务ID"`
	BkBizName                 string `header:"业务"`
	CurrentMonthRMBCostSynced string `header:"已确认账单人民币（元）"`
	CurrentMonthCostSynced    string `header:"已确认账单美金（美元）"`
	CurrentMonthRMBCost       string `header:"当前账单人民币（元）"`
	CurrentMonthCost          string `header:"当前账单美金（美元）"`
}

// GetHeaderValues ...
func (b BillSummaryBizTable) GetHeaderValues() ([]string, error) {
	return parseHeaderFields(b)
}

// GetHeaders ...
func (b BillSummaryBizTable) GetHeaders() ([]string, error) {
	return parseHeader(b)
}
