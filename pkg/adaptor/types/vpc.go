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

package types

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// VpcCreateOption defines create vpc options.
type VpcCreateOption[T VpcCreateExt] struct {
	AccountID string  `json:"account_id" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo,omitempty" validate:"omitempty"`
	Extension *T      `json:"extension" validate:"required"`
}

// VpcCreateExt defines vpc extensional info.
type VpcCreateExt interface {
	TCloudVpcCreateExt | AwsVpcCreateExt | GcpVpcCreateExt | AzureVpcCreateExt | HuaWeiVpcCreateExt
}

// TCloudVpcCreateExt defines tencent cloud vpc extensional info.
type TCloudVpcCreateExt struct {
	Region   string `json:"region" validate:"required"`
	IPv4Cidr string `json:"ipv4_cidr" validate:"required,cidrv4"`
}

// AwsVpcCreateExt defines aws vpc extensional info.
type AwsVpcCreateExt struct {
	Region                      string `json:"region" validate:"required"`
	IPv4Cidr                    string `json:"ipv4_cidr" validate:"required,cidrv4"`
	AmazonProvidedIpv6CidrBlock bool   `json:"amazon_provided_ipv6_cidr_block" validate:"-"`
	InstanceTenancy             string `json:"instance_tenancy" validate:"required"`
}

// GcpVpcCreateExt defines gcp vpc extensional info.
type GcpVpcCreateExt struct {
	AutoCreateSubnetworks bool   `json:"auto_create_subnetworks" validate:"-"`
	EnableUlaInternalIpv6 bool   `json:"enable_ula_internal_ipv6" validate:"-"`
	InternalIpv6Range     string `json:"internal_ipv6_range" validate:"-"`
	RoutingMode           string `json:"routing_mode,omitempty" validate:"omitempty"`
}

// AzureVpcCreateExt defines azure vpc extensional info.
type AzureVpcCreateExt struct {
	Region        string                               `json:"region" validate:"required"`
	ResourceGroup string                               `json:"resource_group" validate:"required"`
	IPv4Cidr      []string                             `json:"ipv4_cidr" validate:"required,dive,cidrv4"`
	IPv6Cidr      []string                             `json:"ipv6_cidr" validate:"omitempty,dive,cidrv6"`
	Subnets       []adtysubnet.AzureSubnetCreateOption `json:"subnets" validate:"min=1,max=100"`
}

// HuaWeiVpcCreateExt defines huawei vpc extensional info.
type HuaWeiVpcCreateExt struct {
	Region              string  `json:"region" validate:"required"`
	IPv4Cidr            string  `json:"ipv4_cidr" validate:"required,cidrv4"`
	EnterpriseProjectID *string `json:"enterprise_project_id" validate:"omitempty"`
}

// TCloudVpcCreateOption defines tencent cloud create vpc options.
type TCloudVpcCreateOption VpcCreateOption[TCloudVpcCreateExt]

