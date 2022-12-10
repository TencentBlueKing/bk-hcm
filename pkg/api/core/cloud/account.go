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

// Account 云账号
type Account struct {
	ID         uint64             `json:"id"`
	Vendor     enumor.Vendor      `json:"vendor"`
	Spec       *AccountSpec       `json:"spec"`
	Extension  *AccountExtension  `json:"extension"`
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

// AccountExtension define account extension.
type AccountExtension struct {
	TCloud *TCloudAccountExtension `json:"tcloud,omitempty"`
	Aws    *AwsAccountExtension    `json:"aws,omitempty"`
	HuaWei *HuaWeiAccountExtension `json:"huawei,omitempty"`
	Gcp    *GcpAccountExtension    `json:"gcp,omitempty"`
	Azure  *AzureAccountExtension  `json:"azure,omitempty"`
}

// TCloudAccountExtension define tcloud account extension.
type TCloudAccountExtension struct {
	MainAccountCid string      `json:"main_account_cid"`
	SubAccountCid  string      `json:"sub_account_cid"`
	Secret         *BaseSecret `json:"secret"`
}

// AwsAccountExtension define aws account extension.
type AwsAccountExtension struct {
	AccountCid  string      `json:"account_cid"`
	IamUserName string      `json:"iam_user_name"`
	Secret      *BaseSecret `json:"secret"`
}

// HuaWeiAccountExtension define huawei account extension.
type HuaWeiAccountExtension struct {
	MainAccountName string      `json:"main_account_name"`
	SubAccountCid   string      `json:"sub_account_cid"`
	SubAccountName  string      `json:"sub_account_name"`
	IamUserCid      string      `json:"iam_user_cid"`
	IamUserName     string      `json:"iam_user_name"`
	Secret          *BaseSecret `json:"secret"`
}

// GcpAccountExtension define gcp account extension.
type GcpAccountExtension struct {
	ProjectName        string         `json:"project_name"`
	ServiceAccountCid  string         `json:"service_account_cid"`
	ServiceAccountName string         `json:"service_account_name"`
	Secret             *GcpCredential `json:"secret"`
}

// GcpCredential define gcp credential.
type GcpCredential struct {
	ProjectCid string `json:"project_cid"`
	Cid        string `json:"cid"`
	Key        string `json:"json"`
}

// AzureAccountExtension define azure credential.
type AzureAccountExtension struct {
	SubscriptionName string           `json:"subscription_name"`
	ApplicationName  string           `json:"application_name"`
	Secret           *AzureCredential `json:"secret"`
}

// AzureCredential define azure credential.
type AzureCredential struct {
	TenantCid       string `json:"tenant_cid"`
	SubscriptionCid string `json:"subscription_cid"`
	ClientCid       string `json:"client_cid"`
	ClientSecret    string `json:"client_secret"`
}

// BaseSecret define base secret.
type BaseSecret struct {
	Cid string `json:"cid"`
	Key string `json:"key"`
}
