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

// Package accountset defines account-server api call protocols.
package accountset

import (
	"encoding/json"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// use core.ListWithoutFields

// -------------------------- Add --------------------------

// RootAccountCommonAddReq ...
type RootAccountCommonAddReq struct {
	Name        string                     `json:"name" validate:"required,min=3,max=64"`
	Vendor      enumor.Vendor              `json:"vendor" validate:"required"`
	Email       string                     `json:"email" validate:"required"`
	Managers    []string                   `json:"managers" validate:"required,max=5"`
	BakManagers []string                   `json:"bak_managers" validate:"required,max=5"`
	Site        enumor.RootAccountSiteType `json:"site" validate:"required"`
	DeptID      int64                      `json:"dept_id" validate:"omitempty"`
	Memo        *string                    `json:"memo" validate:"omitempty"`
}

// RootAccountAddReq ...
type RootAccountAddReq struct {
	RootAccountCommonAddReq `json:",inline"`
	Extension               map[string]string `json:"extension" validate:"omitempty"`
}

// Validate ...
func (req *RootAccountAddReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	return nil
}

// -------------------------- Update --------------------------

// RootAccountUpdateReq ...
type RootAccountUpdateReq struct {
	Name        string          `json:"name" validate:"omitempty"`
	Managers    []string        `json:"managers" validate:"omitempty,max=5"`
	BakManagers []string        `json:"bak_managers" validate:"omitempty,max=5"`
	Memo        *string         `json:"memo" validate:"omitempty"`
	DeptID      int64           `json:"dept_id" validate:"required,min=1"`
	Extension   json.RawMessage `json:"extension" validate:"omitempty"`
}

// Validate ...
func (req *RootAccountUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	return nil
}

// AwsRootAccountExtensionUpdateReq	...
type AwsRootAccountExtensionUpdateReq struct {
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AwsRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// GcpAccountExtensionUpdateReq ...
type GcpRootAccountExtensionUpdateReq struct {
	CloudProjectName        string `json:"cloud_project_name" validate:"omitempty"`
	CloudServiceAccountID   string `json:"cloud_service_account_id" validate:"omitempty"`
	CloudServiceAccountName string `json:"cloud_service_account_name" validate:"omitempty"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id" validate:"omitempty"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *GcpRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// AzureAccountExtensionUpdateReq ...
type AzureRootAccountExtensionUpdateReq struct {
	CloudTenantID         string `json:"cloud_tenant_id" validate:"omitempty"`
	CloudSubscriptionName string `json:"cloud_subscription_name" validate:"omitempty"`
	CloudApplicationID    string `json:"cloud_application_id" validate:"omitempty"`
	CloudApplicationName  string `json:"cloud_application_name" validate:"omitempty"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *AzureRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// HuaWeiAccountExtensionUpdateReq ...
type HuaWeiRootAccountExtensionUpdateReq struct {
	CloudSubAccountName string `json:"cloud_sub_account_name" validate:"required"`
	CloudIamUserID      string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername    string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID       string `json:"cloud_secret_id" validate:"omitempty"`
	CloudSecretKey      string `json:"cloud_secret_key" validate:"omitempty"`
}

// Validate ...
func (req *HuaWeiRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// ZenlayerRootAccountExtensionUpdateReq ...
type ZenlayerRootAccountExtensionUpdateReq struct {
}

// Validate ...
func (req *ZenlayerRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// KaopuRootAccountExtensionUpdateReq ...
type KaopuRootAccountExtensionUpdateReq struct {
}

// Validate ...
func (req *KaopuRootAccountExtensionUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}
