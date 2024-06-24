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
// RootAccountCreateReq main account extension create req.
type RootAccountExtensionCreateReq interface {
	AwsRootAccountExtensionCreateReq | GcpRootAccountExtensionCreateReq |
		AzureRootAccountExtensionCreateReq | HuaWeiRootAccountExtensionCreateReq |
		ZenlayerRootAccountExtensionCreateReq | KaopuRootAccountExtensionCreateReq
}

// AwsRootAccountExtensionCreateReq ...
type AwsRootAccountExtensionCreateReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"omitempty"`
}

// EncryptSecretKey encrypt secret key
func (req *AwsRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudSecretKey = cipher.EncryptToBase64(req.CloudSecretKey)
}

// GcpRootAccountExtensionCreateReq ...
type GcpRootAccountExtensionCreateReq struct {
	Email                   string `json:"email" validate:"omitempty"`
	CloudProjectID          string `json:"cloud_project_id" validate:"required"`
	CloudProjectName        string `json:"cloud_project_name" validate:"required"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"omitempty"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"omitempty"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"omitempty"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"omitempty"`
}

// EncryptSecretKey encrypt secret key
func (req *GcpRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudServiceSecretKey = cipher.EncryptToBase64(req.CloudServiceSecretKey)
}

// AzureRootAccountExtensionCreateReq ...
type AzureRootAccountExtensionCreateReq struct {
	DisplayNameName       string `json:"display_name_name" validate:"omitempty"`
	CloudTenantID         string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID   string `json:"cloud_subscription_id" validate:"required"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"required"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"omitempty"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"omitempty"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"omitempty"`
}

// EncryptSecretKey encrypt secret key
func (req *AzureRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudClientSecretKey = cipher.EncryptToBase64(req.CloudClientSecretKey)
}

// HuaWeiRootAccountExtensionCreateReq ...
type HuaWeiRootAccountExtensionCreateReq struct {
	CloudSubAccountID   string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName string `json:"cloud_sub_account_name" validate:"required"`
	CloudSecretID       string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey      string `json:"cloud_secret_key" validate:"omitempty"`
	CloudIamUserID      string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername    string `json:"cloud_iam_username" validate:"required"`
}

// EncryptSecretKey ...
func (req *HuaWeiRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudSecretKey = cipher.EncryptToBase64(req.CloudSecretKey)
}

// ZenlayerRootAccountExtensionCreateReq ...
type ZenlayerRootAccountExtensionCreateReq struct {
	CloudAccountID string `json:"cloud_account_id" validate:"required"`
}

// EncryptSecretKey ...
func (req *ZenlayerRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {

}

// KaopuRootAccountExtensionCreateReq ...
type KaopuRootAccountExtensionCreateReq struct {
	CloudAccountID string `json:"cloud_account_id" validate:"required"`
}

// EncryptSecretKey ...
func (req *KaopuRootAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
}

// RootAccountCreateReq ...
type RootAccountCreateReq[T RootAccountExtensionCreateReq] struct {
	Name        string                     `json:"name" validate:"required"`
	CloudID     string                     `json:"cloud_id" validate:"required"`
	Email       string                     `json:"email" validate:"required"`
	Managers    []string                   `json:"managers" validate:"required"`
	BakManagers []string                   `json:"bak_managers" validate:"required"`
	Site        enumor.RootAccountSiteType `json:"site" validate:"required"`
	DeptID      int64                      `json:"dept_id" validate:"required"`
	Memo        *string                    `json:"memo" validate:"required"`
	Extension   *T                         `json:"extension" validate:"required"`
}

// Validate ...
func (c *RootAccountCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// RootAccountExtensionUpdateReq ...
type RootAccountExtensionUpdateReq interface {
	AwsRootAccountExtensionUpdateReq | GcpRootAccountExtensionUpdateReq |
		HuaWeiRootAccountExtensionUpdateReq | AzureRootAccountExtensionUpdateReq |
		ZenlayerRootAccountExtensionUpdateReq | KaopuRootAccountExtensionUpdateReq
}

// AwsRootAccountExtensionUpdateReq ...
type AwsRootAccountExtensionUpdateReq struct {
	CloudAccountID   string  `json:"cloud_account_id,omitempty" validate:"omitempty"`
	CloudIamUsername string  `json:"cloud_iam_username,omitempty" validate:"omitempty"`
	CloudSecretID    *string `json:"cloud_secret_id,omitempty" validate:"omitempty"`
	CloudSecretKey   *string `json:"cloud_secret_key,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *AwsRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudSecretKey != nil {
		encryptedCloudSecretKey := cipher.EncryptToBase64(*req.CloudSecretKey)
		req.CloudSecretKey = &encryptedCloudSecretKey
	}
}

