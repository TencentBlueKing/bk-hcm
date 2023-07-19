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

// AwsSubnetCreateExt defines create aws subnet extensional info.
type AwsSubnetCreateExt struct {
	Region   string  `json:"region" validate:"required"`
	Zone     *string `json:"zone" validate:"omitempty"`
	IPv4Cidr *string `json:"ipv4_cidr" validate:"omitempty,cidrv4"`
	IPv6Cidr *string `json:"ipv6_cidr" validate:"omitempty,cidrv6"`
}

// AwsSubnetCreateOption defines aws create subnet options.
type AwsSubnetCreateOption SubnetCreateOption[AwsSubnetCreateExt]

// Validate AwsSubnetCreateOption.
func (c AwsSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// AwsDefaultSubnetCreateOption defines create default aws subnet extensional info.
type AwsDefaultSubnetCreateOption struct {
	Region string `json:"region" validate:"required"`
	Zone   string `json:"zone" validate:"required"`
}

// Validate AwsDefaultSubnetCreateOption.
func (c AwsDefaultSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// AwsSubnetUpdateOption defines aws update subnet options.
type AwsSubnetUpdateOption struct{}

// Validate AwsSubnetUpdateOption.
func (s AwsSubnetUpdateOption) Validate() error {
	return nil
}

// AwsSubnetListResult defines aws list subnet result.
type AwsSubnetListResult struct {
	NextToken *string     `json:"next_token,omitempty"`
	Details   []AwsSubnet `json:"details"`
}

// AwsSubnetExtension defines aws subnet extensional info.
type AwsSubnetExtension struct {
	State                       string `json:"state"`
	Region                      string `json:"region"`
	Zone                        string `json:"zone"`
	IsDefault                   bool   `json:"is_default"`
	MapPublicIpOnLaunch         bool   `json:"map_public_ip_on_launch"`
	AssignIpv6AddressOnCreation bool   `json:"assign_ipv6_address_on_creation"`
	HostnameType                string `json:"hostname_type"`
	AvailableIPAddressCount     int64  `json:"available_ip_address_count"`
	TotalIpAddressCount         int64  `json:"total_ip_address_count"`
	UsedIpAddressCount          int64  `json:"used_ip_address_count"`
}

// AwsSubnet defines aws subnet.
type AwsSubnet Subnet[AwsSubnetExtension]

// GetCloudID ...
func (vpc AwsSubnet) GetCloudID() string {
	return vpc.CloudID
}
