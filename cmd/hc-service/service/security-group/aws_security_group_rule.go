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

package securitygroup

import (
	"fmt"

	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// BatchCreateAwsSGRule batch create aws security group rule.
func (g *securityGroup) BatchCreateAwsSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.AwsSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, id: %s, err: %v, rid: %s", sgID, err,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.Aws(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AwsCreateOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]securitygrouprule.AwsCreate, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, convAwsRule(rule))
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]securitygrouprule.AwsCreate, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, convAwsRule(rule))
		}
	}
	rules, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := convAwsDSReq(rules, req.AccountID, sg.ID, sg.Region)
	result, err := g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		createReq, sgID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// convAwsDSReq convert aws security group rule to data service request.
func convAwsDSReq(rules []*ec2.SecurityGroupRule, accountID, sgID, region string) *protocloud.AwsSGRuleCreateReq {
	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             rule.FromPort,
			ToPort:               rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            accountID,
			Region:               region,
			SecurityGroupID:      sgID,
		}

		if *rule.IsEgress {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}

		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}

		list = append(list, one)
	}

	createReq := &protocloud.AwsSGRuleCreateReq{
		Rules: list,
	}

	return createReq
}

// convAwsRule convert aws security group rule create request to adaptor request.
func convAwsRule(rule proto.AwsSGRuleCreate) securitygrouprule.AwsCreate {
	return securitygrouprule.AwsCreate{
		IPv4Cidr:                   rule.IPv4Cidr,
		IPv6Cidr:                   rule.IPv6Cidr,
		Description:                rule.Memo,
		FromPort:                   rule.FromPort,
		ToPort:                     rule.ToPort,
		Protocol:                   rule.Protocol,
		CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
	}
}

// UpdateAwsSGRule update aws security group rule.
func (g *securityGroup) UpdateAwsSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcservice.AwsSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rule, err := g.getAwsSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Aws(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AwsUpdateOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		RuleSet: []securitygrouprule.AwsSGRuleUpdate{
			{
				CloudID:                    rule.CloudID,
				IPv4Cidr:                   req.IPv4Cidr,
				IPv6Cidr:                   req.IPv6Cidr,
				Description:                req.Memo,
				FromPort:                   req.FromPort,
				ToPort:                     req.ToPort,
				Protocol:                   req.Protocol,
				CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
			},
		},
	}

	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update aws security group rule failed, opt: %v, err: %v, rid: %s", opt, err,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.AwsSGRuleBatchUpdateReq{
		Rules: []protocloud.AwsSGRuleUpdate{
			{
				ID:                         id,
				CloudID:                    rule.CloudID,
				IPv4Cidr:                   req.IPv4Cidr,
				IPv6Cidr:                   req.IPv6Cidr,
				Memo:                       req.Memo,
				FromPort:                   req.FromPort,
				ToPort:                     req.ToPort,
				Protocol:                   req.Protocol,
				CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
				Type:                       rule.Type,
				CloudSecurityGroupID:       rule.CloudSecurityGroupID,
				CloudGroupOwnerID:          rule.CloudGroupOwnerID,
				AccountID:                  rule.AccountID,
				Region:                     rule.Region,
				SecurityGroupID:            rule.SecurityGroupID,
			},
		},
	}
	err = g.dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getAwsSGRuleByID get aws security group rule by id.
func (g *securityGroup) getAwsSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	AwsSecurityGroupRule, error) {

	listReq := &protocloud.AwsSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := g.dataCli.Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, id: %s, err: %v, rid: %s", id, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// DeleteAwsSGRule delete aws security group rule.
func (g *securityGroup) DeleteAwsSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := g.getAwsSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Aws(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AwsDeleteOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.CloudEgressRuleIDs = []string{rule.CloudID}

	case enumor.Ingress:
		opt.CloudIngressRuleIDs = []string{rule.CloudID}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete aws security group rule failed, opt: %v, err: %v, rid: %s", opt, err,
			cts.Kit.Rid)
		return nil, err
	}

	deleteReq := &protocloud.AwsSGRuleBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = g.dataCli.Aws.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// diffAwsSGRuleSyncAdd add tcloud security group rule.
func (g *securityGroup) diffAwsSGRuleSyncAdd(cts *rest.Contexts, ids []string,
	req *proto.SecurityGroupSyncReq) error {

	client, err := g.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sg, err := g.dataCli.Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			logs.Errorf("request dataservice get aws security group failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
			return err
		}

		opt := &securitygrouprule.AwsListOption{
			Region:               req.Region,
			CloudSecurityGroupID: sg.CloudID,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := genAwsRulesList(rules, req, id)
		createReq := &protocloud.AwsSGRuleCreateReq{
			Rules: list,
		}

		if len(createReq.Rules) <= 0 {
			continue
		}

		_, err = g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// genAwsRules gen aws rule list
func genAwsRulesList(rules []securitygrouprule.AwsSGRule, req *proto.SecurityGroupSyncReq,
	id string) []protocloud.AwsSGRuleBatchCreate {
	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             rule.FromPort,
			ToPort:               rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            req.AccountID,
			Region:               req.Region,
			SecurityGroupID:      id,
		}
		if *rule.IsEgress {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}
		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}
		list = append(list, one)
	}
	return list
}
