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

package cloud

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseAccount 云账号
type BaseAccount struct {
	ID         uint64             `json:"id"`
	Vendor     enumor.Vendor      `json:"vendor"`
	Spec       *AccountSpec       `json:"spec"`
	Attachment *AccountAttachment `json:"attachment,omitempty"`
	Revision   *core.Revision     `json:"revision"`
}

// AccountSpec define account spec.
type AccountSpec struct {
	Name         string                   `json:"name"`
	Managers     []string                 `json:"managers"`
	DepartmentID int64                    `json:"department_id"`
	Type         enumor.AccountType       `json:"type"`
	Site         enumor.AccountSiteType   `json:"site"`
	SyncStatus   enumor.AccountSyncStatus `json:"sync_status"`
	Price        string                   `json:"price"`
	PriceUnit    string                   `json:"price_unit"`
	Memo         *string                  `json:"memo"`
}

// AccountAttachment account attachment.
type AccountAttachment struct {
	BkBizIDs []int64 `json:"bk_biz_ids"`
}

// TCloudAccountExtension define tcloud account extension.
type TCloudAccountExtension struct {
	MainAccountID string `json:"main_account_id"`
	SubAccountID  string `json:"sub_account_id"`
	SecretID      string `json:"secret_id"`
	SecretKey     string `json:"secret_key"`
}

// TCloudAccount ...
type TCloudAccount struct {
	BaseAccount
	Extension *TCloudAccountExtension `json:"extension"`
}

// AwsAccountExtension define aws account extension.
type AwsAccountExtension struct {
	AccountID   string `json:"account_id"`
	IamUsername string `json:"iam_username"`
	SecretID    string `json:"secret_id"`
	SecretKey   string `json:"secret_key"`
}

// AwsAccount ...
type AwsAccount struct {
	BaseAccount
	Extension *AwsAccountExtension `json:"extension"`
}

// HuaWeiAccountExtension define huawei account extension.
type HuaWeiAccountExtension struct {
	MainAccountName string `json:"main_account_name"`
	SubAccountID    string `json:"sub_account_id"`
	SubAccountName  string `json:"sub_account_name"`
	SecretID        string `json:"secret_id"`
	SecretKey       string `json:"secret_key"`
}

// HuaWeiAccount ...
type HuaWeiAccount struct {
	BaseAccount
	Extension *HuaWeiAccountExtension `json:"extension"`
}

// GcpAccountExtension define gcp account extension.
type GcpAccountExtension struct {
	ProjectID          string `json:"project_id"`
	ProjectName        string `json:"project_name"`
	ServiceAccountID   string `json:"service_account_cid"`
	ServiceAccountName string `json:"service_account_name"`
	ServiceSecretID    string `json:"service_secret_id"`
	ServiceSecretKey   string `json:"service_secret_key"`
}

// GcpAccount ...
type GcpAccount struct {
	BaseAccount
	Extension *GcpAccountExtension `json:"extension"`
}

// AzureAccountExtension ...
type AzureAccountExtension struct {
	TenantID         string `json:"tenant_id"`
	SubscriptionID   string `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
	ApplicationID    string `json:"application_id"`
	ApplicationName  string `json:"application_name"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
}
