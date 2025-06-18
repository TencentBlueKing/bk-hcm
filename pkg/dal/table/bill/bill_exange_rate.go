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
	cvt "hcm/pkg/tools/converter"
)

// AccountBillExchangeRateColumns defines account_bill_exchange_rate's columns.
var AccountBillExchangeRateColumns = utils.MergeColumns(nil, AccountBillExchangeRateColumnDescriptor)

// AccountBillExchangeRateColumnDescriptor is account_bill_exchange_rate's column descriptors.
var AccountBillExchangeRateColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "from_currency", NamedC: "from_currency", Type: enumor.String},
	{Column: "to_currency", NamedC: "to_currency", Type: enumor.String},
	{Column: "exchange_rate", NamedC: "exchange_rate", Type: enumor.Numeric},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillExchangeRate 汇率表
type AccountBillExchangeRate struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`

	// Year 账单年份
	Year int `db:"year" json:"year"`
	// Month 账单月份
	Month int `db:"month" json:"month"`
	// FromCurrency 原币种
	FromCurrency enumor.CurrencyCode `db:"from_currency" json:"from_currency"`
	// ToCurrency 转换后币种
	ToCurrency enumor.CurrencyCode `db:"to_currency" json:"to_currency"`
	// ExchangeRate 汇率
	ExchangeRate *types.Decimal `db:"exchange_rate" json:"exchange_rate"`

	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// Reviser 修改人
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回汇率表名
func (rate *AccountBillExchangeRate) TableName() table.Name {
	return table.AccountBillExchangeRateTable
}

// InsertValidate validate exchange rate on insert
func (rate *AccountBillExchangeRate) InsertValidate() error {
	if len(rate.ID) == 0 {
		return errors.New("id is required")
	}

	if rate.Year == 0 {
		return errors.New("year is required")
	}
	if rate.Month == 0 {
		return errors.New("month is required")
	}

	if len(rate.ToCurrency) == 0 {
		return errors.New("to currency is required")
	}
	if len(rate.FromCurrency) == 0 {
		return errors.New("from currency is required")
	}
	if cvt.PtrToVal(rate.ExchangeRate).IsZero() {
		return errors.New("exchange rate is required")
	}
	if len(rate.Creator) == 0 {
		return errors.New("creator is required")
	}
	if len(rate.Reviser) == 0 {
		return errors.New("reviser is required")
	}
	return validator.Validate.Struct(rate)
}

// UpdateValidate validate exchange rate on update
func (rate *AccountBillExchangeRate) UpdateValidate() error {
	if len(rate.ID) == 0 {
		return errors.New("id is required")
	}
	if len(rate.Reviser) == 0 {
		return errors.New("reviser is required")
	}
	if len(rate.Creator) != 0 {
		return errors.New("creator is not allowed")
	}
	return validator.Validate.Struct(rate)
}
