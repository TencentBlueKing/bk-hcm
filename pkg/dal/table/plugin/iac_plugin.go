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

// Package tableplugin defines IaC plugin database table models.
package tableplugin

import (
	"encoding/json"
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// IaCPluginColumns defines all the iac_plugin table's columns.
var IaCPluginColumns = utils.MergeColumns(nil, IaCPluginColumnDescriptor)

// IaCPluginColumnDescriptor is iac_plugin's column descriptors.
var IaCPluginColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "plugin_name", NamedC: "plugin_name", Type: enumor.String},
	{Column: "plugin_version", NamedC: "plugin_version", Type: enumor.String},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "message", NamedC: "message", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// IaCPluginTable IaC 插件注册表，每个插件的每个版本一条记录，平台级，不区分租户。
type IaCPluginTable struct {
	// ID 主键ID
	ID string `db:"id" json:"id"`
	// PluginName 插件名
	PluginName string `db:"plugin_name" json:"plugin_name"`
	// PluginVersion 插件版本
	PluginVersion string `db:"plugin_version" json:"plugin_version"`
	// Protocol 插件协议：tf_provider / http
	Protocol enumor.IaCPluginProtocol `db:"protocol" json:"protocol"`
	// State 状态：init / enabled / disabled / failed
	State enumor.IaCPluginState `db:"state" json:"state"`
	// Message 系统写入的状态说明，failed 时写明原因。用指针以便更新时能区分"不更新"与"清空"
	Message *string `db:"message" json:"message"`
	// Extension 协议相关信息，结构随 protocol 不同
	Extension types.JsonField `db:"extension" json:"extension"`
	// Memo 备注
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

// TableName return iac_plugin table name.
func (p IaCPluginTable) TableName() table.Name {
	return table.IaCPluginTable
}

// InsertValidate validate iac_plugin on insert.
func (p IaCPluginTable) InsertValidate() error {
	if len(p.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(p.PluginName) == 0 {
		return errors.New("plugin_name is required")
	}

	if len(p.PluginVersion) == 0 {
		return errors.New("plugin_version is required")
	}

	if err := p.Protocol.Validate(); err != nil {
		return err
	}

	if err := p.State.Validate(); err != nil {
		return err
	}

	if p.Message == nil {
		return errors.New("message is required")
	}

	if err := validateJSONObject("extension", p.Extension); err != nil {
		return err
	}

	if len(p.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(p.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(p.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(p.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate iac_plugin on update.
func (p IaCPluginTable) UpdateValidate() error {
	if len(p.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(p.Creator) != 0 {
		return errors.New("creator can not update")
	}

	// 身份字段创建后不可修改。
	if len(p.PluginName) != 0 {
		return errors.New("plugin_name can not update")
	}

	if len(p.PluginVersion) != 0 {
		return errors.New("plugin_version can not update")
	}

	if len(p.Protocol) != 0 {
		return errors.New("protocol can not update")
	}

	if len(p.State) != 0 {
		if err := p.State.Validate(); err != nil {
			return err
		}
	}

	if len(p.Extension) != 0 {
		if err := validateJSONObject("extension", p.Extension); err != nil {
			return err
		}
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
