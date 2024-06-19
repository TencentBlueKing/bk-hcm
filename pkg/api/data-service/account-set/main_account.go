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

// Package accountset
package accountset

import (
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/cryptography"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------
// MainAccountExtensionCreateReq main account extension create req.
type MainAccountExtensionCreateReq interface {
	AwsMainAccountExtensionCreateReq | GcpMainAccountExtensionCreateReq |
		AzureMainAccountExtensionCreateReq | HuaWeiMainAccountExtensionCreateReq |
		ZenlayerMainAccountExtensionCreateReq | KaopuMainAccountExtensionCreateReq
}

// AwsMainAccountExtensionCreateReq ...
type AwsMainAccountExtensionCreateReq struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// EncryptSecretKey encrypt secret key
func (req *AwsMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudInitPassword = cipher.EncryptToBase64(req.CloudInitPassword)
}

// GcpMainAccountExtensionCreateReq ...
type GcpMainAccountExtensionCreateReq struct {
	CloudProjectID   string `json:"cloud_project_id"`
	CloudProjectName string `json:"cloud_project_name"`
}

// EncryptSecretKey encrypt secret key
func (req *GcpMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	// nothing to encrypt
}

// AzureMainAccountExtensionCreateReq ...
type AzureMainAccountExtensionCreateReq struct {
	CloudSubscriptionID   string `json:"cloud_subscription_id"`
	CloudSubscriptionName string `json:"cloud_subscription_name"`
	CloudInitPassword     string `json:"cloud_init_password"`
}

// EncryptSecretKey encrypt secret key
func (req *AzureMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	// nothing to encrypt
	req.CloudInitPassword = cipher.EncryptToBase64(req.CloudInitPassword)
}

// HuaWeiMainAccountExtensionCreateReq ...
type HuaWeiMainAccountExtensionCreateReq struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// EncryptSecretKey ...
func (req *HuaWeiMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudInitPassword = cipher.EncryptToBase64(req.CloudInitPassword)
}

// ZenlayerMainAccountExtensionCreateReq ...
type ZenlayerMainAccountExtensionCreateReq struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// EncryptSecretKey ...
func (req *ZenlayerMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudInitPassword = cipher.EncryptToBase64(req.CloudInitPassword)
}

// KaopuMainAccountExtensionCreateReq ...
type KaopuMainAccountExtensionCreateReq struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// EncryptSecretKey ...
func (req *KaopuMainAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudInitPassword = cipher.EncryptToBase64(req.CloudInitPassword)
}

// MainAccountCreateReq ...
type MainAccountCreateReq[T MainAccountExtensionCreateReq] struct {
	CloudID           string                         `json:"cloud_id" validate:"required"`
	Email             string                         `json:"email" validate:"required"`
	Managers          []string                       `json:"managers" validate:"required"`
	BakManagers       []string                       `json:"bak_managers" validate:"required"`
	Site              enumor.MainAccountSiteType     `json:"site" validate:"required"`
	BusinessType      enumor.MainAccountBusinessType `json:"business_type" validate:"required"`
	Status            enumor.MainAccountStatus       `json:"status" validate:"required"`
	ParentAccountName string                         `json:"parent_account_name" validate:"required"`
	ParentAccountID   string                         `json:"parent_account_id" validate:"required"`
	DeptID            int64                          `json:"dept_id" validate:"required"`
	BkBizID           int64                          `json:"bk_biz_id" validate:"required"`
	OpProductID       int64                          `json:"op_product_id" validate:"required"`
	Memo              *string                        `json:"memo" validate:"required"`
	Extension         *T                             `json:"extension" validate:"required"`
}

// Validate ...
func (c *MainAccountCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// MainAccountUpdateReq 不允许对extension更新，允许更新字段：负责人/备份负责人/组织架构/运营产品/业务，单独更新状态
type MainAccountUpdateReq struct {
	Managers    []string                 `json:"managers" validate:"omitempty"`
	BakManagers []string                 `json:"bak_managers" validate:"omitempty"`
	Status      enumor.MainAccountStatus `json:"status" validate:"omitempty"`
	DeptID      int64                    `json:"dept_id" validate:"omitempty"`
	BkBizID     int64                    `json:"bk_biz_id" validate:"omitempty"`
	OpProductID int64                    `json:"op_product_id" validate:"omitempty"`
}

// Validate ...
func (u *MainAccountUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Get --------------------------
// MainAccountGetBaseResult
type MainAccountGetBaseResult struct {
	protocore.BaseMainAccount `json:",inline"`
}

// MainAccountGetBaseResp ...
type MainAccountGetBaseResp struct {
	rest.BaseResp `json:",inline"`
	Data          *MainAccountGetBaseResult `json:"data"`
}

// MainAccountExtensionGetResp main account extension
type MainAccountExtensionGetResp interface {
	protocore.AwsMainAccountExtension | protocore.GcpMainAccountExtension |
		protocore.HuaWeiMainAccountExtension | protocore.AzureMainAccountExtension |
		protocore.ZenlayerMainAccountExtension | protocore.KaopuMainAccountExtension
}

// MainAccountGetResult defines get main account result.
type MainAccountGetResult[T MainAccountExtensionGetResp] struct {
	protocore.BaseMainAccount `json:",inline"`
	Extension                 *T `json:"extension"`
}

// MainAccountGetResp defines get main account response.
type MainAccountGetResp[T MainAccountExtensionGetResp] struct {
	rest.BaseResp `json:",inline"`
	Data          *MainAccountGetResult[T] `json:"data"`
}

// -------------------------- List --------------------------

// MainAccountListResult defines list main account result.
type MainAccountListResult core.ListResultT[*protocore.BaseMainAccount]

// MainAccountListResp ...
type MainAccountListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *MainAccountListResult `json:"data"`
}

// -------------------------- Secret Encrypt And Decrypt --------------------------

// SecretEncryptor 用于加密"泛型"Extension密钥
type SecretEncryptor[T MainAccountExtensionCreateReq] interface {
	// EncryptSecretKey 加密约束，将密钥进行加密设置
	EncryptSecretKey(cryptography.Crypto)
	*T
}

// SecretDecryptor 用于解密"泛型"Extension密钥
type SecretDecryptor[T MainAccountExtensionGetResp] interface {
	// DecryptSecretKey 解密约束，需要支持将加密的密钥还原成明文
	DecryptSecretKey(cryptography.Crypto) error
	*T
}
