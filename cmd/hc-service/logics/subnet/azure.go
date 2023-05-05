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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ConvAzureCreateReq convert hc-service azure subnet create request to adaption options.
func ConvAzureCreateReq(req *subnet.SubnetCreateReq[subnet.AzureSubnetCreateExt]) *types.AzureSubnetCreateOption {
	return &types.AzureSubnetCreateOption{
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
}

// AzureSubnetSync sync azure subnets.
func (s *Subnet) AzureSubnetSync(kt *kit.Kit, req *AzureSubnetSyncOptions) (*core.BatchCreateResult, error) {
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReqs := make([]cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt], 0, len(req.Subnets))

	routeTableIDs := make([]string, 0, len(req.Subnets))
	for _, subnet := range req.Subnets {
		createReqs = append(createReqs, convertAzureSubnetCreateReq(&subnet, req.AccountID, constant.UnassignedBiz))
		if subnet.Extension.CloudRouteTableID != nil {
			routeTableIDs = append(routeTableIDs, *subnet.Extension.CloudRouteTableID)
		}
	}

	syncOpt := &hcservice.AzureResourceSyncReq{
		AccountID:         req.AccountID,
		ResourceGroupName: req.ResourceGroup,
		CloudVpcID:        req.CloudVpcID,
	}

	res, err := syncsubnet.BatchCreateAzureSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync azure subnet failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(routeTableIDs) != 0 {
		rtSyncOpt := &hcroutetable.AzureRouteTableSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroup,
			CloudIDs:          routeTableIDs,
		}
		_, err = routetable.AzureRouteTableSync(kt, rtSyncOpt, s.adaptor, s.client.DataService())
		if err != nil {
			return nil, err
		}
	}

	return res, nil
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
