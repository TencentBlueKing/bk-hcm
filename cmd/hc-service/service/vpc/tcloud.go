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
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/subnet"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// TCloudVpcCreate create tcloud vpc.
func (v vpc) TCloudVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.TCloudVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create tcloud vpc
	opt := &types.TCloudVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.TCloudVpcCreateExt{
			Region:   req.Extension.Region,
			IPv4Cidr: req.Extension.IPv4Cidr,
		},
	}
	data, err := cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.TCloudVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.TCloudVpcCreateExt]{convertTCloudVpcCreateReq(req, data)},
	}
	result, err := v.cs.DataService().TCloud.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		return nil, errf.New(errf.Aborted, "create result is invalid")
	}

	// create tcloud subnets
	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: result.IDs[0]}, nil
	}

	subnetCreateOpt := &subnet.TCloudSubnetBatchCreateReq{
		BkBizID:    constant.UnassignedBiz,
		AccountID:  req.AccountID,
		Region:     data.Region,
		CloudVpcID: data.CloudID,
		Subnets:    make([]subnet.TCloudOneSubnetCreateReq, 0, len(req.Extension.Subnets)),
	}

	for _, subnetCreateReq := range req.Extension.Subnets {
		subnetCreateOpt.Subnets = append(subnetCreateOpt.Subnets, subnet.TCloudOneSubnetCreateReq{
			IPv4Cidr:          subnetCreateReq.IPv4Cidr,
			Name:              subnetCreateReq.Name,
			Zone:              subnetCreateReq.Zone,
			CloudRouteTableID: subnetCreateReq.CloudRouteTableID,
			Memo:              subnetCreateReq.Memo,
		})
	}
	_, err = v.subnet.TCloudSubnetCreate(cts.Kit, subnetCreateOpt)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

func convertTCloudVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.TCloudVpcCreateExt],
	data *types.TCloudVpc) cloud.VpcCreateReq[cloud.TCloudVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.TCloudVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.TCloudVpcCreateExt{
			Cidr:            make([]cloud.TCloudCidr, 0, len(data.Extension.Cidr)),
			IsDefault:       data.Extension.IsDefault,
			EnableMulticast: data.Extension.EnableMulticast,
			DnsServerSet:    data.Extension.DnsServerSet,
			DomainName:      data.Extension.DomainName,
		},
	}

	for _, cidr := range data.Extension.Cidr {
		vpcReq.Extension.Cidr = append(vpcReq.Extension.Cidr, cloud.TCloudCidr{
			Type:     cidr.Type,
			Cidr:     cidr.Cidr,
			Category: cidr.Category,
		})
	}

	return vpcReq
}

// TCloudVpcUpdate update tencent cloud vpc.
func (v vpc) TCloudVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().TCloud.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.TCloud(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.TCloudVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.TCloudVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().TCloud.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudVpcDelete delete tencent cloud vpc.
func (v vpc) TCloudVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().TCloud.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.TCloud(cts.Kit, getRes.AccountID)
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
