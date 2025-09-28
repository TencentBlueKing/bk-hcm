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
	"hcm/cmd/data-service/service/cloud/logics"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/route-table"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitRouteTableService initialize the route table service.
func InitRouteTableService(cap *capability.Capability) {
	svc := &routeTableSvc{
		dao: cap.Dao,
	}

	initRouteTableService(svc, cap)
	initTCloudRouteService(svc, cap)
	initAwsRouteService(svc, cap)
	initAzureRouteService(svc, cap)
	initHuaWeiRouteService(svc, cap)
	initGcpRouteService(svc, cap)

}

// initRouteTableService initialize the route table service.
func initRouteTableService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateRouteTable", "POST", "/vendors/{vendor}/route_tables/batch/create",
		svc.BatchCreateRouteTable)
	h.Add("BatchUpdateRouteTableBaseInfo", "PATCH", "/route_tables/base/batch",
		svc.BatchUpdateRouteTableBaseInfo)
	h.Add("GetRouteTable", "GET", "/vendors/{vendor}/route_tables/{id}", svc.GetRouteTable)
	h.Add("ListRouteTable", "POST", "/route_tables/list", svc.ListRouteTable)
	h.Add("ListRouteTableWithExtension", "POST", "/vendors/{vendor}/route_tables/list",
		svc.ListRouteTableWithExtension)
	h.Add("BatchDeleteRouteTable", "DELETE", "/route_tables/batch", svc.BatchDeleteRouteTable)
	h.Add("CountRouteTableSubnets", "POST", "/route_tables/subnets/count", svc.CountRouteTableSubnets)

	h.Load(cap.WebService)
}

type routeTableSvc struct {
	dao dao.Set
}

// TODO sync vpc id

// BatchCreateRouteTable batch create route table.
func (svc *routeTableSvc) BatchCreateRouteTable(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateRouteTable[protocloud.TCloudRouteTableCreateExt](cts, vendor, svc)
	case enumor.Aws:
		return batchCreateRouteTable[protocloud.AwsRouteTableCreateExt](cts, vendor, svc)
	case enumor.HuaWei:
		return batchCreateRouteTable[protocloud.HuaWeiRouteTableCreateExt](cts, vendor, svc)
	case enumor.Azure:
		return batchCreateRouteTable[protocloud.AzureRouteTableCreateExt](cts, vendor, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s is invalid", vendor)
	}
}

