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

package azure

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateVpc create vpc.
// reference: https://docs.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/create-or-update
func (a *Azure) CreateVpc(kt *kit.Kit, opt *types.AzureVpcCreateOption) (*types.AzureVpc, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := armnetwork.VirtualNetwork{
		ExtendedLocation: nil,
		Location:         &opt.Extension.Region,
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{AddressPrefixes: converter.SliceToPtr(
				append(opt.Extension.IPv4Cidr, opt.Extension.IPv6Cidr...))},
			BgpCommunities:         nil,
			DdosProtectionPlan:     nil,
			DhcpOptions:            nil,
			EnableDdosProtection:   nil,
			EnableVMProtection:     nil,
			Encryption:             nil,
			FlowTimeoutInMinutes:   nil,
			IPAllocations:          nil,
			Subnets:                make([]*armnetwork.Subnet, 0, len(opt.Extension.Subnets)),
			VirtualNetworkPeerings: nil,
		},
	}
	for idx := range opt.Extension.Subnets {
		req.Properties.Subnets = append(req.Properties.Subnets, convertSubnetCreateReq(&opt.Extension.Subnets[idx]))
	}
	resp, err := client.BeginCreateOrUpdate(kt.Ctx, opt.Extension.ResourceGroup, opt.Name, req, nil)
	if err != nil {
		logs.Errorf("create azure vpc failed, err: %v, kt: %s", err, kt.Rid)
		return nil, errorf(err)
	}

	res, err := resp.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return nil, err
	}

	return convertVpc(&res.VirtualNetwork, opt.Extension.ResourceGroup), nil
}

// UpdateVpc update vpc.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Azure) UpdateVpc(kt *kit.Kit, opt *types.AzureVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://docs.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/delete
func (a *Azure) DeleteVpc(kt *kit.Kit, opt *core.AzureDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	poller, err := vpcClient.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.ResourceID, nil)
	if err != nil {
		logs.Errorf("delete azure vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/list
func (a *Azure) ListVpc(kt *kit.Kit, opt *core.AzureListOption) (*types.AzureVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := new(armnetwork.VirtualNetworksClientListOptions)

	pager := vpcClient.NewListPager(opt.ResourceGroupName, req)
	if err != nil {
		logs.Errorf("list azure vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure vpc failed, err: %v", err)
	}

	details := make([]types.AzureVpc, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure vpc but get next page failed, err: %v", err)
		}

		for _, vpc := range page.Value {
			details = append(details, converter.PtrToVal(convertVpc(vpc, opt.ResourceGroupName)))
		}
	}

	return &types.AzureVpcListResult{Details: details}, nil
}

// ListVpcByPage list vpc.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/list
func (a *Azure) ListVpcByPage(kt *kit.Kit, opt *core.AzureListOption) (
	*Pager[armnetwork.VirtualNetworksClientListResponse, types.AzureVpc], error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := new(armnetwork.VirtualNetworksClientListOptions)

	azurePager := vpcClient.NewListPager(opt.ResourceGroupName, req)

	pager := &Pager[armnetwork.VirtualNetworksClientListResponse, types.AzureVpc]{
		pager: azurePager,
		resultHandler: &vpcResultHandler{
			resGroupName: opt.ResourceGroupName,
		},
	}

	return pager, nil
}

type vpcResultHandler struct {
	resGroupName string
}

func (handler *vpcResultHandler) BuildResult(resp armnetwork.VirtualNetworksClientListResponse) []types.AzureVpc {
	details := make([]types.AzureVpc, 0, len(resp.Value))
	for _, vpc := range resp.Value {
		details = append(details, converter.PtrToVal(convertVpc(vpc, handler.resGroupName)))
	}

	return details
}

// ListVpcByID list vpc.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/list
func (a *Azure) ListVpcByID(kt *kit.Kit, opt *core.AzureListByIDOption) (*types.AzureVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	pager := vpcClient.NewListPager(opt.ResourceGroupName, new(armnetwork.VirtualNetworksClientListOptions))
	idMap := converter.StringSliceToMap(opt.CloudIDs)
	details := make([]types.AzureVpc, 0, len(idMap))
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure vpc but get next page failed, err: %v", err)
		}

		for _, one := range nextResult.Value {
			id := SPtrToLowerSPtr(one.ID)
			if _, exist := idMap[*id]; exist {
				details = append(details, converter.PtrToVal(convertVpc(one, opt.ResourceGroupName)))
				delete(idMap, *id)

				if len(idMap) == 0 {
					return &types.AzureVpcListResult{Details: details}, nil
				}
			}
		}
	}

	return &types.AzureVpcListResult{Details: details}, nil
}

func convertVpc(data *armnetwork.VirtualNetwork, resourceGroup string) *types.AzureVpc {
	if data == nil {
		return nil
	}

	v := &types.AzureVpc{
		CloudID: SPtrToLowerStr(data.ID),
		Name:    SPtrToLowerStr(data.Name),
		Region:  SPtrToLowerNoSpaceStr(data.Location),
		Extension: &types.AzureVpcExtension{
			ResourceGroupName: strings.ToLower(resourceGroup),
			DNSServers:        make([]string, 0),
			Cidr:              nil,
		},
	}

	if data.Properties == nil {
		return v
	}

	if data.Properties.DhcpOptions != nil {
		v.Extension.DNSServers = converter.PtrToSlice(data.Properties.DhcpOptions.DNSServers)
	}

	if data.Properties.AddressSpace != nil {
		for _, prefix := range data.Properties.AddressSpace.AddressPrefixes {
			if prefix == nil || *prefix == "" {
				continue
			}

			addressType, err := cidr.CidrIPAddressType(*prefix)
			if err != nil {
				logs.Errorf("get cidr ip address type failed, cidr: %v, err: %v", *prefix, err)
				continue
			}

			v.Extension.Cidr = append(v.Extension.Cidr, cloud.AzureCidr{
				Type: addressType,
				Cidr: *prefix,
			})
		}
	}

	for _, subnet := range data.Properties.Subnets {
		if subnet == nil {
			continue
		}

		v.Extension.Subnets = append(v.Extension.Subnets, *convertSubnet(subnet, resourceGroup, v.CloudID))
	}

	return v
}

// ListVpcUsage list vpc usage.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/list-usage?tabs=HTTP
func (a *Azure) ListVpcUsage(kt *kit.Kit, opt *types.AzureVpcListUsageOption) ([]types.VpcUsage,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	usageClient, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new usage client failed, err: %v", err)
	}

	req := new(armnetwork.VirtualNetworksClientListUsageOptions)

	vpcName := parseIDToName(opt.VpcID)
	pager := usageClient.NewListUsagePager(opt.ResourceGroupName, vpcName, req)
	if err != nil {
		logs.Errorf("list azure vpc usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure vpc usage failed, err: %v", err)
	}

	details := make([]armnetwork.VirtualNetworkUsage, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure vpc usage but get next page failed, err: %v", err)
		}

		details = append(details, converter.PtrToSlice(page.Value)...)
	}

	typesDetails := make([]types.VpcUsage, 0)
	for _, v := range details {
		tmp := types.VpcUsage{
			ID:           SPtrToLowerSPtr(v.ID),
			Limit:        v.Limit,
			CurrentValue: v.CurrentValue,
		}
		typesDetails = append(typesDetails, tmp)
	}

	return typesDetails, nil
}
