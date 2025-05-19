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
	subnetlogics "hcm/cmd/hc-service/logics/subnet"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// AzureVpcCreate create azure vpc.
func (v vpc) AzureVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.AzureVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create azure vpc
	opt := &types.AzureVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.AzureVpcCreateExt{
			Region:        req.Extension.Region,
			ResourceGroup: req.Extension.ResourceGroup,
			IPv4Cidr:      req.Extension.IPv4Cidr,
			IPv6Cidr:      req.Extension.IPv6Cidr,
			Subnets:       make([]adtysubnet.AzureSubnetCreateOption, 0, len(req.Extension.Subnets)),
		},
	}
	for _, subnet := range req.Extension.Subnets {
		opt.Extension.Subnets = append(opt.Extension.Subnets, *subnetlogics.ConvAzureCreateReq(&subnet))
	}
	data, err := cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// create hcm vpc & subnets
	createReq := &cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{convertVpcCreateReq(req, data)},
	}
	result, err := v.cs.DataService().Azure.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		return nil, errf.New(errf.Aborted, "create result is invalid")
	}

	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: result.IDs[0]}, nil
	}

	// 业务下申请通过申请单的交付流程 转到对应业务id下，这里不处理
	subnetSyncOpt := &subnetlogics.AzureSubnetSyncOptions{
		BkBizID:       constant.UnassignedBiz,
		AccountID:     req.AccountID,
		CloudVpcID:    data.CloudID,
		ResourceGroup: data.Extension.ResourceGroupName,
		Subnets:       data.Extension.Subnets,
	}
	_, err = v.subnet.AzureSubnetSync(cts.Kit, subnetSyncOpt)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

func convertVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.AzureVpcCreateExt],
	data *types.AzureVpc) cloud.VpcCreateReq[cloud.AzureVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.AzureVpcCreateExt{
			ResourceGroupName: data.Extension.ResourceGroupName,
			DNSServers:        data.Extension.DNSServers,
			Cidr:              make([]cloud.AzureCidr, 0, len(data.Extension.Cidr)),
		},
	}
	for _, cidr := range data.Extension.Cidr {
		vpcReq.Extension.Cidr = append(vpcReq.Extension.Cidr, cloud.AzureCidr{
			Type: cidr.Type,
			Cidr: cidr.Cidr,
		})
	}

	return vpcReq
}

// AzureVpcUpdate update azure vpc.
func (v vpc) AzureVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AzureVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Azure.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureVpcDelete delete azure vpc.
func (v vpc) AzureVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.AzureDeleteOption{
		BaseDeleteOption:  adcore.BaseDeleteOption{ResourceID: getRes.Name},
		ResourceGroupName: getRes.Extension.ResourceGroupName,
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
