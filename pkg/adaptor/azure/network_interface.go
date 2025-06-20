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
	"hcm/pkg/adaptor/types/eip"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CountNI count ni.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) CountNI(kt *kit.Kit) (int32, error) {

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return 0, fmt.Errorf("new netwrok interface client failed, err: %v", err)
	}

	var count int32
	pager := client.NewListAllPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("list network interface next page failed, err: %v, rid: %s", err, kt.Rid)
			return 0, fmt.Errorf("failed to advance page: %v", err)
		}

		count += int32(len(nextResult.Value))
	}

	return count, nil
}

// ListNetworkInterface list all network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) ListNetworkInterface(kt *kit.Kit) (*typesniproto.AzureInterfaceListResult, error) {
	client, err := az.clientSet.networkInterfaceClient()
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
			details = append(details, converter.PtrToVal(az.ConvertCloudNetworkInterface(kt, item)))
		}
	}

	return &typesniproto.AzureInterfaceListResult{Details: details}, nil
}

// ListNetworkInterfaceByPage list all network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) ListNetworkInterfaceByPage(kt *kit.Kit) (
	*Pager[armnetwork.InterfacesClientListAllResponse, typesniproto.AzureNI], error) {

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	azurePager := client.NewListAllPager(nil)

	pager := &Pager[armnetwork.InterfacesClientListAllResponse, typesniproto.AzureNI]{
		pager: azurePager,
		resultHandler: &niResultHandler{
			kt:  kt,
			cli: az,
		},
	}

	return pager, nil
}

type niResultHandler struct {
	kt  *kit.Kit
	cli *Azure
}

// BuildResult ...
func (handler *niResultHandler) BuildResult(resp armnetwork.InterfacesClientListAllResponse) []typesniproto.AzureNI {
	details := make([]typesniproto.AzureNI, 0, len(resp.Value))
	for _, one := range resp.Value {
		details = append(details, converter.PtrToVal(handler.cli.ConvertCloudNetworkInterface(handler.kt, one)))

	}

	return details
}

// ListNetworkInterfaceByID list all network interface by id.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) ListNetworkInterfaceByID(kt *kit.Kit, opt *core.AzureListByIDOption) (
	*typesniproto.AzureInterfaceListResult, error,
) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.networkInterfaceClient()
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
			id := SPtrToLowerSPtr(one.ID)
			if _, exist := idMap[*id]; exist {
				details = append(details, converter.PtrToVal(az.ConvertCloudNetworkInterface(kt, one)))
				delete(idMap, *id)

				if len(idMap) == 0 {
					return &typesniproto.AzureInterfaceListResult{Details: details}, nil
				}
			}
		}
	}

	return &typesniproto.AzureInterfaceListResult{Details: details}, nil
}

// ListRawNetworkInterfaceByIDs ...
func (az *Azure) ListRawNetworkInterfaceByIDs(kt *kit.Kit, opt *core.AzureListByIDOption) (
	[]*armnetwork.Interface, error,
) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	tmpCloudIDMap := converter.StringSliceToMap(opt.CloudIDs)
	networks := make([]*armnetwork.Interface, 0)
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list raw network interface failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			one.ID = converter.ValToPtr(converter.StrToLowerNoSpaceStr(*one.ID))
			if _, exist := tmpCloudIDMap[*one.ID]; exist {
				networks = append(networks, one)
				delete(tmpCloudIDMap, *one.ID)
			}
		}
	}

	return networks, nil
}

// ConvertCloudNetworkInterface ...
func (az *Azure) ConvertCloudNetworkInterface(kt *kit.Kit, data *armnetwork.Interface) *typesniproto.AzureNI {
	if data == nil {
		return nil
	}

	v := &typesniproto.AzureNI{
		Name:    SPtrToLowerSPtr(data.Name),
		Region:  SPtrToLowerSPtr(data.Location),
		CloudID: SPtrToLowerSPtr(data.ID),
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
		v.Extension.CloudVirtualMachineID = SPtrToLowerSPtr(data.Properties.VirtualMachine.ID)
	}

	cloudIDArr := strings.Split(converter.PtrToVal(data.ID), "/")
	if len(cloudIDArr) > 4 {
		v.Extension.ResourceGroupName = strings.ToLower(cloudIDArr[4])
	}
	if data.Properties.VirtualMachine != nil {
		v.InstanceID = SPtrToLowerSPtr(data.Properties.VirtualMachine.ID)
	}
	az.getExtensionData(kt, data, v)
	return v
}

