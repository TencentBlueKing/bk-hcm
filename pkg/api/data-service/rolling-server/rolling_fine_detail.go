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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	rs "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// BatchCreateRollingFineDetailReq batch create request
type BatchCreateRollingFineDetailReq struct {
	FineDetails []RollingFineDetailCreateReq `json:"fine_details" validate:"required,max=100"`
}

// Validate ...
func (c *BatchCreateRollingFineDetailReq) Validate() error {
	if len(c.FineDetails) == 0 || len(c.FineDetails) > 100 {
		return errf.Newf(errf.InvalidParameter, "fine_details count should between 1 and 100")
	}
	for _, item := range c.FineDetails {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(c)
}

// RollingFineDetailCreateReq create request
type RollingFineDetailCreateReq struct {
	BkBizID              int64           `json:"bk_biz_id" validate:"required"`
	AppliedRecordID      string          `json:"applied_record_id" validate:"required"`
	OrderID              uint64          `json:"order_id" validate:"required"`
	SubOrderID           string          `json:"suborder_id" validate:"required"`
	Year                 int             `json:"year" validate:"required"`
	Month                int             `json:"month" validate:"required"`
	Day                  int             `json:"day" validate:"required"`
	DeliveredCore        uint64          `json:"delivered_core"`
	ReturnedCore         uint64          `json:"returned_core"`
	ExemptedReturnedCore uint64          `json:"exempted_returned_core"`
	Fine                 decimal.Decimal `json:"fine"`
}

// Validate ...
func (c *RollingFineDetailCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// RollingFineDetailListReq list request
type RollingFineDetailListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *RollingFineDetailListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RollingFineDetailListResult list result
type RollingFineDetailListResult = core.ListResultT[*rs.RollingFineDetailTable]
