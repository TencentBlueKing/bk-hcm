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

package devicetype

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/thirdparty/cvmapi"
)

// DeviceTypeColumns defines all the device type table's columns.
var DeviceTypeColumns = utils.MergeColumns(nil, DeviceTypeColumnDescriptor)

// DeviceTypeColumnDescriptor is DeviceTypeTable's column descriptors.
var DeviceTypeColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "device_type_class", NamedC: "device_type_class", Type: enumor.String},
	{Column: "device_class", NamedC: "device_class", Type: enumor.String},
	{Column: "device_family", NamedC: "device_family", Type: enumor.String},
	{Column: "core_type", NamedC: "core_type", Type: enumor.String},
	{Column: "cpu_core", NamedC: "cpu_core", Type: enumor.Numeric},
	{Column: "memory", NamedC: "memory", Type: enumor.Numeric},
	{Column: "technical_class", NamedC: "technical_class", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "disable", NamedC: "disable", Type: enumor.Boolean},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DeviceTypeTable is used to save resource's device type information.
type DeviceTypeTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// DeviceType 机型
	DeviceType string `db:"device_type" json:"device_type" validate:"lte=64"`
	// DeviceClass 机型分类
	DeviceClass string `db:"device_class" json:"device_class" validate:"lte=64"`
	// DeviceFamily 机型族
	DeviceFamily string `db:"device_family" json:"device_family" validate:"lte=64"`
	// CoreType 核心类型
	CoreType enumor.CoreType `db:"core_type" json:"core_type" validate:"lte=64"`
	// CpuCore CPU核心数，单位：核
	CpuCore int64 `db:"cpu_core" json:"cpu_core"`
	// Memory 内存大小，单位：GB
	Memory int64 `db:"memory" json:"memory"`
	// DeviceTypeClass 通/专用机型，SpecialType专用，CommonType通用
	DeviceTypeClass cvmapi.InstanceTypeClass `db:"device_type_class" json:"device_type_class" validate:"lte=64"`
	// TechnicalClass 技术分类
	TechnicalClass string `db:"technical_class" json:"technical_class" validate:"lte=64"`
	// Region 地域
	Region string `db:"region" json:"region" validate:"lte=64"`
	// Zone 可用区
	Zone string `db:"zone" json:"zone" validate:"lte=64"`
	// Disable 是否不使用
	Disable *bool `db:"disable" json:"disable"`
	// Source 机型来源
	Source enumor.DeviceTypeSource `db:"source" json:"source" validate:"lte=64"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser 修改人
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the device type's database table name.
func (t DeviceTypeTable) TableName() table.Name {
	return table.DeviceTypeTable
}

// InsertValidate validate device type on insertion.
func (t DeviceTypeTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(t.DeviceType) == 0 {
		return errors.New("device type can not be empty")
	}

	if len(t.DeviceClass) == 0 {
		return errors.New("device class can not be empty")
	}

	if len(t.DeviceFamily) == 0 {
		return errors.New("device family can not be empty")
	}

	if err := t.CoreType.Validate(); err != nil {
		return err
	}

	if len(t.Region) == 0 {
		return errors.New("region can not be empty")
	}

	if len(t.Zone) == 0 {
		return errors.New("zone can not be empty")
	}

	if t.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if t.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	if err := t.Source.Validate(); err != nil {
		return err
	}

	if len(t.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate device type on update.
func (t DeviceTypeTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if t.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if t.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	return nil
}
