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

// -------------------------- Create --------------------------

type AccountExtensionCreateReq interface {
	TCloudAccountExtensionCreateReq | AwsAccountExtensionCreateReq | HuaWeiAccountExtensionCreateReq | GcpAccountExtensionCreateReq | AzureAccountExtensionCreateReq
}

// TCloudAccountExtensionCreateReq ...
type TCloudAccountExtensionCreateReq struct {
	CloudMainAccountID string `json:"cloud_main_account_id" validate:"required"`
	CloudSubAccountID  string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID      string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey     string `json:"cloud_secret_key" validate:"required"`
}

// AwsAccountExtensionCreateReq ...
type AwsAccountExtensionCreateReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// HuaWeiAccountExtensionCreateReq ...
type HuaWeiAccountExtensionCreateReq struct {
	CloudMainAccountName string `json:"cloud_main_account_name" validate:"required"`
	CloudSubAccountID    string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName  string `json:"cloud_sub_account_name" validate:"required"`
	CloudSecretID        string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey       string `json:"cloud_secret_key" validate:"required"`
}

// GcpAccountExtensionCreateReq ...
type GcpAccountExtensionCreateReq struct {
	CloudProjectID          string `json:"cloud_project_id" validate:"required"`
	CloudProjectName        string `json:"cloud_project_name" validate:"required"`
	CloudServiceAccountID   string `json:"cloud_service_account_cid" validate:"required"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"required"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"required"`
}

// AzureAccountExtensionCreateReq ...
type AzureAccountExtensionCreateReq struct {
	CloudTenantID         string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID   string `json:"cloud_subscription_id" validate:"required"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"required"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"required"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"required"`
	CloudClientID         string `json:"cloud_client_id" validate:"required"`
	CloudClientSecret     string `json:"cloud_client_secret" validate:"required"`
}

// AccountSpecCreateReq ...
type AccountSpecCreateReq struct {
	Name         string                 `json:"name" validate:"required"`
	Managers     []string               `json:"managers" validate:"required"`
	DepartmentID int64                  `json:"department_id" validate:"required"`
	Type         enumor.AccountType     `json:"type" validate:"required"`
	Site         enumor.AccountSiteType `json:"site" validate:"required"`
	Memo         *string                `json:"memo" validate:"required"`
}

// AccountAttachmentCreateReq ...
type AccountAttachmentCreateReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"required"`
}

// AccountCreateReq ...
type AccountCreateReq[T AccountExtensionCreateReq] struct {
	Spec       *AccountSpecCreateReq       `json:"spec" validate:"required"`
	Extension  *T                          `json:"extension" validate:"required"`
	Attachment *AccountAttachmentCreateReq `json:"attachment" validate:"required"`
}

// Validate ...
func (c *AccountCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

type AccountExtensionUpdateReq interface {
	TCloudAccountExtensionUpdateReq | AwsAccountExtensionUpdateReq | HuaWeiAccountExtensionUpdateReq | GcpAccountExtensionUpdateReq | AzureAccountExtensionUpdateReq
}

type TCloudAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"omitempty"`
}

type AwsAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"omitempty"`
}
type HuaWeiAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"omitempty"`
}
type GcpAccountExtensionUpdateReq struct {
	CloudServiceSecretID  string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}
type AzureAccountExtensionUpdateReq struct {
	CloudClientID     string `json:"cloud_client_id" validate:"required"`
	CloudClientSecret string `json:"cloud_client_secret" validate:"required"`
}

// AccountSpecUpdateReq ...
type AccountSpecUpdateReq struct {
	Name         string   `json:"name" validate:"omitempty"`
	Managers     []string `json:"managers" validate:"omitempty"`
	DepartmentID int64    `json:"department_id" validate:"omitempty"`
	SyncStatus   string   `json:"sync_status" validate:"omitempty"`
	Price        string   `json:"price" validate:"omitempty"`
	PriceUnit    string   `json:"price_unit" validate:"omitempty"`
	Memo         *string  `json:"memo" validate:"omitempty"`
}

// AccountUpdateReq ...
type AccountUpdateReq[T AccountExtensionUpdateReq] struct {
	Spec      *AccountSpecUpdateReq `json:"spec" validate:"omitempty"`
	Extension *T                    `json:"extension" validate:"omitempty"`
}

// Validate ...
func (u *AccountUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Get --------------------------

type AccountExtensionGetResp interface {
	cloud.TCloudAccountExtension | cloud.AwsAccountExtension | cloud.HuaWeiAccountExtension | cloud.GcpAccountExtension | cloud.AzureAccountExtension
}

type AccountGetResult[T AccountExtensionGetResp] struct {
	cloud.BaseAccount `json:",inline"`
	Extension         *T `json:"extension"`
}

type AccountGetResp[T AccountExtensionGetResp] struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountGetResult[T] `json:"data"`
}

// -------------------------- List --------------------------

// AccountListReq ...
type AccountListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate ...
func (l *AccountListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// BaseAccountListReq ...
type BaseAccountListReq struct {
	ID     uint64             `json:"id"`
	Vendor enumor.Vendor      `json:"vendor"`
	Spec   *cloud.AccountSpec `json:"spec"`
}

// AccountListResult defines list instances for iam pull resource callback result.
type AccountListResult struct {
	Count uint64 `json:"count,omitempty"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []*BaseAccountListReq `json:"details,omitempty"`
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

// UpdateAccountBizRelReq ...
type UpdateAccountBizRelReq struct {
	AccountID uint64   `json:"account_id" validate:"required"`
	BkBizIDs  []uint64 `json:"bk_biz_ids" validate:"required"`
}

// Validate ...
func (req *UpdateAccountBizRelReq) Validate() error {
	return validator.Validate.Struct(req)
}
