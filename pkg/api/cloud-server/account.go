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

package cloudserver

import (
	"encoding/json"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// AccountAttachmentCreateReq account attachment.
type AccountAttachmentCreateReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"required"`
}

// AccountSpecCreateReq ...
type AccountSpecCreateReq struct {
	Name         string                 `json:"name" validate:"required,min=3,max=32"`
	Managers     []string               `json:"managers" validate:"required"`
	DepartmentID int64                  `json:"department_id" validate:"required"`
	Type         enumor.AccountType     `json:"type" validate:"required"`
	Site         enumor.AccountSiteType `json:"site" validate:"required"`
	Memo         *string                `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AccountSpecCreateReq) Validate() error {
	if err := req.Type.Validate(); err != nil {
		return err
	}

	if err := req.Site.Validate(); err != nil {
		return err
	}

	return nil
}

// TCloudAccountExtensionCreateReq ...
type TCloudAccountExtensionCreateReq struct {
	CloudMainAccountID string `json:"cloud_main_account_id" validate:"required"`
	CloudSubAccountID  string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID      string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey     string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *TCloudAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountExtensionCreateReq ...
type AwsAccountExtensionCreateReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *AwsAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountExtensionCreateReq ...
type HuaWeiAccountExtensionCreateReq struct {
	CloudMainAccountName string `json:"cloud_main_account_name" validate:"required"`
	CloudSubAccountID    string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName  string `json:"cloud_sub_account_name" validate:"required"`
	CloudSecretID        string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey       string `json:"cloud_secret_key" validate:"required"`
	CloudIamUserID       string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername     string `json:"cloud_iam_username" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(r)
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

// Validate ...
func (r *GcpAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(r)
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

// Validate ...
func (r *AzureAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AccountCreateReq ...
type AccountCreateReq struct {
	Vendor enumor.Vendor         `json:"vendor" validate:"required"`
	Spec   *AccountSpecCreateReq `json:"spec" validate:"required"`
	// Extension 各云差异化比较大，延后解析成对应结果进行校验
	Extension  json.RawMessage             `json:"extension" validate:"required"`
	Attachment *AccountAttachmentCreateReq `json:"attachment" validate:"required"`
}

// Validate create account request.
func (req *AccountCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	if err := req.Spec.Validate(); err != nil {
		return err
	}

	return nil
}

// -------------------------- Check --------------------------

// AccountCheckReq ...
type AccountCheckReq struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	// Extension 各云差异化比较大，延后解析成对应结果进行校验
	Extension json.RawMessage `json:"extension" validate:"required"`
}

// Validate check account request.
func (req *AccountCheckReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	return nil
}

// TCloudAccountExtensionCheckByIDReq ...
type TCloudAccountExtensionCheckByIDReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *TCloudAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountExtensionCheckByIDReq ...
type AwsAccountExtensionCheckByIDReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *AwsAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountExtensionCheckByIDReq ...
type HuaWeiAccountExtensionCheckByIDReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountExtensionCheckByIDReq ...
type GcpAccountExtensionCheckByIDReq struct {
	CloudServiceSecretID  string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate ...
func (r *GcpAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureAccountExtensionCheckByIDReq ...
type AzureAccountExtensionCheckByIDReq struct {
	CloudClientID     string `json:"cloud_client_id" validate:"required"`
	CloudClientSecret string `json:"cloud_client_secret" validate:"required"`
}

// Validate ...
func (r *AzureAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AccountCheckByIDReq ...
type AccountCheckByIDReq struct {
	// Extension 各云差异化比较大，延后解析成对应结果进行校验
	Extension json.RawMessage `json:"extension" validate:"required"`
}

// Validate ...
func (req *AccountCheckByIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AccountListReq ...
type AccountListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate ...
func (req *AccountListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// TCloudAccountExtensionUpdateReq ...
type TCloudAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *TCloudAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountExtensionUpdateReq ...
type AwsAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *AwsAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountExtensionUpdateReq ...
type HuaWeiAccountExtensionUpdateReq struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountExtensionUpdateReq ...
type GcpAccountExtensionUpdateReq struct {
	CloudServiceSecretID  string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate ...
func (r *GcpAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureAccountExtensionUpdateReq ...
type AzureAccountExtensionUpdateReq struct {
	CloudClientID     string `json:"cloud_client_id" validate:"required"`
	CloudClientSecret string `json:"cloud_client_secret" validate:"required"`
}

// Validate ...
func (r *AzureAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AccountSpecUpdateReq ...
type AccountSpecUpdateReq struct {
	Name         string   `json:"name" validate:"omitempty"`
	Managers     []string `json:"managers" validate:"omitempty"`
	DepartmentID int64    `json:"department_id" validate:"omitempty"`
	Memo         *string  `json:"memo" validate:"omitempty"`
}

// AccountAttachmentUpdateReq ...
type AccountAttachmentUpdateReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"omitempty"`
}

// AccountUpdateReq ...
type AccountUpdateReq struct {
	Spec       *AccountSpecUpdateReq       `json:"spec" validate:"omitempty"`
	Extension  *json.RawMessage            `json:"extension" validate:"omitempty"`
	Attachment *AccountAttachmentUpdateReq `json:"attachment" validate:"omitempty"`
}

// Validate ...
func (req *AccountUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
