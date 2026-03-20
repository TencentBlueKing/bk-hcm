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

// Package plan ...
package plan

import (
	"encoding/json"
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	gpusuborder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	cvt "hcm/pkg/tools/converter"
)

// UpdateResPlanDemandGpuSubOrderItem is GPU demand suborder update item.
type UpdateResPlanDemandGpuSubOrderItem struct {
	SubOrderID  string                           `json:"suborder_id" validate:"required"`
	Status      enumor.RPDemandGPUSubOrderStatus `json:"status"`
	Comment     []json.RawMessage                `json:"comment"`
	DemandYear  *int64                           `json:"demand_year"`
	DemandMonth *int64                           `json:"demand_month"`
	GPUNum      *int64                           `json:"gpu_num"`
	QpmMax      *int64                           `json:"qpm_max"`
	Extension   []json.RawMessage                `json:"extension"`
}

// BatchUpdateResPlanDemandGpuSubOrderReq is GPU demand suborder batch update request.
type BatchUpdateResPlanDemandGpuSubOrderReq struct {
	SubOrderData []UpdateResPlanDemandGpuSubOrderItem `json:"suborder_data" validate:"required,min=1,max=100,dive"`
}

// ValidateBiz validate biz update request.
func (r *BatchUpdateResPlanDemandGpuSubOrderReq) ValidateBiz() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.SubOrderData {
		if err := item.ValidateBiz(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateResource validate resource update request.
func (r *BatchUpdateResPlanDemandGpuSubOrderReq) ValidateResource() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.SubOrderData {
		if err := item.ValidateResource(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBiz validate biz update item.
func (i *UpdateResPlanDemandGpuSubOrderItem) ValidateBiz() error {
	if err := validator.Validate.Struct(i); err != nil {
		return err
	}

	if i.Extension == nil {
		return errors.New("extension is required")
	}

	if i.Status != "" || i.Comment != nil {
		return errors.New("biz update only supports extension")
	}

	if err := i.validateDemandFields(); err != nil {
		return err
	}

	return nil
}

// validateDemandFields 校验需求年份、月份、GPU卡数、峰值QPM 字段。
func (i *UpdateResPlanDemandGpuSubOrderItem) validateDemandFields() error {
	if i.DemandYear != nil && cvt.PtrToVal(i.DemandYear) < 0 {
		return errors.New("demand_year should be >= 0")
	}

	if i.DemandMonth != nil && (cvt.PtrToVal(i.DemandMonth) < 0 || cvt.PtrToVal(i.DemandMonth) > 12) {
		return errors.New("demand_month should be >= 0 and <= 12")
	}

	if i.GPUNum != nil && cvt.PtrToVal(i.GPUNum) < 0 {
		return errors.New("gpu_num should be >= 0")
	}

	if i.QpmMax != nil && cvt.PtrToVal(i.QpmMax) < 0 {
		return errors.New("qpm_max should be >= 0")
	}

	return nil
}

// ValidateResource validate resource update item.
func (i *UpdateResPlanDemandGpuSubOrderItem) ValidateResource() error {
	if err := validator.Validate.Struct(i); err != nil {
		return err
	}

	hasExtension := i.Extension != nil
	hasStatus := i.Status != ""
	hasComment := i.Comment != nil

	if hasExtension && (hasStatus || hasComment) {
		return errors.New("extension cannot be used with status or comment")
	}

	switch {
	case hasExtension:
		return i.validateDemandFields()
	case hasStatus:
		if i.Status.Validate() != nil {
			return errors.New("status is invalid")
		}
		if i.Status != enumor.RPDemandGPUSubOrderStatusDone && i.Status != enumor.RPDemandGPUSubOrderStatusReject {
			return errors.New("status should be DONE or REJECT")
		}
		return nil
	default:
		return errors.New("invalid suborder update mode")
	}
}

// BatchTerminateResPlanDemandGpuSubOrderReq is batch terminate GPU demand suborders request.
type BatchTerminateResPlanDemandGpuSubOrderReq struct {
	SubOrderIDs []string `json:"suborder_ids" validate:"required,min=1,max=100,dive,required"`
}

// Validate validate.
func (r *BatchTerminateResPlanDemandGpuSubOrderReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResPlanDemandGpuTplConfig is GPU demand template config response item.
type ResPlanDemandGpuTplConfig struct {
	ID     string          `json:"id"`
	Sheets json.RawMessage `json:"sheets"`
}

// ListResPlanDemandGpuSubOrderResp is list GPU demand suborder response.
type ListResPlanDemandGpuSubOrderResp struct {
	Count     uint64                                      `json:"count"`
	Details   []gpusuborder.ResPlanDemandGpuSubOrderTable `json:"details"`
	TplConfig []ResPlanDemandGpuTplConfig                 `json:"tpl_config"`
}
