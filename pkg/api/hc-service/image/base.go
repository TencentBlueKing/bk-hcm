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

package image

import (
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/criteria/validator"
)

// TCloudImageSyncReq image sync request
type TCloudImageSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate image sync request.
func (req *TCloudImageSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiImageSyncReq image sync request
type HuaWeiImageSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate image sync request.
func (req *HuaWeiImageSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsImageSyncReq image sync request
type AwsImageSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate image sync request.
func (req *AwsImageSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureImageSyncReq image sync request
type AzureImageSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	Region            string `json:"region" validate:"omitempty"`
	ResourceGroupName string `json:"resource_group_name" validate:"omitempty"`
}

// Validate image sync request.
func (req *AzureImageSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpImageSyncReq image sync request
type GcpImageSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate image sync request.
func (req *GcpImageSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudImageListOption ...
type TCloudImageListOption struct {
	AccountID                    string `json:"account_id" validate:"required"`
	*image.TCloudImageListOption `json:",inline"`
}

// Validate image list option.
func (opt *TCloudImageListOption) Validate() error {
	err := opt.TCloudImageListOption.Validate()
	if err != nil {
		return err
	}
	return validator.Validate.Struct(opt)
}
