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
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

// CreateSubnet create subnet.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0001.html
func (h *HuaWei) CreateSubnet(kt *kit.Kit, opt *adtysubnet.HuaWeiSubnetCreateOption) (*adtysubnet.HuaWeiSubnet, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := h.clientSet.vpcClientV2(opt.Extension.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := &model.CreateSubnetRequest{
		Body: &model.CreateSubnetRequestBody{
			Subnet: &model.CreateSubnetOption{
				Name:             opt.Name,
				Description:      opt.Memo,
				Cidr:             opt.Extension.IPv4Cidr,
				VpcId:            opt.CloudVpcID,
				GatewayIp:        opt.Extension.GatewayIp,
				Ipv6Enable:       &opt.Extension.Ipv6Enable,
				DhcpEnable:       nil,
				PrimaryDns:       nil,
				SecondaryDns:     nil,
				DnsList:          nil,
				AvailabilityZone: opt.Extension.Zone,
				ExtraDhcpOpts:    nil,
			},
		},
	}

	resp, err := subnetClient.CreateSubnet(req)
	if err != nil {
		logs.Errorf("create huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	handler := &createSubnetPollingHandler{
		opt.Extension.Region,
		converter.ValToPtr(resp.Subnet.VpcId),
	}
	respPoller := poller.Poller[*HuaWei, []model.Subnet, []*adtysubnet.HuaWeiSubnet]{Handler: handler}
	results, err := respPoller.PollUntilDone(h, kt, []*string{converter.ValToPtr(resp.Subnet.Id)},
		types.NewBatchCreateSubnetPollerOption())
	if err != nil {
		return nil, err
	}

	if len(converter.PtrToVal(results)) <= 0 {
		return nil, fmt.Errorf("create subnet failed")
	}

	return (converter.PtrToVal(results))[0], nil
}

// UpdateSubnet update subnet.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0004.html
func (h *HuaWei) UpdateSubnet(kt *kit.Kit, opt *adtysubnet.HuaWeiSubnetUpdateOption) error {
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
func (h *HuaWei) DeleteSubnet(kt *kit.Kit, opt *adtysubnet.HuaWeiSubnetDeleteOption) error {
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
func (h *HuaWei) ListSubnet(kt *kit.Kit, opt *adtysubnet.HuaWeiSubnetListOption) (*adtysubnet.HuaWeiSubnetListResult, error) {
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

	if len(opt.CloudVpcID) != 0 {
		req.VpcId = &opt.CloudVpcID
	}

	resp, err := vpcClient.ListSubnets(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return new(adtysubnet.HuaWeiSubnetListResult), nil
		}
		logs.Errorf("list huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei subnet failed, err: %v", err)
	}

	subnets := converter.PtrToVal(resp.Subnets)
	details := make([]adtysubnet.HuaWeiSubnet, 0, len(subnets))

	for _, data := range subnets {
		details = append(details, converter.PtrToVal(convertSubnet(&data, opt.Region)))
	}

	return &adtysubnet.HuaWeiSubnetListResult{Details: details}, nil
}

// ListSubnetByID list subnet by id.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_subnet01_0003.html
func (h *HuaWei) ListSubnetByID(kt *kit.Kit, opt *adtysubnet.HuaWeiSubnetListByIDOption) (*adtysubnet.HuaWeiSubnetListResult,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	idMap := make(map[string]struct{}, len(opt.CloudIDs))
	for _, id := range opt.CloudIDs {
		idMap[id] = struct{}{}
	}

	subnets := make([]adtysubnet.HuaWeiSubnet, 0, len(opt.CloudIDs))
	req := &model.ListSubnetsRequest{
		Limit:  converter.ValToPtr(int32(core.HuaWeiQueryLimit)),
		Marker: nil,
		VpcId:  &opt.CloudVpcID,
	}
	for {
		resp, err := vpcClient.ListSubnets(req)
		if err != nil {
			logs.Errorf("list huawei subnet failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list huawei subnet failed, err: %v", err)
		}

		if resp.Subnets == nil || len(*resp.Subnets) == 0 {
			break
		}

		for _, one := range *resp.Subnets {
			if _, exist := idMap[one.Id]; exist {
				subnets = append(subnets, converter.PtrToVal(convertSubnet(&one, opt.Region)))
				delete(idMap, one.Id)

				if len(idMap) == 0 {
					break
				}
			}
		}

		if len(idMap) == 0 {
			break
		}

		if len(*resp.Subnets) < core.HuaWeiQueryLimit {
			break
		}

		tmp := *resp.Subnets
		req.Marker = converter.ValToPtr(tmp[len(tmp)-1].Id)
	}

	return &adtysubnet.HuaWeiSubnetListResult{Details: subnets}, nil
}

func convertSubnet(data *model.Subnet, region string) *adtysubnet.HuaWeiSubnet {
	if data == nil {
		return nil
	}

	s := &adtysubnet.HuaWeiSubnet{
		CloudVpcID: data.VpcId,
		CloudID:    data.Id,
		Name:       data.Name,
		Memo:       &data.Description,
		Extension: &adtysubnet.HuaWeiSubnetExtension{
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

type createSubnetPollingHandler struct {
	region string
	vpcID  *string
}

// Done ...
func (h *createSubnetPollingHandler) Done(subnets []model.Subnet) (bool, *[]*adtysubnet.HuaWeiSubnet) {
	results := make([]*adtysubnet.HuaWeiSubnet, 0)
	flag := true
	for _, subnet := range subnets {
		if subnet.Status.Value() != "ACTIVE" {
			flag = false
			continue
		}
		results = append(results, convertSubnet(converter.ValToPtr(subnet), h.region))
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createSubnetPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.Subnet, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	subnets := make([]model.Subnet, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		vpcClient, err := client.clientSet.vpcClientV2(h.region)
		if err != nil {
			return nil, fmt.Errorf("new subnet client failed, err: %v", err)
		}

		req := new(model.ListSubnetsRequest)
		req.VpcId = h.vpcID

		resp, err := vpcClient.ListSubnets(req)
		if err != nil {
			if strings.Contains(err.Error(), ErrDataNotFound) {
				return make([]model.Subnet, 0), nil
			}
			return nil, err
		}

		for _, subnet := range *resp.Subnets {
			for _, id := range partIDs {
				if subnet.Id == converter.PtrToVal(id) {
					subnets = append(subnets, subnet)
					break
				}
			}
		}
	}

	if len(subnets) != len(cloudIDs) {
		return nil, fmt.Errorf("query subnet count: %d not equal return count: %d", len(cloudIDs), len(subnets))
	}

	return subnets, nil
}
