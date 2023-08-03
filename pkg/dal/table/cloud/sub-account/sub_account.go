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

package tablesubaccount

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// Columns defines all the account table's columns.
var Columns = utils.MergeColumns(nil, ColumnDescriptor)

// ColumnDescriptor is SubAccountID's column descriptors.
var ColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "site", NamedC: "site", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "managers", NamedC: "managers", Type: enumor.Json},
	{Column: "bk_biz_ids", NamedC: "bk_biz_ids", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// Table 子账号表
type Table struct {
	// ID 账号 ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// CloudID 云账号ID
	CloudID string `db:"cloud_id" json:"cloud_id" validate:"lte=255"`
	// Name 子账号名称
	Name string `db:"name" json:"name" validate:"lte=255"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor" validate:"lte=16"`
	// Site 站点(中国站｜国际站)
	Site enumor.AccountSiteType `db:"site" json:"site" validate:"lte=32"`
	// AccountID 归属资源账号ID
	AccountID string `db:"account_id" json:"account_id" validate:"lte=64"`
	// Extension 云厂商账号差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Managers 责任人
	Managers types.StringArray `db:"managers" json:"managers"`
	// BkBizIDs 业务ID
	BkBizIDs types.Int64Array `db:"bk_biz_ids" json:"bk_biz_ids"`
	// Memo 账号信息备注
	Memo *string `db:"memo" json:"memo"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"lte=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName return account table name.
func (a Table) TableName() table.Name {
	return table.SubAccountTable
}

// InsertValidate validate account table on insert.
func (a Table) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(a.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(a.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(a.Site) == 0 {
		return errors.New("site is required")
	}

	if len(a.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(a.Extension) == 0 {
		return errors.New("extension is required")
	}

	if len(a.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate account table on update.
func (a Table) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
