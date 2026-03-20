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

package plan

import (
	"errors"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"
)

// BatchGpuOrderStatusReq 批量变更GPU需求主单状态请求
type BatchGpuOrderStatusReq struct {
	OrderIDs []string `json:"order_ids" validate:"required,min=1,max=100"`
}

// Validate 校验请求参数
func (r *BatchGpuOrderStatusReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(r.OrderIDs))
	for _, id := range r.OrderIDs {
		if _, ok := seen[id]; ok {
			return errors.New("order_ids contains duplicate values")
		}
		seen[id] = struct{}{}
	}

	return nil
}

// ListGpuDemandOrderReq GPU需求主单列表查询请求
type ListGpuDemandOrderReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate 校验请求参数
func (r *ListGpuDemandOrderReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	return r.Page.Validate()
}

// GpuDemandOrderItem GPU需求主单列表响应单条记录（含子单聚合字段）
type GpuDemandOrderItem struct {
	ID            string                             `json:"id"`
	BkBizID       int64                              `json:"bk_biz_id"`
	OpProductID   int64                              `json:"op_product_id"`
	OpProductName string                             `json:"op_product_name"`
	TemplateID    string                             `json:"template_id"`
	Status        enumor.ResPlanDemandGpuOrderStatus `json:"status"`
	Remark        string                             `json:"remark"`
	// TotalGpuNum 需求卡数，由关联子单的 gpu_num 汇总求和得出
	TotalGpuNum int64 `json:"total_gpu_num"`
	// TotalQpmMax QPM（月调用峰值），由关联子单的 qpm_max 汇总求和得出
	TotalQpmMax int64      `json:"total_qpm_max"`
	Creator     string     `json:"creator"`
	Reviser     string     `json:"reviser"`
	CreatedAt   types.Time `json:"created_at"`
	UpdatedAt   types.Time `json:"updated_at"`
}

// ListGpuDemandOrderResult GPU需求主单列表查询响应
type ListGpuDemandOrderResult struct {
	Count   uint64               `json:"count"`
	Details []GpuDemandOrderItem `json:"details"`
}
