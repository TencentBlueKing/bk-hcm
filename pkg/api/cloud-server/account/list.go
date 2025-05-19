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

package account

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// AccountListReq ...
type AccountListReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"omitempty"`
}

// Validate ...
func (req *AccountListReq) Validate() error {
	if req.Page != nil {
		if err := req.Page.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// AccountListResourceReq ...
type AccountListResourceReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *AccountListResourceReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AccountListWithExtReq ...
type AccountListWithExtReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *AccountListWithExtReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListSecretKeyReq ...
type ListSecretKeyReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate ...
func (req *ListSecretKeyReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SecretKeyData ...
type SecretKeyData struct {
	ID        string `json:"id"`
	SecretKey string `json:"secret_key"`
}
