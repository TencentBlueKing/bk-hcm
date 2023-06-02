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

package aws

import (
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateSecurityGroupRule create security group rule.
func (a *Aws) CreateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AwsCreateOption) (
	[]*ec2.SecurityGroupRule, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(opt.EgressRuleSet) != 0 {
		return a.createEgressSGRule(kt, opt)
	}

	if len(opt.IngressRuleSet) != 0 {
		return a.createIngressSGRule(kt, opt)
	}

	return nil, nil
}

// createEgressSGRule create egress security group rule.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupEgress.html
func (a *Aws) createEgressSGRule(kt *kit.Kit, opt *securitygrouprule.AwsCreateOption) (
	[]*ec2.SecurityGroupRule, error) {

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: aws.String(opt.CloudSecurityGroupID),
	}

	ips := make([]*ec2.IpPermission, 0, len(opt.EgressRuleSet))
	for _, rule := range opt.EgressRuleSet {
		ip := &ec2.IpPermission{
			FromPort:   rule.FromPort,
			IpProtocol: rule.Protocol,
			ToPort:     rule.ToPort,
		}

		if rule.IPv4Cidr != nil && len(*rule.IPv4Cidr) != 0 {
			ip.IpRanges = []*ec2.IpRange{
				{
					Description: rule.Description,
					CidrIp:      rule.IPv4Cidr,
				},
			}
		}

		if rule.IPv6Cidr != nil && len(*rule.IPv6Cidr) != 0 {
			ip.Ipv6Ranges = []*ec2.Ipv6Range{
				{
					Description: rule.Description,
					CidrIpv6:    rule.IPv6Cidr,
				},
			}
		}

		if rule.CloudTargetSecurityGroupID != nil && len(*rule.CloudTargetSecurityGroupID) != 0 {
			ip.UserIdGroupPairs = []*ec2.UserIdGroupPair{
				{
					Description: rule.Description,
					GroupId:     rule.CloudTargetSecurityGroupID,
				},
			}
		}

		ips = append(ips, ip)
	}
	req.IpPermissions = ips

	resp, err := client.AuthorizeSecurityGroupEgressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws security group egress rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.SecurityGroupRules, nil
}

// createIngressSGRule create ingress security group rule.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html
func (a *Aws) createIngressSGRule(kt *kit.Kit, opt *securitygrouprule.AwsCreateOption) (
	[]*ec2.SecurityGroupRule, error) {

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(opt.CloudSecurityGroupID),
	}

	ips := make([]*ec2.IpPermission, 0, len(opt.IngressRuleSet))
	for _, rule := range opt.IngressRuleSet {
		ip := &ec2.IpPermission{
			FromPort:   rule.FromPort,
			IpProtocol: rule.Protocol,
			ToPort:     rule.ToPort,
		}

		if rule.IPv4Cidr != nil && len(*rule.IPv4Cidr) != 0 {
			ip.IpRanges = []*ec2.IpRange{
				{
					Description: rule.Description,
					CidrIp:      rule.IPv4Cidr,
				},
			}
		}

		if rule.IPv6Cidr != nil && len(*rule.IPv6Cidr) != 0 {
			ip.Ipv6Ranges = []*ec2.Ipv6Range{
				{
					Description: rule.Description,
					CidrIpv6:    rule.IPv6Cidr,
				},
			}
		}

		if rule.CloudTargetSecurityGroupID != nil && len(*rule.CloudTargetSecurityGroupID) != 0 {
			ip.UserIdGroupPairs = []*ec2.UserIdGroupPair{
				{
					Description: rule.Description,
					GroupId:     rule.CloudTargetSecurityGroupID,
				},
			}
		}

		ips = append(ips, ip)
	}
	req.IpPermissions = ips

	resp, err := client.AuthorizeSecurityGroupIngressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws security group ingress rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.SecurityGroupRules, nil
}

// DeleteSecurityGroupRule delete security group rule.
// Egress: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_RevokeSecurityGroupEgress.html
// Ingress: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_RevokeSecurityGroupIngress.html
func (a *Aws) DeleteSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AwsDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	if len(opt.CloudEgressRuleIDs) != 0 {
		req := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:              aws.String(opt.CloudSecurityGroupID),
			SecurityGroupRuleIds: aws.StringSlice(opt.CloudEgressRuleIDs),
		}

		_, err = client.RevokeSecurityGroupEgressWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("revoke aws security group egress rule failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		return nil
	}

	if len(opt.CloudIngressRuleIDs) != 0 {
		req := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:              aws.String(opt.CloudSecurityGroupID),
			SecurityGroupRuleIds: aws.StringSlice(opt.CloudIngressRuleIDs),
		}

		_, err = client.RevokeSecurityGroupIngressWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("revoke aws security group ingress rule failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		return nil
	}

	return nil
}

// ListSecurityGroupRule list security group rule.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DescribeSecurityGroupRules.html
func (a *Aws) ListSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AwsListOption) (
	[]securitygrouprule.AwsSGRule, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.DescribeSecurityGroupRulesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: aws.StringSlice([]string{opt.CloudSecurityGroupID}),
			},
		},
	}

	if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeSecurityGroupRulesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws security group rules failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	results := make([]securitygrouprule.AwsSGRule, 0)
	for _, one := range resp.SecurityGroupRules {
		results = append(results, securitygrouprule.AwsSGRule{one})
	}

	return results, nil
}

// UpdateSecurityGroupRule update security group rule.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_ModifySecurityGroupRules.html
func (a *Aws) UpdateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.AwsUpdateOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.ModifySecurityGroupRulesInput{
		GroupId: aws.String(opt.CloudSecurityGroupID),
	}

	rules := make([]*ec2.SecurityGroupRuleUpdate, 0, len(opt.RuleSet))
	for _, rule := range opt.RuleSet {
		rules = append(rules, &ec2.SecurityGroupRuleUpdate{
			SecurityGroupRuleId: aws.String(rule.CloudID),
			SecurityGroupRule: &ec2.SecurityGroupRuleRequest{
				CidrIpv4:          rule.IPv4Cidr,
				CidrIpv6:          rule.IPv6Cidr,
				Description:       rule.Description,
				FromPort:          rule.FromPort,
				IpProtocol:        rule.Protocol,
				ReferencedGroupId: rule.CloudTargetSecurityGroupID,
				ToPort:            rule.ToPort,
			},
		})
	}
	req.SecurityGroupRules = rules

	_, err = client.ModifySecurityGroupRulesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("modify aws security group rules failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
