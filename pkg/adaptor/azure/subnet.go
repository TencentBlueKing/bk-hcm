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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Azure) UpdateSubnet(kt *kit.Kit, opt *types.AzureSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/delete?tabs=HTTP
func (a *Azure) DeleteSubnet(kt *kit.Kit, opt *types.AzureSubnetDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	subnetClient, err := a.clientSet.subnetClient()
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	vpcName := parseIDToName(opt.VpcID)
	poller, err := subnetClient.BeginDelete(kt.Ctx, opt.ResourceGroupName, vpcName, opt.ResourceID, nil)
	if err != nil {
		logs.Errorf("delete azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/list?tabs=HTTP
func (a *Azure) ListSubnet(kt *kit.Kit, opt *types.AzureSubnetListOption) (*types.AzureSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := a.clientSet.subnetClient()
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := new(armnetwork.SubnetsClientListOptions)

	vpcName := parseIDToName(opt.VpcID)
	pager := subnetClient.NewListPager(opt.ResourceGroupName, vpcName, req)
	if err != nil {
		logs.Errorf("list azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure subnet failed, err: %v", err)
	}

	details := make([]types.AzureSubnet, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure subnet but get next page failed, err: %v", err)
		}

		for _, subnet := range page.Value {
			details = append(details, converter.PtrToVal(convertSubnet(subnet,
				a.clientSet.credential.CloudSubscriptionID, opt.ResourceGroupName, vpcName)))
		}
	}

	return &types.AzureSubnetListResult{Details: details}, nil
}

func convertSubnet(data *armnetwork.Subnet, subscription, resourceGroup, vpc string) *types.AzureSubnet {
	if data == nil {
		return nil
	}

	s := &types.AzureSubnet{
		CloudVpcID: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s",
			subscription, resourceGroup, vpc),
		CloudID: converter.PtrToVal(data.ID),
		Name:    converter.PtrToVal(data.Name),
		Extension: &cloud.AzureSubnetExtension{
			ResourceGroup: resourceGroup,
		},
	}

	if data.Properties == nil {
		return s
	}

	for _, prefix := range append(data.Properties.AddressPrefixes, data.Properties.AddressPrefix) {
		if prefix != nil && *prefix != "" {
			addressType, err := cidr.CidrIPAddressType(*prefix)
			if err != nil {
				return nil
			}

			switch addressType {
			case enumor.Ipv4:
				s.Ipv4Cidr = append(s.Ipv4Cidr, *prefix)
			case enumor.Ipv6:
				s.Ipv6Cidr = append(s.Ipv6Cidr, *prefix)
			}
		}
	}

	if data.Properties.NatGateway != nil {
		s.Extension.NatGateway = converter.PtrToVal(data.Properties.NatGateway.ID)
	}

	if data.Properties.NetworkSecurityGroup != nil {
		s.Extension.NetworkSecurityGroup = converter.PtrToVal(data.Properties.NetworkSecurityGroup.ID)
	}

	return s
}
