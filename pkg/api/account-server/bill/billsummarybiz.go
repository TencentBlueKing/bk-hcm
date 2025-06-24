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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// BizSummaryListReq list request for Biz summary
type BizSummaryListReq struct {
	BillYear  int            `json:"bill_year" validate:"required"`
	BillMonth int            `json:"bill_month" validate:"required"`
	BKBizIDs  []int64        `json:"bk_biz_ids" validate:"required"`
	Page      *core.BasePage `json:"page" validate:"omitempty"`
}

// Validate ...
func (req *BizSummaryListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BizSummaryExportReq export request for biz summary
type BizSummaryExportReq struct {
	BillYear    int     `json:"bill_year" validate:"required"`
	BillMonth   int     `json:"bill_month" validate:"required"`
	ExportLimit uint64  `json:"export_limit" validate:"omitempty"`
	BKBizIDs    []int64 `json:"bk_biz_ids" validate:"required"`
}

// Validate ...
func (r *BizSummaryExportReq) Validate() error {
	if r.ExportLimit > constant.ExcelExportLimit {
		return errors.New("export limit exceed")
	}

	if len(r.BKBizIDs) > 1000 {
		return errors.New("bk biz ids exceed")
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
