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
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListRoute list routes.
func (svc *routeTableSvc) ListRoute(cts *rest.Contexts) (interface{}, error) {
	return svc.listRoute(cts, handler.ResOperateAuth)
}

// ListBizRoute list biz routes.
func (svc *routeTableSvc) ListBizRoute(cts *rest.Contexts) (interface{}, error) {
	return svc.listRoute(cts, handler.BizOperateAuth)
}

func (svc *routeTableSvc) listRoute(cts *rest.Contexts, validator handler.ValidWithAuthHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	tableID := cts.PathParameter("route_table_id").String()
	if len(tableID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.RouteTableCloudResType, tableID)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validator(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.RouteTable,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// list routes
	switch vendor {
	case enumor.TCloud:
		res, err := svc.client.DataService().TCloud.RouteTable.ListRoute(cts.Kit.Ctx, cts.Kit.Header(), tableID, req)
		if err != nil {
			logs.Errorf("list tcloud route failed, err: %v, table id: %s, rid: %s", err, tableID, cts.Kit.Rid)
			return nil, err
		}
		return res, nil
	case enumor.Aws:
		res, err := svc.client.DataService().Aws.RouteTable.ListRoute(cts.Kit.Ctx, cts.Kit.Header(), tableID, req)
		if err != nil {
			logs.Errorf("list aws route failed, err: %v, table id: %s, rid: %s", err, tableID, cts.Kit.Rid)
			return nil, err
		}
		return res, nil
	case enumor.Azure:
		res, err := svc.client.DataService().Azure.RouteTable.ListRoute(cts.Kit.Ctx, cts.Kit.Header(), tableID, req)
		if err != nil {
			logs.Errorf("list azure route failed, err: %v, table id: %s, rid: %s", err, tableID, cts.Kit.Rid)
			return nil, err
		}
		return res, nil
	case enumor.HuaWei:
		res, err := svc.client.DataService().HuaWei.RouteTable.ListRoute(cts.Kit.Ctx, cts.Kit.Header(), tableID, req)
		if err != nil {
			logs.Errorf("list huawei route failed, err: %v, table id: %s, rid: %s", err, tableID, cts.Kit.Rid)
			return nil, err
		}
		return res, nil
	case enumor.Gcp:
		// TODO confirm if gcp list route operation needs route table id
		req := &routetable.GcpRouteListReq{
			ListReq: &core.ListReq{
				Filter: req.Filter,
				Page:   req.Page,
			},
			RouteTableID: tableID,
		}
		res, err := svc.client.DataService().Gcp.RouteTable.ListRoute(cts.Kit.Ctx, cts.Kit.Header(), req)
		if err != nil {
			logs.Errorf("list gcp route failed, err: %v, table id: %s, rid: %s", err, tableID, cts.Kit.Rid)
			return nil, err
		}
		return res, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported cloud vendor: %s", vendor)
	}
}
