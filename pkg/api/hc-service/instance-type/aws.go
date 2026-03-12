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

package instancetype

import (
	"encoding/json"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// AwsInstanceTypeListReq ...
type AwsInstanceTypeListReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate ...
func (req *AwsInstanceTypeListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsInstanceTypeResp ...
type AwsInstanceTypeResp struct {
	InstanceFamily     string `json:"instance_family"`
	InstanceType       string `json:"instance_type"`
	GPU                int64  `json:"gpu"`
	GPUMemory          int64  `json:"gpu_memory"`
	GPUName            string `json:"gpu_name"`
	GPUManufacturer    string `json:"gpu_manufacturer"`
	CPU                int64  `json:"cpu"`
	Memory             int64  `json:"memory"`
	FPGA               int64  `json:"fpga"`
	NetworkPerformance string `json:"network_performance"`
	DiskSizeInGB       int64  `json:"disk_size_in_gb"`
	Architecture       string `json:"architecture"`
	DiskType           string `json:"disk_type"`
}

// AwsInstanceTypeListResp ...
type AwsInstanceTypeListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*AwsInstanceTypeResp `json:"data"`
}

// AwsAssumeRoleInstanceTypeListReq is the request for listing instance types via AssumeRole.
type AwsAssumeRoleInstanceTypeListReq struct {
	RootAccountID string   `json:"root_account_id" validate:"required"`
	MainAccountID string   `json:"main_account_id" validate:"required"`
	RoleChain     []string `json:"role_chain" validate:"required,min=1"`
	Region        string   `json:"region" validate:"required"`
	ExternalID    string   `json:"external_id,omitempty"`
}

// Validate ...
func (req *AwsAssumeRoleInstanceTypeListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleInstanceListReq is the request for listing instances via AssumeRole.
type AwsAssumeRoleInstanceListReq struct {
	RootAccountID string   `json:"root_account_id" validate:"required"`
	MainAccountID string   `json:"main_account_id" validate:"required"`
	RoleChain     []string `json:"role_chain" validate:"required,min=1"`
	Region        string   `json:"region" validate:"required"`
	ExternalID    string   `json:"external_id,omitempty"`
}

// Validate ...
func (req *AwsAssumeRoleInstanceListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleInstanceTypeListResp wraps a list of AwsInstanceTypeResp for AssumeRole instance type queries.
type AwsAssumeRoleInstanceTypeListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*AwsInstanceTypeResp `json:"data"`
}

// AwsAssumeRoleInstanceListResp wraps the raw AWS EC2 DescribeInstances response for transparent pass-through.
type AwsAssumeRoleInstanceListResp struct {
	rest.BaseResp `json:",inline"`
	Data          json.RawMessage `json:"data"`
}
