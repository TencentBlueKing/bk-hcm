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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// initTCloudRouteService initialize the tcloud route service.
func initTCloudRouteService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	// TODO confirm if we should allow batch operation without route table id
	h.Path("/vendors/tcloud/route_tables/{route_table_id}/routes")

	h.Add("BatchCreateTCloudRoute", "POST", "/batch/create", svc.BatchCreateTCloudRoute)
	h.Add("BatchUpdateTCloudRoute", "PATCH", "/batch", svc.BatchUpdateTCloudRoute)
	h.Add("ListTCloudRoute", "POST", "/list", svc.ListTCloudRoute)
	h.Add("ListAllTCloudRoute", "POST", "/list/all", svc.ListAllTCloudRoute)
	h.Add("BatchDeleteTCloudRoute", "DELETE", "/batch", svc.BatchDeleteTCloudRoute)

	h.Load(cap.WebService)
}

// BatchCreateTCloudRoute batch create route.
func (svc *routeTableSvc) BatchCreateTCloudRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.TCloudRouteBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check if all routes are in the route table
	cloudTableID := req.TCloudRoutes[0].CloudRouteTableID
	for _, createReq := range req.TCloudRoutes {
		if createReq.CloudRouteTableID != cloudTableID {
			return nil, errf.New(errf.InvalidParameter, "cloud route table ids are not the same")
		}
	}

	if err := svc.validateTCloudRouteTable(cts.Kit, tableID, cloudTableID); err != nil {
		return nil, err
	}

	// add routes
	routeIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		routes := make([]tablecloud.TCloudRouteTable, 0, len(req.TCloudRoutes))
		for _, createReq := range req.TCloudRoutes {
			route := tablecloud.TCloudRouteTable{
				CloudID:                  createReq.CloudID,
				RouteTableID:             tableID,
				CloudRouteTableID:        cloudTableID,
				DestinationCidrBlock:     createReq.DestinationCidrBlock,
				DestinationIpv6CidrBlock: createReq.DestinationIpv6CidrBlock,
				GatewayType:              createReq.GatewayType,
				CloudGatewayID:           createReq.CloudGatewayID,
				Enabled:                  &createReq.Enabled,
				RouteType:                createReq.RouteType,
				PublishedToVbc:           &createReq.PublishedToVbc,
				Memo:                     createReq.Memo,
				Creator:                  cts.Kit.User,
				Reviser:                  cts.Kit.User,
			}

			routes = append(routes, route)
		}

		routeID, err := svc.dao.Route().TCloud().BatchCreateWithTx(cts.Kit, txn, routes)
		if err != nil {
			return nil, fmt.Errorf("create tcloud route failed, err: %v", err)
		}

		return routeID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create tcloud route but return ids type %s is not string array",
			reflect.TypeOf(routeIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *routeTableSvc) validateTCloudRouteTable(kt *kit.Kit, tableID string, cloudTableID string) error {
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
	tableRes, err := svc.dao.RouteTable().List(kt, tableOpt)
	if err != nil {
		logs.Errorf("validate route table(%s/%s) failed, err: %v, rid: %s", tableID, cloudTableID, err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	if tableRes.Count != 1 {
		return errf.New(errf.RecordNotFound, "route table not exists")
	}
	return nil
}

// BatchUpdateTCloudRoute batch update route.
func (svc *routeTableSvc) BatchUpdateTCloudRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.TCloudRouteBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.TCloudRoutes))
	for _, route := range req.TCloudRoutes {
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
	listRes, err := svc.dao.Route().TCloud().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud route failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.TCloudRoutes)) {
		return nil, fmt.Errorf("list tcloud route failed, some route(ids=%+v) doesn't exist", ids)
	}

	// update route
	route := &tablecloud.TCloudRouteTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.TCloudRoutes {
		route.DestinationCidrBlock = updateReq.DestinationCidrBlock
		route.DestinationIpv6CidrBlock = updateReq.DestinationIpv6CidrBlock
		route.GatewayType = updateReq.GatewayType
		route.CloudGatewayID = updateReq.CloudGatewayID
		route.Enabled = updateReq.Enabled
		route.RouteType = updateReq.RouteType
		route.PublishedToVbc = updateReq.PublishedToVbc
		route.Memo = updateReq.Memo

		err = svc.dao.Route().TCloud().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), route)
		if err != nil {
			logs.Errorf("update tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update tcloud route failed, err: %v", err)
		}
	}
	return nil, nil
}

// ListTCloudRoute list routes.
func (svc *routeTableSvc) ListTCloudRoute(cts *rest.Contexts) (interface{}, error) {
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

	daoTCloudRouteResp, err := svc.dao.Route().TCloud().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.TCloudRouteListResult{Count: daoTCloudRouteResp.Count}, nil
	}

	details := make([]protocore.TCloudRoute, 0, len(daoTCloudRouteResp.Details))
	for _, route := range daoTCloudRouteResp.Details {
		details = append(details, protocore.TCloudRoute{
			ID:                       route.ID,
			RouteTableID:             route.RouteTableID,
			CloudID:                  route.CloudID,
			CloudRouteTableID:        route.CloudRouteTableID,
			DestinationCidrBlock:     route.DestinationCidrBlock,
			DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
			GatewayType:              route.GatewayType,
			CloudGatewayID:           route.CloudGatewayID,
			Enabled:                  converter.PtrToVal(route.Enabled),
			RouteType:                route.RouteType,
			PublishedToVbc:           converter.PtrToVal(route.PublishedToVbc),
			Memo:                     route.Memo,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.TCloudRouteListResult{Details: details}, nil
}

// ListAllTCloudRoute list routes.
func (svc *routeTableSvc) ListAllTCloudRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.TCloudRouteListReq)
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

	daoTCloudRouteResp, err := svc.dao.Route().TCloud().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.TCloudRouteListResult{Count: daoTCloudRouteResp.Count}, nil
	}

	details := make([]protocore.TCloudRoute, 0, len(daoTCloudRouteResp.Details))
	for _, route := range daoTCloudRouteResp.Details {
		details = append(details, protocore.TCloudRoute{
			ID:                       route.ID,
			RouteTableID:             route.RouteTableID,
			CloudID:                  route.CloudID,
			CloudRouteTableID:        route.CloudRouteTableID,
			DestinationCidrBlock:     route.DestinationCidrBlock,
			DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
			GatewayType:              route.GatewayType,
			CloudGatewayID:           route.CloudGatewayID,
			Enabled:                  converter.PtrToVal(route.Enabled),
			RouteType:                route.RouteType,
			PublishedToVbc:           converter.PtrToVal(route.PublishedToVbc),
			Memo:                     route.Memo,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.TCloudRouteListResult{Details: details}, nil
}

// BatchDeleteTCloudRoute batch delete routes.
func (svc *routeTableSvc) BatchDeleteTCloudRoute(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.Route().TCloud().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud route failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delTCloudRouteIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delTCloudRouteIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delTCloudRouteFilter := tools.ContainersExpression("id", delTCloudRouteIDs)
		if err := svc.dao.Route().TCloud().BatchDeleteWithTx(cts.Kit, txn, delTCloudRouteFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete tcloud route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
