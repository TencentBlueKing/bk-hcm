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
	"hcm/pkg/cryptography"
)

// BaseAccount 云账号
type BaseAccount struct {
	ID                 string                 `json:"id"`
	Vendor             enumor.Vendor          `json:"vendor"`
	Name               string                 `json:"name"`
	Managers           []string               `json:"managers"`
	Type               enumor.AccountType     `json:"type"`
	Site               enumor.AccountSiteType `json:"site"`
	Price              string                 `json:"price"`
	PriceUnit          string                 `json:"price_unit"`
	Memo               *string                `json:"memo"`
	BkBizIDs           []int64                `json:"bk_biz_ids"`
	SyncStatus         string                 `json:"sync_status"`
	SyncFailedReason   string                 `json:"sync_failed_reason"`
	RecycleReserveTime int                    `json:"recycle_reserve_time"`
	core.Revision      `json:",inline"`
}

// TCloudAccountExtension define tcloud account extension.
type TCloudAccountExtension struct {
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
	CloudSecretID      string `json:"cloud_secret_id"`
	CloudSecretKey     string `json:"cloud_secret_key,omitempty"`
}

// DecryptSecretKey ...
func (e *TCloudAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudSecretKey != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudSecretKey)
		if err != nil {
			return err
		}
		e.CloudSecretKey = plainSecretKey
	}

	return nil
}

// AwsAccountExtension define aws account extension.
type AwsAccountExtension struct {
	CloudAccountID   string `json:"cloud_account_id"`
	CloudIamUsername string `json:"cloud_iam_username"`
	CloudSecretID    string `json:"cloud_secret_id"`
	CloudSecretKey   string `json:"cloud_secret_key,omitempty"`
}

// DecryptSecretKey ...
func (e *AwsAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudSecretKey != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudSecretKey)
		if err != nil {
			return err
		}
		e.CloudSecretKey = plainSecretKey
	}

	return nil
}

// HuaWeiAccountExtension define huawei account extension.
type HuaWeiAccountExtension struct {
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudSubAccountID    string `json:"cloud_sub_account_id"`
	CloudSubAccountName  string `json:"cloud_sub_account_name"`
	CloudSecretID        string `json:"cloud_secret_id"`
	CloudSecretKey       string `json:"cloud_secret_key,omitempty"`
	CloudIamUserID       string `json:"cloud_iam_user_id" `
	CloudIamUsername     string `json:"cloud_iam_username"`
}

// DecryptSecretKey ...
func (e *HuaWeiAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudSecretKey != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudSecretKey)
		if err != nil {
			return err
		}
		e.CloudSecretKey = plainSecretKey
	}

	return nil
}

// GcpAccountExtension define gcp account extension.
type GcpAccountExtension struct {
	Email                   string `json:"email"`
	CloudProjectID          string `json:"cloud_project_id"`
	CloudProjectName        string `json:"cloud_project_name"`
	CloudServiceAccountID   string `json:"cloud_service_account_id"`
	CloudServiceAccountName string `json:"cloud_service_account_name"`
	CloudServiceSecretID    string `json:"cloud_service_secret_id"`
	CloudServiceSecretKey   string `json:"cloud_service_secret_key,omitempty"`
}

// DecryptSecretKey ...
func (e *GcpAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudServiceSecretKey != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudServiceSecretKey)
		if err != nil {
			return err
		}
		e.CloudServiceSecretKey = plainSecretKey
	}

	return nil
}

// AzureAccountExtension ...
type AzureAccountExtension struct {
	DisplayNameName       string `json:"display_name_name"`
	CloudTenantID         string `json:"cloud_tenant_id"`
	CloudSubscriptionID   string `json:"cloud_subscription_id"`
	CloudSubscriptionName string `json:"cloud_subscription_name"`
	CloudApplicationID    string `json:"cloud_application_id"`
	CloudApplicationName  string `json:"cloud_application_name"`
	CloudClientSecretID   string `json:"cloud_client_secret_id"`
	CloudClientSecretKey  string `json:"cloud_client_secret_key,omitempty"`
}

// DecryptSecretKey ...
func (e *AzureAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudClientSecretKey != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudClientSecretKey)
		if err != nil {
			return err
		}
		e.CloudClientSecretKey = plainSecretKey
	}

	return nil
}

// OtherAccountExtension define other account extension.
type OtherAccountExtension struct {
}

// DecryptSecretKey ...
func (o OtherAccountExtension) DecryptSecretKey(crypto cryptography.Crypto) error {
	return nil
}
