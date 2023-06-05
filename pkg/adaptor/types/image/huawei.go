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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
)

// HuaWeiImageListOption ...
type HuaWeiImageListOption struct {
	Region   string                          `json:"region" validate:"required"`
	Platform model.ListImagesRequestPlatform `json:"platform" validate:"required"`
	CloudID  string                          `json:"cloud_id" validate:"omitempty"`
	Page     *core.HuaWeiPage                `json:"page" validate:"omitempty"`
}

// Validate huawei cvm list option.
func (opt HuaWeiImageListOption) Validate() error {
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

// HuaWeiImageListResult ...
type HuaWeiImageListResult struct {
	Details []HuaWeiImage `json:"details"`
}

// HuaWeiImage ...
type HuaWeiImage struct {
	CloudID      string `json:"cloud_id"`
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Platform     string `json:"platform"`
	State        string `json:"state"`
	Type         string `json:"type"`
}

// GetCloudID ...
func (image HuaWeiImage) GetCloudID() string {
	return image.CloudID
}
