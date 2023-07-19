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
	"fmt"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// HuaWeiSubnetCreateExt defines create huawei subnet extensional info.
type HuaWeiSubnetCreateExt struct {
	Region     string  `json:"region" validate:"required"`
	Zone       *string `json:"zone" validate:"omitempty"`
	IPv4Cidr   string  `json:"ipv4_cidr" validate:"required,cidrv4"`
	Ipv6Enable bool    `json:"ipv6_enable" validate:"omitempty"`
	GatewayIp  string  `json:"gateway_ip" validate:"required"`
}

// HuaWeiSubnetCreateOption defines HuaWei create subnet options.
type HuaWeiSubnetCreateOption SubnetCreateOption[HuaWeiSubnetCreateExt]

// Validate HuaWeiSubnetCreateOption.
func (c HuaWeiSubnetCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// HuaWeiSubnetUpdateOption defines huawei update subnet options.
type HuaWeiSubnetUpdateOption struct {
	SubnetUpdateOption `json:",inline"`
	Region             string `json:"region"`
	Name               string `json:"name"`
	VpcID              string `json:"vpc_id"`
}

// Validate HuaWeiSubnetUpdateOption.
func (s HuaWeiSubnetUpdateOption) Validate() error {
	if err := s.Data.Validate(); err != nil {
		return err
	}

	if len(s.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(s.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}
	return nil
}

// HuaWeiSubnetDeleteOption defines huawei delete subnet options.
type HuaWeiSubnetDeleteOption struct {
	core.BaseRegionalDeleteOption `json:",inline"`
	VpcID                         string `json:"vpc_id"`
}

// Validate HuaWeiSubnetDeleteOption.
func (s HuaWeiSubnetDeleteOption) Validate() error {
	if err := s.BaseRegionalDeleteOption.Validate(); err != nil {
		return err
	}

	if len(s.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}
	return nil
}

// HuaWeiSubnetListOption defines huawei list subnet options.
type HuaWeiSubnetListOption struct {
	Region     string           `json:"region"`
	Page       *core.HuaWeiPage `json:"page,omitempty"`
	CloudVpcID string           `json:"vpc_id,omitempty"`
}

// Validate huawei list option.
func (s HuaWeiSubnetListOption) Validate() error {
	if len(s.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if s.Page != nil {
		if err := s.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiSubnetListByIDOption ...
type HuaWeiSubnetListByIDOption struct {
	Region     string   `json:"region" validate:"required"`
	CloudVpcID string   `json:"cloud_vpc_id" validate:"required"`
	CloudIDs   []string `json:"cloud_ids" validate:"required"`
}

// Validate HuaWeiSubnetListByIDOption.
func (opt HuaWeiSubnetListByIDOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// HuaWeiSubnetListResult defines huawei list subnet result.
type HuaWeiSubnetListResult struct {
	Details []HuaWeiSubnet `json:"details"`
}

// HuaWeiSubnetExtension defines huawei subnet extensional info.
type HuaWeiSubnetExtension struct {
	Region       string   `json:"region"`
	Status       string   `json:"status"`
	DhcpEnable   bool     `json:"dhcp_enable"`
	GatewayIp    string   `json:"gateway_ip"`
	DnsList      []string `json:"dns_list"`
	NtpAddresses []string `json:"ntp_addresses"`
}

// HuaWeiSubnet defines huawei subnet.
type HuaWeiSubnet Subnet[HuaWeiSubnetExtension]

// GetCloudID ...
func (vpc HuaWeiSubnet) GetCloudID() string {
	return vpc.CloudID
}
