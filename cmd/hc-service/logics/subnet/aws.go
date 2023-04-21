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

// Package subnet defines subnet logics.
package subnet

import (
	syncsubnet "hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AwsSubnetCreate create aws subnet.
func (s *Subnet) AwsSubnetCreate(kt *kit.Kit, opt *SubnetCreateOptions[hcservice.AwsSubnetCreateExt]) (
	*core.BatchCreateResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.Aws(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	// create aws subnets
	createReqs := make([]cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt], 0, len(opt.CreateReqs))
	for _, req := range opt.CreateReqs {
		awsCreateOpt := &types.AwsSubnetCreateOption{
			Name:       req.Name,
			Memo:       req.Memo,
			CloudVpcID: opt.CloudVpcID,
			Extension: &types.AwsSubnetCreateExt{
				Region:   opt.Region,
				Zone:     req.Extension.Zone,
				IPv4Cidr: req.Extension.IPv4Cidr,
				IPv6Cidr: req.Extension.IPv6Cidr,
			},
		}
		awsCreateRes, err := cli.CreateSubnet(kt, awsCreateOpt)
		if err != nil {
			return nil, err
		}

		createReqs = append(createReqs, convertAwsSubnetCreateReq(awsCreateRes, opt.AccountID, opt.BkBizID))
	}

	// create hcm subnets
	syncOpt := &syncsubnet.SyncAwsOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
	}
	res, err := syncsubnet.BatchCreateAwsSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync aws subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	return res, nil
}

func convertAwsSubnetCreateReq(data *types.AwsSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: data.CloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Extension.Region,
		Zone:       data.Extension.Zone,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.AwsSubnetCreateExt{
			State:                       data.Extension.State,
			IsDefault:                   data.Extension.IsDefault,
			MapPublicIpOnLaunch:         data.Extension.MapPublicIpOnLaunch,
			AssignIpv6AddressOnCreation: data.Extension.AssignIpv6AddressOnCreation,
			HostnameType:                data.Extension.HostnameType,
		},
	}

	return subnetReq
}