// Validate TCloudVpcCreateOption.
func (c TCloudVpcCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// AwsVpcCreateOption defines aws create vpc options.
type AwsVpcCreateOption VpcCreateOption[AwsVpcCreateExt]

// Validate AwsVpcCreateOption.
func (c AwsVpcCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// GcpVpcCreateOption defines gcp create vpc options.
type GcpVpcCreateOption VpcCreateOption[GcpVpcCreateExt]

// Validate GcpVpcCreateOption.
func (c GcpVpcCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureVpcCreateOption defines azure create vpc options.
type AzureVpcCreateOption VpcCreateOption[AzureVpcCreateExt]

// Validate AzureVpcCreateOption.
func (c AzureVpcCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// HuaWeiVpcCreateOption defines HuaWei create vpc options.
type HuaWeiVpcCreateOption VpcCreateOption[HuaWeiVpcCreateExt]

// Validate HuaWeiVpcCreateOption.
func (c HuaWeiVpcCreateOption) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// VpcUpdateOption defines update vpc options.
type VpcUpdateOption struct {
	ResourceID string             `json:"resource_id"`
	Data       *BaseVpcUpdateData `json:"data"`
}

// BaseVpcUpdateData defines the basic update vpc instance data.
type BaseVpcUpdateData struct {
	Memo *string `json:"memo"`
}

// Validate BaseVpcUpdateData.
func (v BaseVpcUpdateData) Validate() error {
	if v.Memo == nil {
		return errf.New(errf.InvalidParameter, "memo is required")
	}
	return nil
}

// TCloudVpcUpdateOption defines tencent cloud update vpc options.
type TCloudVpcUpdateOption struct{}

// Validate TCloudVpcUpdateOption.
func (v TCloudVpcUpdateOption) Validate() error {
	return nil
}

// AwsVpcUpdateOption defines aws update vpc options.
type AwsVpcUpdateOption struct{}

// Validate AwsVpcUpdateOption.
func (v AwsVpcUpdateOption) Validate() error {
	return nil
}

// GcpVpcUpdateOption defines gcp update vpc options.
type GcpVpcUpdateOption VpcUpdateOption

// Validate GcpVpcUpdateOption.
func (v GcpVpcUpdateOption) Validate() error {
	if len(v.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if v.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := v.Data.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureVpcUpdateOption defines azure update vpc options.
type AzureVpcUpdateOption struct{}

// Validate AzureVpcUpdateOption.
func (v AzureVpcUpdateOption) Validate() error {
	return nil
}

// HuaWeiVpcUpdateOption defines huawei update vpc options.
type HuaWeiVpcUpdateOption struct {
	VpcUpdateOption `json:",inline"`
	Region          string `json:"region"`
}

// Validate HuaWeiVpcUpdateOption.
func (v HuaWeiVpcUpdateOption) Validate() error {
	if len(v.Region) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if len(v.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if v.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := v.Data.Validate(); err != nil {
		return err
	}
	return nil
}

// -------------------------- List --------------------------

// GcpListOption defines options to list gcp vpc.
type GcpListOption struct {
	CloudIDs  []string      `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string      `json:"self_links" validate:"omitempty"`
	Page      *core.GcpPage `json:"page" validate:"required"`
}

// Validate gcp cvm list option.
func (opt GcpListOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudVpcListResult defines tencent cloud list vpc result.
type TCloudVpcListResult struct {
	Count   *uint64     `json:"count,omitempty"`
	Details []TCloudVpc `json:"details"`
}

// AwsVpcListResult defines aws list vpc result.
type AwsVpcListResult struct {
	NextToken *string  `json:"next_token,omitempty"`
	Details   []AwsVpc `json:"details"`
}

// GcpVpcListResult defines gcp list vpc result.
type GcpVpcListResult struct {
	NextPageToken string   `json:"next_page_token,omitempty"`
	Details       []GcpVpc `json:"details"`
}

// AzureVpcListResult defines azure list vpc result.
type AzureVpcListResult struct {
	Details []AzureVpc `json:"details"`
}

// HuaWeiVpcListOption defines huawei list vpc options.
type HuaWeiVpcListOption struct {
	core.HuaWeiListOption `json:",inline"`
	Names                 []string `json:"names,omitempty"`
}

// Validate huawei list option.
func (v HuaWeiVpcListOption) Validate() error {
	if err := v.HuaWeiListOption.Validate(); err != nil {
		return err
	}

	return nil
}

// HuaWeiVpcListResult defines huawei list vpc result.
type HuaWeiVpcListResult struct {
	NextMarker *string     `json:"next_marker,omitempty"`
	Details    []HuaWeiVpc `json:"details"`
}

// Vpc defines vpc struct.
type Vpc[T VpcExtension] struct {
	CloudID   string  `json:"cloud_id"`
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	Memo      *string `json:"memo,omitempty"`
	Extension *T      `json:"extension"`
}

// AzureVpcExtension defines azure vpc extensional info.
type AzureVpcExtension struct {
	ResourceGroupName string            `json:"resource_group_name"`
	DNSServers        []string          `json:"dns_servers"`
	Cidr              []cloud.AzureCidr `json:"cidr"`
	Subnets           []adtysubnet.AzureSubnet
}

// VpcExtension defines vpc extensional info.
type VpcExtension interface {
	cloud.TCloudVpcExtension | cloud.AwsVpcExtension | cloud.GcpVpcExtension | AzureVpcExtension |
		cloud.HuaWeiVpcExtension
}

// TCloudVpc defines tencent cloud vpc.
type TCloudVpc Vpc[cloud.TCloudVpcExtension]

// GetCloudID ...
func (vpc TCloudVpc) GetCloudID() string {
	return vpc.CloudID
}

// AwsVpc defines aws vpc.
type AwsVpc Vpc[cloud.AwsVpcExtension]

// GetCloudID ...
func (vpc AwsVpc) GetCloudID() string {
	return vpc.CloudID
}

// GcpVpc defines gcp vpc.
type GcpVpc Vpc[cloud.GcpVpcExtension]

// GetCloudID ...
func (vpc GcpVpc) GetCloudID() string {
	return vpc.CloudID
}

// AzureVpc defines azure vpc.
type AzureVpc Vpc[AzureVpcExtension]

// GetCloudID ...
func (vpc AzureVpc) GetCloudID() string {
	return vpc.CloudID
}

// HuaWeiVpc defines huawei vpc.
type HuaWeiVpc Vpc[cloud.HuaWeiVpcExtension]

// GetCloudID ...
func (vpc HuaWeiVpc) GetCloudID() string {
	return vpc.CloudID
}

// VpcUsage define vpc usage.
type VpcUsage struct {
	ID           *string  `json:"id"`
	Limit        *float64 `json:"limit"`
	CurrentValue *float64 `json:"current_value"`
}

// -------------------------- IP --------------------------

// HuaWeiVpcIPAvailGetOption get huawei vcp ip availabilities option.
type HuaWeiVpcIPAvailGetOption struct {
	Region   string `json:"region"`
	SubnetID string `json:"subnet_id"`
}

// Validate HuaWeiVpcIPAvailGetOption.
func (v HuaWeiVpcIPAvailGetOption) Validate() error {
	if len(v.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(v.SubnetID) == 0 {
		return errf.New(errf.InvalidParameter, "subnetID id is required")
	}

	return nil
}

// AzureVpcListUsageOption defines azure list vpc usage options
type AzureVpcListUsageOption struct {
	ResourceGroupName string `json:"resource_group_name"`
	VpcID             string `json:"vpc_name"`
}

// Validate AzureVpcIPAvailGetOption.
func (v AzureVpcListUsageOption) Validate() error {
	if len(v.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group is required")
	}

	if len(v.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}

	return nil
}

// HuaweiListPortOption defines huawei list port options.
type HuaweiListPortOption struct {
	Region           string   `json:"region" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required"`
	Marker           string   `json:"marker" validate:"omitempty"`
}

// Validate ...
func (v HuaweiListPortOption) Validate() error {
	return validator.Validate.Struct(v)
}
