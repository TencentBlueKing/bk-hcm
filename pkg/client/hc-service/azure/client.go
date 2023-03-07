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
	"hcm/pkg/rest"
)

// Client is a azure api client
type Client struct {
	Account       *AccountClient
	SecurityGroup *SecurityGroupClient
	Vpc           *VpcClient
	Subnet        *SubnetClient
	Eip           *EipClient
	Disk          *DiskClient
	Region        *RegionClient
	ResourceGroup *ResourceGroupClient
	Image         *ImageClient
	Cvm           *CvmClient
	RouteTable    *RouteTableClient
	InstanceType  *InstanceTypeClient
}

// NewClient create a new azure api client.
func NewClient(client rest.ClientInterface) *Client {
	return &Client{
		Account:       NewAccountClient(client),
		SecurityGroup: NewCloudSecurityGroupClient(client),
		Vpc:           NewVpcClient(client),
		Subnet:        NewSubnetClient(client),
		Eip:           NewEipClient(client),
		Disk:          NewCloudDiskClient(client),
		Region:        NewRegionClient(client),
		ResourceGroup: NewResourceGroupClient(client),
		Cvm:           NewCvmClient(client),
		Image:         NewCloudPublicClient(client),
		RouteTable:    NewRouteTableClient(client),
		InstanceType:  NewInstanceTypeClient(client),
	}
}
