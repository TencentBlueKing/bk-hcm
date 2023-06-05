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
	"hcm/pkg/criteria/validator"
)

// AzureImageListResult ...
type AzureImageListResult struct {
	Count   *uint64      `json:"count,omitempty"`
	Details []AzureImage `json:"details"`
}

// AzureImage ...
type AzureImage struct {
	CloudID      string `json:"cloud_id"`
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Platform     string `json:"platform"`
	State        string `json:"state"`
	Type         string `json:"type"`
	Sku          string `json:"sku"`
	ImageSize    int64  `json:"image_size"`
	ImageSource  string `json:"image_source"`
}

// GetCloudID ...
func (image AzureImage) GetCloudID() string {
	return image.CloudID
}

// AzureImageListOption define tcloud image list option.
type AzureImageListOption struct {
	Region    string `json:"region" validate:"required"`
	Publisher string `json:"publisher" validate:"required"`
	Offer     string `json:"offer" validate:"required"`
}

// Validate tcloud image option.
func (opt AzureImageListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}
