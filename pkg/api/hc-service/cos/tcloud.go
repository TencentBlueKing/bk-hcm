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

// Package cos ...
package cos

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// TCloudCreateBucketReq tcloud create bucket req.
type TCloudCreateBucketReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Name      string `json:"name" validate:"required"`

	XCosACL                   string                     `json:"x_cos_acl" validate:"omitempty"`
	XCosGrantRead             string                     `json:"x_cos_grant_read" validate:"omitempty"`
	XCosGrantWrite            string                     `json:"x_cos_grant_write" validate:"omitempty"`
	XCosGrantFullControl      string                     `json:"x_cos_grant_full_control" validate:"omitempty"`
	XCosGrantReadACP          string                     `json:"x_cos_grant_read_acp" validate:"omitempty"`
	XCosGrantWriteACP         string                     `json:"x_cos_grant_write_acp" validate:"omitempty"`
	XCosTagging               string                     `json:"x_cos_tagging" validate:"omitempty"`
	CreateBucketConfiguration *CreateBucketConfiguration `json:"create_bucket_configuration" validate:"omitempty"`
}

// Validate TCloudCreateBucketReq.
func (req *TCloudCreateBucketReq) Validate() error {
	if req.CreateBucketConfiguration != nil {
		if err := req.CreateBucketConfiguration.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(req)
}

// CreateBucketConfiguration create bucket configuration.
type CreateBucketConfiguration struct {
	BucketAZConfig enumor.BucketAZConfig `json:"bucket_az_config" validate:"required"`
}

// Validate CreateBucketConfiguration.
func (c CreateBucketConfiguration) Validate() error {
	if err := c.BucketAZConfig.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(c)
}

// TCloudDeleteBucketReq tcloud delete bucket req.
type TCloudDeleteBucketReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

// Validate TCloudDeleteBucketReq.
func (req *TCloudDeleteBucketReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudBucketListReq tcloud list bucket request.
type TCloudBucketListReq struct {
	AccountID string `json:"account_id" validate:"required"`

	TagKey     *string `json:"tag_key" validate:"omitempty"`
	TagValue   *string `json:"tag_value" validate:"omitempty"`
	MaxKeys    *int64  `json:"max_keys" validate:"omitempty"`
	Marker     *string `json:"marker" validate:"omitempty"`
	Range      *string `json:"range" validate:"omitempty"`
	CreateTime *int64  `json:"create_time" validate:"omitempty"`
	Region     *string `json:"region" validate:"omitempty"`
}

// Validate TCloudBucketListReq.
func (c TCloudBucketListReq) Validate() error {
	return validator.Validate.Struct(c)
}
