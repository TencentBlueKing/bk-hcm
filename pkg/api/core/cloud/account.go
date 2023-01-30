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
	ID         string             `json:"id"`
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
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
	CloudSecretID      string `json:"cloud_secret_id"`
	CloudSecretKey     string `json:"cloud_secret_key"`
}

// TCloudAccount ...
type TCloudAccount struct {
	BaseAccount `json:",inline"`
	Extension   *TCloudAccountExtension `json:"extension"`
}

// AwsAccountExtension define aws account extension.
type AwsAccountExtension struct {
	CloudAccountID   string `json:"cloud_account_id"`
	CloudIamUsername string `json:"cloud_iam_username"`
	CloudSecretID    string `json:"cloud_secret_id"`
	CloudSecretKey   string `json:"cloud_secret_key"`
}

// AwsAccount ...
type AwsAccount struct {
	BaseAccount `json:",inline"`
	Extension   *AwsAccountExtension `json:"extension"`
}

// HuaWeiAccountExtension define huawei account extension.
type HuaWeiAccountExtension struct {
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudSubAccountID    string `json:"cloud_sub_account_id"`
	CloudSubAccountName  string `json:"cloud_sub_account_name"`
	CloudSecretID        string `json:"cloud_secret_id"`
	CloudSecretKey       string `json:"cloud_secret_key"`
	CloudIamUserID       string `json:"cloud_iam_user_id" `
	CloudIamUsername     string `json:"cloud_iam_username"`
}

// HuaWeiAccount ...
type HuaWeiAccount struct {
	BaseAccount `json:",inline"`
	Extension   *HuaWeiAccountExtension `json:"extension"`
}

// GcpAccountExtension define gcp account extension.
type GcpAccountExtension struct {
	CloudProjectID          string `json:"cloud_project_id"`
	CloudProjectName        string `json:"cloud_project_name"`
	CloudServiceAccountID   string `json:"cloud_service_account_id"`
	CloudServiceAccountName string `json:"cloud_service_account_name"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key"`
}

// GcpAccount ...
type GcpAccount struct {
	BaseAccount `json:",inline"`
	Extension   *GcpAccountExtension `json:"extension"`
}

// AzureAccountExtension ...
type AzureAccountExtension struct {
	CloudTenantID         string `json:"cloud_tenant_id"`
	CloudSubscriptionID   string `json:"cloud_subscription_id"`
	CloudSubscriptionName string `json:"cloud_subscription_name"`
	CloudApplicationID    string `json:"cloud_application_id"`
	CloudApplicationName  string `json:"cloud_application_name"`
	CloudClientSecretID   string `json:"cloud_client_secret_id"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key"`
}
