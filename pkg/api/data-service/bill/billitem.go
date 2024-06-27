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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"

	"github.com/shopspring/decimal"
)

// BatchBillItemCreateReq batch bill item create request
type BatchBillItemCreateReq[E bill.BillItemExtension] []BillItemCreateReq[E]

// BillItemCreateReq create request
type BillItemCreateReq[E bill.BillItemExtension] struct {
	RootAccountID string              `json:"root_account_id" validate:"required"`
	MainAccountID string              `json:"main_account_id" validate:"required"`
	Vendor        enumor.Vendor       `json:"vendor" validate:"required"`
	ProductID     int64               `json:"product_id" validate:"omitempty"`
	BkBizID       int64               `json:"bk_biz_id" validate:"omitempty"`
	BillYear      int                 `json:"bill_year" validate:"required"`
	BillMonth     int                 `json:"bill_month" validate:"required"`
	BillDay       int                 `json:"bill_day" validate:"required"`
	VersionID     int                 `json:"version_id" validate:"required"`
	Currency      enumor.CurrencyCode `json:"currency" validate:"required"`
	Cost          decimal.Decimal     `json:"cost" validate:"required"`
	HcProductCode string              `json:"hc_product_code,omitempty"`
	HcProductName string              `json:"hc_product_name,omitempty"`
	ResAmount     decimal.Decimal     `json:"res_amount,omitempty"`
	ResAmountUnit string              `json:"res_amount_unit,omitempty"`
	Extension     *E                  `json:"extension"`
}

// Validate ...
func (c *BillItemCreateReq[E]) Validate() error {
	return validator.Validate.Struct(c)
}

// BillItemListReq list request
type BillItemListReq = core.ListReq

// BillItemBaseListResult ...
type BillItemBaseListResult = core.ListResultT[*bill.BaseBillItem]

// TCloudBillItemListResult ...
type TCloudBillItemListResult = core.ListResultT[*bill.TCloudBillItem]

// GcpBillItemListResult ...
type GcpBillItemListResult = core.ListResultT[*bill.GcpBillItem]

// AwsBillItemListResult ...
type AwsBillItemListResult = core.ListResultT[*bill.AwsBillItem]

// AzureBillItemListResult ...
type AzureBillItemListResult = core.ListResultT[*bill.AzureBillItem]

// HuaweiBillItemListResult ...
type HuaweiBillItemListResult = core.ListResultT[*bill.HuaweiBillItem]

// KaopuBillItemListResult ...
type KaopuBillItemListResult = core.ListResultT[*bill.KaopuBillItem]

// ZenlayerBillItemListResult ...
type ZenlayerBillItemListResult = core.ListResultT[*bill.ZenlayerBillItem]

// BillItemUpdateReq update request
type BillItemUpdateReq struct {
	ID            string              `json:"id,omitempty" validate:"required"`
	Currency      enumor.CurrencyCode `json:"currency" validate:"required"`
	Cost          *decimal.Decimal     `json:"cost" validate:"required"`
	HcProductCode string              `json:"hc_product_code,omitempty"`
	HcProductName string              `json:"hc_product_name,omitempty"`
	ResAmount     *decimal.Decimal     `json:"res_amount,omitempty"`
	ResAmountUnit string              `json:"res_amount_unit,omitempty"`
	Extension     types.JsonField     `json:"extension"`
}

// Validate ...
func (req *BillItemUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
