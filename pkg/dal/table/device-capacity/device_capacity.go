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

// Package devicecapacity ...
package devicecapacity

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// DeviceCapacityColumns defines all the device_capacity table's columns.
var DeviceCapacityColumns = utils.MergeColumns(nil, DeviceCapacityColumnDescriptor)

// DeviceCapacityColumnDescriptor is device_capacity column descriptors.
var DeviceCapacityColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "require_type", NamedC: "require_type", Type: enumor.Numeric},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "capacity", NamedC: "capacity", Type: enumor.Numeric},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DeviceCapacityTable is used to save device capacity information.
type DeviceCapacityTable struct {
	ID          string             `db:"id" json:"id" validate:"max=64"`
	RequireType enumor.RequireType `db:"require_type" json:"require_type"`
	Region      string             `db:"region" json:"region" validate:"max=64"`
	Zone        string             `db:"zone" json:"zone" validate:"max=64"`
	DeviceType  string             `db:"device_type" json:"device_type" validate:"max=64"`
	Capacity    *int64             `db:"capacity" json:"capacity"`
	Extension   types.JsonField    `db:"extension" json:"extension"`
	Creator     string             `db:"creator" validate:"max=64" json:"creator"`
	Reviser     string             `db:"reviser" validate:"max=64" json:"reviser"`
	CreatedAt   types.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   types.Time         `db:"updated_at" json:"updated_at"`
}

// TableName is the device_capacity database table name.
func (d DeviceCapacityTable) TableName() table.Name {
	return table.DeviceCapacityTable
}

// InsertValidate validate device capacity on insertion.
func (d DeviceCapacityTable) InsertValidate() error {
	if len(d.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(d.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if err := validator.Validate.Struct(d); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate device capacity on update.
func (d DeviceCapacityTable) UpdateValidate() error {
	if len(d.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(d.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if err := validator.Validate.Struct(d); err != nil {
		return err
	}
	return nil
}
