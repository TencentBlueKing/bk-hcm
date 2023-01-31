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
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// HuaWeiSGRuleCreateReq define huawei security group create request.
type HuaWeiSGRuleCreateReq struct {
	Rules []HuaWeiSGRuleBatchCreate `json:"rules" validate:"required"`
}

// HuaWeiSGRuleBatchCreate define huawei security group rule when create.
type HuaWeiSGRuleBatchCreate struct {
	CloudID                   string                       `json:"cloud_id"`
	Memo                      *string                      `json:"memo"`
	Protocol                  string                       `json:"protocol"`
	Ethertype                 string                       `json:"ethertype"`
	CloudRemoteGroupID        string                       `json:"cloud_remote_group_id"`
	RemoteIPPrefix            string                       `json:"remote_ip_prefix"`
	CloudRemoteAddressGroupID string                       `json:"cloud_remote_address_group_id"`
	Port                      string                       `json:"port"`
	Priority                  int64                        `json:"priority"`
	Action                    string                       `json:"action"`
	Type                      enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID      string                       `json:"cloud_security_group_id"`
	CloudProjectID            string                       `json:"cloud_project_id"`
	AccountID                 string                       `json:"account_id"`
	Region                    string                       `json:"region"`
	SecurityGroupID           string                       `json:"security_group_id"`
}

// Validate huawei security group rule create request.
func (req *HuaWeiSGRuleCreateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Update --------------------------

// HuaWeiSGRuleBatchUpdateReq define huawei security group batch update request.
type HuaWeiSGRuleBatchUpdateReq struct {
	Rules []HuaWeiSGRuleBatchUpdate `json:"rules" validate:"required"`
}

// HuaWeiSGRuleBatchUpdate huawei security group batch update option.
type HuaWeiSGRuleBatchUpdate struct {
	ID                        string                       `json:"id" validate:"required"`
	CloudID                   string                       `json:"cloud_id"`
	Memo                      *string                      `json:"memo"`
	Protocol                  string                       `json:"protocol"`
	Ethertype                 string                       `json:"ethertype"`
	CloudRemoteGroupID        string                       `json:"cloud_remote_group_id"`
	RemoteIPPrefix            string                       `json:"remote_ip_prefix"`
	CloudRemoteAddressGroupID string                       `json:"cloud_remote_address_group_id"`
	Port                      string                       `json:"port"`
	Priority                  int64                        `json:"priority"`
	Action                    string                       `json:"action"`
	Type                      enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID      string                       `json:"cloud_security_group_id"`
	CloudProjectID            string                       `json:"cloud_project_id"`
	AccountID                 string                       `json:"account_id"`
	Region                    string                       `json:"region"`
	SecurityGroupID           string                       `json:"security_group_id"`
}

// Validate huawei security group rule batch update request.
func (req *HuaWeiSGRuleBatchUpdateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// HuaWeiSGRuleListReq huawei security group rule list req.
type HuaWeiSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
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

// HuaWeiSGRuleBatchDeleteReq huawei security group rule delete request.
type HuaWeiSGRuleBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate huawei security group rule delete request.
func (req *HuaWeiSGRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
