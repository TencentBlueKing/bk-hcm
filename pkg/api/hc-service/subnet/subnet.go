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

// Package subnet ...
package subnet

import (
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// SubnetCreateReq defines create subnet request.
type SubnetCreateReq[T SubnetCreateExt] struct {
	*BaseSubnetCreateReq `json:",inline" validate:"required"`
	Extension            *T `json:"extension" validate:"required"`
}

// BaseSubnetCreateReq defines base create subnet request info.
type BaseSubnetCreateReq struct {
	AccountID  string  `json:"account_id" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Memo       *string `json:"memo,omitempty" validate:"omitempty"`
	CloudVpcID string  `json:"cloud_vpc_id" validate:"required"`
	BkBizID    int64   `json:"bk_biz_id" validate:"required"`
}

// Validate SubnetCreateReq.
func (c SubnetCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// SubnetCreateExt defines create subnet extensional info.
type SubnetCreateExt interface {
	AwsSubnetCreateExt | GcpSubnetCreateExt | AzureSubnetCreateExt | HuaWeiSubnetCreateExt
}

// AwsSubnetCreateExt defines create aws subnet extensional info.
type AwsSubnetCreateExt struct {
	Region   string  `json:"region" validate:"required"`
	Zone     *string `json:"zone" validate:"omitempty"`
	IPv4Cidr *string `json:"ipv4_cidr" validate:"omitempty,cidrv4"`
	IPv6Cidr *string `json:"ipv6_cidr" validate:"omitempty,cidrv6"`
}

// GcpSubnetCreateExt defines create gcp subnet extensional info.
type GcpSubnetCreateExt struct {
	Region                string `json:"region" validate:"required"`
	IPv4Cidr              string `json:"ipv4_cidr" validate:"required,cidrv4"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access" validate:"omitempty"`
	EnableFlowLogs        bool   `json:"enable_flow_logs" validate:"omitempty"`
}

// AzureSubnetCreateExt defines create azure subnet extensional info.
type AzureSubnetCreateExt struct {
	ResourceGroup        string   `json:"resource_group" validate:"required"`
	IPv4Cidr             []string `json:"ipv4_cidr" validate:"required,dive,cidrv4"`
	IPv6Cidr             []string `json:"ipv6_cidr" validate:"omitempty,dive,cidrv6"`
	CloudRouteTableID    string   `json:"cloud_route_table_id,omitempty" validate:"omitempty"`
	NatGateway           string   `json:"nat_gateway,omitempty" validate:"omitempty"`
	NetworkSecurityGroup string   `json:"network_security_group,omitempty" validate:"omitempty"`
}

// HuaWeiSubnetCreateExt defines create huawei subnet extensional info.
type HuaWeiSubnetCreateExt struct {
	Region     string  `json:"region" validate:"required"`
	Zone       *string `json:"zone" validate:"omitempty"`
	IPv4Cidr   string  `json:"ipv4_cidr" validate:"required,cidrv4"`
	Ipv6Enable bool    `json:"ipv6_enable" validate:"omitempty"`
	GatewayIp  string  `json:"gateway_ip" validate:"required"`
}

// TCloudSubnetBatchCreateReq defines batch create tencent cloud subnets request.
type TCloudSubnetBatchCreateReq struct {
	BkBizID    int64                      `json:"bk_biz_id" validate:"required"`
	AccountID  string                     `json:"account_id" validate:"required"`
	Region     string                     `json:"region" validate:"required"`
	CloudVpcID string                     `json:"cloud_vpc_id" validate:"required"`
	Subnets    []TCloudOneSubnetCreateReq `json:"subnets" validate:"min=1"`
}

// TCloudOneSubnetCreateReq defines create one tencent cloud subnets request for TCloudSubnetBatchCreateReq.
type TCloudOneSubnetCreateReq struct {
	IPv4Cidr          string  `json:"ipv4_cidr" validate:"required,cidrv4"`
	Name              string  `json:"name" validate:"required"`
	Zone              string  `json:"zone" validate:"required"`
	CloudRouteTableID string  `json:"cloud_route_table_id" validate:"omitempty"`
	Memo              *string `json:"memo"`
}

// Validate TCloudSubnetBatchCreateReq.
func (c TCloudSubnetBatchCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// ------------------------- Update -------------------------

// SubnetUpdateReq defines update subnet request.
type SubnetUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate SubnetUpdateReq.
func (u *SubnetUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}
