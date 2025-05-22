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

// Package zone ...
package zone

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ZoneColumns defines all the zone table's columns.
var ZoneColumns = utils.MergeColumns(nil, ZoneTableColumnDescriptor)

// ZoneTableColumnDescriptor is zone's column descriptors.
var ZoneTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "name_cn", NamedC: "name_cn", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ZoneTable define zone table.
type ZoneTable struct {
	ID        string          `db:"id" json:"id" validate:"lte=64"`
	Vendor    enumor.Vendor   `db:"vendor" json:"vendor"`
	CloudID   string          `db:"cloud_id" json:"cloud_id" validate:"lte=255"`
	Name      string          `db:"name" json:"name" validate:"lte=64"`
	Region    string          `db:"region" json:"region" validate:"lte=64"`
	NameCn    string          `db:"name_cn" json:"name_cn" validate:"lte=64"`
	State     string          `db:"state" json:"state" validate:"lte=64"`
	Extension types.JsonField `db:"extension" json:"extension"`
	Creator   string          `db:"creator" json:"creator" validate:"lte=64"`
	Reviser   string          `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt types.Time      `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt types.Time      `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return zone table name.
func (z ZoneTable) TableName() table.Name {
	return table.ZoneTable
}

// InsertValidate zone table when insert.
func (t ZoneTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.State) == 0 {
		return errors.New("state is required")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.Extension) == 0 {
		return errors.New("extension is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate zone table when update.
func (t ZoneTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
