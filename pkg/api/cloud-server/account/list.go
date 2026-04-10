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
	"hcm/pkg/api/data-service/cloud"
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

// AccountBizListReq 业务级账号列表查询请求
type AccountBizListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate 验证请求参数
func (req *AccountBizListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	return req.Page.Validate()
}

// AccountWithOtherInfo 带其他信息的账号数据
type AccountWithOtherInfo struct {
	*cloud.BaseAccountWithExtensionListResp `json:",inline"`
	SubAccountCount                         uint64 `json:"sub_account_count"`
	AccountSecretCount                      uint64 `json:"account_secret_count"`
}

// AccountBizListResult 业务级账号列表响应结果
type AccountBizListResult struct {
	Count   uint64                  `json:"count"`
	Details []*AccountWithOtherInfo `json:"details"`
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
