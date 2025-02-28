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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// VpcColumns defines all the vpc table's columns.
var VpcColumns = utils.MergeColumns(nil, VpcColumnDescriptor)

// VpcColumnDescriptor is Vpc's column descriptors.
var VpcColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "category", NamedC: "category", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// VpcTable vpc表
type VpcTable struct {
	// ID vpc ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"max=64" json:"account_id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=255" json:"cloud_id"`
	// Name vpc名称
	Name *string `db:"name" validate:"omitempty,max=128" json:"name"`
	// Region 地域
	Region string `db:"region" validate:"max=255" json:"region"`
	// Category 类别
	Category enumor.VpcCategory `db:"category" validate:"max=32" json:"category"`
	// Memo 备注
	Memo *string `db:"memo" validate:"omitempty,max=255" json:"memo"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" validate:"-" json:"extension"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName return vpc table name.
func (v VpcTable) TableName() table.Name {
	return table.VpcTable
}

// InsertValidate validate vpc table on insert.
func (v VpcTable) InsertValidate() error {
	if err := v.Vendor.Validate(); err != nil {
		return err
	}

	if len(v.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if len(v.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if v.Name == nil {
		return errors.New("name can not be nil")
	}

	if err := v.Category.Validate(); err != nil {
		return err
	}

	if v.BkBizID == 0 {
		return errors.New("biz id can not be empty")
	}

	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate vpc table on update.
func (v VpcTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	if v.Name == nil && len(v.Category) == 0 && len(v.Extension) == 0 && v.BkBizID == 0 && v.Memo == nil {
		return errors.New("at least one of the update fields must be set")
	}

	if len(v.AccountID) != 0 {
		return errors.New("account id can not update")
	}

	if len(v.CloudID) != 0 {
		return errors.New("cloud id can not update")
	}

	if len(v.Region) != 0 {
		return errors.New("region can not update")
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(v)
}
