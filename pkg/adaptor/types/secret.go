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

package types

import "hcm/pkg/criteria/errf"

// BaseSecret defines the hybrid cloud's base secret info.
type BaseSecret struct {
	// CloudSecretID is the secret id to do credential.
	CloudSecretID string `json:"cloud_secret_id"`
	// CloudSecretKey is the secret key to do credential.
	CloudSecretKey string `json:"cloud_secret_key"`
}

// Validate BaseSecret.
func (b BaseSecret) Validate() error {
	if len(b.CloudSecretID) == 0 {
		return errf.New(errf.InvalidParameter, "secret id is required")
	}

	if len(b.CloudSecretKey) == 0 {
		return errf.New(errf.InvalidParameter, "secret key is required")
	}

	return nil
}

// GcpCredential define gcp credential information.
type GcpCredential struct {
	CloudProjectID string `json:"cloud_project_id"`
	Json           []byte `json:"json,omitempty"`
}

// AzureCredential define azure credential information.
type AzureCredential struct {
	CloudTenantID       string `json:"cloud_tenant_id"`
	CloudSubscriptionID string `json:"cloud_subscription_id"`
	CloudClientID       string `json:"cloud_client_id"`
	CloudClientSecret   string `json:"cloud_client_secret"`
}

// Validate AzureCredential.
func (a AzureCredential) Validate() error {
	if len(a.CloudTenantID) == 0 {
		return errf.New(errf.InvalidParameter, "tenant id is required")
	}

	if len(a.CloudSubscriptionID) == 0 {
		return errf.New(errf.InvalidParameter, "subscription id is required")
	}

	if len(a.CloudClientID) == 0 {
		return errf.New(errf.InvalidParameter, "client id is required")
	}

	if len(a.CloudClientSecret) == 0 {
		return errf.New(errf.InvalidParameter, "client secret is required")
	}

	return nil
}
