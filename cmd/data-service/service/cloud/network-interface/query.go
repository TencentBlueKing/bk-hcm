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

package networkinterface

import (
	"fmt"

	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	datacloudniproto "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListNetworkInterface list network interface.
func (svc *NetworkInterfaceSvc) ListNetworkInterface(cts *rest.Contexts) (interface{}, error) {
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
	daoResp, err := svc.dao.NetworkInterface().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list network interface failed, err: %v", err)
	}

	if req.Page.Count {
		return &datacloudniproto.NetworkInterfaceListResult{Count: daoResp.Count}, nil
	}

	details := make([]coreni.BaseNetworkInterface, 0, len(daoResp.Details))
	for _, item := range daoResp.Details {
		details = append(details, converter.PtrToVal(convertBaseNetworkInterface(&item)))
	}

	return &datacloudniproto.NetworkInterfaceListResult{Details: details}, nil
}

// ListNetworkInterfaceAssociate list network interface associate.
func (svc *NetworkInterfaceSvc) ListNetworkInterfaceAssociate(cts *rest.Contexts) (interface{}, error) {
	req := new(datacloudniproto.NetworkInterfaceListReq)
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
	daoResp, err := svc.dao.NetworkInterface().ListAssociate(cts.Kit, opt, req.IsAssociate)
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list network interface failed, err: %v", err)
	}

	if req.Page.Count {
		return &datacloudniproto.NetworkInterfaceAssociateListResult{Count: converter.PtrToVal(daoResp.Count)}, nil
	}

	details := make([]coreni.NetworkInterfaceAssociate, 0, len(daoResp.Details))
	for _, item := range daoResp.Details {
		details = append(details, convertBaseNIAssociate(item))
	}

	return &datacloudniproto.NetworkInterfaceAssociateListResult{Details: details}, nil
}

