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
	"log"
	"strings"

	"hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// ListNetworkInterface list all network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (a *Azure) ListNetworkInterface(kt *kit.Kit) (*typesniproto.AzureInterfaceListResult, error) {
	client, err := a.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	details := make([]typesniproto.AzureNI, 0)
	pager := client.NewListAllPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, item := range nextResult.Value {
			details = append(details, converter.PtrToVal(convertCloudNetworkInterface(item)))
		}
	}

	return &typesniproto.AzureInterfaceListResult{Details: details}, nil
}

// ListNetworkInterfaceByID list all network interface by id.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (a *Azure) ListNetworkInterfaceByID(kt *kit.Kit, opt *core.AzureListByIDOption) (
	*typesniproto.AzureInterfaceListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	details := make([]typesniproto.AzureNI, 0, len(idMap))
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			if _, exist := idMap[*one.ID]; exist {
				details = append(details, converter.PtrToVal(convertCloudNetworkInterface(one)))
				delete(idMap, *one.ID)

				if len(idMap) == 0 {
					return &typesniproto.AzureInterfaceListResult{Details: details}, nil
				}
			}
		}
	}

	return &typesniproto.AzureInterfaceListResult{Details: details}, nil
}

func convertCloudNetworkInterface(data *armnetwork.Interface) *typesniproto.AzureNI {
	if data == nil {
		return nil
	}

	v := &typesniproto.AzureNI{
		Name:    data.Name,
		Region:  data.Location,
		CloudID: data.ID,
	}

	if data.Properties == nil {
		return v
	}

	v.Extension = &coreni.AzureNIExtension{
		Type:                        data.Type,
		EnableAcceleratedNetworking: data.Properties.EnableAcceleratedNetworking,
		EnableIPForwarding:          data.Properties.EnableIPForwarding,
		MacAddress:                  data.Properties.MacAddress,
	}
	if data.Properties.DNSSettings != nil {
		v.Extension.DNSSettings = &coreni.InterfaceDNSSettings{
			DNSServers:        data.Properties.DNSSettings.DNSServers,
			AppliedDNSServers: data.Properties.DNSSettings.AppliedDNSServers,
		}
	}
	if data.Properties.VirtualMachine != nil {
		v.Extension.CloudVirtualMachineID = data.Properties.VirtualMachine.ID
	}

	cloudIDArr := strings.Split(converter.PtrToVal(data.ID), "/")
	if len(cloudIDArr) > 4 {
		v.Extension.ResourceGroupName = cloudIDArr[4]
	}
	if data.Properties.VirtualMachine != nil {
		v.InstanceID = data.Properties.VirtualMachine.ID
	}
	getExtensionData(data, v)
	return v
}

func getExtensionData(data *armnetwork.Interface, v *typesniproto.AzureNI) {
	if data.Properties.NetworkSecurityGroup != nil {
		v.Extension.CloudSecurityGroupID = data.Properties.NetworkSecurityGroup.ID
	}
	getIpConfigExtensionData(data, v)

	if data.Properties.DNSSettings != nil {
		v.Extension.DNSSettings = &coreni.InterfaceDNSSettings{
			DNSServers:        data.Properties.DNSSettings.DNSServers,
			AppliedDNSServers: data.Properties.DNSSettings.AppliedDNSServers,
		}
	}
}

// getIpConfigExtensionData get ipconfig extension data
func getIpConfigExtensionData(data *armnetwork.Interface, v *typesniproto.AzureNI) {
	if data == nil || data.Properties == nil || data.Properties.IPConfigurations == nil {
		return
	}

	tmpArr := make([]*coreni.InterfaceIPConfiguration, 0)
	for _, item := range data.Properties.IPConfigurations {
		tmpIP := &coreni.InterfaceIPConfiguration{
			CloudID: item.ID,
			Name:    item.Name,
			Type:    item.Type,
		}
		if item.Properties != nil {
			tmpIP.Properties = &coreni.InterfaceIPConfigurationPropertiesFormat{
				Primary:                   item.Properties.Primary,
				PrivateIPAddress:          item.Properties.PrivateIPAddress,
				PrivateIPAddressVersion:   (*coreni.IPVersion)(item.Properties.PrivateIPAddressVersion),
				PrivateIPAllocationMethod: (*coreni.IPAllocationMethod)(item.Properties.PrivateIPAllocationMethod),
			}

			getIpConfigSubnetData(item, tmpIP, v)

			if converter.PtrToVal(tmpIP.Properties.Primary) {
				v.PrivateIP = tmpIP.Properties.PrivateIPAddress
			}
			if item.Properties.GatewayLoadBalancer != nil {
				tmpIP.Properties.CloudGatewayLoadBalancerID = item.Properties.GatewayLoadBalancer.ID
				v.Extension.CloudGatewayLoadBalancerID = tmpIP.Properties.CloudGatewayLoadBalancerID
			}
			if item.Properties.PublicIPAddress != nil {
				tmpPublicIPAddress := item.Properties.PublicIPAddress
				tmpIP.Properties.PublicIPAddress = &coreni.PublicIPAddress{
					CloudID:  tmpPublicIPAddress.ID,
					Location: tmpPublicIPAddress.Location,
					Zones:    tmpPublicIPAddress.Zones,
					Name:     tmpPublicIPAddress.Name,
					Type:     tmpPublicIPAddress.Type,
				}
				if tmpPublicIPAddress.Properties != nil {
					tmpIP.Properties.PublicIPAddress.Properties = &coreni.PublicIPAddressPropertiesFormat{
						IPAddress: tmpPublicIPAddress.Properties.IPAddress,
						PublicIPAddressVersion: (*coreni.IPVersion)(
							tmpPublicIPAddress.Properties.PublicIPAddressVersion),
						PublicIPAllocationMethod: (*coreni.IPAllocationMethod)(
							tmpPublicIPAddress.Properties.PublicIPAllocationMethod),
					}
				}
			}
		}
		tmpArr = append(tmpArr, tmpIP)
	}
	v.Extension.IPConfigurations = tmpArr
}

