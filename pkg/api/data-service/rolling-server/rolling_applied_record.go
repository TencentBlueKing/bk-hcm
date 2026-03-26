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

// Package rollingserver ...
package rollingserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	rs "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/runtime/filter"
)

// BatchCreateRollingAppliedRecordReq batch create request
type BatchCreateRollingAppliedRecordReq struct {
	AppliedRecords []RollingAppliedRecordCreateReq `json:"applied_records" validate:"required,max=100"`
}

// Validate ...
func (c *BatchCreateRollingAppliedRecordReq) Validate() error {
	if len(c.AppliedRecords) == 0 || len(c.AppliedRecords) > 100 {
		return errf.Newf(errf.InvalidParameter, "applied_records count should between 1 and 100")
	}
	for _, item := range c.AppliedRecords {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(c)
}

// RollingAppliedRecordCreateReq create request
type RollingAppliedRecordCreateReq struct {
	RequireType   enumor.RequireType `json:"require_type" validate:"required"`
	AppliedType   enumor.AppliedType `json:"applied_type" validate:"required"`
	BkBizID       int64              `json:"bk_biz_id" validate:"required"`
	OrderID       uint64             `json:"order_id" validate:"required"`
	SubOrderID    string             `json:"suborder_id" validate:"required"`
	Year          int                `json:"year" validate:"required"`
	Month         int                `json:"month" validate:"required"`
	Day           int                `json:"day" validate:"required"`
	AppliedCore   int64              `json:"applied_core" validate:"required"`
	DeliveredCore int64              `json:"delivered_core" validate:"omitempty"`
	InstanceGroup string             `json:"instance_group" validate:"required"`
	CoreType      enumor.CoreType    `json:"core_type" validate:"required"`
	NotNotice     bool               `json:"not_notice" validate:"omitempty"`
}

// Validate ...
func (c *RollingAppliedRecordCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// RollingAppliedRecordListReq list request
type RollingAppliedRecordListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *RollingAppliedRecordListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RollingAppliedRecordListResult list result
type RollingAppliedRecordListResult = core.ListResultT[*rs.RollingAppliedRecord]

// BatchUpdateRollingAppliedRecordReq batch update request
type BatchUpdateRollingAppliedRecordReq struct {
	AppliedRecords []RollingAppliedRecordUpdateReq `json:"applied_records" validate:"required,max=100"`
}

// Validate ...
func (c *BatchUpdateRollingAppliedRecordReq) Validate() error {
	if len(c.AppliedRecords) == 0 || len(c.AppliedRecords) > 100 {
		return errf.Newf(errf.InvalidParameter, "applied_records count should between 1 and 100")
	}
	for _, item := range c.AppliedRecords {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(c)
}

// RollingAppliedRecordUpdateReq update request
type RollingAppliedRecordUpdateReq struct {
	ID                   string             `json:"id" validate:"required"`
	AppliedType          enumor.AppliedType `json:"applied_type" validate:"omitempty"`
	AppliedCore          *int64             `json:"applied_core" validate:"omitempty"`
	DeliveredCore        *int64             `json:"delivered_core" validate:"omitempty"`
	NotNotice            *bool              `json:"not_notice" validate:"omitempty"`
	ExemptedReturnedCore *int64             `json:"exempted_returned_core" validate:"omitempty,gte=0"`
}

// Validate ...
func (req *RollingAppliedRecordUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RollingCpuCoreSummaryResult get rolling cpu core summary result
type RollingCpuCoreSummaryResult = core.BaseResp[*RollingCpuCoreSummaryItem]

// RollingCpuCoreSummaryItem wrapper for rolling cpu core summary item
type RollingCpuCoreSummaryItem struct {
	SumDeliveredCore       int64 `json:"sum_delivered_core" db:"sum_delivered_core"`
	SumReturnedAppliedCore int64 `json:"sum_returned_applied_core" db:"sum_returned_applied_core"`
}

// AppliedRecordUpdateNoticeStateReq update request
type AppliedRecordUpdateNoticeStateReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate ...
func (req *AppliedRecordUpdateNoticeStateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AppliedRecordUpdateExemptedReturnedCoreReq update request
type AppliedRecordUpdateExemptedReturnedCoreReq struct {
	IDs                  []string `json:"ids" validate:"required,min=1,max=100"`
	ExemptedReturnedCore int64    `json:"exempted_returned_core" validate:"gte=0"`
}

// Validate ...
func (req *AppliedRecordUpdateExemptedReturnedCoreReq) Validate() error {
	return validator.Validate.Struct(req)
}
