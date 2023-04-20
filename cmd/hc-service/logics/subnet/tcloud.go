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
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// TCloudSubnetCreate create tcloud subnet.
func (s *Subnet) TCloudSubnetCreate(kt *kit.Kit, req *hcservice.TCloudSubnetBatchCreateReq) (
	*core.BatchCreateResult, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create tencent cloud subnets
	subnetReqs := make([]types.TCloudOneSubnetCreateOpt, 0, len(req.Subnets))
	for _, subnetReq := range req.Subnets {
		subnetReqs = append(subnetReqs, types.TCloudOneSubnetCreateOpt{
			IPv4Cidr:          subnetReq.IPv4Cidr,
			Name:              subnetReq.Name,
			Zone:              subnetReq.Zone,
			CloudRouteTableID: subnetReq.CloudRouteTableID,
		})
	}

	tcloudCreateOpt := &types.TCloudSubnetsCreateOption{
		AccountID:  req.AccountID,
		Region:     req.Region,
		CloudVpcID: req.CloudVpcID,
		Subnets:    subnetReqs,
	}
	createdSubnets, err := cli.CreateSubnets(kt, tcloudCreateOpt)
	if err != nil {
		return nil, err
	}

	// sync hcm subnets and related route tables
	sync.SleepBeforeSync()

	createReqs := make([]cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt], 0, len(createdSubnets))
	cloudRTIDs := make([]string, 0, len(req.Subnets))
	for _, subnet := range createdSubnets {
		createReq := convertTCloudSubnetCreateReq(&subnet, req.AccountID, constant.UnassignedBiz)
		createReqs = append(createReqs, createReq)
		cloudRTIDs = append(cloudRTIDs, createReq.CloudRouteTableID)
	}

	syncOpt := &syncsubnet.SyncTCloudOption{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	res, err := syncsubnet.BatchCreateTCloudSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync tcloud subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	rtSyncOpt := &hcroutetable.TCloudRouteTableSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudRTIDs,
	}
	_, err = routetable.TCloudRouteTableSync(kt, rtSyncOpt, s.adaptor, s.client.DataService())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func convertTCloudSubnetCreateReq(data *types.TCloudSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]{
		AccountID:         accountID,
		CloudVpcID:        data.CloudVpcID,
		BkBizID:           bizID,
		CloudRouteTableID: converter.PtrToVal(data.Extension.CloudRouteTableID),
		CloudID:           data.CloudID,
		Name:              &data.Name,
		Region:            data.Extension.Region,
		Zone:              data.Extension.Zone,
		Ipv4Cidr:          data.Ipv4Cidr,
		Ipv6Cidr:          data.Ipv6Cidr,
		Memo:              data.Memo,
		Extension: &cloud.TCloudSubnetCreateExt{
			IsDefault:         data.Extension.IsDefault,
			CloudNetworkAclID: data.Extension.CloudNetworkAclID,
		},
	}

	return subnetReq
}
