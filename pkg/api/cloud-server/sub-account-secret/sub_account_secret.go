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

// Package subaccountsecret defines cloud-server sub account secret api types.
package subaccountsecret

import (
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/validator"
)

// CreateReq defines create sub account secret request.
type CreateReq struct {
	SubAccountID string `json:"sub_account_id" validate:"required"`
}

// Validate create sub account secret request.
func (req *CreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateResult defines create sub account secret result.
type CreateResult struct {
	ID        string      `json:"id"`
	Extension interface{} `json:"extension"`
}

// TCloudCreateExtension defines tcloud create sub account secret extension in response.
type TCloudCreateExtension struct {
	CloudSecretID  string `json:"cloud_secret_id"`
	CloudSecretKey string `json:"cloud_secret_key"`
}

// ListSubAccountSecretReq defines list sub account secret request.
type ListSubAccountSecretReq struct {
	protocloud.SubAccountSecretFilters `json:",inline"`
	Page                               *core.BasePage `json:"page" validate:"required"`
}

// Validate list request.
func (req *ListSubAccountSecretReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.SubAccountSecretFilters.Validate(); err != nil {
		return err
	}

	return req.Page.Validate(core.NewDefaultPageOption())
}

// BizSubAccountSecretJoinExtDetail defines one row in biz-scoped sub account secret join list with operable.
type BizSubAccountSecretJoinExtDetail struct {
	protocloud.SubAccountSecretJoinExtDetail `json:",inline"`
	Operable                                 bool `json:"operable"`
}

// BizSubAccountSecretJoinExtListResult defines biz-scoped sub account secret join list response.
type BizSubAccountSecretJoinExtListResult struct {
	Count   uint64                             `json:"count"`
	Details []BizSubAccountSecretJoinExtDetail `json:"details"`
}
