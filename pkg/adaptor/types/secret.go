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

// Secret defines the hybrid cloud's secret info.
type Secret struct {
	TCloud *BaseSecret      `json:"tcloud,omitempty"`
	Aws    *BaseSecret      `json:"aws,omitempty"`
	HuaWei *BaseSecret      `json:"huawei,omitempty"`
	Azure  *AzureCredential `json:"azure,omitempty"`
	Gcp    *GcpCredential   `json:"gcp,omitempty"`
}

// BaseSecret defines the hybrid cloud's base secret info.
type BaseSecret struct {
	// ID is the secret id to do credential
	ID string `json:"id,omitempty"`
	// Key is the secret key to do credential
	Key string `json:"key,omitempty"`
}

// Validate BaseSecret.
func (b BaseSecret) Validate() error {
	if len(b.ID) == 0 {
		return errf.New(errf.InvalidParameter, "secret id is required")
	}

	if len(b.Key) == 0 {
		return errf.New(errf.InvalidParameter, "secret key is required")
	}

	return nil
}

// GcpCredential define gcp credential information.
type GcpCredential struct {
	ProjectID string `json:"project_id,omitempty"`
	Json      []byte `json:"json,omitempty"`
}

// Validate GcpCredential.
func (g GcpCredential) Validate() error {
	if len(g.ProjectID) == 0 {
		return errf.New(errf.InvalidParameter, "project id is required")
	}

	if len(g.Json) == 0 {
		return errf.New(errf.InvalidParameter, "credential json is required")
	}

	return nil
}

// AzureCredential define azure credential information.
type AzureCredential struct {
	TenantID       string `json:"tenant_id,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
}

// Validate AzureCredential.
func (a AzureCredential) Validate() error {
	if len(a.TenantID) == 0 {
		return errf.New(errf.InvalidParameter, "tenant id is required")
	}

	if len(a.SubscriptionID) == 0 {
		return errf.New(errf.InvalidParameter, "subscription id is required")
	}

	if len(a.ClientID) == 0 {
		return errf.New(errf.InvalidParameter, "client id is required")
	}

	if len(a.ClientSecret) == 0 {
		return errf.New(errf.InvalidParameter, "client secret is required")
	}

	return nil
}
