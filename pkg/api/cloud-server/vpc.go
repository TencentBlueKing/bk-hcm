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

package cloudserver

import (
	"encoding/json"

	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// VpcCreateReq defines create vpc request.
type VpcCreateReq struct {
	Vendor    enumor.Vendor      `json:"vendor" validate:"required"`
	AccountID string             `json:"account_id" validate:"required"`
	Name      string             `json:"name" validate:"required"`
	Region    string             `json:"region" validate:"required"`
	Category  enumor.VpcCategory `json:"category" validate:"required"`
	Memo      *string            `json:"memo,omitempty" validate:"omitempty"`
	BkCloudID int64              `json:"bk_cloud_id" validate:"required"`
	Extension json.RawMessage    `json:"extension" validate:"required"`
}

// Validate VpcCreateReq.
func (c VpcCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// VpcCreateExt defines vpc extensional info.
type VpcCreateExt interface {
	TCloudVpcCreateExt | AwsVpcCreateExt | GcpVpcCreateExt | AzureVpcCreateExt | HuaWeiVpcCreateExt
}

// TODO add subnets options in vpc extensions

// TCloudVpcCreateExt defines tencent cloud vpc extensional info.
type TCloudVpcCreateExt struct {
	IPv4Cidr string `json:"ipv4_cidr" validate:"required"`
}

// AwsVpcCreateExt defines aws vpc extensional info.
type AwsVpcCreateExt struct {
	IPv4Cidr                    string `json:"ipv4_cidr" validate:"required"`
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
	ResourceGroup string   `json:"resource_group" validate:"required"`
	IPv4Cidr      []string `json:"ipv4_cidr" validate:"omitempty"`
	IPv6Cidr      []string `json:"ipv6_cidr" validate:"omitempty"`
}

// HuaWeiVpcCreateExt defines huawei vpc extensional info.
type HuaWeiVpcCreateExt struct {
	IPv4Cidr            string  `json:"ipv4_cidr" validate:"required"`
	EnterpriseProjectID *string `json:"enterprise_project_id" validate:"omitempty"`
}

// -------------------------- Update --------------------------

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq struct {
	Memo *string `json:"memo" validate:"required"`
}

// Validate VpcUpdateReq.
func (u VpcUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- List --------------------------

// VpcListResult defines list vpc result.
type VpcListResult struct {
	Count   uint64          `json:"count"`
	Details []cloud.BaseVpc `json:"details"`
}

// -------------------------- Relation ------------------------

// AssignVpcToBizReq assign vpcs to biz request.
type AssignVpcToBizReq struct {
	VpcIDs  []string `json:"vpc_ids"`
	BkBizID int64    `json:"bk_biz_id"`
}

// Validate AssignVpcToBizReq.
func (a AssignVpcToBizReq) Validate() error {
	if len(a.VpcIDs) == 0 {
		return errf.New(errf.InvalidParameter, "vpc ids are required")
	}

	if a.BkBizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is required")
	}

	return nil
}

// BindVpcWithCloudAreaReq bind vpcs with bizs request.
type BindVpcWithCloudAreaReq []VpcCloudAreaRelation

// VpcCloudAreaRelation vpc and cloud area relation.
type VpcCloudAreaRelation struct {
	VpcID     string `json:"vpc_id"`
	BkCloudID int64  `json:"bk_cloud_id"`
}

// Validate BindVpcWithCloudAreaReq.
func (b BindVpcWithCloudAreaReq) Validate() error {
	if len(b) == 0 {
		return errf.New(errf.InvalidParameter, "bind vpc with cloud area request can not be empty")
	}

	for _, relation := range b {
		if len(relation.VpcID) == 0 {
			return errf.New(errf.InvalidParameter, "vpc id is required")
		}

		if relation.BkCloudID == 0 {
			return errf.New(errf.InvalidParameter, "cloud id is required")
		}
	}

	return nil
}
