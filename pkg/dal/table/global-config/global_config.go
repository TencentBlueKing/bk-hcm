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

// Package tablegconf global config table
package tablegconf

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// GlobalConfigTableColumns defines all the cvm table's columns.
var GlobalConfigTableColumns = utils.MergeColumns(nil, GlobalConfigTableColumnDescriptors)

// GlobalConfigTableColumnDescriptors is cvm table column descriptors.
var GlobalConfigTableColumnDescriptors = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "config_key", NamedC: "config_key", Type: enumor.String},
	{Column: "config_value", NamedC: "config_value", Type: enumor.Json},
	{Column: "config_type", NamedC: "config_type", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.String},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.String},
}

// GlobalConfigTable define cvm table.
type GlobalConfigTable struct {
	// ID global config id
	ID string `db:"id" json:"id"`
	// ConfigKey global config key, key+type is unique
	ConfigKey string `db:"config_key" json:"config_key"`
	// ConfigValue global config value, json format
	ConfigValue types.JsonField `db:"config_value" json:"config_value"`
	// ConfigType global config type, enum: string/int/float/bool/map/slice
	ConfigType string `db:"config_type" json:"config_type"`
	// Memo global config memo
	Memo *string `db:"memo" json:"memo"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// UniqueKey returns the unique key of BizOrgRel table.
func (t GlobalConfigTable) UniqueKey() string {
	return fmt.Sprintf("(%s,%s)", t.ConfigType, t.ConfigKey)
}

// Columns return cvm table columns.
func (t GlobalConfigTable) Columns() *utils.Columns {
	return GlobalConfigTableColumns
}

// ColumnDescriptors define cvm table column descriptor.
func (t GlobalConfigTable) ColumnDescriptors() utils.ColumnDescriptors {
	return GlobalConfigTableColumnDescriptors
}

// TableName return cvm table name.
func (t GlobalConfigTable) TableName() table.Name {
	return table.GlobalConfigTable
}

// InsertValidate cvm table when insert.
func (t GlobalConfigTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.ConfigKey) == 0 {
		return errors.New("key is required")
	}

	if len(t.ConfigValue) == 0 {
		return errors.New("value is required")
	}

	if len(t.ConfigType) == 0 {
		return errors.New("type is required")
	}

	if err := validator.ValidateMemo(t.Memo, false); err != nil {
		return err
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate cvm table when update.
func (t GlobalConfigTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}
	if len(t.ID) != 0 {
		return errors.New("id can not be updated")
	}

	// key+type 组成一个全局唯一键，所以不能更新
	if len(t.ConfigKey) != 0 {
		return errors.New("key can not be updated")
	}

	if len(t.ConfigType) != 0 {
		return errors.New("type can not be updated")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}
