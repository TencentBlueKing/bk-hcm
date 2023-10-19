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

// Package bill ...
package bill

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AwsBillColumns defines all the Aws bill table's columns.
var AwsBillColumns = utils.MergeColumns(nil, AwsBillColumnDescriptor)

// AwsBillColumnDescriptor is AwsBill's column descriptors.
var AwsBillColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_database_name", NamedC: "cloud_database_name", Type: enumor.String},
	{Column: "cloud_table_name", NamedC: "cloud_table_name", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.Numeric},
	{Column: "err_msg", NamedC: "err_msg", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillConfigTable account_bill_config表
type AccountBillConfigTable struct {
	// ID 自增ID
	ID string `db:"id" json:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"max=64" json:"account_id"`
	// CloudDatabaseName 云账单数据库名称
	CloudDatabaseName string `db:"cloud_database_name" validate:"max=64" json:"cloud_database_name"`
	// CloudTableName 云账单数据表名称
	CloudTableName string `db:"cloud_table_name" validate:"max=64" json:"cloud_table_name"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Status 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
	Status int64 `db:"status" json:"status"`
	// ErrMsg 错误描述
	ErrMsg types.JsonField `db:"err_msg" json:"err_msg"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return account bill config table name.
func (a AccountBillConfigTable) TableName() table.Name {
	return table.AccountBillConfigTable
}

// InsertValidate validate account bill config table on insert.
func (a AccountBillConfigTable) InsertValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(a.AccountID) == 0 {
		return errors.New("account_id can not be empty")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate account bill config table on update.
func (a AccountBillConfigTable) UpdateValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
