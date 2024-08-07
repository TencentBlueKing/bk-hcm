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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"

	"github.com/shopspring/decimal"
)

// SyncRecord 同步记录
type SyncRecord struct {
	ID        string               `json:"id,omitempty"`
	Vendor    enumor.Vendor        `json:"vendor"`
	BillYear  int                  `json:"bill_year"`
	BillMonth int                  `json:"bill_month"`
	State     enumor.BillSyncState `json:"state"`
	Currency  enumor.CurrencyCode  `json:"currency" `
	Cost      decimal.Decimal      `json:"cost"`
	RMBCost   decimal.Decimal      `json:"rmb_cost"`
	Detail    string               `json:"detail"`
	Operator  string               `json:"operator"`
	core.Revision
}
