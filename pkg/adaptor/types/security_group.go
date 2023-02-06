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

package types

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Common --------------------------

// SecurityGroupDeleteOption security group delete option.
type SecurityGroupDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate security group delete option.
func (opt SecurityGroupDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- TCloud --------------------------

// TCloudSecurityGroupCreateOption define security group create option.
type TCloudSecurityGroupCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"required,lte=60"`
	Description *string `json:"description" validate:"omitempty,lte=100"`
}

// Validate security group create option.
func (opt TCloudSecurityGroupCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudSecurityGroupUpdateOption define tcloud security group update option.
type TCloudSecurityGroupUpdateOption struct {
	CloudID     string  `json:"cloud_id" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"omitempty,lte=60"`
	Description *string `json:"description" validate:"omitempty,lte=100"`
}

// Validate security group update option.
func (opt TCloudSecurityGroupUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudSecurityGroupListOption define tcloud security group list option.
type TCloudSecurityGroupListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud security group list option.
func (opt TCloudSecurityGroupListOption) Validate() error {
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

// -------------------------- Aws --------------------------

// AwsSecurityGroupCreateOption define aws security group create option.
type AwsSecurityGroupCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	CloudVpcID  string  `json:"cloud_vpc_id" validate:"omitempty"`
	Name        string  `json:"name" validate:"required,lte=255"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate aws security group create option.
func (opt AwsSecurityGroupCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AwsSecurityGroupListOption define aws security group list option.
type AwsSecurityGroupListOption struct {
	Region   string        `json:"region" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Page     *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate security group list option.
func (opt AwsSecurityGroupListOption) Validate() error {
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

// -------------------------- HuaWei --------------------------

// HuaWeiSecurityGroupCreateOption define huawei security group create option.
type HuaWeiSecurityGroupCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"required,lte=64"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate huawei security group create option.
func (opt HuaWeiSecurityGroupCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiSecurityGroupUpdateOption define huawei security group update option.
type HuaWeiSecurityGroupUpdateOption struct {
	CloudID     string  `json:"cloud_id" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"omitempty,lte=64"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate security group update option.
func (opt HuaWeiSecurityGroupUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiSecurityGroupListOption define huawei security group list option.
type HuaWeiSecurityGroupListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.HuaWeiPage `json:"page" validate:"omitempty"`
}

// Validate huawei security group list option.
func (opt HuaWeiSecurityGroupListOption) Validate() error {
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

// -------------------------- Azure --------------------------

// AzureSecurityGroupOption define security group update option.
type AzureSecurityGroupOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Region            string `json:"region" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// Validate azure security group update option.
func (opt AzureSecurityGroupOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureSecurityGroupListOption define azure security group list option.
type AzureSecurityGroupListOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate huawei security group list option.
func (opt AzureSecurityGroupListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}
