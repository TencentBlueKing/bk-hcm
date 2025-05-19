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

	"hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateSecurityGroup create security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) CreateSecurityGroup(kt *kit.Kit, opt *securitygroup.AzureOption) (*securitygroup.AzureSecurityGroup,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	sg := armnetwork.SecurityGroup{
		Location: &opt.Region,
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.Name, sg, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return az.converCloudToSecurityGroup(&resp.SecurityGroup), nil
}

// DeleteSecurityGroup delete security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/delete?tabs=HTTP
func (az *Azure) DeleteSecurityGroup(kt *kit.Kit, opt *securitygroup.AzureOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	poller, err := client.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.Name, nil)
	if err != nil {
		logs.Errorf("request to BeginDelete failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginDelete result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

type sgResultHandler struct {
	resGroupName string
	az           *Azure
}

// BuildResult ...
func (handler *sgResultHandler) BuildResult(resp armnetwork.SecurityGroupsClientListResponse) []securitygroup.AzureSecurityGroup {
	sgs := make([]securitygroup.AzureSecurityGroup, 0, len(resp.Value))

	for _, one := range resp.Value {
		sgs = append(sgs, converter.PtrToVal(handler.az.converCloudToSecurityGroup(one)))
	}

	return sgs
}

// CountSecurityGroup count security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) CountSecurityGroup(kt *kit.Kit) (int32, error) {

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return 0, fmt.Errorf("new security group client failed, err: %v", err)
	}

	var count int32
	pager := client.NewListAllPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("list security group next page failed, err: %v, rid: %s", err, kt.Rid)
			return 0, fmt.Errorf("failed to advance page: %v", err)
		}

		count += int32(len(nextResult.Value))
	}

	return count, nil
}

// ListSecurityGroupByPage ...
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListSecurityGroupByPage(kt *kit.Kit, opt *securitygroup.AzureListOption) (
	*Pager[armnetwork.SecurityGroupsClientListResponse, securitygroup.AzureSecurityGroup], error) {

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	azurePager := client.NewListPager(opt.ResourceGroupName, nil)

	pager := &Pager[armnetwork.SecurityGroupsClientListResponse, securitygroup.AzureSecurityGroup]{
		pager: azurePager,
		resultHandler: &sgResultHandler{
			resGroupName: opt.ResourceGroupName,
		},
	}

	return pager, nil
}

// ListSecurityGroup list security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListSecurityGroup(kt *kit.Kit, opt *securitygroup.AzureListOption) (
	[]*securitygroup.AzureSecurityGroup, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	securityGroups := make([]*armnetwork.SecurityGroup, 0)
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		securityGroups = append(securityGroups, nextResult.Value...)
	}

	typesSecurityGroups := make([]*securitygroup.AzureSecurityGroup, 0)
	for _, v := range securityGroups {
		typesSecurityGroups = append(typesSecurityGroups, az.converCloudToSecurityGroup(v))
	}

	return typesSecurityGroups, nil
}

// ListSecurityGroupByID list security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListSecurityGroupByID(kt *kit.Kit, opt *core.AzureListByIDOption) (
	[]*securitygroup.AzureSecurityGroup, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	securityGroups := make([]*armnetwork.SecurityGroup, 0, len(idMap))
	typesSecurityGroups := make([]*securitygroup.AzureSecurityGroup, 0)
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			if len(opt.CloudIDs) > 0 {
				id := SPtrToLowerSPtr(one.ID)
				if _, exist := idMap[*id]; exist {
					securityGroups = append(securityGroups, one)
					delete(idMap, *id)

					if len(idMap) == 0 {
						for _, v := range securityGroups {
							typesSecurityGroups = append(typesSecurityGroups, az.converCloudToSecurityGroup(v))
						}
						return typesSecurityGroups, nil
					}
				}
			} else {
				securityGroups = append(securityGroups, one)
			}
		}
	}

	for _, v := range securityGroups {
		typesSecurityGroups = append(typesSecurityGroups, az.converCloudToSecurityGroup(v))
	}

	return typesSecurityGroups, nil
}

