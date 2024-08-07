/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"

	"github.com/shopspring/decimal"
)

// BatchCreateBillExchangeRateReq ...
type BatchCreateBillExchangeRateReq struct {
	ExchangeRates []ExchangeRateCreate `json:"exchange_rates" validate:"required,dive,required"`
}

// ExchangeRateCreate ...
type ExchangeRateCreate struct {
	// Year 账单年份
	Year int `json:"year" validate:"required,gt=0"`
	// Month 账单月份
	Month int `json:"month" validate:"required,gte=1,lte=12"`
	// FromCurrency 原币种
	FromCurrency enumor.CurrencyCode `json:"from_currency" validate:"required"`
	// ToCurrency 转换后币种
	ToCurrency enumor.CurrencyCode `json:"to_currency" validate:"required"`
	// ExchangeRate 汇率
	ExchangeRate *decimal.Decimal `json:"exchange_rate" validate:"required"`
}

// Validate ...
func (r *BatchCreateBillExchangeRateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// Validate ...
func (r *ExchangeRateCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// ExchangeRateUpdateReq ...
type ExchangeRateUpdateReq struct {
	ID string `json:"id" validate:"required"`
	// Year 账单年份
	Year int `json:"year" validate:"omitempty"`
	// Month 账单月份
	Month int `json:"month" validate:"omitempty,gte=1,lte=12"`
	// FromCurrency 原币种
	FromCurrency enumor.CurrencyCode `json:"from_currency"`
	// ToCurrency 转换后币种
	ToCurrency enumor.CurrencyCode `json:"to_currency"`
	// ExchangeRate 汇率
	ExchangeRate *decimal.Decimal `json:"exchange_rate" `
}

// Validate ...
func (r *ExchangeRateUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ExchangeRateListResult ...
type ExchangeRateListResult = core.ListResultT[bill.ExchangeRate]
