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

	"github.com/shopspring/decimal"
)

// AccountBillItemColumns defines account_bill_item's columns.
var AccountBillItemColumns = utils.MergeColumns(nil, AccountBillItemColumnDescriptor)

// AccountBillItemColumnDescriptor is account_bill_item's column descriptors.
var AccountBillItemColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "first_account_id", NamedC: "first_account_id", Type: enumor.String},
	{Column: "second_account_id", NamedC: "second_account_id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "bill_day", NamedC: "bill_day", Type: enumor.Numeric},
	{Column: "version_id", NamedC: "version_id", Type: enumor.String},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "rmb_cost", NamedC: "rmb_cost", Type: enumor.Numeric},
	{Column: "hc_product_code", NamedC: "hc_product_code", Type: enumor.String},
	{Column: "hc_product_name", NamedC: "hc_product_name", Type: enumor.String},
	{Column: "res_amount", NamedC: "res_amount", Type: enumor.Numeric},
	{Column: "res_amount_unit", NamedC: "res_amount_unit", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillItem 存储分账后的明细
type AccountBillItem struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// FirstAccountID 一级账号ID
	FirstAccountID string `db:"first_account_id" json:"first_account_id"`
	// SecondAccountID 账号ID
	SecondAccountID string `db:"second_account_id" json:"second_account_id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// ProductID 运营产品ID
	ProductID int64 `db:"product_id" json:"product_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BillYear 账单年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 账单月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// BillDay 账单日期
	BillDay int `db:"bill_day" json:"bill_day"`
	// VersionID AccountBillSummary VersionID
	VersionID string `db:"version_id" json:"version_id"`
	// Currency 币种
	Currency string `db:"currency" json:"currency"`
	// Cost 费用
	Cost decimal.Decimal `db:"cost" json:"cost"`
	// RMBCost 费用
	RMBCost decimal.Decimal `db:"rmb_cost" json:"rmb_cost"`
	// HcProductCode 云服务代号
	HcProductCode string `db:"hc_product_code" json:"hc_product_code"`
	// HcProductName 云服务名字
	HcProductName string `db:"hc_product_name" json:"hc_product_name"`
	// ResAmount 用量，部分云账单可能没有
	ResAmount decimal.Decimal `db:"res_amount" json:"res_amount,omitempty"`
	// ResAmountUnit 用量单位
	ResAmountUnit string `db:"res_amount_unit" json:"res_amount_unit,omitempty"`
	// Extension 云原始字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回账单明细表名
func (abs *AccountBillItem) TableName() table.Name {
	return table.AccountBillItemTable
}

// InsertValidate validate account bill item on insert
func (abs *AccountBillItem) InsertValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.Vendor) == 0 {
		return errors.New("vendor is required")
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
	if len(abs.VersionID) == 0 {
		return errors.New("version_ib is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill item on update
func (abs *AccountBillItem) UpdateValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.Vendor) == 0 {
		return errors.New("vendor is required")
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
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}
