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

package types

import (
	"errors"

	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// TCloudSGRuleCreateOption tcloud security group rule create option.
type TCloudSGRuleCreateOption struct {
	Region               string         `json:"region" validate:"required"`
	CloudSecurityGroupID string         `json:"cloud_security_group_id" validate:"required"`
	EgressRuleSet        []TCloudSGRule `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []TCloudSGRule `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate tcloud security group rule create option.
func (opt TCloudSGRuleCreateOption) Validate() error {
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

// TCloudSGRule tcloud security group rule.
type TCloudSGRule struct {
	Protocol                   *string `json:"protocol"`
	Port                       *string `json:"port"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
	Action                     string  `json:"action"`
	Description                *string `json:"description"`
}

// -------------------------- Delete --------------------------

// TCloudSGRuleDeleteOption tcloud security group delete option.
type TCloudSGRuleDeleteOption struct {
	Region               string  `json:"region" validate:"required"`
	CloudSecurityGroupID string  `json:"cloud_security_group_id" validate:"required"`
	Version              string  `json:"version" validate:"required"`
	EgressRuleIndexes    []int64 `json:"egress_rule_indexes" validate:"omitempty"`
	IngressRuleIndexes   []int64 `json:"ingress_rule_indexes" validate:"omitempty"`
}

// Validate tcloud security group rule delete option.
func (opt TCloudSGRuleDeleteOption) Validate() error {
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

// TCloudSGRuleUpdateOption tcloud security group rule update option.
type TCloudSGRuleUpdateOption struct {
	Region               string                   `json:"region" validate:"required"`
	CloudSecurityGroupID string                   `json:"cloud_security_group_id" validate:"required"`
	Version              string                   `json:"version" validate:"required"`
	EgressRuleSet        []TCloudSGRuleUpdateSpec `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []TCloudSGRuleUpdateSpec `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate tcloud security group rule delete option.
func (opt TCloudSGRuleUpdateOption) Validate() error {
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

// TCloudSGRuleUpdateSpec tcloud security group rule when update.
type TCloudSGRuleUpdateSpec struct {
	PolicyIndex                int64   `json:"policy_index" validate:"required"`
	Protocol                   *string `json:"protocol"`
	Port                       *string `json:"port"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
	Action                     string  `json:"action"`
	Description                *string `json:"memo"`
}

// -------------------------- List --------------------------

// TCloudSGRuleListOption define tcloud security group rule list option.
type TCloudSGRuleListOption struct {
	Region               string `json:"region" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
}

// Validate tcloud security group rule list option.
func (opt TCloudSGRuleListOption) Validate() error {
	return validator.Validate.Struct(opt)
}
