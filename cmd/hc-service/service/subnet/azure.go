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

// Package subnet defines subnet service.
package subnet

import (
	subnetlogics "hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// AzureSubnetCreate create azure subnet.
func (s subnet) AzureSubnetCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.SubnetCreateReq[hcservice.AzureSubnetCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	azureCreateOpt := subnetlogics.ConvAzureCreateReq(req)
	azureCreateRes, err := cli.CreateSubnet(cts.Kit, azureCreateOpt)
	if err != nil {
		return nil, err
	}

	// sync hcm subnets and related route tables
	sync.SleepBeforeSync()

	subnetSyncOpt := &subnetlogics.AzureSubnetSyncOptions{
		BkBizID:       req.BkBizID,
		AccountID:     req.AccountID,
		CloudVpcID:    azureCreateRes.CloudID,
		ResourceGroup: azureCreateRes.Extension.ResourceGroupName,
		Subnets:       []types.AzureSubnet{*azureCreateRes},
	}
	res, err := s.subnet.AzureSubnetSync(cts.Kit, subnetSyncOpt)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: res.IDs[0]}, nil
}

// AzureSubnetUpdate update azure subnet.
func (s subnet) AzureSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AzureSubnetUpdateOption)
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().Azure.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureSubnetDelete delete azure subnet.
func (s subnet) AzureSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &types.AzureSubnetDeleteOption{
		AzureDeleteOption: adcore.AzureDeleteOption{
			BaseDeleteOption:  adcore.BaseDeleteOption{ResourceID: getRes.Name},
			ResourceGroupName: getRes.Extension.ResourceGroupName,
		},
		VpcID: getRes.CloudVpcID,
	}
	err = cli.DeleteSubnet(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureSubnetCountIP count azure subnets' available ips.
func (s subnet) AzureSubnetCountIP(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	getRes, err := s.cs.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	usageOpt := &types.AzureVpcListUsageOption{
		ResourceGroupName: getRes.Extension.ResourceGroupName,
		VpcID:             getRes.CloudVpcID,
	}
	usages, err := cli.ListVpcUsage(cts.Kit, usageOpt)
	if err != nil {
		return nil, err
	}

	for _, usage := range usages {
		if converter.PtrToVal(usage.ID) == getRes.CloudID {
			return &hcservice.SubnetCountIPResult{
				AvailableIPv4Count: uint64(converter.PtrToVal(usage.Limit) - converter.PtrToVal(usage.CurrentValue)),
			}, nil
		}
	}

	return nil, errf.New(errf.InvalidParameter, "subnet ip count is not found")
}
