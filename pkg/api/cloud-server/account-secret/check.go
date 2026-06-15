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

// Package accountsecret defines account secret api call protocols.
package accountsecret

import (
	"encoding/json"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// AccountSecretCheckReq defines account secret check request.
type AccountSecretCheckReq struct {
	AccountID string                   `json:"account_id" validate:"required"`
	Type      enumor.AccountSecretType `json:"type" validate:"required"`
	Extension json.RawMessage          `json:"extension" validate:"required"`
}

// Validate account secret check request.
func (req *AccountSecretCheckReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return req.Type.Validate()
}

// TCloudAccountSecretExtension defines tcloud account secret extension.
type TCloudAccountSecretExtension struct {
	CloudSecretID  string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey string `json:"cloud_secret_key" validate:"required"`
}

// Validate tcloud account secret extension.
func (ext *TCloudAccountSecretExtension) Validate() error {
	return validator.Validate.Struct(ext)
}

// TCloudAccountSecretCheckResult defines tcloud account secret check result.
type TCloudAccountSecretCheckResult struct {
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
}
