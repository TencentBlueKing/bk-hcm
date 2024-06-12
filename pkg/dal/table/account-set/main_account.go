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

// Package accountset 账号表
package accountset

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// MainAccountColumns defines all the account tables's columns.
var MainAccountColumns = utils.MergeColumns(nil, MainAccountColumnDescriptor)

// MainAccountColumnDescriptor is MainAccountID's column descriptors.
var MainAccountColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "email", NamedC: "email", Type: enumor.String},
	{Column: "managers", NamedC: "managers", Type: enumor.Json},
	{Column: "bak_managers", NamedC: "bak_managers", Type: enumor.Json},
	{Column: "site", NamedC: "site", Type: enumor.String},
	{Column: "business_type", NamedC: "business_type", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "parent_account_name", NamedC: "parent_account_name", Type: enumor.String},
	{Column: "parent_account_id", NamedC: "parent_account_id", Type: enumor.String},
	{Column: "dept_id", NamedC: "dept_id", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// MainAccountTable 主账号（二级账号）表
type MainAccountTable struct {
	// ID 账号 ID
	ID string `db:"id" json:"id"`
	// Vendor 云厂商
	Vendor string `db:"vendor" json:"vendor"`
	// CloudID 云账号ID
	CloudID string `db:"cloud_id" json:"cloud_id"`
	// Email 邮箱
	Email string `db:"email" json:"email"`
	// Managers 责任人
	Managers types.StringArray `db:"managers" json:"managers"`
	// BakManagers 备份责任人
	BakManagers types.StringArray `db:"bak_managers" json:"bak_managers"`
	// Site 站点(中国站｜国际站)
	Site string `db:"site"`
	// Type 账号类型(国内业务|国际业务)
	BusinessType string `db:"business_type" json:"business_type"`
	// Status 状态（RUNNING｜DELETED｜SUSPENDED｜）
	Status string `db:"status" json:"status"`
	// ParentAccountName 所属账号
	ParentAccountName string `db:"parent_account_name" json:"parent_account_name"`
	// ParentAccountID 所属账号ID
	ParentAccountID string `db:"parent_account_id" json:"parent_account_id"`
	// DeptID 部门id
	DeptID int64 `db:"dept_id" json:"dept_id"`
	// BkBizID 业务id
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// OpProductID 运营产品id
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// Memo 账号信息备注
	Memo *string `db:"memo" json:"memo"`
	// Extension 云厂商账号差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName return account table name.
func (a MainAccountTable) TableName() table.Name {
	return table.MainAccountTable
}

// InsertValidate validate account table on insert.
func (a MainAccountTable) InsertValidate() error {
	if len(a.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(a.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	if len(a.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(a.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(a.Email) == 0 {
		return errors.New("email is required")
	}

	if len(a.Managers) == 0 {
		return errors.New("managers is required")
	}

	if len(a.Site) == 0 {
		return errors.New("site is required")
	}

	if len(a.BusinessType) == 0 {
		return errors.New("business_type is required")
	}

	if len(a.Status) == 0 {
		return errors.New("status is required")
	}

	if len(a.ParentAccountName) == 0 {
		return errors.New("parent_account_name is required")
	}

	if len(a.ParentAccountID) == 0 {
		return errors.New("parent_account_id is required")
	}

	return nil
}

// UpdateValidate validate account table on update.
func (a MainAccountTable) UpdateValidate() error {
	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
