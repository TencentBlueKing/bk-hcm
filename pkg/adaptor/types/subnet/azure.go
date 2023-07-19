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

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// AzureSubnetCreateExt defines create azure subnet extensional info.
type AzureSubnetCreateExt struct {
	ResourceGroup        string   `json:"resource_group" validate:"required"`
	IPv4Cidr             []string `json:"ipv4_cidr" validate:"required,dive,cidrv4"`
	IPv6Cidr             []string `json:"ipv6_cidr" validate:"omitempty,dive,cidrv6"`
	CloudRouteTableID    string   `json:"cloud_route_table_id,omitempty" validate:"omitempty"`
	NatGateway           string   `json:"nat_gateway,omitempty" validate:"omitempty"`
	NetworkSecurityGroup string   `json:"network_security_group,omitempty" validate:"omitempty"`
}

// AzureSubnetCreateOption defines azure create subnet options.
type AzureSubnetCreateOption SubnetCreateOption[AzureSubnetCreateExt]

// Validate AzureSubnetCreateOption.
func (c AzureSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureSubnetUpdateOption defines azure update subnet options.
type AzureSubnetUpdateOption struct{}

// Validate AzureSubnetUpdateOption.
func (s AzureSubnetUpdateOption) Validate() error {
	return nil
}

// AzureSubnetDeleteOption defines azure delete subnet options.
type AzureSubnetDeleteOption struct {
	core.AzureDeleteOption `json:",inline"`
	VpcID                  string `json:"vpc_id"`
}

// Validate AzureSubnetDeleteOption.
func (a AzureSubnetDeleteOption) Validate() error {
	if err := a.AzureDeleteOption.Validate(); err != nil {
		return err
	}

	if len(a.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id must be set")
	}

	return nil
}

// AzureSubnetListOption defines azure list subnet options.
type AzureSubnetListOption struct {
	core.AzureListOption `json:",inline"`
	CloudVpcID           string `json:"cloud_vpc_id"`
}

// Validate AzureSubnetListOption.
func (a AzureSubnetListOption) Validate() error {
	if err := a.AzureListOption.Validate(); err != nil {
		return err
	}

	if len(a.CloudVpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id must be set")
	}

	return nil
}

// AzureSubnetListByIDOption defines azure list subnet options.
type AzureSubnetListByIDOption struct {
	core.AzureListByIDOption `json:",inline"`
	CloudVpcID               string `json:"cloud_vpc_id"`
}

// Validate AzureSubnetListOption.
func (a AzureSubnetListByIDOption) Validate() error {
	if err := a.AzureListByIDOption.Validate(); err != nil {
		return err
	}

	if len(a.CloudVpcID) == 0 {
		return errf.New(errf.InvalidParameter, "cloud vpc id must be set")
	}

	return nil
}

// AzureSubnetListResult defines azure list subnet result.
type AzureSubnetListResult struct {
	Details []AzureSubnet `json:"details"`
}

// AzureSubnetExtension defines azure subnet extensional info.
type AzureSubnetExtension struct {
	ResourceGroupName    string  `json:"resource_group"`
	CloudRouteTableID    *string `json:"cloud_route_table_id,omitempty"`
	NatGateway           string  `json:"nat_gateway,omitempty"`
	NetworkSecurityGroup string  `json:"network_security_group,omitempty"`
}

// AzureSubnet defines azure subnet.
type AzureSubnet Subnet[AzureSubnetExtension]

// GetCloudID ...
func (vpc AzureSubnet) GetCloudID() string {
	return vpc.CloudID
}
