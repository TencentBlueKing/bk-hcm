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

package adtysubnet

import (
	"hcm/pkg/criteria/errf"
)

// -------------------------- Create --------------------------

// SubnetCreateOption defines create subnet options.
type SubnetCreateOption[T SubnetCreateExt] struct {
	Name       string  `json:"name" validate:"required"`
	Memo       *string `json:"memo,omitempty" validate:"omitempty"`
	CloudVpcID string  `json:"cloud_vpc_id" validate:"required"`
	Extension  *T      `json:"extension" validate:"required"`
}

// SubnetCreateExt defines create subnet extensional info.
type SubnetCreateExt interface {
	TCloudSubnetCreateExt | AwsSubnetCreateExt | GcpSubnetCreateExt | AzureSubnetCreateExt | HuaWeiSubnetCreateExt
}

// -------------------------- Update --------------------------

// SubnetUpdateOption defines update subnet options.
type SubnetUpdateOption struct {
	ResourceID string                `json:"resource_id"`
	Data       *BaseSubnetUpdateData `json:"data"`
}

// BaseSubnetUpdateData defines the basic update subnet instance data.
type BaseSubnetUpdateData struct {
	Memo *string `json:"memo"`
}

// Validate BaseSubnetUpdateData.
func (s BaseSubnetUpdateData) Validate() error {
	if s.Memo == nil {
		return errf.New(errf.InvalidParameter, "memo is required")
	}
	return nil
}

// -------------------------- List --------------------------

// Subnet defines subnet struct.
type Subnet[T SubnetExtension] struct {
	// TODO: gcp 添加 vpcSelfLink字段，不要和 CloudVpcID 字段混用
	// CloudVpcID gcp 该字段为 self_link
	CloudVpcID string   `json:"cloud_vpc_id"`
	CloudID    string   `json:"cloud_id"`
	Name       string   `json:"name"`
	Ipv4Cidr   []string `json:"ipv4_cidr,omitempty"`
	Ipv6Cidr   []string `json:"ipv6_cidr,omitempty"`
	Memo       *string  `json:"memo,omitempty"`
	Extension  *T       `json:"extension"`
}

// SubnetExtension defines subnet extensional info.
type SubnetExtension interface {
	TCloudSubnetExtension | AwsSubnetExtension | GcpSubnetExtension | AzureSubnetExtension | HuaWeiSubnetExtension
}
