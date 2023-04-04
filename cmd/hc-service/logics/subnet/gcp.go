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
	"strconv"

	syncsubnet "hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GcpSubnetCreate create gcp subnet.
func (s *Subnet) GcpSubnetCreate(kt *kit.Kit, opt *SubnetCreateOptions[hcservice.GcpSubnetCreateExt]) (
	*core.BatchCreateResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.Gcp(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	// get gcp vpc self link by cloud id
	vpcReq := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", opt.CloudVpcID),
		Page:   core.DefaultBasePage,
		Fields: []string{"extension"},
	}
	vpcRes, err := s.client.DataService().Gcp.Vpc.ListVpcExt(kt.Ctx, kt.Header(), vpcReq)
	if err != nil {
		logs.Errorf("get vpc by cloud id %s failed, err: %v, rid: %s", opt.CloudVpcID, err, kt.Rid)
		return nil, err
	}

	if len(vpcRes.Details) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "gcp vpc(cloud id: %s) not exists", opt.CloudVpcID)
	}

	// create gcp subnets
	createdIDs := make([]string, 0, len(opt.CreateReqs))
	for _, req := range opt.CreateReqs {
		gcpCreateOpt := &types.GcpSubnetCreateOption{
			Name:       req.Name,
			Memo:       req.Memo,
			CloudVpcID: vpcRes.Details[0].Extension.SelfLink,
			Extension: &types.GcpSubnetCreateExt{
				Region:                req.Extension.Region,
				IPv4Cidr:              req.Extension.IPv4Cidr,
				PrivateIpGoogleAccess: req.Extension.PrivateIpGoogleAccess,
				EnableFlowLogs:        req.Extension.EnableFlowLogs,
			},
		}
		createdID, err := cli.CreateSubnet(kt, gcpCreateOpt)
		if err != nil {
			return nil, err
		}

		cloudID := strconv.FormatUint(createdID, 10)
		createdIDs = append(createdIDs, cloudID)
	}

	// get created subnets
	subnetRes, err := cli.ListSubnet(kt, &types.GcpSubnetListOption{
		GcpListOption: adcore.GcpListOption{CloudIDs: createdIDs},
		Region:        opt.Region,
	})
	if err != nil {
		logs.Errorf("get subnet failed, err: %v,s, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(subnetRes.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "created subnets are not found")
	}

	createReqs := make([]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt], 0, len(subnetRes.Details))
	for _, subnet := range subnetRes.Details {
		createReqs = append(createReqs, convertGcpSubnetCreateReq(&subnet, opt.AccountID, opt.CloudVpcID,
			opt.BkBizID))
	}

	// create hcm subnets
	sync.SleepBeforeSync()

	syncOpt := &syncsubnet.SyncGcpOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
	}
	res, err := syncsubnet.BatchCreateGcpSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync gcp subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	return res, nil
}

func convertGcpSubnetCreateReq(data *types.GcpSubnet, accountID, cloudVpcID string,
	bizID int64) cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: cloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Extension.Region,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.GcpSubnetCreateExt{
			VpcSelfLink:           data.CloudVpcID,
			SelfLink:              data.Extension.SelfLink,
			StackType:             data.Extension.StackType,
			Ipv6AccessType:        data.Extension.Ipv6AccessType,
			GatewayAddress:        data.Extension.GatewayAddress,
			PrivateIpGoogleAccess: data.Extension.PrivateIpGoogleAccess,
			EnableFlowLogs:        data.Extension.EnableFlowLogs,
		},
	}

	return subnetReq
}