// ListRawSecurityGroupByID list security group, return raw response.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListRawSecurityGroupByID(kt *kit.Kit, opt *core.AzureListByIDOption) (
	[]*armnetwork.SecurityGroup, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(opt.CloudIDs) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cloud_ids is required")
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	securityGroups := make([]*armnetwork.SecurityGroup, 0, len(idMap))
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			id := SPtrToLowerSPtr(one.ID)
			if _, exist := idMap[*id]; exist {
				securityGroups = append(securityGroups, one)
				delete(idMap, *id)

				if len(idMap) == 0 {
					return securityGroups, nil
				}
			}
		}
	}

	return securityGroups, nil
}

func (az *Azure) converCloudToSecurityGroup(cloud *armnetwork.SecurityGroup) *securitygroup.AzureSecurityGroup {
	respSecurityGroup := &securitygroup.AzureSecurityGroup{
		ID:              SPtrToLowerSPtr(cloud.ID),
		Location:        SPtrToLowerNoSpaceSPtr(cloud.Location),
		Name:            SPtrToLowerSPtr(cloud.Name),
		Etag:            cloud.Etag,
		FlushConnection: nil,
		ResourceGUID:    nil,
	}
	if cloud.Properties != nil {
		respSecurityGroup.FlushConnection = cloud.Properties.FlushConnection
		respSecurityGroup.ResourceGUID = cloud.Properties.ResourceGUID
	}

	return respSecurityGroup
}

func (az *Azure) getSecurityGroupByCloudID(kt *kit.Kit, resGroupName, cloudID string) (*securitygroup.AzureSecurityGroup,
	error) {

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	resp, err := client.Get(kt.Ctx, resGroupName, parseIDToName(cloudID), new(armnetwork.SecurityGroupsClientGetOptions))
	if err != nil {
		logs.Errorf("get security group failed, err: %v, resGroupName: %s, cloudID: %s, rid: %s",
			err, resGroupName, cloudID, kt.Rid)
		return nil, err
	}

	sg := &securitygroup.AzureSecurityGroup{
		ID:              SPtrToLowerSPtr(resp.SecurityGroup.ID),
		Location:        SPtrToLowerNoSpaceSPtr(resp.SecurityGroup.Location),
		Name:            SPtrToLowerSPtr(resp.SecurityGroup.Name),
		Etag:            resp.SecurityGroup.Etag,
		FlushConnection: nil,
		ResourceGUID:    nil,
	}
	if resp.SecurityGroup.Properties != nil {
		sg.FlushConnection = resp.SecurityGroup.Properties.FlushConnection
		sg.ResourceGUID = resp.SecurityGroup.Properties.ResourceGUID
		sg.SecurityRules = resp.SecurityGroup.Properties.SecurityRules
	}

	return sg, nil
}

