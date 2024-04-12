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

package cloudserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- List --------------------------

// AuditListReq define audit list req.
type AuditListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate audit list req.
func (req *AuditListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List Audit Async Flow --------------------------

// AuditAsyncFlowListReq define audit async flow list req.
type AuditAsyncFlowListReq struct {
	AuditID uint64 `json:"audit_id" validate:"required"`
	FlowID  string `json:"flow_id" validate:"required"`
}

// Validate validate audit async task list req.
func (req *AuditAsyncFlowListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List Audit Async Task --------------------------

// AuditAsyncTaskListReq define audit async task list req.
type AuditAsyncTaskListReq struct {
	AuditID  uint64 `json:"audit_id" validate:"required"`
	FlowID   string `json:"flow_id" validate:"required"`
	ActionID string `json:"action_id" validate:"required"`
}

// Validate validate audit async task list req.
func (req *AuditAsyncTaskListReq) Validate() error {
	return validator.Validate.Struct(req)
}
