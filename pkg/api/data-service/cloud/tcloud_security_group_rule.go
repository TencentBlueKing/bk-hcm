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

// TCloudSGRuleCreateReq define tcloud security group create request.
type TCloudSGRuleCreateReq struct {
	Rules []TCloudSGRuleBatchCreate `json:"rules" validate:"required"`
}

// Validate tcloud security group rule create request.
func (req *TCloudSGRuleCreateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d, but got: %d",
			constant.BatchOperationMaxLimit, len(req.Rules))
	}

	return nil
}

// TCloudSGRuleBatchCreate define tcloud security group rule when create.
type TCloudSGRuleBatchCreate struct {
	CloudPolicyIndex           int64                        `json:"cloud_policy_index"`
	Version                    string                       `json:"version"`
	Protocol                   *string                      `json:"protocol"`
	Port                       *string                      `json:"port"`
	ServiceID                  *string                      `json:"service_id"`
	CloudServiceID             *string                      `json:"cloud_service_id"`
	ServiceGroupID             *string                      `json:"service_group_id"`
	CloudServiceGroupID        *string                      `json:"cloud_service_group_id"`
	IPv4Cidr                   *string                      `json:"ipv4_cidr"`
	IPv6Cidr                   *string                      `json:"ipv6_cidr"`
	CloudTargetSecurityGroupID *string                      `json:"cloud_target_security_group_id"`
	AddressID                  *string                      `json:"address_id"`
	CloudAddressID             *string                      `json:"cloud_address_id"`
	AddressGroupID             *string                      `json:"address_group_id"`
	CloudAddressGroupID        *string                      `json:"cloud_address_group_id"`
	Action                     string                       `json:"action"`
	Memo                       *string                      `json:"memo"`
	Type                       enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID       string                       `json:"cloud_security_group_id"`
	SecurityGroupID            string                       `json:"security_group_id"`
	Region                     string                       `json:"region"`
	AccountID                  string                       `json:"account_id"`
}

// -------------------------- Update --------------------------

// TCloudSGRuleBatchUpdateReq define tcloud security group batch update request.
type TCloudSGRuleBatchUpdateReq struct {
	Rules []TCloudSGRuleBatchUpdate `json:"rules" validate:"required"`
}

// TCloudSGRuleBatchUpdate tcloud security group batch update option.
type TCloudSGRuleBatchUpdate struct {
	ID                         string                       `json:"id" validate:"required"`
	CloudPolicyIndex           int64                        `json:"cloud_policy_index"`
	Version                    string                       `json:"version"`
	Protocol                   *string                      `json:"protocol"`
	Port                       *string                      `json:"port"`
	ServiceID                  *string                      `json:"service_id"`
	CloudServiceID             *string                      `json:"cloud_service_id"`
	ServiceGroupID             *string                      `json:"service_group_id"`
	CloudServiceGroupID        *string                      `json:"cloud_service_group_id"`
	IPv4Cidr                   *string                      `json:"ipv4_cidr"`
	IPv6Cidr                   *string                      `json:"ipv6_cidr"`
	CloudTargetSecurityGroupID *string                      `json:"cloud_target_security_group_id"`
	AddressID                  *string                      `json:"address_id"`
	CloudAddressID             *string                      `json:"cloud_address_id"`
	AddressGroupID             *string                      `json:"address_group_id"`
	CloudAddressGroupID        *string                      `json:"cloud_address_group_id"`
	Action                     string                       `json:"action"`
	Memo                       *string                      `json:"memo"`
	Type                       enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID       string                       `json:"cloud_security_group_id"`
	SecurityGroupID            string                       `json:"security_group_id"`
	Region                     string                       `json:"region"`
	AccountID                  string                       `json:"account_id"`
}

// Validate tcloud security group rule batch update request.
func (req *TCloudSGRuleBatchUpdateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// TCloudSGRuleListReq tcloud security group rule list req.
type TCloudSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
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

// TCloudSGRuleListExtResult define tcloud security group rule list ext result.
type TCloudSGRuleListExtResult struct {
	Count             uint64                              `json:"count,omitempty"`
	SecurityGroup     []corecloud.BaseSecurityGroup       `json:"security_group,omitempty"`
	SecurityGroupRule []corecloud.TCloudSecurityGroupRule `json:"security_group_rule,omitempty"`
}

// TCloudSGRuleListExtResp define tcloud security group rule list ext resp.
type TCloudSGRuleListExtResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudSGRuleListExtResult `json:"data"`
}

// -------------------------- Delete --------------------------

// TCloudSGRuleBatchDeleteReq tcloud security group rule delete request.
type TCloudSGRuleBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate tcloud security group rule delete request.
func (req *TCloudSGRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