func (az *Azure) getExtensionData(kt *kit.Kit, data *armnetwork.Interface, v *typesniproto.AzureNI) {
	if data.Properties.NetworkSecurityGroup != nil {
		v.Extension.CloudSecurityGroupID = SPtrToLowerSPtr(data.Properties.NetworkSecurityGroup.ID)
	}
	az.getIpConfigExtensionData(kt, data, v)

	if data.Properties.DNSSettings != nil {
		v.Extension.DNSSettings = &coreni.InterfaceDNSSettings{
			DNSServers:        data.Properties.DNSSettings.DNSServers,
			AppliedDNSServers: data.Properties.DNSSettings.AppliedDNSServers,
		}
	}
}

// getIpConfigExtensionData get ipconfig extension data
func (az *Azure) getIpConfigExtensionData(kt *kit.Kit, data *armnetwork.Interface, v *typesniproto.AzureNI) {
	if data == nil || data.Properties == nil || data.Properties.IPConfigurations == nil {
		return
	}

	tmpArr := make([]*coreni.InterfaceIPConfiguration, 0)
	for _, item := range data.Properties.IPConfigurations {
		tmpIP := &coreni.InterfaceIPConfiguration{
			CloudID: SPtrToLowerSPtr(item.ID),
			Name:    SPtrToLowerSPtr(item.Name),
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

			if *item.Properties.PrivateIPAddressVersion == armnetwork.IPVersionIPv4 {
				v.PrivateIPv4 = append(v.PrivateIPv4, converter.PtrToVal(tmpIP.Properties.PrivateIPAddress))
			} else {
				v.PrivateIPv6 = append(v.PrivateIPv6, converter.PtrToVal(tmpIP.Properties.PrivateIPAddress))
			}
			if item.Properties.GatewayLoadBalancer != nil {
				tmpIP.Properties.CloudGatewayLoadBalancerID = SPtrToLowerSPtr(item.Properties.GatewayLoadBalancer.ID)
				v.Extension.CloudGatewayLoadBalancerID = SPtrToLowerSPtr(tmpIP.Properties.CloudGatewayLoadBalancerID)
			}
			if item.Properties.PublicIPAddress != nil {
				tmpPublicIPAddress := item.Properties.PublicIPAddress
				tmpIP.Properties.PublicIPAddress = &coreni.PublicIPAddress{
					CloudID:  SPtrToLowerSPtr(tmpPublicIPAddress.ID),
					Location: SPtrToLowerNoSpaceSPtr(tmpPublicIPAddress.Location),
					Zones:    tmpPublicIPAddress.Zones,
					Name:     SPtrToLowerSPtr(tmpPublicIPAddress.Name),
					Type:     tmpPublicIPAddress.Type,
				}
				if len(v.Extension.ResourceGroupName) != 0 && tmpIP.Properties.PublicIPAddress.CloudID != nil {
					eipInfo, _ := az.GetEipByCloudID(kt, v.Extension.ResourceGroupName,
						converter.PtrToVal(tmpIP.Properties.PublicIPAddress.CloudID))
					if eipInfo != nil {
						tmpIP.Properties.PublicIPAddress.Name = eipInfo.Name
						tmpPublicIPAddress.Properties = &armnetwork.PublicIPAddressPropertiesFormat{
							IPAddress:              eipInfo.PublicIp,
							PublicIPAddressVersion: (*armnetwork.IPVersion)(eipInfo.PublicIPAddressVersion),
						}
					}
				}
				if tmpPublicIPAddress.Properties != nil && tmpPublicIPAddress.Properties.IPAddress != nil {
					if converter.PtrToVal(tmpPublicIPAddress.Properties.PublicIPAddressVersion) ==
						armnetwork.IPVersionIPv4 {
						v.PublicIPv4 = append(v.PublicIPv4, converter.PtrToVal(tmpPublicIPAddress.Properties.IPAddress))
					} else {
						v.PublicIPv6 = append(v.PublicIPv6, converter.PtrToVal(tmpPublicIPAddress.Properties.IPAddress))
					}
				}
			}
		}
		tmpArr = append(tmpArr, tmpIP)
	}
	v.Extension.IPConfigurations = tmpArr
}

