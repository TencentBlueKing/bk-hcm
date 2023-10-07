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

package recyclerecord

import (
	rr "hcm/pkg/api/core/recycle-record"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Recycle --------------------------

// BatchRecycleReq defines batch recycle cloud resource request.
type BatchRecycleReq struct {
	ResType            enumor.CloudResourceType `json:"resource_type" validate:"required"`
	DefaultRecycleTime uint                     `json:"default_recycle_time" validate:"required"`
	RecycleType        enumor.RecycleType       `json:"recycle_type,omitempty"`
	Infos              []RecycleReq             `json:"infos" validate:"min=1,max=100"`
}

// RecycleReq defines recycle one cloud resource request.
type RecycleReq struct {
	ID     string      `json:"id" validate:"required"`
	Detail interface{} `json:"detail" validate:"required"`
}

// Validate BatchRecycleReq.
func (c *BatchRecycleReq) Validate() error {
	return validator.Validate.Struct(c)
}

// RecycleResp defines recycle cloud resource response.
type RecycleResp struct {
	rest.BaseResp `json:",inline"`
	Data          string `json:"data"`
}

// -------------------------- Recover --------------------------

// BatchRecoverReq defines batch recover cloud resource request.
type BatchRecoverReq struct {
	ResType   enumor.CloudResourceType `json:"res_type" validate:"required"`
	RecordIDs []string                 `json:"record_ids" validate:"min=1,max=100"`
}

// Validate BatchRecoverReq.
func (c *BatchRecoverReq) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- List --------------------------

// ListResp defines list recycle record response.
type ListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListResult `json:"data"`
}

// ListResult defines list recycle record result.
type ListResult struct {
	Count   uint64             `json:"count"`
	Details []rr.RecycleRecord `json:"details"`
}

// -------------------------- Update --------------------------

// BatchUpdateReq defines batch update recycle record request.
type BatchUpdateReq struct {
	Data []UpdateReq `json:"data" validate:"min=1,max=100"`
}

// UpdateReq defines update recycle record request.
type UpdateReq struct {
	ID     string                     `json:"id" validate:"required"`
	Status enumor.RecycleRecordStatus `json:"status" validate:"omitempty"`
	Detail interface{}                `json:"detail" validate:"omitempty"`
}

// Validate BatchUpdateReq.
func (c *BatchUpdateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// BatchUpdateRecycleStatusReq defines batch update cloud resource recycle status request.
type BatchUpdateRecycleStatusReq struct {
	ResType       enumor.CloudResourceType   `json:"res_type" validate:"required"`
	IDs           []string                   `json:"ids" validate:"min=1,max=100"`
	RecycleStatus enumor.RecycleRecordStatus `json:"recycle_status"  validate:"required"`
}

// Validate BatchRecoverReq.
func (c *BatchUpdateRecycleStatusReq) Validate() error {
	return validator.Validate.Struct(c)
}
