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

package aws

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cidrtools "hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateSubnet create subnet.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateSubnet.html
func (a *Aws) CreateSubnet(kt *kit.Kit, opt *adtysubnet.AwsSubnetCreateOption) (*adtysubnet.AwsSubnet, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Extension.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.CreateSubnetInput{
		AvailabilityZone:   opt.Extension.Zone,
		AvailabilityZoneId: nil,
		CidrBlock:          opt.Extension.IPv4Cidr,
		DryRun:             nil,
		Ipv6CidrBlock:      opt.Extension.IPv6Cidr,
		Ipv6Native:         nil,
		OutpostArn:         nil,
		TagSpecifications:  genNameTags(subnetTagResType, opt.Name),
		VpcId:              aws.String(opt.CloudVpcID),
	}

	resp, err := client.CreateSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	handler := &createSubnetPollingHandler{
		opt.Extension.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Subnet, []*adtysubnet.AwsSubnet]{Handler: handler}
	results, err := respPoller.PollUntilDone(a, kt, []*string{resp.Subnet.SubnetId},
		types.NewBatchCreateSubnetPollerOption())
	if err != nil {
		return nil, err
	}

	if len(converter.PtrToVal(results)) <= 0 {
		return nil, fmt.Errorf("create subnet failed")
	}

	return (converter.PtrToVal(results))[0], nil
}

// CreateDefaultSubnet create default subnet.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateDefaultSubnet.html
func (a *Aws) CreateDefaultSubnet(kt *kit.Kit, opt *adtysubnet.AwsDefaultSubnetCreateOption) (*adtysubnet.AwsSubnet, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.CreateDefaultSubnetInput{
		AvailabilityZone: aws.String(opt.Zone),
		DryRun:           nil,
		Ipv6Native:       nil,
	}

	resp, err := client.CreateDefaultSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return convertSubnet(resp.Subnet, opt.Region), nil
}

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Aws) UpdateSubnet(_ *kit.Kit, _ *adtysubnet.AwsSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DeleteSubnet.html
func (a *Aws) DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(opt.ResourceID),
	}
	_, err = client.DeleteSubnetWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete aws subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DescribeSubnets.html
func (a *Aws) ListSubnet(kt *kit.Kit, opt *core.AwsListOption) (*adtysubnet.AwsSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeSubnetsInput)

	if len(opt.CloudIDs) != 0 {
		req.SubnetIds = aws.StringSlice(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.NextToken = opt.Page.NextToken
		req.MaxResults = opt.Page.MaxResults
	}

	resp, err := client.DescribeSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		if !strings.Contains(err.Error(), ErrSubnetNotFound) {
			logs.Errorf("list aws subnet failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, err
	}

	details := make([]adtysubnet.AwsSubnet, 0, len(resp.Subnets))
	for _, subnet := range resp.Subnets {
		details = append(details, converter.PtrToVal(convertSubnet(subnet, opt.Region)))
	}

	return &adtysubnet.AwsSubnetListResult{NextToken: resp.NextToken, Details: details}, nil
}

func convertSubnet(data *ec2.Subnet, region string) *adtysubnet.AwsSubnet {
	if data == nil {
		return nil
	}

	s := &adtysubnet.AwsSubnet{
		CloudVpcID: converter.PtrToVal(data.VpcId),
		CloudID:    converter.PtrToVal(data.SubnetId),
		Extension: &adtysubnet.AwsSubnetExtension{
			State:                       converter.PtrToVal(data.State),
			Region:                      region,
			Zone:                        converter.PtrToVal(data.AvailabilityZone),
			IsDefault:                   converter.PtrToVal(data.DefaultForAz),
			MapPublicIpOnLaunch:         converter.PtrToVal(data.MapPublicIpOnLaunch),
			AssignIpv6AddressOnCreation: converter.PtrToVal(data.AssignIpv6AddressOnCreation),
			AvailableIPAddressCount:     converter.PtrToVal(data.AvailableIpAddressCount),
		},
	}

	name, _ := parseTags(data.Tags)
	s.Name = name

	if data.CidrBlock != nil && converter.PtrToVal(data.CidrBlock) != "" {
		cidr := converter.PtrToVal(data.CidrBlock)
		s.Ipv4Cidr = []string{cidr}
		ips, err := cidrtools.CidrIPCounts(cidr)
		if err == nil {
			s.Extension.TotalIpAddressCount = int64(ips)
			s.Extension.UsedIpAddressCount = int64(ips) - converter.PtrToVal(data.AvailableIpAddressCount)
		}
	}

	for _, association := range data.Ipv6CidrBlockAssociationSet {
		if association != nil && association.Ipv6CidrBlock != nil && *association.Ipv6CidrBlock != "" {
			s.Ipv6Cidr = append(s.Ipv6Cidr, *association.Ipv6CidrBlock)
		}
	}
	if data.PrivateDnsNameOptionsOnLaunch != nil {
		s.Extension.HostnameType = converter.PtrToVal(data.PrivateDnsNameOptionsOnLaunch.HostnameType)
	}

	return s
}

type createSubnetPollingHandler struct {
	region string
}

// Done ...
func (h *createSubnetPollingHandler) Done(subnets []*ec2.Subnet) (bool, *[]*adtysubnet.AwsSubnet) {
	results := make([]*adtysubnet.AwsSubnet, 0)
	flag := true
	for _, subnet := range subnets {
		if converter.PtrToVal(subnet.State) == "pending" {
			flag = false
			continue
		}

		results = append(results, convertSubnet(subnet, h.region))
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createSubnetPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Subnet, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.AwsQueryLimit)

	subnets := make([]*ec2.Subnet, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(ec2.DescribeSubnetsInput)
		req.SubnetIds = aws.StringSlice(converter.PtrToSlice(partIDs))

		client, err := client.clientSet.ec2Client(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := client.DescribeSubnetsWithContext(kt.Ctx, req)
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, resp.Subnets...)
	}

	if len(subnets) != len(cloudIDs) {
		return nil, fmt.Errorf("query subnet count: %d not equal return count: %d", len(cloudIDs), len(subnets))
	}

	return subnets, nil
}
