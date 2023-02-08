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

package azure

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateSecurityGroupRule create security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) CreateSecurityGroupRule(kt *kit.Kit, opt *types.AzureSGRuleCreateOption) ([]*armnetwork.SecurityRule,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := az.getSecurityGroupByCloudID(kt, opt.ResourceGroupName, opt.CloudSecurityGroupID)
	if err != nil {
		return nil, err
	}

	var rules []*armnetwork.SecurityRule
	var nameMap map[string]bool
	if len(opt.EgressRuleSet) != 0 {
		rules, nameMap = convSecurityGroupRule(armnetwork.SecurityRuleDirectionOutbound, opt.EgressRuleSet)
	}

	if len(opt.IngressRuleSet) != 0 {
		rules, nameMap = convSecurityGroupRule(armnetwork.SecurityRuleDirectionInbound, opt.IngressRuleSet)
	}
	rules = append(sg.Properties.SecurityRules, rules...)

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	req := armnetwork.SecurityGroup{
		Location: &opt.Region,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: rules,
		},
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, *sg.Name, req, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]*armnetwork.SecurityRule, 0)
	for _, rule := range resp.SecurityGroup.Properties.SecurityRules {
		if nameMap[*rule.Name] {
			result = append(result, rule)
		}
	}

	return result, nil
}

func convSecurityGroupRule(direction armnetwork.SecurityRuleDirection, rules []types.AzureSGRuleCreate,
) ([]*armnetwork.SecurityRule, map[string]bool) {

	result := make([]*armnetwork.SecurityRule, 0, len(rules))
	nameMap := make(map[string]bool, len(rules))
	for _, one := range rules {
		nameMap[one.Name] = true

		rule := &armnetwork.SecurityRule{
			Name: &one.Name,
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:                one.Description,
				DestinationAddressPrefix:   one.DestinationAddressPrefix,
				DestinationAddressPrefixes: one.DestinationAddressPrefixes,
				DestinationPortRange:       one.DestinationPortRange,
				DestinationPortRanges:      one.DestinationPortRanges,
				Priority:                   &one.Priority,
				SourceAddressPrefix:        one.SourceAddressPrefix,
				SourceAddressPrefixes:      one.SourceAddressPrefixes,
				SourcePortRange:            one.SourcePortRange,
				SourcePortRanges:           one.SourcePortRanges,
			},
		}
		access := armnetwork.SecurityRuleAccess(one.Access)
		rule.Properties.Access = &access
		rule.Properties.Direction = &direction
		protocol := armnetwork.SecurityRuleProtocol(one.Protocol)
		rule.Properties.Protocol = &protocol

		if len(one.CloudDestinationSecurityGroupIDs) != 0 {
			rule.Properties.DestinationApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
				len(one.CloudDestinationSecurityGroupIDs))

			for _, id := range one.CloudDestinationSecurityGroupIDs {
				rule.Properties.DestinationApplicationSecurityGroups = append(rule.Properties.
					DestinationApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
					ID: id,
				})
			}
		}

		if len(one.CloudSourceSecurityGroupIDs) != 0 {
			rule.Properties.SourceApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
				len(one.CloudSourceSecurityGroupIDs))

			for _, id := range one.CloudSourceSecurityGroupIDs {
				rule.Properties.SourceApplicationSecurityGroups = append(rule.Properties.
					SourceApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
					ID: id,
				})
			}
		}

		result = append(result, rule)
	}

	return result, nameMap
}

// UpdateSecurityGroupRule update security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) UpdateSecurityGroupRule(kt *kit.Kit, opt *types.AzureSGRuleUpdateOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := az.getSecurityGroupByCloudID(kt, opt.ResourceGroupName, opt.CloudSecurityGroupID)
	if err != nil {
		return err
	}

	exist := false
	for _, rule := range sg.Properties.SecurityRules {
		if *rule.ID == opt.Rule.CloudID {
			exist = true
			access := armnetwork.SecurityRuleAccess(opt.Rule.Access)
			rule.Properties.Access = &access
			protocol := armnetwork.SecurityRuleProtocol(opt.Rule.Protocol)
			rule.Properties.Protocol = &protocol
			rule.Properties.Description = opt.Rule.Description
			rule.Name = &opt.Rule.Name
			rule.Properties.DestinationAddressPrefix = opt.Rule.DestinationAddressPrefix
			rule.Properties.DestinationAddressPrefixes = opt.Rule.DestinationAddressPrefixes
			rule.Properties.DestinationPortRange = opt.Rule.DestinationPortRange
			rule.Properties.DestinationPortRanges = opt.Rule.DestinationPortRanges
			rule.Properties.Priority = &opt.Rule.Priority
			rule.Properties.SourceAddressPrefix = opt.Rule.SourceAddressPrefix
			rule.Properties.SourceAddressPrefixes = opt.Rule.SourceAddressPrefixes
			rule.Properties.SourcePortRange = opt.Rule.SourcePortRange
			rule.Properties.SourcePortRanges = opt.Rule.SourcePortRanges

			if len(opt.Rule.CloudDestinationSecurityGroupIDs) != 0 {
				rule.Properties.DestinationApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
					len(opt.Rule.CloudDestinationSecurityGroupIDs))

				for _, id := range opt.Rule.CloudDestinationSecurityGroupIDs {
					rule.Properties.DestinationApplicationSecurityGroups = append(rule.Properties.
						DestinationApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
						ID: id,
					})
				}
			}

			if len(opt.Rule.CloudSourceSecurityGroupIDs) != 0 {
				rule.Properties.SourceApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
					len(opt.Rule.CloudSourceSecurityGroupIDs))

				for _, id := range opt.Rule.CloudSourceSecurityGroupIDs {
					rule.Properties.SourceApplicationSecurityGroups = append(rule.Properties.
						SourceApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
						ID: id,
					})
				}
			}
		}
	}

	if !exist {
		return errf.Newf(errf.RecordNotFound, "security group rule: %s not found", opt.Rule.CloudID)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	req := armnetwork.SecurityGroup{
		Location: &opt.Region,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: sg.Properties.SecurityRules,
		},
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, *sg.Name, req, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSecurityGroupRule delete security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) DeleteSecurityGroupRule(kt *kit.Kit, opt *types.AzureSGRuleDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := az.getSecurityGroupByCloudID(kt, opt.ResourceGroupName, opt.CloudSecurityGroupID)
	if err != nil {
		return err
	}

	exist := false
	rules := make([]*armnetwork.SecurityRule, 0)
	for _, rule := range sg.Properties.SecurityRules {
		if *rule.ID == opt.CloudRuleID {
			exist = true
			continue
		}

		rules = append(rules, rule)
	}

	if !exist {
		return errf.Newf(errf.RecordNotFound, "security group rule: %s not found", opt.CloudRuleID)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	req := armnetwork.SecurityGroup{
		Location: &opt.Region,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: rules,
		},
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, *sg.Name, req, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroupRule list security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListSecurityGroupRule(kt *kit.Kit, opt *types.AzureSGRuleListOption) ([]*armnetwork.SecurityRule,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := az.getSecurityGroupByCloudID(kt, opt.ResourceGroupName, opt.CloudSecurityGroupID)
	if err != nil {
		return nil, err
	}

	return sg.Properties.SecurityRules, nil
}
