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

package tcloudziyanpmdevicetype

import (
	"hcm/pkg/api/core"
	pmdevicetype "hcm/pkg/api/core/tcloud-ziyan-pm-device-type"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CreateTCloudZiyanPmDeviceTypeReq define create tcloud ziyan pm device type request.
type CreateTCloudZiyanPmDeviceTypeReq struct {
	Items []CreateTCloudZiyanPmDeviceTypeField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate CreateTCloudZiyanPmDeviceTypeReq.
func (req CreateTCloudZiyanPmDeviceTypeReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Items) > constant.BatchOperationMaxLimit {
		return errf.Newf(errf.InvalidParameter, "items length must be less than %d", constant.BatchOperationMaxLimit)
	}

	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CreateTCloudZiyanPmDeviceTypeField define tcloud ziyan pm device type create field.
type CreateTCloudZiyanPmDeviceTypeField struct {
	DeviceType string `json:"device_type" validate:"required"`
	Raid       string `json:"raid" validate:"required"`
	CpuCore    int    `json:"cpu_core" validate:"required"`
	Memory     int    `json:"memory" validate:"required"`
	Disable    bool   `json:"disable"`
}

// Validate CreateTCloudZiyanPmDeviceTypeField.
func (req CreateTCloudZiyanPmDeviceTypeField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateTCloudZiyanPmDeviceTypeReq define update tcloud ziyan pm device type request.
type UpdateTCloudZiyanPmDeviceTypeReq struct {
	Items []UpdateTCloudZiyanPmDeviceTypeField `json:"items" validate:"required,min=1,max=100,dive,required"`
}

// Validate UpdateTCloudZiyanPmDeviceTypeReq.
func (req UpdateTCloudZiyanPmDeviceTypeReq) Validate() error {
	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// UpdateTCloudZiyanPmDeviceTypeField define tcloud ziyan pm device type update field.
type UpdateTCloudZiyanPmDeviceTypeField struct {
	ID string `json:"id" validate:"required"`

	DeviceType *string `json:"device_type"`
	Raid       *string `json:"raid"`
	CpuCore    *int    `json:"cpu_core"`
	Memory     *int    `json:"memory"`
	Disable    *bool   `json:"disable"`
}

// Validate UpdateTCloudZiyanPmDeviceTypeField.
func (req UpdateTCloudZiyanPmDeviceTypeField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListTCloudZiyanPmDeviceTypeResult defines list result.
type ListTCloudZiyanPmDeviceTypeResult = core.ListResultT[pmdevicetype.TCloudZiyanPmDeviceType]

// -------------------------- Delete --------------------------

// DeleteTCloudZiyanPmDeviceTypeReq tcloud ziyan pm device type delete request.
type DeleteTCloudZiyanPmDeviceTypeReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate tcloud ziyan pm device type delete request.
func (req *DeleteTCloudZiyanPmDeviceTypeReq) Validate() error {
	return validator.Validate.Struct(req)
}
