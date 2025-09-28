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

// AwsBillItemHeaders is the headers of Aws bill item.
var AwsBillItemHeaders []string

func init() {
	var err error
	AwsBillItemHeaders, err = AwsBillItemTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("GetAwsHeader failed: %v", err)
	}
}

var _ Table = (*AwsBillItemTable)(nil)

// AwsBillItemTable aws账单导出表结构
type AwsBillItemTable struct {
	Site        string `header:"站点类型"`
	AccountDate string `header:"核算年月"`

	BizID   string `header:"业务"`
	BizName string `header:"业务名称"`

	RootAccountName string `header:"一级账号名称"`
	MainAccountName string `header:"二级账号名称"`
	Region          string `header:"地域"`

	LocationName        string `header:"地区名称"`
	BillInvoiceIC       string `header:"发票ID"`
	BillEntity          string `header:"账单实体"`
	ProductCode         string `header:"产品代号"`
	ProductFamily       string `header:"服务组"`
	ProductName         string `header:"产品名称"`
	ApiOperation        string `header:"API操作"`
	ProductUsageType    string `header:"产品规格"`
	InstanceType        string `header:"实例类型"`
	ResourceId          string `header:"资源ID"`
	PricingTerm         string `header:"计费方式"`
	LineItemType        string `header:"计费类型"`
	LineItemDescription string `header:"计费说明"`
	UsageAmount         string `header:"用量"`
	PricingUnit         string `header:"单位"`
	Cost                string `header:"折扣前成本（外币）"`
	Currency            string `header:"外币种类"`
	RMBCost             string `header:"人民币成本（元）"`
	Rate                string `header:"汇率"`
}

// GetHeaders ...
func (c AwsBillItemTable) GetHeaders() ([]string, error) {
	return parseHeader(c)
}

// GetHeaderValues 获取表头对应的数据
func (c AwsBillItemTable) GetHeaderValues() ([]string, error) {
	return parseHeaderFields(c)
}
