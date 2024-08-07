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

package bill

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tablebill "hcm/pkg/dal/table/bill"
)

// ListAccountBillSummaryMainDetails list account bill config details.
type ListAccountBillSummaryMainDetails struct {
	Count   uint64                             `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryMain `json:"details,omitempty"`
}

// ListAccountBillSummaryVersionDetails list account bill summary version details.
type ListAccountBillSummaryVersionDetails struct {
	Count   uint64                                `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryVersion `json:"details,omitempty"`
}

// ListAccountBillSummaryDailyDetails list account bill summary daily details.
type ListAccountBillSummaryDailyDetails struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryDaily `json:"details,omitempty"`
}

// ListAccountBillItemDetails list account bill summary daily details.
type ListAccountBillItemDetails struct {
	Count   uint64                      `json:"count,omitempty"`
	Details []tablebill.AccountBillItem `json:"details,omitempty"`
}

// ListAccountBillMonthPullTaskDetails list account bill month pull details
type ListAccountBillMonthPullTaskDetails struct {
	Count   uint64                           `json:"count,omitempty"`
	Details []tablebill.AccountBillMonthTask `json:"details,omitempty"`
}

// ListAccountBillDailyPullTaskDetails list account bill daily pull task details
type ListAccountBillDailyPullTaskDetails struct {
	Count   uint64                               `json:"count,omitempty"`
	Details []tablebill.AccountBillDailyPullTask `json:"details,omitempty"`
}

// ListAccountBillAdjustmentItemDetails list account bill adjustment item details
type ListAccountBillAdjustmentItemDetails struct {
	Count   uint64                                `json:"count,omitempty"`
	Details []tablebill.AccountBillAdjustmentItem `json:"details,omitempty"`
}

// ListAccountBillSummaryRootDetails list account bill adjustment item details
type ListAccountBillSummaryRootDetails struct {
	Count   uint64                             `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryRoot `json:"details,omitempty"`
}

// ListRootAccountBillConfigDetails list account bill config details.
type ListRootAccountBillConfigDetails struct {
	Count   uint64                                 `json:"count,omitempty"`
	Details []tablebill.RootAccountBillConfigTable `json:"details,omitempty"`
}

// ListAccountBillExchangeRateDetails list account bill adjustment item details
type ListAccountBillExchangeRateDetails struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []tablebill.AccountBillExchangeRate `json:"details,omitempty"`
}

// ListAccountBillSyncRecordDetails list account bill sync record details
type ListAccountBillSyncRecordDetails struct {
	Count   uint64                            `json:"count,omitempty"`
	Details []tablebill.AccountBillSyncRecord `json:"details,omitempty"`
}

// ItemCommonOpt  bill item table partition parameters
type ItemCommonOpt struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	Year   int           `json:"year" validate:"required"`
	Month  int           `json:"month" validate:"required,min=1,max=12"`
}

// Validate ...
func (p *ItemCommonOpt) Validate() error {
	if p == nil {
		return errf.New(errf.InvalidParameter, "bill partition params is required")
	}
	if len(p.Vendor) == 0 {
		return errf.New(errf.InvalidParameter, "vendor is required")
	}
	if p.Year == 0 {
		return errf.New(errf.InvalidParameter, "year is required")
	}
	if p.Month == 0 {
		return errf.New(errf.InvalidParameter, "month is required")
	}
	if p.Month > 12 || p.Month < 0 {
		return errf.New(errf.InvalidParameter, "month must between 1 and 12")
	}
	return nil
}
