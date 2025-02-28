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
	logicsrt "hcm/cmd/hc-service/logics/route-table"
	"hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	subnetproto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/api/hc-service/sync"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AwsVpcCreate create aws vpc.
func (v vpc) AwsVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.AwsVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create aws vpc
	opt := &types.AwsVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.AwsVpcCreateExt{
			Region:                      req.Extension.Region,
			IPv4Cidr:                    req.Extension.IPv4Cidr,
			AmazonProvidedIpv6CidrBlock: req.Extension.AmazonProvidedIpv6CidrBlock,
			InstanceTenancy:             req.Extension.InstanceTenancy,
		},
	}
	data, err := cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.AwsVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{convertAwsVpcCreateReq(req, data)},
	}
	result, err := v.cs.DataService().Aws.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		return nil, errf.New(errf.Aborted, "create result is invalid")
	}

	// create aws subnets
	if len(req.Extension.Subnets) > 0 {
		subnetCreateOpt := &subnet.SubnetCreateOptions[subnetproto.AwsSubnetCreateExt]{
			BkBizID:    constant.UnassignedBiz,
			AccountID:  req.AccountID,
			Region:     data.Region,
			CloudVpcID: data.CloudID,
			CreateReqs: req.Extension.Subnets,
		}
		_, err = v.subnet.AwsSubnetCreate(cts.Kit, subnetCreateOpt)
		if err != nil {
			return nil, err
		}
	}

	syncCli, err := v.syncCli.Aws(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("build aws sync client failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = handler.ResourceSync(cts, &logicsrt.AwsRouteTableHandler{
		DisablePrepare: true,
		Cli:            v.syncCli,
		Request: &sync.AwsSyncReq{
			AccountID: req.AccountID,
			Region:    req.Extension.Region,
		},
		SyncCli: syncCli,
	})
	if err != nil {
		logs.Errorf("route table sync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

func convertAwsVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.AwsVpcCreateExt],
	data *types.AwsVpc) cloud.VpcCreateReq[cloud.AwsVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.AwsVpcCreateExt{
			Cidr:               make([]cloud.AwsCidr, 0, len(data.Extension.Cidr)),
			State:              data.Extension.State,
			InstanceTenancy:    data.Extension.InstanceTenancy,
			IsDefault:          data.Extension.IsDefault,
			EnableDnsHostnames: data.Extension.EnableDnsHostnames,
			EnableDnsSupport:   data.Extension.EnableDnsSupport,
		},
	}
	for _, cidr := range data.Extension.Cidr {
		vpcReq.Extension.Cidr = append(vpcReq.Extension.Cidr, cloud.AwsCidr{
			Type:        cidr.Type,
			Cidr:        cidr.Cidr,
			AddressPool: cidr.AddressPool,
			State:       cidr.State,
		})
	}

	return vpcReq
}

// AwsVpcUpdate update aws vpc.
func (v vpc) AwsVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Aws.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Aws(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AwsVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Aws.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AwsVpcDelete delete aws vpc.
func (v vpc) AwsVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Aws.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Aws(cts.Kit, getRes.AccountID)
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
