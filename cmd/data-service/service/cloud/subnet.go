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

package cloud

import (
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitSubnetService initialize the subnet service.
func InitSubnetService(cap *capability.Capability) {
	svc := &subnetSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateSubnet", "POST", "/vendors/{vendor}/subnets/batch/create", svc.BatchCreateSubnet)
	h.Add("BatchUpdateSubnet", "PATCH", "/vendors/{vendor}/subnets/batch", svc.BatchUpdateSubnet)
	h.Add("BatchUpdateSubnetBaseInfo", "PATCH", "/subnets/base/batch", svc.BatchUpdateSubnetBaseInfo)
	h.Add("GetSubnet", "GET", "/vendors/{vendor}/subnets/{id}", svc.GetSubnet)
	h.Add("ListSubnet", "POST", "/subnets/list", svc.ListSubnet)
	h.Add("ListSubnetExt", "POST", "/vendors/{vendor}/subnets/list", svc.ListSubnetExt)
	h.Add("DeleteSubnet", "DELETE", "/subnets/batch", svc.BatchDeleteSubnet)

	h.Load(cap.WebService)
}

type subnetSvc struct {
	dao dao.Set
}

// BatchCreateSubnet batch create subnet.
func (svc *subnetSvc) BatchCreateSubnet(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateSubnet[protocloud.TCloudSubnetCreateExt](cts, vendor, svc)
	case enumor.Aws:
		return batchCreateSubnet[protocloud.AwsSubnetCreateExt](cts, vendor, svc)
	case enumor.Gcp:
		return batchCreateSubnet[protocloud.GcpSubnetCreateExt](cts, vendor, svc)
	case enumor.HuaWei:
		return batchCreateSubnet[protocloud.HuaWeiSubnetCreateExt](cts, vendor, svc)
	case enumor.Azure:
		return batchCreateSubnet[protocloud.AzureSubnetCreateExt](cts, vendor, svc)
	}

	return nil, nil
}

// batchCreateSubnet batch create subnet.
func batchCreateSubnet[T protocloud.SubnetCreateExtension](cts *rest.Contexts, vendor enumor.Vendor, svc *subnetSvc) (
	interface{}, error) {

	req := new(protocloud.SubnetBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get vpc & route table cloud id to id mapping
	vpcCloudIDs := make([]string, 0, len(req.Subnets))
	rtCloudIDs := make([]string, 0)
	for _, subnet := range req.Subnets {
		vpcCloudIDs = append(vpcCloudIDs, subnet.CloudVpcID)
		if len(subnet.CloudRouteTableID) != 0 {
			rtCloudIDs = append(rtCloudIDs, subnet.CloudRouteTableID)
		}
	}

	subnetIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		subnets := make([]tablecloud.SubnetTable, 0, len(req.Subnets))
		for _, createReq := range req.Subnets {
			ext, err := tabletype.NewJsonField(createReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			subnet := tablecloud.SubnetTable{
				Vendor:            vendor,
				AccountID:         createReq.AccountID,
				CloudVpcID:        createReq.CloudVpcID,
				CloudRouteTableID: &createReq.CloudRouteTableID,
				CloudID:           createReq.CloudID,
				Name:              createReq.Name,
				VpcID:             createReq.VpcID,
				Region:            createReq.Region,
				Zone:              createReq.Zone,
				Ipv4Cidr:          createReq.Ipv4Cidr,
				Ipv6Cidr:          createReq.Ipv6Cidr,
				Memo:              createReq.Memo,
				Extension:         ext,
				RouteTableID:      &createReq.RouteTableID,
				BkBizID:           createReq.BkBizID,
				Creator:           cts.Kit.User,
				Reviser:           cts.Kit.User,
			}

			subnets = append(subnets, subnet)
		}

		subnetID, err := svc.dao.Subnet().BatchCreateWithTx(cts.Kit, txn, subnets)
		if err != nil {
			return nil, fmt.Errorf("create subnet failed, err: %v", err)
		}

		return subnetID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := subnetIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create subnet but return ids type %s is not string array",
			reflect.TypeOf(subnetIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateSubnet batch update subnet.
func (svc *subnetSvc) BatchUpdateSubnet(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateSubnet[protocloud.TCloudSubnetUpdateExt](cts, svc)
	case enumor.Aws:
		return batchUpdateSubnet[protocloud.AwsSubnetUpdateExt](cts, svc)
	case enumor.Gcp:
		return batchUpdateSubnet[protocloud.GcpSubnetUpdateExt](cts, svc)
	case enumor.HuaWei:
		return batchUpdateSubnet[protocloud.HuaWeiSubnetUpdateExt](cts, svc)
	case enumor.Azure:
		return batchUpdateSubnet[protocloud.AzureSubnetUpdateExt](cts, svc)
	}

	return nil, nil
}

// batchUpdateSubnet batch update subnet.
func batchUpdateSubnet[T protocloud.SubnetUpdateExtension](cts *rest.Contexts, svc *subnetSvc) (
	interface{}, error) {

	req := new(protocloud.SubnetBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Subnets))
	for _, subnet := range req.Subnets {
		ids = append(ids, subnet.ID)
	}

	// check if all subnets exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Subnet().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list subnet failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.Subnets)) {
		return nil, fmt.Errorf("list subnet failed, some subnet(ids=%+v) doesn't exist", ids)
	}

	// update subnet
	subnet := &tablecloud.SubnetTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Subnets {
		subnet.Region = updateReq.Region
		subnet.Name = updateReq.Name
		subnet.Ipv4Cidr = updateReq.Ipv4Cidr
		subnet.Ipv6Cidr = updateReq.Ipv6Cidr
		subnet.Memo = updateReq.Memo
		subnet.BkBizID = updateReq.BkBizID

		// update extension
		if updateReq.Extension != nil {
			dbAccount, err := getSubnetFromTable(cts.Kit, svc.dao, updateReq.ID)
			if err != nil {
				return nil, err
			}

			updatedExtension, err := json.UpdateMerge(updateReq.Extension, string(dbAccount.Extension))
			if err != nil {
				return nil, fmt.Errorf("extension update merge failed, err: %v", err)
			}

			subnet.Extension = tabletype.JsonField(updatedExtension)
		}

		err = svc.dao.Subnet().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), subnet)
		if err != nil {
			logs.Errorf("update subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update subnet failed, err: %v", err)
		}
	}
	return nil, nil
}

