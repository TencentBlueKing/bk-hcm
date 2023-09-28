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
	coreimage "hcm/pkg/api/core/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// BatchCreateReq define batch create req.
type BatchCreateReq[T coreimage.Extension] struct {
	Items []ImageCreate[T] `json:"items" validate:"required,min=1,max=100"`
}

// Validate ...
func (req *BatchCreateReq[T]) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, one := range req.Items {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ImageCreate ...
type ImageCreate[T coreimage.Extension] struct {
	CloudID      string        `json:"cloud_id"`
	Name         string        `json:"name"`
	Architecture string        `json:"architecture"`
	Platform     string        `json:"platform"`
	State        string        `json:"state"`
	Type         string        `json:"type"`
	OsType       enumor.OsType `json:"os_type"`
	Extension    *T            `json:"extension"`
}

// Validate ...
func (req *ImageCreate[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchUpdateReq define batch update req.
type BatchUpdateReq[T coreimage.Extension] struct {
	Items []ImageUpdate[T] `json:"items,min=1,max=100"`
}

// Validate ...
func (req *BatchUpdateReq[T]) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, one := range req.Items {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ImageUpdate ...
type ImageUpdate[T coreimage.Extension] struct {
	ID        string        `json:"id" validate:"required"`
	State     string        `json:"state"`
	OsType    enumor.OsType `json:"os_type"`
	Extension *T            `json:"extension"`
}

// Validate ...
func (req *ImageUpdate[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// DeleteReq ...
type DeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *DeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
