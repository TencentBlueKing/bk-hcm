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

package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	rpgpu "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
)

// ResPlanDemandGpuOrderBatchCreateReq batch create request.
type ResPlanDemandGpuOrderBatchCreateReq struct {
	Items []ResPlanDemandGpuOrderCreateReq `json:"items" validate:"required,min=1,max=100"`
}

// Validate validate.
func (r *ResPlanDemandGpuOrderBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandGpuOrderCreateReq create request.
type ResPlanDemandGpuOrderCreateReq struct {
	BkBizID       int64                              `json:"bk_biz_id" validate:"required"`
	OpProductID   int64                              `json:"op_product_id" validate:"required"`
	OpProductName string                             `json:"op_product_name" validate:"required,lte=64"`
	TemplateID    string                             `json:"template_id" validate:"required,lte=32"`
	Status        enumor.ResPlanDemandGpuOrderStatus `json:"status" validate:"required"`
	Remark        string                             `json:"remark" validate:"omitempty,lte=255"`
}

// Validate validate.
func (r *ResPlanDemandGpuOrderCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	return r.Status.Validate()
}

// ResPlanDemandGpuOrderBatchUpdateReq batch update request.
type ResPlanDemandGpuOrderBatchUpdateReq struct {
	Items []ResPlanDemandGpuOrderUpdateReq `json:"items" validate:"required,min=1,max=100"`
}

// Validate validate.
func (r *ResPlanDemandGpuOrderBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, item := range r.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanDemandGpuOrderUpdateReq update request.
type ResPlanDemandGpuOrderUpdateReq struct {
	ID            string                             `json:"id" validate:"required,lte=64"`
	OpProductID   int64                              `json:"op_product_id" validate:"omitempty"`
	OpProductName string                             `json:"op_product_name" validate:"omitempty,lte=64"`
	TemplateID    string                             `json:"template_id" validate:"omitempty,lte=32"`
	Status        enumor.ResPlanDemandGpuOrderStatus `json:"status" validate:"omitempty"`
	Remark        string                             `json:"remark" validate:"omitempty,lte=255"`
}

// Validate validate.
func (r *ResPlanDemandGpuOrderUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Status) > 0 {
		return r.Status.Validate()
	}

	return nil
}

// ResPlanDemandGpuOrderListReq list request.
type ResPlanDemandGpuOrderListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate.
func (r *ResPlanDemandGpuOrderListReq) Validate() error {
	return r.ListReq.Validate()
}

// ResPlanDemandGpuOrderListResult list result.
type ResPlanDemandGpuOrderListResult types.ListResult[rpgpu.ResPlanDemandGpuOrderTable]

