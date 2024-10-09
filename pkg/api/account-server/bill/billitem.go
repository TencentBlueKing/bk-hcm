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
	"encoding/json"
	"errors"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/runtime/filter"
)

// ExportBillItemReq ...
type ExportBillItemReq struct {
	BillYear    int                `json:"bill_year" validate:"required"`
	BillMonth   int                `json:"bill_month" validate:"required"`
	ExportLimit uint64             `json:"export_limit" validate:"required"`
	Filter      *filter.Expression `json:"filter" validate:"omitempty"`
}

// Validate ListBillItemReq
func (r *ExportBillItemReq) Validate() error {
	if r.ExportLimit > constant.ExcelExportLimit {
		return errors.New("export limit exceed")
	}
	if r.Filter != nil {
		err := r.Filter.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillItemColumns.ColumnTypes())))
		if err != nil {
			return err
		}
	}
	if r.BillYear == 0 {
		return errors.New("year is required")
	}
	if r.BillMonth == 0 {
		return errors.New("month is required")
	}
	if r.BillMonth > 12 || r.BillMonth < 0 {
		return errors.New("month must between 1 and 12")
	}
	return validator.Validate.Struct(r)
}

// ListBillItemReq ...
type ListBillItemReq struct {
	BillYear  int                `json:"bill_year" validate:"required"`
	BillMonth int                `json:"bill_month" validate:"required"`
	Filter    *filter.Expression `json:"filter" validate:"required"`
	Page      *core.BasePage     `json:"page" validate:"required"`
}

// Validate ListBillItemReq
func (r *ListBillItemReq) Validate() error {

	return validator.Validate.Struct(r)
}

// ListBillItemByVendorReq ...
type ListBillItemByVendorReq struct {
	BillYear  uint `json:"bill_year" validate:"required"`
	BillMonth uint `json:"bill_month" validate:"required,min=1,max=12"`

	RootAccountIds      []string `json:"root_account_ids" validate:"max=10"`
	RootAccountCloudIds []string `json:"root_account_cloud_ids" validate:"max=10"`
	MainAccountIds      []string `json:"main_account_ids" validate:"max=10"`
	MainAccountCloudIds []string `json:"main_account_cloud_ids" validate:"max=10"`

	BeginBillDay *uint `json:"begin_bill_day" validate:"omitempty,min=1,max=31"`
	EndBillDay   *uint `json:"end_bill_day" validate:"omitempty,min=1,max=31"`

	Page *core.BasePage `json:"page" validate:"required"`
}

// Validate ListBillItemReq
func (r *ListBillItemByVendorReq) Validate() error {
	if r.BeginBillDay != nil && r.EndBillDay == nil || r.BeginBillDay == nil && r.EndBillDay != nil {
		return errf.New(errf.InvalidParameter, "begin_bill_day and end_bill_day must be both empty or not empty")
	}
	if r.BeginBillDay != nil && r.EndBillDay != nil && *r.BeginBillDay > *r.EndBillDay {
		return errf.New(errf.InvalidParameter, "begin_bill_day must be less or equal than end_bill_day")
	}

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

const (
	importFileSizeMaxLimit1MB = 1 * 1024 * 1024
	importBillItemsMaxLimit   = 200000
)

// Base64String base64 string
type Base64String string

func (b Base64String) checkSize(expectedSize int) error {
	if len(b) > expectedSize {
		return errors.New("file size exceed limit")
	}
	return nil
}

// ImportBillItemPreviewReq 账单明细预览
type ImportBillItemPreviewReq struct {
	BillYear  int `json:"bill_year" validate:"required"`
	BillMonth int `json:"bill_month" validate:"required"`
	// 调账 文件上传
	ExcelFileBase64 Base64String `json:"excel_file_base64" validate:"required"`
}

// Validate ...
func (r *ImportBillItemPreviewReq) Validate() error {
	if err := r.ExcelFileBase64.checkSize(importFileSizeMaxLimit1MB); err != nil {
		return err
	}
	return validator.Validate.Struct(r)
}

// ImportBillItemPreviewResult 账单明细预览结果
type ImportBillItemPreviewResult struct {
	Items   []dsbill.BillItemCreateReq[json.RawMessage]    `json:"items" validate:"required"`
	CostMap map[enumor.CurrencyCode]*bill.CostWithCurrency `json:"cost_map"`
}

// ImportBillItemReq 账单明细
type ImportBillItemReq struct {
	BillYear  int                                         `json:"bill_year" validate:"required"`
	BillMonth int                                         `json:"bill_month" validate:"required"`
	Items     []dsbill.BillItemCreateReq[json.RawMessage] `json:"items" validate:"required"`
}

// Validate ...
func (r *ImportBillItemReq) Validate() error {
	if len(r.Items) == 0 {
		return errf.New(errf.InvalidParameter, "items is empty")
	}
	if len(r.Items) > importBillItemsMaxLimit {
		return errf.New(errf.InvalidParameter, "items count exceed limit importBillItemsMaxLimit")
	}
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

// FileDownloadResp define file download resp.
type FileDownloadResp struct {
	ContentTypeStr        string
	ContentDispositionStr string
	FilePath              string
}

// ContentType ...
func (f *FileDownloadResp) ContentType() string {
	return f.ContentTypeStr
}

// ContentDisposition ...
func (f *FileDownloadResp) ContentDisposition() string {
	return f.ContentDispositionStr
}

// Filepath return file path.
func (f *FileDownloadResp) Filepath() string {
	return f.FilePath
}

// IsDeleteFile is true, file will be deleted after download.
func (f *FileDownloadResp) IsDeleteFile() bool {
	return true
}
