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

package networkinterface

import (
	"hcm/pkg/adaptor/types/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// AzureInterfaceListResult defines azure list result.
type AzureInterfaceListResult struct {
	Details []AzureNI `json:"details"`
}

// HuaWeiInterfaceListResult defines huawei list result.
type HuaWeiInterfaceListResult struct {
	Details []HuaWeiNI `json:"details"`
}

// GcpInterfaceListResult defines gcp list result.
type GcpInterfaceListResult struct {
	Details []GcpNI `json:"details"`
}

// AzureNI defines azure network interface.
type AzureNI CloudNetworkInterface[coreni.AzureNIExtension]

// GetCloudID ...
func (ni AzureNI) GetCloudID() string {
	return *ni.CloudID
}

// HuaWeiNI defines huawei network interface.
type HuaWeiNI CloudNetworkInterface[coreni.HuaWeiNIExtension]

// GetCloudID ...
func (ni HuaWeiNI) GetCloudID() string {
	return *ni.CloudID
}

// GcpNI defines gcp network interface.
type GcpNI CloudNetworkInterface[coreni.GcpNIExtension]

// GetCloudID ...
func (ni GcpNI) GetCloudID() string {
	return *ni.CloudID
}

// CloudNetworkInterface defines network interface struct.
type CloudNetworkInterface[T coreni.NetworkInterfaceExtension] struct {
	Name          *string  `json:"name"`
	AccountID     *string  `json:"account_id"`
	Region        *string  `json:"region"`
	Zone          *string  `json:"zone"`
	CloudID       *string  `json:"cloud_id"`
	VpcID         *string  `json:"vpc_id"`
	CloudVpcID    *string  `json:"cloud_vpc_id"`
	SubnetID      *string  `json:"subnet_id"`
	CloudSubnetID *string  `json:"cloud_subnet_id"`
	PrivateIPv4   []string `json:"private_ipv4"`
	PrivateIPv6   []string `json:"private_ipv6"`
	PublicIPv4    []string `json:"public_ipv4"`
	PublicIPv6    []string `json:"public_ipv6"`
	BkBizID       *int64   `json:"bk_biz_id"`
	InstanceID    *string  `json:"instance_id"`
	Extension     *T       `json:"extension"`
}

// AzureInterfaceListResponse defines azure list network interface response.
type AzureInterfaceListResponse struct {
	NextPageToken string `json:"next_page_token,omitempty"`
}

// AzureNetworkInterfaceListOption define azure network interface list option.
type AzureNetworkInterfaceListOption struct {
	Region string           `json:"region" validate:"required"`
	Page   *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate azure network interface list option.
func (opt AzureNetworkInterfaceListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiNIListOption defines huawei network interface list options.
type HuaWeiNIListOption struct {
	ServerID string `json:"server_id" validate:"required"`
	Region   string `json:"region" validate:"required"`
}

// Validate huawei network interface list option.
func (v HuaWeiNIListOption) Validate() error {
	if err := v.Validate(); err != nil {
		return err
	}

	return nil
}

// HuaWeiEipListOption defines huawei eip list options.
type HuaWeiEipListOption struct {
	Region      string   `json:"region" validate:"required"`
	VnicPortIDs []string `json:"vnic_port_ids" validate:"omitempty"`
}

// Validate huawei huawei list option.
func (v HuaWeiEipListOption) Validate() error {
	if err := v.Validate(); err != nil {
		return err
	}

	return nil
}

// HuaWeiPortInfoOption defines huawei port info options.
type HuaWeiPortInfoOption struct {
	Region         string          `json:"region" validate:"required"`
	PortID         string          `json:"port_id" validate:"omitempty"`
	NetID          string          `json:"net_id" validate:"omitempty"`
	IPv4AddressMap map[string]bool `json:"ipv4_address" validate:"omitempty"`
	IPv6AddressMap map[string]bool `json:"ipv6_address" validate:"omitempty"`
}

// Validate port info option.
func (v HuaWeiPortInfoOption) Validate() error {
	if err := v.Validate(); err != nil {
		return err
	}

	return nil
}

// GcpListByCvmIDOption defines basic gcp list options.
type GcpListByCvmIDOption struct {
	CloudCvmIDs []string `json:"cloud_cvm_ids" validate:"required"`
	Zone        string   `json:"zone" validate:"required"`
}

// Validate gcp list option.
func (opt GcpListByCvmIDOption) Validate() error {
	if len(opt.CloudCvmIDs) > core.GcpQueryLimit {
		return errf.Newf(errf.InvalidParameter, "resource ids length should <= %d", len(opt.CloudCvmIDs))
	}

	if len(opt.CloudCvmIDs) == 0 {
		return errf.New(errf.InvalidParameter, "cloud cvm ids is required")
	}

	if len(opt.Zone) == 0 {
		return errf.New(errf.InvalidParameter, "zone is required")
	}
	return nil
}
