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

	securitygroup "hcm/pkg/adaptor/types/security-group"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateSecurityGroupRule create security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) CreateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AzureCreateOption) (
	[]*securitygrouprule.AzureSGRule, error) {

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
	rules = append(sg.SecurityRules, rules...)

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

	result := make([]*securitygrouprule.AzureSGRule, 0)
	for _, v := range resp.SecurityGroup.Properties.SecurityRules {
		if nameMap[converter.PtrToVal(v.Name)] {
			result = append(result, az.converCloudToSecurityRule(v))
		}
	}

	return result, nil
}

func convSecurityGroupRule(direction armnetwork.SecurityRuleDirection, rules []securitygrouprule.AzureCreate,
) ([]*armnetwork.SecurityRule, map[string]bool) {

	result := make([]*armnetwork.SecurityRule, 0, len(rules))
	nameMap := make(map[string]bool, len(rules))
	for _, one := range rules {
		nameMap[one.Name] = true

		rule := &armnetwork.SecurityRule{
			Name: converter.ValToPtr(one.Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:                one.Description,
				DestinationAddressPrefix:   one.DestinationAddressPrefix,
				DestinationAddressPrefixes: one.DestinationAddressPrefixes,
				DestinationPortRange:       one.DestinationPortRange,
				DestinationPortRanges:      one.DestinationPortRanges,
				Priority:                   converter.ValToPtr(one.Priority),
				SourceAddressPrefix:        one.SourceAddressPrefix,
				SourceAddressPrefixes:      one.SourceAddressPrefixes,
				SourcePortRange:            one.SourcePortRange,
				SourcePortRanges:           one.SourcePortRanges,
			},
		}
		access := armnetwork.SecurityRuleAccess(one.Access)
		rule.Properties.Access = converter.ValToPtr(access)
		rule.Properties.Direction = converter.ValToPtr(direction)
		protocol := armnetwork.SecurityRuleProtocol(one.Protocol)
		rule.Properties.Protocol = converter.ValToPtr(protocol)

		if len(one.CloudDestinationAppSecurityGroupIDs) != 0 {
			rule.Properties.DestinationApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
				len(one.CloudDestinationAppSecurityGroupIDs))

			for _, id := range one.CloudDestinationAppSecurityGroupIDs {
				rule.Properties.DestinationApplicationSecurityGroups = append(rule.Properties.
					DestinationApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
					ID: SPtrToLowerSPtr(id),
				})
			}
		}

		if len(one.CloudSourceAppSecurityGroupIDs) != 0 {
			rule.Properties.SourceApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
				len(one.CloudSourceAppSecurityGroupIDs))

			for _, id := range one.CloudSourceAppSecurityGroupIDs {
				rule.Properties.SourceApplicationSecurityGroups = append(rule.Properties.
					SourceApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
					ID: SPtrToLowerSPtr(id),
				})
			}
		}

		result = append(result, rule)
	}

	return result, nameMap
}

// UpdateSecurityGroupRule update security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) UpdateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AzureUpdateOption) error {

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

	if err = modifySgRule(sg, opt); err != nil {
		return err
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	req := armnetwork.SecurityGroup{
		Location: &opt.Region,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: sg.SecurityRules,
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

func modifySgRule(sg *securitygroup.AzureSecurityGroup, opt *securitygrouprule.AzureUpdateOption) error {
	exist := false
	for _, rule := range sg.SecurityRules {
		if SPtrToLowerNoSpaceStr(rule.ID) == opt.Rule.CloudID {
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

			if len(opt.Rule.CloudDestinationAppSecurityGroupIDs) != 0 {
				rule.Properties.DestinationApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
					len(opt.Rule.CloudDestinationAppSecurityGroupIDs))

				for _, id := range opt.Rule.CloudDestinationAppSecurityGroupIDs {
					rule.Properties.DestinationApplicationSecurityGroups = append(rule.Properties.
						DestinationApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
						ID: SPtrToLowerSPtr(id),
					})
				}
			}

			if len(opt.Rule.CloudSourceAppSecurityGroupIDs) != 0 {
				rule.Properties.SourceApplicationSecurityGroups = make([]*armnetwork.ApplicationSecurityGroup, 0,
					len(opt.Rule.CloudSourceAppSecurityGroupIDs))

				for _, id := range opt.Rule.CloudSourceAppSecurityGroupIDs {
					rule.Properties.SourceApplicationSecurityGroups = append(rule.Properties.
						SourceApplicationSecurityGroups, &armnetwork.ApplicationSecurityGroup{
						ID: SPtrToLowerSPtr(id),
					})
				}
			}
		}
	}

	if !exist {
		return errf.Newf(errf.RecordNotFound, "security group rule: %s not found", opt.Rule.CloudID)
	}

	return nil
}

// DeleteSecurityGroupRule delete security group rule.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) DeleteSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AzureDeleteOption) error {

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
	for _, rule := range sg.SecurityRules {
		if SPtrToLowerNoSpaceStr(rule.ID) == opt.CloudRuleID {
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
func (az *Azure) ListSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AzureListOption) ([]*securitygrouprule.AzureSGRule,
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

	securityRules := make([]*securitygrouprule.AzureSGRule, 0)
	for _, v := range sg.SecurityRules {
		securityRules = append(securityRules, az.converCloudToSecurityRule(v))
	}

	return securityRules, nil
}

func (az *Azure) converCloudToSecurityRule(cloud *armnetwork.SecurityRule) *securitygrouprule.AzureSGRule {
	return &securitygrouprule.AzureSGRule{
		ID:                                   SPtrToLowerSPtr(cloud.ID),
		Etag:                                 cloud.Etag,
		Name:                                 SPtrToLowerSPtr(cloud.Name),
		Description:                          cloud.Properties.Description,
		DestinationAddressPrefix:             cloud.Properties.DestinationAddressPrefix,
		DestinationAddressPrefixes:           cloud.Properties.DestinationAddressPrefixes,
		DestinationPortRange:                 cloud.Properties.DestinationPortRange,
		DestinationPortRanges:                cloud.Properties.DestinationPortRanges,
		Protocol:                             cloud.Properties.Protocol,
		ProvisioningState:                    cloud.Properties.ProvisioningState,
		SourceAddressPrefix:                  cloud.Properties.SourceAddressPrefix,
		SourceAddressPrefixes:                cloud.Properties.SourceAddressPrefixes,
		SourcePortRange:                      cloud.Properties.SourcePortRange,
		SourcePortRanges:                     cloud.Properties.SourcePortRanges,
		Priority:                             cloud.Properties.Priority,
		Access:                               cloud.Properties.Access,
		Direction:                            cloud.Properties.Direction,
		DestinationApplicationSecurityGroups: cloud.Properties.DestinationApplicationSecurityGroups,
		SourceApplicationSecurityGroups:      cloud.Properties.SourceApplicationSecurityGroups,
	}
}
