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

// AccountBillAdjustmentItemColumns defines account_bill_adjustment_item's columns.
var AccountBillAdjustmentItemColumns = utils.MergeColumns(nil, AccountBillAdjustmentItemColumnDescriptor)

// AccountBillAdjustmentItemColumnDescriptor is account_bill_summary_daily's column descriptors.
var AccountBillAdjustmentItemColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "first_account_id", NamedC: "first_account_id", Type: enumor.String},
	{Column: "second_account_id", NamedC: "second_account_id", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "bill_day", NamedC: "bill_day", Type: enumor.Numeric},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "rmb_cost", NamedC: "rmb_cost", Type: enumor.Numeric},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillAdjustmentItem 存储手动调账明细
type AccountBillAdjustmentItem struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// FirstAccountID 一级账号ID
	FirstAccountID string `db:"first_account_id" json:"first_account_id"`
	// SecondAccountID 账号ID
	SecondAccountID string `db:"second_account_id" json:"second_account_id"`
	// ProductID 运营产品ID
	ProductID int64 `db:"product_id" json:"product_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BillYear 账单年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 账单月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// BillDay 账单天
	BillDay int `db:"bill_day" json:"bill_day"`
	// Type 调账类型
	Type string `db:"type" json:"type"`
	// Memo 注解
	Memo string `db:"memo" json:"memo"`
	// Operator 调账类型
	Operator string `db:"operator" json:"operator"`
	// Currency 币种
	Currency string `db:"currency" json:"currency"`
	// Cost 费用
	Cost *types.Decimal `db:"cost" json:"cost"`
	// RMBCost 费用
	RMBCost *types.Decimal `db:"rmb_cost" json:"rmb_cost"`
	// State 状态，未确定、已确定
	State string `db:"string" json:"string"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回账单明细表名
func (abs *AccountBillAdjustmentItem) TableName() table.Name {
	return table.AccountBillAdjustmentItemTable
}

// InsertValidate validate account bill item on insert
func (abs *AccountBillAdjustmentItem) InsertValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.FirstAccountID) == 0 {
		return errors.New("first_account_id is required")
	}
	if len(abs.SecondAccountID) == 0 {
		return errors.New("second_account_id is required")
	}
	if abs.BkBizID == 0 && abs.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if abs.BillYear == 0 {
		return errors.New("bill_year is required")
	}
	if abs.BillMonth == 0 {
		return errors.New("bill_month is required")
	}
	if abs.BillDay == 0 {
		return errors.New("bill_day is required")
	}
	if len(abs.Type) == 0 {
		return errors.New("type is required")
	}
	if len(abs.Operator) == 0 {
		return errors.New("operator is required")
	}
	if len(abs.Currency) == 0 {
		return errors.New("currency is required")
	}
	if abs.Cost.IsZero() {
		return errors.New("cost is required")
	}
	if len(abs.State) == 0 {
		return errors.New("state is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill item on update
func (abs *AccountBillAdjustmentItem) UpdateValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.FirstAccountID) == 0 {
		return errors.New("first_account_id is required")
	}
	if len(abs.SecondAccountID) == 0 {
		return errors.New("second_account_id is required")
	}
	if abs.BkBizID == 0 && abs.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if abs.BillYear == 0 {
		return errors.New("bill_year is required")
	}
	if abs.BillMonth == 0 {
		return errors.New("bill_month is required")
	}
	if abs.BillDay == 0 {
		return errors.New("bill_day is required")
	}
	if len(abs.Type) == 0 {
		return errors.New("type is required")
	}
	if len(abs.Operator) == 0 {
		return errors.New("operator is required")
	}
	if len(abs.Currency) == 0 {
		return errors.New("currency is required")
	}
	if abs.Cost.IsZero() {
		return errors.New("cost is required")
	}
	if len(abs.State) == 0 {
		return errors.New("state is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}
