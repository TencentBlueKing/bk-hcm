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

import (
	"errors"
	"strconv"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// TCloudListPoliciesOption defines options for listing CAM policies.
// reference: https://cloud.tencent.com/document/product/598/34570
type TCloudListPoliciesOption struct {
	Region string `json:"region" validate:"omitempty"`
	Page   uint64 `json:"page" validate:"required,min=1,max=200"`
	Rp     uint64 `json:"rp" validate:"required,min=1,max=200"`
	// All获取所有策略，QCS只获取预设策略，Local只获取自定义策略，默认ALL
	Scope string `json:"scope" validate:"omitempty"`
}

// Validate TCloudListPoliciesOption.
func (opt *TCloudListPoliciesOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudPolicyItem defines a single policy item returned by GetPolicyList.
type TCloudPolicyItem struct {
	PolicyID    uint64                  `json:"policy_id"`
	PolicyName  string                  `json:"policy_name"`
	Description string                  `json:"description"`
	PolicyType  enumor.TCloudPolicyType `json:"policy_type"`
	CreateTime  string                  `json:"create_time"`
}

// TCloudGetPolicyDetailOption defines options for getting a single CAM policy detail.
// reference: https://cloud.tencent.com/document/product/598/34570
type TCloudGetPolicyDetailOption struct {
	PolicyID uint64 `json:"policy_id" validate:"required"`
	Region   string `json:"region" validate:"omitempty"`
}

// Validate TCloudGetPolicyDetailOption.
func (opt *TCloudGetPolicyDetailOption) Validate() error {
	if opt.PolicyID == 0 {
		return errors.New("policy_id is required")
	}
	return validator.Validate.Struct(opt)
}

// TCloudPolicyDetail defines the full detail of a CAM policy including PolicyDocument.
type TCloudPolicyDetail struct {
	PolicyID       uint64                  `json:"policy_id"`
	PolicyName     string                  `json:"policy_name"`
	PolicyDocument string                  `json:"policy_document"`
	Description    string                  `json:"description"`
	PolicyType     enumor.TCloudPolicyType `json:"policy_type"`
	CreateTime     string                  `json:"create_time"`
}

// GetCloudID implements CloudResType interface.
func (t TCloudPolicyDetail) GetCloudID() string {
	return strconv.FormatUint(t.PolicyID, 10)
}
