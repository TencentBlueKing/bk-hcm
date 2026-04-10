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

package account

import "hcm/pkg/criteria/validator"

// TCloudCreatePolicyOption defines options for creating a CAM policy.
type TCloudCreatePolicyOption struct {
	Region         string `json:"region"`
	PolicyName     string `json:"policy_name" validate:"required"`
	PolicyDocument string `json:"policy_document" validate:"required"`
	Description    string `json:"description"`
}

// Validate TCloudCreatePolicyOption.
func (opt *TCloudCreatePolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudCreatePolicyResult defines the result of creating a CAM policy.
type TCloudCreatePolicyResult struct {
	PolicyID uint64 `json:"policy_id"`
}

// TCloudUpdatePolicyOption defines options for updating a CAM policy.
type TCloudUpdatePolicyOption struct {
	Region         string  `json:"region"`
	PolicyID       uint64  `json:"policy_id" validate:"required"`
	PolicyDocument *string `json:"policy_document"`
	Description    *string `json:"description"`
}

// Validate TCloudUpdatePolicyOption.
func (opt *TCloudUpdatePolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}
