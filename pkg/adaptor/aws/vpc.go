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
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateVpc create vpc.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_CreateVpc.html
func (a *Aws) CreateVpc(kt *kit.Kit, opt *types.AwsVpcCreateOption) (*types.AwsVpc, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Extension.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock:     aws.Bool(opt.Extension.AmazonProvidedIpv6CidrBlock),
		CidrBlock:                       aws.String(opt.Extension.IPv4Cidr),
		DryRun:                          nil,
		InstanceTenancy:                 aws.String(opt.Extension.InstanceTenancy),
		Ipv4IpamPoolId:                  nil,
		Ipv4NetmaskLength:               nil,
		Ipv6CidrBlock:                   nil,
		Ipv6CidrBlockNetworkBorderGroup: nil,
		Ipv6IpamPoolId:                  nil,
		Ipv6NetmaskLength:               nil,
		Ipv6Pool:                        nil,
		TagSpecifications:               genNameTags(vpcTagResType, opt.Name),
	}

	resp, err := client.CreateVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	handler := &createVpcPollingHandler{
		opt.Extension.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Vpc, []*types.AwsVpc]{Handler: handler}
	results, err := respPoller.PollUntilDone(a, kt, []*string{resp.Vpc.VpcId},
		types.NewBatchCreateVpcPollerOption())
	if err != nil {
		return nil, err
	}

	if len(converter.PtrToVal(results)) <= 0 {
		return nil, fmt.Errorf("create vpc failed")
	}

	return (converter.PtrToVal(results))[0], nil
}

