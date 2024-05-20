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

// ApplicationCreateReq ...
type ApplicationCreateReq struct {
	Source         enumor.ApplicationSource `json:"source" validate:"required"`
	SN             string                   `json:"sn" validate:"required"`
	Type           enumor.ApplicationType   `json:"type" validate:"required"`
	Status         enumor.ApplicationStatus `json:"status" validate:"required"`
	Applicant      string                   `json:"applicant" validate:"required"`
	Content        string                   `json:"content" validate:"required"`
	DeliveryDetail string                   `json:"delivery_detail" validate:"required"`
	Memo           *string                  `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *ApplicationCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApplicationUpdateReq ...
type ApplicationUpdateReq struct {
	Status         enumor.ApplicationStatus `json:"status" validate:"required"`
	DeliveryDetail *string                  `json:"delivery_detail" validate:"omitempty"`
}

// Validate ...
func (req *ApplicationUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ApplicationResp ...
type ApplicationResp struct {
	ID             string                   `json:"id"`
	Source         enumor.ApplicationSource `json:"source"`
	SN             string                   `json:"sn"`
	Type           enumor.ApplicationType   `json:"type"`
	Status         enumor.ApplicationStatus `json:"status"`
	Applicant      string                   `json:"applicant"`
	Content        string                   `json:"content"`
	DeliveryDetail string                   `json:"delivery_detail"`
	Memo           *string                  `json:"memo"`
	core.Revision  `json:",inline"`
}

// ApplicationGetResp ...
type ApplicationGetResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ApplicationResp `json:"data"`
}

// ApplicationListReq ...
type ApplicationListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (l *ApplicationListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ApplicationListResult defines list instances for iam pull resource callback result.
type ApplicationListResult struct {
	Count uint64 `json:"count"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []*ApplicationResp `json:"details"`
}

// ApplicationListResp ...
type ApplicationListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ApplicationListResult `json:"data"`
}
