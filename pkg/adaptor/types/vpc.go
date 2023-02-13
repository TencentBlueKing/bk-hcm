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

// VpcUpdateOption defines update vpc options.
type VpcUpdateOption struct {
	ResourceID string             `json:"resource_id"`
	Data       *BaseVpcUpdateData `json:"data"`
}

// BaseVpcUpdateData defines the basic update vpc instance data.
type BaseVpcUpdateData struct {
	Memo *string `json:"memo"`
}

// Validate BaseVpcUpdateData.
func (v BaseVpcUpdateData) Validate() error {
	if v.Memo == nil {
		return errf.New(errf.InvalidParameter, "memo is required")
	}
	return nil
}

// TCloudVpcUpdateOption defines tencent cloud update vpc options.
type TCloudVpcUpdateOption struct{}

// Validate TCloudVpcUpdateOption.
func (v TCloudVpcUpdateOption) Validate() error {
	return nil
}

// AwsVpcUpdateOption defines aws update vpc options.
type AwsVpcUpdateOption struct{}

// Validate AwsVpcUpdateOption.
func (v AwsVpcUpdateOption) Validate() error {
	return nil
}

// GcpVpcUpdateOption defines gcp update vpc options.
type GcpVpcUpdateOption VpcUpdateOption

// Validate GcpVpcUpdateOption.
func (v GcpVpcUpdateOption) Validate() error {
	if len(v.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if v.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := v.Data.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureVpcUpdateOption defines azure update vpc options.
type AzureVpcUpdateOption struct{}

// Validate AzureVpcUpdateOption.
func (v AzureVpcUpdateOption) Validate() error {
	return nil
}

// HuaWeiVpcUpdateOption defines huawei update vpc options.
type HuaWeiVpcUpdateOption struct {
	VpcUpdateOption `json:",inline"`
	Region          string `json:"region"`
}

// Validate HuaWeiVpcUpdateOption.
func (v HuaWeiVpcUpdateOption) Validate() error {
	if len(v.Region) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if len(v.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if v.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := v.Data.Validate(); err != nil {
		return err
	}
	return nil
}

// -------------------------- List --------------------------

// TCloudVpcListResult defines tencent cloud list vpc result.
type TCloudVpcListResult struct {
	Count   *uint64     `json:"count,omitempty"`
	Details []TCloudVpc `json:"details"`
}

// AwsVpcListResult defines aws list vpc result.
type AwsVpcListResult struct {
	NextToken *string  `json:"next_token,omitempty"`
	Details   []AwsVpc `json:"details"`
}

// GcpVpcListResult defines gcp list vpc result.
type GcpVpcListResult struct {
	NextPageToken string   `json:"next_page_token,omitempty"`
	Details       []GcpVpc `json:"details"`
}

// AzureVpcListResult defines azure list vpc result.
type AzureVpcListResult struct {
	Details []AzureVpc `json:"details"`
}

// HuaWeiVpcListOption defines huawei list vpc options.
type HuaWeiVpcListOption struct {
	core.HuaWeiListOption `json:",inline"`
	Names                 []string `json:"names,omitempty"`
}

// Validate huawei list option.
func (v HuaWeiVpcListOption) Validate() error {
	if err := v.HuaWeiListOption.Validate(); err != nil {
		return err
	}

	return nil
}

// HuaWeiVpcListResult defines huawei list vpc result.
type HuaWeiVpcListResult struct {
	NextMarker *string     `json:"next_marker,omitempty"`
	Details    []HuaWeiVpc `json:"details"`
}

// Vpc defines vpc struct.
type Vpc[T cloud.VpcExtension] struct {
	CloudID   string  `json:"cloud_id"`
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	Memo      *string `json:"memo,omitempty"`
	Extension *T      `json:"extension"`
}

// TCloudVpc defines tencent cloud vpc.
type TCloudVpc Vpc[cloud.TCloudVpcExtension]

// AwsVpc defines aws vpc.
type AwsVpc Vpc[cloud.AwsVpcExtension]

// GcpVpc defines gcp vpc.
type GcpVpc Vpc[cloud.GcpVpcExtension]

// AzureVpc defines azure vpc.
type AzureVpc Vpc[cloud.AzureVpcExtension]

// HuaWeiVpc defines huawei vpc.
type HuaWeiVpc Vpc[cloud.HuaWeiVpcExtension]

// -------------------------- IP --------------------------

// HuaWeiVpcIPAvailGetOption get huawei vcp ip availabilities option.
type HuaWeiVpcIPAvailGetOption struct {
	Region   string `json:"region"`
	SubnetID string `json:"subnet_id"`
}

// Validate HuaWeiVpcIPAvailGetOption.
func (v HuaWeiVpcIPAvailGetOption) Validate() error {
	if len(v.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(v.SubnetID) == 0 {
		return errf.New(errf.InvalidParameter, "subnetID id is required")
	}

	return nil
}

// AzureVpcListUsageOption defines azure list vpc usage options
type AzureVpcListUsageOption struct {
	ResourceGroupName string `json:"resource_group_name"`
	VpcID             string `json:"vpc_name"`
}

// Validate AzureVpcIPAvailGetOption.
func (v AzureVpcListUsageOption) Validate() error {
	if len(v.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group is required")
	}

	if len(v.VpcID) == 0 {
		return errf.New(errf.InvalidParameter, "vpc id is required")
	}

	return nil
}
