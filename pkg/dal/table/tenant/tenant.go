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

// Package tenant ...
package tenant

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// TenantColumns defines all the tenant table's columns.
var TenantColumns = utils.MergeColumns(nil, TenantColumnDescriptor)

// TenantColumnDescriptor is tenant column descriptors.
var TenantColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "tenant_id", NamedC: "tenant_id", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TenantTable is used to save tenant information.
type TenantTable struct {
	ID        string              `db:"id" json:"id" validate:"max=64"`
	TenantID  string              `db:"tenant_id" json:"tenant_id" validate:"max=64"`
	Status    enumor.TenantStatus `db:"status" json:"status"`
	Creator   string              `db:"creator" validate:"max=64" json:"creator"`
	Reviser   string              `db:"reviser" validate:"max=64" json:"reviser"`
	CreatedAt types.Time          `db:"created_at" validate:"isdefault" json:"created_at"`
	UpdatedAt types.Time          `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the tenant database table name.
func (d TenantTable) TableName() table.Name {
	return table.TenantTable
}

// InsertValidate validate tenant on insertion.
func (d TenantTable) InsertValidate() error {
	if len(d.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(d.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if err := validator.Validate.Struct(d); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate tenant on update.
func (d TenantTable) UpdateValidate() error {
	if len(d.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(d.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if err := validator.Validate.Struct(d); err != nil {
		return err
	}
	return nil
}
