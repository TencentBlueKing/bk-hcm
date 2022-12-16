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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// AzureSGRuleCreateReq define azure security group create request.
type AzureSGRuleCreateReq struct {
	AccountID      string            `json:"account_id"`
	EgressRuleSet  []AzureSGRuleSpec `json:"egress_rule_set" validate:"required"`
	IngressRuleSet []AzureSGRuleSpec `json:"ingress_rule_set" validate:"required"`
}

// Validate azure security group rule create request.
func (req *AzureSGRuleCreateReq) Validate() error {
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

// AzureSGRuleSpec define azure sg rule spec when create.
type AzureSGRuleSpec struct {
	Name                             string                       `json:"name"`
	Memo                             *string                      `json:"description"`
	DestinationAddressPrefix         *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes       []*string                    `json:"destination_address_prefixes"`
	CloudDestinationSecurityGroupIDs []*string                    `json:"cloud_destination_security_group_ids"`
	DestinationPortRange             *string                      `json:"destination_port_range"`
	DestinationPortRanges            []*string                    `json:"destination_port_ranges"`
	Protocol                         string                       `json:"protocol"`
	SourceAddressPrefix              *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes            []*string                    `json:"source_address_prefixes"`
	CloudSourceSecurityGroupIDs      []*string                    `json:"cloud_source_security_group_ids"`
	SourcePortRange                  *string                      `json:"source_port_range"`
	SourcePortRanges                 []*string                    `json:"source_port_ranges"`
	Priority                         int32                        `json:"priority"`
	Type                             enumor.SecurityGroupRuleType `json:"type"`
	Access                           string                       `json:"access"`
}

// -------------------------- Update --------------------------

// AzureSGRuleUpdateReq define azure security group update request.
type AzureSGRuleUpdateReq struct {
	Spec *AzureSGRuleSpec `json:"spec" validate:"required"`
}

// Validate azure security group rule update request.
func (req *AzureSGRuleUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
