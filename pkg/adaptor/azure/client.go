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

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	armcomputev4 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

type clientSet struct {
	credential *types.AzureCredential
}

func newClientSet(credential *types.AzureCredential) *clientSet {
	return &clientSet{credential}
}

func (c *clientSet) subscriptionClient() (*armsubscription.SubscriptionsClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armsubscription.NewSubscriptionsClient(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure subscription client failed, err: %v", err)
	}

	return client, nil
}

func (c *clientSet) vpcClient() (*armnetwork.VirtualNetworksClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewVirtualNetworksClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure vpc client failed, err: %v", err)
	}
	return client, nil
}

func (c *clientSet) usageClient() (*armnetwork.UsagesClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewUsagesClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure usage client failed, err: %v", err)
	}
	return client, nil
}

func (c *clientSet) subnetClient() (*armnetwork.SubnetsClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewSubnetsClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure vpc client failed, err: %v", err)
	}
	return client, nil
}

func (c *clientSet) diskClient() (*armcompute.DisksClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	return armcompute.NewDisksClient(c.credential.CloudSubscriptionID, credential, nil)
}

func (c *clientSet) imageClient() (*armcomputev4.VirtualMachineImagesClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	return armcomputev4.NewVirtualMachineImagesClient(c.credential.CloudSubscriptionID, credential, nil)
}

func (c *clientSet) newClientSecretCredential() (*azidentity.ClientSecretCredential, error) {
	return azidentity.NewClientSecretCredential(
		c.credential.CloudTenantID,
		c.credential.CloudApplicationID,
		c.credential.CloudClientSecretKey, nil)
}

func (c *clientSet) securityGroupClient() (*armnetwork.SecurityGroupsClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewSecurityGroupsClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure security group client failed, err: %v", err)
	}

	return client, nil
}

func (c *clientSet) virtualMachineClient() (*armcompute.VirtualMachinesClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armcompute.NewVirtualMachinesClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure virtual machines client failed, err: %v", err)
	}

	return client, nil
}

func (c *clientSet) resourceGroupsClient() (*armresources.ResourceGroupsClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armresources.NewResourceGroupsClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init resourceGroups client failed, err: %v", err)
	}

	return client, nil
}

func (c *clientSet) regionClient() (*armsubscriptions.Client, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armsubscriptions.NewClient(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init region client failed, err: %v", err)
	}

	return client, nil
}

func (c *clientSet) routeTableClient() (*armnetwork.RouteTablesClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewRouteTablesClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure vpc client failed, err: %v", err)
	}
	return client, nil
}

func (c *clientSet) routeClient() (*armnetwork.RoutesClient, error) {
	credential, err := c.newClientSecretCredential()
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armnetwork.NewRoutesClient(c.credential.CloudSubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure vpc client failed, err: %v", err)
	}
	return client, nil
}
