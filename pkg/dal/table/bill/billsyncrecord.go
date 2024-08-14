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

package bill

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AccountBillSyncRecordColumns defines account_bill_adjustment_item's columns.
var AccountBillSyncRecordColumns = utils.MergeColumns(nil, AccountBillSyncRecordColumnDescriptor)

// AccountBillSyncRecordColumnDescriptor is account_bill_summary_daily's column descriptors.
var AccountBillSyncRecordColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "count", NamedC: "count", Type: enumor.Numeric},
	{Column: "rmb_cost", NamedC: "rmb_cost", Type: enumor.Numeric},
	{Column: "detail", NamedC: "detail", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillSyncRecord account bill sync record
type AccountBillSyncRecord struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// BillYear 出账年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 出账月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// State 同步状态
	State enumor.BillSyncState `db:"state" json:"state"`
	// Currency 币种
	Currency enumor.CurrencyCode `db:"currency" json:"currency"`
	// Count 账单数量
	Count *uint `db:"count" json:"count"`
	// Cost 账单
	Cost *types.Decimal `db:"cost" json:"cost"`
	// RMBCost 人民币账单
	RMBCost *types.Decimal `db:"rmb_cost" json:"rmb_cost"`
	// Detail 同步详情
	Detail string `db:"detail" json:"detail"`
	// Operator 操作人
	Operator string `db:"operator" validate:"max=64" json:"operator"`

	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回每天汇总账单版本表名
func (absr *AccountBillSyncRecord) TableName() table.Name {
	return table.AccountBillSyncRecordTable
}

// InsertValidate validate account bill sync record on insert
func (absr *AccountBillSyncRecord) InsertValidate() error {
	if len(absr.ID) == 0 {
		return errors.New("id is required")
	}
	if len(absr.Vendor) == 0 {
		return errors.New("vendor is required")
	}
	if absr.BillYear == 0 {
		return errors.New("bill_year is required")
	}
	if absr.BillMonth == 0 {
		return errors.New("bill_month is required")
	}
	if len(absr.State) == 0 {
		return errors.New("state is required")
	}
	return validator.Validate.Struct(absr)
}

// UpdateValidate validate account bill day on update
func (absr *AccountBillSyncRecord) UpdateValidate() error {
	if len(absr.ID) == 0 {
		return errors.New("id is required")
	}
	return validator.Validate.Struct(absr)
}
