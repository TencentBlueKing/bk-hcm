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

package hcservice

import (
	"hcm/pkg/criteria/validator"
)

// TCloudAccountCheckReq ...
type TCloudAccountCheckReq struct {
	MainAccountID string `json:"main_account_id" validate:"required"`
	SubAccountID  string `json:"sub_account_id" validate:"required"`
	SecretID      string `json:"secret_id" validate:"required"`
	SecretKey     string `json:"secret_key" validate:"required"`
}

// Validate ...
func (r *TCloudAccountCheckReq) Validate() error {
	// TODO: 是否还需要添加其他规则校验呢？
	return validator.Validate.Struct(r)
}

// AwsAccountCheckReq ...
type AwsAccountCheckReq struct {
	AccountID   string `json:"account_id" validate:"required"`
	IamUsername string `json:"iam_username" validate:"required"`
	SecretID    string `json:"secret_id" validate:"required"`
	SecretKey   string `json:"secret_key" validate:"required"`
}

// Validate ...
func (r *AwsAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiAccountCheckReq ...
type HuaWeiAccountCheckReq struct {
	MainAccountName string `json:"main_account_name" validate:"required"`
	SubAccountID    string `json:"sub_account_id" validate:"required"`
	SubAccountName  string `json:"sub_account_name" validate:"required"`
	SecretID        string `json:"secret_id" validate:"required"`
	SecretKey       string `json:"secret_key" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}

// // GcpAccountCheckReq ...
// type GcpAccountCheckReq struct {
// 	ProjectID          string `json:"project_id" validate:"required"`
// 	ProjectName        string `json:"project_name" validate:"required"`
// 	ServiceAccountID   string `json:"service_account_cid" validate:"required"`
// 	ServiceAccountName string `json:"service_account_name" validate:"required"`
// 	ServiceSecretID    string `json:"service_secret_id" validate:"required"`
// 	ServiceSecretKey   string `json:"service_secret_key" validate:"required"`
// }
//
// // Validate ...
// func (r *GcpAccountCheckReq) Validate() error {
// 	return validator.Validate.Struct(r)
// }
//
// // AzureAccountCheckReq ...
// type AzureAccountCheckReq struct {
// 	TenantID         string `json:"tenant_id" validate:"required"`
// 	SubscriptionID   string `json:"subscription_id" validate:"required"`
// 	SubscriptionName string `json:"subscription_name" validate:"required"`
// 	ApplicationID    string `json:"application_id" validate:"required"`
// 	ApplicationName  string `json:"application_name" validate:"required"`
// 	ClientID         string `json:"client_id" validate:"required"`
// 	ClientSecret     string `json:"client_secret" validate:"required"`
// }
//
// // Validate ...
// func (r *AzureAccountCheckReq) Validate() error {
// 	return validator.Validate.Struct(r)
// }
