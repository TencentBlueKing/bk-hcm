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

package registry

import (
	accountset "hcm/pkg/api/core/account-set"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"

	"github.com/shopspring/decimal"
)

// PullDailyBillOption define daily bill pull option
type PullDailyBillOption struct {
	RootAccountID string `json:"root_account_id" validate:"required"`
	MainAccountID string `json:"main_account_id" validate:"required"`
	// 主账号云id
	BillAccountID string        `json:"bill_account_id" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	BillDay       int           `json:"bill_day" validate:"required"`
	VersionID     int           `json:"version_id" validate:"required"`
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
	MainAccount   *accountset.BaseMainAccount
}

// PullerRegistry registry of pullers
var PullerRegistry = make(map[enumor.Vendor]Puller)

// PullerResult puller result
type PullerResult struct {
	// Count 账单条目数量
	Count int64 `db:"count" json:"count"`
	// Currency 币种
	Currency enumor.CurrencyCode `db:"currency" json:"currency"`
	// Cost 金额，单位：元
	Cost decimal.Decimal `db:"cost" json:"cost"`
}

// Puller puller interface
type Puller interface {
	Pull(kt run.ExecuteKit, opt *PullDailyBillOption) (*PullerResult, error)
}
