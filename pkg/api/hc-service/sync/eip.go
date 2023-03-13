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

package sync

import "hcm/pkg/criteria/validator"

// SyncTCloudEipReq define sync tcloud eip req.
type SyncTCloudEipReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncTCloudEipReq
func (req SyncTCloudEipReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncHuaWeiEipReq define sync huawei eip req.
type SyncHuaWeiEipReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiEipReq
func (req SyncHuaWeiEipReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncGcpEipReq define sync gcp eip req.
type SyncGcpEipReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncGcpEipReq
func (req SyncGcpEipReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAzureEipReq ...
type SyncAzureEipReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAzureEipReq
func (req SyncAzureEipReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAwsEipReq define sync aws eip req.
type SyncAwsEipReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAwsEipReq
func (req SyncAwsEipReq) Validate() error {
	return validator.Validate.Struct(req)
}
