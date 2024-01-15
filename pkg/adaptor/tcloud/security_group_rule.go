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

package tcloud

import (
	"fmt"

	"hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateSecurityGroupRule create security group rule.
// reference: https://cloud.tencent.com/document/api/215/15807
func (t *TCloudImpl) CreateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudCreateOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule create option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateSecurityGroupPoliciesRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudSecurityGroupID)
	req.SecurityGroupPolicySet = new(vpc.SecurityGroupPolicySet)

	if opt.EgressRuleSet != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.EgressRuleSet))

		for _, rule := range opt.EgressRuleSet {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				Protocol:          rule.Protocol,
				Port:              rule.Port,
				CidrBlock:         rule.IPv4Cidr,
				Ipv6CidrBlock:     rule.IPv6Cidr,
				SecurityGroupId:   rule.CloudTargetSecurityGroupID,
				Action:            aws.String(rule.Action),
				PolicyDescription: rule.Description,
			})
		}

		req.SecurityGroupPolicySet.Egress = policies
	}

	if opt.IngressRuleSet != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.IngressRuleSet))

		for _, rule := range opt.IngressRuleSet {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				Protocol:          rule.Protocol,
				Port:              rule.Port,
				CidrBlock:         rule.IPv4Cidr,
				Ipv6CidrBlock:     rule.IPv6Cidr,
				SecurityGroupId:   rule.CloudTargetSecurityGroupID,
				Action:            common.StringPtr(rule.Action),
				PolicyDescription: rule.Description,
			})
		}

		req.SecurityGroupPolicySet.Ingress = policies
	}

	_, err = client.CreateSecurityGroupPoliciesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSecurityGroupRule delete security group.
// reference: https://cloud.tencent.com/document/api/215/15809
func (t *TCloudImpl) DeleteSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteSecurityGroupPoliciesRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudSecurityGroupID)
	req.SecurityGroupPolicySet = new(vpc.SecurityGroupPolicySet)
	req.SecurityGroupPolicySet = &vpc.SecurityGroupPolicySet{
		Version: common.StringPtr(opt.Version),
	}

	if opt.EgressRuleIndexes != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.EgressRuleIndexes))

		for _, index := range opt.EgressRuleIndexes {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				PolicyIndex: common.Int64Ptr(index),
			})
		}

		req.SecurityGroupPolicySet.Egress = policies
	}

	if opt.IngressRuleIndexes != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.IngressRuleIndexes))

		for _, index := range opt.IngressRuleIndexes {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				PolicyIndex: common.Int64Ptr(index),
			})
		}

		req.SecurityGroupPolicySet.Ingress = policies
	}

	_, err = client.DeleteSecurityGroupPoliciesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateSecurityGroupRule update security group.
// reference: https://cloud.tencent.com/document/api/215/15811
func (t *TCloudImpl) UpdateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudUpdateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewReplaceSecurityGroupPolicyRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudSecurityGroupID)
	req.SecurityGroupPolicySet = new(vpc.SecurityGroupPolicySet)
	req.SecurityGroupPolicySet = &vpc.SecurityGroupPolicySet{
		Version: common.StringPtr(opt.Version),
	}

	if opt.EgressRuleSet != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.EgressRuleSet))

		for _, rule := range opt.EgressRuleSet {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				PolicyIndex:       common.Int64Ptr(rule.CloudPolicyIndex),
				Protocol:          rule.Protocol,
				Port:              rule.Port,
				CidrBlock:         rule.IPv4Cidr,
				Ipv6CidrBlock:     rule.IPv6Cidr,
				SecurityGroupId:   rule.CloudTargetSecurityGroupID,
				Action:            aws.String(rule.Action),
				PolicyDescription: rule.Description,
			})
		}

		req.SecurityGroupPolicySet.Egress = policies
	}

	if opt.IngressRuleSet != nil {
		policies := make([]*vpc.SecurityGroupPolicy, 0, len(opt.IngressRuleSet))

		for _, rule := range opt.IngressRuleSet {
			policies = append(policies, &vpc.SecurityGroupPolicy{
				PolicyIndex:       common.Int64Ptr(rule.CloudPolicyIndex),
				Protocol:          rule.Protocol,
				Port:              rule.Port,
				CidrBlock:         rule.IPv4Cidr,
				Ipv6CidrBlock:     rule.IPv6Cidr,
				SecurityGroupId:   rule.CloudTargetSecurityGroupID,
				Action:            common.StringPtr(rule.Action),
				PolicyDescription: rule.Description,
			})
		}

		req.SecurityGroupPolicySet.Ingress = policies
	}

	_, err = client.ReplaceSecurityGroupPolicyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("replace tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroupRule list tcloud security group rule.
// reference: https://cloud.tencent.com/document/api/215/15804
func (t *TCloudImpl) ListSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudListOption) (
	*vpc.SecurityGroupPolicySet, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeSecurityGroupPoliciesRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudSecurityGroupID)

	resp, err := client.DescribeSecurityGroupPoliciesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud security group rules failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Response.SecurityGroupPolicySet, nil
}
