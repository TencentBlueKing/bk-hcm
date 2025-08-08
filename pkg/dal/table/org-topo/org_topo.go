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

// Package orgtopo org topo
package orgtopo

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// OrgTopoColumns defines all the org_topo table's columns.
var OrgTopoColumns = utils.MergeColumns(nil, OrgTopoColumnsDescriptor)

// OrgTopoColumnsDescriptor is OrgTopo's column descriptors.
var OrgTopoColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "dept_id", NamedC: "dept_id", Type: enumor.String},
	{Column: "dept_name", NamedC: "dept_name", Type: enumor.String},
	{Column: "full_name", NamedC: "full_name", Type: enumor.String},
	{Column: "level", NamedC: "level", Type: enumor.Numeric},
	{Column: "parent", NamedC: "parent", Type: enumor.String},
	{Column: "has_children", NamedC: "has_children", Type: enumor.Numeric},

	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// OrgTopo defines the OrgTopo database table.
type OrgTopo struct {
	// ID is the unique identity id.
	ID string `db:"id" json:"id" validate:"max=16"`
	// DeptID 部门ID
	DeptID string `db:"dept_id" json:"dept_id" validate:"max=64"`
	// DeptName 部门名称
	DeptName string `db:"dept_name" json:"dept_name" validate:"max=64"`
	// FullName 部门完整名称
	FullName string `db:"full_name" json:"full_name" validate:"max=256"`
	// Level 部门所属等级
	Level int64 `db:"level" json:"level"`
	// Parent 上级部门ID
	Parent string `db:"parent" json:"parent" validate:"max=64"`
	// HasChildren 是否有下级部门(0:否1:有)
	HasChildren *int64 `db:"has_children" json:"has_children"`

	// Memo is used to set the memo information if needed.
	Memo *string `db:"memo" json:"memo"`
	// Creator is the one who create this resource.
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser is the one who revise this resource.
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt is the time when the resource is created.
	CreatedAt types.Time `db:"created_at" json:"created_at" validate:"excluded_unless"`
	// UpdatedAt is the time when the resource update this resource.
	UpdatedAt types.Time `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// TableName return the OrgTopo table's name
func (ot OrgTopo) TableName() table.Name {
	return table.OrgTopoTable
}

// ValidateInsert validate the inserted data is valid or not
func (ot OrgTopo) ValidateInsert() error {
	if err := validator.Validate.Struct(ot); err != nil {
		return err
	}

	if len(ot.ID) != 0 {
		return errors.New("invalid id, can not be set")
	}

	if len(ot.DeptID) == 0 {
		return errors.New("invalid dept_id, can not be empty")
	}

	if len(ot.DeptName) <= 0 {
		return errors.New("invalid dept_name, can not be empty")
	}

	if ot.Level < 0 {
		return errors.New("invalid level,  should be >= 0")
	}

	if len(ot.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(ot.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// ValidateUpdate validate if the to be updated data is valid or not.
func (ot OrgTopo) ValidateUpdate() error {
	if err := validator.Validate.Struct(ot); err != nil {
		return err
	}

	if len(ot.ID) == 0 {
		return errors.New("id is required, but not set")
	}

	if len(ot.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if len(ot.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}
