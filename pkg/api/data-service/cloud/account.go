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
	"encoding/json"

	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	jsoniter "github.com/json-iterator/go"
)

// -------------------------- Create --------------------------

type CreateAccountExtensionReq interface {
	CreateTCloudAccountExtensionReq | CreateAwsAccountExtensionReq | CreateHuaWeiAccountExtensionReq | CreateGcpAccountExtensionReq | CreateAzureAccountExtensionReq
}

// CreateTCloudAccountExtensionReq ...
type CreateTCloudAccountExtensionReq struct {
	MainAccountID string `json:"main_account_id" validate:"required"`
	SubAccountID  string `json:"sub_account_id" validate:"required"`
	SecretID      string `json:"secret_id" validate:"required"`
	SecretKey     string `json:"secret_key" validate:"required"`
}

// CreateAwsAccountExtensionReq ...
type CreateAwsAccountExtensionReq struct {
	AccountID   string `json:"account_id" validate:"required"`
	IamUsername string `json:"iam_username" validate:"required"`
	SecretID    string `json:"secret_id" validate:"required"`
	SecretKey   string `json:"secret_key" validate:"required"`
}

// CreateHuaWeiAccountExtensionReq ...
type CreateHuaWeiAccountExtensionReq struct {
	MainAccountName string `json:"main_account_name" validate:"required"`
	SubAccountID    string `json:"sub_account_id" validate:"required"`
	SubAccountName  string `json:"sub_account_name" validate:"required"`
	SecretID        string `json:"secret_id" validate:"required"`
	SecretKey       string `json:"secret_key" validate:"required"`
}

// CreateGcpAccountExtensionReq ...
type CreateGcpAccountExtensionReq struct {
	ProjectID          string `json:"project_id" validate:"required"`
	ProjectName        string `json:"project_name" validate:"required"`
	ServiceAccountID   string `json:"service_account_cid" validate:"required"`
	ServiceAccountName string `json:"service_account_name" validate:"required"`
	ServiceSecretID    string `json:"service_secret_id" validate:"required"`
	ServiceSecretKey   string `json:"service_secret_key" validate:"required"`
}

// CreateAzureAccountExtensionReq ...
type CreateAzureAccountExtensionReq struct {
	TenantID         string `json:"tenant_id" validate:"required"`
	SubscriptionID   string `json:"subscription_id" validate:"required"`
	SubscriptionName string `json:"subscription_name" validate:"required"`
	ApplicationID    string `json:"application_id" validate:"required"`
	ApplicationName  string `json:"application_name" validate:"required"`
	ClientID         string `json:"client_id" validate:"required"`
	ClientSecret     string `json:"client_secret" validate:"required"`
}

// CreateAccountSpecReq ...
type CreateAccountSpecReq struct {
	Name         string                 `json:"name" validate:"required"`
	Managers     []string               `json:"managers" validate:"required"`
	DepartmentID int64                  `json:"department_id" validate:"required"`
	Type         enumor.AccountType     `json:"type" validate:"required"`
	Site         enumor.AccountSiteType `json:"site" validate:"required"`
	Memo         *string                `json:"memo" validate:"required"`
}

// CreateAccountAttachmentReq ...
type CreateAccountAttachmentReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"required"`
}

// CreateAccountReq ...
type CreateAccountReq[T CreateAccountExtensionReq] struct {
	Spec       *CreateAccountSpecReq       `json:"spec" validate:"required"`
	Extension  *T                          `json:"extension" validate:"required"`
	Attachment *CreateAccountAttachmentReq `json:"attachment" validate:"required"`
}

// Validate ...
func (c *CreateAccountReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

type UpdateAccountExtensionReq interface {
	UpdateTCloudAccountExtensionReq | UpdateAwsAccountExtensionReq | UpdateHuaWeiAccountExtensionReq | UpdateGcpAccountExtensionReq | UpdateAzureAccountExtensionReq
}

type UpdateTCloudAccountExtensionReq struct {
	SecretID  string `json:"secret_id" validate:"omitempty"`
	SecretKey string `json:"secret_key" validate:"omitempty"`
}

type UpdateAwsAccountExtensionReq struct {
	SecretID  string `json:"secret_id" validate:"omitempty"`
	SecretKey string `json:"secret_key" validate:"omitempty"`
}
type UpdateHuaWeiAccountExtensionReq struct {
	SecretID  string `json:"secret_id" validate:"omitempty"`
	SecretKey string `json:"secret_key" validate:"omitempty"`
}
type UpdateGcpAccountExtensionReq struct {
	ServiceSecretID  string `json:"service_secret_id" validate:"required"`
	ServiceSecretKey string `json:"service_secret_key" validate:"required"`
}
type UpdateAzureAccountExtensionReq struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// UpdateAccountSpecReq ...
type UpdateAccountSpecReq struct {
	Name         string   `json:"name" validate:"omitempty"`
	Managers     []string `json:"managers" validate:"omitempty"`
	DepartmentID int64    `json:"department_id" validate:"omitempty"`
	SyncStatus   string   `json:"sync_status" validate:"omitempty"`
	Price        string   `json:"price" validate:"omitempty"`
	PriceUnit    string   `json:"price_unit" validate:"omitempty"`
	Memo         *string  `json:"memo" validate:"omitempty"`
}

// UpdateAccountReq ...
type UpdateAccountReq[T UpdateAccountExtensionReq] struct {
	Spec      *UpdateAccountSpecReq `json:"spec" validate:"omitempty"`
	Extension *T                    `json:"extension" validate:"omitempty"`
}

// Validate ...
func (u *UpdateAccountReq[T]) Validate() error {
	return validator.Validate.Struct(u)
}

// ExtensionToMap ...
func (u *UpdateAccountReq[T]) ExtensionToMap() (m map[string]interface{}, err error) {
	if u.Extension == nil {
		return
	}
	b, err := jsoniter.Marshal(u.Extension)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &m)
	return
}

// -------------------------- List --------------------------

// ListAccountReq ...
type ListAccountReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate ...
func (l *ListAccountReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ListBaseAccountReq ...
type ListBaseAccountReq struct {
	ID     uint64             `json:"id"`
	Vendor enumor.Vendor      `json:"vendor"`
	Spec   *cloud.AccountSpec `json:"spec"`
}

// ListAccountResult defines list instances for iam pull resource callback result.
type ListAccountResult struct {
	Count uint64 `json:"count,omitempty"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []*ListBaseAccountReq `json:"details,omitempty"`
}

// -------------------------- Delete --------------------------

// DeleteAccountReq ...
type DeleteAccountReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (d *DeleteAccountReq) Validate() error {
	return validator.Validate.Struct(d)
}

// ListAccountResp ...
type ListAccountResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListAccountResult `json:"data"`
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
