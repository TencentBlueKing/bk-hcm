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

// AzureSGRuleCreateReq define azure security group create request.
type AzureSGRuleCreateReq struct {
	Rules []corecloud.AzureSecurityGroupRuleSpec `json:"rules" validate:"required"`
}

// Validate azure security group rule create request.
func (req *AzureSGRuleCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// AzureSGRuleBatchUpdateReq define azure security group batch update request.
type AzureSGRuleBatchUpdateReq struct {
	Rules []AzureSGRuleUpdate `json:"rules" validate:"required"`
}

// AzureSGRuleUpdate azure security group batch update option.
type AzureSGRuleUpdate struct {
	ID   string                                `json:"id" validate:"required"`
	Spec *corecloud.AzureSecurityGroupRuleSpec `json:"spec" validate:"required"`
}

// Validate azure security group rule batch update request.
func (req *AzureSGRuleBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AzureSGRuleListReq azure security group rule list req.
type AzureSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate azure security group rule list request.
func (req *AzureSGRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSGRuleListResult define azure security group rule list result.
type AzureSGRuleListResult struct {
	Count   uint64                             `json:"count,omitempty"`
	Details []corecloud.AzureSecurityGroupRule `json:"details,omitempty"`
}

// AzureSGRuleListResp define azure security group rule list resp.
type AzureSGRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AzureSGRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// AzureSGRuleDeleteReq azure security group rule delete request.
type AzureSGRuleDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate azure security group rule delete request.
func (req *AzureSGRuleDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
