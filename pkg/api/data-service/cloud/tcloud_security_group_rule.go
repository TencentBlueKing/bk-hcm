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

// TCloudSGRuleCreateReq define tcloud security group create request.
type TCloudSGRuleCreateReq struct {
	Rules []corecloud.TCloudSecurityGroupRuleSpec `json:"rules" validate:"required"`
}

// Validate tcloud security group rule create request.
func (req *TCloudSGRuleCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// TCloudSGRuleBatchUpdateReq define tcloud security group batch update request.
type TCloudSGRuleBatchUpdateReq struct {
	Rules []TCloudSGRuleBatchUpdateOption `json:"rules" validate:"required"`
}

// TCloudSGRuleBatchUpdateOption tcloud security group batch update option.
type TCloudSGRuleBatchUpdateOption struct {
	ID   string                                 `json:"id" validate:"required"`
	Spec *corecloud.TCloudSecurityGroupRuleSpec `json:"spec" validate:"required"`
}

// Validate tcloud security group rule batch update request.
func (req *TCloudSGRuleBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// TCloudSGRuleListReq tcloud security group rule list req.
type TCloudSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate tcloud security group rule list request.
func (req *TCloudSGRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudSGRuleListResult define tcloud security group rule list result.
type TCloudSGRuleListResult struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []corecloud.TCloudSecurityGroupRule `json:"details,omitempty"`
}

// TCloudSGRuleListResp define tcloud security group rule list resp.
type TCloudSGRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudSGRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// TCloudSGRuleDeleteReq tcloud security group rule delete request.
type TCloudSGRuleDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate tcloud security group rule delete request.
func (req *TCloudSGRuleDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