// BatchUpdateSubnetBaseInfo batch update subnet basic info.
func (svc *subnetSvc) BatchUpdateSubnetBaseInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SubnetBaseInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0)
	rtCloudIDs := make([]string, 0)
	for _, subnet := range req.Subnets {
		ids = append(ids, subnet.IDs...)
		if subnet.Data.CloudRouteTableID != nil && *subnet.Data.CloudRouteTableID != "" {
			rtCloudIDs = append(rtCloudIDs, *subnet.Data.CloudRouteTableID)
		}
	}

	// check if all subnets exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Subnet().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list subnet failed, err: %v", err)
	}
	if listRes.Count != uint64(len(ids)) {
		return nil, fmt.Errorf("list subnet failed, some subnet(ids=%+v) doesn't exist", ids)
	}

	// update subnet
	subnet := &tablecloud.SubnetTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Subnets {
		subnet.Name = updateReq.Data.Name
		subnet.Ipv4Cidr = updateReq.Data.Ipv4Cidr
		subnet.Ipv6Cidr = updateReq.Data.Ipv6Cidr
		subnet.Memo = updateReq.Data.Memo
		subnet.BkBizID = updateReq.Data.BkBizID
		if updateReq.Data.CloudRouteTableID != nil {
			subnet.RouteTableID = updateReq.Data.RouteTableID
			subnet.CloudRouteTableID = updateReq.Data.CloudRouteTableID
		}
		err = svc.dao.Subnet().Update(cts.Kit, tools.ContainersExpression("id", updateReq.IDs), subnet)
		if err != nil {
			logs.Errorf("update subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update subnet failed, err: %v", err)
		}

	}

	return nil, nil
}

// GetSubnet get subnet details.
func (svc *subnetSvc) GetSubnet(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	subnetID := cts.PathParameter("id").String()

	dbSubnet, err := getSubnetFromTable(cts.Kit, svc.dao, subnetID)
	if err != nil {
		return nil, err
	}

	base := convertBaseSubnet(dbSubnet)

	switch vendor {
	case enumor.TCloud:
		return convertToSubnetResult[protocore.TCloudSubnetExtension](base, dbSubnet.Extension)
	case enumor.Aws:
		return convertToSubnetResult[protocore.AwsSubnetExtension](base, dbSubnet.Extension)
	case enumor.Gcp:
		return convertToSubnetResult[protocore.GcpSubnetExtension](base, dbSubnet.Extension)
	case enumor.HuaWei:
		return convertToSubnetResult[protocore.HuaWeiSubnetExtension](base, dbSubnet.Extension)
	case enumor.Azure:
		return convertToSubnetResult[protocore.AzureSubnetExtension](base, dbSubnet.Extension)
	}

	return nil, nil
}

