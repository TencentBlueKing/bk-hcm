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

// HuaWeiOperateSyncReq cvm oprate sync request.
type HuaWeiOperateSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate cvm operate sync request.
func (req *HuaWeiOperateSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	if len(req.CloudIDs) <= 0 {
		return fmt.Errorf("operate sync count should > 0")
	}

	return validator.Validate.Struct(req)
}

// HuaWeiBatchDeleteReq define batch delete req.
type HuaWeiBatchDeleteReq struct {
	AccountID      string   `json:"account_id" validate:"required"`
	Region         string   `json:"region" validate:"required"`
	IDs            []string `json:"ids" validate:"required"`
	DeletePublicIP bool     `json:"delete_public_ip" validate:"required"`
	DeleteDisk     bool     `json:"delete_disk" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HuaWeiBatchStartReq define batch start req.
type HuaWeiBatchStartReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchStartReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HuaWeiBatchStopReq define batch stop req.
type HuaWeiBatchStopReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
	Force     bool     `json:"force" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchStopReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HuaWeiBatchRebootReq define batch reboot req.
type HuaWeiBatchRebootReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
	Force     bool     `json:"force" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchRebootReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HuaWeiBatchResetPwdReq tcloud batch reset pwd req.
type HuaWeiBatchResetPwdReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
	Password  string   `json:"password" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchResetPwdReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// HuaWeiCvmSpec 计费/规格字段（询价/创建共用）
type HuaWeiCvmSpec struct {
	DryRun           bool                          `json:"dry_run" validate:"omitempty"`
	InstanceType     string                        `json:"instance_type" validate:"required"`
	CloudImageID     string                        `json:"cloud_image_id" validate:"required"`
	RequiredCount    int32                         `json:"required_count" validate:"required"`
	Description      *string                       `json:"description" validate:"omitempty"`
	RootVolume       *typecvm.HuaWeiVolume         `json:"root_volume" validate:"required"`
	DataVolume       []typecvm.HuaWeiVolume        `json:"data_volume" validate:"omitempty"`
	InstanceCharge   *typecvm.HuaWeiInstanceCharge `json:"instance_charge" validate:"required"`
	PublicIPAssigned bool                          `json:"public_ip_assigned" validate:"omitempty"`
	Eip              *typecvm.HuaWeiEip            `json:"eip" validate:"omitempty"`
}

// ValidateSpec 校验规格/计费字段
func (spec *HuaWeiCvmSpec) ValidateSpec() error {
	if spec == nil {
		return fmt.Errorf("spec is required")
	}
	if spec.RequiredCount > constant.BatchCreateCvmFromCloudMaxLimit {
		return fmt.Errorf("required_count should <= %d", constant.BatchCreateCvmFromCloudMaxLimit)
	}
	return validator.Validate.Struct(spec)
}

// HuaWeiBatchCreateReq batch create req.
type HuaWeiBatchCreateReq struct {
	AccountID             string   `json:"account_id" validate:"required"`
	Region                string   `json:"region" validate:"required"`
	Zone                  string   `json:"zone" validate:"required"`
	Name                  string   `json:"name" validate:"required"`
	Password              string   `json:"password" validate:"required"`
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids" validate:"required"`
	ClientToken           *string  `json:"client_token" validate:"omitempty"`
	CloudVpcID            string   `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID         string   `json:"cloud_subnet_id" validate:"required"`
	HuaWeiCvmSpec         `json:",inline" validate:"required"`
}

// Validate request.
func (req *HuaWeiBatchCreateReq) Validate() error {
	if err := req.HuaWeiCvmSpec.ValidateSpec(); err != nil {
		return err
	}
	if req.CloudVpcID == "" {
		return fmt.Errorf("cloud_vpc_id is required")
	}
	if req.CloudSubnetID == "" {
		return fmt.Errorf("cloud_subnet_id is required")
	}
	return validator.Validate.Struct(req)
}

// HuaWeiCvmInquiryReq huawei cvm inquiry req
type HuaWeiCvmInquiryReq struct {
	AccountID     string `json:"account_id" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Zone          string `json:"zone" validate:"required"`
	HuaWeiCvmSpec `json:",inline" validate:"required"`
}

// Validate inquiry request.
func (req *HuaWeiCvmInquiryReq) Validate() error {
	if err := req.HuaWeiCvmSpec.ValidateSpec(); err != nil {
		return err
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Monitor --------------------------

// HuaWeiMonitorDataReq defines request to get huawei monitor data.
type HuaWeiMonitorDataReq struct {
	AccountID   string   `json:"account_id" validate:"required"`
	Region      string   `json:"region" validate:"required"`
	MetricName  string   `json:"metric_name" validate:"required"`
	Period      int64    `json:"period" validate:"required,min=1"`
	StartTime   int64    `json:"start_time" validate:"required,min=1"`
	EndTime     int64    `json:"end_time" validate:"required,min=1"`
	Namespace   string   `json:"namespace" validate:"omitempty"`
	Filter      string   `json:"filter" validate:"omitempty"`
	Dimension   string   `json:"dimension" validate:"omitempty"`
	InstanceIDs []string `json:"instance_ids" validate:"required,min=1"`
}

// Validate request.
func (req *HuaWeiMonitorDataReq) Validate() error {
	if req.StartTime >= req.EndTime {
		return fmt.Errorf("start_time should < end_time")
	}

	return validator.Validate.Struct(req)
}

// HuaWeiMonitorDataResp defines response of huawei monitor data.
type HuaWeiMonitorDataResp struct {
	DataPoints []*MonitorDataPointResp `json:"data_points"`
}
