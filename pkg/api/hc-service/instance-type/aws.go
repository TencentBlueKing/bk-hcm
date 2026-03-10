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

// AwsGpuInstanceTypeListReq is the request for listing GPU instance types via AssumeRole.
type AwsGpuInstanceTypeListReq struct {
	CloudID  string `json:"cloud_id" validate:"required"`
	RoleName string `json:"role_name" validate:"required"`
	Region   string `json:"region" validate:"required"`
}

// Validate ...
func (req *AwsGpuInstanceTypeListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsGpuInstanceListReq is the request for listing GPU instances via AssumeRole.
type AwsGpuInstanceListReq struct {
	CloudID  string `json:"cloud_id" validate:"required"`
	RoleName string `json:"role_name" validate:"required"`
	Region   string `json:"region" validate:"required"`
}

// Validate ...
func (req *AwsGpuInstanceListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsGpuInstanceResp represents a single EC2 instance in the GPU instance list response.
type AwsGpuInstanceResp struct {
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	State        string `json:"state"`
	PrivateIP    string `json:"private_ip"`
	PublicIP     string `json:"public_ip"`
	Region       string `json:"region"`
	Zone         string `json:"zone"`
}

// AwsGpuInstanceTypeListResp wraps a list of AwsInstanceTypeResp for GPU instance type queries.
type AwsGpuInstanceTypeListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*AwsInstanceTypeResp `json:"data"`
}

// AwsGpuInstanceListResp wraps a list of AwsGpuInstanceResp.
type AwsGpuInstanceListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*AwsGpuInstanceResp `json:"data"`
}
