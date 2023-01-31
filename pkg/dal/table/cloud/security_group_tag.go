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
	"hcm/pkg/dal/table/utils"
)

// SecurityGroupTagColumns defines all the security group tag table's columns.
var SecurityGroupTagColumns = utils.MergeColumns(nil, SecurityGroupTagColumnDescriptor)

// SecurityGroupTagColumnDescriptor is Security Group Tag's column descriptors.
var SecurityGroupTagColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "key", NamedC: "key", Type: enumor.String},
	{Column: "value", NamedC: "value", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// SecurityGroupTagTable define security group tag table.
type SecurityGroupTagTable struct {
	ID              string     `db:"id" validate:"lte=64"`
	SecurityGroupID string     `db:"security_group_id" validate:"lte=64"`
	Key             string     `db:"key" validate:"lte=255"`
	Value           string     `db:"value" validate:"lte=255"`
	Creator         string     `db:"creator" validate:"lte=64"`
	Reviser         string     `db:"reviser" validate:"lte=64"`
	CreatedAt       *time.Time `db:"created_at" validate:"excluded_unless"`
	UpdatedAt       *time.Time `db:"updated_at" validate:"excluded_unless"`
}

// TableName return security group tag table name.
func (t SecurityGroupTagTable) TableName() table.Name {
	return table.SecurityGroupTagTable
}

// InsertValidate security group tag table when insert.
func (t SecurityGroupTagTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.SecurityGroupID) == 0 {
		return errors.New("security_group_id is required")
	}

	if len(t.Key) == 0 {
		return errors.New("key is required")
	}

	if len(t.Value) == 0 {
		return errors.New("value is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if t.CreatedAt != nil {
		return errors.New("created_at can not set")
	}

	if t.UpdatedAt != nil {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate security group tag table when update.
func (t SecurityGroupTagTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
