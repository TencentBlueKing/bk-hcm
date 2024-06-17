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
	"hcm/pkg/criteria/constant"
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
