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
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// SubnetBatchCreateReq defines batch create subnet request.
type SubnetBatchCreateReq[T SubnetCreateExtension] struct {
	Subnets []SubnetCreateReq[T] `json:"subnets" validate:"required"`
}

// SubnetCreateReq defines create subnet request.
type SubnetCreateReq[T SubnetCreateExtension] struct {
	AccountID         string   `json:"account_id" validate:"required"`
	CloudVpcID        string   `json:"cloud_vpc_id" validate:"required"`
	VpcID             string   `json:"vpc_id" validate:"required"`
	BkBizID           int64    `json:"bk_biz_id" validate:"required"`
	CloudRouteTableID string   `json:"cloud_route_table_id" validate:"omitempty"`
	RouteTableID      string   `json:"route_table_id" validate:"omitempty"`
	CloudID           string   `json:"cloud_id" validate:"required"`
	Name              *string  `json:"name,omitempty" validate:"required"`
	Region            string   `json:"region" validate:"omitempty"`
	Zone              string   `json:"zone" validate:"omitempty"`
	Ipv4Cidr          []string `json:"ipv4_cidr,omitempty" validate:"required"`
	Ipv6Cidr          []string `json:"ipv6_cidr,omitempty" validate:"omitempty"`
	Memo              *string  `json:"memo,omitempty" validate:"omitempty"`
	Extension         *T       `json:"extension" validate:"required"`
}

