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

package handlers

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloudzone "hcm/pkg/api/core/cloud/zone"
	dataprotozone "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
)

// GetZone 查询可用区
func (a *BaseApplicationHandler) GetZone(vendor enumor.Vendor, region, zone string) (*corecloudzone.BaseZone, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			filter.AtomRule{Field: "name", Op: filter.Equal.Factory(), Value: zone},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Zone.ListZone(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotozone.ZoneListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s zone by region(%s) and zone cloud_id(%s)", vendor, region, zone)
	}

	return &resp.Details[0], nil
}

// GetZones 查询多个可用区
func (a *BaseApplicationHandler) GetZones(vendor enumor.Vendor, region string, zoneNames []string) (
	[]corecloudzone.BaseZone, error) {

	reqFilter := tools.ExpressionAnd(
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("region", region),
		tools.RuleIn("name", zoneNames),
	)
	// 查询
	resp, err := a.Client.DataService().Global.Zone.ListZone(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotozone.ZoneListReq{
			Filter: reqFilter,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: uint(len(zoneNames))},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s zone by region(%s) and zone names(%v)", vendor, region, zoneNames)
	}

	return resp.Details, nil
}
