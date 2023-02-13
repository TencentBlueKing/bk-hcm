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

package huawei

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

// UpdateSubnet update subnet.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0004.html
func (h *HuaWei) UpdateSubnet(kt *kit.Kit, opt *types.HuaWeiSubnetUpdateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := &model.UpdateSubnetRequest{
		SubnetId: opt.ResourceID,
		VpcId:    opt.VpcID,
		Body: &model.UpdateSubnetRequestBody{
			Subnet: &model.UpdateSubnetOption{
				Name:        opt.Name,
				Description: opt.Data.Memo,
			},
		},
	}

	_, err = vpcClient.UpdateSubnet(req)
	if err != nil {
		logs.Errorf("create huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSubnet delete subnet.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0005.html
func (h *HuaWei) DeleteSubnet(kt *kit.Kit, opt *types.HuaWeiSubnetDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := &model.DeleteSubnetRequest{
		VpcId:    opt.VpcID,
		SubnetId: opt.ResourceID,
	}

	_, err = vpcClient.DeleteSubnet(req)
	if err != nil {
		logs.Errorf("delete huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0003.html
func (h *HuaWei) ListSubnet(kt *kit.Kit, opt *types.HuaWeiSubnetListOption) (*types.HuaWeiSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := new(model.ListSubnetsRequest)

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	if len(opt.VpcID) != 0 {
		req.VpcId = &opt.VpcID
	}

	resp, err := vpcClient.ListSubnets(req)
	if err != nil {
		logs.Errorf("list huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei subnet failed, err: %v", err)
	}

	subnets := converter.PtrToVal(resp.Subnets)
	details := make([]types.HuaWeiSubnet, 0, len(subnets))

	for _, data := range subnets {
		details = append(details, converter.PtrToVal(convertSubnet(&data, opt.Region)))
	}

	return &types.HuaWeiSubnetListResult{Details: details}, nil
}

func convertSubnet(data *model.Subnet, region string) *types.HuaWeiSubnet {
	if data == nil {
		return nil
	}

	s := &types.HuaWeiSubnet{
		CloudVpcID: data.VpcId,
		CloudID:    data.Id,
		Name:       data.Name,
		Memo:       &data.Description,
		Extension: &types.HuaWeiSubnetExtension{
			Region:     region,
			Status:     data.Status.Value(),
			DhcpEnable: data.DhcpEnable,
			GatewayIp:  data.GatewayIp,
			DnsList:    data.DnsList,
		},
	}

	if data.Cidr != "" {
		s.Ipv4Cidr = []string{data.Cidr}
	}

	if data.CidrV6 != "" {
		s.Ipv6Cidr = []string{data.CidrV6}
	}

	for _, opt := range data.ExtraDhcpOpts {
		switch opt.OptName {
		case model.GetExtraDhcpOptionOptNameEnum().NTP:
			if opt.OptValue != nil {
				s.Extension.NtpAddresses = append(s.Extension.NtpAddresses, *opt.OptValue)
			}
		}
	}

	return s
}

// GetSubnetIPAvailabilities get subnet ip availabilities.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_natworkip_0001.html
func (h *HuaWei) GetSubnetIPAvailabilities(kt *kit.Kit, opt *types.HuaWeiVpcIPAvailGetOption) (
	*model.NetworkIpAvailability, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.ShowNetworkIpAvailabilitiesRequest{
		NetworkId: opt.SubnetID,
	}

	resp, err := vpcClient.ShowNetworkIpAvailabilities(req)
	if err != nil {
		logs.Errorf("get huawei vpc ip availabilities failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.NetworkIpAvailability, nil
}
