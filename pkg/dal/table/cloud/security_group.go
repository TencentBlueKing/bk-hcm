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
	"errors"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// SecurityGroupColumns defines all the security group table's columns.
var SecurityGroupColumns = utils.MergeColumns(nil, SecurityGroupColumnDescriptor)

// SecurityGroupColumnDescriptor is Security Group's column descriptors.
var SecurityGroupColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "assigned", NamedC: "assigned", Type: enumor.Boolean},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// SecurityGroupTable define security group table.
type SecurityGroupTable struct {
	ID        string          `db:"id" validate:"lte=64"`
	Vendor    string          `db:"vendor" validate:"lte=16"`
	CloudID   string          `db:"cloud_id" validate:"lte=255"`
	Assigned  bool            `db:"assigned"`
	Region    string          `db:"region" validate:"lte=20"`
	Name      string          `db:"name" validate:"lte=60"`
	Memo      *string         `db:"memo" validate:"lte=omitempty,255"`
	AccountID string          `db:"account_id" validate:"lte=64"`
	Extension types.JsonField `db:"extension"`
	Creator   string          `db:"creator" validate:"lte=64"`
	Reviser   string          `db:"reviser" validate:"lte=64"`
	CreatedAt *time.Time      `db:"created_at" validate:"excluded_unless"`
	UpdatedAt *time.Time      `db:"updated_at" validate:"excluded_unless"`
}

// TableName return security group table name.
func (t SecurityGroupTable) TableName() table.Name {
	return table.SecurityGroupTable
}

// InsertValidate security group table when insert.
func (t SecurityGroupTable) InsertValidate() error {
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

	if len(t.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(t.Region) == 0 {
		return errors.New("region is required")
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

// UpdateValidate security group table when update.
func (t SecurityGroupTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
