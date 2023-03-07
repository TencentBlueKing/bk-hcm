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

package cvm

import (
	"fmt"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// GcpListOption defines options to list gcp cvm instances.
type GcpListOption struct {
	Region   string        `json:"region" validate:"required"`
	Zone     string        `json:"zone" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Page     *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate gcp cvm list option.
func (opt GcpListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if len(opt.CloudIDs) != 0 && opt.Page != nil {
		return fmt.Errorf("list by cloud_ids not support page")
	}

	if len(opt.CloudIDs) > core.GcpQueryLimit {
		return fmt.Errorf("cloud_ids should <= %d", core.GcpQueryLimit)
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// GcpDeleteOption defines options to operation cvm instances.
type GcpDeleteOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Stop --------------------------

// GcpStopOption defines options to operation cvm instances.
type GcpStopOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// GcpStartOption defines options to operation cvm instances.
type GcpStartOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reset --------------------------

// GcpResetOption defines options to operation cvm instances.
type GcpResetOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpResetOption) Validate() error {
	return validator.Validate.Struct(opt)
}
