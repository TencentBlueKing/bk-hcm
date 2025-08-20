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

// initAwsRouteService initialize the aws route service.
func initAwsRouteService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	// TODO confirm if we should allow batch operation without route table id
	h.Path("/vendors/aws/route_tables/{route_table_id}/routes")

	h.Add("BatchCreateAwsRoute", "POST", "/batch/create", svc.BatchCreateAwsRoute)
	h.Add("BatchUpdateAwsRoute", "PATCH", "/batch", svc.BatchUpdateAwsRoute)
	h.Add("ListAwsRoute", "POST", "/list", svc.ListAwsRoute)
	h.Add("ListAllAwsRoute", "POST", "/list/all", svc.ListAllAwsRoute)
	h.Add("BatchDeleteAwsRoute", "DELETE", "/batch", svc.BatchDeleteAwsRoute)

	h.Load(cap.WebService)
}

// BatchCreateAwsRoute batch create route.
func (svc *routeTableSvc) BatchCreateAwsRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.AwsRouteBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check if all routes are in the route table
	cloudTableID := req.AwsRoutes[0].CloudRouteTableID
	for _, createReq := range req.AwsRoutes {
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

	ids, err := svc.addAwsRoute(cts.Kit, tableID, cloudTableID, req.AwsRoutes)
	if err != nil {
		logs.Errorf("create aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return &core.BatchCreateResult{IDs: ids}, nil
}

// addAwsRoute add aws route to the database.
func (svc *routeTableSvc) addAwsRoute(kt *kit.Kit, tableID string, cloudTableID string,
	createReqs []protocloud.AwsRouteCreateReq) ([]string, error) {

	// add routes
	routeIDs, err := svc.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		routes := make([]tablecloud.AwsRouteTable, 0, len(createReqs))
		for _, createReq := range createReqs {
			route := tablecloud.AwsRouteTable{
				RouteTableID:                     tableID,
				CloudRouteTableID:                cloudTableID,
				DestinationCidrBlock:             createReq.DestinationCidrBlock,
				DestinationIpv6CidrBlock:         createReq.DestinationIpv6CidrBlock,
				CloudDestinationPrefixListID:     createReq.CloudDestinationPrefixListID,
				CloudCarrierGatewayID:            createReq.CloudCarrierGatewayID,
				CoreNetworkArn:                   createReq.CoreNetworkArn,
				CloudEgressOnlyInternetGatewayID: createReq.CloudEgressOnlyInternetGatewayID,
				CloudGatewayID:                   createReq.CloudGatewayID,
				CloudInstanceID:                  createReq.CloudInstanceID,
				CloudInstanceOwnerID:             createReq.CloudInstanceOwnerID,
				CloudLocalGatewayID:              createReq.CloudLocalGatewayID,
				CloudNatGatewayID:                createReq.CloudNatGatewayID,
				CloudNetworkInterfaceID:          createReq.CloudNetworkInterfaceID,
				CloudTransitGatewayID:            createReq.CloudTransitGatewayID,
				CloudVpcPeeringConnectionID:      createReq.CloudVpcPeeringConnectionID,
				State:                            createReq.State,
				Propagated:                       &createReq.Propagated,
				Creator:                          kt.User,
				Reviser:                          kt.User,
			}

			routes = append(routes, route)
		}
		routeIDs, err := svc.dao.Route().Aws().BatchCreateWithTx(kt, txn, routes)
		if err != nil {
			return nil, fmt.Errorf("create aws route failed, err: %v", err)
		}

		return routeIDs, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create aws route but return ids type %s is not string array",
			reflect.TypeOf(routeIDs).String())
	}
	return ids, nil
}

// BatchUpdateAwsRoute batch update route.
func (svc *routeTableSvc) BatchUpdateAwsRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.AwsRouteBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.AwsRoutes))
	idsExists := make(map[string]struct{}, 0)
	for _, route := range req.AwsRoutes {
		ids = append(ids, route.ID)
		if _, ok := idsExists[route.ID]; !ok {
			idsExists[route.ID] = struct{}{}
		}
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
	listRes, err := svc.dao.Route().Aws().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws route failed, err: %v", err)
	}
	if listRes.Count != uint64(len(idsExists)) {
		return nil, fmt.Errorf("list aws route failed, some route(ids=%+v) doesn't exist", ids)
	}

	// update route
	route := &tablecloud.AwsRouteTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.AwsRoutes {
		route.CloudCarrierGatewayID = updateReq.CloudCarrierGatewayID
		route.CoreNetworkArn = updateReq.CoreNetworkArn
		route.CloudEgressOnlyInternetGatewayID = updateReq.CloudEgressOnlyInternetGatewayID
		route.CloudGatewayID = updateReq.CloudGatewayID
		route.CloudInstanceID = updateReq.CloudInstanceID
		route.CloudInstanceOwnerID = updateReq.CloudInstanceOwnerID
		route.CloudLocalGatewayID = updateReq.CloudLocalGatewayID
		route.CloudNatGatewayID = updateReq.CloudNatGatewayID
		route.CloudNetworkInterfaceID = updateReq.CloudNetworkInterfaceID
		route.CloudTransitGatewayID = updateReq.CloudTransitGatewayID
		route.CloudVpcPeeringConnectionID = updateReq.CloudVpcPeeringConnectionID
		route.State = updateReq.State
		route.Propagated = updateReq.Propagated

		err = svc.dao.Route().Aws().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), route)
		if err != nil {
			logs.Errorf("update aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update aws route failed, err: %v", err)
		}
	}
	return nil, nil
}