// batchCreateRouteTable batch create route table.
func batchCreateRouteTable[T protocloud.RouteTableCreateExtension](cts *rest.Contexts, vendor enumor.Vendor, svc *routeTableSvc) (
	interface{}, error) {

	req := new(protocloud.RouteTableBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get vpc cloud id to id mapping
	vpcCloudIDs := make([]string, 0)
	for _, subnet := range req.RouteTables {
		if len(subnet.CloudVpcID) != 0 {
			vpcCloudIDs = append(vpcCloudIDs, subnet.CloudVpcID)
		}
	}

	vpcIDMap, err := logics.GetVpcIDByCloudID(cts.Kit, svc.dao, vendor, vpcCloudIDs)
	if err != nil {
		return nil, err
	}

	routeTableIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		routeTables := make([]tablecloud.RouteTableTable, 0, len(req.RouteTables))
		for _, createReq := range req.RouteTables {
			ext, err := tabletype.NewJsonField(createReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			routeTable := tablecloud.RouteTableTable{
				Vendor:     vendor,
				AccountID:  createReq.AccountID,
				CloudID:    createReq.CloudID,
				CloudVpcID: createReq.CloudVpcID,
				Name:       createReq.Name,
				Region:     createReq.Region,
				Memo:       createReq.Memo,
				BkBizID:    createReq.BkBizID,
				Extension:  ext,
				Creator:    cts.Kit.User,
				Reviser:    cts.Kit.User,
			}

			vpcID, exists := vpcIDMap[createReq.CloudVpcID]
			if !exists {
				vpcID = constant.NotFoundVpc
			}

			routeTable.VpcID = vpcID

			routeTables = append(routeTables, routeTable)
		}

		routeTableID, err := svc.dao.RouteTable().BatchCreateWithTx(cts.Kit, txn, routeTables)
		if err != nil {
			return nil, fmt.Errorf("create route table failed, err: %v", err)
		}

		return routeTableID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeTableIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create route table but return ids type %s is not string array",
			reflect.TypeOf(routeTableIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateRouteTableBaseInfo batch update route table base info.
func (svc *routeTableSvc) BatchUpdateRouteTableBaseInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.RouteTableBaseInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0)
	for _, routeTable := range req.RouteTables {
		ids = append(ids, routeTable.IDs...)
	}

	// check if all route tables exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.RouteTable().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list route table failed, err: %v", err)
	}
	if listRes.Count != uint64(len(ids)) {
		return nil, fmt.Errorf("list route table failed, some route table(ids=%+v) doesn't exist", ids)
	}

	// update route table
	routeTable := &tablecloud.RouteTableTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.RouteTables {
		routeTable.Name = updateReq.Data.Name
		routeTable.Memo = updateReq.Data.Memo
		routeTable.BkBizID = updateReq.Data.BkBizID

		err = svc.dao.RouteTable().Update(cts.Kit, tools.ContainersExpression("id", updateReq.IDs), routeTable)
		if err != nil {
			logs.Errorf("update route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update route table failed, err: %v", err)
		}

	}

	return nil, nil
}

// GetRouteTable get route table details.
func (svc *routeTableSvc) GetRouteTable(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeTableID := cts.PathParameter("id").String()

	dbRouteTable, err := getRouteTableFromTable(cts.Kit, svc.dao, routeTableID)
	if err != nil {
		return nil, err
	}

	base := convertBaseRouteTable(dbRouteTable)

	switch vendor {
	case enumor.TCloud:
		return convertToRouteTableResult[protocore.TCloudRouteTableExtension](base, dbRouteTable.Extension)
	case enumor.Aws:
		return convertToRouteTableResult[protocore.AwsRouteTableExtension](base, dbRouteTable.Extension)
	case enumor.Gcp:
		return base, nil
	case enumor.HuaWei:
		return convertToRouteTableResult[protocore.HuaWeiRouteTableExtension](base, dbRouteTable.Extension)
	case enumor.Azure:
		return convertToRouteTableResult[protocore.AzureRouteTableExtension](base, dbRouteTable.Extension)
	}

	return nil, nil
}

func convertToRouteTableResult[T protocore.RouteTableExtension](baseRouteTable *protocore.BaseRouteTable,
	dbExtension tabletype.JsonField) (*protocore.RouteTable[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &protocore.RouteTable[T]{
		BaseRouteTable: *baseRouteTable,
		Extension:      extension,
	}, nil
}

func getRouteTableFromTable(kt *kit.Kit, dao dao.Set, routeTableID string) (*tablecloud.RouteTableTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", routeTableID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.RouteTable().List(kt, opt)
	if err != nil {
		logs.Errorf("list route table failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list route table failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list route table failed, route table(id=%s) doesn't exist", routeTableID)
	}

	return &details[0], nil
}

// ListRouteTable list route tables.
func (svc *routeTableSvc) ListRouteTable(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
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
	daoRouteTableResp, err := svc.dao.RouteTable().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list route table failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.RouteTableListResult{Count: daoRouteTableResp.Count}, nil
	}

	details := make([]protocore.BaseRouteTable, 0, len(daoRouteTableResp.Details))
	for _, routeTable := range daoRouteTableResp.Details {
		details = append(details, converter.PtrToVal(convertBaseRouteTable(&routeTable)))
	}

	return &protocloud.RouteTableListResult{Details: details}, nil
}

func convertBaseRouteTable(dbRouteTable *tablecloud.RouteTableTable) *protocore.BaseRouteTable {
	if dbRouteTable == nil {
		return nil
	}

	return &protocore.BaseRouteTable{
		ID:         dbRouteTable.ID,
		Vendor:     dbRouteTable.Vendor,
		AccountID:  dbRouteTable.AccountID,
		CloudID:    dbRouteTable.CloudID,
		CloudVpcID: dbRouteTable.CloudVpcID,
		Name:       converter.PtrToVal(dbRouteTable.Name),
		Region:     dbRouteTable.Region,
		Memo:       dbRouteTable.Memo,
		VpcID:      dbRouteTable.VpcID,
		BkBizID:    dbRouteTable.BkBizID,
		Revision: &core.Revision{
			Creator:   dbRouteTable.Creator,
			Reviser:   dbRouteTable.Reviser,
			CreatedAt: dbRouteTable.CreatedAt.String(),
			UpdatedAt: dbRouteTable.UpdatedAt.String(),
		},
	}
}

// BatchDeleteRouteTable batch delete route tables.
func (svc *routeTableSvc) BatchDeleteRouteTable(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listResp, err := svc.listRouteTablesForDelete(cts.Kit, req.Filter)
	if err != nil {
		return nil, err
	}
	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delRouteTableIDs := make([]string, len(listResp.Details))
	delRouteTableIDMap := make(map[enumor.Vendor][]string)
	for index, one := range listResp.Details {
		delRouteTableIDs[index] = one.ID
		delRouteTableIDMap[one.Vendor] = append(delRouteTableIDMap[one.Vendor], one.ID)
	}

	// check if all route tables are not bound with subnet
	err = svc.checkRouteTableBinding(cts.Kit, delRouteTableIDs)
	if err != nil {
		return nil, err
	}
	// delete route table routes firstly, then delete route table itself.
	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for vendor, ids := range delRouteTableIDMap {

			delRouteFilter := tools.ContainersExpression("route_table_id", ids)

			switch vendor {
			case enumor.TCloud:
				if err := svc.dao.Route().TCloud().BatchDeleteWithTx(cts.Kit, txn, delRouteFilter); err != nil {
					return nil, err
				}
			case enumor.Aws:
				if err := svc.dao.Route().Aws().BatchDeleteWithTx(cts.Kit, txn, delRouteFilter); err != nil {
					return nil, err
				}
			case enumor.Azure:
				if err := svc.dao.Route().Azure().BatchDeleteWithTx(cts.Kit, txn, delRouteFilter); err != nil {
					return nil, err
				}
			case enumor.HuaWei:
				if err := svc.dao.Route().HuaWei().BatchDeleteWithTx(cts.Kit, txn, delRouteFilter); err != nil {
					return nil, err
				}
			default:
				return nil, errf.Newf(errf.InvalidParameter, "vendor %s is invalid", vendor)
			}

		}

		delRouteTableFilter := tools.ContainersExpression("id", delRouteTableIDs)
		if err := svc.dao.RouteTable().BatchDeleteWithTx(cts.Kit, txn, delRouteTableFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *routeTableSvc) listRouteTablesForDelete(kt *kit.Kit, fil *filter.Expression) (
	*types.RouteTableListResult, error) {

	opt := &types.ListOption{
		Filter: fil,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.RouteTable().List(kt, opt)
	if err != nil {
		logs.Errorf("list route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list route table failed, err: %v", err)
	}
	return listResp, nil
}

func (svc *routeTableSvc) checkRouteTableBinding(kt *kit.Kit, routeTableIDs []string) error {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("route_table_id", routeTableIDs),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Subnet().List(kt, opt)
	if err != nil {
		logs.Errorf("count subnet failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("count subnet failed, err: %v", err)
	}

	if listRes.Count > 0 {
		logs.Errorf("some route table is bound with subnet, ids: %+v, rid: %s", routeTableIDs, kt.Rid)
		return fmt.Errorf("delete route table failed, some route table is bound with subnet")
	}

	return nil
}

// CountRouteTableSubnets count route tables' subnets.
func (svc *routeTableSvc) CountRouteTableSubnets(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.CountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.CountOption{
		Filter:  req.Filter,
		GroupBy: "route_table_id",
	}
	res, err := svc.dao.Subnet().Count(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list route table failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list route table failed, err: %v", err)
	}

	counts := make([]protocloud.RouteTableSubnetsCountResult, 0, len(res))
	for _, cnt := range res {
		counts = append(counts, protocloud.RouteTableSubnetsCountResult{
			Count: cnt.Count,
			ID:    cnt.GroupField,
		})
	}

	return counts, nil
}

// ListRouteTableWithExtension list route table extension
func (svc *routeTableSvc) ListRouteTableWithExtension(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListReq)
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
	data, err := svc.dao.RouteTable().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list route table extension failed, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list route table extension failed, err: %v", err)
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoRouteTableExt[protocore.TCloudRouteTableExtension](data)
	case enumor.Azure:
		return toProtoRouteTableExt[protocore.AzureRouteTableExtension](data)
	case enumor.HuaWei:
		return toProtoRouteTableExt[protocore.HuaWeiRouteTableExtension](data)
	case enumor.Aws:
		return toProtoRouteTableExt[protocore.AwsRouteTableExtension](data)
	case enumor.Gcp:
		return data, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}
