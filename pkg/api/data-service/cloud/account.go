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

// Package cloud 包提供各类云资源的请求与返回序列化器
package cloud

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateAccountReq ...
type CreateAccountReq struct {
	Vendor     enumor.Vendor            `json:"vendor" validate:"required"`
	Spec       *cloud.AccountSpec       `json:"spec" validate:"required"`
	Extension  *cloud.AccountExtension  `json:"extension" validate:"required"`
	Attachment *cloud.AccountAttachment `json:"attachment" validate:"required"`
}

// Validate ...
func (c *CreateAccountReq) Validate() error {
	return validator.Validate.Struct(c)
}

// UpdateAccountReq ...
type UpdateAccountReq struct {
	Spec      *cloud.AccountSpec      `json:"spec" validate:"required"`
	Extension *cloud.AccountExtension `json:"extension" validate:"required"`
	Filter    *filter.Expression      `json:"filter" validate:"required"`
}

// Validate ...
func (u *UpdateAccountReq) Validate() error {
	return validator.Validate.Struct(u)
}

// ListAccountReq ...
type ListAccountReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate ...
func (l *ListAccountReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ListAccountResult defines list instances for iam pull resource callback result.
type ListAccountResult struct {
	Count   uint64           `json:"count,omitempty"`
	Details []*cloud.Account `json:"details,omitempty"`
}

// DeleteAccountReq ...
type DeleteAccountReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (d *DeleteAccountReq) Validate() error {
	return validator.Validate.Struct(d)
}

// ListAccountResp ...
type ListAccountResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListAccountResult `json:"data"`
}

// UpdateAccountBizRelReq ...
type UpdateAccountBizRelReq struct {
	AccountID uint64   `json:"account_id" validate:"required"`
	BkBizIDs  []uint64 `json:"bk_biz_ids" validate:"required"`
}

// Validate ...
func (req *UpdateAccountBizRelReq) Validate() error {
	return validator.Validate.Struct(req)
}
