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

	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateSubnet create subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/create-or-update?tabs=HTTP
func (a *Azure) CreateSubnet(kt *kit.Kit, opt *adtysubnet.AzureSubnetCreateOption) (*adtysubnet.AzureSubnet, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.subnetClient()
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := converter.PtrToVal(convertSubnetCreateReq(opt))
	vpc := parseIDToName(opt.CloudVpcID)
	resp, err := client.BeginCreateOrUpdate(kt.Ctx, opt.Extension.ResourceGroup, vpc, opt.Name, req, nil)
	if err != nil {
		return nil, err
	}

	res, err := resp.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return nil, err
	}

	subnet := convertSubnet(&res.Subnet, opt.Extension.ResourceGroup, opt.CloudVpcID)
	return subnet, nil
}

func convertSubnetCreateReq(opt *adtysubnet.AzureSubnetCreateOption) *armnetwork.Subnet {
	if opt == nil {
		return nil
	}

	subnet := &armnetwork.Subnet{
		Name: &opt.Name,
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: nil,
			AddressPrefixes: converter.SliceToPtr(append(opt.Extension.IPv4Cidr,
				opt.Extension.IPv6Cidr...)),
			ApplicationGatewayIPConfigurations: nil,
			Delegations:                        nil,
			IPAllocations:                      nil,
			NatGateway:                         nil,
			NetworkSecurityGroup:               nil,
			PrivateEndpointNetworkPolicies:     nil,
			PrivateLinkServiceNetworkPolicies:  nil,
			RouteTable:                         nil,
			ServiceEndpointPolicies:            nil,
			ServiceEndpoints:                   nil,
			IPConfigurationProfiles:            nil,
			IPConfigurations:                   nil,
			PrivateEndpoints:                   nil,
			ProvisioningState:                  nil,
			Purpose:                            nil,
			ResourceNavigationLinks:            nil,
			ServiceAssociationLinks:            nil,
		},
	}

	if len(opt.Extension.CloudRouteTableID) > 0 {
		subnet.Properties.RouteTable = &armnetwork.RouteTable{ID: &opt.Extension.CloudRouteTableID}
	}

	if len(opt.Extension.NetworkSecurityGroup) > 0 {
		subnet.Properties.NetworkSecurityGroup = &armnetwork.SecurityGroup{ID: &opt.Extension.NetworkSecurityGroup}
	}

	if len(opt.Extension.NatGateway) > 0 {
		subnet.Properties.NatGateway = &armnetwork.SubResource{ID: &opt.Extension.NatGateway}
	}

	return subnet
}

// UpdateSubnet update subnet.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Azure) UpdateSubnet(_ *kit.Kit, _ *adtysubnet.AzureSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/delete?tabs=HTTP
func (a *Azure) DeleteSubnet(kt *kit.Kit, opt *adtysubnet.AzureSubnetDeleteOption) error {
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
func (a *Azure) ListSubnet(kt *kit.Kit, opt *adtysubnet.AzureSubnetListOption) (*adtysubnet.AzureSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := a.clientSet.subnetClient()
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := new(armnetwork.SubnetsClientListOptions)

	vpcName := parseIDToName(opt.CloudVpcID)
	pager := subnetClient.NewListPager(opt.ResourceGroupName, vpcName, req)
	if err != nil {
		logs.Errorf("list azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure subnet failed, err: %v", err)
	}

	details := make([]adtysubnet.AzureSubnet, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure subnet but get next page failed, err: %v", err)
		}

		for _, subnet := range page.Value {
			details = append(details, converter.PtrToVal(convertSubnet(subnet, opt.ResourceGroupName, opt.CloudVpcID)))
		}
	}

	return &adtysubnet.AzureSubnetListResult{Details: details}, nil
}

// ListSubnetByPage list subnet by page.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/list?tabs=HTTP
func (a *Azure) ListSubnetByPage(kt *kit.Kit, opt *adtysubnet.AzureSubnetListOption) (
	*Pager[armnetwork.SubnetsClientListResponse, adtysubnet.AzureSubnet], error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := a.clientSet.subnetClient()
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	req := new(armnetwork.SubnetsClientListOptions)

	vpcName := parseIDToName(opt.CloudVpcID)
	azurePager := subnetClient.NewListPager(opt.ResourceGroupName, vpcName, req)

	pager := &Pager[armnetwork.SubnetsClientListResponse, adtysubnet.AzureSubnet]{
		pager: azurePager,
		resultHandler: &subnetResultHandler{
			resGroupName: opt.ResourceGroupName,
			cloudVpcID:   opt.CloudVpcID,
		},
	}

	return pager, nil
}

type subnetResultHandler struct {
	resGroupName string
	cloudVpcID   string
}

func (handler *subnetResultHandler) BuildResult(resp armnetwork.SubnetsClientListResponse) []adtysubnet.AzureSubnet {
	details := make([]adtysubnet.AzureSubnet, 0, len(resp.Value))
	for _, subnet := range resp.Value {
		details = append(details, converter.PtrToVal(convertSubnet(subnet, handler.resGroupName, handler.cloudVpcID)))
	}

	return details
}

// ListSubnetByID list subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/list?tabs=HTTP
func (a *Azure) ListSubnetByID(kt *kit.Kit, opt *adtysubnet.AzureSubnetListByIDOption) (
	*adtysubnet.AzureSubnetListResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetClient, err := a.clientSet.subnetClient()
	if err != nil {
		return nil, fmt.Errorf("new subnet client failed, err: %v", err)
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	req := new(armnetwork.SubnetsClientListOptions)
	vpcName := parseIDToName(opt.CloudVpcID)
	pager := subnetClient.NewListPager(opt.ResourceGroupName, vpcName, req)
	details := make([]adtysubnet.AzureSubnet, 0, len(idMap))
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure subnet but get next page failed, err: %v", err)
		}

		for _, one := range nextResult.Value {
			id := SPtrToLowerSPtr(one.ID)
			if _, exist := idMap[*id]; exist {
				details = append(details, converter.PtrToVal(convertSubnet(one, opt.ResourceGroupName, opt.CloudVpcID)))
				delete(idMap, *id)

				if len(idMap) == 0 {
					return &adtysubnet.AzureSubnetListResult{Details: details}, nil
				}
			}
		}
	}

	return &adtysubnet.AzureSubnetListResult{Details: details}, nil
}

func convertSubnet(data *armnetwork.Subnet, resourceGroup, cloudVpcID string) *adtysubnet.AzureSubnet {
	if data == nil {
		return nil
	}

	s := &adtysubnet.AzureSubnet{
		CloudVpcID: strings.ToLower(cloudVpcID),
		CloudID:    SPtrToLowerStr(data.ID),
		Name:       SPtrToLowerStr(data.Name),
		Extension: &adtysubnet.AzureSubnetExtension{
			ResourceGroupName: strings.ToLower(resourceGroup),
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
		s.Extension.NatGateway = SPtrToLowerStr(data.Properties.NatGateway.ID)
	}

	if data.Properties.NetworkSecurityGroup != nil {
		s.Extension.NetworkSecurityGroup = SPtrToLowerStr(data.Properties.NetworkSecurityGroup.ID)
	}

	if data.Properties.RouteTable != nil {
		s.Extension.CloudRouteTableID = SPtrToLowerSPtr(data.Properties.RouteTable.ID)
	}

	return s
}
