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
type VpcBatchCreateReq[T cloud.VpcExtension] struct {
	Vpcs []VpcCreateReq[T] `json:"vpcs" validate:"required"`
}

// VpcCreateReq defines create vpc request.
type VpcCreateReq[T cloud.VpcExtension] struct {
	Spec      *VpcCreateSpec `json:"spec" validate:"required"`
	Extension *T             `json:"extension" validate:"required"`
}

// VpcCreateSpec defines create vpc request spec.
type VpcCreateSpec struct {
	AccountID string             `json:"account_id" validate:"required"`
	CloudID   string             `json:"cloud_id" validate:"required"`
	Name      *string            `json:"name" validate:"omitempty"`
	Category  enumor.VpcCategory `json:"category" validate:"required"`
	Memo      *string            `json:"memo,omitempty" validate:"omitempty"`
}

// Validate VpcBatchCreateReq.
func (c *VpcBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// VpcBatchUpdateReq defines batch update vpc request.
type VpcBatchUpdateReq[T VpcUpdateExtension] struct {
	Vpcs []VpcUpdateReq[T] `json:"vpcs" validate:"required"`
}

// Validate VpcBatchUpdateReq.
func (u *VpcBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq[T VpcUpdateExtension] struct {
	ID        string         `json:"id" validate:"required"`
	Spec      *VpcUpdateSpec `json:"spec" validate:"required_without=extension"`
	Extension *T             `json:"extension" validate:"required_without=spec"`
}

// VpcUpdateSpec defines update vpc request spec.
type VpcUpdateSpec struct {
	Name     *string            `json:"name" validate:"omitempty"`
	Category enumor.VpcCategory `json:"category" validate:"omitempty"`
	Memo     *string            `json:"memo" validate:"omitempty"`
}

// VpcUpdateExtension defines vpc update request extensional info.
type VpcUpdateExtension interface {
	TCloudVpcUpdateExt | AwsVpcUpdateExt | GcpVpcUpdateExt | AzureVpcUpdateExt | HuaWeiVpcUpdateExt
}

// TCloudVpcUpdateExt defines tencent cloud vpc extensional info.
type TCloudVpcUpdateExt struct {
	Region          string       `json:"region"`
	Cidr            []TCloudCidr `json:"cidr"`
	IsDefault       bool         `json:"is_default"`
	EnableMulticast bool         `json:"enable_multicast"`
	DnsServerSet    []string     `json:"dns_server_set"`
	DomainName      *string      `json:"domain_name,omitempty"`
}

// TCloudCidr tencent cloud cidr
type TCloudCidr struct {
	Type     enumor.IPAddressType      `json:"type"`
	Cidr     string                    `json:"cidr"`
	Category enumor.TCloudCidrCategory `json:"category"`
}

// AwsVpcUpdateExt defines aws vpc extensional info.
type AwsVpcUpdateExt struct {
	Region             string    `json:"region"`
	Cidr               []AwsCidr `json:"cidr"`
	State              string    `json:"state"`
	InstanceTenancy    string    `json:"instance_tenancy"`
	IsDefault          bool      `json:"is_default"`
	EnableDnsHostnames bool      `json:"enable_dns_hostnames"`
	EnableDnsSupport   bool      `json:"enable_dns_support"`
}

// AwsCidr aws cidr
type AwsCidr struct {
	Type        enumor.IPAddressType `json:"type"`
	Cidr        string               `json:"cidr"`
	AddressPool string               `json:"address_pool"`
	State       string               `json:"state"`
}

// GcpVpcUpdateExt defines gcp vpc extensional info.
type GcpVpcUpdateExt struct {
	AutoCreateSubnetworks bool   `json:"auto_create_subnetworks"`
	EnableUlaInternalIpv6 bool   `json:"enable_ula_internal_ipv6"`
	Mtu                   int64  `json:"mtu"`
	RoutingMode           string `json:"routing_mode"`
}

// AzureVpcUpdateExt defines azure vpc extensional info.
type AzureVpcUpdateExt struct {
	ResourceGroup string      `json:"resource_group"`
	Region        string      `json:"region"`
	DNSServers    []*string   `json:"dns_servers"`
	Cidr          []AzureCidr `json:"cidr"`
}

// AzureCidr azure cidr
type AzureCidr struct {
	Type enumor.IPAddressType `json:"type"`
	Cidr string               `json:"cidr"`
}

// HuaWeiVpcUpdateExt defines huawei vpc extensional info.
type HuaWeiVpcUpdateExt struct {
	Region              string       `json:"region"`
	Cidr                []HuaWeiCidr `json:"cidr"`
	Status              string       `json:"status"`
	EnterpriseProjectId string       `json:"enterprise_project_id"`
}

// HuaWeiCidr huawei cidr
type HuaWeiCidr struct {
	Type enumor.IPAddressType `json:"type"`
	Cidr string               `json:"cidr"`
}

// VpcAttachment vpc attachment.
type VpcAttachment struct {
	BkCloudID int64 `json:"bk_cloud_id"`
	BkBizID   int64 `json:"bk_biz_id"`
}

// VpcAttachmentBatchUpdateReq defines batch update vpc attachment request.
type VpcAttachmentBatchUpdateReq struct {
	Attachments []VpcAttachmentUpdateReq `json:"attachments" validate:"required"`
}

// Validate VpcAttachmentBatchUpdateReq.
func (u *VpcAttachmentBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// VpcAttachmentUpdateReq defines update vpc attachment request.
type VpcAttachmentUpdateReq struct {
	ID         string               `json:"id" validate:"required"`
	Attachment *VpcUpdateAttachment `json:"attachment" validate:"required"`
}

// VpcUpdateAttachment defines update vpc attachment.
type VpcUpdateAttachment struct {
	BkCloudID int64 `json:"bk_cloud_id" validate:"required"`
	BkBizID   int64 `json:"bk_biz_id" validate:"required"`
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
