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

// HuaweiBillItemHeaders is the headers of GCP bill item.
var HuaweiBillItemHeaders [][]string

func init() {
	var err error
	HuaweiBillItemHeaders, err = HuaweiBillItemTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("GetHuaweiHeader failed: %v", err)
	}
}

var _ table.Table = (*HuaweiBillItemTable)(nil)

// HuaweiBillItemTable huawei账单导出表结构
type HuaweiBillItemTable struct {
	Site        string `header:"站点类型"`
	AccountDate string `header:"核算年月"`

	BizID   string `header:"业务"`
	BizName string `header:"业务名称"`

	RootAccountName string `header:"一级账号名称"`
	MainAccountName string `header:"二级账号名称"`
	RegionName      string `header:"地域"`

	ProductName          string `header:"产品名称"`
	Region               string `header:"云服务区名称"`
	MeasureID            string `header:"金额单位"`
	UsageType            string `header:"使用量类型"`
	UsageMeasureID       string `header:"使用量度量单位"`
	CloudServiceType     string `header:"云服务类型编码"`
	CloudServiceTypeName string `header:"云服务类型名称"`
	ResourceType         string `header:"资源类型编码"`
	ResourceTypeName     string `header:"资源类型名称"`
	ChargeMode           string `header:"计费模式"`
	BillType             string `header:"账单类型"`
	FreeResourceUsage    string `header:"套餐内使用量"`
	Usage                string `header:"使用量"`
	RiUsage              string `header:"预留实例使用量"`
	Currency             string `header:"币种"`
	ExchangeRate         string `header:"汇率"`
	Cost                 string `header:"本期应付外币金额（元）"`
	CostRMB              string `header:"本期应付人民币金额（元）"`
}

// GetHeaders ...
func (h HuaweiBillItemTable) GetHeaders() ([][]string, error) {
	return table.GetHeaders(h)
}

// GetValuesByHeader ...
func (h HuaweiBillItemTable) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(h)
}
