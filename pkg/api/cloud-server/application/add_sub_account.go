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

package application

import (
	"encoding/json"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// SubAccountBaseReq is the common base for all sub-account application requests.
type SubAccountBaseReq struct {
	Vendor  enumor.Vendor `json:"vendor" validate:"required"`
	BkBizID int64         `json:"bk_biz_id" validate:"required,gt=0"`
}

// ValidateBase validates the common fields shared by all sub-account requests.
func (req *SubAccountBaseReq) ValidateBase() error {
	return req.Vendor.Validate()
}

// SubAccountBatchAddReq define sub account batch add request.
type SubAccountBatchAddReq struct {
	SubAccountBaseReq `json:",inline"`
	SubAccounts       []SubAccountAddReq `json:"sub_accounts" validate:"required,min=1,max=100"`
}

// Validate sub account batch add request.
func (req *SubAccountBatchAddReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ValidateBase(); err != nil {
		return err
	}

	for i, item := range req.SubAccounts {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("sub_accounts[%d] validate failed, err: %w", i, err)
		}
	}

	return nil
}

// SubAccountAddReq define sub account create request for a single sub-account.
type SubAccountAddReq struct {
	AccountID             string          `json:"account_id" validate:"required"`
	Name                  string          `json:"name" validate:"required"`
	PermissionTemplateIDs []string        `json:"permission_template_ids" validate:"required,min=1,dive,required"`
	ReceiveEmail          string          `json:"receive_email" validate:"required,email"`
	Email                 string          `json:"email" validate:"omitempty,email"`
	PhoneNum              string          `json:"phone_num" validate:"omitempty"`
	CountryCode           string          `json:"country_code" validate:"omitempty"`
	Managers              []string        `json:"managers" validate:"required,min=1"`
	Memo                  *string         `json:"memo" validate:"omitempty"`
	Extension             json.RawMessage `json:"extension" validate:"required"`
}

// Validate sub account create request.
func (item *SubAccountAddReq) Validate() error {
	return validator.Validate.Struct(item)
}

// TCloudSubAccountAddExtension defines the TCloud-specific extension fields for sub account creation.
type TCloudSubAccountAddExtension struct {
	ConsoleLogin *enumor.SubAccountConsoleLogin `json:"console_login" validate:"required"`
}

// Validate validates TCloudSubAccountAddExtension.
func (ext *TCloudSubAccountAddExtension) Validate() error {
	if err := validator.Validate.Struct(ext); err != nil {
		return err
	}

	return ext.ConsoleLogin.Validate()
}
