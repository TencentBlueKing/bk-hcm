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
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/route-table"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// initHuaWeiRouteService initialize the huawei route service.
func initHuaWeiRouteService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	// TODO confirm if we should allow batch operation without route table id
	h.Path("/vendors/huawei/route_tables/{route_table_id}/routes")

	h.Add("BatchCreateHuaWeiRoute", "POST", "/batch/create", svc.BatchCreateHuaWeiRoute)
	h.Add("BatchUpdateHuaWeiRoute", "PATCH", "/batch", svc.BatchUpdateHuaWeiRoute)
	h.Add("ListHuaWeiRoute", "POST", "/list", svc.ListHuaWeiRoute)
	h.Add("ListAllHuaWeiRoute", "POST", "/list/all", svc.ListAllHuaWeiRoute)
	h.Add("BatchDeleteHuaWeiRoute", "DELETE", "/batch", svc.BatchDeleteHuaWeiRoute)

	h.Load(cap.WebService)
}

// BatchCreateHuaWeiRoute batch create route.
func (svc *routeTableSvc) BatchCreateHuaWeiRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.HuaWeiRouteBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check if all routes are in the route table
	cloudTableID := req.HuaWeiRoutes[0].CloudRouteTableID
	for _, createReq := range req.HuaWeiRoutes {
		if createReq.CloudRouteTableID != cloudTableID {
			return nil, errf.New(errf.InvalidParameter, "cloud route table ids are not the same")
		}
	}

	tableOpt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: tableID},
				filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudTableID},
			},
		},
		Page: &core.BasePage{Count: true},
	}
	tableRes, err := svc.dao.RouteTable().List(cts.Kit, tableOpt)
	if err != nil {
		logs.Errorf("validate route table(%s/%s) failed, err: %v, rid: %s", tableID, cloudTableID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if tableRes.Count != 1 {
		return nil, errf.New(errf.RecordNotFound, "route table not exists")
	}

	// add routes
	routeIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		routes := make([]tablecloud.HuaWeiRouteTable, 0, len(req.HuaWeiRoutes))
		for _, createReq := range req.HuaWeiRoutes {
			route := tablecloud.HuaWeiRouteTable{
				RouteTableID:      tableID,
				CloudRouteTableID: cloudTableID,
				Type:              createReq.Type,
				Destination:       createReq.Destination,
				NextHop:           createReq.NextHop,
				Memo:              createReq.Memo,
				Creator:           cts.Kit.User,
				Reviser:           cts.Kit.User,
			}

			routes = append(routes, route)
		}

		routeID, err := svc.dao.Route().HuaWei().BatchCreateWithTx(cts.Kit, txn, routes)
		if err != nil {
			return nil, fmt.Errorf("create huawei route failed, err: %v", err)
		}

		return routeID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create huawei route but return ids type %s is not string array",
			reflect.TypeOf(routeIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateHuaWeiRoute batch update route.
func (svc *routeTableSvc) BatchUpdateHuaWeiRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.HuaWeiRouteBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.HuaWeiRoutes))
	for _, route := range req.HuaWeiRoutes {
		ids = append(ids, route.ID)
	}

	// check if all routes exists in route table
	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids},
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: tableID},
			},
		},
		Page: &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Route().HuaWei().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei route failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.HuaWeiRoutes)) {
		return nil, fmt.Errorf("list huawei route failed, some route(ids=%+v) doesn't exist", ids)
	}

	// update route
	route := &tablecloud.HuaWeiRouteTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.HuaWeiRoutes {
		route.Type = updateReq.Type
		route.Destination = updateReq.Destination
		route.NextHop = updateReq.NextHop
		route.Memo = updateReq.Memo

		err = svc.dao.Route().HuaWei().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), route)
		if err != nil {
			logs.Errorf("update huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update huawei route failed, err: %v", err)
		}
	}
	return nil, nil
}

// ListHuaWeiRoute list routes.
func (svc *routeTableSvc) ListHuaWeiRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: tableID},
				req.Filter,
			},
		},
		Page:   req.Page,
		Fields: req.Fields,
	}

	daoHuaWeiRouteResp, err := svc.dao.Route().HuaWei().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.HuaWeiRouteListResult{Count: daoHuaWeiRouteResp.Count}, nil
	}

	details := make([]protocore.HuaWeiRoute, 0, len(daoHuaWeiRouteResp.Details))
	for _, route := range daoHuaWeiRouteResp.Details {
		details = append(details, protocore.HuaWeiRoute{
			ID:                route.ID,
			RouteTableID:      route.RouteTableID,
			CloudRouteTableID: route.CloudRouteTableID,
			Type:              route.Type,
			Destination:       route.Destination,
			NextHop:           route.NextHop,
			Memo:              route.Memo,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt,
				UpdatedAt: route.UpdatedAt,
			},
		})
	}

	return &protocloud.HuaWeiRouteListResult{Details: details}, nil
}

// ListAllHuaWeiRoute list routes.
func (svc *routeTableSvc) ListAllHuaWeiRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.HuaWeiRouteListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	if len(req.RouteTableID) != 0 {
		opt.Filter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: req.RouteTableID},
				req.Filter,
			},
		}
	}

	daoHuaWeiRouteResp, err := svc.dao.Route().HuaWei().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.HuaWeiRouteListResult{Count: daoHuaWeiRouteResp.Count}, nil
	}

	details := make([]protocore.HuaWeiRoute, 0, len(daoHuaWeiRouteResp.Details))
	for _, route := range daoHuaWeiRouteResp.Details {
		details = append(details, protocore.HuaWeiRoute{
			ID:                route.ID,
			RouteTableID:      route.RouteTableID,
			CloudRouteTableID: route.CloudRouteTableID,
			Type:              route.Type,
			Destination:       route.Destination,
			NextHop:           route.NextHop,
			Memo:              route.Memo,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt,
				UpdatedAt: route.UpdatedAt,
			},
		})
	}

	return &protocloud.HuaWeiRouteListResult{Details: details}, nil
}

// BatchDeleteHuaWeiRoute batch delete routes.
func (svc *routeTableSvc) BatchDeleteHuaWeiRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: tableID},
				req.Filter,
			},
		},
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.Route().HuaWei().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei route failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delHuaWeiRouteIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delHuaWeiRouteIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delHuaWeiRouteFilter := tools.ContainersExpression("id", delHuaWeiRouteIDs)
		if err := svc.dao.Route().HuaWei().BatchDeleteWithTx(cts.Kit, txn, delHuaWeiRouteFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete huawei route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
