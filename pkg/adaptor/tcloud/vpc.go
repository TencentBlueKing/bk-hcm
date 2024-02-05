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
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateVpc create vpc.
// reference: https://cloud.tencent.com/document/api/215/15774
func (t *TCloudImpl) CreateVpc(kt *kit.Kit, opt *types.TCloudVpcCreateOption) (*types.TCloudVpc, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := t.clientSet.VpcClient(opt.Extension.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateVpcRequest()
	req.VpcName = converter.ValToPtr(opt.Name)
	req.CidrBlock = converter.ValToPtr(opt.Extension.IPv4Cidr)

	resp, err := vpcClient.CreateVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tencent cloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	handler := &createVpcPollingHandler{
		opt.Extension.Region,
	}
	respPoller := poller.Poller[*TCloudImpl, []*vpc.Vpc, []*types.TCloudVpc]{Handler: handler}
	results, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.Vpc.VpcId},
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
func (t *TCloudImpl) UpdateVpc(_ *kit.Kit, _ *types.TCloudVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://cloud.tencent.com/document/api/215/15775
func (t *TCloudImpl) DeleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteVpcRequest()
	req.VpcId = converter.ValToPtr(opt.ResourceID)

	_, err = VpcClient.DeleteVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tencent cloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://cloud.tencent.com/document/api/215/15778
func (t *TCloudImpl) ListVpc(kt *kit.Kit, opt *core.TCloudListOption) (*types.TCloudVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeVpcsRequest()
	if len(opt.CloudIDs) != 0 {
		req.VpcIds = converter.SliceToPtr(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.Offset = converter.ValToPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = converter.ValToPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := VpcClient.DescribeVpcsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tencent cloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tencent cloud vpc failed, err: %v", err)
	}

	details := make([]types.TCloudVpc, 0, len(resp.Response.VpcSet))

	for _, data := range resp.Response.VpcSet {
		details = append(details, converter.PtrToVal(convertVpc(data, opt.Region)))
	}

	return &types.TCloudVpcListResult{Count: resp.Response.TotalCount, Details: details}, nil
}

// CountVpc 基于 DescribeVpcsWithContext
// reference: https://cloud.tencent.com/document/api/215/15778
func (t *TCloudImpl) CountVpc(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeVpcsRequest()
	req.Limit = converter.ValToPtr("1")
	resp, err := client.DescribeVpcsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud vpc failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

func convertVpc(data *vpc.Vpc, region string) *types.TCloudVpc {
	if data == nil {
		return nil
	}

	v := &types.TCloudVpc{
		CloudID: converter.PtrToVal(data.VpcId),
		Name:    converter.PtrToVal(data.VpcName),
		Region:  region,
		Extension: &cloud.TCloudVpcExtension{
			Cidr:            nil,
			IsDefault:       converter.PtrToVal(data.IsDefault),
			EnableMulticast: converter.PtrToVal(data.EnableMulticast),
			DnsServerSet:    converter.PtrToSlice(data.DnsServerSet),
			DomainName:      converter.PtrToVal(data.DomainName),
		},
	}

	if data.CidrBlock != nil && len(*data.CidrBlock) != 0 {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.TCloudCidr{
			Type:     enumor.Ipv4,
			Cidr:     *data.CidrBlock,
			Category: enumor.MasterTCloudCidr,
		})
	}

	if data.Ipv6CidrBlock != nil && len(*data.Ipv6CidrBlock) != 0 {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.TCloudCidr{
			Type:     enumor.Ipv6,
			Cidr:     *data.Ipv6CidrBlock,
			Category: enumor.MasterTCloudCidr,
		})
	}

	for _, asstCidr := range data.AssistantCidrSet {
		if asstCidr == nil {
			continue
		}

		cidrBlock := converter.PtrToVal(asstCidr.CidrBlock)
		addressType, err := cidr.CidrIPAddressType(cidrBlock)
		if err != nil {
			logs.Errorf("get cidr ip address type failed, cidr: %v, err: %v", cidrBlock, err)
		}

		tcloudCidr := cloud.TCloudCidr{
			Type: addressType,
			Cidr: cidrBlock,
		}

		switch converter.PtrToVal(asstCidr.AssistantType) {
		case 0:
			tcloudCidr.Category = enumor.AssistantTCloudCidr
		case 1:
			tcloudCidr.Category = enumor.ContainerTCloudCidr
		}

		v.Extension.Cidr = append(v.Extension.Cidr, tcloudCidr)
	}

	return v
}

type createVpcPollingHandler struct {
	region string
}

// Done ...
func (h *createVpcPollingHandler) Done(vpcs []*vpc.Vpc) (bool, *[]*types.TCloudVpc) {
	results := make([]*types.TCloudVpc, 0)
	flag := true
	for _, vpc := range vpcs {
		if converter.PtrToVal(vpc.VpcId) == "" {
			flag = false
			continue
		}
		results = append(results, convertVpc(vpc, h.region))
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createVpcPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]*vpc.Vpc, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.TCloudQueryLimit)

	vpcs := make([]*vpc.Vpc, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := vpc.NewDescribeVpcsRequest()
		req.VpcIds = partIDs
		req.Limit = converter.ValToPtr(strconv.FormatUint(uint64(core.TCloudQueryLimit), 10))

		VpcClient, err := client.clientSet.VpcClient(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := VpcClient.DescribeVpcsWithContext(kt.Ctx, req)
		if err != nil {
			return nil, err
		}

		vpcs = append(vpcs, resp.Response.VpcSet...)
	}

	if len(vpcs) != len(cloudIDs) {
		return nil, fmt.Errorf("query vpc count: %d not equal return count: %d", len(cloudIDs), len(vpcs))
	}

	return vpcs, nil
}
