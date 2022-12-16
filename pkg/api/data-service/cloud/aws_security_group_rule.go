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

package cloud

import (
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// AwsSGRuleCreateReq define aws security group create request.
type AwsSGRuleCreateReq struct {
	Rules []corecloud.AwsSecurityGroupRuleSpec `json:"rules" validate:"required"`
}

// Validate aws security group rule create request.
func (req *AwsSGRuleCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// AwsSGRuleBatchUpdateReq define aws security group batch update request.
type AwsSGRuleBatchUpdateReq struct {
	Rules []AwsSGRuleUpdate `json:"rules" validate:"required"`
}

// AwsSGRuleUpdate aws security group batch update option.
type AwsSGRuleUpdate struct {
	ID   string                              `json:"id" validate:"required"`
	Spec *corecloud.AwsSecurityGroupRuleSpec `json:"spec" validate:"required"`
}

// Validate aws security group rule batch update request.
func (req *AwsSGRuleBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AwsSGRuleListReq aws security group rule list req.
type AwsSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate aws security group rule list request.
func (req *AwsSGRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsSGRuleListResult define aws security group rule list result.
type AwsSGRuleListResult struct {
	Count   uint64                           `json:"count,omitempty"`
	Details []corecloud.AwsSecurityGroupRule `json:"details,omitempty"`
}

// AwsSGRuleListResp define aws security group rule list resp.
type AwsSGRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AwsSGRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// AwsSGRuleDeleteReq aws security group rule delete request.
type AwsSGRuleDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate aws security group rule delete request.
func (req *AwsSGRuleDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
