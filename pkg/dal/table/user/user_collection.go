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

package tableuser

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// UserCollTableColumns defines all the cvm table's columns.
var UserCollTableColumns = utils.MergeColumns(nil, UserCollTableColumnDescriptor)

// UserCollTableColumnDescriptor is cvm table column descriptors.
var UserCollTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "user", NamedC: "user", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// UserCollTable define cvm table.
type UserCollTable struct {
	ID        string                       `db:"id" json:"id" validate:"lte=64"`
	User      string                       `db:"user" json:"user" validate:"lte=64"`
	ResType   enumor.UserCollectionResType `db:"res_type" json:"res_type" validate:"lte=50"`
	ResID     string                       `db:"res_id" json:"res_id" validate:"lte=64"`
	Creator   string                       `db:"creator" json:"creator" validate:"lte=64"`
	CreatedAt types.Time                   `db:"created_at" validate:"excluded_unless" json:"created_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return cvm table name.
func (t UserCollTable) TableName() table.Name {
	return table.UserCollectionTable
}

// InsertValidate cvm table when insert.
func (t UserCollTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if err := t.ResType.Validate(); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.User) == 0 {
		return errors.New("user is required")
	}

	if len(t.ResType) == 0 {
		return errors.New("res type is required")
	}

	if len(t.ResID) == 0 {
		return errors.New("res id is required")
	}

	return nil
}
