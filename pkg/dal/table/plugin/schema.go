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

// Package tableplugin 定义 IaC 插件与插件 Schema 的数据库表模型。
package tableplugin

import (
	"encoding/json"
	"errors"
	"fmt"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// IacPluginSchemaTableColumns iac_plugin_schema 表全部列。
var IacPluginSchemaTableColumns = utils.MergeColumns(nil, IacPluginSchemaTableColumnDescriptors)

// IacPluginSchemaTableColumnDescriptors iac_plugin_schema 表列描述符。
var IacPluginSchemaTableColumnDescriptors = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "plugin_name", NamedC: "plugin_name", Type: enumor.String},
	{Column: "plugin_version", NamedC: "plugin_version", Type: enumor.String},
	{Column: "schema_hash", NamedC: "schema_hash", Type: enumor.String},
	{Column: "schema_detail", NamedC: "schema_detail", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// IacPluginSchemaTable iac_plugin_schema 表模型。
type IacPluginSchemaTable struct {
	ID            string          `db:"id" json:"id" validate:"lte=64"`
	PluginName    string          `db:"plugin_name" json:"plugin_name" validate:"lte=64"`
	PluginVersion string          `db:"plugin_version" json:"plugin_version" validate:"lte=64"`
	SchemaHash    string          `db:"schema_hash" json:"schema_hash" validate:"lte=80"`
	SchemaDetail  types.JsonField `db:"schema_detail" json:"schema_detail"`
	Creator       string          `db:"creator" json:"creator" validate:"lte=64"`
	Reviser       string          `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt     types.Time      `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt     types.Time      `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// UniqueKey 返回 iac_plugin_schema 表唯一键。
func (t IacPluginSchemaTable) UniqueKey() string {
	return fmt.Sprintf("(%s,%s)", t.PluginName, t.PluginVersion)
}

// Columns 返回 iac_plugin_schema 表列定义。
func (t IacPluginSchemaTable) Columns() *utils.Columns {
	return IacPluginSchemaTableColumns
}

// ColumnDescriptors 返回 iac_plugin_schema 表列描述符。
func (t IacPluginSchemaTable) ColumnDescriptors() utils.ColumnDescriptors {
	return IacPluginSchemaTableColumnDescriptors
}

// TableName 返回 iac_plugin_schema 表名。
func (t IacPluginSchemaTable) TableName() table.Name {
	return table.IacPluginSchemaTable
}

// InsertValidate 插入前校验 IacPluginSchemaTable。
func (t IacPluginSchemaTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be set")
	}

	if len(t.PluginName) == 0 {
		return errors.New("plugin_name is required")
	}

	if len(t.PluginVersion) == 0 {
		return errors.New("plugin_version is required")
	}

	if err := coreplugin.ValidateSchemaHashFormat(t.SchemaHash); err != nil {
		return err
	}

	if err := validateJSONObject("schema_detail", t.SchemaDetail); err != nil {
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

// UpdateValidate 更新前校验 IacPluginSchemaTable。
func (t IacPluginSchemaTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be updated")
	}

	if len(t.PluginName) != 0 {
		return errors.New("plugin_name can not be updated")
	}

	if len(t.PluginVersion) != 0 {
		return errors.New("plugin_version can not be updated")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if len(t.SchemaHash) != 0 {
		if err := coreplugin.ValidateSchemaHashFormat(t.SchemaHash); err != nil {
			return err
		}
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// validateJSONObject 校验 JSON 列必须是合法的 JSON 对象（不接受 null、数组、标量）。
func validateJSONObject(name string, field types.JsonField) error {
	if len(field) == 0 {
		return errors.New(name + " is required")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(field), &obj); err != nil || obj == nil {
		return errors.New(name + " must be a json object")
	}

	return nil
}
