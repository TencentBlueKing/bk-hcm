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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"

	"github.com/shopspring/decimal"
)

// RawBillItem raw bill item
type RawBillItem struct {
	Region        string              `json:"region,omitempty" validate:"required"`
	HcProductCode string              `json:"hcProductCode,omitempty"`                    // 云服务代号
	HcProductName string              `json:"hcProductName,omitempty"`                    // 云服务名字
	BillCurrency  enumor.CurrencyCode `json:"billCurrency,omitempty" validate:"required"` // 币种
	BillCost      decimal.Decimal     `json:"billCost,omitempty" validate:"required"`     // 原币种消费（元）
	ResAmount     decimal.Decimal     `json:"resAmount,omitempty"`                        // 用量，部分云账单可能没有
	ResAmountUnit string              `json:"resAmountUnit,omitempty"`                    // 用量单位
	Extension     types.JsonField     `json:"extension" validate:"required"`              // 存储云原始账单信息
}

// RawBillCreateReq create request
type RawBillCreateReq struct {
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
	RootAccountID string        `json:"root_account_id" validate:"required"`
	AccountID     string        `json:"account_id" validate:"required"`
	BillYear      string        `json:"bill_year" validate:"required"`
	BillMonth     string        `json:"bill_month" validate:"required"`
	Version       string        `json:"version" validate:"required"`
	BillDate      string        `json:"bill_date" validate:"required"`
	// FileName cos写入的文件名
	FileName string        `json:"file_name" validate:"required"`
	Items    []RawBillItem `json:"items" validate:"required"`
}

// Validate RawBillCreateReq.
func (c *RawBillCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// RawBillItemQueryReq request for query bill item content
// only used in client
type RawBillItemQueryReq struct {
	Vendor         enumor.Vendor `json:"vendor" validate:"required"`
	FirstAccountID string        `json:"first_account_id" validate:"required"`
	AccountID      string        `json:"account_id" validate:"required"`
	BillYear       string        `json:"bill_year" validate:"required"`
	BillMonth      string        `json:"bill_month" validate:"required"`
	Version        string        `json:"version" validate:"required"`
	BillDate       string        `json:"bill_date" validate:"required"`
	// FileName cos写入的文件名
	FileName string `json:"file_name" validate:"required"`
}

// RawBillItemQueryResult query item list
type RawBillItemQueryResult struct {
	Count   *uint64        `json:"count,omitempty"`
	Details []*RawBillItem `json:"details"`
}

// RawBillItemNameListReq request for list bill name list
// only used in client
type RawBillItemNameListReq struct {
	Vendor         enumor.Vendor `json:"vendor" validate:"required"`
	FirstAccountID string        `json:"first_account_id" validate:"required"`
	AccountID      string        `json:"account_id" validate:"required"`
	BillYear       string        `json:"bill_year" validate:"required"`
	BillMonth      string        `json:"bill_month" validate:"required"`
	Version        string        `json:"version" validate:"required"`
	BillDate       string        `json:"bill_date" validate:"required"`
}

// RawBillItemNameListResult list filenames
type RawBillItemNameListResult struct {
	Filenames []string `json:"filenames,omitempty"`
}
