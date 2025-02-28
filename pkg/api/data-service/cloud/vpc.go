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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// VpcBatchCreateReq defines batch create vpc request.
type VpcBatchCreateReq[T VpcCreateExtension] struct {
	Vpcs []VpcCreateReq[T] `json:"vpcs" validate:"min=1,max=100"`
}

// VpcCreateReq defines create vpc request.
type VpcCreateReq[T VpcCreateExtension] struct {
	AccountID string             `json:"account_id" validate:"required"`
	CloudID   string             `json:"cloud_id" validate:"required"`
	BkBizID   int64              `json:"bk_biz_id" validate:"required"`
	Name      *string            `json:"name,omitempty" validate:"omitempty"`
	Region    string             `json:"region" validate:"omitempty"`
	Category  enumor.VpcCategory `json:"category" validate:"required"`
	Memo      *string            `json:"memo,omitempty" validate:"omitempty"`
	Extension *T                 `json:"extension" validate:"required"`
}

// VpcCreateExtension defines create vpc extensional info.
type VpcCreateExtension interface {
	TCloudVpcCreateExt | AwsVpcCreateExt | GcpVpcCreateExt | AzureVpcCreateExt | HuaWeiVpcCreateExt
}

// TCloudVpcCreateExt defines create tencent cloud vpc extensional info.
type TCloudVpcCreateExt struct {
	Cidr            []TCloudCidr `json:"cidr" validate:"required"`
	IsDefault       bool         `json:"is_default" validate:"omitempty"`
	EnableMulticast bool         `json:"enable_multicast" validate:"omitempty"`
	DnsServerSet    []string     `json:"dns_server_set" validate:"omitempty"`
	DomainName      string       `json:"domain_name,omitempty" validate:"omitempty"`
}

// AwsVpcCreateExt defines create aws vpc extensional info.
type AwsVpcCreateExt struct {
	Cidr               []AwsCidr `json:"cidr" validate:"required"`
	State              string    `json:"state" validate:"required"`
	InstanceTenancy    string    `json:"instance_tenancy" validate:"omitempty"`
	IsDefault          bool      `json:"is_default" validate:"omitempty"`
	EnableDnsHostnames bool      `json:"enable_dns_hostnames" validate:"omitempty"`
	EnableDnsSupport   bool      `json:"enable_dns_support" validate:"omitempty"`
}

// GcpVpcCreateExt defines gcp vpc extensional info.
type GcpVpcCreateExt struct {
	SelfLink              string `json:"self_link" validate:"required"`
	AutoCreateSubnetworks bool   `json:"auto_create_subnetworks" validate:"omitempty"`
	EnableUlaInternalIpv6 bool   `json:"enable_ula_internal_ipv6" validate:"omitempty"`
	InternalIpv6Range     string `json:"internal_ipv6_range" validate:"omitempty"`
	Mtu                   int64  `json:"mtu" validate:"required"`
	RoutingMode           string `json:"routing_mode" validate:"omitempty"`
}

// AzureVpcCreateExt defines azure vpc extensional info.
type AzureVpcCreateExt struct {
	ResourceGroupName string      `json:"resource_group_name" validate:"required"`
	DNSServers        []string    `json:"dns_servers" validate:"omitempty"`
	Cidr              []AzureCidr `json:"cidr" validate:"required,min=1"`
}

// HuaWeiVpcCreateExt defines huawei vpc extensional info.
type HuaWeiVpcCreateExt struct {
	Cidr                []HuaWeiCidr `json:"cidr" validate:"required,min=1"`
	Status              string       `json:"status" validate:"required"`
	EnterpriseProjectID string       `json:"enterprise_project_id" validate:"omitempty"`
}

