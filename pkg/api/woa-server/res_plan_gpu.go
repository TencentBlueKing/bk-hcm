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

package woaserver

import (
	"fmt"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/tools/excel"
)

// --- GPU demand excel import ---

// GpuDemandExcelImportResp GPU需求Excel导入预览响应
type GpuDemandExcelImportResp struct {
	Sheets  []excel.Sheet                `json:"sheets"`
	Details []GpuDemandExcelImportDetail `json:"details"`
}

// GpuDemandExcelImportDetail GPU需求Excel导入预览的单行数据
type GpuDemandExcelImportDetail struct {
	// Name sheet名称，对应tpl_schema中的sheet名称
	Name string `json:"name"`
	// RawData 原始行数据数组，按fixed_headers+headers中可见列顺序排列
	RawData []interface{} `json:"raw_data"`
	// ValidateResult 校验结果详情，空数组表示校验通过
	ValidateResult []string `json:"validate_result"`
}

// --- GPU demand order create ---

// CreateGpuDemandOrderReq 创建GPU需求提报主单请求
type CreateGpuDemandOrderReq struct {
	OpProductID   int64                        `json:"op_product_id" validate:"required"`
	OpProductName string                       `json:"op_product_name" validate:"required,lte=64"`
	Details       []CreateGpuDemandOrderDetail `json:"details" validate:"required"`
}

// Validate validates CreateGpuDemandOrderReq.
func (r *CreateGpuDemandOrderReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Details) == 0 {
		return fmt.Errorf("details count must be greater than 0")
	}

	// 限制子单的数量不超过1000
	if len(r.Details) > 1000 {
		return fmt.Errorf("details count must be less than 1000, now is %d", len(r.Details))
	}

	for i, d := range r.Details {
		if err := d.Validate(); err != nil {
			return fmt.Errorf("details[%d]: %w", i, err)
		}
	}

	return nil
}

// CreateGpuDemandOrderDetail 创建GPU需求提报子单明细
type CreateGpuDemandOrderDetail struct {
	DemandType  string `json:"demand_type" validate:"required,lte=64"`
	DemandYear  int64  `json:"demand_year" validate:"required"`
	DemandMonth int64  `json:"demand_month" validate:"required,min=1,max=12"`
	GPUNum      int64  `json:"gpu_num" validate:"min=0"`
	QpmMax      int64  `json:"qpm_max" validate:"min=0"`
	// Extension 扩展数据数组，按tpl_schema中headers定义的列顺序排列
	Extension types.JsonField `json:"extension" validate:"required"`
}

// Validate validates CreateGpuDemandOrderDetail.
func (d *CreateGpuDemandOrderDetail) Validate() error {
	if err := validator.Validate.Struct(d); err != nil {
		return err
	}

	return nil
}

// --- GPU demand order overwrite ---

// OverwriteGpuDemandOrderReq 覆盖上传GPU需求提报主单请求
type OverwriteGpuDemandOrderReq struct {
	OrderID string                       `json:"order_id" validate:"required,lte=64"`
	Details []CreateGpuDemandOrderDetail `json:"details" validate:"required"`
}

// Validate validates OverwriteGpuDemandOrderReq.
func (r *OverwriteGpuDemandOrderReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Details) == 0 {
		return fmt.Errorf("details count must be greater than 0")
	}

	// 限制子单的数量不超过1000
	if len(r.Details) > 1000 {
		return fmt.Errorf("details count must be less than 1000, now is %d", len(r.Details))
	}

	for i, d := range r.Details {
		if err := d.Validate(); err != nil {
			return fmt.Errorf("details[%d]: %w", i, err)
		}
	}

	return nil
}
