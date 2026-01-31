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

package cloud

import (
	"hcm/pkg/api/core"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/cryptography"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ------------------------ Extension 加密接口 ------------------------

// AccountSecretExtension 账号密钥扩展接口（用于加密）
type AccountSecretExtension interface {
	EncryptSecretKey(cipher cryptography.Crypto)
}

// TCloudAccountSecretExtension 腾讯云账号密钥扩展（实现加密接口）
type TCloudAccountSecretExtension struct {
	coreas.TCloudAccountSecretExtension
}

// EncryptSecretKey 加密密钥
func (ext *TCloudAccountSecretExtension) EncryptSecretKey(cipher cryptography.Crypto) {
	ext.CloudSecretKey = cipher.EncryptToBase64(ext.CloudSecretKey)
}

// ------------------------ Batch Create ------------------------

// AccountSecretBatchCreateReq defines batch create account secret request.
type AccountSecretBatchCreateReq[T coreas.Extension] struct {
	AccountSecrets []AccountSecretCreate[T] `json:"account_secrets" validate:"required,min=1,max=100"`
}

// AccountSecretCreate defines create account secret.
type AccountSecretCreate[T coreas.Extension] struct {
	AccountID string                     `json:"account_id" validate:"required"`
	Type      enumor.AccountSecretType   `json:"type" validate:"required"`
	Status    enumor.AccountSecretStatus `json:"status" validate:"required"`
	Extension *T                         `json:"extension" validate:"required"`
}

// Validate account secret batch create request.
func (req *AccountSecretBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ Batch Update ------------------------

// AccountSecretBatchUpdateReq defines batch update account secret request.
type AccountSecretBatchUpdateReq[T coreas.Extension] struct {
	AccountSecrets []AccountSecretUpdate[T] `json:"account_secrets" validate:"required,min=1,max=100"`
}

// AccountSecretUpdate defines update account secret.
type AccountSecretUpdate[T coreas.Extension] struct {
	ID        string                      `json:"id" validate:"required"`
	Type      *enumor.AccountSecretType   `json:"type" validate:"omitempty"`
	Status    *enumor.AccountSecretStatus `json:"status,omitempty"`
	Extension *T                          `json:"extension,omitempty"`
}

// Validate account secret batch update request.
func (req *AccountSecretBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ Batch Delete ------------------------

// AccountSecretBatchDeleteReq defines batch delete account secret request.
type AccountSecretBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate account secret batch delete request.
func (req *AccountSecretBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ List ------------------------

// AccountSecretListReq defines list account secret request.
type AccountSecretListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate account secret list request.
func (req *AccountSecretListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AccountSecretListResult defines list account secret result.
type AccountSecretListResult struct {
	Count   uint64                     `json:"count"`
	Details []coreas.BaseAccountSecret `json:"details"`
}

// AccountSecretListResp defines list account secret response.
type AccountSecretListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountSecretListResult `json:"data"`
}

// AccountSecretExtListReq defines list account secret with extension request.
type AccountSecretExtListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate account secret list with extension request.
func (req *AccountSecretExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AccountSecretExtListResult defines list account secret with extension result.
type AccountSecretExtListResult[T coreas.Extension] struct {
	Count   uint64                    `json:"count"`
	Details []coreas.AccountSecret[T] `json:"details"`
}

// AccountSecretExtListResp defines list account secret with extension response.
type AccountSecretExtListResp[T coreas.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountSecretExtListResult[T] `json:"data"`
}
