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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateSubnet create subnet.
// reference: https://cloud.tencent.com/document/api/215/15782
func (t *TCloudImpl) CreateSubnet(kt *kit.Kit, opt *adtysubnet.TCloudSubnetCreateOption) (*adtysubnet.TCloudSubnet,
	error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := t.clientSet.VpcClient(opt.Extension.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewCreateSubnetRequest()
	req.VpcId = common.StringPtr(opt.CloudVpcID)
	req.SubnetName = common.StringPtr(opt.Name)
	req.CidrBlock = common.StringPtr(opt.Extension.IPv4Cidr)
	req.Zone = common.StringPtr(opt.Extension.Zone)

	resp, err := subnetClient.CreateSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return convertSubnet(resp.Response.Subnet, opt.Extension.Region), nil
}

// CreateSubnets create subnets.
// reference: https://cloud.tencent.com/document/api/215/31960
func (t *TCloudImpl) CreateSubnets(kt *kit.Kit, opt *adtysubnet.TCloudSubnetsCreateOption) ([]adtysubnet.TCloudSubnet,
	error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewCreateSubnetsRequest()
	req.VpcId = common.StringPtr(opt.CloudVpcID)
	for _, subnet := range opt.Subnets {
		one := &vpc.SubnetInput{
			CidrBlock:  common.StringPtr(subnet.IPv4Cidr),
			SubnetName: common.StringPtr(subnet.Name),
			Zone:       common.StringPtr(subnet.Zone),
		}

		if len(subnet.CloudRouteTableID) != 0 {
			one.RouteTableId = common.StringPtr(subnet.CloudRouteTableID)
		}
		req.Subnets = append(req.Subnets, one)
	}

	resp, err := subnetClient.CreateSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subnetIds := make([]*string, 0)
	for _, v := range resp.Response.SubnetSet {
		subnetIds = append(subnetIds, v.SubnetId)
	}

	handler := &createSubnetPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*TCloudImpl, []*vpc.Subnet, []adtysubnet.TCloudSubnet]{Handler: handler}
	results, err := respPoller.PollUntilDone(t, kt, subnetIds, types.NewBatchCreateSubnetPollerOption())
	if err != nil {
		return nil, err
	}

	return converter.PtrToVal(results), nil
}

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (t *TCloudImpl) UpdateSubnet(_ *kit.Kit, _ *adtysubnet.TCloudSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://cloud.tencent.com/document/api/215/15783
func (t *TCloudImpl) DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewDeleteSubnetRequest()
	req.SubnetId = converter.ValToPtr(opt.ResourceID)

	_, err = VpcClient.DeleteSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://cloud.tencent.com/document/api/215/15784
func (t *TCloudImpl) ListSubnet(kt *kit.Kit, opt *core.TCloudListOption) (*adtysubnet.TCloudSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := vpc.NewDescribeSubnetsRequest()
	if len(opt.CloudIDs) != 0 {
		req.SubnetIds = converter.SliceToPtr(opt.CloudIDs)
		req.Limit = converter.ValToPtr(strconv.FormatUint(core.TCloudQueryLimit, 10))
	}

	if opt.Page != nil {
		req.Offset = converter.ValToPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = converter.ValToPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := VpcClient.DescribeSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tencent cloud subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tencent cloud subnet failed, err: %v", err)
	}

	details := make([]adtysubnet.TCloudSubnet, 0, len(resp.Response.SubnetSet))

	for _, data := range resp.Response.SubnetSet {
		details = append(details, converter.PtrToVal(convertSubnet(data, opt.Region)))
	}

	return &adtysubnet.TCloudSubnetListResult{Count: resp.Response.TotalCount, Details: details}, nil
}

// CountSubnet 基于 DescribeSubnetsWithContext
// reference: https://cloud.tencent.com/document/api/215/15784
func (t *TCloudImpl) CountSubnet(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeSubnetsRequest()
	req.Limit = converter.ValToPtr("1")
	resp, err := client.DescribeSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud subnet failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

func convertSubnet(data *vpc.Subnet, region string) *adtysubnet.TCloudSubnet {
	if data == nil {
		return nil
	}

	s := &adtysubnet.TCloudSubnet{
		CloudVpcID: converter.PtrToVal(data.VpcId),
		CloudID:    converter.PtrToVal(data.SubnetId),
		Name:       converter.PtrToVal(data.SubnetName),
		Region:     region,
		Extension: &adtysubnet.TCloudSubnetExtension{
			IsDefault:               converter.PtrToVal(data.IsDefault),
			Zone:                    converter.PtrToVal(data.Zone),
			CloudRouteTableID:       data.RouteTableId,
			CloudNetworkAclID:       data.NetworkAclId,
			AvailableIPAddressCount: converter.PtrToVal(data.AvailableIpAddressCount),
			TotalIpAddressCount:     converter.PtrToVal(data.TotalIpAddressCount),
		},
	}

	if converter.PtrToVal(data.TotalIpAddressCount) > converter.PtrToVal(data.AvailableIpAddressCount) {
		s.Extension.UsedIpAddressCount = converter.PtrToVal(data.TotalIpAddressCount) -
			converter.PtrToVal(data.AvailableIpAddressCount)
	}

	if data.CidrBlock != nil && *data.CidrBlock != "" {
		s.Ipv4Cidr = append(s.Ipv4Cidr, *data.CidrBlock)
	}

	if data.Ipv6CidrBlock != nil && *data.Ipv6CidrBlock != "" {
		s.Ipv6Cidr = append(s.Ipv6Cidr, *data.Ipv6CidrBlock)
	}

	return s
}

type createSubnetPollingHandler struct {
	region string
}

// Done ...
func (h *createSubnetPollingHandler) Done(subnets []*vpc.Subnet) (bool, *[]adtysubnet.TCloudSubnet) {
	results := make([]adtysubnet.TCloudSubnet, 0)
	flag := true
	for _, subnet := range subnets {
		if converter.PtrToVal(subnet.SubnetId) == "" {
			flag = false
			continue
		}
		results = append(results, converter.PtrToVal(convertSubnet(subnet, h.region)))
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createSubnetPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]*vpc.Subnet, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.TCloudQueryLimit)

	subnets := make([]*vpc.Subnet, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := vpc.NewDescribeSubnetsRequest()
		req.SubnetIds = partIDs
		req.Limit = converter.ValToPtr(strconv.FormatUint(core.TCloudQueryLimit, 10))

		VpcClient, err := client.clientSet.VpcClient(h.region)
		if err != nil {
			return nil, fmt.Errorf("new subnet client failed, err: %v", err)
		}

		resp, err := VpcClient.DescribeSubnetsWithContext(kt.Ctx, req)
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, resp.Response.SubnetSet...)
	}

	if len(subnets) != len(cloudIDs) {
		return nil, fmt.Errorf("query subnet count: %d not equal return count: %d", len(cloudIDs), len(subnets))
	}

	return subnets, nil
}
