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

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	hcproto "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// InitRouteTableService initialize the route table service.
func InitRouteTableService(c *capability.Capability) {
	svc := &routeTableSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("GetRouteTable", "GET", "/route_tables/{id}", svc.GetRouteTable)
	h.Add("ListRouteTable", "POST", "/route_tables/list", svc.ListRouteTable)
	h.Add("CountRouteTableSubnets", "POST", "/route_tables/subnets/count", svc.CountRouteTableSubnets)
	h.Add("AssignRouteTableToBiz", "POST", "/route_tables/assign/bizs", svc.AssignRouteTableToBiz)

	h.Add("ListRoute", "POST", "/vendors/{vendor}/route_tables/{route_table_id}/routes/list", svc.ListRoute)

	// route table & route apis in biz
	h.Add("GetBizRouteTable", "GET", "/bizs/{bk_biz_id}/route_tables/{id}", svc.GetBizRouteTable)
	h.Add("ListBizRouteTable", "POST", "/bizs/{bk_biz_id}/route_tables/list", svc.ListBizRouteTable)
	h.Add("CountBizRTSubnets", "POST", "/bizs/{bk_biz_id}/route_tables/subnets/count", svc.CountBizRTSubnets)

	h.Add("ListBizRoute", "POST", "/bizs/{bk_biz_id}/vendors/{vendor}/route_tables/{route_table_id}/routes/list",
		svc.ListBizRoute)

	h.Load(c.WebService)
}

type routeTableSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// UpdateRouteTable update route table.
func (svc *routeTableSvc) UpdateRouteTable(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.RouteTableUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.RouteTableCloudResType, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RouteTable, Action: meta.Update,
		ResourceID: basicInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// update route table
	switch basicInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.RouteTable.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.Aws:
		err = svc.client.HCService().Aws.RouteTable.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.Gcp:
		updateReq := &hcproto.RouteTableUpdateReq{
			Memo: req.Memo,
		}
		err = svc.client.HCService().Gcp.RouteTable.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Azure:
		err = svc.client.HCService().Azure.RouteTable.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.HuaWei:
		updateReq := &hcproto.RouteTableUpdateReq{
			Memo: req.Memo,
		}
		err = svc.client.HCService().HuaWei.RouteTable.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetRouteTable get route table details.
func (svc *routeTableSvc) GetRouteTable(cts *rest.Contexts) (interface{}, error) {
	return svc.getRouteTable(cts, handler.ResOperateAuth)
}

// GetBizRouteTable get biz route table details.
func (svc *routeTableSvc) GetBizRouteTable(cts *rest.Contexts) (interface{}, error) {
	return svc.getRouteTable(cts, handler.BizOperateAuth)
}

func (svc *routeTableSvc) getRouteTable(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.RouteTableCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.RouteTable,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// get route table detail info
	switch basicInfo.Vendor {
	case enumor.TCloud:
		routeTable, err := svc.client.DataService().TCloud.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return routeTable, err
	case enumor.Aws:
		routeTable, err := svc.client.DataService().Aws.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return routeTable, err
	case enumor.Gcp:
		listReq := &core.ListReq{
			Filter: tools.EqualExpression("id", id),
			Page:   &core.BasePage{Limit: 1},
		}

		routeTableRes, err := svc.client.DataService().Global.RouteTable.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
		if err != nil {
			return nil, err
		}

		if len(routeTableRes.Details) != 1 {
			return nil, errf.New(errf.InvalidParameter, "route table not exists")
		}

		return routeTableRes.Details[0], err
	case enumor.HuaWei:
		routeTable, err := svc.client.DataService().HuaWei.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return routeTable, err
	case enumor.Azure:
		routeTable, err := svc.client.DataService().Azure.RouteTable.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return routeTable, err
	}

	return nil, nil
}

// ListRouteTable list route table.
func (svc *routeTableSvc) ListRouteTable(cts *rest.Contexts) (interface{}, error) {
	return svc.listRouteTable(cts, handler.ListResourceAuthRes)
}

// ListBizRouteTable list biz route table.
func (svc *routeTableSvc) ListBizRouteTable(cts *rest.Contexts) (interface{}, error) {
	return svc.listRouteTable(cts, handler.ListBizAuthRes)
}

func (svc *routeTableSvc) listRouteTable(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.RouteTable, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &cloudserver.RouteTableListResult{Count: 0, Details: make([]corecloud.BaseRouteTable, 0)}, nil
	}
	req.Filter = expr

	// list route tables
	res, err := svc.client.DataService().Global.RouteTable.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &cloudserver.RouteTableListResult{Count: res.Count, Details: res.Details}, nil
}

// CountRouteTableSubnets count subnets in route tables. **NOTICE** only for ui.
func (svc *routeTableSvc) CountRouteTableSubnets(cts *rest.Contexts) (interface{}, error) {
	return svc.countRouteTableSubnets(cts, handler.ResOperateAuth)
}

// CountBizRTSubnets count subnets in route tables for biz use. **NOTICE** only for ui.
func (svc *routeTableSvc) CountBizRTSubnets(cts *rest.Contexts) (interface{}, error) {
	return svc.countRouteTableSubnets(cts, handler.BizOperateAuth)
}

func (svc *routeTableSvc) countRouteTableSubnets(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.CountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.RouteTableCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.RouteTable,
		Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// get route tables' subnet counts
	countReq := &dataservice.CountReq{Filter: tools.ContainersExpression("route_table_id", req.IDs)}
	result, err := svc.client.DataService().Global.RouteTable.CountSubnets(cts.Kit.Ctx, cts.Kit.Header(), countReq)
	if err != nil {
		logs.Errorf("count assigned route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// AssignRouteTableToBiz assign route tables to biz.
func (svc *routeTableSvc) AssignRouteTableToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.AssignRouteTableToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	err := svc.authorizeRouteTableAssignOp(cts.Kit, req.RouteTableIDs, req.BkBizID)
	if err != nil {
		return nil, err
	}

	// check if all route tables are not assigned to biz, right now assigning resource twice is not allowed
	routeTableFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.RouteTableIDs}
	err = svc.checkRouteTablesInBiz(cts.Kit, routeTableFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.RouteTableAuditResType, req.RouteTableIDs, req.BkBizID)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update route table biz relations
	createReq := &routetable.RouteTableBaseInfoBatchUpdateReq{
		RouteTables: []routetable.RouteTableBaseInfoUpdateReq{{
			IDs: req.RouteTableIDs,
			Data: &routetable.RouteTableUpdateBaseInfo{
				BkBizID: req.BkBizID,
			},
		}},
	}

	err = svc.client.DataService().Global.RouteTable.BatchUpdateBaseInfo(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *routeTableSvc) authorizeRouteTableAssignOp(kt *kit.Kit, ids []string, bizID int64) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.RouteTableCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
	if err != nil {
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RouteTable, Action: meta.Assign,
			ResourceID: info.AccountID}, BizID: bizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}

// checkRouteTablesInBiz check if routeTables are in the specified biz.
func (svc *routeTableSvc) checkRouteTablesInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Global.RouteTable.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count route tables that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d route tables are already assigned", result.Count)
	}

	return nil
}
