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

package securitygroup

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// HuaWeiCreateOption define huawei security group create option.
type HuaWeiCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"required,lte=64"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate huawei security group create option.
func (opt HuaWeiCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update --------------------------

// HuaWeiUpdateOption define huawei security group update option.
type HuaWeiUpdateOption struct {
	CloudID     string  `json:"cloud_id" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"omitempty,lte=64"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate security group update option.
func (opt HuaWeiUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// HuaWeiListOption define huawei security group list option.
type HuaWeiListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.HuaWeiPage `json:"page" validate:"omitempty"`
}

// Validate huawei security group list option.
func (opt HuaWeiListOption) Validate() error {
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

// HuaWeiDeleteOption huawei security group delete option.
type HuaWeiDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate security group delete option.
func (opt HuaWeiDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Associate --------------------------

// HuaWeiAssociateCvmOption define security group bind cvm option.
type HuaWeiAssociateCvmOption struct {
	Region               string `json:"region" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
	CloudCvmID           string `json:"cloud_cvm_id" validate:"required"`
}

// Validate security group cvm bind option.
func (opt HuaWeiAssociateCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}
