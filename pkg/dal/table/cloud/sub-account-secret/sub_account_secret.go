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

package tablesubaccountsecret

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// Columns defines all the sub account secret table's columns.
var Columns = utils.MergeColumns(nil, ColumnDescriptor)

// ColumnDescriptor is SubAccountSecret's column descriptors.
var ColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "sub_account_id", NamedC: "sub_account_id", Type: enumor.String},
	{Column: "cloud_created_at", NamedC: "cloud_created_at", Type: enumor.Time},
	{Column: "disabled_time", NamedC: "disabled_time", Type: enumor.Time},
	{Column: "last_used_time", NamedC: "last_used_time", Type: enumor.Time},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// Table 子账号密钥表
type Table struct {
	// ID 密钥ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor" validate:"lte=16"`
	// Status 密钥状态(enabled:启用 disabled:禁用)
	Status enumor.SubAccountSecretStatus `db:"status" json:"status" validate:"lte=16"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// AccountID 账号id
	AccountID string `db:"account_id" json:"account_id" validate:"lte=64"`
	// SubAccountID 子账号id
	SubAccountID string `db:"sub_account_id" json:"sub_account_id" validate:"lte=64"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id" validate:"lte=64"`
	// CloudCreatedAt 云上创建时间
	CloudCreatedAt types.Time `db:"cloud_created_at" json:"cloud_created_at"`
	// DisabledTime 本地禁用时间
	DisabledTime types.Time `db:"disabled_time" json:"disabled_time"`
	// LastUsedTime 密钥上次调用时间
	LastUsedTime types.Time `db:"last_used_time" json:"last_used_time"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"lte=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName return sub account secret table name.
func (t Table) TableName() table.Name {
	return table.SubAccountSecretTable
}

// InsertValidate validate sub account secret table on insert.
func (t Table) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.Status) == 0 {
		return errors.New("status is required")
	}

	if err := t.Status.Validate(); err != nil {
		return err
	}

	if len(t.Extension) == 0 {
		return errors.New("extension is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(t.SubAccountID) == 0 {
		return errors.New("sub_account_id is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate sub account secret table on update.
func (t Table) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("created_at can not update")
	}

	return nil
}
