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

import "hcm/pkg/criteria/validator"

// -------------------------- Update --------------------------

// AzureOption define security group update option.
type AzureOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Region            string `json:"region" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// Validate azure security group update option.
func (opt AzureOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// AzureListOption define azure security group list option.
type AzureListOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate huawei security group list option.
func (opt AzureListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// -------------------------- Delete --------------------------

// AzureDeleteOption azure security group delete option.
type AzureDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate security group delete option.
func (opt AzureDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}
