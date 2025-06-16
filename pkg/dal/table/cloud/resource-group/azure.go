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

package resourcegroup

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AzureRGColumns defines all the azure resource group table's columns.
var AzureRGColumns = utils.MergeColumns(nil, AzureRGTableColumnDescriptor)

// AzureRGTableColumnDescriptor is azure resource group's column descriptors.
var AzureRGTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "location", NamedC: "location", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AzureRGTable azure资源表
type AzureRGTable struct {
	// ID 账号 ID
	ID string `db:"id"`
	// name 资源名称
	Name string `db:"name"`
	// Type 资源类型
	Type string `db:"type"`
	// Location 地域
	Location string `db:"location"`
	// AccountID 账号id
	AccountID string `db:"account_id"`
	// Creator 创建者
	Creator string `db:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return azure resource group table name.
func (a AzureRGTable) TableName() table.Name {
	return table.AzureRGTable
}

// InsertValidate azure resource group table when insert.
func (t AzureRGTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.Type) == 0 {
		return errors.New("type is required")
	}

	if len(t.Location) == 0 {
		return errors.New("location is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("accout_id is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate azure security group rule table when update.
func (t AzureRGTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
