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

package types

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
)

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

// TCloudSubnetUpdateOption defines tencent cloud update subnet options.
type TCloudSubnetUpdateOption struct{}

// Validate TCloudSubnetUpdateOption.
func (s TCloudSubnetUpdateOption) Validate() error {
	return nil
}

// AwsSubnetUpdateOption defines aws update subnet options.
type AwsSubnetUpdateOption struct{}

// Validate AwsSubnetUpdateOption.
func (s AwsSubnetUpdateOption) Validate() error {
	return nil
}

// GcpSubnetUpdateOption defines gcp update subnet options.
type GcpSubnetUpdateOption struct {
	SubnetUpdateOption `json:",inline"`
	Region             string `json:"region"`
}

// Validate GcpSubnetUpdateOption.
func (s GcpSubnetUpdateOption) Validate() error {
	if len(s.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if s.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := s.Data.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureSubnetUpdateOption defines azure update subnet options.
type AzureSubnetUpdateOption struct{}

// Validate AzureSubnetUpdateOption.
func (s AzureSubnetUpdateOption) Validate() error {
	return nil
}

// HuaweiSubnetUpdateOption defines huawei update subnet options.
type HuaweiSubnetUpdateOption struct {
	SubnetUpdateOption `json:",inline"`
	Region             string `json:"region"`
	VpcID              string `json:"vpc_id"`
}

// Validate HuaweiSubnetUpdateOption.
func (s HuaweiSubnetUpdateOption) Validate() error {
	if err := s.Data.Validate(); err != nil {
		return err
	}

	if len(s.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(s.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}
	return nil
}

// ------------------------- Delete -------------------------

// HuaweiSubnetDeleteOption defines huawei delete subnet options.
type HuaweiSubnetDeleteOption struct {
	core.BaseRegionalDeleteOption `json:",inline"`
	VpcID                         string `json:"vpc_id"`
}

// Validate HuaweiSubnetDeleteOption.
func (s HuaweiSubnetDeleteOption) Validate() error {
	if err := s.BaseRegionalDeleteOption.Validate(); err != nil {
		return err
	}

	if len(s.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}
	return nil
}

// -------------------------- List --------------------------

// TCloudSubnetListResult defines tencent cloud list subnet result.
type TCloudSubnetListResult struct {
	Count   *uint64        `json:"count,omitempty"`
	Details []TCloudSubnet `json:"details"`
}

// AwsSubnetListResult defines aws list subnet result.
type AwsSubnetListResult struct {
	NextToken *string     `json:"next_token,omitempty"`
	Details   []AwsSubnet `json:"details"`
}

// GcpSubnetListOption basic gcp list subnet options.
type GcpSubnetListOption struct {
	core.GcpListOption `json:",inline"`
	Region             string `json:"region"`
}

// Validate gcp list subnet option.
func (g GcpSubnetListOption) Validate() error {
	if err := g.GcpListOption.Validate(); err != nil {
		return err
	}

	if len(g.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region can be empty")
	}

	return nil
}

// GcpSubnetListResult defines gcp list subnet result.
type GcpSubnetListResult struct {
	NextPageToken string      `json:"next_page_token,omitempty"`
	Details       []GcpSubnet `json:"details"`
}

// AzureSubnetListOption defines azure list subnet options.
type AzureSubnetListOption struct {
	core.AzureListOption `json:",inline"`
	VpcName              string `json:"vpc_name"`
}

// AzureSubnetListResult defines azure list subnet result.
type AzureSubnetListResult struct {
	Details []AzureSubnet `json:"details"`
}

// HuaweiSubnetListOption defines huawei list subnet options.
type HuaweiSubnetListOption struct {
	Region string           `json:"region"`
	Page   *core.HuaweiPage `json:"page,omitempty"`
	VpcID  string           `json:"vpc_id,omitempty"`
}

// Validate huawei list option.
func (s HuaweiSubnetListOption) Validate() error {
	if len(s.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if s.Page != nil {
		if err := s.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaweiSubnetListResult defines huawei list subnet result.
type HuaweiSubnetListResult struct {
	Details []HuaWeiSubnet `json:"details"`
}

// Subnet defines subnet struct.
type Subnet[T cloud.SubnetExtension] struct {
	CloudVpcID string   `json:"cloud_vpc_id"`
	CloudID    string   `json:"cloud_id"`
	Name       string   `json:"name"`
	Ipv4Cidr   []string `json:"ipv4_cidr,omitempty"`
	Ipv6Cidr   []string `json:"ipv6_cidr,omitempty"`
	Memo       *string  `json:"memo,omitempty"`
	Extension  *T       `json:"extension"`
}

// TCloudSubnet defines tencent cloud subnet.
type TCloudSubnet Subnet[cloud.TCloudSubnetExtension]

// AwsSubnet defines aws subnet.
type AwsSubnet Subnet[cloud.AwsSubnetExtension]

// GcpSubnet defines gcp subnet.
type GcpSubnet Subnet[cloud.GcpSubnetExtension]

// AzureSubnet defines azure subnet.
type AzureSubnet Subnet[cloud.AzureSubnetExtension]

// HuaWeiSubnet defines huawei subnet.
type HuaWeiSubnet Subnet[cloud.HuaWeiSubnetExtension]
