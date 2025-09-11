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

package dataservice

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ApprovalProcessCreateReq ...
type ApprovalProcessCreateReq struct {
	ApplicationType enumor.ApplicationType `json:"application_type" validate:"required"`
	ServiceID       int64                  `json:"service_id" validate:"required,min=1"`
}

// Validate ...
func (req *ApprovalProcessCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApprovalProcessUpdateReq ...
type ApprovalProcessUpdateReq struct {
	ServiceID int64 `json:"service_id" validate:"required,min=1"`
}

// Validate ...
func (req *ApprovalProcessUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApprovalProcessListReq ...
type ApprovalProcessListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (l *ApprovalProcessListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ApprovalProcessResp ...
type ApprovalProcessResp struct {
	ID              string                 `json:"id"`
	ApplicationType enumor.ApplicationType `json:"application_type"`
	ServiceID       int64                  `json:"service_id"`
	WorkflowKey     string                 `json:"workflow_key"`
	Managers        string                 `json:"managers"`
	core.Revision   `json:",inline"`
}

// ApprovalProcessListResult defines list instances for iam pull resource callback result.
type ApprovalProcessListResult struct {
	Count uint64 `json:"count"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []*ApprovalProcessResp `json:"details"`
}

// ApprovalProcessListResp ...
type ApprovalProcessListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ApprovalProcessListResult `json:"data"`
}
