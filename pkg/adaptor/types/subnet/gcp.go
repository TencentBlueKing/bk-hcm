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

// Package adtysubnet is adaptor types subnet
package adtysubnet

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// GcpSubnetCreateExt defines create gcp subnet extensional info.
type GcpSubnetCreateExt struct {
	Region                string `json:"region" validate:"required"`
	IPv4Cidr              string `json:"ipv4_cidr" validate:"required,cidrv4"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access" validate:"omitempty"`
	EnableFlowLogs        bool   `json:"enable_flow_logs" validate:"omitempty"`
}

// GcpSubnetCreateOption defines gcp create subnet options.
type GcpSubnetCreateOption SubnetCreateOption[GcpSubnetCreateExt]

// Validate GcpSubnetCreateOption.
func (c GcpSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// GcpSubnetUpdateOption defines gcp update subnet options.
type GcpSubnetUpdateOption struct {
	SubnetUpdateOption `json:",inline"`
	Region             string `json:"region"`
}

// Validate GcpSubnetUpdateOption.
func (s GcpSubnetUpdateOption) Validate() error {
	if len(s.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if s.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := s.Data.Validate(); err != nil {
		return err
	}

	return nil
}

// GcpSubnetListOption basic gcp list subnet options.
type GcpSubnetListOption struct {
	core.GcpListOption `json:",inline"`
	Region             string `json:"region"`
}

// Validate gcp list subnet option.
func (g GcpSubnetListOption) Validate() error {
	if err := g.GcpListOption.Validate(); err != nil {
		return err
	}

	if len(g.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region can be empty")
	}

	return nil
}

// GcpSubnetListResult defines gcp list subnet result.
type GcpSubnetListResult struct {
	NextPageToken string      `json:"next_page_token,omitempty"`
	Details       []GcpSubnet `json:"details"`
}

// GcpSubnetExtension defines gcp subnet extensional info.
type GcpSubnetExtension struct {
	SelfLink              string `json:"self_link"`
	Region                string `json:"region"`
	StackType             string `json:"stack_type"`
	Ipv6AccessType        string `json:"ipv6_access_type"`
	GatewayAddress        string `json:"gateway_address"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access"`
	EnableFlowLogs        bool   `json:"enable_flow_logs"`

	// 默认不返回
	AvailableIPAddressCount uint64 `json:"available_ip_address_count,omitempty"`
	TotalIpAddressCount     uint64 `json:"total_ip_address_count,omitempty"`
	UsedIpAddressCount      uint64 `json:"used_ip_address_count,omitempty"`
}

// GcpSubnet defines gcp subnet.
type GcpSubnet Subnet[GcpSubnetExtension]

// GetCloudID ...
func (vpc GcpSubnet) GetCloudID() string {
	return vpc.CloudID
}
