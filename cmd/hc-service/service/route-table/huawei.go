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

// Package routetable defines route table service.
package routetable

import (
	adcore "hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	hcservice "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// HuaWeiRouteTableUpdate update huawei route table.
func (r routeTable) HuaWeiRouteTableUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.RouteTableUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := r.cs.DataService().HuaWei.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := r.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &routetable.HuaWeiRouteTableUpdateOption{
		RouteTableUpdateOption: routetable.RouteTableUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &routetable.BaseRouteTableUpdateData{Memo: req.Memo},
		},
		Region: getRes.Region,
	}
	err = cli.UpdateRouteTable(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &dataproto.RouteTableBaseInfoBatchUpdateReq{
		RouteTables: []dataproto.RouteTableBaseInfoUpdateReq{{
			IDs: []string{id},
			Data: &dataproto.RouteTableUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = r.cs.DataService().Global.RouteTable.BatchUpdateBaseInfo(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiRouteTableDelete delete huawei route table.
func (r routeTable) HuaWeiRouteTableDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := r.cs.DataService().HuaWei.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := r.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Region,
	}
	err = cli.DeleteRouteTable(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = r.cs.DataService().Global.RouteTable.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