// ListNetworkInterfaceExt list network interface with extension.
func (svc *NetworkInterfaceSvc) ListNetworkInterfaceExt(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.NetworkInterface().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list network interface extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.Azure:
		return convertNIExtListResult[coreni.AzureNIExtension](listResp.Details)
	case enumor.HuaWei:
		return convertNIExtListResult[coreni.HuaWeiNIExtension](listResp.Details)
	case enumor.Gcp:
		return convertNIExtListResult[coreni.GcpNIExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func convertNIExtListResult[T coreni.NetworkInterfaceExtension](
	tables []tableni.NetworkInterfaceTable) (*datacloudniproto.NetworkInterfaceExtListResult[T], error) {

	details := make([]coreni.NetworkInterface[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		err := json.UnmarshalFromString(string(one.Extension), &extension)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString network interface json extension failed, err: %v", err)
		}

		details = append(details, coreni.NetworkInterface[T]{
			BaseNetworkInterface: *convertBaseNetworkInterface(&one),
			Extension:            extension,
		})
	}

	return &datacloudniproto.NetworkInterfaceExtListResult[T]{
		Details: details,
	}, nil
}

// GetNetworkInterface get network interface detail.
func (svc *NetworkInterfaceSvc) GetNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	netID := cts.PathParameter("id").String()
	dbDetail, err := getNetworkInterfaceFromTable(cts.Kit, svc.dao, netID)
	if err != nil {
		return nil, err
	}

	base := convertBaseNetworkInterface(dbDetail)
	switch vendor {
	case enumor.Azure:
		return convertToNIResult[coreni.AzureNIExtension](base, dbDetail.Extension)
	case enumor.HuaWei:
		return convertToNIResult[coreni.HuaWeiNIExtension](base, dbDetail.Extension)
	case enumor.Gcp:
		return convertToNIResult[coreni.GcpNIExtension](base, dbDetail.Extension)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func convertToNIResult[T coreni.NetworkInterfaceExtension](baseNI *coreni.BaseNetworkInterface,
	dbExtension tabletype.JsonField) (*coreni.NetworkInterface[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString network interface db extension failed, err: %v", err)
	}

	return &coreni.NetworkInterface[T]{
		BaseNetworkInterface: *baseNI,
		Extension:            extension,
	}, nil
}

// convertBaseNIAssociate convert db detail to associate type.
func convertBaseNIAssociate(dbDetail *types.NetworkInterfaceWithCvmID) coreni.NetworkInterfaceAssociate {
	if dbDetail == nil {
		return coreni.NetworkInterfaceAssociate{}
	}

	tmpPrivateIPv4, tmpPrivateIPv6, tmpPublicIPv4, tmpPublicIPv6 := ConvertIPJSONToArr(dbDetail.PrivateIPv4,
		dbDetail.PrivateIPv6, dbDetail.PublicIPv4, dbDetail.PublicIPv6)

	return coreni.NetworkInterfaceAssociate{
		BaseNetworkInterface: coreni.BaseNetworkInterface{
			ID:            dbDetail.ID,
			Vendor:        dbDetail.Vendor,
			Name:          dbDetail.Name,
			AccountID:     dbDetail.AccountID,
			Region:        dbDetail.Region,
			Zone:          dbDetail.Zone,
			CloudID:       dbDetail.CloudID,
			VpcID:         dbDetail.VpcID,
			CloudVpcID:    dbDetail.CloudVpcID,
			SubnetID:      dbDetail.SubnetID,
			CloudSubnetID: dbDetail.CloudSubnetID,
			PrivateIPv4:   tmpPrivateIPv4,
			PrivateIPv6:   tmpPrivateIPv6,
			PublicIPv4:    tmpPublicIPv4,
			PublicIPv6:    tmpPublicIPv6,
			BkBizID:       dbDetail.BkBizID,
			InstanceID:    dbDetail.InstanceID,
			Revision: &core.Revision{
				Creator:   dbDetail.Creator,
				Reviser:   dbDetail.Reviser,
				CreatedAt: dbDetail.CreatedAt.String(),
				UpdatedAt: dbDetail.UpdatedAt.String(),
			},
		},
		CvmID:        dbDetail.CvmID,
		RelCreator:   dbDetail.RelCreator,
		RelCreatedAt: dbDetail.RelCreatedAt.String(),
	}
}

// convertBaseNetworkInterface convert db detail to base type.
func convertBaseNetworkInterface(dbDetail *tableni.NetworkInterfaceTable) *coreni.BaseNetworkInterface {
	if dbDetail == nil {
		return nil
	}

	tmpPrivateIPv4, tmpPrivateIPv6, tmpPublicIPv4, tmpPublicIPv6 := ConvertIPJSONToArr(dbDetail.PrivateIPv4,
		dbDetail.PrivateIPv6, dbDetail.PublicIPv4, dbDetail.PublicIPv6)

	return &coreni.BaseNetworkInterface{
		ID:            dbDetail.ID,
		Vendor:        dbDetail.Vendor,
		Name:          dbDetail.Name,
		AccountID:     dbDetail.AccountID,
		Region:        dbDetail.Region,
		Zone:          dbDetail.Zone,
		CloudID:       dbDetail.CloudID,
		VpcID:         dbDetail.VpcID,
		CloudVpcID:    dbDetail.CloudVpcID,
		SubnetID:      dbDetail.SubnetID,
		CloudSubnetID: dbDetail.CloudSubnetID,
		PrivateIPv4:   tmpPrivateIPv4,
		PrivateIPv6:   tmpPrivateIPv6,
		PublicIPv4:    tmpPublicIPv4,
		PublicIPv6:    tmpPublicIPv6,
		BkBizID:       dbDetail.BkBizID,
		InstanceID:    dbDetail.InstanceID,

		Revision: &core.Revision{
			Creator:   dbDetail.Creator,
			Reviser:   dbDetail.Reviser,
			CreatedAt: dbDetail.CreatedAt.String(),
			UpdatedAt: dbDetail.UpdatedAt.String(),
		},
	}
}

// getNetworkInterfaceFromTable get network interface from table by netID.
func getNetworkInterfaceFromTable(kt *kit.Kit, dao dao.Set, netID string) (*tableni.NetworkInterfaceTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", netID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.NetworkInterface().List(kt, opt)
	if err != nil {
		logs.Errorf("get network interface failed, netID: %s, err: %v, rid: %s", netID, kt.Rid)
		return nil, fmt.Errorf("get network interface failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("get list network interface failed, network(id=%s) doesn't exist", netID)
	}

	return &details[0], nil
}

// ConvertIPJSONToArr convert ip json to array
func ConvertIPJSONToArr(privateIPv4, privateIPv6, publicIPv4, publicIPv6 tabletype.JsonField) ([]string, []string,
	[]string, []string) {

	var tmpPrivateIPv4 []string
	json.UnmarshalFromString(string(privateIPv4), &tmpPrivateIPv4)
	var tmpPrivateIPv6 []string
	json.UnmarshalFromString(string(privateIPv6), &tmpPrivateIPv6)
	var tmpPublicIPv4 []string
	json.UnmarshalFromString(string(publicIPv4), &tmpPublicIPv4)
	var tmpPublicIPv6 []string
	json.UnmarshalFromString(string(publicIPv6), &tmpPublicIPv6)

	return tmpPrivateIPv4, tmpPrivateIPv6, tmpPublicIPv4, tmpPublicIPv6
}