func getIpConfigSubnetData(
	item *armnetwork.InterfaceIPConfiguration,
	tmpIP *coreni.InterfaceIPConfiguration,
	v *typesniproto.AzureNI,
) {
	if item.Properties.Subnet == nil {
		return
	}

	tmpIP.Properties.CloudSubnetID = item.Properties.Subnet.ID
	ipSubnetArr := strings.Split(converter.PtrToVal(tmpIP.Properties.CloudSubnetID), "/")
	if len(ipSubnetArr) > 9 {
		v.CloudVpcID = SPtrToLowerSPtr(converter.ValToPtr(strings.Join(ipSubnetArr[:9], "/")))
	}

	v.CloudSubnetID = SPtrToLowerSPtr(tmpIP.Properties.CloudSubnetID)
}

// GetNetworkInterface get one network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/get
func (az *Azure) GetNetworkInterface(kt *kit.Kit, opt *core.AzureListOption) (*typesniproto.AzureNI, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.NetworkInterfaceName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "network interface name must be set")
	}

	client, err := az.clientSet.networkInterfaceClient()
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
	return az.ConvertCloudNetworkInterface(kt, niDetail), nil
}

// ListNetworkSecurityGroup list network security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/
// list-effective-network-security-groups
func (az *Azure) ListNetworkSecurityGroup(kt *kit.Kit, opt *core.AzureListOption) (interface{}, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.NetworkInterfaceName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "network interface name must be set")
	}

	client, err := az.clientSet.networkInterfaceClient()
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
func (az *Azure) ListIP(kt *kit.Kit, opt *core.AzureListOption) ([]*coreni.InterfaceIPConfiguration, error) {
	client, err := az.clientSet.networkInterfaceIPConfigClient()
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
			details = append(details, convertCloudNIIPConfig(item))
		}
	}

	return details, nil
}

func convertCloudNIIPConfig(data *armnetwork.InterfaceIPConfiguration) *coreni.InterfaceIPConfiguration {
	if data == nil {
		return nil
	}

	v := &coreni.InterfaceIPConfiguration{
		CloudID: SPtrToLowerSPtr(data.ID),
		Name:    SPtrToLowerSPtr(data.Name),
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
			CloudID:  SPtrToLowerSPtr(data.Properties.PublicIPAddress.ID),
			Name:     SPtrToLowerSPtr(data.Properties.PublicIPAddress.Name),
			Location: SPtrToLowerNoSpaceSPtr(data.Properties.PublicIPAddress.Location),
			Zones:    data.Properties.PublicIPAddress.Zones,
		}
	}
	return v
}

// ListNetworkInterfacePage list network interface page.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) ListNetworkInterfacePage() (*runtime.Pager[armnetwork.InterfacesClientListAllResponse], error) {

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface cloud client failed, err: %v", err)
	}

	pager := client.NewListAllPager(nil)
	return pager, nil
}

// ListNetworkInterfaceByIDPage list network interface by id page.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-interfaces/list-all
func (az *Azure) ListNetworkInterfaceByIDPage(opt *core.AzureListByIDOption) (
	*runtime.Pager[armnetwork.InterfacesClientListResponse], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "new network interface client list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return nil, fmt.Errorf("new network interface client failed, err: %v", err)
	}

	pager := client.NewListPager(opt.ResourceGroupName, nil)
	return pager, nil
}

// GetEipByCloudID get eip info by cloudid
func (az *Azure) GetEipByCloudID(kt *kit.Kit, resourceGroupName, cloudPublicIP string) (*eip.AzureEip, error) {
	opt := &core.AzureListByIDOption{
		ResourceGroupName: resourceGroupName,
		CloudIDs:          []string{cloudPublicIP},
	}
	datas, err := az.ListEipByID(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(datas.Details) == 0 {
		return &eip.AzureEip{}, nil
	}

	return datas.Details[0], nil
}
