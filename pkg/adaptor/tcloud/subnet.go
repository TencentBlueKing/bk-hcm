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

package tcloud

import (
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (t *TCloud) UpdateSubnet(_ *kit.Kit, _ *types.TCloudSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://cloud.tencent.com/document/api/215/15783
func (t *TCloud) DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := t.clientSet.vpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewDeleteSubnetRequest()
	req.SubnetId = converter.ValToPtr(opt.ResourceID)

	_, err = vpcClient.DeleteSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://cloud.tencent.com/document/api/215/15784
func (t *TCloud) ListSubnet(kt *kit.Kit, opt *core.TCloudListOption) (*types.TCloudSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := t.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewDescribeSubnetsRequest()
	if len(opt.ResourceIDs) != 0 {
		req.SubnetIds = converter.SliceToPtr(opt.ResourceIDs)
	}

	if opt.Page != nil {
		req.Offset = converter.ValToPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = converter.ValToPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := vpcClient.DescribeSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tencent cloud subnet failed, err: %v", err)
	}

	details := make([]types.TCloudSubnet, 0, len(resp.Response.SubnetSet))

	for _, data := range resp.Response.SubnetSet {
		details = append(details, converter.PtrToVal(convertSubnet(data, opt.Region)))
	}

	return &types.TCloudSubnetListResult{Count: resp.Response.TotalCount, Details: details}, nil
}

func convertSubnet(data *vpc.Subnet, region string) *types.TCloudSubnet {
	if data == nil {
		return nil
	}

	s := &types.TCloudSubnet{
		CloudVpcID: converter.PtrToVal(data.VpcId),
		CloudID:    converter.PtrToVal(data.SubnetId),
		Name:       converter.PtrToVal(data.SubnetName),
		Extension: &types.TCloudSubnetExtension{
			IsDefault:               converter.PtrToVal(data.IsDefault),
			Region:                  region,
			Zone:                    converter.PtrToVal(data.Zone),
			CloudRouteTableID:       data.RouteTableId,
			CloudNetworkAclID:       data.NetworkAclId,
			AvailableIPAddressCount: converter.PtrToVal(data.AvailableIpAddressCount),
		},
	}

	if data.CidrBlock != nil && *data.CidrBlock != "" {
		s.Ipv4Cidr = append(s.Ipv4Cidr, *data.CidrBlock)
	}

	if data.Ipv6CidrBlock != nil && *data.Ipv6CidrBlock != "" {
		s.Ipv6Cidr = append(s.Ipv6Cidr, *data.Ipv6CidrBlock)
	}

	return s
}
