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

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// AwsSGRuleCreateOption aws security group rule create option.
type AwsSGRuleCreateOption struct {
	Region               string            `json:"region" validate:"required"`
	CloudSecurityGroupID string            `json:"cloud_security_group_id" validate:"required"`
	EgressRuleSet        []AwsSGRuleCreate `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []AwsSGRuleCreate `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate aws security group rule create option.
func (opt AwsSGRuleCreateOption) Validate() error {
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

// AwsSGRuleCreate aws security group rule.
type AwsSGRuleCreate struct {
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Description                *string `json:"description"`
	FromPort                   int64   `json:"from_port"`
	ToPort                     int64   `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}

// -------------------------- Delete --------------------------

// AwsSGRuleDeleteOption aws security group delete option.
type AwsSGRuleDeleteOption struct {
	Region               string   `json:"region" validate:"required"`
	CloudSecurityGroupID string   `json:"cloud_security_group_id" validate:"required"`
	CloudEgressRuleIDs   []string `json:"cloud_egress_rule_ids" validate:"omitempty"`
	CloudIngressRuleIDs  []string `json:"cloud_ingress_rule_ids" validate:"omitempty"`
}

// Validate aws security group rule delete option.
func (opt AwsSGRuleDeleteOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudEgressRuleIDs) == 0 && len(opt.CloudIngressRuleIDs) == 0 {
		return errors.New("egress rule ids or ingress rule ids is required")
	}

	if opt.CloudEgressRuleIDs != nil && opt.CloudIngressRuleIDs != nil {
		return errors.New("egress rule ids or ingress rule ids only one is allowed")
	}

	return nil
}

// -------------------------- List --------------------------

// AwsSGRuleListOption define aws security group rule list option.
type AwsSGRuleListOption struct {
	Region               string        `json:"region" validate:"required"`
	CloudSecurityGroupID string        `json:"cloud_security_group_id" validate:"required"`
	Page                 *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate security group rule list option.
func (opt AwsSGRuleListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Update --------------------------

// AwsSGRuleUpdateOption aws security group rule update option.
type AwsSGRuleUpdateOption struct {
	Region               string            `json:"region" validate:"required"`
	CloudSecurityGroupID string            `json:"cloud_security_group_id" validate:"required"`
	RuleSet              []AwsSGRuleUpdate `json:"rule_set" validate:"required"`
}

// Validate aws security group rule delete option.
func (opt AwsSGRuleUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AwsSGRuleUpdate aws security group rule when update.
type AwsSGRuleUpdate struct {
	CloudID                    string  `json:"cloud_id"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Description                *string `json:"description"`
	FromPort                   int64   `json:"from_port"`
	ToPort                     int64   `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}
