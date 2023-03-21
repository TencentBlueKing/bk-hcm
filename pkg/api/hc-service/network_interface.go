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

import "hcm/pkg/criteria/validator"

// -------------------------- Sync --------------------------

// AzureNetworkInterfaceSyncReq defines sync resource request.
type AzureNetworkInterfaceSyncReq struct {
	AccountID            string   `json:"account_id" validate:"required"`
	ResourceGroupName    string   `json:"resource_group_name" validate:"omitempty"`
	NetworkInterfaceName string   `json:"network_interface_name" validate:"omitempty"`
	CloudIDs             []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *AzureNetworkInterfaceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpNetworkInterfaceSyncReq defines sync resource request.
type GcpNetworkInterfaceSyncReq struct {
	AccountID   string   `json:"account_id" validate:"required"`
	Zone        string   `json:"zone" validate:"required"`
	CloudCvmIDs []string `json:"cloud_cvm_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *GcpNetworkInterfaceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiNetworkInterfaceSyncReq defines sync resource request.
type HuaWeiNetworkInterfaceSyncReq struct {
	AccountID   string   `json:"account_id" validate:"required"`
	Region      string   `json:"region" validate:"required"`
	CloudCvmIDs []string `json:"cloud_cvm_ids" validate:"required"`
}

// Validate validate sync vpc request.
func (r *HuaWeiNetworkInterfaceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// NetworkInterfaceSyncResult defines sync result.
type NetworkInterfaceSyncResult struct {
	TaskID string `json:"task_id"`
}
