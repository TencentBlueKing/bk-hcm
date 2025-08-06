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

// BillAdjustmentTableHeaders 账单调整导出表头
var BillAdjustmentTableHeaders [][]string

var _ table.Table = (*BillAdjustmentTable)(nil)

func init() {
	var err error
	BillAdjustmentTableHeaders, err = BillAdjustmentTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("bill adjustment table header init failed: %v", err)
	}
}

// BillAdjustmentTable 账单调整导出表头结构
type BillAdjustmentTable struct {
	UpdateTime string `header:"更新时间"`
	BillID     string `header:"调账ID"`

	BKBizID   string `header:"业务"`
	BKBizName string `header:"业务名称"`

	MainAccountName string `header:"二级账号名称"`
	AdjustType      string `header:"调账类型"`
	Operator        string `header:"操作人"`
	Cost            string `header:"金额"`
	Currency        string `header:"币种"`
	AdjustStatus    string `header:"调账状态"`
}

// GetValuesByHeader ...
func (b BillAdjustmentTable) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(b)
}

// GetHeaders ...
func (b BillAdjustmentTable) GetHeaders() ([][]string, error) {
	return table.GetHeaders(b)
}
