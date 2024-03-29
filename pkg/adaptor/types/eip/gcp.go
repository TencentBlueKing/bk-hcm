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

package eip

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	"google.golang.org/api/compute/v1"
)

const (
	// GcpGlobalRegion ...
	GcpGlobalRegion = "global"
	// DefaultExternalNatName ...
	DefaultExternalNatName = "external-nat"
)

// GcpEipListOption ...
type GcpEipListOption struct {
	Region    string        `json:"region" validate:"required"`
	CloudIDs  []string      `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string      `json:"self_links" validate:"omitempty"`
	Page      *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate ...
func (o *GcpEipListOption) Validate() error {
	return validator.Validate.Struct(o)
}

// GcpEipAggregatedListOption ...
type GcpEipAggregatedListOption struct {
	IPAddresses []string `json:"ip_addresses" validate:"required"`
}

// Validate ...
func (o *GcpEipAggregatedListOption) Validate() error {
	return validator.Validate.Struct(o)
}

// GcpEipListResult ...
type GcpEipListResult struct {
	NextPageToken string
	Details       []*GcpEip
}

// GcpEip ...
type GcpEip struct {
	CloudID      string
	Name         *string
	Region       string
	Status       *string
	PublicIp     *string
	PrivateIp    *string
	AddressType  string
	Description  string
	IpVersion    string
	NetworkTier  string
	PrefixLength int64
	Purpose      string
	Network      string
	Subnetwork   string
	SelfLink     string
	Users        []string
}

// GetCloudID ...
func (eip *GcpEip) GetCloudID() string {
	return eip.CloudID
}

// GcpEipDeleteOption ...
type GcpEipDeleteOption struct {
	Region  string `json:"region" validate:"required"`
	EipName string `json:"eip_name" validate:"required"`
}

// Validate ...
func (opt *GcpEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpEipAssociateOption ...
type GcpEipAssociateOption struct {
	Zone                 string `json:"zone" validate:"required"`
	CvmName              string `json:"cvm_name" validate:"required"`
	NetworkInterfaceName string `json:"network_interface_name" validate:"required"`
	PublicIp             string `json:"public_ip" validate:"required"`
	CloudID              string `json:"cloud_id" validate:"omitempty"`
	Region               string `json:"region" validate:"omitempty"`
}

// Validate ...
func (opt *GcpEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpEipDisassociateOption ...
type GcpEipDisassociateOption struct {
	Zone                 string `json:"zone" validate:"required"`
	CvmName              string `json:"cvm_name" validate:"required"`
	NetworkInterfaceName string `json:"network_interface_name" validate:"required"`
	AccessConfigName     string `json:"access_config_name" validate:"required"`
	CloudID              string `json:"cloud_id" validate:"omitempty"`
	Region               string `json:"region" validate:"omitempty"`
}

// Validate ...
func (opt *GcpEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpEipCreateOption ...
type GcpEipCreateOption struct {
	EipName     string `json:"eip_name" validate:"required"`
	Region      string `json:"region" validate:"required"`
	NetworkTier string `json:"network_tier" validate:"required,eq=PREMIUM|eq=STANDARD"`
	IpVersion   string `json:"ip_version" validate:"required"`
}

// Validate ...
func (opt *GcpEipCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAddress ...
func (opt *GcpEipCreateOption) ToAddress() (*compute.Address, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &compute.Address{NetworkTier: opt.NetworkTier}
	if opt.Region == GcpGlobalRegion {
		req.IpVersion = opt.IpVersion
	}
	req.Name = opt.EipName

	return req, nil
}
