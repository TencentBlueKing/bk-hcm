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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// TCloudEipListOption ...
type TCloudEipListOption struct {
	Region   string           `json:"region" validate:"required"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Ips      []string         `json:"ips" validate:"omitempty"`
}

// Validate ...
func (o *TCloudEipListOption) Validate() error {
	if err := validator.Validate.Struct(o); err != nil {
		return err
	}
	if o.Page != nil {
		if err := o.Page.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// TCloudEipListResult ...
type TCloudEipListResult struct {
	Count   *uint64
	Details []*TCloudEip
}

// TCloudEip ...
type TCloudEip struct {
	CloudID                 string
	Name                    *string
	Region                  string
	InstanceId              *string
	Status                  *string
	PublicIp                *string
	PrivateIp               *string
	Bandwidth               *uint64
	InternetChargeType      *string
	InternetServiceProvider *string
}

// GetCloudID ...
func (eip *TCloudEip) GetCloudID() string {
	return eip.CloudID
}

// TCloudEipDeleteOption ...
type TCloudEipDeleteOption struct {
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	Region   string   `json:"region" validate:"required"`
}

// Validate ...
func (opt *TCloudEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToReleaseAddressesRequest ...
func (opt *TCloudEipDeleteOption) ToReleaseAddressesRequest() (*vpc.ReleaseAddressesRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewReleaseAddressesRequest()
	req.AddressIds = common.StringPtrs(opt.CloudIDs)

	return req, nil
}

// TCloudEipAssociateOption ...
type TCloudEipAssociateOption struct {
	Region                  string `json:"region" validate:"required"`
	CloudEipID              string `json:"cloud_eip_id" validate:"required"`
	CloudCvmID              string `json:"cloud_cvm_id" validate:"required"`
	CloudNetworkInterfaceID string `json:"cloud_network_interface_id"`
}

// Validate ...
func (opt *TCloudEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAssociateAddressRequest ...
func (opt *TCloudEipAssociateOption) ToAssociateAddressRequest() (*vpc.AssociateAddressRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewAssociateAddressRequest()
	req.AddressId = common.StringPtr(opt.CloudEipID)

	if opt.CloudNetworkInterfaceID != "" {
		req.NetworkInterfaceId = &opt.CloudNetworkInterfaceID
	} else {
		req.InstanceId = common.StringPtr(opt.CloudCvmID)
	}

	return req, nil
}

// TCloudEipDisassociateOption ...
type TCloudEipDisassociateOption struct {
	Region     string `json:"region" validate:"required"`
	CloudEipID string `json:"cloud_eip_id" validate:"required"`
}

// Validate ...
func (opt *TCloudEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDisassociateAddressRequest ...
func (opt *TCloudEipDisassociateOption) ToDisassociateAddressRequest() (*vpc.DisassociateAddressRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewDisassociateAddressRequest()
	req.AddressId = common.StringPtr(opt.CloudEipID)
	return req, nil
}

// TCloudEipCreateOption ...
type TCloudEipCreateOption struct {
	Region          string  `json:"region" validate:"required"`
	EipName         *string `json:"eip_name"`
	EipCount        int64   `json:"eip_count"  validate:"required"`
	ServiceProvider string  `json:"service_provider" validate:"required,eq=BGP"`
	AddressType     string  `json:"address_type" validate:"required,eq=EIP"`
}

// Validate ...
func (opt *TCloudEipCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAllocateAddressesRequest ...
func (opt *TCloudEipCreateOption) ToAllocateAddressesRequest() (*vpc.AllocateAddressesRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewAllocateAddressesRequest()
	req.AddressCount = common.Int64Ptr(opt.EipCount)
	req.InternetServiceProvider = common.StringPtr(opt.ServiceProvider)
	req.AddressName = opt.EipName
	req.AddressType = common.StringPtr(opt.AddressType)

	return req, nil
}

// TCloudAddressChargePrepaid ...
type TCloudAddressChargePrepaid struct {
	Period        int64 `json:"period"`
	AutoRenewFlag int64 `json:"auto_renew_flag"`
}
