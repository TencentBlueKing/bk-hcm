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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// ImageExtCreateReq ...
type ImageExtCreateReq[T ImageExtensionCreateReq] struct {
	CloudID      string `json:"cloud_id"`
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Platform     string `json:"platform"`
	State        string `json:"state"`
	Type         string `json:"type"`
	Extension    *T     `json:"extension"`
}

// Validate ...
func (req *ImageExtCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ImageExtensionCreateReq ...
type ImageExtensionCreateReq interface {
	TCloudImageExtensionCreateReq | AwsImageExtensionCreateReq | GcpImageExtensionCreateReq | HuaWeiImageExtensionCreateReq | AzureImageExtensionCreateReq
}

// ImageExtBatchCreateReq ...
type ImageExtBatchCreateReq[T ImageExtensionCreateReq] []*ImageExtCreateReq[T]

// Validate ...
func (req *ImageExtBatchCreateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ImageListReq ...
type ImageListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *ImageListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ImageExtUpdateReq ...
type ImageExtUpdateReq[T ImageExtensionUpdateReq] struct {
	ID        string `json:"id" validate:"required"`
	State     string `json:"state"`
	Extension *T     `json:"extension"`
}

// Validate ...
func (req *ImageExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ImageExtensionUpdateReq ...
type ImageExtensionUpdateReq interface {
	TCloudImageExtensionUpdateReq | AwsImageExtensionUpdateReq | GcpImageExtensionUpdateReq | HuaWeiImageExtensionUpdateReq | AzureImageExtensionUpdateReq
}

// ImageExtBatchUpdateReq ...
type ImageExtBatchUpdateReq[T ImageExtensionUpdateReq] []*ImageExtUpdateReq[T]

// Validate ...
func (req *ImageExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ImageDeleteReq ...
type ImageDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *ImageDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
