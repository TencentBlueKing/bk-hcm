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

package devicecapacity

import (
	"hcm/pkg/api/core"
	devicecapacity "hcm/pkg/api/core/device-capacity"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// -------------------------- Create --------------------------

// CreateDeviceCapacityReq define create device capacity request.
type CreateDeviceCapacityReq struct {
	Items []CreateDeviceCapacityField `json:"items" validate:"required,min=1,max=100"`
}

// Validate CreateDeviceCapacityReq.
func (req CreateDeviceCapacityReq) Validate() error {
	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// CreateDeviceCapacityField define device capacity create field.
type CreateDeviceCapacityField struct {
	RequireType enumor.RequireType `json:"require_type" validate:"required"`
	Region      string             `json:"region" validate:"required"`
	Zone        string             `json:"zone" validate:"required"`
	DeviceType  string             `json:"device_type" validate:"required"`
	Capacity    *int64             `json:"capacity" validate:"required"`
	Extension   types.JsonField    `json:"extension"`
}

// Validate CreateDeviceCapacityField.
func (req CreateDeviceCapacityField) Validate() error {
	if err := req.RequireType.Validate(); err != nil {
		return err
	}
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateDeviceCapacityReq define update device capacity request.
type UpdateDeviceCapacityReq struct {
	Items []UpdateDeviceCapacityField `json:"items" validate:"required,min=1,max=100"`
}

// Validate UpdateDeviceCapacityReq.
func (req UpdateDeviceCapacityReq) Validate() error {
	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// UpdateDeviceCapacityField define device capacity update field.
type UpdateDeviceCapacityField struct {
	ID string `json:"id" validate:"required"`

	RequireType *enumor.RequireType `json:"require_type"`
	Region      *string             `json:"region"`
	Zone        *string             `json:"zone"`
	DeviceType  *string             `json:"device_type"`
	Capacity    *int64              `json:"capacity"`
	Extension   *types.JsonField    `json:"extension"`
}

// Validate UpdateDeviceCapacityField.
func (req UpdateDeviceCapacityField) Validate() error {
	if req.RequireType != nil {
		requireType := converter.PtrToVal(req.RequireType)
		if err := requireType.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListDeviceCapacityResult defines list result.
type ListDeviceCapacityResult = core.ListResultT[devicecapacity.DeviceCapacity]

// -------------------------- List With Device Info --------------------------

// ListCapacityWithDeviceInfoReq define list device capacity with device info request.
type ListCapacityWithDeviceInfoReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields"`
}

// Validate ListCapacityWithDeviceInfoReq.
func (req *ListCapacityWithDeviceInfoReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// ListCapacityWithDeviceInfoResult defines list with device info result.
type ListCapacityWithDeviceInfoResult = core.ListResultT[devicecapacity.CapacityWithDeviceInfo]

// -------------------------- Delete --------------------------

// DeleteDeviceCapacityReq device capacity delete request.
type DeleteDeviceCapacityReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate device capacity delete request.
func (req *DeleteDeviceCapacityReq) Validate() error {
	return validator.Validate.Struct(req)
}
