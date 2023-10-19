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

// Package vpc defines vpc service.
package vpc

import (
	"hcm/cmd/hc-service/logics/subnet"
	syncroutetable "hcm/cmd/hc-service/logics/sync/route-table"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	subnetproto "hcm/pkg/api/hc-service/subnet"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// HuaWeiVpcCreate create huawei vpc.
func (v vpc) HuaWeiVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create huawei vpc
	opt := &types.HuaWeiVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.HuaWeiVpcCreateExt{
			Region:              req.Extension.Region,
			IPv4Cidr:            req.Extension.IPv4Cidr,
			EnterpriseProjectID: req.Extension.EnterpriseProjectID,
		},
	}
	err = cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// get created vpc info
	listOpt := &types.HuaWeiVpcListOption{
		HuaWeiListOption: adcore.HuaWeiListOption{
			Region: req.Extension.Region,
		},
		Names: []string{req.Name},
	}
	listRes, err := cli.ListVpc(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(listRes.Details) != 1 {
		return nil, errf.Newf(errf.Aborted, "get created vpc detail, but result count is invalid")
	}

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.HuaWeiVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{convertHuaWeiVpcCreateReq(req, &listRes.Details[0])},
	}
	result, err := v.cs.DataService().HuaWei.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		return nil, errf.New(errf.Aborted, "create result is invalid")
	}

	// create huawei subnets
	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: result.IDs[0]}, nil
	}

	subnetCreateOpt := &subnet.SubnetCreateOptions[subnetproto.HuaWeiSubnetCreateExt]{
		BkBizID:    constant.UnassignedBiz,
		AccountID:  req.AccountID,
		Region:     listRes.Details[0].Region,
		CloudVpcID: listRes.Details[0].CloudID,
		CreateReqs: req.Extension.Subnets,
	}
	_, err = v.subnet.HuaWeiSubnetCreate(cts.Kit, subnetCreateOpt)
	if err != nil {
		return nil, err
	}

	// TODO: sync-todo change to 3.0 sync route table
	rtReq := &hcroutetable.HuaWeiRouteTableSyncReq{
		AccountID: req.AccountID,
		Region:    req.Extension.Region,
	}
	if _, err = syncroutetable.HuaWeiRouteTableSync(cts.Kit, rtReq, v.ad, v.cs.DataService()); err != nil {
		return nil, err
	}
	return core.CreateResult{ID: result.IDs[0]}, nil
}

func convertHuaWeiVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt],
	data *types.HuaWeiVpc) cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		BkCloudID: req.BkCloudID,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.HuaWeiVpcCreateExt{
			Cidr:                make([]cloud.HuaWeiCidr, 0, len(data.Extension.Cidr)),
			Status:              data.Extension.Status,
			EnterpriseProjectID: data.Extension.EnterpriseProjectId,
		},
	}

	for _, cidr := range data.Extension.Cidr {
		vpcReq.Extension.Cidr = append(vpcReq.Extension.Cidr, cloud.HuaWeiCidr{
			Type: cidr.Type,
			Cidr: cidr.Cidr,
		})
	}

	return vpcReq
}

// HuaWeiVpcUpdate update huawei vpc.
func (v vpc) HuaWeiVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.HuaWeiVpcUpdateOption{
		VpcUpdateOption: types.VpcUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
		},
		Region: getRes.Region,
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.HuaWeiVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().HuaWei.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiVpcDelete delete huawei vpc.
func (v vpc) HuaWeiVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Region,
	}
	err = cli.DeleteVpc(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
