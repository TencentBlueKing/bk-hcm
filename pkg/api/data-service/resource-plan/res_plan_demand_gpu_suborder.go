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

// Package resourceplan ...
package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	dtypes "hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	ttypes "hcm/pkg/dal/table/types"
)

// ResPlanDemandGpuSubOrderBatchCreateReq create request.
type ResPlanDemandGpuSubOrderBatchCreateReq struct {
	SubOrders []ResPlanDemandGpuSubOrderCreateReq `json:"sub_orders" validate:"required,min=1,max=1000"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.SubOrders {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandGpuSubOrderCreateReq create request.
type ResPlanDemandGpuSubOrderCreateReq struct {
	OrderID       string                           `json:"order_id" validate:"required"`
	BkBizID       int64                            `json:"bk_biz_id" validate:"required"`
	OpProductID   int64                            `json:"op_product_id" validate:"required"`
	OpProductName string                           `json:"op_product_name" validate:"required,lte=64"`
	TemplateID    string                           `json:"template_id" validate:"required,lte=32"`
	DemandType    string                           `json:"demand_type" validate:"required"`
	DemandYear    int64                            `json:"demand_year" validate:"min=0"`
	DemandMonth   int64                            `json:"demand_month" validate:"min=1,max=12"`
	GPUNum        int64                            `json:"gpu_num" validate:"min=0"`
	QpmMax        int64                            `json:"qpm_max" validate:"min=0"`
	Status        enumor.RPDemandGPUSubOrderStatus `json:"status" validate:"required"`
	Comment       *ttypes.JsonField                `json:"comment"`
	Extension     ttypes.JsonField                 `json:"extension" validate:"required"`
	Remark        string                           `json:"remark" validate:"omitempty,max=255"`
	Creator       string                           `json:"creator" validate:"omitempty,max=64"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderCreateReq) Validate() error {
	if err := r.Status.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(r)
}

// ResPlanDemandGpuSubOrderBatchUpdateReq batch update request.
type ResPlanDemandGpuSubOrderBatchUpdateReq struct {
	SubOrders []ResPlanDemandGpuSubOrderUpdateReq `json:"sub_orders" validate:"required,min=1,max=100"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.SubOrders {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandGpuSubOrderUpdateReq batch update request.
type ResPlanDemandGpuSubOrderUpdateReq struct {
	ID            string                           `json:"id" validate:"required"`
	BkBizID       int64                            `json:"bk_biz_id"`
	OpProductID   int64                            `json:"op_product_id" validate:"omitempty"`
	OpProductName string                           `json:"op_product_name" validate:"omitempty,lte=64"`
	TemplateID    string                           `json:"template_id" validate:"omitempty,lte=32"`
	DemandType    string                           `json:"demand_type"`
	DemandYear    int64                            `json:"demand_year"`
	DemandMonth   int64                            `json:"demand_month"`
	GPUNum        int64                            `json:"gpu_num"`
	QpmMax        int64                            `json:"qpm_max"`
	Status        enumor.RPDemandGPUSubOrderStatus `json:"status"`
	Comment       *ttypes.JsonField                `json:"comment"`
	Extension     *ttypes.JsonField                `json:"extension"`
	Remark        string                           `json:"remark" validate:"omitempty,max=255"`
	Reviser       string                           `json:"reviser" validate:"omitempty,max=64"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderUpdateReq) Validate() error {
	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(r)
}

// ResPlanDemandGpuSubOrderBatchUpdateStatusReq batch update sub orders to the same status by id list.
type ResPlanDemandGpuSubOrderBatchUpdateStatusReq struct {
	IDs     []string                         `json:"ids" validate:"required,min=1,max=100"`
	Status  enumor.RPDemandGPUSubOrderStatus `json:"status" validate:"required"`
	Comment *ttypes.JsonField                `json:"comment"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderBatchUpdateStatusReq) Validate() error {
	if err := r.Status.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(r)
}

// ResPlanDemandGpuSubOrderListResult list result.
type ResPlanDemandGpuSubOrderListResult dtypes.ListResult[tablers.ResPlanDemandGpuSubOrderTable]

// ResPlanDemandGpuSubOrderListReq list request.
type ResPlanDemandGpuSubOrderListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderListReq) Validate() error {
	return r.ListReq.Validate()
}

// ResPlanDemandGpuSubOrderOverwriteReq atomically replaces all sub orders of an order with new ones.
type ResPlanDemandGpuSubOrderOverwriteReq struct {
	OrderID   string                              `json:"order_id" validate:"required"`
	SubOrders []ResPlanDemandGpuSubOrderCreateReq `json:"sub_orders" validate:"required,min=1,max=1000"`
}

// Validate validate.
func (r *ResPlanDemandGpuSubOrderOverwriteReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.SubOrders {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}
