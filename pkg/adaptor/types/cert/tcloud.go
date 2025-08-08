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

package cert

import (
	"hcm/pkg/adaptor/types/core"
	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// -------------------------- List --------------------------

// TCloudListOption defines options to list tcloud cert instances.
type TCloudListOption struct {
	SearchKey string           `json:"search_key" validate:"omitempty"`
	Page      *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud cert list option.
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

// -------------------------- Delete --------------------------

// TCloudDeleteOption defines options to operation tcloud cert instances.
type TCloudDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate tcloud cert operation option.
func (opt TCloudDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------

// TCloudCreateOption defines options to create aws cert instances.
type TCloudCreateOption struct {
	Name       string            `json:"name" validate:"required"`
	CertType   string            `json:"cert_type" validate:"required"`
	PublicKey  string            `json:"public_key" validate:"required"`
	PrivateKey string            `json:"private_key" validate:"omitempty"`
	Repeatable bool              `json:"repeatable" validate:"omitempty"`
	Tags       []apicore.TagPair `json:"tags,omitempty"`
}

// Validate tcloud cert operation option.
func (opt TCloudCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudCert for cert Instance
type TCloudCert struct {
	*ssl.Certificates
}

// GetCloudID ...
func (cert TCloudCert) GetCloudID() string {
	return converter.PtrToVal(cert.CertificateId)
}
