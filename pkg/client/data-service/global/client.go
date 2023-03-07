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

package global

import (
	"hcm/pkg/rest"
)

// Client is a global api client
type Client struct {
	*restClient
	Auth                   *AuthClient
	Account                *AccountClient
	Cloud                  *CloudClient
	SecurityGroup          *SecurityGroupClient
	Vpc                    *VpcClient
	VpcCvmRel              *VpcCvmRelClient
	Subnet                 *SubnetClient
	SubnetCvmRel           *SubnetCvmRelClient
	Audit                  *AuditClient
	Zone                   *ZoneClient
	Cvm                    *CvmClient
	RouteTable             *RouteTableClient
	SGCvmRel               *SGCvmRelClient
	NetworkInterface       *NetworkInterfaceClient
	NetworkInterfaceCvmRel *NetworkInterfaceCvmRelClient
	Application            *ApplicationClient
	ApprovalProcess        *ApprovalProcessClient
}

type restClient struct {
	client rest.ClientInterface
}

// NewClient create a new global api client.
func NewClient(client rest.ClientInterface) *Client {
	return &Client{
		restClient:             &restClient{client: client},
		Auth:                   NewAuthClient(client),
		Account:                NewAccountClient(client),
		Cloud:                  NewCloudClient(client),
		SecurityGroup:          NewCloudSecurityGroupClient(client),
		Vpc:                    NewVpcClient(client),
		VpcCvmRel:              NewVpcCvmRelClient(client),
		Subnet:                 NewSubnetClient(client),
		SubnetCvmRel:           NewSubnetCvmRelClient(client),
		Audit:                  NewAuditClient(client),
		Zone:                   NewZoneClient(client),
		Cvm:                    NewCloudCvmClient(client),
		RouteTable:             NewRouteTableClient(client),
		SGCvmRel:               NewCloudSGCvmRelClient(client),
		NetworkInterface:       NewNetworkInterfaceClient(client),
		NetworkInterfaceCvmRel: NewNetworkInterfaceCvmRelClient(client),
		Application:            NewApplicationClient(client),
		ApprovalProcess:        NewApprovalProcessClient(client),
	}
}
