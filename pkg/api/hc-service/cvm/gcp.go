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

package hccvm

import (
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// GcpOperateSyncReq cvm oprate sync request.
type GcpOperateSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	Zone      string   `json:"zone" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate cvm operate sync request.
func (req *GcpOperateSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	if len(req.CloudIDs) <= 0 {
		return fmt.Errorf("operate sync count should > 0")
	}

	return validator.Validate.Struct(req)
}

// GcpBatchCreateReq gcp batch create req.
type GcpBatchCreateReq struct {
	AccountID     string `json:"account_id" validate:"required"`
	NamePrefix    string `json:"name_prefix" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Zone          string `json:"zone" validate:"required"`
	InstanceType  string `json:"instance_type" validate:"required"`
	CloudImageID  string `json:"cloud_image_id" validate:"required"`
	Password      string `json:"password" validate:"required"`
	RequiredCount int64  `json:"required_count" validate:"required"`
	// RequestID 唯一标识支持生产请求
	RequestID        string                `json:"request_id" validate:"omitempty"`
	CloudVpcID       string                `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID    string                `json:"cloud_subnet_id" validate:"required"`
	Description      string                `json:"description" validate:"omitempty"`
	SystemDisk       *typecvm.GcpOsDisk    `json:"system_disk" validate:"required"`
	DataDisk         []typecvm.GcpDataDisk `json:"data_disk" validate:"omitempty"`
	PublicIPAssigned bool                  `json:"public_ip_assigned" validate:"omitempty"`
}

// Validate request.
func (req *GcpBatchCreateReq) Validate() error {
	if req.RequiredCount > constant.BatchCreateCvmFromCloudMaxLimit {
		return fmt.Errorf("required_count should <= %d", constant.BatchCreateCvmFromCloudMaxLimit)
	}

	return validator.Validate.Struct(req)
}
