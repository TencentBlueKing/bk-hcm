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

import (
	"hcm/pkg/criteria/validator"
)

// SyncTCloudDiskReq define sync tcloud disk req.
type SyncTCloudDiskReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncTCloudDiskReq
func (req SyncTCloudDiskReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncHuaWeiDiskReq define sync huawei disk req.
type SyncHuaWeiDiskReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncHuaWeiDiskReq
func (req SyncHuaWeiDiskReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncGcpDiskReq define sync gcp disk req.
type SyncGcpDiskReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate SyncGcpDiskReq
func (req SyncGcpDiskReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAzureDiskReq define sync azure disk req.
type SyncAzureDiskReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate SyncAzureDiskReq
func (req SyncAzureDiskReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAwsDiskReq define sync aws disk req.
type SyncAwsDiskReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncAwsDiskReq
func (req SyncAwsDiskReq) Validate() error {
	return validator.Validate.Struct(req)
}
