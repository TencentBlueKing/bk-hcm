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

// AwsOperateSyncReq cvm oprate sync request.
type AwsOperateSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate cvm operate sync request.
func (req *AwsOperateSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	if len(req.CloudIDs) <= 0 {
		return fmt.Errorf("operate sync count should > 0")
	}

	return validator.Validate.Struct(req)
}

// AwsBatchDeleteReq define batch delete req.
type AwsBatchDeleteReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *AwsBatchDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// AwsBatchStartReq define batch start req.
type AwsBatchStartReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *AwsBatchStartReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// AwsBatchStopReq define batch stop req.
type AwsBatchStopReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
	Force     bool     `json:"force" validate:"required"`
	Hibernate bool     `json:"hibernate" validate:"omitempty"`
}

// Validate request.
func (req *AwsBatchStopReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// AwsBatchRebootReq define batch reboot req.
type AwsBatchRebootReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *AwsBatchRebootReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// AwsBatchCreateReq aws batch create req.
type AwsBatchCreateReq struct {
	DryRun                bool                            `json:"dry_run" validate:"omitempty"`
	AccountID             string                          `json:"account_id" validate:"required"`
	Region                string                          `json:"region" validate:"required"`
	Zone                  string                          `json:"zone" validate:"required"`
	Name                  string                          `json:"name" validate:"required"`
	InstanceType          string                          `json:"instance_type" validate:"required"`
	CloudImageID          string                          `json:"cloud_image_id" validate:"required"`
	CloudSubnetID         string                          `json:"cloud_subnet_id" validate:"required"`
	PublicIPAssigned      bool                            `json:"public_ip_assigned" validate:"omitempty"`
	CloudSecurityGroupIDs []string                        `json:"cloud_security_group_ids" validate:"required"`
	BlockDeviceMapping    []typecvm.AwsBlockDeviceMapping `json:"block_device_mapping" validate:"required"`
	Password              string                          `json:"password" validate:"required"`
	RequiredCount         int64                           `json:"required_count" validate:"required"`
	ClientToken           *string                         `json:"client_token" validate:"omitempty"`
}

// Validate request.
func (req *AwsBatchCreateReq) Validate() error {
	if req.RequiredCount > constant.BatchCreateCvmFromCloudMaxLimit {
		return fmt.Errorf("required_count should <= %d", constant.BatchCreateCvmFromCloudMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// AwsCvmBatchAssociateSecurityGroupReq aws batch associate security group req.
type AwsCvmBatchAssociateSecurityGroupReq struct {
	AccountID        string   `json:"account_id" validate:"required"`
	Region           string   `json:"region" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required"`
	CvmID            string   `json:"cvm_id" validate:"required"`
}

// Validate ...
func (req *AwsCvmBatchAssociateSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(req)
}
