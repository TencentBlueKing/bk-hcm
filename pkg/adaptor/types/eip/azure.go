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
	Status            *string
	PublicIp          *string
	PrivateIp         *string
	IpConfigurationID *string
	SKU               *string
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
