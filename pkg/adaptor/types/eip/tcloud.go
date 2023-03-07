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
	CloudID            string
	Name               *string
	Region             string
	InstanceId         *string
	Status             *string
	PublicIp           *string
	PrivateIp          *string
	Bandwidth          *uint64
	InternetChargeType *string
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

	return &vpc.ReleaseAddressesRequest{AddressIds: common.StringPtrs(opt.CloudIDs)}, nil
}

// TCloudEipAssociateOption ...
type TCloudEipAssociateOption struct {
	Region     string `json:"region" validate:"required"`
	CloudEipID string `json:"cloud_eip_id" validate:"required"`
	CloudCvmID string `json:"cloud_cvm_id" validate:"required"`
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

	return &vpc.AssociateAddressRequest{
		AddressId:  common.StringPtr(opt.CloudEipID),
		InstanceId: common.StringPtr(opt.CloudCvmID),
	}, nil
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

	return &vpc.DisassociateAddressRequest{
		AddressId: common.StringPtr(opt.CloudEipID),
	}, nil
}
