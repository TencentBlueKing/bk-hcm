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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/cryptography"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// AccountExtensionCreateReq account extension create req.
type AccountExtensionCreateReq interface {
	TCloudAccountExtensionCreateReq | AwsAccountExtensionCreateReq | HuaWeiAccountExtensionCreateReq |
		GcpAccountExtensionCreateReq | AzureAccountExtensionCreateReq | OtherAccountExtensionCreateReq
}

// TCloudAccountExtensionCreateReq ...
type TCloudAccountExtensionCreateReq struct {
	CloudMainAccountID string `json:"cloud_main_account_id" validate:"required"`
	CloudSubAccountID  string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID      string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey     string `json:"cloud_secret_key" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *TCloudAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudSecretKey = cipher.EncryptToBase64(req.CloudSecretKey)
}

// AwsAccountExtensionCreateReq ...
type AwsAccountExtensionCreateReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *AwsAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudSecretKey = cipher.EncryptToBase64(req.CloudSecretKey)
}

// HuaWeiAccountExtensionCreateReq ...
type HuaWeiAccountExtensionCreateReq struct {
	CloudSubAccountID   string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName string `json:"cloud_sub_account_name" validate:"required"`
	CloudSecretID       string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey      string `json:"cloud_secret_key" validate:"omitempty"`
	CloudIamUserID      string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername    string `json:"cloud_iam_username" validate:"required"`
}

// EncryptSecretKey ...
func (req *HuaWeiAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudSecretKey = cipher.EncryptToBase64(req.CloudSecretKey)
}

// GcpAccountExtensionCreateReq ...
type GcpAccountExtensionCreateReq struct {
	Email                   string `json:"email" validate:"omitempty"`
	CloudProjectID          string `json:"cloud_project_id" validate:"required"`
	CloudProjectName        string `json:"cloud_project_name" validate:"required"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"omitempty"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"omitempty"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"omitempty"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *GcpAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudServiceSecretKey = cipher.EncryptToBase64(req.CloudServiceSecretKey)
}

// AzureAccountExtensionCreateReq ...
type AzureAccountExtensionCreateReq struct {
	DisplayNameName       string `json:"display_name_name" validate:"omitempty"`
	CloudTenantID         string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID   string `json:"cloud_subscription_id" validate:"required"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"required"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"omitempty"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"omitempty"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *AzureAccountExtensionCreateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	req.CloudClientSecretKey = cipher.EncryptToBase64(req.CloudClientSecretKey)
}

// AccountCreateReq ...
type AccountCreateReq[T AccountExtensionCreateReq] struct {
	Name      string                 `json:"name" validate:"required"`
	Managers  []string               `json:"managers" validate:"required"`
	Type      enumor.AccountType     `json:"type" validate:"required"`
	Site      enumor.AccountSiteType `json:"site" validate:"required"`
	Memo      *string                `json:"memo" validate:"required"`
	Extension *T                     `json:"extension" validate:"required"`
	BkBizIDs  []int64                `json:"bk_biz_ids" validate:"required"`
}

// Validate ...
func (c *AccountCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// AccountExtensionUpdateReq Note: DataService的更新是与业务无关的，所以必须支持调用方根据场景需求来更新部分字段
// Note: 对于允许为空字符串的字段，则其类型需要定义为指针，正常情况下，Json合并时空值会被忽略
type AccountExtensionUpdateReq interface {
	TCloudAccountExtensionUpdateReq | AwsAccountExtensionUpdateReq | HuaWeiAccountExtensionUpdateReq |
		GcpAccountExtensionUpdateReq | AzureAccountExtensionUpdateReq
}

// TCloudAccountExtensionUpdateReq ...
type TCloudAccountExtensionUpdateReq struct {
	CloudMainAccountID string  `json:"cloud_main_account_id,omitempty" validate:"omitempty"`
	CloudSubAccountID  string  `json:"cloud_sub_account_id,omitempty" validate:"omitempty"`
	CloudSecretID      *string `json:"cloud_secret_id,omitempty" validate:"omitempty"`
	CloudSecretKey     *string `json:"cloud_secret_key,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *TCloudAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudSecretKey != nil {
		encryptedCloudSecretKey := cipher.EncryptToBase64(*req.CloudSecretKey)
		req.CloudSecretKey = &encryptedCloudSecretKey
	}
}

// AwsAccountExtensionUpdateReq ...
type AwsAccountExtensionUpdateReq struct {
	CloudAccountID   string  `json:"cloud_account_id,omitempty" validate:"omitempty"`
	CloudIamUsername string  `json:"cloud_iam_username,omitempty" validate:"omitempty"`
	CloudSecretID    *string `json:"cloud_secret_id,omitempty" validate:"omitempty"`
	CloudSecretKey   *string `json:"cloud_secret_key,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *AwsAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudSecretKey != nil {
		encryptedCloudSecretKey := cipher.EncryptToBase64(*req.CloudSecretKey)
		req.CloudSecretKey = &encryptedCloudSecretKey
	}
}

// HuaWeiAccountExtensionUpdateReq ...
type HuaWeiAccountExtensionUpdateReq struct {
	CloudSubAccountID   string  `json:"cloud_sub_account_id,omitempty" validate:"omitempty"`
	CloudSubAccountName string  `json:"cloud_sub_account_name,omitempty" validate:"omitempty"`
	CloudSecretID       *string `json:"cloud_secret_id,omitempty" validate:"omitempty"`
	CloudSecretKey      *string `json:"cloud_secret_key,omitempty" validate:"omitempty"`
	CloudIamUserID      string  `json:"cloud_iam_user_id,omitempty" validate:"omitempty"`
	CloudIamUsername    string  `json:"cloud_iam_username,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *HuaWeiAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudSecretKey != nil {
		encryptedCloudSecretKey := cipher.EncryptToBase64(*req.CloudSecretKey)
		req.CloudSecretKey = &encryptedCloudSecretKey
	}
}

