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

package hcservice

import (
	"errors"

	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// AwsSGRuleCreateReq define aws security group create request.
type AwsSGRuleCreateReq struct {
	AccountID      string            `json:"account_id" validate:"required"`
	EgressRuleSet  []AwsSGRuleCreate `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet []AwsSGRuleCreate `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate aws security group rule create request.
func (req *AwsSGRuleCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.EgressRuleSet) == 0 && len(req.IngressRuleSet) == 0 {
		return errors.New("egress rule or ingress rule is required")
	}

	if len(req.EgressRuleSet) != 0 && len(req.IngressRuleSet) != 0 {
		return errors.New("egress rule or ingress rule only one is allowed")
	}

	return nil
}

// AwsSGRuleCreate define aws sg rule spec when create.
type AwsSGRuleCreate struct {
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Memo                       *string `json:"memo"`
	FromPort                   *int64  `json:"from_port"`
	ToPort                     *int64  `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}

// -------------------------- Update --------------------------

// AwsSGRuleUpdateReq define aws security group update request.
type AwsSGRuleUpdateReq struct {
	IPv4Cidr                   *string `json:"ipv4_cidr"`
	IPv6Cidr                   *string `json:"ipv6_cidr"`
	Memo                       *string `json:"memo"`
	FromPort                   *int64  `json:"from_port"`
	ToPort                     *int64  `json:"to_port"`
	Protocol                   *string `json:"protocol"`
	CloudTargetSecurityGroupID *string `json:"cloud_target_security_group_id"`
}

// Validate aws security group rule update request.
func (req *AwsSGRuleUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
