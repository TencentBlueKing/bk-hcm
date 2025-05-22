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

package tableargstpl

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ArgumentTplTableColumns defines all the argument template table's columns.
var ArgumentTplTableColumns = utils.MergeColumns(nil, ArgumentTplTableColumnDescriptor)

// ArgumentTplTableColumnDescriptor is argument template table column descriptors.
var ArgumentTplTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "templates", NamedC: "templates", Type: enumor.Json},
	{Column: "group_templates", NamedC: "group_templates", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.String},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.String},
}

// ArgumentTemplateTable DB表
type ArgumentTemplateTable struct {
	// ID scheme ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=255" json:"cloud_id"`
	// Name 名称
	Name string `db:"name" validate:"max=255" json:"name"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
	// AccountID 账号ID
	AccountID string `json:"account_id" db:"account_id"`
	// Type 参数模版类型
	Type enumor.TemplateType `db:"type" json:"type"`
	// Templates 参数模版的参数数组
	Templates types.JsonField `db:"templates" json:"templates"`
	// GroupTemplates 参数模版组的参数数组
	GroupTemplates types.JsonField `db:"group_templates" json:"group_templates"`
	// Memo 备注
	Memo *string `db:"memo" json:"memo"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return argument template table name.
func (v ArgumentTemplateTable) TableName() table.Name {
	return table.ArgumentTemplateTable
}

// InsertValidate validate argument template table on insert.
func (v ArgumentTemplateTable) InsertValidate() error {
	if len(v.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if len(v.Name) == 0 {
		return errors.New("name can not be nil")
	}

	if len(v.Vendor) == 0 {
		return errors.New("vendor can not be empty")
	}

	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate argument template table on update.
func (v ArgumentTemplateTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
