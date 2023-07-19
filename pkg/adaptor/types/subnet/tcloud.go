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

package adtysubnet

import "hcm/pkg/criteria/validator"

// TCloudSubnetCreateExt defines tencent cloud create subnet extensional info.
type TCloudSubnetCreateExt struct {
	Region   string `json:"region" validate:"required"`
	Zone     string `json:"zone" validate:"required"`
	IPv4Cidr string `json:"ipv4_cidr" validate:"required,cidrv4"`
}

// TCloudSubnetCreateOption defines tcloud create subnet options.
type TCloudSubnetCreateOption SubnetCreateOption[TCloudSubnetCreateExt]

// Validate TCloudSubnetCreateOption.
func (c TCloudSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudSubnetsCreateOption defines create tencent cloud subnets options.
type TCloudSubnetsCreateOption struct {
	AccountID  string                     `json:"account_id" validate:"required"`
	Region     string                     `json:"region" validate:"required"`
	CloudVpcID string                     `json:"cloud_vpc_id" validate:"required"`
	Subnets    []TCloudOneSubnetCreateOpt `json:"subnets" validate:"min=1,max=100"`
}

// TCloudOneSubnetCreateOpt defines create one tencent cloud subnets options for TCloudSubnetsCreateOption.
type TCloudOneSubnetCreateOpt struct {
	IPv4Cidr          string `json:"ipv4_cidr" validate:"required,cidrv4"`
	Name              string `json:"name" validate:"required"`
	Zone              string `json:"zone" validate:"required"`
	CloudRouteTableID string `json:"cloud_route_table_id" validate:"omitempty"`
}

// Validate TCloudSubnetsCreateOption.
func (c TCloudSubnetsCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudSubnetUpdateOption defines tencent cloud update subnet options.
type TCloudSubnetUpdateOption struct{}

// Validate TCloudSubnetUpdateOption.
func (s TCloudSubnetUpdateOption) Validate() error {
	return nil
}

// TCloudSubnetListResult defines tencent cloud list subnet result.
type TCloudSubnetListResult struct {
	Count   *uint64        `json:"count,omitempty"`
	Details []TCloudSubnet `json:"details"`
}

// TCloudSubnetExtension defines tcloud subnet extensional info.
type TCloudSubnetExtension struct {
	IsDefault               bool    `json:"is_default"`
	Region                  string  `json:"region"`
	Zone                    string  `json:"zone"`
	CloudRouteTableID       *string `json:"cloud_route_table_id,omitempty"`
	CloudNetworkAclID       *string `json:"cloud_network_acl_id,omitempty"`
	AvailableIPAddressCount uint64  `json:"available_ip_address_count,omitempty"`
	TotalIpAddressCount     uint64  `json:"total_ip_address_count,omitempty"`
	UsedIpAddressCount      uint64  `json:"used_ip_address_count,omitempty"`
}

// TCloudSubnet defines tencent cloud subnet.
type TCloudSubnet Subnet[TCloudSubnetExtension]

// GetCloudID ...
func (vpc TCloudSubnet) GetCloudID() string {
	return vpc.CloudID
}
