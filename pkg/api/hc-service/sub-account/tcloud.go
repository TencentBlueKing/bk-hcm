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

// Package hssubaccount defines hc-service sub account api types.
package hssubaccount

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// ------------------------- List Account -------------------------

// TCloudListAccountReq define tcloud list account request for hc-service.
type TCloudListAccountReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate tcloud list account request.
func (req *TCloudListAccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudListAccountResult directly reuses adaptor TCloudAccount slice.
type TCloudListAccountResult = []typeaccount.TCloudAccount

// TCloudListAccountResp define tcloud list account response.
type TCloudListAccountResp struct {
	rest.BaseResp `json:",inline"`
	Data          TCloudListAccountResult `json:"data"`
}

// ------------------------- Create Sub Account -------------------------

// TCloudCreateSubAccountReq define tcloud create sub account request for hc-service.
type TCloudCreateSubAccountReq struct {
	AccountID    string                         `json:"account_id" validate:"required"`
	Name         string                         `json:"name" validate:"required"`
	Remark       string                         `json:"remark" validate:"omitempty"`
	Email        string                         `json:"email" validate:"omitempty,email"`
	PhoneNum     string                         `json:"phone_num" validate:"omitempty"`
	ConsoleLogin *enumor.SubAccountConsoleLogin `json:"console_login" validate:"required"`
	CountryCode  string                         `json:"country_code" validate:"omitempty"`
}

// Validate tcloud create sub account request.
func (req *TCloudCreateSubAccountReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ConsoleLogin.Validate(); err != nil {
		return fmt.Errorf("console_login validate failed, err: %w", err)
	}

	return nil
}

// TCloudCreateSubAccountResult directly reuses adaptor AddUserResult.
type TCloudCreateSubAccountResult = typeaccount.AddUserResult

// TCloudCreateSubAccountResp define tcloud create sub account response.
type TCloudCreateSubAccountResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudCreateSubAccountResult `json:"data"`
}

// ------------------------- Update Sub Account -------------------------

// TCloudUpdateSubAccountReq define tcloud update sub account request for hc-service.
// Pointer fields use nil to indicate "no change".
type TCloudUpdateSubAccountReq struct {
	AccountID   string  `json:"account_id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Remark      *string `json:"remark" validate:"omitempty"`
	Email       *string `json:"email" validate:"omitempty"`
	PhoneNum    *string `json:"phone_num" validate:"omitempty"`
	CountryCode *string `json:"country_code" validate:"omitempty"`
}

// Validate tcloud update sub account request.
func (req *TCloudUpdateSubAccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------- Delete Sub Account -------------------------

// TCloudDeleteSubAccountReq define tcloud delete sub account request for hc-service.
type TCloudDeleteSubAccountReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

// Validate tcloud delete sub account request.
func (req *TCloudDeleteSubAccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------- Describe Sub Accounts -------------------------

// TCloudDescribeSubAccountsReq define tcloud describe sub accounts request for hc-service.
// reference: https://cloud.tencent.com/document/api/598/53486
type TCloudDescribeSubAccountsReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	SubUin    []uint64 `json:"sub_uin" validate:"required"`
}

// Validate tcloud describe sub accounts request.
func (req *TCloudDescribeSubAccountsReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.SubUin) < 1 {
		return fmt.Errorf("sub_uin count %d is less than 1", len(req.SubUin))
	}

	if len(req.SubUin) > typeaccount.DescribeSubAccountsMaxUIN {
		return fmt.Errorf("sub_uin count %d exceeds max %d",
			len(req.SubUin), typeaccount.DescribeSubAccountsMaxUIN)
	}

	return nil
}

// ------------------------- Describe Safe Auth Flag -------------------------

// TCloudDescribeSafeAuthFlagReq define tcloud describe sub-account safe auth flag request for hc-service.
type TCloudDescribeSafeAuthFlagReq struct {
	AccountID string `json:"account_id" validate:"required"`
	SubUin    uint64 `json:"sub_uin" validate:"required"`
}

// Validate tcloud describe safe auth flag request.
func (req *TCloudDescribeSafeAuthFlagReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudDescribeSafeAuthFlagResult directly reuses adaptor SafeAuthFlagResult.
type TCloudDescribeSafeAuthFlagResult = typeaccount.SafeAuthFlagResult

// TCloudDescribeSafeAuthFlagResp define tcloud describe sub-account safe auth flag response.
type TCloudDescribeSafeAuthFlagResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudDescribeSafeAuthFlagResult `json:"data"`
}

// ------------------------- Set MFA Flag -------------------------

// TCloudSetMfaFlagReq define tcloud set sub-account MFA flag request for hc-service.
type TCloudSetMfaFlagReq struct {
	AccountID  string                       `json:"account_id" validate:"required"`
	OpUin      uint64                       `json:"op_uin" validate:"required"`
	LoginFlag  *typeaccount.LoginActionFlag `json:"login_flag" validate:"omitempty"`
	ActionFlag *typeaccount.LoginActionFlag `json:"action_flag" validate:"omitempty"`
}

// Validate tcloud set MFA flag request.
func (req *TCloudSetMfaFlagReq) Validate() error {
	return validator.Validate.Struct(req)
}
