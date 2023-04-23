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
	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// GetTCloudAccountZoneQuotaReq ...
type GetTCloudAccountZoneQuotaReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate ...
func (opt *GetTCloudAccountZoneQuotaReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// GetTCloudAccountZoneQuotaResp ...
type GetTCloudAccountZoneQuotaResp struct {
	rest.BaseResp `json:",inline"`
	Data          *typeaccount.TCloudAccountQuota `json:"data"`
}

// GetHuaWeiAccountRegionQuotaReq ...
type GetHuaWeiAccountRegionQuotaReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate ...
func (opt *GetHuaWeiAccountRegionQuotaReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// GetHuaWeiAccountQuotaResp ...
type GetHuaWeiAccountQuotaResp struct {
	rest.BaseResp `json:",inline"`
	Data          *typeaccount.HuaWeiAccountQuota `json:"data"`
}

// GetGcpAccountRegionQuotaReq ...
type GetGcpAccountRegionQuotaReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate ...
func (opt *GetGcpAccountRegionQuotaReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// GetGcpAccountQuotaResp ...
type GetGcpAccountQuotaResp struct {
	rest.BaseResp `json:",inline"`
	Data          *typeaccount.GcpProjectQuota `json:"data"`
}
