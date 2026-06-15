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
	"hcm/pkg/rest"
)

// TCloudBatchDeleteReq define batch delete req.
type TCloudBatchDeleteReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *TCloudBatchDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudBatchStartReq define batch start req.
type TCloudBatchStartReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
}

// Validate request.
func (req *TCloudBatchStartReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudBatchStopReq define batch stop req.
type TCloudBatchStopReq struct {
	AccountID   string              `json:"account_id" validate:"required"`
	Region      string              `json:"region" validate:"required"`
	IDs         []string            `json:"ids" validate:"required"`
	StopType    typecvm.StopType    `json:"stop_type" validate:"required"`
	StoppedMode typecvm.StoppedMode `json:"stopped_mode" validate:"required"`
}

// Validate request.
func (req *TCloudBatchStopReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudBatchRebootReq define batch reboot req.
type TCloudBatchRebootReq struct {
	AccountID string           `json:"account_id" validate:"required"`
	Region    string           `json:"region" validate:"required"`
	IDs       []string         `json:"ids" validate:"required"`
	StopType  typecvm.StopType `json:"stop_type" validate:"required"`
}

// Validate request.
func (req *TCloudBatchRebootReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudBatchResetPwdReq tcloud batch reset pwd req.
type TCloudBatchResetPwdReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	IDs       []string `json:"ids" validate:"required"`
	UserName  string   `json:"user_name" validate:"required"`
	Password  string   `json:"password" validate:"required"`
	ForceStop bool     `json:"force_stop" validate:"required"`
}

// Validate request.
func (req *TCloudBatchResetPwdReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("batch operation resource count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudCvmSpec 计费/规格字段（询价/创建共用，不含名称/密码/安全组等非计费字段）
type TCloudCvmSpec struct {
	DryRun                  bool                                 `json:"dry_run" validate:"omitempty"`
	InstanceType            string                               `json:"instance_type" validate:"required"`
	CloudImageID            string                               `json:"cloud_image_id" validate:"required"`
	RequiredCount           int64                                `json:"required_count" validate:"required"`
	CloudVpcID              string                               `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           string                               `json:"cloud_subnet_id" validate:"required"`
	InstanceChargeType      typecvm.TCloudInstanceChargeType     `json:"instance_charge_type" validate:"required"`
	InstanceChargePrepaid   *typecvm.TCloudInstanceChargePrepaid `json:"instance_charge_prepaid" validate:"omitempty"`
	SystemDisk              *typecvm.TCloudSystemDisk            `json:"system_disk" validate:"required"`
	DataDisk                []typecvm.TCloudDataDisk             `json:"data_disk" validate:"omitempty"`
	PublicIPAssigned        bool                                 `json:"public_ip_assigned" validate:"omitempty"`
	InternetMaxBandwidthOut int64                                `json:"internet_max_bandwidth_out" validate:"omitempty"`
	InternetChargeType      typecvm.TCloudInternetChargeType     `json:"internet_charge_type" validate:"omitempty"`
	BandwidthPackageID      *string                              `json:"bandwidth_package_id" validate:"omitempty"`
}

// ValidateSpec 校验规格/计费字段
func (spec *TCloudCvmSpec) ValidateSpec() error {
	if spec == nil {
		return fmt.Errorf("spec is required")
	}
	if spec.RequiredCount > constant.BatchCreateCvmFromCloudMaxLimit {
		return fmt.Errorf("required_count should <= %d", constant.BatchCreateCvmFromCloudMaxLimit)
	}
	return validator.Validate.Struct(spec)
}

// TCloudBatchCreateReq tcloud batch create req.
type TCloudBatchCreateReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
	Name      string `json:"name" validate:"required"`

	TCloudCvmSpec `json:",inline" validate:"required"`

	Password              string   `json:"password" validate:"required"`
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids" validate:"required"`
	ClientToken           *string  `json:"client_token" validate:"omitempty"`
}

// Validate request.
func (req *TCloudBatchCreateReq) Validate() error {
	if err := req.TCloudCvmSpec.ValidateSpec(); err != nil {
		return err
	}
	return validator.Validate.Struct(req)
}

// TCloudCvmInquiryReq tcloud cvm inquiry req
type TCloudCvmInquiryReq struct {
	AccountID     string `json:"account_id" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Zone          string `json:"zone" validate:"required"`
	TCloudCvmSpec `json:",inline" validate:"required"`
}

// Validate inquiry request.
func (req *TCloudCvmInquiryReq) Validate() error {
	if err := req.TCloudCvmSpec.ValidateSpec(); err != nil {
		return err
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

// BatchCreateResp ...
type BatchCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *BatchCreateResult `json:"data"`
}

// TCloudBatchResetReq defines options to reset cvm request.
type TCloudBatchResetReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
	ImageID   string   `json:"image_id" validate:"required"`
	Password  string   `json:"password" validate:"required,min=12,max=30"`
}

// Validate reset cvm request.
func (opt TCloudBatchResetReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudCvmBatchAssociateSecurityGroupReq defines options to associate security group to cvm request.
type TCloudCvmBatchAssociateSecurityGroupReq struct {
	AccountID        string   `json:"account_id" validate:"required"`
	Region           string   `json:"region" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required"`
	CvmID            string   `json:"cvm_id" validate:"required"`
}

// Validate associate security group to cvm request.
func (opt TCloudCvmBatchAssociateSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudInstanceConfigListOption ...
type TCloudInstanceConfigListOption struct {
	AccountID                              string `json:"account_id" validate:"required"`
	typecvm.TCloudInstanceConfigListOption `json:",inline"`
}

// Validate instance config list option.
func (opt *TCloudInstanceConfigListOption) Validate() error {
	if err := opt.TCloudInstanceConfigListOption.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(opt)
}

// -------------------------- Monitor --------------------------

// TCloudMonitorDataReq defines request to get tcloud monitor data.
type TCloudMonitorDataReq struct {
	AccountID   string   `json:"account_id" validate:"required"`
	Region      string   `json:"region" validate:"required"`
	MetricName  string   `json:"metric_name" validate:"required"`
	Period      int64    `json:"period" validate:"required,min=60"`
	StartTime   string   `json:"start_time" validate:"required"` // DateTimeLayout format: 2006-01-02 15:04:05
	EndTime     string   `json:"end_time" validate:"required"`   // DateTimeLayout format: 2006-01-02 15:04:05
	InstanceIDs []string `json:"instance_ids" validate:"required,min=1"`
}

// Validate request.
func (req *TCloudMonitorDataReq) Validate() error {
	if len(req.InstanceIDs) > constant.MonitorMaxInstanceLimit {
		return fmt.Errorf("instances count should <= %d", constant.MonitorMaxInstanceLimit)
	}

	return validator.Validate.Struct(req)
}

// TCloudMonitorDataResp defines response of tcloud monitor data.
type TCloudMonitorDataResp struct {
	DataPoints []*MonitorDataPointResp `json:"data_points"`
}

// MonitorDataPointResp defines a single monitor data point response.
type MonitorDataPointResp struct {
	Dimensions []*MonitorDimensionResp `json:"dimensions"`
	Timestamps []int64                 `json:"timestamps"`
	Values     []float64               `json:"values"`
	Extensions map[string]interface{}  `json:"extensions,omitempty"`
}

// MonitorDimensionResp defines monitor dimension response.
type MonitorDimensionResp struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
