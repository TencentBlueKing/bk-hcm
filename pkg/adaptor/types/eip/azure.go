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
	"hcm/pkg/criteria/validator"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// PublicIPAddressSKUEnum ...
var PublicIPAddressSKUEnum = map[string]armnetwork.PublicIPAddressSKUName{
	"Standard": armnetwork.PublicIPAddressSKUNameStandard,
	"Basic":    armnetwork.PublicIPAddressSKUNameBasic,
}

// PublicIPAddressSKUTierEnum ...
var PublicIPAddressSKUTierEnum = map[string]armnetwork.PublicIPAddressSKUTier{
	"Global":   armnetwork.PublicIPAddressSKUTierGlobal,
	"Regional": armnetwork.PublicIPAddressSKUTierRegional,
}

// IPAllocationMethodEnum ...
var IPAllocationMethodEnum = map[string]armnetwork.IPAllocationMethod{
	"Dynamic": armnetwork.IPAllocationMethodDynamic,
	"Static":  armnetwork.IPAllocationMethodStatic,
}

// IPVersionEnum ...
var IPVersionEnum = map[string]armnetwork.IPVersion{
	"ipv6": armnetwork.IPVersionIPv6,
	"ipv4": armnetwork.IPVersionIPv4,
}

// AzureEipListOption ...
type AzureEipListOption struct {
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	Ips      []string `json:"ips" validate:"omitempty"`
}

// AzureEipListResult ...
type AzureEipListResult struct {
	Details []*AzureEip
}

// AzureEip ...
type AzureEip struct {
	CloudID           string
	Name              *string
	Region            string
	InstanceId        *string
	InstanceType      string
	Status            *string
	PublicIp          *string
	PrivateIp         *string
	IpConfigurationID *string
	SKU               *string
	SKUTier           *string
	ResourceGroupName string
	Location          *string
	Fqdn              *string
	Zone              string
}

// AzureEipDeleteOption ...
type AzureEipDeleteOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	EipName           string `json:"eip_name" validate:"required"`
}

// Validate ...
func (opt *AzureEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureEipAssociateOption ...
type AzureEipAssociateOption struct {
	ResourceGroupName string                `json:"resource_group_name" validate:"required"`
	CloudEipID        string                `json:"cloud_eip_id" validate:"required"`
	NetworkInterface  *armnetwork.Interface `json:",inline"  validate:"required"`
}

// Validate ...
func (opt *AzureEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToInterfaceParams ...
func (opt *AzureEipAssociateOption) ToInterfaceParams() (*armnetwork.Interface, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	params := opt.NetworkInterface

	firstIpConfig := params.Properties.IPConfigurations[0]
	firstIpConfig.Properties.PublicIPAddress = &armnetwork.PublicIPAddress{ID: to.Ptr(opt.CloudEipID)}

	return params, nil
}

// AzureEipDisassociateOption ...
type AzureEipDisassociateOption struct {
	ResourceGroupName string                `json:"resource_group_name" validate:"required"`
	CloudEipID        string                `json:"cloud_eip_id" validate:"required"`
	NetworkInterface  *armnetwork.Interface `json:",inline"  validate:"required"`
}

// Validate ...
func (opt *AzureEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToInterfaceParams ...
func (opt *AzureEipDisassociateOption) ToInterfaceParams() (*armnetwork.Interface, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	params := opt.NetworkInterface

	firstIpConfig := params.Properties.IPConfigurations[0]
	firstIpConfig.Properties.PublicIPAddress = nil

	return params, nil
}

// AzureEipCreateOption ...
type AzureEipCreateOption struct {
	ResourceGroupName    string `json:"resource_group_name" validate:"required"`
	EipName              string `json:"eip_name" validate:"required"`
	Region               string `json:"region" validate:"required"`
	Zone                 string `json:"zone" validate:"omitempty"`
	SKUName              string `json:"sku_name" validate:"required,eq=Standard|eq=Basic"`
	SKUTier              string `json:"sku_tier" validate:"required,eq=Regional|eq=Global"`
	AllocationMethod     string `json:"allocation_method" validate:"required,eq=Dynamic|eq=Static"`
	IPVersion            string `json:"ip_version" validate:"required,eq=ipv6|eq=ipv4"`
	IdleTimeoutInMinutes int32  `json:"idle_timeout_in_minutes" validate:"required"`
}

// Validate ...
func (opt *AzureEipCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToPublicIPAddress ...
func (opt *AzureEipCreateOption) ToPublicIPAddress() (*armnetwork.PublicIPAddress, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &armnetwork.PublicIPAddress{Location: to.Ptr(opt.Region)}

	allocationMethod := IPAllocationMethodEnum[opt.AllocationMethod]
	ipVersion := IPVersionEnum[opt.IPVersion]
	req.Properties = &armnetwork.PublicIPAddressPropertiesFormat{
		IdleTimeoutInMinutes:     to.Ptr(opt.IdleTimeoutInMinutes),
		PublicIPAllocationMethod: to.Ptr(allocationMethod),
		PublicIPAddressVersion:   to.Ptr(ipVersion),
	}
	req.SKU = &armnetwork.PublicIPAddressSKU{
		Name: to.Ptr(PublicIPAddressSKUEnum[opt.SKUName]),
		Tier: to.Ptr(PublicIPAddressSKUTierEnum[opt.SKUTier]),
	}

	if opt.Zone != "" {
		req.Zones = to.SliceOfPtrs[string](opt.Zone)
	}

	return req, nil
}
