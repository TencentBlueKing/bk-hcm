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

	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// AzureSGRuleCreateReq define azure security group create request.
type AzureSGRuleCreateReq struct {
	Rules []AzureSGRuleBatchCreate `json:"rules" validate:"required"`
}

// AzureSGRuleBatchCreate define azure security group rule when create.
type AzureSGRuleBatchCreate struct {
	CloudID                          string                       `json:"cloud_id"`
	Etag                             *string                      `json:"etag"`
	Name                             string                       `json:"name"`
	Memo                             *string                      `json:"memo"`
	DestinationAddressPrefix         *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes       []*string                    `json:"destination_address_prefixes"`
	CloudDestinationSecurityGroupIDs []*string                    `json:"cloud_destination_security_group_ids"`
	DestinationPortRange             *string                      `json:"destination_port_range"`
	DestinationPortRanges            []*string                    `json:"destination_port_ranges"`
	Protocol                         string                       `json:"protocol"`
	ProvisioningState                string                       `json:"provisioning_state"`
	SourceAddressPrefix              *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes            []*string                    `json:"source_address_prefixes"`
	CloudSourceSecurityGroupIDs      []*string                    `json:"cloud_source_security_group_ids"`
	SourcePortRange                  *string                      `json:"source_port_range"`
	SourcePortRanges                 []*string                    `json:"source_port_ranges"`
	Priority                         int32                        `json:"priority"`
	Type                             enumor.SecurityGroupRuleType `json:"type"`
	Access                           string                       `json:"access"`
	CloudSecurityGroupID             string                       `json:"cloud_security_group_id"`
	AccountID                        string                       `json:"account_id"`
	Region                           string                       `json:"region"`
	SecurityGroupID                  string                       `json:"security_group_id"`
}

// Validate azure security group rule create request.
func (req *AzureSGRuleCreateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Update --------------------------

// AzureSGRuleBatchUpdateReq define azure security group batch update request.
type AzureSGRuleBatchUpdateReq struct {
	Rules []AzureSGRuleUpdate `json:"rules" validate:"required"`
}

// AzureSGRuleUpdate azure security group batch update option.
type AzureSGRuleUpdate struct {
	ID                               string                       `json:"id" validate:"required"`
	CloudID                          string                       `json:"cloud_id"`
	Etag                             *string                      `json:"etag"`
	Name                             string                       `json:"name"`
	Memo                             *string                      `json:"memo"`
	DestinationAddressPrefix         *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes       []*string                    `json:"destination_address_prefixes"`
	CloudDestinationSecurityGroupIDs []*string                    `json:"cloud_destination_security_group_ids"`
	DestinationPortRange             *string                      `json:"destination_port_range"`
	DestinationPortRanges            []*string                    `json:"destination_port_ranges"`
	Protocol                         string                       `json:"protocol"`
	ProvisioningState                string                       `json:"provisioning_state"`
	SourceAddressPrefix              *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes            []*string                    `json:"source_address_prefixes"`
	CloudSourceSecurityGroupIDs      []*string                    `json:"cloud_source_security_group_ids"`
	SourcePortRange                  *string                      `json:"source_port_range"`
	SourcePortRanges                 []*string                    `json:"source_port_ranges"`
	Priority                         int32                        `json:"priority"`
	Type                             enumor.SecurityGroupRuleType `json:"type"`
	Access                           string                       `json:"access"`
	CloudSecurityGroupID             string                       `json:"cloud_security_group_id"`
	AccountID                        string                       `json:"account_id"`
	Region                           string                       `json:"region"`
	SecurityGroupID                  string                       `json:"security_group_id"`
}

// Validate azure security group rule batch update request.
func (req *AzureSGRuleBatchUpdateReq) Validate() error {
	if len(req.Rules) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Rules) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// AzureSGRuleListReq azure security group rule list req.
type AzureSGRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate azure security group rule list request.
func (req *AzureSGRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSGRuleListResult define azure security group rule list result.
type AzureSGRuleListResult struct {
	Count   uint64                             `json:"count,omitempty"`
	Details []corecloud.AzureSecurityGroupRule `json:"details,omitempty"`
}

// AzureSGRuleListResp define azure security group rule list resp.
type AzureSGRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AzureSGRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// AzureSGRuleBatchDeleteReq azure security group rule delete request.
type AzureSGRuleBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate azure security group rule delete request.
func (req *AzureSGRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
