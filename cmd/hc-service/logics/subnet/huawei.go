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
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// HuaWeiSubnetCreate create huawei subnet.
func (s *Subnet) HuaWeiSubnetCreate(kt *kit.Kit, opt *SubnetCreateOptions[hcservice.HuaWeiSubnetCreateExt]) (
	*core.BatchCreateResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.HuaWei(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	// create huawei subnets
	createReqs := make([]cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt], 0, len(opt.CreateReqs))
	for _, req := range opt.CreateReqs {
		huaweiCreateOpt := &types.HuaWeiSubnetCreateOption{
			Name:       req.Name,
			Memo:       req.Memo,
			CloudVpcID: opt.CloudVpcID,
			Extension: &types.HuaWeiSubnetCreateExt{
				Region:     opt.Region,
				Zone:       req.Extension.Zone,
				IPv4Cidr:   req.Extension.IPv4Cidr,
				Ipv6Enable: req.Extension.Ipv6Enable,
				GatewayIp:  req.Extension.GatewayIp,
			},
		}
		huaweiCreateRes, err := cli.CreateSubnet(kt, huaweiCreateOpt)
		if err != nil {
			return nil, err
		}

		createReqs = append(createReqs, convertHuaWeiSubnetCreateReq(huaweiCreateRes, opt.AccountID, opt.BkBizID))
	}

	// create hcm subnets
	sync.SleepBeforeSync()

	syncOpt := &syncsubnet.SyncHuaWeiOption{
		AccountID:  opt.AccountID,
		Region:     opt.Region,
		CloudVpcID: opt.CloudVpcID,
	}
	res, err := syncsubnet.BatchCreateHuaWeiSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync huawei subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	return res, nil
}

func convertHuaWeiSubnetCreateReq(data *types.HuaWeiSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: data.CloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Extension.Region,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.HuaWeiSubnetCreateExt{
			Status:       data.Extension.Status,
			DhcpEnable:   data.Extension.DhcpEnable,
			GatewayIp:    data.Extension.GatewayIp,
			DnsList:      data.Extension.DnsList,
			NtpAddresses: data.Extension.NtpAddresses,
		},
	}

	return subnetReq
}
