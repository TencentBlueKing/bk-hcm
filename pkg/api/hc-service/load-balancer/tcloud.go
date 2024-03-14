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

package hclb

import (
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudBatchCreateReq tcloud batch create req.
type TCloudBatchCreateReq struct {
	AccountID        string                        `json:"account_id" validate:"required"`
	Region           string                        `json:"region" validate:"required"`
	LoadBalancerType typelb.TCloudLoadBalancerType `json:"load_balancer_type" validate:"required"`
	Name             *string                       `json:"name" validate:"required,max=60"`
	// 公网	单可用区		传递zones（单元素数组）
	// 公网	主备可用区	传递zones（单元素数组），以及backup_zones
	Zones                   []string                       `json:"zones" validate:"omitempty"`
	BackupZones             []string                       `json:"backup_zones" validate:"omitempty"`
	AddressIPVersion        *typelb.TCloudAddressIPVersion `json:"address_ip_version" validate:"required"`
	CloudVpcID              *string                        `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           *string                        `json:"cloud_subnet_id" validate:"omitempty"`
	Vip                     *string                        `json:"vip" validate:"omitempty"`
	CloudEipID              *string                        `json:"cloud_eip_id" validate:"omitempty"`
	VipIsp                  *string                        `json:"vip_isp" validate:"omitempty"`
	InternetChargeType      *string                        `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64                         `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthPackageID      *string                        `json:"bandwidth_package_id" validate:"omitempty"`
	SlaType                 *string                        `json:"sla_type" validate:"omitempty"`
	AutoRenew               *bool                          `json:"auto_renew" validate:"omitempty"`
	RequireCount            *uint64                        `json:"require_count" validate:"omitempty"`
	Memo                    string                         `json:"memo" validate:"omitempty"`
}

// Validate request.
func (req *TCloudBatchCreateReq) Validate() error {
	switch req.LoadBalancerType {
	case typelb.InternalLoadBalancerType:
		// 内网校验
		if converter.PtrToVal(req.CloudSubnetID) == "" {
			return errors.New("subnet id is required for load balancer type 'INTERNAL'")
		}
	case typelb.OpenLoadBalancerType:
		if converter.PtrToVal(req.CloudEipID) != "" {
			return errors.New("eip id only support load balancer type 'INTERNAL'")
		}
	default:
		return fmt.Errorf("unknown load balancer type: '%s'", req.LoadBalancerType)
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

// TCloudDescribeResourcesOption ...
type TCloudDescribeResourcesOption struct {
	AccountID                             string `json:"account_id" validate:"required"`
	*typelb.TCloudDescribeResourcesOption `json:",inline" validate:"required"`
}

// Validate tcloud clb list option.
func (opt TCloudDescribeResourcesOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// --------------------------[Associate 设置负载均衡实例的安全组]--------------------------

// TCloudSetClbSecurityGroupReq defines options to set tcloud clb security-group request.
type TCloudSetClbSecurityGroupReq struct {
	LbID             string   `json:"lb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate tcloud clb security-group option.
func (opt TCloudSetClbSecurityGroupReq) Validate() error {
	if len(opt.LbID) == 0 {
		return errors.New("lb_id is required")
	}

	if len(opt.SecurityGroupIDs) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("invalid security_group_ids max value: %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return validator.Validate.Struct(opt)
}

// --------------------------[DisAssociate 设置负载均衡实例的安全组]--------------------------

// TCloudDisAssociateClbSecurityGroupReq defines options to DisAssociate tcloud clb security-group request.
type TCloudDisAssociateClbSecurityGroupReq struct {
	LbID            string `json:"lb_id" validate:"required"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
}

// Validate tcloud clb security-group option.
func (opt TCloudDisAssociateClbSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudUpdateReq security group update request.
type TCloudUpdateReq struct {
	ClbID string `json:"load_balancer_id" validate:"omitempty"`

	Name *string `json:"name" validate:"omitempty"`
	Memo *string `json:"memo" validate:"omitempty"`

	// 网络计费相关参数
	InternetChargeType      *string `json:"internet_charge_type" validate:"omitempty"`
	InternetMaxBandwidthOut *int64  `json:"internet_max_bandwidth_out" validate:"omitempty"`
	BandwidthpkgSubType     *string `json:"bandwidthpkg_sub_type" validate:"omitempty"`

	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组。
	LoadBalancerPassToTarget *bool `json:"load_balancer_pass_to_target" validate:"omitempty"`
	SnatPro                  *bool `json:"snat_pro" validate:"omitempty"`
	DeleteProtect            *bool `json:"delete_protect" validate:"omitempty"`
	ModifyClassicDomain      *bool `json:"modify_classic_domain" validate:"omitempty"`
}

// Validate tcloud security group update request.
func (req *TCloudUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
