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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Aws) UpdateSubnet(_ *kit.Kit, _ *types.AwsSubnetUpdateOption) error {
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
func (a *Aws) ListSubnet(kt *kit.Kit, opt *core.AwsListOption) (*types.AwsSubnetListResult, error) {
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
	} else {
		req.NextToken = opt.Page.NextToken
		req.MaxResults = opt.Page.MaxResults
	}

	resp, err := client.DescribeSubnetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]types.AwsSubnet, 0, len(resp.Subnets))
	for _, subnet := range resp.Subnets {
		details = append(details, converter.PtrToVal(convertSubnet(subnet, opt.Region)))
	}

	return &types.AwsSubnetListResult{NextToken: resp.NextToken, Details: details}, nil
}

func convertSubnet(data *ec2.Subnet, region string) *types.AwsSubnet {
	if data == nil {
		return nil
	}

	s := &types.AwsSubnet{
		CloudVpcID: converter.PtrToVal(data.VpcId),
		CloudID:    converter.PtrToVal(data.SubnetId),
		Extension: &types.AwsSubnetExtension{
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

	if data.CidrBlock != nil && *data.CidrBlock != "" {
		s.Ipv4Cidr = []string{*data.CidrBlock}
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
