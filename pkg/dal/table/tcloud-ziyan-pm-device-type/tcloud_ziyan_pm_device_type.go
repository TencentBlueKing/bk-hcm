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

// Package tcloudziyanpmdevicetype ...
package tcloudziyanpmdevicetype

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// TCloudZiyanPmDeviceTypeColumns defines all the tcloud_ziyan_pm_device_type table's columns.
var TCloudZiyanPmDeviceTypeColumns = utils.MergeColumns(nil, TCloudZiyanPmDeviceTypeColumnDescriptor)

// TCloudZiyanPmDeviceTypeColumnDescriptor is tcloud_ziyan_pm_device_type column descriptors.
var TCloudZiyanPmDeviceTypeColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "raid", NamedC: "raid", Type: enumor.String},
	{Column: "cpu_core", NamedC: "cpu_core", Type: enumor.Numeric},
	{Column: "memory", NamedC: "memory", Type: enumor.Numeric},
	{Column: "disable", NamedC: "disable", Type: enumor.Boolean},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TCloudZiyanPmDeviceTypeTable is used to save tcloud ziyan pm device type information.
type TCloudZiyanPmDeviceTypeTable struct {
	ID         string     `db:"id" json:"id" validate:"max=64"`
	DeviceType string     `db:"device_type" json:"device_type" validate:"max=64"`
	Raid       string     `db:"raid" json:"raid" validate:"max=64"`
	CpuCore    int        `db:"cpu_core" json:"cpu_core"`
	Memory     int        `db:"memory" json:"memory"`
	Disable    bool       `db:"disable" json:"disable"`
	Creator    string     `db:"creator" json:"creator" validate:"max=64"`
	Reviser    string     `db:"reviser" json:"reviser" validate:"max=64"`
	CreatedAt  types.Time `db:"created_at" json:"created_at"`
	UpdatedAt  types.Time `db:"updated_at" json:"updated_at"`
}

// TableName is the tcloud_ziyan_pm_device_type database table name.
func (t TCloudZiyanPmDeviceTypeTable) TableName() table.Name {
	return table.TCloudZiyanPmDeviceTypeTable
}

// InsertValidate validate tcloud ziyan pm device type on insertion.
func (t TCloudZiyanPmDeviceTypeTable) InsertValidate() error {
	if len(t.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(t.DeviceType) == 0 {
		return errors.New("device_type can not be empty")
	}

	if len(t.Raid) == 0 {
		return errors.New("raid can not be empty")
	}

	if t.CpuCore < 0 {
		return errors.New("cpu_core should be >= 0")
	}

	if t.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if err := validator.Validate.Struct(t); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate tcloud ziyan pm device type on update.
func (t TCloudZiyanPmDeviceTypeTable) UpdateValidate() error {
	if len(t.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if t.CpuCore < 0 {
		return errors.New("cpu_core should be >= 0")
	}

	if t.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	if err := validator.Validate.Struct(t); err != nil {
		return err
	}
	return nil
}