// Validate SubnetBatchCreateReq.
func (c *SubnetBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// SubnetCreateExtension defines create subnet extensional info.
type SubnetCreateExtension interface {
	TCloudSubnetCreateExt | AwsSubnetCreateExt | GcpSubnetCreateExt | AzureSubnetCreateExt | HuaWeiSubnetCreateExt
}

// TCloudSubnetCreateExt defines create tencent cloud subnet extensional info.
type TCloudSubnetCreateExt struct {
	IsDefault         bool    `json:"is_default" validate:"required"`
	CloudNetworkAclID *string `json:"cloud_network_acl_id,omitempty" validate:"omitempty"`
}

// AwsSubnetCreateExt defines create aws subnet extensional info.
type AwsSubnetCreateExt struct {
	State                       string `json:"state" validate:"required"`
	IsDefault                   bool   `json:"is_default" validate:"required"`
	MapPublicIpOnLaunch         bool   `json:"map_public_ip_on_launch" validate:"required"`
	AssignIpv6AddressOnCreation bool   `json:"assign_ipv6_address_on_creation" validate:"required"`
	HostnameType                string `json:"hostname_type" validate:"required"`
}

// GcpSubnetCreateExt defines create gcp subnet extensional info.
type GcpSubnetCreateExt struct {
	SelfLink              string `json:"self_link" validate:"required"`
	StackType             string `json:"stack_type" validate:"required"`
	Ipv6AccessType        string `json:"ipv6_access_type" validate:"required"`
	GatewayAddress        string `json:"gateway_address" validate:"required"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access" validate:"required"`
	EnableFlowLogs        bool   `json:"enable_flow_logs" validate:"required"`
}

// AzureSubnetCreateExt defines create azure subnet extensional info.
type AzureSubnetCreateExt struct {
	ResourceGroupName    string `json:"resource_group_name" validate:"required"`
	NatGateway           string `json:"nat_gateway,omitempty" validate:"omitempty"`
	CloudSecurityGroupID string `json:"cloud_security_group_id,omitempty" validate:"omitempty"`
	SecurityGroupID      string `json:"security_group_id,omitempty" validate:"omitempty"`
}

// HuaWeiSubnetCreateExt defines create huawei subnet extensional info.
type HuaWeiSubnetCreateExt struct {
	Status       string   `json:"status" validate:"required"`
	DhcpEnable   bool     `json:"dhcp_enable" validate:"required"`
	GatewayIp    string   `json:"gateway_ip" validate:"required"`
	DnsList      []string `json:"dns_list" validate:"required"`
	NtpAddresses []string `json:"ntp_addresses" validate:"required"`
}

// -------------------------- Update --------------------------

// SubnetBatchUpdateReq defines batch update subnet request.
type SubnetBatchUpdateReq[T SubnetUpdateExtension] struct {
	Subnets []SubnetUpdateReq[T] `json:"subnets" validate:"required"`
}

// Validate SubnetBatchUpdateReq.
func (u *SubnetBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// SubnetUpdateReq defines update subnet request.
type SubnetUpdateReq[T SubnetUpdateExtension] struct {
	ID                   string `json:"id" validate:"required"`
	SubnetUpdateBaseInfo `json:",inline" validate:"omitempty"`
	Extension            *T `json:"extension" validate:"omitempty"`
}

// SubnetUpdateBaseInfo defines update subnet request base info.
type SubnetUpdateBaseInfo struct {
	Name              *string  `json:"name,omitempty" validate:"omitempty"`
	Ipv4Cidr          []string `json:"ipv4_cidr,omitempty" validate:"omitempty"`
	Ipv6Cidr          []string `json:"ipv6_cidr,omitempty" validate:"omitempty"`
	Memo              *string  `json:"memo,omitempty" validate:"omitempty"`
	BkBizID           int64    `json:"bk_biz_id,omitempty" validate:"omitempty"`
	CloudRouteTableID *string  `json:"cloud_route_table_id,omitempty" validate:"omitempty"`
	RouteTableID      *string  `json:"route_table_id" validate:"omitempty"`
}

// SubnetUpdateExtension defines subnet update request extensional info.
type SubnetUpdateExtension interface {
	TCloudSubnetUpdateExt | AwsSubnetUpdateExt | GcpSubnetUpdateExt | AzureSubnetUpdateExt | HuaWeiSubnetUpdateExt
}

// TCloudSubnetUpdateExt defines update tencent cloud subnet extensional info.
type TCloudSubnetUpdateExt struct {
	IsDefault         bool    `json:"is_default" validate:"omitempty"`
	Region            string  `json:"region" validate:"omitempty"`
	Zone              string  `json:"zone" validate:"omitempty"`
	CloudNetworkAclID *string `json:"cloud_network_acl_id,omitempty" validate:"omitempty"`
}

// AwsSubnetUpdateExt defines update aws subnet extensional info.
type AwsSubnetUpdateExt struct {
	State                       string `json:"state,omitempty" validate:"omitempty"`
	Region                      string `json:"region,omitempty" validate:"omitempty"`
	Zone                        string `json:"zone,omitempty" validate:"omitempty"`
	IsDefault                   *bool  `json:"is_default,omitempty" validate:"omitempty"`
	MapPublicIpOnLaunch         *bool  `json:"map_public_ip_on_launch,omitempty" validate:"omitempty"`
	AssignIpv6AddressOnCreation *bool  `json:"assign_ipv6_address_on_creation,omitempty" validate:"omitempty"`
	HostnameType                string `json:"hostname_type,omitempty" validate:"omitempty"`
}

// GcpSubnetUpdateExt defines update gcp subnet extensional info.
type GcpSubnetUpdateExt struct {
	StackType             string `json:"stack_type,omitempty" validate:"omitempty"`
	Ipv6AccessType        string `json:"ipv6_access_type,omitempty" validate:"omitempty"`
	GatewayAddress        string `json:"gateway_address,omitempty" validate:"omitempty"`
	PrivateIpGoogleAccess *bool  `json:"private_ip_google_access,omitempty" validate:"omitempty"`
	EnableFlowLogs        *bool  `json:"enable_flow_logs,omitempty" validate:"omitempty"`
}

// AzureSubnetUpdateExt defines update azure subnet extensional info.
type AzureSubnetUpdateExt struct {
	NatGateway           *string `json:"nat_gateway,omitempty" validate:"omitempty"`
	CloudSecurityGroupID *string `json:"cloud_security_group_id,omitempty" validate:"omitempty"`
	SecurityGroupID      *string `json:"security_group_id,omitempty" validate:"omitempty"`
}

// HuaWeiSubnetUpdateExt defines update huawei subnet extensional info.
type HuaWeiSubnetUpdateExt struct {
	Status       string   `json:"status,omitempty" validate:"omitempty"`
	DhcpEnable   *bool    `json:"dhcp_enable,omitempty" validate:"omitempty"`
	GatewayIp    string   `json:"gateway_ip,omitempty" validate:"omitempty"`
	DnsList      []string `json:"dns_list,omitempty" validate:"omitempty"`
	NtpAddresses []string `json:"ntp_addresses,omitempty" validate:"omitempty"`
}

// SubnetBaseInfoBatchUpdateReq defines batch update subnet base info request.
type SubnetBaseInfoBatchUpdateReq struct {
	Subnets []SubnetBaseInfoUpdateReq `json:"subnets" validate:"required"`
}

// Validate SubnetBaseInfoBatchUpdateReq.
func (u *SubnetBaseInfoBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// SubnetBaseInfoUpdateReq defines update subnet base info request.
type SubnetBaseInfoUpdateReq struct {
	IDs  []string              `json:"id" validate:"required"`
	Data *SubnetUpdateBaseInfo `json:"data" validate:"required"`
}

// -------------------------- Get --------------------------

// SubnetGetResp defines get subnet response.
type SubnetGetResp[T cloud.SubnetExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *cloud.Subnet[T] `json:"data"`
}

// -------------------------- List --------------------------

// SubnetListResp defines list subnet response.
type SubnetListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SubnetListResult `json:"data"`
}

// SubnetListResult defines list subnet result.
type SubnetListResult struct {
	Count   uint64             `json:"count"`
	Details []cloud.BaseSubnet `json:"details"`
}

// SubnetExtListResult define subnet with extension list result.
type SubnetExtListResult[T cloud.SubnetExtension] struct {
	Count   uint64            `json:"count,omitempty"`
	Details []cloud.Subnet[T] `json:"details,omitempty"`
}

// SubnetExtListResp define list resp.
type SubnetExtListResp[T cloud.SubnetExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *SubnetExtListResult[T] `json:"data"`
}
