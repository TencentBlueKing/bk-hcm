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

// SyncTCloudSecurityGroupReq define sync tcloud sg and sg rule req.
type SyncTCloudSecurityGroupReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncTCloudSecurityGroupReq
func (req SyncTCloudSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncHuaWeiSecurityGroupReq define sync huawei sg and sg rule req.
type SyncHuaWeiSecurityGroupReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncHuaWeiSecurityGroupReq
func (req SyncHuaWeiSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAzureSecurityGroupReq define sync azure sg and sg rule req.
type SyncAzureSecurityGroupReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate SyncAzureSecurityGroupReq
func (req SyncAzureSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAwsSecurityGroupReq define sync aws sg and sg rule req.
type SyncAwsSecurityGroupReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncAwsSecurityGroupReq
func (req SyncAwsSecurityGroupReq) Validate() error {
	return validator.Validate.Struct(req)
}
