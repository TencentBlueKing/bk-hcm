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

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

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
