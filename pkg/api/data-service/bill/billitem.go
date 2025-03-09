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
	rawjson "encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	typesbill "hcm/pkg/dal/dao/types/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BatchRawBillItemCreateReq ...
type BatchRawBillItemCreateReq = BatchBillItemCreateReq[rawjson.RawMessage]

// BatchBillItemCreateReq batch bill item create request
type BatchBillItemCreateReq[E bill.BillItemExtension] struct {
	*ItemCommonOpt `json:",inline" validate:"required"`
	Items          []BillItemCreateReq[E] `json:"items" validate:"required,dive"`
}

// BillItemCreateReq create request
type BillItemCreateReq[E bill.BillItemExtension] struct {
	RootAccountID string        `json:"root_account_id" validate:"required"`
	MainAccountID string        `json:"main_account_id" validate:"required"`
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
	ProductID     int64         `json:"product_id" validate:"omitempty"`
	BkBizID       int64         `json:"bk_biz_id" validate:"omitempty"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	// allow bill day `zero` for monthly bill
	BillDay       int                 `json:"bill_day" `
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
func (req *BatchBillItemCreateReq[E]) Validate() error {
	if len(req.Items) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("bill item  count should <= %d", constant.BatchOperationMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// Validate ...
func (c *BillItemCreateReq[E]) Validate() error {
	return validator.Validate.Struct(c)
}

// BillItemListReq list request
type BillItemListReq struct {
	*ItemCommonOpt `json:",inline" validate:"required"`

	*core.ListReq `json:",inline"`
}

// Validate ...
func (r *BillItemListReq) Validate() error {
	if err := r.ItemCommonOpt.Validate(); err != nil {
		return err
	}
	return r.ListReq.Validate()
}

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
	*ItemCommonOpt `json:",inline" validate:"required"`

	ID            string              `json:"id,omitempty" validate:"required"`
	Currency      enumor.CurrencyCode `json:"currency" validate:"required"`
	Cost          *decimal.Decimal    `json:"cost" validate:"required"`
	HcProductCode string              `json:"hc_product_code,omitempty"`
	HcProductName string              `json:"hc_product_name,omitempty"`
	ResAmount     *decimal.Decimal    `json:"res_amount,omitempty"`
	ResAmountUnit string              `json:"res_amount_unit,omitempty"`
	Extension     types.JsonField     `json:"extension"`
}

// Validate ...
func (req *BillItemUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BillItemDeleteReq ...
type BillItemDeleteReq struct {
	*ItemCommonOpt `json:",inline" validate:"required"`
	Filter         *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (r *BillItemDeleteReq) Validate() error {

	if r.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}
	if r.ItemCommonOpt == nil {
		return errf.New(errf.InvalidParameter, "item common option is required")
	}
	return r.ItemCommonOpt.Validate()
}

// ItemCommonOpt general option for all bill item operations
type ItemCommonOpt = typesbill.ItemCommonOpt

// BillItemSumReq ...
type BillItemSumReq struct {
	*ItemCommonOpt `json:",inline" validate:"required"`
	Filter         *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (r *BillItemSumReq) Validate() error {

	if r.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}
	if r.ItemCommonOpt == nil {
		return errf.New(errf.InvalidParameter, "item common option is required")
	}
	return r.ItemCommonOpt.Validate()
}

// BillItemSumResult sum bill item result
type BillItemSumResult struct {
	Count    uint64              `json:"count,omitempty"`
	Cost     decimal.Decimal     `json:"cost"`
	Currency enumor.CurrencyCode `json:"currency"`
}
