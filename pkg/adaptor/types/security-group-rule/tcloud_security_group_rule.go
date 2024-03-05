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

package securitygrouprule

import (
	"errors"

	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// TCloudCreateOption tcloud security group rule create option.
type TCloudCreateOption struct {
	Region               string   `json:"region" validate:"required"`
	CloudSecurityGroupID string   `json:"cloud_security_group_id" validate:"required"`
	EgressRuleSet        []TCloud `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []TCloud `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate tcloud security group rule create option.
func (opt TCloudCreateOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.EgressRuleSet) == 0 && len(opt.IngressRuleSet) == 0 {
		return errors.New("egress rule or ingress rule is required")
	}

	if len(opt.EgressRuleSet) != 0 && len(opt.IngressRuleSet) != 0 {
		return errors.New("egress rule or ingress rule only one is allowed")
	}

	return nil
}

// TCloud tcloud security group rule.
type TCloud struct {
	Protocol                   *string `json:"protocol"`
	Port                       *string `json:"port"`
	CloudServiceID             *string `json:"cloud_service_id"`
	CloudServiceGroupID        *string `json:"cloud_service_group_id"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	CloudAddressID             *string `json:"cloud_address_id"`
	CloudAddressGroupID        *string `json:"cloud_address_group_id"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
	Action                     string  `json:"action"`
	Description                *string `json:"description"`
}

// -------------------------- Delete --------------------------

// TCloudDeleteOption tcloud security group delete option.
type TCloudDeleteOption struct {
	Region               string  `json:"region" validate:"required"`
	CloudSecurityGroupID string  `json:"cloud_security_group_id" validate:"required"`
	Version              string  `json:"version" validate:"required"`
	EgressRuleIndexes    []int64 `json:"egress_rule_indexes" validate:"omitempty"`
	IngressRuleIndexes   []int64 `json:"ingress_rule_indexes" validate:"omitempty"`
}

// Validate tcloud security group rule delete option.
func (opt TCloudDeleteOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.EgressRuleIndexes) == 0 && len(opt.IngressRuleIndexes) == 0 {
		return errors.New("egress rule index or ingress rule index is required")
	}

	if len(opt.EgressRuleIndexes) != 0 && len(opt.IngressRuleIndexes) != 0 {
		return errors.New("egress rule index or ingress rule index only one is allowed")
	}

	return nil
}

// -------------------------- Update --------------------------

// TCloudUpdateOption tcloud security group rule update option.
type TCloudUpdateOption struct {
	Region               string             `json:"region" validate:"required"`
	CloudSecurityGroupID string             `json:"cloud_security_group_id" validate:"required"`
	Version              string             `json:"version" validate:"required"`
	EgressRuleSet        []TCloudUpdateSpec `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []TCloudUpdateSpec `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate tcloud security group rule delete option.
func (opt TCloudUpdateOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.EgressRuleSet) == 0 && len(opt.IngressRuleSet) == 0 {
		return errors.New("egress rule or ingress rule is required")
	}

	if len(opt.EgressRuleSet) != 0 && len(opt.IngressRuleSet) != 0 {
		return errors.New("egress rule or ingress rule only one is allowed")
	}

	return nil
}

// TCloudUpdateSpec tcloud security group rule when update.
type TCloudUpdateSpec struct {
	CloudPolicyIndex           int64   `json:"cloud_policy_index" validate:"required"`
	Protocol                   *string `json:"protocol"`
	CloudServiceID             *string `json:"cloud_service_id"`
	CloudServiceGroupID        *string `json:"cloud_service_group_id"`
	Port                       *string `json:"port"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	CloudAddressID             *string `json:"cloud_address_id"`
	CloudAddressGroupID        *string `json:"cloud_address_group_id"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
	Action                     string  `json:"action"`
	Description                *string `json:"memo"`
}

// -------------------------- List --------------------------

// TCloudListOption define tcloud security group rule list option.
type TCloudListOption struct {
	Region               string `json:"region" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
}

// Validate tcloud security group rule list option.
func (opt TCloudListOption) Validate() error {
	return validator.Validate.Struct(opt)
}
