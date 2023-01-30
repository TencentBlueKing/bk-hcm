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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// Vpc defines vpc info.
type Vpc[T VpcExtension] struct {
	BaseVpc   `json:",inline"`
	Extension *T `json:"extension"`
}

// BaseVpc defines base vpc info.
type BaseVpc struct {
<<<<<<< HEAD
	ID        string             `json:"id"`
	Vendor    enumor.Vendor      `json:"vendor"`
	AccountID string             `json:"account_id"`
	CloudID   string             `json:"cloud_id"`
	Name      string             `json:"name"`
	Category  enumor.VpcCategory `json:"category"`
	Memo      *string            `json:"memo,omitempty"`
	BkCloudID int64              `json:"bk_cloud_id"`
	BkBizID   int64              `json:"bk_biz_id"`
	Revision  *core.Revision     `json:"revision"`
=======
	ID             string             `json:"id"`
	Vendor         enumor.Vendor      `json:"vendor"`
	AccountID      string             `json:"account_id"`
	CloudID        string             `json:"cloud_id"`
	Name           string             `json:"name"`
	Category       enumor.VpcCategory `json:"category"`
	Memo           *string            `json:"memo,omitempty"`
	BkCloudID      int64              `json:"bk_cloud_id"`
	BkBizID        int64              `json:"bk_biz_id"`
	*core.Revision `json:",inline"`
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
}

// VpcExtension defines vpc extensional info.
type VpcExtension interface {
	TCloudVpcExtension | AwsVpcExtension | GcpVpcExtension | AzureVpcExtension | HuaWeiVpcExtension
}

// TCloudVpcExtension defines tencent cloud vpc extensional info.
type TCloudVpcExtension struct {
	Region          string       `json:"region"`
	Cidr            []TCloudCidr `json:"cidr"`
	IsDefault       bool         `json:"is_default"`
	EnableMulticast bool         `json:"enable_multicast"`
	DnsServerSet    []string     `json:"dns_server_set"`
	DomainName      string       `json:"domain_name,omitempty"`
}

// TCloudCidr tencent cloud cidr
type TCloudCidr struct {
	Type     enumor.IPAddressType      `json:"type"`
	Cidr     string                    `json:"cidr"`
	Category enumor.TCloudCidrCategory `json:"category"`
}

// AwsVpcExtension defines aws vpc extensional info.
type AwsVpcExtension struct {
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

// GcpVpcExtension defines gcp vpc extensional info.
type GcpVpcExtension struct {
<<<<<<< HEAD
=======
	SelfLink              string `json:"self_link"`
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	AutoCreateSubnetworks bool   `json:"auto_create_subnetworks"`
	EnableUlaInternalIpv6 bool   `json:"enable_ula_internal_ipv6"`
	Mtu                   int64  `json:"mtu"`
	RoutingMode           string `json:"routing_mode"`
}

// AzureVpcExtension defines azure vpc extensional info.
type AzureVpcExtension struct {
	ResourceGroup string      `json:"resource_group"`
	Region        string      `json:"region"`
	DNSServers    []string    `json:"dns_servers"`
	Cidr          []AzureCidr `json:"cidr"`
}

// AzureCidr azure cidr
type AzureCidr struct {
	Type enumor.IPAddressType `json:"type"`
	Cidr string               `json:"cidr"`
}

// HuaWeiVpcExtension defines huawei vpc extensional info.
type HuaWeiVpcExtension struct {
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

// TCloudVpc defines tencent cloud vpc.
type TCloudVpc Vpc[TCloudVpcExtension]

// AwsVpc defines aws vpc.
type AwsVpc Vpc[AwsVpcExtension]

// GcpVpc defines gcp vpc.
type GcpVpc Vpc[GcpVpcExtension]

// AzureVpc defines azure vpc.
type AzureVpc Vpc[AzureVpcExtension]

// HuaWeiVpc defines huawei vpc.
type HuaWeiVpc Vpc[HuaWeiVpcExtension]
