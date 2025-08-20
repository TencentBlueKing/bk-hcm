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

	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// TCloudAccountExtensionUpdateReq ...
type TCloudAccountExtensionUpdateReq struct {
	CloudSubAccountID string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID     string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey    string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *TCloudAccountExtensionUpdateReq) Validate(accountType enumor.AccountType) error {
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
func (req *TCloudAccountExtensionUpdateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// AwsAccountExtensionUpdateReq ...
type AwsAccountExtensionUpdateReq struct {
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AwsAccountExtensionUpdateReq) Validate(accountType enumor.AccountType) error {
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
func (req *AwsAccountExtensionUpdateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// HuaWeiAccountExtensionUpdateReq ...
type HuaWeiAccountExtensionUpdateReq struct {
	CloudSubAccountName string `json:"cloud_sub_account_name" validate:"required"`
	CloudIamUserID      string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername    string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID       string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey      string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *HuaWeiAccountExtensionUpdateReq) Validate(accountType enumor.AccountType) error {
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
func (req *HuaWeiAccountExtensionUpdateReq) IsFull() bool {
	return req.CloudSecretID != "" && req.CloudSecretKey != ""
}

// GcpAccountExtensionUpdateReq ...
type GcpAccountExtensionUpdateReq struct {
	CloudProjectName        string `json:"cloud_project_name" validate:"omitempty"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"omitempty"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"omitempty"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"omitempty"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *GcpAccountExtensionUpdateReq) Validate(accountType enumor.AccountType) error {
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
func (req *GcpAccountExtensionUpdateReq) IsFull() bool {
	return req.CloudServiceSecretID != "" &&
		req.CloudServiceSecretKey != "" &&
		req.CloudServiceAccountID != "" &&
		req.CloudServiceAccountName != ""
}

// AzureAccountExtensionUpdateReq ...
type AzureAccountExtensionUpdateReq struct {
	CloudTenantID         string `json:"cloud_tenant_id" validate:"omitempty"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"omitempty"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"omitempty"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"omitempty"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AzureAccountExtensionUpdateReq) Validate(accountType enumor.AccountType) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 登记账号密钥可为空，其他类型则必填
	if accountType != enumor.RegistrationAccount && !req.IsFull() {
		return errors.New("ApplicationID/ApplicationName/SecretID/SecretKey can not be empty")
	}

	return nil
}

// IsFull  对于不同账号类型，有些字段是允许为空的，这里返回是否所有字段都有值
func (req *AzureAccountExtensionUpdateReq) IsFull() bool {
	return req.CloudClientSecretKey != "" &&
		req.CloudApplicationID != "" &&
		req.CloudApplicationName != ""
}

// OtherAccountExtensionUpdateReq ...
type OtherAccountExtensionUpdateReq struct {
	// placeholder
	CloudID     string `json:"cloud_id" validate:"omitempty"`
	CloudSecKey string `json:"cloud_sec_key" validate:"omitempty"`
}

// AccountUpdateReq ...
type AccountUpdateReq struct {
	Name               string          `json:"name" validate:"omitempty"`
	Managers           []string        `json:"managers" validate:"omitempty,max=5"`
	Memo               *string         `json:"memo" validate:"omitempty"`
	RecycleReserveTime int             `json:"recycle_reserve_time" validate:"omitempty"`
	BkBizID            int64           `json:"bk_biz_id" validate:"omitempty"`
	UsageBizIDs        []int64         `json:"usage_biz_ids" validate:"omitempty"`
	Extension          json.RawMessage `json:"extension" validate:"omitempty"`
}

// Validate ...
func (req *AccountUpdateReq) Validate(accountInfo *cloud.BaseAccount) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 名称有限制特定规则
	if err := validateAccountName(req.Name); err != nil {
		return err
	}

	// 使用业务合法性校验
	if err := validateUsageBizIDs(req.UsageBizIDs); err != nil {
		return err
	}

	// 根据账号类型进一步校验管理业务和使用业务的合法性
	if err := req.validateBizIDAndUsageBizIDs(accountInfo); err != nil {
		return err
	}
	return nil
}

func (req *AccountUpdateReq) validateBizIDAndUsageBizIDs(accountInfo *cloud.BaseAccount) error {
	if accountInfo.Type != enumor.ResourceAccount {
		return validateNonResAccountBizIDs(req.BkBizID, req.UsageBizIDs)
	}

	// 确定要使用的 bizID 和 usageBizIDs
	bizID := req.BkBizID
	if bizID == 0 {
		bizID = accountInfo.BkBizID
	}

	usageBizIDs := req.UsageBizIDs
	if usageBizIDs == nil {
		usageBizIDs = accountInfo.UsageBizIDs
	}

	// 特殊校验：不能设置为 all biz
	if bizID == constant.AttachedAllBiz {
		return fmt.Errorf("bk_biz_id can not set all biz")
	}

	// 统一校验逻辑
	return validateBizIDInUsageBizIDs(bizID, usageBizIDs)
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