// SecurityGroupSubnetAssociate associate subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/create-or-update?tabs=HTTP
func (az *Azure) SecurityGroupSubnetAssociate(kt *kit.Kit, opt *securitygroup.AzureAssociateSubnetOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "associate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.subnetClient()
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	vpcName := parseIDToName(opt.CloudVpcID)
	pager := client.NewListPager(opt.ResourceGroupName, vpcName, new(armnetwork.SubnetsClientListOptions))
	if err != nil {
		logs.Errorf("list azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("list azure subnet failed, err: %v", err)
	}

	var subnet *armnetwork.Subnet
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return fmt.Errorf("list azure subnet but get next page failed, err: %v", err)
		}

		for _, one := range page.Value {
			if SPtrToLowerStr(one.ID) == opt.CloudSubnetID {
				subnet = one
			}
		}
	}

	if subnet == nil {
		return fmt.Errorf("subnet: %s not found", opt.CloudSubnetID)
	}

	subnet.Properties.NetworkSecurityGroup = &armnetwork.SecurityGroup{
		ID: converter.ValToPtr(opt.CloudSecurityGroupID),
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, vpcName, *subnet.Name, *subnet, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate subnet result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupSubnetDisassociate disassociate subnet.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/create-or-update?tabs=HTTP
func (az *Azure) SecurityGroupSubnetDisassociate(kt *kit.Kit, opt *securitygroup.AzureAssociateSubnetOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "disassociate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.subnetClient()
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	vpcName := parseIDToName(opt.CloudVpcID)
	pager := client.NewListPager(opt.ResourceGroupName, vpcName, new(armnetwork.SubnetsClientListOptions))
	if err != nil {
		logs.Errorf("list azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("list azure subnet failed, err: %v", err)
	}

	var subnet *armnetwork.Subnet
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return fmt.Errorf("list azure subnet but get next page failed, err: %v", err)
		}

		for _, one := range page.Value {
			if SPtrToLowerStr(one.ID) == opt.CloudSubnetID {
				subnet = one
			}
		}
	}

	if subnet == nil {
		return fmt.Errorf("subnet: %s not found", opt.CloudSubnetID)
	}

	if subnet.Properties.NetworkSecurityGroup == nil || subnet.Properties.NetworkSecurityGroup.ID == nil ||
		SPtrToLowerStr(subnet.Properties.NetworkSecurityGroup.ID) != opt.CloudSecurityGroupID {
		return fmt.Errorf("subnet: %s not associate security group: %s", opt.CloudSubnetID, opt.CloudSecurityGroupID)
	}

	subnet.Properties.NetworkSecurityGroup = nil
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, vpcName, *subnet.Name, *subnet, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate subnet result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupNetworkInterfaceAssociate associate network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/create-or-update?tabs=Go
func (az *Azure) SecurityGroupNetworkInterfaceAssociate(kt *kit.Kit,
	opt *securitygroup.AzureAssociateNetworkInterfaceOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "associate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	pager := client.NewListPager(opt.ResourceGroupName, new(armnetwork.InterfacesClientListOptions))
	if err != nil {
		logs.Errorf("list azure network interface failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("list azure interface failed, err: %v", err)
	}

	var inter *armnetwork.Interface
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return fmt.Errorf("list azure interface but get next page failed, err: %v", err)
		}

		for _, one := range page.Value {
			if SPtrToLowerStr(one.ID) == opt.CloudNetworkInterfaceID {
				inter = one
			}
		}
	}

	if inter == nil {
		return fmt.Errorf("network interface: %s not found", opt.CloudNetworkInterfaceID)
	}

	if inter.Properties.NetworkSecurityGroup != nil && inter.Properties.NetworkSecurityGroup.ID != nil {
		return fmt.Errorf("network interface: %s already associated security group: %s",
			opt.CloudNetworkInterfaceID, opt.CloudSecurityGroupID)
	}

	inter.Properties.NetworkSecurityGroup = &armnetwork.SecurityGroup{
		ID: converter.ValToPtr(opt.CloudSecurityGroupID),
	}

	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, *inter.Name, *inter,
		new(armnetwork.InterfacesClientBeginCreateOrUpdateOptions))
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate interface failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate interface result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupNetworkInterfaceDisassociate disassociate network interface.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/subnets/create-or-update?tabs=Go
func (az *Azure) SecurityGroupNetworkInterfaceDisassociate(kt *kit.Kit,
	opt *securitygroup.AzureAssociateNetworkInterfaceOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "disassociate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return fmt.Errorf("new subnet client failed, err: %v", err)
	}

	pager := client.NewListPager(opt.ResourceGroupName, new(armnetwork.InterfacesClientListOptions))
	if err != nil {
		logs.Errorf("list azure interface failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("list azure interface failed, err: %v", err)
	}

	var inter *armnetwork.Interface
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return fmt.Errorf("list azure interface but get next page failed, err: %v", err)
		}

		for _, one := range page.Value {
			if SPtrToLowerStr(one.ID) == opt.CloudNetworkInterfaceID {
				inter = one
			}
		}
	}

	if inter == nil {
		return fmt.Errorf("network interface: %s not found", opt.CloudNetworkInterfaceID)
	}

	if inter.Properties.NetworkSecurityGroup == nil || inter.Properties.NetworkSecurityGroup.ID == nil ||
		SPtrToLowerStr(inter.Properties.NetworkSecurityGroup.ID) != opt.CloudSecurityGroupID {
		return fmt.Errorf("network interface: %s not associate security group: %s", opt.CloudNetworkInterfaceID,
			opt.CloudSecurityGroupID)
	}

	inter.Properties.NetworkSecurityGroup = nil
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, *inter.Name, *inter,
		new(armnetwork.InterfacesClientBeginCreateOrUpdateOptions))
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate interface failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate interface result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