// UpdateVpc update vpc.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Aws) UpdateVpc(kt *kit.Kit, opt *types.AwsVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DeleteVpc.html
func (a *Aws) DeleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.DeleteVpcInput{
		VpcId: aws.String(opt.ResourceID),
	}
	_, err = client.DeleteVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete aws vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc. 如果查询的ID不存在，会报错：InvalidVpcID.NotFound
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DescribeVpcs.html
func (a *Aws) ListVpc(kt *kit.Kit, opt *core.AwsListOption) (*types.AwsVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeVpcsInput)

	if len(opt.CloudIDs) != 0 {
		req.VpcIds = aws.StringSlice(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.NextToken = opt.Page.NextToken
		req.MaxResults = opt.Page.MaxResults
	}

	resp, err := client.DescribeVpcsWithContext(kt.Ctx, req)
	if err != nil {
		if !strings.Contains(err.Error(), ErrVpcNotFound) {
			logs.Errorf("list aws vpc failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, err
	}

	details := make([]types.AwsVpc, 0, len(resp.Vpcs))
	for _, vpc := range resp.Vpcs {
		details = append(details, converter.PtrToVal(convertVpc(vpc, opt.Region)))
	}

	return &types.AwsVpcListResult{NextToken: resp.NextToken, Details: details}, nil
}

// GetVpcAttribute get vpc enableDnsHostnames and enableDnsSupport attribute.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DescribeVpcAttribute.html
func (a *Aws) GetVpcAttribute(kt *kit.Kit, vpcID, region string) (bool, bool, error) {
	if len(vpcID) == 0 {
		return false, false, errf.New(errf.InvalidParameter, "vpc id can not be empty")
	}

	if len(region) == 0 {
		return false, false, errf.New(errf.InvalidParameter, "region can not be empty")
	}

	client, err := a.clientSet.ec2Client(region)
	if err != nil {
		return false, false, err
	}

	hostNameAttrReq := &ec2.DescribeVpcAttributeInput{
		Attribute: converter.ValToPtr(ec2.VpcAttributeNameEnableDnsHostnames),
		VpcId:     converter.ValToPtr(vpcID),
	}
	hostNameAttr, err := client.DescribeVpcAttributeWithContext(kt.Ctx, hostNameAttrReq)
	if err != nil {
		return false, false, err
	}

	enableDnsHostnames := false
	if hostNameAttr != nil && hostNameAttr.EnableDnsHostnames != nil && hostNameAttr.EnableDnsHostnames.Value != nil {
		enableDnsHostnames = *hostNameAttr.EnableDnsHostnames.Value
	}

	supportAttrReq := &ec2.DescribeVpcAttributeInput{
		Attribute: converter.ValToPtr(ec2.VpcAttributeNameEnableDnsSupport),
		VpcId:     converter.ValToPtr(vpcID),
	}
	supportAttr, err := client.DescribeVpcAttributeWithContext(kt.Ctx, supportAttrReq)
	if err != nil {
		return false, false, err
	}
	enableDnsSupport := false
	if supportAttr != nil && supportAttr.EnableDnsSupport != nil && supportAttr.EnableDnsSupport.Value != nil {
		enableDnsSupport = *supportAttr.EnableDnsSupport.Value
	}

	return enableDnsHostnames, enableDnsSupport, nil
}

func convertVpc(data *ec2.Vpc, region string) *types.AwsVpc {
	if data == nil {
		return nil
	}

	v := &types.AwsVpc{
		CloudID: converter.PtrToVal(data.VpcId),
		Region:  region,
		Extension: &cloud.AwsVpcExtension{
			State:           converter.PtrToVal(data.State),
			InstanceTenancy: converter.PtrToVal(data.InstanceTenancy),
			IsDefault:       converter.PtrToVal(data.IsDefault),
		},
	}

	name, _ := parseTags(data.Tags)
	v.Name = name

	if data.CidrBlock != nil {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.AwsCidr{
			Type: enumor.Ipv4,
			Cidr: *data.CidrBlock,
		})
	}

	for _, asst := range data.CidrBlockAssociationSet {
		if asst == nil || asst.CidrBlock == nil || *asst.CidrBlock == "" {
			continue
		}

		// update primary cidr state if cidr equals
		if *asst.CidrBlock == converter.PtrToVal(data.CidrBlock) {
			if asst.CidrBlockState != nil && asst.CidrBlockState.State != nil {
				v.Extension.Cidr[0].State = *asst.CidrBlockState.State
			}
			continue
		}

		cidr := cloud.AwsCidr{
			Type: enumor.Ipv4,
			Cidr: converter.PtrToVal(asst.CidrBlock),
		}

		if asst.CidrBlockState != nil && asst.CidrBlockState.State != nil {
			cidr.State = *asst.CidrBlockState.State
		}

		v.Extension.Cidr = append(v.Extension.Cidr, cidr)
	}

	for _, asst := range data.Ipv6CidrBlockAssociationSet {
		cidr := cloud.AwsCidr{
			Type:        enumor.Ipv6,
			Cidr:        converter.PtrToVal(asst.Ipv6CidrBlock),
			AddressPool: converter.PtrToVal(asst.Ipv6Pool),
		}

		if asst.Ipv6CidrBlockState != nil && asst.Ipv6CidrBlockState.State != nil {
			cidr.State = *asst.Ipv6CidrBlockState.State
		}

		v.Extension.Cidr = append(v.Extension.Cidr, cidr)
	}

	return v
}

type createVpcPollingHandler struct {
	region string
}

// Done ...
func (h *createVpcPollingHandler) Done(vpcs []*ec2.Vpc) (bool, *[]*types.AwsVpc) {
	results := make([]*types.AwsVpc, 0)
	flag := true
	for _, vpc := range vpcs {
		if converter.PtrToVal(vpc.State) == "pending" {
			flag = false
			continue
		}

		results = append(results, convertVpc(vpc, h.region))
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createVpcPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Vpc, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.AwsQueryLimit)

	vpcs := make([]*ec2.Vpc, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(ec2.DescribeVpcsInput)
		req.VpcIds = partIDs

		client, err := client.clientSet.ec2Client(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := client.DescribeVpcsWithContext(kt.Ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), ErrDataNotFound) {
				return make([]*ec2.Vpc, 0), nil
			}
			return nil, err
		}

		vpcs = append(vpcs, resp.Vpcs...)
	}

	if len(vpcs) != len(cloudIDs) {
		return nil, fmt.Errorf("query vpc count: %d not equal return count: %d", len(cloudIDs), len(vpcs))
	}

	return vpcs, nil
}