// GcpAccountExtensionUpdateReq ...
type GcpAccountExtensionUpdateReq struct {
	Email                   string  `json:"email" validate:"omitempty"`
	CloudProjectID          string  `json:"cloud_project_id,omitempty" validate:"omitempty"`
	CloudProjectName        string  `json:"cloud_project_name,omitempty" validate:"omitempty"`
	CloudServiceAccountID   *string `json:"cloud_service_account_id,omitempty" validate:"omitempty"`
	CloudServiceAccountName *string `json:"cloud_service_account_name,omitempty" validate:"omitempty"`
	CloudServiceSecretID    *string `json:"cloud_service_secret_id,omitempty" validate:"omitempty"`
	CloudServiceSecretKey   *string `json:"cloud_service_secret_key,omitempty" validate:"omitempty"`
}

// EncryptSecretKey ...
func (req *GcpAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudServiceSecretKey != nil {
		encryptedCloudServiceSecretKey := cipher.EncryptToBase64(*req.CloudServiceSecretKey)
		req.CloudServiceSecretKey = &encryptedCloudServiceSecretKey
	}
}

// AzureAccountExtensionUpdateReq ...
type AzureAccountExtensionUpdateReq struct {
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
func (req *AzureAccountExtensionUpdateReq) EncryptSecretKey(cipher cryptography.Crypto) {
	if req.CloudClientSecretKey != nil {
		encryptedCloudClientSecretKey := cipher.EncryptToBase64(*req.CloudClientSecretKey)
		req.CloudClientSecretKey = &encryptedCloudClientSecretKey
	}
}

// AccountUpdateReq ...
type AccountUpdateReq[T AccountExtensionUpdateReq] struct {
	Name               string   `json:"name" validate:"omitempty"`
	Managers           []string `json:"managers" validate:"omitempty"`
	Price              string   `json:"price" validate:"omitempty"`
	PriceUnit          string   `json:"price_unit" validate:"omitempty"`
	Memo               *string  `json:"memo" validate:"omitempty"`
	RecycleReserveTime int      `json:"recycle_reserve_time" validate:"omitempty"`
	Extension          *T       `json:"extension" validate:"omitempty"`
}

// Validate ...
func (u *AccountUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Get --------------------------

// AccountExtensionGetResp ...
type AccountExtensionGetResp interface {
	cloud.TCloudAccountExtension | cloud.AwsAccountExtension | cloud.HuaWeiAccountExtension |
		cloud.GcpAccountExtension | cloud.AzureAccountExtension | cloud.OtherAccountExtension
}

// AccountGetResult ...
type AccountGetResult[T AccountExtensionGetResp] struct {
	cloud.BaseAccount `json:",inline"`
	Extension         *T `json:"extension"`
}

// AccountGetResp ...
type AccountGetResp[T AccountExtensionGetResp] struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountGetResult[T] `json:"data"`
}

// -------------------------- List --------------------------

// AccountListReq ...
type AccountListReq = core.ListReq

// BaseAccountListResp ...
type BaseAccountListResp struct {
	ID            string                 `json:"id"`
	Vendor        enumor.Vendor          `json:"vendor"`
	Name          string                 `json:"name"`
	Managers      []string               `json:"managers"`
	Type          enumor.AccountType     `json:"type"`
	Site          enumor.AccountSiteType `json:"site"`
	Price         string                 `json:"price"`
	PriceUnit     string                 `json:"price_unit"`
	Memo          *string                `json:"memo"`
	core.Revision `json:",inline"`
}

// AccountListResult defines list instances for iam pull resource callback result.
type AccountListResult struct {
	Count uint64 `json:"count"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []*cloud.BaseAccount `json:"details"`
}

// AccountListResp ...
type AccountListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// AccountDeleteReq ...
type AccountDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (d *AccountDeleteReq) Validate() error {
	return validator.Validate.Struct(d)
}

// AccountDeleteValidateResp ...
type AccountDeleteValidateResp struct {
	rest.BaseResp `json:",inline"`
	Data          map[string]uint64 `json:"data"`
}

// -------------------------- List With Extension/Secret--------------------------

// BaseAccountWithExtensionListResp ...
type BaseAccountWithExtensionListResp struct {
	cloud.BaseAccount `json:",inline"`
	Extension         map[string]interface{} `json:"extension"`
}

// AccountWithExtensionListResult defines list instances for iam pull resource callback result.
type AccountWithExtensionListResult struct {
	Count   uint64                              `json:"count"`
	Details []*BaseAccountWithExtensionListResp `json:"details"`
}

// AccountWithExtensionListResp ...
type AccountWithExtensionListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountWithExtensionListResult `json:"data"`
}

// -------------------------- Secret Encrypt And Decrypt --------------------------

// SecretEncryptor 用于加密"泛型"Extension密钥
type SecretEncryptor[T AccountExtensionCreateReq | AccountExtensionUpdateReq] interface {
	// EncryptSecretKey 加密约束，将密钥进行加密设置
	EncryptSecretKey(cryptography.Crypto)
	*T
}

// SecretDecryptor 用于解密"泛型"Extension密钥
type SecretDecryptor[T AccountExtensionGetResp] interface {
	// DecryptSecretKey 解密约束，需要支持将加密的密钥还原成明文
	DecryptSecretKey(cryptography.Crypto) error
	*T
}

// OtherAccountExtensionCreateReq ...
type OtherAccountExtensionCreateReq struct {
}