// ListAwsRoute list routes.
func (svc *routeTableSvc) ListAwsRoute(cts *rest.Contexts) (interface{}, error) {
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

	daoAwsRouteResp, err := svc.dao.Route().Aws().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AwsRouteListResult{Count: daoAwsRouteResp.Count}, nil
	}

	details := make([]protocore.AwsRoute, 0, len(daoAwsRouteResp.Details))
	for _, route := range daoAwsRouteResp.Details {
		details = append(details, protocore.AwsRoute{
			ID:                               route.ID,
			RouteTableID:                     route.RouteTableID,
			CloudRouteTableID:                route.CloudRouteTableID,
			DestinationCidrBlock:             route.DestinationCidrBlock,
			DestinationIpv6CidrBlock:         route.DestinationIpv6CidrBlock,
			CloudDestinationPrefixListID:     route.CloudDestinationPrefixListID,
			CloudCarrierGatewayID:            route.CloudCarrierGatewayID,
			CoreNetworkArn:                   route.CoreNetworkArn,
			CloudEgressOnlyInternetGatewayID: route.CloudEgressOnlyInternetGatewayID,
			CloudGatewayID:                   route.CloudGatewayID,
			CloudInstanceID:                  route.CloudInstanceID,
			CloudInstanceOwnerID:             route.CloudInstanceOwnerID,
			CloudLocalGatewayID:              route.CloudLocalGatewayID,
			CloudNatGatewayID:                route.CloudNatGatewayID,
			CloudNetworkInterfaceID:          route.CloudNetworkInterfaceID,
			CloudTransitGatewayID:            route.CloudTransitGatewayID,
			CloudVpcPeeringConnectionID:      route.CloudVpcPeeringConnectionID,
			State:                            route.State,
			Propagated:                       converter.PtrToVal(route.Propagated),
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.AwsRouteListResult{Details: details}, nil
}

// ListAllAwsRoute list routes.
func (svc *routeTableSvc) ListAllAwsRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AwsRouteListReq)
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

	daoAwsRouteResp, err := svc.dao.Route().Aws().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AwsRouteListResult{Count: daoAwsRouteResp.Count}, nil
	}

	details := make([]protocore.AwsRoute, 0, len(daoAwsRouteResp.Details))
	for _, route := range daoAwsRouteResp.Details {
		details = append(details, protocore.AwsRoute{
			ID:                               route.ID,
			RouteTableID:                     route.RouteTableID,
			CloudRouteTableID:                route.CloudRouteTableID,
			DestinationCidrBlock:             route.DestinationCidrBlock,
			DestinationIpv6CidrBlock:         route.DestinationIpv6CidrBlock,
			CloudDestinationPrefixListID:     route.CloudDestinationPrefixListID,
			CloudCarrierGatewayID:            route.CloudCarrierGatewayID,
			CoreNetworkArn:                   route.CoreNetworkArn,
			CloudEgressOnlyInternetGatewayID: route.CloudEgressOnlyInternetGatewayID,
			CloudGatewayID:                   route.CloudGatewayID,
			CloudInstanceID:                  route.CloudInstanceID,
			CloudInstanceOwnerID:             route.CloudInstanceOwnerID,
			CloudLocalGatewayID:              route.CloudLocalGatewayID,
			CloudNatGatewayID:                route.CloudNatGatewayID,
			CloudNetworkInterfaceID:          route.CloudNetworkInterfaceID,
			CloudTransitGatewayID:            route.CloudTransitGatewayID,
			CloudVpcPeeringConnectionID:      route.CloudVpcPeeringConnectionID,
			State:                            route.State,
			Propagated:                       converter.PtrToVal(route.Propagated),
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.AwsRouteListResult{Details: details}, nil
}

// BatchDeleteAwsRoute batch delete routes.
func (svc *routeTableSvc) BatchDeleteAwsRoute(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.Route().Aws().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list aws route failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delAwsRouteIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delAwsRouteIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delAwsRouteFilter := tools.ContainersExpression("id", delAwsRouteIDs)
		if err := svc.dao.Route().Aws().BatchDeleteWithTx(cts.Kit, txn, delAwsRouteFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete aws route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
