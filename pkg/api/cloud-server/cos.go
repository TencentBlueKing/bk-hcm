/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package cloudserver

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// TCloudBizCreateCosBucketReq 业务视角创建 COS 存储桶请求
type TCloudBizCreateCosBucketReq struct {
	AccountID  string `json:"account_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Manager    string `json:"manager" validate:"required"`
	BakManager string `json:"bak_manager" validate:"required"`

	XCosACL                   string                        `json:"x_cos_acl" validate:"omitempty"`
	XCosGrantRead             string                        `json:"x_cos_grant_read" validate:"omitempty"`
	XCosGrantWrite            string                        `json:"x_cos_grant_write" validate:"omitempty"`
	XCosGrantFullControl      string                        `json:"x_cos_grant_full_control" validate:"omitempty"`
	XCosGrantReadACP          string                        `json:"x_cos_grant_read_acp" validate:"omitempty"`
	XCosGrantWriteACP         string                        `json:"x_cos_grant_write_acp" validate:"omitempty"`
	CreateBucketConfiguration *BizCreateBucketConfiguration `json:"create_bucket_configuration" validate:"omitempty"`
}

// BizCreateBucketConfiguration 创建存储桶配置
type BizCreateBucketConfiguration struct {
	BucketAZConfig enumor.BucketAZConfig `json:"bucket_az_config" validate:"required"`
}

// Validate TCloudBizCreateCosBucketReq.
func (req *TCloudBizCreateCosBucketReq) Validate() error {
	if req.CreateBucketConfiguration != nil {
		if err := req.CreateBucketConfiguration.BucketAZConfig.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// TCloudBizDeleteCosBucketReq 业务视角删除 COS 存储桶请求
type TCloudBizDeleteCosBucketReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	CloudName string `json:"cloud_name" validate:"required"`
}

// Validate TCloudBizDeleteCosBucketReq.
func (req *TCloudBizDeleteCosBucketReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudBizListCosBucketReq 业务视角查询 COS 存储桶列表请求
type TCloudBizListCosBucketReq struct {
	// AccountID 账号 ID
	AccountID string `json:"account_id" validate:"required"`
	// Region 支持根据地域过滤存储桶
	Region string `json:"region" validate:"required"`
	// MaxKeys 单次返回最大的条目数量，默认值为2000，最大为2000。
	MaxKeys *int64 `json:"max_keys" validate:"omitempty"`
	// Marker 起始标记，从该标记之后（不含）按照 UTF-8 字典序返回存储桶条目
	Marker *string `json:"marker" validate:"omitempty"`
	// Range 和 create-time 参数一起使用，支持根据创建时间过滤存储桶，支持枚举值 lt、gt、lte、gte
	Range *string `json:"range" validate:"omitempty"`
	// CreateTime GMT 时间戳，和 range 参数一起使用，支持根据创建时间过滤存储桶
	CreateTime *int64 `json:"create_time" validate:"omitempty"`
}

// Validate TCloudBizListCosBucketReq.
func (req *TCloudBizListCosBucketReq) Validate() error {
	return validator.Validate.Struct(req)
}
