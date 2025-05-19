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

package vpc

import (
	"hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// VpcCreateReq defines create vpc request.
type VpcCreateReq[T VpcCreateExt] struct {
	*BaseVpcCreateReq `json:",inline" validate:"required"`
	Extension         *T `json:"extension" validate:"required"`
}

// BaseVpcCreateReq defines base create vpc request info.
type BaseVpcCreateReq struct {
	AccountID string             `json:"account_id" validate:"required"`
	Name      string             `json:"name" validate:"required"`
	Category  enumor.VpcCategory `json:"category" validate:"required"`
	Memo      *string            `json:"memo,omitempty" validate:"omitempty"`
	BkBizID   int64              `json:"bk_biz_id" validate:"omitempty"`
}

// Validate VpcCreateReq.
func (c VpcCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// VpcCreateExt defines vpc extensional info.
type VpcCreateExt interface {
	TCloudVpcCreateExt | AwsVpcCreateExt | GcpVpcCreateExt | AzureVpcCreateExt | HuaWeiVpcCreateExt
}

// TCloudVpcCreateExt defines tencent cloud vpc extensional info.
type TCloudVpcCreateExt struct {
	Region   string `json:"region" validate:"required"`
	IPv4Cidr string `json:"ipv4_cidr" validate:"required"`
	// Subnets defines subnets that should be created in vpc. **not supported in cloud, only for product demand**
	Subnets []subnet.TCloudOneSubnetCreateReq `json:"subnets" validate:"omitempty,max=100"`
}

// AwsVpcCreateExt defines aws vpc extensional info.
type AwsVpcCreateExt struct {
	Region                      string `json:"region" validate:"required"`
	IPv4Cidr                    string `json:"ipv4_cidr" validate:"required"`
	AmazonProvidedIpv6CidrBlock bool   `json:"amazon_provided_ipv6_cidr_block" validate:"-"`
	InstanceTenancy             string `json:"instance_tenancy" validate:"required"`
	// TODO dns选项，用ModifyVpcAttribute操作的
	// Subnets defines subnets that should be created in vpc. **not supported in cloud, only for product demand**
	Subnets []subnet.SubnetCreateReq[subnet.AwsSubnetCreateExt] `json:"subnets" validate:"omitempty,max=100"`
}

// GcpVpcCreateExt defines gcp vpc extensional info.
type GcpVpcCreateExt struct {
	AutoCreateSubnetworks bool   `json:"auto_create_subnetworks" validate:"-"`
	EnableUlaInternalIpv6 bool   `json:"enable_ula_internal_ipv6" validate:"-"`
	InternalIpv6Range     string `json:"internal_ipv6_range" validate:"-"`
	RoutingMode           string `json:"routing_mode,omitempty" validate:"omitempty"`
	// Subnets defines subnets that should be created in vpc. **not supported in cloud, only for product demand**
	Subnets []subnet.SubnetCreateReq[subnet.GcpSubnetCreateExt] `json:"subnets" validate:"omitempty,max=100"`
}

// AzureVpcCreateExt defines azure vpc extensional info.
type AzureVpcCreateExt struct {
	Region        string   `json:"region" validate:"required"`
	ResourceGroup string   `json:"resource_group" validate:"required"`
	IPv4Cidr      []string `json:"ipv4_cidr" validate:"omitempty"`
	IPv6Cidr      []string `json:"ipv6_cidr" validate:"omitempty"`
	// TODO BastionHost 等选项，暂时不支持启用，先不支持
	// Subnets defines subnets that should be created in vpc. **required**
	Subnets []subnet.SubnetCreateReq[subnet.AzureSubnetCreateExt] `json:"subnets" validate:"min=1,max=100"`
}

// HuaWeiVpcCreateExt defines huawei vpc extensional info.
type HuaWeiVpcCreateExt struct {
	Region              string  `json:"region" validate:"required"`
	IPv4Cidr            string  `json:"ipv4_cidr"`
	EnterpriseProjectID *string `json:"enterprise_project_id" validate:"omitempty"`
	// Subnets defines subnets that should be created in vpc. **not supported in cloud, only for product demand**
	Subnets []subnet.SubnetCreateReq[subnet.HuaWeiSubnetCreateExt] `json:"subnets" validate:"omitempty,max=100"`
}

// ------------------------- Update -------------------------

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate VpcUpdateReq.
func (u *VpcUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Sync --------------------------

// TCloudResourceSyncReq defines sync resource request.
type TCloudResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *TCloudResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiResourceSyncReq defines sync resource request.
type HuaWeiResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	VpcID     string   `json:"vpc_id" validate:"omitempty"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *HuaWeiResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpResourceSyncReq defines sync resource request.
type GcpResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *GcpResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureResourceSyncReq defines sync resource request.
type AzureResourceSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	CloudVpcID        string `json:"cloud_vpc_id" validate:"omitempty"`
	// CloudIDs 仅hcservice内部使用
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *AzureResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AwsResourceSyncReq defines sync resource request.
type AwsResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *AwsResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResourceSyncResult defines sync vpc result.
type ResourceSyncResult struct {
	TaskID string `json:"task_id"`
}
