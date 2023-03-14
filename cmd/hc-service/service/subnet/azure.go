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
	routetable "hcm/cmd/hc-service/logics/sync/route-table"
	syncsubnet "hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
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

	azureCreateOpt := &types.AzureSubnetCreateOption{
		Name:       req.Name,
		Memo:       req.Memo,
		CloudVpcID: req.CloudVpcID,
		Extension: &types.AzureSubnetCreateExt{
			ResourceGroup:        req.Extension.ResourceGroup,
			IPv4Cidr:             req.Extension.IPv4Cidr,
			IPv6Cidr:             req.Extension.IPv6Cidr,
			CloudRouteTableID:    req.Extension.CloudRouteTableID,
			NatGateway:           req.Extension.NatGateway,
			NetworkSecurityGroup: req.Extension.NetworkSecurityGroup,
		},
	}
	azureCreateRes, err := cli.CreateSubnet(cts.Kit, azureCreateOpt)
	if err != nil {
		return nil, err
	}

	// sync hcm subnets and related route tables
	sync.SleepBeforeSync()

	syncOpt := &hcservice.AzureResourceSyncReq{
		AccountID:         req.AccountID,
		ResourceGroupName: req.Extension.ResourceGroup,
		CloudVpcID:        req.CloudVpcID,
	}
	createReqs := []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{convertAzureSubnetCreateReq(azureCreateRes,
		req.AccountID, req.BkBizID)}
	res, err := syncsubnet.BatchCreateAzureSubnet(cts.Kit, createReqs, s.cs.DataService(), s.ad, syncOpt)
	if err != nil {
		logs.Errorf("sync azure subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, cts.Kit.Rid)
		return nil, err
	}

	if azureCreateRes.Extension.CloudRouteTableID != nil {
		rtSyncOpt := &hcroutetable.AzureRouteTableSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: req.Extension.ResourceGroup,
			CloudIDs:          []string{*azureCreateRes.Extension.CloudRouteTableID},
		}
		_, err = routetable.AzureRouteTableSync(cts.Kit, rtSyncOpt, s.ad, s.cs.DataService())
		if err != nil {
			return nil, err
		}
	}

	return core.CreateResult{ID: res.IDs[0]}, nil
}

func convertAzureSubnetCreateReq(data *types.AzureSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
		AccountID:         accountID,
		CloudVpcID:        data.CloudVpcID,
		CloudID:           data.CloudID,
		Name:              &data.Name,
		Ipv4Cidr:          data.Ipv4Cidr,
		Ipv6Cidr:          data.Ipv6Cidr,
		Memo:              data.Memo,
		BkBizID:           bizID,
		CloudRouteTableID: converter.PtrToVal(data.Extension.CloudRouteTableID),
		Extension: &cloud.AzureSubnetCreateExt{
			ResourceGroupName:    data.Extension.ResourceGroupName,
			NatGateway:           data.Extension.NatGateway,
			CloudSecurityGroupID: data.Extension.NetworkSecurityGroup,
		},
	}

	return subnetReq
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
