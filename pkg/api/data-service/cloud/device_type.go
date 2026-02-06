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

package cloud

import (
	"hcm/pkg/api/core"
	coredevicetype "hcm/pkg/api/core/cloud/device-type"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	dt "hcm/pkg/dal/table/cloud/device-type"
	"hcm/pkg/thirdparty/cvmapi"
)

// DeviceTypeListReq list request
type DeviceTypeListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *DeviceTypeListReq) Validate() error {
	return r.ListReq.Validate()
}

// DistinctDeviceTypeListReq list distinct request
type DistinctDeviceTypeListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *DistinctDeviceTypeListReq) Validate() error {
	return r.ListReq.Validate()
}

// DeviceTypeTableListResult list device type table result
type DeviceTypeTableListResult types.ListResult[dt.DeviceTypeTable]

// DeviceTypeListResult list result
type DeviceTypeListResult types.ListResult[coredevicetype.DeviceType]

// DistinctDeviceTypeListResult distinct device type list result
type DistinctDeviceTypeListResult types.ListResult[coredevicetype.DistinctDeviceType]

// DeviceTypeCreate create device type request
type DeviceTypeCreate struct {
	// DeviceType 机型
	DeviceType string `json:"device_type" validate:"required,lte=64"`
	// DeviceClass 机型分类
	DeviceClass string `json:"device_class" validate:"required,lte=64"`
	// DeviceFamily 机型族
	DeviceFamily string `json:"device_family" validate:"required,lte=64"`
	// CoreType 核心类型
	CoreType enumor.CoreType `json:"core_type" validate:"required,lte=64"`
	// CpuCore CPU核心数，单位：核
	CpuCore int64 `json:"cpu_core" validate:"gte=0"`
	// Memory 内存大小，单位：GB
	Memory int64 `json:"memory" validate:"gte=0"`
	// DeviceTypeClass 通/专用机型，SpecialType专用，CommonType通用
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class" validate:"required,lte=64"`
	// TechnicalClass 技术分类
	TechnicalClass string `json:"technical_class" validate:"required,lte=64"`
	// Region 地域
	Region string `json:"region" validate:"required,lte=64"`
	// Zone 可用区
	Zone string `json:"zone" validate:"required,lte=64"`
	// Disable 是否不使用
	Disable bool `json:"disable"`
	// Source 机型来源
	Source enumor.DeviceTypeSource `json:"source"`
}

// Validate validate
func (r *DeviceTypeCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// DeviceTypeUpdate update device type request
type DeviceTypeUpdate struct {
	// ID 唯一ID
	ID string `json:"id" validate:"required,lte=64"`
	// DeviceType 机型
	DeviceType *string `json:"device_type,omitempty" validate:"omitempty,lte=64"`
	// DeviceClass 机型分类
	DeviceClass *string `json:"device_class,omitempty" validate:"omitempty,lte=64"`
	// DeviceFamily 机型族
	DeviceFamily *string `json:"device_family,omitempty" validate:"omitempty,lte=64"`
	// CoreType 核心类型
	CoreType *enumor.CoreType `json:"core_type,omitempty" validate:"omitempty,lte=64"`
	// CpuCore CPU核心数，单位：核
	CpuCore *int64 `json:"cpu_core,omitempty" validate:"omitempty,gte=0"`
	// Memory 内存大小，单位：GB
	Memory *int64 `json:"memory,omitempty" validate:"omitempty,gte=0"`
	// DeviceTypeClass 通/专用机型，SpecialType专用，CommonType通用
	DeviceTypeClass *cvmapi.InstanceTypeClass `json:"device_type_class,omitempty" validate:"omitempty,lte=64"`
	// TechnicalClass 技术分类
	TechnicalClass *string `json:"technical_class,omitempty" validate:"omitempty,lte=64"`
	// Region 地域
	Region *string `json:"region,omitempty" validate:"omitempty,lte=64"`
	// Zone 可用区
	Zone *string `json:"zone,omitempty" validate:"omitempty,lte=64"`
	// Disable 是否不使用
	Disable *bool `json:"disable,omitempty"`
	// Source 机型来源
	Source *enumor.DeviceTypeSource `json:"source,omitempty" validate:"omitempty,lte=64"`
}

// Validate validate
func (r *DeviceTypeUpdate) Validate() error {
	return validator.Validate.Struct(r)
}

// DeviceTypeBatchCreateReq create request
type DeviceTypeBatchCreateReq struct {
	DeviceTypes []DeviceTypeCreate `json:"device_types" validate:"required,min=1,max=100,dive"`
}

// Validate validate
func (r *DeviceTypeBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}
	for _, dt := range r.DeviceTypes {
		if err := dt.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DeviceTypeBatchUpdateReq batch update request
type DeviceTypeBatchUpdateReq struct {
	DeviceTypes []DeviceTypeUpdate `json:"device_types" validate:"required,min=1,max=100,dive"`
}

// Validate validate
func (r *DeviceTypeBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}
	for _, dt := range r.DeviceTypes {
		if err := dt.Validate(); err != nil {
			return err
		}
	}
	return nil
}
