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

package hspermissiontemplate

import (
	"errors"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// CreateCAMPolicyReq defines the request to create a CAM policy via hc-service.
type CreateCAMPolicyReq struct {
	AccountID      string `json:"account_id" validate:"required"`
	PolicyName     string `json:"policy_name" validate:"required"`
	PolicyDocument string `json:"policy_document" validate:"required"`
	Description    string `json:"description"`
}

// Validate CreateCAMPolicyReq.
func (req *CreateCAMPolicyReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateCAMPolicyResult defines the result of creating a CAM policy.
type CreateCAMPolicyResult struct {
	PolicyID uint64 `json:"policy_id"`
}

// CreateCAMPolicyResp defines the response for creating a CAM policy.
type CreateCAMPolicyResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateCAMPolicyResult `json:"data"`
}

// UpdateCAMPolicyReq defines the request to update a CAM policy via hc-service.
type UpdateCAMPolicyReq struct {
	AccountID      string  `json:"account_id" validate:"required"`
	PolicyID       uint64  `json:"policy_id" validate:"required"`
	PolicyDocument *string `json:"policy_document"`
	Description    *string `json:"description"`
}

// Validate UpdateCAMPolicyReq.
func (req *UpdateCAMPolicyReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.PolicyDocument == nil && req.Description == nil {
		return errors.New("at least one of policy_document or description must be provided")
	}

	return nil
}
