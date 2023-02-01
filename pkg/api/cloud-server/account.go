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
	"errors"
	"fmt"
	"regexp"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

var (
	validAccountNameRegex   = regexp.MustCompile("^[a-z][a-z0-9-]{1,30}[a-z0-9]$")
	accountNameInvalidError = errors.New("invalid account name: name should begin with a lowercase letter, " +
		"contains lowercase letters(a-z), numbers(0-9) or hyphen(-), end with a lowercase letter or number, " +
		"length should be 3 to 32 letters")
)

// -------------------------- Create --------------------------

// AccountAttachmentCreateReq account attachment.
type AccountAttachmentCreateReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"required"`
}

// Validate ...
func (req *AccountAttachmentCreateReq) Validate() error {
	bizCount := len(req.BkBizIDs)
	for _, bizID := range req.BkBizIDs {
		// 校验是否非法业务ID
		if !(bizID == constant.AttachedAllBiz || bizID > 0) {
			return fmt.Errorf("invalid biz id: %d", bizID)
		}
		// 选择全业务时不可选择其他具体业务，即全业务时业务数量只能是1
		if bizID == constant.AttachedAllBiz && bizCount > 1 {
			return errors.New("can't choose specific biz when choose all biz")
		}
	}

	return nil
}

// AccountSpecCreateReq ...
type AccountSpecCreateReq struct {
	Name         string                 `json:"name" validate:"required,min=3,max=32"`
	Managers     []string               `json:"managers" validate:"required,max=5"`
	DepartmentID int64                  `json:"department_id" validate:"required,min=1"`
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

	if !validAccountNameRegex.MatchString(req.Name) {
		return accountNameInvalidError
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
	CloudIamUserID       string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername     string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID        string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey       string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountExtensionCreateReq ...
type GcpAccountExtensionCreateReq struct {
	CloudProjectID          string `json:"cloud_project_id" validate:"required"`
	CloudProjectName        string `json:"cloud_project_name" validate:"required"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"required"`
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
	CloudClientSecretID   string `json:"cloud_client_secret_id" validate:"required"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"required"`
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

	if err := req.Attachment.Validate(); err != nil {
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
	CloudSubAccountID string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID     string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey    string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *TCloudAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountExtensionCheckByIDReq ...
type AwsAccountExtensionCheckByIDReq struct {
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *AwsAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountExtensionCheckByIDReq ...
type HuaWeiAccountExtensionCheckByIDReq struct {
	CloudIamUserID   string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountExtensionCheckByIDReq ...
type GcpAccountExtensionCheckByIDReq struct {
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"required"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"required"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate ...
func (r *GcpAccountExtensionCheckByIDReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureAccountExtensionCheckByIDReq ...
type AzureAccountExtensionCheckByIDReq struct {
	CloudApplicationID   string `json:"cloud_application_id" validate:"required"`
	CloudApplicationName string `json:"cloud_application_name" validate:"required"`
	CloudClientSecretID  string `json:"cloud_client_secret_id" validate:"required"`
	CloudClientSecretKey string `json:"cloud_client_secret_key" validate:"required"`
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
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *AccountListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// TCloudAccountExtensionUpdateReq ...
type TCloudAccountExtensionUpdateReq struct {
	CloudSubAccountID string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID     string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey    string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *TCloudAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountExtensionUpdateReq ...
type AwsAccountExtensionUpdateReq struct {
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (req *AwsAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountExtensionUpdateReq ...
type HuaWeiAccountExtensionUpdateReq struct {
	CloudIamUserID   string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountExtensionUpdateReq ...
type GcpAccountExtensionUpdateReq struct {
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"required"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"required"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"required"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate ...
func (r *GcpAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureAccountExtensionUpdateReq ...
type AzureAccountExtensionUpdateReq struct {
	CloudApplicationID   string `json:"cloud_application_id" validate:"required"`
	CloudApplicationName string `json:"cloud_application_name" validate:"required"`
	CloudClientSecretID  string `json:"cloud_client_secret_id" validate:"required"`
	CloudClientSecretKey string `json:"cloud_client_secret_key" validate:"required"`
}

// Validate ...
func (r *AzureAccountExtensionUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AccountSpecUpdateReq ...
type AccountSpecUpdateReq struct {
	Name         string   `json:"name" validate:"omitempty"`
	Managers     []string `json:"managers" validate:"omitempty,max=5"`
	DepartmentID int64    `json:"department_id" validate:"omitempty,min=1"`
	Memo         *string  `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AccountSpecUpdateReq) Validate() error {
	if req.Name != "" && !validAccountNameRegex.MatchString(req.Name) {
		return accountNameInvalidError
	}
	return nil
}

// AccountAttachmentUpdateReq ...
type AccountAttachmentUpdateReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"omitempty"`
}

// Validate ...
func (req *AccountAttachmentUpdateReq) Validate() error {
	bizCount := len(req.BkBizIDs)
	for _, bizID := range req.BkBizIDs {
		// 校验是否非法业务ID
		if !(bizID == constant.AttachedAllBiz || bizID > 0) {
			return fmt.Errorf("invalid biz id: %d", bizID)
		}
		// 选择全业务时不可选择其他具体业务，即全业务时业务数量只能是1
		if bizID == constant.AttachedAllBiz && bizCount > 1 {
			return errors.New("can't choose specific biz when choose all biz")
		}
	}

	return nil
}

// AccountUpdateReq ...
type AccountUpdateReq struct {
	Spec       *AccountSpecUpdateReq       `json:"spec" validate:"omitempty"`
	Extension  *json.RawMessage            `json:"extension" validate:"omitempty"`
	Attachment *AccountAttachmentUpdateReq `json:"attachment" validate:"omitempty"`
}

// Validate ...
func (req *AccountUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Spec != nil {
		if err := req.Spec.Validate(); err != nil {
			return err
		}
	}

	if req.Attachment != nil {
		if err := req.Attachment.Validate(); err != nil {
			return err
		}
	}

	return nil
}
