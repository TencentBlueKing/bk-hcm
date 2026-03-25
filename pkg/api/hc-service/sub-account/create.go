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

// CreateSubAccountReq define create sub account request for hc-service.
type CreateSubAccountReq struct {
	AccountID    string                        `json:"account_id" validate:"required"`
	Name         string                        `json:"name" validate:"required"`
	Remark       string                        `json:"remark" validate:"omitempty"`
	Email        string                        `json:"email" validate:"omitempty,email"`
	PhoneNum     string                        `json:"phone_num" validate:"omitempty"`
	ConsoleLogin *enumor.SubAccountConsoleLogin `json:"console_login" validate:"required"`
	CountryCode  string                        `json:"country_code" validate:"omitempty"`
}

// Validate create sub account request.
func (req *CreateSubAccountReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ConsoleLogin.Validate(); err != nil {
		return fmt.Errorf("console_login validate failed, err: %w", err)
	}

	return nil
}

// CreateSubAccountResult directly reuses adaptor AddUserResult.
type CreateSubAccountResult = typeaccount.AddUserResult

// CreateSubAccountResp define create sub account response.
type CreateSubAccountResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateSubAccountResult `json:"data"`
}
