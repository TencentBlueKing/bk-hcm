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

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// -------------------------- Create --------------------------

// AwsCreateOption aws security group rule create option.
type AwsCreateOption struct {
	Region               string      `json:"region" validate:"required"`
	CloudSecurityGroupID string      `json:"cloud_security_group_id" validate:"required"`
	EgressRuleSet        []AwsCreate `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []AwsCreate `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate aws security group rule create option.
func (opt AwsCreateOption) Validate() error {
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

// AwsCreate aws security group rule.
type AwsCreate struct {
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Description                *string `json:"description"`
	FromPort                   *int64  `json:"from_port"`
	ToPort                     *int64  `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}

// -------------------------- Delete --------------------------

// AwsDeleteOption aws security group delete option.
type AwsDeleteOption struct {
	Region               string   `json:"region" validate:"required"`
	CloudSecurityGroupID string   `json:"cloud_security_group_id" validate:"required"`
	CloudEgressRuleIDs   []string `json:"cloud_egress_rule_ids" validate:"omitempty"`
	CloudIngressRuleIDs  []string `json:"cloud_ingress_rule_ids" validate:"omitempty"`
}

// Validate aws security group rule delete option.
func (opt AwsDeleteOption) Validate() error {
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

// AwsListOption define aws security group rule list option.
type AwsListOption struct {
	Region               string        `json:"region" validate:"required"`
	CloudSecurityGroupID string        `json:"cloud_security_group_id" validate:"required"`
	Page                 *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate security group rule list option.
func (opt AwsListOption) Validate() error {
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

// AwsUpdateOption aws security group rule update option.
type AwsUpdateOption struct {
	Region               string            `json:"region" validate:"required"`
	CloudSecurityGroupID string            `json:"cloud_security_group_id" validate:"required"`
	RuleSet              []AwsSGRuleUpdate `json:"rule_set" validate:"required"`
}

// Validate aws security group rule delete option.
func (opt AwsUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AwsSGRuleUpdate aws security group rule when update.
type AwsSGRuleUpdate struct {
	CloudID                    string  `json:"cloud_id"`
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Description                *string `json:"description"`
	FromPort                   *int64  `json:"from_port"`
	ToPort                     *int64  `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}

// AwsSGRule for ec2 SecurityGroupRule
type AwsSGRule struct {
	*ec2.SecurityGroupRule
}

// GetCloudID ...
func (sgrule AwsSGRule) GetCloudID() string {
	return converter.PtrToVal(sgrule.SecurityGroupRuleId)
}
