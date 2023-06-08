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

package routetable

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablert "hcm/pkg/dal/table/cloud/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// NewRouteTable new routeTable.
func NewRouteTable(dao dao.Set) *RouteTable {
	return &RouteTable{
		dao: dao,
	}
}

// RouteTable define routeTable audit.
type RouteTable struct {
	dao dao.Set
}

// RouteTableAssignAuditBuild build route table assign audit.
func (ad *RouteTable) RouteTableAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idRouteTableMap, err := ListRouteTable(kt, ad.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		routeTable, exist := idRouteTableMap[one.ResID]
		if !exist {
			continue
		}

		var action enumor.AuditAction
		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			action = enumor.Assign
		case enumor.DeliverAssignedResType:
			action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: routeTable.CloudID,
			ResName:    converter.PtrToVal(routeTable.Name),
			ResType:    enumor.RouteTableAuditResType,
			Action:     action,
			BkBizID:    routeTable.BkBizID,
			Vendor:     routeTable.Vendor,
			AccountID:  routeTable.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		})
	}

	return audits, nil
}

// ListRouteTable list routeTable.
func ListRouteTable(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablert.RouteTableTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.RouteTable().List(kt, opt)
	if err != nil {
		logs.Errorf("list routeTables failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablert.RouteTableTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
