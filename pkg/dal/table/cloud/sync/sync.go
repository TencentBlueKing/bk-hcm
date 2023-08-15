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

package tablessync

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AccountSyncDetailColumns defines all the account_sync_detail table's columns.
var AccountSyncDetailColumns = utils.MergeColumns(nil, AccountSyncDetailColumnDescriptor)

// AccountSyncDetailColumnDescriptor is account_sync_detail's column descriptors.
var AccountSyncDetailColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "res_name", NamedC: "res_name", Type: enumor.String},
	{Column: "res_status", NamedC: "res_status", Type: enumor.String},
	{Column: "res_end_time", NamedC: "res_end_time", Type: enumor.String},
	{Column: "res_failed_reason", NamedC: "res_failed_reason", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountSyncDetailTable define account_sync_detail table.
type AccountSyncDetailTable struct {
	ID              string          `db:"id" json:"id" validate:"lte=64"`
	Vendor          enumor.Vendor   `db:"vendor" json:"vendor"`
	AccountID       string          `db:"account_id" json:"account_id" validate:"lte=64"`
	ResName         string          `db:"res_name" json:"res_name" validate:"lte=64"`
	ResStatus       string          `db:"res_status" json:"res_status" validate:"lte=64"`
	ResEndTime      string          `db:"res_end_time" json:"res_end_time"`
	ResFailedReason types.JsonField `db:"res_failed_reason" json:"res_failed_reason"`
	Creator         string          `db:"creator" json:"creator" validate:"lte=64"`
	Reviser         string          `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt       types.Time      `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt       types.Time      `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// TableName return account_sync_detail table name.
func (a AccountSyncDetailTable) TableName() table.Name {
	return table.AccountSyncDetailTable
}

// InsertValidate account_sync_detail table when insert.
func (a AccountSyncDetailTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) == 0 {
		return errors.New("id is required")
	}

	if len(a.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(a.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(a.ResName) == 0 {
		return errors.New("res_name is required")
	}

	if len(a.ResStatus) == 0 {
		return errors.New("res_status is required")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate account_sync_detail table when update.
func (a AccountSyncDetailTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Vendor) != 0 {
		return errors.New("vendor can not update")
	}

	if len(a.AccountID) != 0 {
		return errors.New("account_id can not update")
	}

	if len(a.ResName) != 0 {
		return errors.New("res_name can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
