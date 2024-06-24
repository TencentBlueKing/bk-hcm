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

package gcp

import (
	"hcm/pkg/rest"
)

// Client is a gcp api client
type Client struct {
	Account          *AccountClient
	Firewall         *FirewallClient
	Vpc              *VpcClient
	Subnet           *SubnetClient
	Disk             *DiskClient
	Cvm              *CvmClient
	Image            *ImageClient
	RouteTable       *RouteTableClient
	Zone             *ZoneClient
	Region           *RegionClient
	Eip              *EipClient
	InstanceType     *InstanceTypeClient
	NetworkInterface *NetworkInterfaceClient
	Bill             *BillClient
	MainAccount      *MainAccountClient
}

// NewClient create a new gcp api client.
func NewClient(client rest.ClientInterface) *Client {
	return &Client{
		Account:          NewAccountClient(client),
		Firewall:         NewFirewallClient(client),
		Vpc:              NewVpcClient(client),
		Subnet:           NewSubnetClient(client),
		Disk:             NewCloudDiskClient(client),
		Cvm:              NewCvmClient(client),
		Image:            NewCloudPublicClient(client),
		RouteTable:       NewRouteTableClient(client),
		Zone:             NewZoneClient(client),
		Region:           NewRegionClient(client),
		Eip:              NewEipClient(client),
		InstanceType:     NewInstanceTypeClient(client),
		NetworkInterface: NewNetworkInterfaceClient(client),
		Bill:             NewBillClient(client),
		MainAccount:      NewMainAccountClient(client),
	}
}
