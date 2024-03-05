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

package hcclb

import (
	"errors"

	typeclb "hcm/pkg/adaptor/types/clb"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudBatchCreateReq tcloud batch create req.
type TCloudBatchCreateReq struct {
	AccountID               string                          `json:"account_id" validate:"required"`
	Region                  string                          `json:"region" validate:"required"`
	LoadBalancerType        typeclb.TCloudLoadBalancerType  `json:"load_balancer_type" validate:"required"`
	Name                    *string                         `json:"name" validate:"required,max=60"`
	Zones                   []string                        `json:"zones" validate:"omitempty"`
	BackupZones             []string                        `json:"backup_zones" validate:"omitempty"`
	AddressIPVersion        *typeclb.TCloudAddressIPVersion `json:"address_ip_version" validate:"required"`
	CloudVpcID              *string                         `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           *string                         `json:"cloud_subnet_id" validate:"omitempty"`
	Vip                     *string                         `json:"vip" validate:"omitempty"`
	VipID                   *string                         `json:"vip_id" validate:"omitempty"`
	VipIsp                  *string                         `json:"vip_isp" validate:"omitempty"`
	InternetChargeType      *string                         `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64                          `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthPackageID      *string                         `json:"bandwidth_package_id" validate:"omitempty"`
	SlaType                 *string                         `json:"sla_type" validate:"omitempty"`
	AutoRenew               *bool                           `json:"auto_renew" validate:"omitempty"`
	RequireCount            *uint64                         `json:"require_count" validate:"omitempty"`
	Memo                    string                          `json:"memo" validate:"omitempty"`
}

// Validate request.
func (req *TCloudBatchCreateReq) Validate() error {
	if req.LoadBalancerType == typeclb.InternalLoadBalancerType {
		// 内网校验
		if converter.PtrToVal(req.CloudSubnetID) == "" {
			return errors.New("subnet id  is required for internal type load balancer")
		}
	}
	return validator.Validate.Struct(req)
}

// BatchCreateResult ...
type BatchCreateResult struct {
	UnknownCloudIDs []string `json:"unknown_cloud_ids"`
	SuccessCloudIDs []string `json:"success_cloud_ids"`
	FailedCloudIDs  []string `json:"failed_cloud_ids"`
	FailedMessage   string   `json:"failed_message"`
}

// -------------------------- List Clb--------------------------

// TCloudListOption defines options to list tcloud clb instances.
type TCloudListOption struct {
	AccountID string           `json:"account_id" validate:"required"`
	Region    string           `json:"region" validate:"required"`
	CloudIDs  []string         `json:"cloud_ids" validate:"omitempty,max=200"`
	Page      *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud clb list option.
func (opt TCloudListOption) Validate() error {
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
