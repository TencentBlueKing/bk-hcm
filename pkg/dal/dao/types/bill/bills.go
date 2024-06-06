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
	tablebill "hcm/pkg/dal/table/bill"
)

// ListAccountBillSummaryDetails list account bill config details.
type ListAccountBillSummaryDetails struct {
	Count   *uint64                        `json:"count,omitempty"`
	Details []tablebill.AccountBillSummary `json:"details,omitempty"`
}

// ListAccountBillSummaryVersionDetails list account bill summary version details.
type ListAccountBillSummaryVersionDetails struct {
	Count   *uint64                               `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryVersion `json:"details,omitempty"`
}

// ListAccountBillSummaryDailyDetails list account bill summary daily details.
type ListAccountBillSummaryDailyDetails struct {
	Count   *uint64                             `json:"count,omitempty"`
	Details []tablebill.AccountBillSummaryDaily `json:"details,omitempty"`
}

// ListAccountBillItemDetails list account bill summary daily details.
type ListAccountBillItemDetails struct {
	Count   *uint64                     `json:"count,omitempty"`
	Details []tablebill.AccountBillItem `json:"details,omitempty"`
}

// ListAccountBillPullerDetails list account bill puller details
type ListAccountBillPullerDetails struct {
	Count   *uint64                       `json:"count,omitempty"`
	Details []tablebill.AccountBillPuller `json:"details,omitempty"`
}

// ListAccountBillDailyPullTaskDetails list account bill daily pull task details
type ListAccountBillDailyPullTaskDetails struct {
	Count   *uint64                              `json:"count,omitempty"`
	Details []tablebill.AccountBillDailyPullTask `json:"details,omitempty"`
}

// ListAccountBillAdjustmentItemDetails list account bill adjustment item details
type ListAccountBillAdjustmentItemDetails struct {
	Count   *uint64                               `json:"count,omitempty"`
	Details []tablebill.AccountBillAdjustmentItem `json:"details,omitempty"`
}
