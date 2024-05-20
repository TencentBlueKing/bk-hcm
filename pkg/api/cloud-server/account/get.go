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

package account

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// SyncDetailRsp ...
type SyncDetailRsp struct {
	IassRes []IassResItem `json:"iass_res"`
}

// IassResItem ...
type IassResItem struct {
	ResName         string `json:"res_name"`
	ResStatus       string `json:"res_status"`
	ResFailedReason string `json:"res_failed_reason"`
	ResEndTime      string `json:"res_end_time"`
}

// BySecretResp 根据秘钥获取的字段
type BySecretResp[T cloud.AccountInfoBySecret] struct {
	rest.BaseResp `json:",inline"`
	Data          *T `json:"data"`
}

// TCloudAccountInfoBySecretReq ...
type TCloudAccountInfoBySecretReq struct {
	DisableCheck        bool `json:"disable_check" validate:"omitempty"`
	*cloud.TCloudSecret `json:",inline" validate:"required"`
}

// Validate ...
func (req *TCloudAccountInfoBySecretReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAccountInfoBySecretReq ...
type AwsAccountInfoBySecretReq struct {
	DisableCheck     bool `json:"disable_check" validate:"omitempty"`
	*cloud.AwsSecret `json:",inline" validate:"required"`
}

// Validate ...
func (req *AwsAccountInfoBySecretReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiAccountInfoBySecretReq ...
type HuaWeiAccountInfoBySecretReq struct {
	DisableCheck        bool `json:"disable_check" validate:"omitempty"`
	*cloud.HuaWeiSecret `json:",inline" validate:"required"`
}

// Validate ...
func (req *HuaWeiAccountInfoBySecretReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureAccountInfoBySecretReq ...
type AzureAccountInfoBySecretReq struct {
	DisableCheck       bool `json:"disable_check" validate:"omitempty"`
	*cloud.AzureSecret `json:",inline" validate:"required"`
}

// Validate ...
func (req *AzureAccountInfoBySecretReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpAccountInfoBySecretReq ...
type GcpAccountInfoBySecretReq struct {
	DisableCheck     bool `json:"disable_check" validate:"omitempty"`
	*cloud.GcpSecret `json:",inline" validate:"required"`
}

// Validate ...
func (req *GcpAccountInfoBySecretReq) Validate() error {
	return validator.Validate.Struct(req)
}
