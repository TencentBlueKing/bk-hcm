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

// HuaWeiSGRuleCreateReq define huawei security group create request.
type HuaWeiSGRuleCreateReq struct {
	Rules []corecloud.HuaWeiSecurityGroupRuleSpec `json:"rules" validate:"required"`
}

// Validate huawei security group rule create request.
func (req *HuaWeiSGRuleCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// HuaWeiSGRuleBatchUpdateReq define huawei security group batch update request.
type HuaWeiSGRuleBatchUpdateReq struct {
	Rules []HuaWeiSGRuleBatchUpdateOption `json:"rules" validate:"required"`
}

// HuaWeiSGRuleBatchUpdateOption huawei security group batch update option.
type HuaWeiSGRuleBatchUpdateOption struct {
	ID   string                                 `json:"id" validate:"required"`
	Spec *corecloud.HuaWeiSecurityGroupRuleSpec `json:"spec" validate:"required"`
}

// Validate huawei security group rule batch update request.
func (req *HuaWeiSGRuleBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// HuaWeiSGRuleListReq huawei security group rule list req.
type HuaWeiSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate huawei security group rule list request.
func (req *HuaWeiSGRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiSGRuleListResult define huawei security group rule list result.
type HuaWeiSGRuleListResult struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []corecloud.HuaWeiSecurityGroupRule `json:"details,omitempty"`
}

// HuaWeiSGRuleListResp define huawei security group rule list resp.
type HuaWeiSGRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *HuaWeiSGRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// HuaWeiSGRuleDeleteReq huawei security group rule delete request.
type HuaWeiSGRuleDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate huawei security group rule delete request.
func (req *HuaWeiSGRuleDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
