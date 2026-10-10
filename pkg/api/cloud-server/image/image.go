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

// Package image defines cloud-server image API protocols.
package image

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// BizImageListReq is the typed request for biz-dimension image list query.
// Vendor is provided via path parameter. It does not accept a raw filter expression.
type BizImageListReq struct {
	// Platform is the image platform filter, e.g. CentOS / TencentOS.
	Platform string `json:"platform" validate:"omitempty"`
	// Name is the image name keyword for fuzzy match.
	Name string `json:"name" validate:"omitempty"`
	// Type is the normalized image type filter, enumeration values such as: public/private/shared.
	Type enumor.ImageTypeNormalized `json:"type" validate:"omitempty"`
	// Region is the region filter.
	Region string `json:"region" validate:"omitempty"`
	// Page is the standard pagination config.
	Page *core.BasePage `json:"page" validate:"required"`
}

// Validate BizImageListReq.
func (req *BizImageListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Type) > 0 {
		if err := req.Type.Validate(); err != nil {
			return err
		}
	}

	return req.Page.Validate()
}
