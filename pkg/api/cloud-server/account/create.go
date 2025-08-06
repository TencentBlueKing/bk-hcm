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
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
)

// TCloudAccountExtensionCreateReq ...
type TCloudAccountExtensionCreateReq struct {
	CloudMainAccountID string `json:"cloud_main_account_id" validate:"required"`
	CloudSubAccountID  string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID      string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey     string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *TCloudAccountExtensionCreateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return secretEmptyError
	}

	return nil
}

// IsFull 对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *TCloudAccountExtensionCreateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// AwsAccountExtensionCreateReq ...
type AwsAccountExtensionCreateReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AwsAccountExtensionCreateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return secretEmptyError
	}

	return nil
}

// IsFull 对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *AwsAccountExtensionCreateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// HuaWeiAccountExtensionCreateReq ...
type HuaWeiAccountExtensionCreateReq struct {
	CloudSubAccountID   string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName string `json:"cloud_sub_account_name" validate:"required"`
	CloudIamUserID      string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername    string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID       string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey      string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *HuaWeiAccountExtensionCreateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return secretEmptyError
	}

	return nil
}

// IsFull 对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *HuaWeiAccountExtensionCreateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// GcpAccountExtensionCreateReq ...
type GcpAccountExtensionCreateReq struct {
	CloudProjectID          string `json:"cloud_project_id" validate:"required"`
	CloudProjectName        string `json:"cloud_project_name" validate:"required"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"omitempty"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"omitempty"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"omitempty"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *GcpAccountExtensionCreateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 检查密钥是否符合要求
	if err := validateGcpCloudServiceSK(req.CloudServiceSecretKey); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return errors.New("AccountID/AccountName/SecretID/SecretKey can not be empty")
	}

	return nil
}

// IsFull  对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *GcpAccountExtensionCreateReq) IsFull() bool {
	return req.CloudServiceSecretID != "" &&
		req.CloudServiceSecretKey != "" &&
		req.CloudServiceAccountID != "" &&
		req.CloudServiceAccountName != ""
}

// AzureAccountExtensionCreateReq ...
type AzureAccountExtensionCreateReq struct {
	CloudTenantID         string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID   string `json:"cloud_subscription_id" validate:"required"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"required"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"omitempty"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"omitempty"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AzureAccountExtensionCreateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return errors.New("ApplicationID/ApplicationName/SecretID/SecretKey can not be empty")
	}
	// 要求订阅id为小写
	if assert.ContainsUpperCase(req.CloudSubscriptionID) {
		return errors.New("CloudSubscriptionID should be lower case")
	}

	return nil
}

// IsFull  对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *AzureAccountExtensionCreateReq) IsFull() bool {
	return req.CloudClientSecretKey != "" &&
		req.CloudApplicationID != "" &&
		req.CloudApplicationName != ""
}

// AccountCommonInfoCreateReq ...
type AccountCommonInfoCreateReq struct {
	Vendor      enumor.Vendor          `json:"vendor" validate:"required"`
	Name        string                 `json:"name" validate:"required,min=3,max=255"`
	Managers    []string               `json:"managers" validate:"required,max=5"`
	Type        enumor.AccountType     `json:"type" validate:"required"`
	Site        enumor.AccountSiteType `json:"site" validate:"required"`
	Memo        *string                `json:"memo" validate:"omitempty"`
	BkBizID     int64                  `json:"bk_biz_id" validate:"omitempty"`
	UsageBizIDs []int64                `json:"usage_biz_ids" validate:"required"`
}

// Validate ...
func (req *AccountCommonInfoCreateReq) Validate() error {
	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	if err := req.Type.Validate(); err != nil {
		return err
	}

	if err := req.Site.Validate(); err != nil {
		return err
	}

	// 部分云只有国际站
	if (req.Vendor == enumor.Gcp || req.Vendor == enumor.Azure || req.Vendor == enumor.HuaWei) &&
		req.Site != enumor.InternationalSite {
		return fmt.Errorf("%s support only international site", req.Vendor)
	}

	// 名称有限制特定规则
	if err := validateAccountName(req.Name); err != nil {
		return err
	}

	// 使用业务合法性校验
	if err := validateUsageBizIDs(req.UsageBizIDs); err != nil {
		return err
	}

	// 资源账号需要进一步对管理业务和使用业务进行校验
	if err := req.validateBizIDAndUsageBizIDs(req.BkBizID, req.UsageBizIDs, req.Type); err != nil {
		return err
	}

	if req.Memo != nil {
		if utf8.RuneCountInString(cvt.PtrToVal(req.Memo)) > 255 {
			return errors.New("invalid account memo, length should less than 255")
		}
	}
	return nil
}

// validateBizIDAndUsageBizIDs 校验管理业务和使用业务的合法性
func (req *AccountCommonInfoCreateReq) validateBizIDAndUsageBizIDs(bizID int64, usageBizIDs []int64, accountType enumor.AccountType) error {
	// 资源账号需要进一步对管理业务和使用业务进行校验
	if accountType == enumor.ResourceAccount {
		if err := req.validateResAccountBizIDs(); err != nil {
			return err
		}
		return nil
	}
	// 对于非资源账号：结合上usageBizIDs在请求体的required，相当于要求使用业务切片长度必须为1
	return validateNonResAccountBizIDs(bizID, usageBizIDs)
}

// validateResAccountBizIDs 校验资源账号管理业务和使用业务的合法性
func (req *AccountCommonInfoCreateReq) validateResAccountBizIDs() error {
	// 管理业务合法性校验
	if err := req.validateBizID(); err != nil {
		return err
	}
	// 校验使用业务是否包含管理业务，要求必须包含
	if err := validateBizIDInUsageBizIDs(req.BkBizID, req.UsageBizIDs); err != nil {
		return err
	}
	return nil
}

func (req *AccountCommonInfoCreateReq) validateBizID() error {
	if req.BkBizID == 0 {
		return fmt.Errorf("bk_biz_id can not be empty")
	}
	if req.BkBizID == constant.AttachedAllBiz {
		return fmt.Errorf("bk_biz_id can not set all biz")
	}
	return nil
}

// AccountCreateReq ...
type AccountCreateReq struct {
	AccountCommonInfoCreateReq `json:",inline"`
	// Extension 各云差异化比较大，延后解析成对应结果进行校验
	Extension json.RawMessage `json:"extension" validate:"required"`
}

// Validate create account request.
func (req *AccountCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.AccountCommonInfoCreateReq.Validate(); err != nil {
		return err
	}

	return nil
}

// -------------------------- Check --------------------------

// AccountCheckReq ...
type AccountCheckReq struct {
	Vendor enumor.Vendor      `json:"vendor" validate:"required"`
	Type   enumor.AccountType `json:"type" validate:"required"`
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

	if err := req.Type.Validate(); err != nil {
		return err
	}

	return nil
}
