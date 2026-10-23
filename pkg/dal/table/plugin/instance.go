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

package tableplugin

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// IacPluginInstanceTableColumns iac_plugin_instance 表全部列。
var IacPluginInstanceTableColumns = utils.MergeColumns(nil, IacPluginInstanceTableColumnDescriptors)

// IacPluginInstanceTableColumnDescriptors iac_plugin_instance 表列描述符。
var IacPluginInstanceTableColumnDescriptors = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "plugin_name", NamedC: "plugin_name", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "attrs", NamedC: "attrs", Type: enumor.Json},
	{Column: "secret_attrs", NamedC: "secret_attrs", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// IacPluginInstanceTable iac_plugin_instance 表模型。
type IacPluginInstanceTable struct {
	ID          string          `db:"id" json:"id" validate:"lte=64"`
	PluginName  string          `db:"plugin_name" json:"plugin_name" validate:"lte=64"`
	Name        string          `db:"name" json:"name" validate:"lte=128"`
	Attrs       types.JsonField `db:"attrs" json:"attrs"`
	SecretAttrs types.JsonField `db:"secret_attrs" json:"secret_attrs"`
	Extension   types.JsonField `db:"extension" json:"extension"`
	Memo        *string         `db:"memo" json:"memo"`
	TenantID    string          `db:"tenant_id" json:"tenant_id" validate:"lte=64"`
	Creator     string          `db:"creator" json:"creator" validate:"lte=64"`
	Reviser     string          `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt   types.Time      `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt   types.Time      `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// Columns 返回 iac_plugin_instance 表列定义。
func (t IacPluginInstanceTable) Columns() *utils.Columns {
	return IacPluginInstanceTableColumns
}

// ColumnDescriptors 返回 iac_plugin_instance 表列描述符。
func (t IacPluginInstanceTable) ColumnDescriptors() utils.ColumnDescriptors {
	return IacPluginInstanceTableColumnDescriptors
}

// TableName 返回 iac_plugin_instance 表名。
func (t IacPluginInstanceTable) TableName() table.Name {
	return table.IacPluginInstanceTable
}

// InsertValidate 插入前校验 IacPluginInstanceTable。
func (t IacPluginInstanceTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be set")
	}

	if len(t.PluginName) == 0 {
		return errors.New("plugin_name is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.Attrs) == 0 {
		return errors.New("attrs is required")
	}

	if len(t.SecretAttrs) == 0 {
		return errors.New("secret_attrs is required")
	}

	if len(t.Extension) == 0 {
		return errors.New("extension is required")
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

// UpdateValidate 更新前校验 IacPluginInstanceTable。
func (t IacPluginInstanceTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be updated")
	}

	if len(t.PluginName) != 0 {
		return errors.New("plugin_name can not be updated")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if err := validator.ValidateMemo(t.Memo, false); err != nil {
		return err
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}
