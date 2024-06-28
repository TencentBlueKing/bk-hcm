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

	"github.com/shopspring/decimal"
)

// BatchBillSyncRecordCreateReq batch create request
type BatchBillSyncRecordCreateReq struct {
	Items []BillSyncRecordCreateReq `json:"items" validate:"required,min=1,max=100,dive,required"`
}

// Validate ...
func (r *BatchBillSyncRecordCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// BillSyncRecordCreateReq create request
type BillSyncRecordCreateReq struct {
	Vendor    enumor.Vendor       `json:"vendor" validate:"required"`
	BillYear  int                 `json:"bill_year" validate:"required"`
	BillMonth int                 `json:"bill_month" validate:"required"`
	State     string              `json:"state" validate:"required"`
	Currency  enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	Cost      decimal.Decimal     `json:"cost" validate:"omitempty"`
	RMBCost   decimal.Decimal     `json:"rmb_cost" validate:"omitempty"`
	Detail    string              `json:"detail" validate:"omitempty"`
	Operator  string              `json:"operator" validate:"max=64" `
}

// Validate ...
func (c *BillSyncRecordCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BillSyncRecordListReq list request
type BillSyncRecordListReq = core.ListReq

// BillSyncRecordListResult list result
type BillSyncRecordListResult = core.ListResultT[*bill.SyncRecord]

// BillSyncRecordUpdateReq update request
type BillSyncRecordUpdateReq struct {
	ID       string              `json:"id"`
	State    string              `json:"state" validate:"omitempty"`
	Currency enumor.CurrencyCode `json:"currency" validate:"omitempty"`
	Cost     *decimal.Decimal    `json:"cost" validate:"omitempty"`
	RMBCost  *decimal.Decimal    `json:"rmb_cost" validate:"omitempty"`
	Detail   string              `json:"detail" validate:"omitempty"`
	Operator string              `json:"operator" validate:"max=64" `
}

// Validate ...
func (req *BillSyncRecordUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
