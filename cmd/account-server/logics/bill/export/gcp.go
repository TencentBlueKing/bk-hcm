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

// GcpBillItemHeaders is the headers of GCP bill item.
var GcpBillItemHeaders []string

func init() {
	var err error
	GcpBillItemHeaders, err = GcpBillItemTable{}.GetHeaders()
	if err != nil {
		logs.Errorf("GetGcpHeader failed: %v", err)
	}
}

var _ Table = (*GcpBillItemTable)(nil)

// GcpBillItemTable is the table structure of GCP bill item.
type GcpBillItemTable struct {
	Site        string `header:"站点类型"`
	AccountDate string `header:"核算年月"`

	BizID   string `header:"业务"`
	BizName string `header:"业务名称"`

	RootAccountName string `header:"一级账号名称"`
	MainAccountName string `header:"二级账号名称"`
	Region          string `header:"地域"`

	RegionName                 string `header:"Region位置"`
	ProjectID                  string `header:"项目ID"`
	ProjectName                string `header:"项目名称"`
	ServiceCategory            string `header:"服务分类"`
	ServiceCategoryDescription string `header:"服务分类名称"`
	SkuDescription             string `header:"Sku名称"`
	Currency                   string `header:"外币类型"`
	UsageUnit                  string `header:"用量单位"`
	UsageAmount                string `header:"用量"`
	Cost                       string `header:"外币成本(元)"`
	ExchangeRate               string `header:"汇率"`
	RMBCost                    string `header:"人民币成本(元)"`
}

// GetHeaderValues ...
func (g GcpBillItemTable) GetHeaderValues() ([]string, error) {
	return parseHeaderFields(g)
}

// GetHeaders ...
func (g GcpBillItemTable) GetHeaders() ([]string, error) {
	return parseHeader(g)
}