// HuaWeiRootAccountExtensionUpdateReq ...
type HuaWeiRootAccountExtensionUpdateReq struct {
	CloudSubAccountID   string  `json:"cloud_sub_account_id,omitempty" validate:"omitempty"`
	CloudSubAccountName string  `json:"cloud_sub_account_name,omitempty" validate:"omitempty"`
	CloudSecretID       *string `json:"cloud_secret_id,omitempty" validate:"omitempty"`
	CloudSecretKey      *string `json:"cloud_secret_key,omitempty" validate:"omitempty"`
	CloudIamUserID      string  `json:"cloud_iam_user_id,omitempty" validate:"omitempty"`
	CloudIamUsername    string  `json:"cloud_iam_username,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *HuaWeiRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudSecretKey != nil {
		encryptedCloudSecretKey := cipher.EncryptToBase64(*req.CloudSecretKey)
		req.CloudSecretKey = &encryptedCloudSecretKey
	}
}

// GcpRootAccountExtensionUpdateReq ...
type GcpRootAccountExtensionUpdateReq struct {
	Email                   string  `json:"email" validate:"omitempty"`
	CloudProjectID          string  `json:"cloud_project_id,omitempty" validate:"omitempty"`
	CloudProjectName        string  `json:"cloud_project_name,omitempty" validate:"omitempty"`
	CloudServiceAccountID   *string `json:"cloud_service_account_id,omitempty" validate:"omitempty"`
	CloudServiceAccountName *string `json:"cloud_service_account_name,omitempty" validate:"omitempty"`
	CloudServiceSecretID    *string `json:"cloud_service_secret_id,omitempty" validate:"omitempty"`
	CloudServiceSecretKey   *string `json:"cloud_service_secret_key,omitempty" validate:"omitempty"`
	CloudBillingAccount     string  `json:"cloud_billing_account,omitempty" validate:"omitempty"`
	CloudOrganization       string  `json:"cloud_organization,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *GcpRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudServiceSecretKey != nil {
		encryptedCloudServiceSecretKey := cipher.EncryptToBase64(*req.CloudServiceSecretKey)
		req.CloudServiceSecretKey = &encryptedCloudServiceSecretKey
	}
}

// AzureRootAccountExtensionUpdateReq ...
type AzureRootAccountExtensionUpdateReq struct {
	DisplayNameName       string  `json:"display_name_name" validate:"omitempty"`
	CloudTenantID         string  `json:"cloud_tenant_id,omitempty" validate:"omitempty"`
	CloudSubscriptionID   string  `json:"cloud_subscription_id,omitempty" validate:"omitempty"`
	CloudSubscriptionName string  `json:"cloud_subscription_name,omitempty" validate:"omitempty"`
	CloudApplicationID    *string `json:"cloud_application_id,omitempty" validate:"omitempty"`
	CloudApplicationName  *string `json:"cloud_application_name,omitempty" validate:"omitempty"`
	CloudClientSecretID   *string `json:"cloud_client_secret_id,omitempty" validate:"omitempty"`
	CloudClientSecretKey  *string `json:"cloud_client_secret_key,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *AzureRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudClientSecretKey != nil {
		encryptedCloudClientSecretKey := cipher.EncryptToBase64(*req.CloudClientSecretKey)
		req.CloudClientSecretKey = &encryptedCloudClientSecretKey
	}
}

// ZenlayerRootAccountExtensionUpdateReq ...
type ZenlayerRootAccountExtensionUpdateReq struct {
}

// EncryptSecretKey ...
func (req *ZenlayerRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {

}

// KaopuRootAccountExtensionUpdateReq ...
type KaopuRootAccountExtensionUpdateReq struct {
}

// EncryptSecretKey ...
func (req *KaopuRootAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {

}

// RootAccountUpdateReq 不允许对extension更新，允许更新字段：负责人/备份负责人/组织架构/运营产品/业务，单独更新状态
type RootAccountUpdateReq[T RootAccountExtensionUpdateReq] struct {
	Name        string   `json:"name" validate:"omitempty"`
	Managers    []string `json:"managers" validate:"omitempty"`
	BakManagers []string `json:"bak_managers" validate:"omitempty"`
	DeptID      int64    `json:"dept_id" validate:"omitempty"`
	Memo        *string  `json:"memo" validate:"omitempty"`
	Extension   *T       `json:"extension" validate:"omitempty"`
}

// Validate ...
func (u *RootAccountUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Get --------------------------

// RootAccountGetBaseResult
type RootAccountGetBaseResult struct {
	protocore.BaseRootAccount `json:",inline"`
}

// RootAccountGetBaseResp ...
type RootAccountGetBaseResp struct {
	rest.BaseResp `json:",inline"`
	Data          *RootAccountGetBaseResult `json:"data"`
}

// RootAccountExtensionGetResp ...
type RootAccountExtensionGetResp interface {
	protocore.AwsRootAccountExtension | protocore.GcpRootAccountExtension |
		protocore.HuaWeiRootAccountExtension | protocore.AzureRootAccountExtension |
		protocore.ZenlayerRootAccountExtension | protocore.KaopuRootAccountExtension
}

// RootAccountGetResult ...
type RootAccountGetResult[T RootAccountExtensionGetResp] struct {
	protocore.BaseRootAccount `json:",inline"`
	Extension                 *T `json:"extension"`
}

// RootAccountGetResp ...
type RootAccountGetResp[T RootAccountExtensionGetResp] struct {
	rest.BaseResp `json:",inline"`
	Data          *RootAccountGetResult[T] `json:"data"`
}

// -------------------------- List --------------------------

// use core.ListWithoutFileds

// RootAccountListResult defines list main account result.
type RootAccountListResult core.ListResultT[*protocore.BaseRootAccount]

// RootAccountListResp ...
type RootAccountListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *RootAccountListResult `json:"data"`
}

// -------------------------- Secret Encrypt And Decrypt --------------------------

// SecretEncryptor 用于加密"泛型"Extension密钥
type RootSecretEncryptor[T RootAccountExtensionCreateReq | RootAccountExtensionUpdateReq] interface {
	// EncryptSecretKey 加密约束，将密钥进行加密设置
	EncryptSecretKey(cryptography.Crypto)
	*T
}

// SecretDecryptor 用于解密"泛型"Extension密钥
type RootSecretDecryptor[T RootAccountExtensionGetResp] interface {
	// DecryptSecretKey 解密约束，需要支持将加密的密钥还原成明文
	DecryptSecretKey(cryptography.Crypto) error
	*T
}