// Validate VpcBatchCreateReq.
func (c *VpcBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// VpcBatchUpdateReq defines batch update vpc request.
type VpcBatchUpdateReq[T VpcUpdateExtension] struct {
	Vpcs []VpcUpdateReq[T] `json:"vpcs" validate:"min=1,max=100"`
}

// Validate VpcBatchUpdateReq.
func (u *VpcBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq[T VpcUpdateExtension] struct {
	ID                string `json:"id" validate:"required"`
	VpcUpdateBaseInfo `json:",inline" validate:"omitempty"`
	Extension         *T `json:"extension" validate:"omitempty"`
}

// VpcUpdateBaseInfo defines update vpc request base info.
type VpcUpdateBaseInfo struct {
	Name     *string            `json:"name,omitempty" validate:"omitempty"`
	Category enumor.VpcCategory `json:"category,omitempty" validate:"omitempty"`
	Memo     *string            `json:"memo,omitempty" validate:"omitempty"`
	BkBizID  int64              `json:"bk_biz_id,omitempty" validate:"omitempty"`
}

// VpcUpdateExtension defines vpc update request extensional info.
type VpcUpdateExtension interface {
	TCloudVpcUpdateExt | AwsVpcUpdateExt | GcpVpcUpdateExt | AzureVpcUpdateExt | HuaWeiVpcUpdateExt
}

// TCloudVpcUpdateExt defines tencent cloud vpc extensional info.
type TCloudVpcUpdateExt struct {
	Cidr            []TCloudCidr `json:"cidr,omitempty" validate:"omitempty"`
	IsDefault       *bool        `json:"is_default,omitempty" validate:"omitempty"`
	EnableMulticast *bool        `json:"enable_multicast,omitempty" validate:"omitempty"`
	DnsServerSet    []string     `json:"dns_server_set" validate:"omitempty"`
	DomainName      *string      `json:"domain_name,omitempty" validate:"omitempty"`
}

// AwsVpcUpdateExt defines aws vpc extensional info.
type AwsVpcUpdateExt struct {
	Cidr               []AwsCidr `json:"cidr,omitempty" validate:"omitempty"`
	State              string    `json:"state,omitempty" validate:"omitempty"`
	InstanceTenancy    *string   `json:"instance_tenancy,omitempty" validate:"omitempty"`
	IsDefault          *bool     `json:"is_default,omitempty" validate:"omitempty"`
	EnableDnsHostnames *bool     `json:"enable_dns_hostnames,omitempty" validate:"omitempty"`
	EnableDnsSupport   *bool     `json:"enable_dns_support,omitempty" validate:"omitempty"`
}

// GcpVpcUpdateExt defines gcp vpc extensional info.
type GcpVpcUpdateExt struct {
	EnableUlaInternalIpv6 *bool   `json:"enable_ula_internal_ipv6,omitempty" validate:"omitempty"`
	InternalIpv6Range     *string `json:"internal_ipv6_range,omitempty" validate:"omitempty"`
	Mtu                   int64   `json:"mtu,omitempty" validate:"omitempty"`
	RoutingMode           *string `json:"routing_mode,omitempty" validate:"omitempty"`
}

// AzureVpcUpdateExt defines azure vpc extensional info.
type AzureVpcUpdateExt struct {
	DNSServers []string    `json:"dns_servers" validate:"omitempty"`
	Cidr       []AzureCidr `json:"cidr,omitempty" validate:"omitempty"`
}

// HuaWeiVpcUpdateExt defines huawei vpc extensional info.
type HuaWeiVpcUpdateExt struct {
	Cidr                []HuaWeiCidr `json:"cidr,omitempty" validate:"omitempty"`
	Status              string       `json:"status,omitempty" validate:"omitempty"`
	EnterpriseProjectId *string      `json:"enterprise_project_id,omitempty" validate:"omitempty"`
}

// VpcBaseInfoBatchUpdateReq defines batch update vpc base info request.
type VpcBaseInfoBatchUpdateReq struct {
	Vpcs []VpcBaseInfoUpdateReq `json:"vpcs" validate:"required"`
}

// Validate VpcBaseInfoBatchUpdateReq.
func (u *VpcBaseInfoBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// VpcBaseInfoUpdateReq defines update vpc base info request.
type VpcBaseInfoUpdateReq struct {
	IDs  []string           `json:"id" validate:"required"`
	Data *VpcUpdateBaseInfo `json:"data" validate:"required"`
}

// -------------------------- Get --------------------------

// VpcGetResp defines get vpc response.
type VpcGetResp[T cloud.VpcExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *cloud.Vpc[T] `json:"data"`
}

// -------------------------- List --------------------------

// VpcListResp defines list vpc response.
type VpcListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *VpcListResult `json:"data"`
}

// VpcListResult defines list vpc result.
type VpcListResult struct {
	Count   uint64          `json:"count"`
	Details []cloud.BaseVpc `json:"details"`
}

// VpcExtListResult define vpc with extension list result.
type VpcExtListResult[T cloud.VpcExtension] struct {
	Count   uint64         `json:"count,omitempty"`
	Details []cloud.Vpc[T] `json:"details,omitempty"`
}

// VpcExtListResp define list resp.
type VpcExtListResp[T cloud.VpcExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *VpcExtListResult[T] `json:"data"`
}

// -------------------------- Cidr --------------------------

// TCloudCidr tencent cloud cidr
type TCloudCidr struct {
	Type     enumor.IPAddressType      `json:"type" validate:"required"`
	Cidr     string                    `json:"cidr" validate:"required"`
	Category enumor.TCloudCidrCategory `json:"category" validate:"required"`
}

// AwsCidr aws cidr
type AwsCidr struct {
	Type        enumor.IPAddressType `json:"type" validate:"required"`
	Cidr        string               `json:"cidr" validate:"required"`
	AddressPool string               `json:"address_pool" validate:"omitempty"`
	State       string               `json:"state" validate:"omitempty"`
}

// AzureCidr azure cidr
type AzureCidr struct {
	Type enumor.IPAddressType `json:"type" validate:"required"`
	Cidr string               `json:"cidr" validate:"required"`
}

// HuaWeiCidr huawei cidr
type HuaWeiCidr struct {
	Type enumor.IPAddressType `json:"type" validate:"required"`
	Cidr string               `json:"cidr" validate:"required"`
}
