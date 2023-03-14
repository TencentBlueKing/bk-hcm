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

package cloudserver

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// BaseSubnetCreateReq defines base create subnet request info.
type BaseSubnetCreateReq struct {
	Vendor     enumor.Vendor `json:"vendor" validate:"required"`
	AccountID  string        `json:"account_id" validate:"required"`
	CloudVpcID string        `json:"cloud_vpc_id" validate:"required"`
	Name       string        `json:"name" validate:"required"`
	Memo       *string       `json:"memo,omitempty" validate:"omitempty"`
}

// TCloudSubnetCreateReq defines tencent cloud create subnet request.
type TCloudSubnetCreateReq struct {
	*BaseSubnetCreateReq `json:",inline"  validate:"required"`
	Region               string `json:"region" validate:"required"`
	Zone                 string `json:"zone" validate:"required"`
	IPv4Cidr             string `json:"ipv4_cidr" validate:"required,cidrv4"`
	CloudRouteTableID    string `json:"cloud_route_table_id" validate:"omitempty"`
}

// Validate TCloudSubnetCreateReq.
func (c TCloudSubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// AwsSubnetCreateReq defines aws create subnet request.
type AwsSubnetCreateReq struct {
	*BaseSubnetCreateReq `json:",inline"  validate:"required"`
	Region               string  `json:"region" validate:"required"`
	Zone                 *string `json:"zone" validate:"omitempty"`
	IPv4Cidr             *string `json:"ipv4_cidr" validate:"omitempty,cidrv4"`
	IPv6Cidr             *string `json:"ipv6_cidr" validate:"omitempty,cidrv6"`
}

// Validate AwsSubnetCreateReq.
func (c AwsSubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// GcpSubnetCreateReq defines gcp create subnet request.
type GcpSubnetCreateReq struct {
	*BaseSubnetCreateReq  `json:",inline"  validate:"required"`
	Region                string `json:"region" validate:"required"`
	IPv4Cidr              string `json:"ipv4_cidr" validate:"required,cidrv4"`
	PrivateIpGoogleAccess bool   `json:"private_ip_google_access" validate:"omitempty"`
	EnableFlowLogs        bool   `json:"enable_flow_logs" validate:"omitempty"`
}

// Validate GcpSubnetCreateReq.
func (c GcpSubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureSubnetCreateReq defines azure create subnet request.
type AzureSubnetCreateReq struct {
	*BaseSubnetCreateReq `json:",inline"  validate:"required"`
	ResourceGroup        string   `json:"resource_group" validate:"required"`
	IPv4Cidr             []string `json:"ipv4_cidr" validate:"required,dive,cidrv4"`
	IPv6Cidr             []string `json:"ipv6_cidr" validate:"omitempty,dive,cidrv6"`
	CloudRouteTableID    string   `json:"cloud_route_table_id,omitempty" validate:"omitempty"`
	NatGateway           string   `json:"nat_gateway,omitempty" validate:"omitempty"`
	NetworkSecurityGroup string   `json:"network_security_group,omitempty" validate:"omitempty"`
}

// Validate AzureSubnetCreateReq.
func (c AzureSubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// HuaWeiSubnetCreateReq defines huawei create subnet request.
type HuaWeiSubnetCreateReq struct {
	*BaseSubnetCreateReq `json:",inline"  validate:"required"`
	Region               string  `json:"region" validate:"required"`
	Zone                 *string `json:"zone" validate:"omitempty"`
	IPv4Cidr             string  `json:"ipv4_cidr" validate:"required,cidrv4"`
	Ipv6Enable           bool    `json:"ipv6_enable" validate:"omitempty"`
	GatewayIp            string  `json:"gateway_ip" validate:"required"`
}

// Validate HuaWeiSubnetCreateReq.
func (c HuaWeiSubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// SubnetUpdateReq defines update subnet request.
type SubnetUpdateReq struct {
	Memo *string `json:"memo" validate:"required"`
}

// Validate SubnetUpdateReq.
func (u SubnetUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- List --------------------------

// SubnetListResult defines list subnet result.
type SubnetListResult struct {
	Count   uint64             `json:"count"`
	Details []cloud.BaseSubnet `json:"details"`
}

// -------------------------- Relation ------------------------

// AssignSubnetToBizReq assign subnets to biz request.
type AssignSubnetToBizReq struct {
	SubnetIDs []string `json:"subnet_ids"`
	BkBizID   int64    `json:"bk_biz_id"`
}

// Validate AssignSubnetToBizReq.
func (a AssignSubnetToBizReq) Validate() error {
	if len(a.SubnetIDs) == 0 {
		return errf.New(errf.InvalidParameter, "subnet ids are required")
	}

	if a.BkBizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is required")
	}

	return nil
}
