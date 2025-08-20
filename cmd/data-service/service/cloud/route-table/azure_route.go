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

// initAzureRouteService initialize the azure route service.
func initAzureRouteService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	// TODO confirm if we should allow batch operation without route table id
	h.Path("/vendors/azure/route_tables/{route_table_id}/routes")

	h.Add("BatchCreateAzureRoute", "POST", "/batch/create", svc.BatchCreateAzureRoute)
	h.Add("BatchUpdateAzureRoute", "PATCH", "/batch", svc.BatchUpdateAzureRoute)
	h.Add("ListAzureRoute", "POST", "/list", svc.ListAzureRoute)
	h.Add("ListAllAzureRoute", "POST", "/list/all", svc.ListAllAzureRoute)
	h.Add("BatchDeleteAzureRoute", "DELETE", "/batch", svc.BatchDeleteAzureRoute)

	h.Load(cap.WebService)
}

// BatchCreateAzureRoute batch create route.
func (svc *routeTableSvc) BatchCreateAzureRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.AzureRouteBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check if all routes are in the route table
	cloudTableID := req.AzureRoutes[0].CloudRouteTableID
	for _, createReq := range req.AzureRoutes {
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
		routes := make([]tablecloud.AzureRouteTable, 0, len(req.AzureRoutes))
		for _, createReq := range req.AzureRoutes {
			route := tablecloud.AzureRouteTable{
				CloudID:           createReq.CloudID,
				RouteTableID:      tableID,
				CloudRouteTableID: cloudTableID,
				Name:              createReq.Name,
				AddressPrefix:     createReq.AddressPrefix,
				NextHopType:       createReq.NextHopType,
				NextHopIPAddress:  createReq.NextHopIPAddress,
				ProvisioningState: createReq.ProvisioningState,
				Creator:           cts.Kit.User,
				Reviser:           cts.Kit.User,
			}

			routes = append(routes, route)
		}

		routeID, err := svc.dao.Route().Azure().BatchCreateWithTx(cts.Kit, txn, routes)
		if err != nil {
			return nil, fmt.Errorf("create azure route failed, err: %v", err)
		}

		return routeID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create azure route but return ids type %s is not string array",
			reflect.TypeOf(routeIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateAzureRoute batch update route.
func (svc *routeTableSvc) BatchUpdateAzureRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(protocloud.AzureRouteBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.AzureRoutes))
	for _, route := range req.AzureRoutes {
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
	listRes, err := svc.dao.Route().Azure().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure route failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.AzureRoutes)) {
		return nil, fmt.Errorf("list azure route failed, some route(ids=%+v) doesn't exist", ids)
	}

	// update route
	route := &tablecloud.AzureRouteTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.AzureRoutes {
		route.AddressPrefix = updateReq.AddressPrefix
		route.NextHopType = updateReq.NextHopType
		route.NextHopIPAddress = updateReq.NextHopIPAddress
		route.ProvisioningState = updateReq.ProvisioningState

		err = svc.dao.Route().Azure().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), route)
		if err != nil {
			logs.Errorf("update azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update azure route failed, err: %v", err)
		}
	}
	return nil, nil
}

// ListAzureRoute list routes.
func (svc *routeTableSvc) ListAzureRoute(cts *rest.Contexts) (interface{}, error) {
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

	daoAzureRouteResp, err := svc.dao.Route().Azure().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AzureRouteListResult{Count: daoAzureRouteResp.Count}, nil
	}

	details := make([]protocore.AzureRoute, 0, len(daoAzureRouteResp.Details))
	for _, route := range daoAzureRouteResp.Details {
		details = append(details, protocore.AzureRoute{
			ID:                route.ID,
			CloudID:           route.CloudID,
			RouteTableID:      route.RouteTableID,
			CloudRouteTableID: route.CloudRouteTableID,
			Name:              route.Name,
			AddressPrefix:     route.AddressPrefix,
			NextHopType:       route.NextHopType,
			NextHopIPAddress:  route.NextHopIPAddress,
			ProvisioningState: route.ProvisioningState,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.AzureRouteListResult{Details: details}, nil
}

// ListAllAzureRoute list routes.
func (svc *routeTableSvc) ListAllAzureRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AzureRouteListReq)
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

	daoAzureRouteResp, err := svc.dao.Route().Azure().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AzureRouteListResult{Count: daoAzureRouteResp.Count}, nil
	}

	details := make([]protocore.AzureRoute, 0, len(daoAzureRouteResp.Details))
	for _, route := range daoAzureRouteResp.Details {
		details = append(details, protocore.AzureRoute{
			ID:                route.ID,
			CloudID:           route.CloudID,
			RouteTableID:      route.RouteTableID,
			CloudRouteTableID: route.CloudRouteTableID,
			Name:              route.Name,
			AddressPrefix:     route.AddressPrefix,
			NextHopType:       route.NextHopType,
			NextHopIPAddress:  route.NextHopIPAddress,
			ProvisioningState: route.ProvisioningState,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.AzureRouteListResult{Details: details}, nil
}

// BatchDeleteAzureRoute batch delete routes.
func (svc *routeTableSvc) BatchDeleteAzureRoute(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.Route().Azure().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure route failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delAzureRouteIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delAzureRouteIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delAzureRouteFilter := tools.ContainersExpression("id", delAzureRouteIDs)
		if err := svc.dao.Route().Azure().BatchDeleteWithTx(cts.Kit, txn, delAzureRouteFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete azure route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
