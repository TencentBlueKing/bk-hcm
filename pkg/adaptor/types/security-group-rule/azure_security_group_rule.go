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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// -------------------------- Create --------------------------

// AzureCreateOption azure security group rule create option.
type AzureCreateOption struct {
	Region               string        `json:"region" validate:"required"`
	ResourceGroupName    string        `json:"resource_group_name" validate:"required"`
	CloudSecurityGroupID string        `json:"cloud_security_group_id" validate:"required"`
	EgressRuleSet        []AzureCreate `json:"egress_rule_set" validate:"omitempty"`
	IngressRuleSet       []AzureCreate `json:"ingress_rule_set" validate:"omitempty"`
}

// Validate azure security group rule create option.
func (opt AzureCreateOption) Validate() error {
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

// AzureCreate azure security group rule.
type AzureCreate struct {
	Name                                string                       `json:"name"`
	Description                         *string                      `json:"description"`
	DestinationAddressPrefix            *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes          []*string                    `json:"destination_address_prefixes"`
	CloudDestinationAppSecurityGroupIDs []*string                    `json:"cloud_destination_app_security_group_ids"`
	DestinationPortRange                *string                      `json:"destination_port_range"`
	DestinationPortRanges               []*string                    `json:"destination_port_ranges"`
	Protocol                            string                       `json:"protocol"`
	SourceAddressPrefix                 *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes               []*string                    `json:"source_address_prefixes"`
	CloudSourceAppSecurityGroupIDs      []*string                    `json:"cloud_source_app_security_group_ids"`
	SourcePortRange                     *string                      `json:"source_port_range"`
	SourcePortRanges                    []*string                    `json:"source_port_ranges"`
	Priority                            int32                        `json:"priority"`
	Type                                enumor.SecurityGroupRuleType `json:"type"`
	Access                              string                       `json:"access"`
}

// -------------------------- Update --------------------------

// AzureUpdateOption azure security group rule update option.
type AzureUpdateOption struct {
	Region               string       `json:"region" validate:"required"`
	ResourceGroupName    string       `json:"resource_group_name" validate:"required"`
	CloudSecurityGroupID string       `json:"cloud_security_group_id" validate:"required"`
	Rule                 *AzureUpdate `json:"rule_set" validate:"required"`
}

// Validate azure security group rule delete option.
func (opt AzureUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureUpdate azure security group rule when update.
type AzureUpdate struct {
	CloudID                             string    `json:"cloud_id"`
	Name                                string    `json:"name"`
	Description                         *string   `json:"description"`
	DestinationAddressPrefix            *string   `json:"destination_address_prefix"`
	DestinationAddressPrefixes          []*string `json:"destination_address_prefixes"`
	CloudDestinationAppSecurityGroupIDs []*string `json:"cloud_destination_app_security_group_ids"`
	DestinationPortRange                *string   `json:"destination_port_range"`
	DestinationPortRanges               []*string `json:"destination_port_ranges"`
	Protocol                            string    `json:"protocol"`
	SourceAddressPrefix                 *string   `json:"source_address_prefix"`
	SourceAddressPrefixes               []*string `json:"source_address_prefixes"`
	CloudSourceAppSecurityGroupIDs      []*string `json:"cloud_source_app_security_group_ids"`
	SourcePortRange                     *string   `json:"source_port_range"`
	SourcePortRanges                    []*string `json:"source_port_ranges"`
	Priority                            int32     `json:"priority"`
	Access                              string    `json:"access"`
}

// -------------------------- Delete --------------------------

// AzureDeleteOption azure security group delete option.
type AzureDeleteOption struct {
	Region               string `json:"region" validate:"required"`
	ResourceGroupName    string `json:"resource_group_name" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
	CloudRuleID          string `json:"cloud_rule_id" validate:"required"`
}

// Validate azure security group rule delete option.
func (opt AzureDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// AzureListOption azure security group list option.
type AzureListOption struct {
	ResourceGroupName    string `json:"resource_group_name" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
}

// Validate azure security group rule list option.
func (opt AzureListOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureSGRule define azure security rule.
type AzureSGRule struct {
	ID                                   *string                                `json:"id"`
	Etag                                 *string                                `json:"etag"`
	Name                                 *string                                `json:"name"`
	Description                          *string                                `json:"description"`
	DestinationAddressPrefix             *string                                `json:"destination_address_prefix"`
	DestinationAddressPrefixes           []*string                              `json:"destination_address_prefixes"`
	DestinationPortRange                 *string                                `json:"destination_port_range"`
	DestinationPortRanges                []*string                              `json:"destination_port_ranges"`
	Protocol                             *armnetwork.SecurityRuleProtocol       `json:"protocol"`
	ProvisioningState                    *armnetwork.ProvisioningState          `json:"provisioning_state"`
	SourceAddressPrefix                  *string                                `json:"source_address_prefix"`
	SourceAddressPrefixes                []*string                              `json:"source_address_prefixes"`
	SourcePortRange                      *string                                `json:"source_port_range"`
	SourcePortRanges                     []*string                              `json:"source_port_ranges"`
	Priority                             *int32                                 `json:"priority"`
	Access                               *armnetwork.SecurityRuleAccess         `json:"access"`
	Direction                            *armnetwork.SecurityRuleDirection      `json:"direction"`
	DestinationApplicationSecurityGroups []*armnetwork.ApplicationSecurityGroup `json:"destination_application_security_groups"`
	SourceApplicationSecurityGroups      []*armnetwork.ApplicationSecurityGroup `json:"source_application_security_groups"`
}

// GetCloudID ...
func (sgrule AzureSGRule) GetCloudID() string {
	return converter.PtrToVal(sgrule.ID)
}