func getIpConfigSubnetData(item *armnetwork.InterfaceIPConfiguration, tmpIP *coreni.InterfaceIPConfiguration,
	v *typesniproto.AzureNI) {

	if item.Properties.Subnet != nil {
		return
	}

	tmpIP.Properties.CloudSubnetID = item.Properties.Subnet.ID
	ipSubnetArr := strings.Split(converter.PtrToVal(tmpIP.Properties.CloudSubnetID), "/")
	if len(ipSubnetArr) > 8 {
		v.CloudVpcID = converter.ValToPtr(ipSubnetArr[8])
	}

	v.CloudSubnetID = tmpIP.Properties.CloudSubnetID
}

// GetNetworkInterface get one network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/get
func (a *Azure) GetNetworkInterface(kt *kit.Kit, opt *core.AzureListOption) (*typesniproto.AzureNI, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.NetworkInterfaceName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "network interface name must be set")
	}

	client, err := a.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	res, err := client.Get(kt.Ctx, opt.ResourceGroupName, opt.NetworkInterfaceName, nil)
	if err != nil {
		logs.Errorf("get one azure network interface failed, rgName: %s, niName: %s, err: %v, rid: %s",
			opt.ResourceGroupName, opt.NetworkInterfaceName, err, kt.Rid)
		return nil, fmt.Errorf("get one azure network interface failed, err: %v", err)
	}

	niDetail := &armnetwork.Interface{
		ID:         res.ID,
		Name:       res.Name,
		Type:       res.Type,
		Location:   res.Location,
		Properties: res.Properties,
	}
	return convertCloudNetworkInterface(niDetail), nil
}

// ListNetworkSecurityGroup list network security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/
// list-effective-network-security-groups
func (a *Azure) ListNetworkSecurityGroup(kt *kit.Kit, opt *core.AzureListOption) (interface{}, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.NetworkInterfaceName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "network interface name must be set")
	}

	client, err := a.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface security_group client failed, err: %v", err)
	}

	poller, err := client.BeginListEffectiveNetworkSecurityGroups(kt.Ctx, opt.ResourceGroupName,
		opt.NetworkInterfaceName, nil)
	if err != nil {
		logs.Errorf("list all azure network interface security_group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list all azure network interface security_group failed, err: %v", err)
	}

	res, err := poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		log.Fatalf("list all azure network interface security_group failed to pull the result: %v", err)
	}

	return res, nil
}

// ListIP list all network interface's ip.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interface-ip-configurations/list
func (a *Azure) ListIP(kt *kit.Kit, opt *core.AzureListOption) ([]*coreni.InterfaceIPConfiguration, error) {
	client, err := a.clientSet.networkInterfaceIPConfigClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface ipconfig client failed, err: %v", err)
	}

	pager := client.NewListPager(opt.ResourceGroupName, opt.NetworkInterfaceName, nil)
	if err != nil {
		logs.Errorf("list all azure network interface ipconfig failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list all azure network interface ipconfig failed, err: %v", err)
	}

	details := make([]*coreni.InterfaceIPConfiguration, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list all azure network interface ipconfig but get next page failed, err: %v",
				err)
		}

		for _, item := range page.Value {
			details = append(details, convertCloudNetworkInterfaceIPConfig(item))
		}
	}

	return details, nil
}

func convertCloudNetworkInterfaceIPConfig(data *armnetwork.InterfaceIPConfiguration) *coreni.InterfaceIPConfiguration {
	if data == nil {
		return nil
	}

	v := &coreni.InterfaceIPConfiguration{
		CloudID: data.ID,
		Name:    data.Name,
		Type:    data.Type,
		Properties: &coreni.InterfaceIPConfigurationPropertiesFormat{
			Primary:                   data.Properties.Primary,
			PrivateIPAddress:          data.Properties.PrivateIPAddress,
			PrivateIPAddressVersion:   (*coreni.IPVersion)(data.Properties.PrivateIPAddressVersion),
			PrivateIPAllocationMethod: (*coreni.IPAllocationMethod)(data.Properties.PrivateIPAllocationMethod),
		},
	}
	if data.Properties.PublicIPAddress != nil {
		v.Properties.PublicIPAddress = &coreni.PublicIPAddress{
			CloudID:  data.Properties.PublicIPAddress.ID,
			Name:     data.Properties.PublicIPAddress.Name,
			Location: data.Properties.PublicIPAddress.Location,
			Zones:    data.Properties.PublicIPAddress.Zones,
		}
	}
	return v
}