func convertToSubnetResult[T protocore.SubnetExtension](baseSubnet *protocore.BaseSubnet, dbExtension tabletype.JsonField) (
	*protocore.Subnet[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &protocore.Subnet[T]{
		BaseSubnet: *baseSubnet,
		Extension:  extension,
	}, nil
}

func getSubnetFromTable(kt *kit.Kit, dao dao.Set, subnetID string) (*tablecloud.SubnetTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", subnetID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.Subnet().List(kt, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list subnet failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list subnet failed, account(id=%s) doesn't exist", subnetID)
	}

	return &details[0], nil
}

// ListSubnet list subnets.
func (svc *subnetSvc) ListSubnet(cts *rest.Contexts) (interface{}, error) {
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
	daoSubnetResp, err := svc.dao.Subnet().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list subnet failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.SubnetListResult{Count: daoSubnetResp.Count}, nil
	}

	details := make([]protocore.BaseSubnet, 0, len(daoSubnetResp.Details))
	for _, subnet := range daoSubnetResp.Details {
		details = append(details, converter.PtrToVal(convertBaseSubnet(&subnet)))
	}

	return &protocloud.SubnetListResult{Details: details}, nil
}

func convertBaseSubnet(dbSubnet *tablecloud.SubnetTable) *protocore.BaseSubnet {
	if dbSubnet == nil {
		return nil
	}

	return &protocore.BaseSubnet{
		ID:                dbSubnet.ID,
		Vendor:            dbSubnet.Vendor,
		AccountID:         dbSubnet.AccountID,
		CloudVpcID:        dbSubnet.CloudVpcID,
		CloudRouteTableID: converter.PtrToVal(dbSubnet.CloudRouteTableID),
		CloudID:           dbSubnet.CloudID,
		Name:              converter.PtrToVal(dbSubnet.Name),
		Region:            dbSubnet.Region,
		Zone:              dbSubnet.Zone,
		Ipv4Cidr:          dbSubnet.Ipv4Cidr,
		Ipv6Cidr:          dbSubnet.Ipv6Cidr,
		Memo:              dbSubnet.Memo,
		VpcID:             dbSubnet.VpcID,
		RouteTableID:      converter.PtrToVal(dbSubnet.RouteTableID),
		BkBizID:           dbSubnet.BkBizID,
		Revision: &core.Revision{
			Creator:   dbSubnet.Creator,
			Reviser:   dbSubnet.Reviser,
			CreatedAt: dbSubnet.CreatedAt.String(),
			UpdatedAt: dbSubnet.UpdatedAt.String(),
		},
	}
}

// BatchDeleteSubnet batch delete subnets.
func (svc *subnetSvc) BatchDeleteSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.Subnet().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list subnet failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delSubnetIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delSubnetIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delSubnetFilter := tools.ContainersExpression("id", delSubnetIDs)
		if err := svc.dao.Subnet().BatchDeleteWithTx(cts.Kit, txn, delSubnetFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListSubnetExt ...
func (svc *subnetSvc) ListSubnetExt(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.Subnet().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return conSubnetExtListResult[protocore.TCloudSubnetExtension](listResp.Details)
	case enumor.Aws:
		return conSubnetExtListResult[protocore.AwsSubnetExtension](listResp.Details)
	case enumor.Azure:
		return conSubnetExtListResult[protocore.AzureSubnetExtension](listResp.Details)
	case enumor.HuaWei:
		return conSubnetExtListResult[protocore.HuaWeiSubnetExtension](listResp.Details)
	case enumor.Gcp:
		return conSubnetExtListResult[protocore.GcpSubnetExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func conSubnetExtListResult[T protocore.SubnetExtension](tables []tablecloud.SubnetTable) (
	*protocloud.SubnetExtListResult[T], error) {

	details := make([]protocore.Subnet[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		err := json.UnmarshalFromString(string(one.Extension), &extension)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString vpc json extension failed, err: %v", err)
		}

		details = append(details, protocore.Subnet[T]{
			BaseSubnet: *convertBaseSubnet(&one),
			Extension:  extension,
		})
	}

	return &protocloud.SubnetExtListResult[T]{
		Details: details,
	}, nil
}
