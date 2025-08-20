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

// Package hccert ...
package hccert

import (
	"hcm/pkg/adaptor/types/core"
	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Delete --------------------------

// TCloudDeleteReq define delete cert req.
type TCloudDeleteReq struct {
	AccountID string `json:"account_id" validate:"required"`
	ID        string `json:"id" validate:"required"`
}

// Validate request.
func (req *TCloudDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create --------------------------

// TCloudCreateReq tcloud create req.
type TCloudCreateReq struct {
	BkBizID    int64             `json:"bk_biz_id" validate:"omitempty"`
	AccountID  string            `json:"account_id" validate:"required"`
	Vendor     string            `json:"vendor" validate:"required"`
	Name       string            `json:"name" validate:"required"`
	CertType   enumor.CertType   `json:"cert_type" validate:"required"`
	PublicKey  string            `json:"public_key" validate:"required"`
	PrivateKey string            `json:"private_key" validate:"omitempty"`
	Memo       string            `json:"memo"`
	Tags       []apicore.TagPair `json:"tags,omitempty"`
}

// Validate request.
func (req *TCloudCreateReq) Validate() error {
	if req.CertType == enumor.SVRServiceCertType && len(req.PrivateKey) == 0 {
		return errf.Newf(errf.InvalidParameter, "private_key is required when cert_type is SVR")
	}

	return validator.Validate.Struct(req)
}

// CreateResult ...
type CreateResult struct {
	CertificateID *string `json:"certificate_id"`
}

// CreateResp ...
type CreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateResult `json:"data"`
}

// -------------------------- List --------------------------

// TCloudListOption defines options to list tcloud instances.
type TCloudListOption struct {
	AccountID string           `json:"account_id" validate:"required"`
	Page      *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud list option.
func (opt TCloudListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}
