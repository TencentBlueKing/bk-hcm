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
	"errors"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// ExportBillItemReq ...
type ExportBillItemReq struct {
	BillYear    int                `json:"bill_year" validate:"required"`
	BillMonth   int                `json:"bill_month" validate:"required"`
	ExportLimit uint64             `json:"export_limit" validate:"omitempty"`
	Filter      *filter.Expression `json:"filter" validate:"omitempty"`
}

// Validate ListBillItemReq
func (r *ExportBillItemReq) Validate() error {
	if r.ExportLimit > constant.ExcelExportLimit {
		return errors.New("export limit exceed")
	}
	return validator.Validate.Struct(r)
}

// ListBillItemReq ...
type ListBillItemReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"omitempty"`
	Page      *core.BasePage     `json:"page" validate:"required"`
}

// Validate ListBillItemReq
func (r *ListBillItemReq) Validate() error {

	return validator.Validate.Struct(r)
}

// ImportBillAdjustmentReq 导入账单调整
type ImportBillAdjustmentReq struct {
	// 调账 文件上传
	ExcelFileBase64 string `json:"excel_file_base64" validate:"required"`
}

// Validate ListBillAdjustmentReq
func (r *ImportBillAdjustmentReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ListBillAdjustmentReq ...
type ListBillAdjustmentReq = core.ListReq

// BatchDeleteReq ...
type BatchDeleteReq struct {
	Ids []string `json:"ids" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *BatchDeleteReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AdjustmentItemResult wrapper for adjustment item
type AdjustmentItemResult struct {
	MainAccountCloudID string `json:"main_account_cloud_id"`
	MainAccountEmail   string `json:"main_account_email"`
	*bill.AdjustmentItem
}

// AdjustmentItemSumReq ...
type AdjustmentItemSumReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *AdjustmentItemSumReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AdjustmentItemSumResult all adjustment item summary get result
type AdjustmentItemSumResult struct {
	Count   uint64                                                                       `json:"count"`
	CostMap map[enumor.BillAdjustmentType]map[enumor.CurrencyCode]*bill.CostWithCurrency `json:"cost_map"`
}
