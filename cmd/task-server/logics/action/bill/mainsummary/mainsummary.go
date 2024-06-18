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

package mainsummary

import (
	"hcm/pkg/criteria/enumor"
)

// MainAccountSummaryActionOption option for main account summary action
type MainAccountSummaryActionOption struct {
	RootAccountID string        `json:"root_account_id" validate:"required"`
	MainAccountID string        `json:"main_account_id" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	VersionID     int           `json:"version_id" validate:"required"`
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
}

// MainAccountSummaryAction define main account summary action
type MainAccountSummaryAction struct{}

// ParameterNew return request params.
func (act MainAccountSummaryAction) ParameterNew() interface{} {
	return new(MainAccountSummaryActionOption)
}

// Name return action name
func (act MainAccountSummaryAction) Name() enumor.ActionName {
	return enumor.ActionMainAccountSummary
}
