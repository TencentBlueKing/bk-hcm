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

package cloud

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// Subnet defines subnet info.
type Subnet[T SubnetExtension] struct {
	BaseSubnet `json:",inline"`
	Extension  *T `json:"extension"`
}

// GetID ...
func (subnet Subnet[T]) GetID() string {
	return subnet.BaseSubnet.ID
}

// GetCloudID ...
func (subnet Subnet[T]) GetCloudID() string {
	return subnet.BaseSubnet.CloudID
}

// BaseSubnet defines base subnet info.
type BaseSubnet struct {
	ID                string        `json:"id"`
	Vendor            enumor.Vendor `json:"vendor"`
	AccountID         string        `json:"account_id"`
	CloudVpcID        string        `json:"cloud_vpc_id"`
	CloudRouteTableID string        `json:"cloud_route_table_id"`
	CloudID           string        `json:"cloud_id"`
	Name              string        `json:"name"`
	Region            string        `json:"region"`
	Zone              string        `json:"zone"`
	Ipv4Cidr          []string      `json:"ipv4_cidr,omitempty"`
	Ipv6Cidr          []string      `json:"ipv6_cidr,omitempty"`
	Memo              *string       `json:"memo,omitempty"`
	VpcID             string        `json:"vpc_id"`
	RouteTableID      string        `json:"route_table_id"`
	BkBizID           int64         `json:"bk_biz_id"`
	*core.Revision    `json:",inline"`
}

// SubnetExtension defines subnet extensional info.
type SubnetExtension interface {
	TCloudSubnetExtension | AwsSubnetExtension | GcpSubnetExtension | AzureSubnetExtension | HuaWeiSubnetExtension
}

// TCloudSubnetExtension defines tencent cloud subnet extensional info.
type TCloudSubnetExtension struct {
	IsDefault         bool    `json:"is_default"`
	CloudNetworkAclId *string `json:"cloud_network_acl_id,omitempty"`
}

// AwsSubnetExtension defines aws subnet extensional info.
type AwsSubnetExtension struct {
	// TODO: state -> status
	State                       string `json:"state"`
	IsDefault                   bool   `json:"is_default"`
	MapPublicIpOnLaunch         bool   `json:"map_public_ip_on_launch"`
	AssignIpv6AddressOnCreation bool   `json:"assign_ipv6_address_on_creation"`
	HostnameType                string `json:"hostname_type"`
}

// GcpSubnetExtension defines gcp subnet extensional info.
type GcpSubnetExtension struct {
	VpcSelfLink           string `json:"vpc_self_link"`
	SelfLink              string `json:"self_link"`
	StackType             string `json:"stack_type"`
	Ipv6AccessType        string `json:"ipv6_access_type"`
	GatewayAddress        string `json:"gateway_address"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access"`
	EnableFlowLogs        bool   `json:"enable_flow_logs"`
}

// AzureSubnetExtension defines azure subnet extensional info.
type AzureSubnetExtension struct {
	ResourceGroupName    string `json:"resource_group_name"`
	NatGateway           string `json:"nat_gateway,omitempty"`
	SecurityGroupID      string `json:"security_group_id,omitempty"`
	CloudSecurityGroupID string `json:"cloud_security_group_id,omitempty"`
}

// HuaWeiSubnetExtension defines huawei subnet extensional info.
type HuaWeiSubnetExtension struct {
	Status       string   `json:"status"`
	DhcpEnable   bool     `json:"dhcp_enable"`
	GatewayIp    string   `json:"gateway_ip"`
	DnsList      []string `json:"dns_list"`
	NtpAddresses []string `json:"ntp_addresses"`
}

// TCloudSubnet defines tencent cloud subnet.
type TCloudSubnet Subnet[TCloudSubnetExtension]

// AwsSubnet defines aws subnet.
type AwsSubnet Subnet[AwsSubnetExtension]

// GcpSubnet defines gcp subnet.
type GcpSubnet Subnet[GcpSubnetExtension]

// AzureSubnet defines azure subnet.
type AzureSubnet Subnet[AzureSubnetExtension]

// HuaWeiSubnet defines huawei subnet.
type HuaWeiSubnet Subnet[HuaWeiSubnetExtension]
