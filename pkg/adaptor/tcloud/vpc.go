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
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// UpdateVpc update vpc.
// TODO right now only memo is supported to update, add other update operations later.
func (t *TCloud) UpdateVpc(_ *kit.Kit, _ *types.TCloudVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://cloud.tencent.com/document/api/215/15775
func (t *TCloud) DeleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := t.clientSet.vpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteVpcRequest()
	req.VpcId = converter.ValToPtr(opt.ResourceID)

	_, err = vpcClient.DeleteVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tencent cloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://cloud.tencent.com/document/api/215/15778
func (t *TCloud) ListVpc(kt *kit.Kit, opt *core.TCloudListOption) (*types.TCloudVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := t.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeVpcsRequest()
	if len(opt.ResourceIDs) != 0 {
		req.VpcIds = converter.SliceToPtr(opt.ResourceIDs)
	}

	if opt.Page != nil {
		req.Offset = converter.ValToPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = converter.ValToPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := vpcClient.DescribeVpcsWithContext(kt.Ctx, req)
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

func convertVpc(data *vpc.Vpc, region string) *types.TCloudVpc {
	if data == nil {
		return nil
	}

	v := &types.TCloudVpc{
		CloudID: converter.PtrToVal(data.VpcId),
		Name:    converter.PtrToVal(data.VpcName),
		Extension: &cloud.TCloudVpcExtension{
			Region:          region,
			Cidr:            nil,
			IsDefault:       converter.PtrToVal(data.IsDefault),
			EnableMulticast: converter.PtrToVal(data.EnableMulticast),
			DnsServerSet:    converter.PtrToSlice(data.DnsServerSet),
			DomainName:      converter.PtrToVal(data.DomainName),
		},
	}

	if data.CidrBlock != nil {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.TCloudCidr{
			Type:     enumor.Ipv4,
			Cidr:     *data.CidrBlock,
			Category: enumor.MasterTCloudCidr,
		})
	}

	if data.Ipv6CidrBlock != nil {
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
