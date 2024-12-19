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
	"hcm/pkg/criteria/validator"
)

// AccountInfoBySecret 根据秘钥获取的账号字段
type AccountInfoBySecret interface {
	TCloudInfoBySecret | AwsInfoBySecret | AzureInfoBySecret | GcpInfoBySecret | HuaWeiInfoBySecret
}

// TCloudInfoBySecret 腾讯云根据秘钥获取的字段
type TCloudInfoBySecret struct {
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
}

// AwsInfoBySecret AWS 根据秘钥获取的字段
type AwsInfoBySecret struct {
	CloudAccountID   string `json:"cloud_account_id"`
	CloudIamUsername string `json:"cloud_iam_username"`
}

// HuaWeiInfoBySecret 华为云 根据秘钥获取的字段
type HuaWeiInfoBySecret struct {
	CloudSubAccountID   string `json:"cloud_sub_account_id"`
	CloudSubAccountName string `json:"cloud_sub_account_name"`
	CloudIamUserID      string `json:"cloud_iam_user_id"`
	CloudIamUsername    string `json:"cloud_iam_username"`
}

// GcpInfoBySecret GCP 根据秘钥获取的字段
type GcpInfoBySecret struct {
	CloudProjectInfos []GcpProjectInfo `json:"cloud_project_infos"`
}

// GcpProjectInfo GCP 单个project的字段信息
type GcpProjectInfo struct {
	Email                   string `json:"email"`
	CloudProjectID          string `json:"cloud_project_id"`
	CloudProjectName        string `json:"cloud_project_name"`
	CloudServiceAccountID   string `json:"cloud_service_account_id"`
	CloudServiceAccountName string `json:"cloud_service_account_name"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id"`
}

// AzureInfoBySecret Azure 根据秘钥获取的字段
type AzureInfoBySecret struct {
	SubscriptionInfos []AzureSubscriptionInfo `json:"cloud_subscription_infos"`
	ApplicationInfos  []AzureApplicationInfo  `json:"cloud_application_infos"`
}

// AzureApplicationInfo Azure 单个应用实例的字段信息
type AzureApplicationInfo struct {
	CloudApplicationID   string `json:"cloud_application_id"`
	CloudApplicationName string `json:"cloud_application_name"`
}

// AzureSubscriptionInfo Azure 单个订阅的字段信息
type AzureSubscriptionInfo struct {
	CloudSubscriptionID   string `json:"cloud_subscription_id"`
	CloudSubscriptionName string `json:"cloud_subscription_name"`
}

// AccountSecret 账号所需秘钥
type AccountSecret interface {
	TCloudSecret | AwsSecret | HuaWeiSecret | GcpSecret | AzureSecret
	Validate() error
}

// TCloudSecret 腾讯云秘钥
type TCloudSecret struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (sk TCloudSecret) Validate() error {
	return validator.Validate.Struct(sk)
}

// AwsSecret AWS 秘钥
type AwsSecret struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (sk AwsSecret) Validate() error {
	return validator.Validate.Struct(sk)
}

// HuaWeiSecret 华为云秘钥
type HuaWeiSecret struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (sk HuaWeiSecret) Validate() error {
	return validator.Validate.Struct(sk)
}

// GcpSecret GCP 秘钥
type GcpSecret struct {
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}

// GcpCredential gcp credential
type GcpCredential struct {
	CloudProjectID        string `json:"cloud_project_id" validate:"required"`
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate GcpCredential
func (g *GcpCredential) Validate() error {
	return validator.Validate.Struct(g)
}

// Validate ...
func (sk GcpSecret) Validate() error {
	return validator.Validate.Struct(sk)
}

// AzureSecret Azure 秘钥
type AzureSecret struct {
	CloudTenantID        string `json:"cloud_tenant_id" validate:"required"`
	CloudApplicationID   string `json:"cloud_application_id" validate:"required"`
	CloudClientSecretKey string `json:"cloud_client_secret_key" validate:"required"`
}

// Validate ...
func (sk AzureSecret) Validate() error {
	return validator.Validate.Struct(sk)
}

// AzureAuthSecret Azure 秘钥
type AzureAuthSecret struct {
	CloudTenantID        string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID  string `json:"cloud_subscription_id" validate:"required"`
	CloudApplicationID   string `json:"cloud_application_id" validate:"required"`
	CloudClientSecretKey string `json:"cloud_client_secret_key" validate:"required"`
}

// Validate ...
func (sk AzureAuthSecret) Validate() error {
	return validator.Validate.Struct(sk)
}
