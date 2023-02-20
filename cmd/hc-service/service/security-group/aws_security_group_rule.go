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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	apicore "hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

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

	opt := &types.AwsSGRuleCreateOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]types.AwsSGRuleCreate, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, types.AwsSGRuleCreate{
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				Description:                rule.Memo,
				FromPort:                   rule.FromPort,
				ToPort:                     rule.ToPort,
				Protocol:                   rule.Protocol,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			})
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]types.AwsSGRuleCreate, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, types.AwsSGRuleCreate{
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				Description:                rule.Memo,
				FromPort:                   rule.FromPort,
				ToPort:                     rule.ToPort,
				Protocol:                   rule.Protocol,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			})
		}
	}
	rules, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             *rule.FromPort,
			ToPort:               *rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            req.AccountID,
			Region:               sg.Region,
			SecurityGroupID:      sg.ID,
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
	result, err := g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
	if err != nil {
		return nil, err
	}

	return result, nil
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

	opt := &types.AwsSGRuleUpdateOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		RuleSet: []types.AwsSGRuleUpdate{
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

func (g *securityGroup) getAwsSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	AwsSecurityGroupRule, error) {

	listReq := &protocloud.AwsSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.DefaultBasePage,
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

	opt := &types.AwsSGRuleDeleteOption{
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

		opt := &types.AwsSGRuleListOption{
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
func genAwsRulesList(rules []*ec2.SecurityGroupRule, req *proto.SecurityGroupSyncReq,
	id string) []protocloud.AwsSGRuleBatchCreate {
	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             *rule.FromPort,
			ToPort:               *rule.ToPort,
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

// diffAwsSGRuleSyncUpdate update aws security group rule.
func (g *securityGroup) diffAwsSGRuleSyncUpdate(cts *rest.Contexts, updateCloudIDs []string,
	req *proto.SecurityGroupSyncReq, dsMap map[string]*proto.SecurityGroupSyncDS) error {

	client, err := g.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range updateCloudIDs {

		sgID := dsMap[id].HcSecurityGroup.ID

		opt := &types.AwsSGRuleListOption{
			Region:               req.Region,
			CloudSecurityGroupID: id,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := g.genAwsUpdateRulesList(rules, req, sgID, cts)
		createReq := &protocloud.AwsSGRuleBatchUpdateReq{
			Rules: list,
		}
		if len(createReq.Rules) <= 0 {
			continue
		}
		err = g.dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

// genAwsUpdateRulesList gen AwsSGRuleUpdate list
func (g *securityGroup) genAwsUpdateRulesList(rules []*ec2.SecurityGroupRule, req *proto.SecurityGroupSyncReq,
	sgID string, cts *rest.Contexts) []protocloud.AwsSGRuleUpdate {

	list := make([]protocloud.AwsSGRuleUpdate, 0, len(rules))

	for _, rule := range rules {
		cOne, err := g.getAwsSGRuleByCid(cts, *rule.SecurityGroupRuleId, sgID)
		if err != nil || cOne == nil {
			logs.Errorf("aws gen update RulesList getAwsSGRuleByCid failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}
		if cOne.CloudID == *rule.SecurityGroupRuleId &&
			cOne.IPv4Cidr == rule.CidrIpv4 &&
			cOne.IPv6Cidr == rule.CidrIpv6 &&
			cOne.Memo == rule.Description &&
			cOne.FromPort == *rule.FromPort &&
			cOne.ToPort == *rule.ToPort &&
			cOne.Protocol == rule.IpProtocol &&
			cOne.CloudPrefixListID == rule.PrefixListId &&
			cOne.CloudSecurityGroupID == *rule.GroupId &&
			cOne.CloudGroupOwnerID == *rule.GroupOwnerId {
			continue
		}
		one := protocloud.AwsSGRuleUpdate{
			ID:                   cOne.ID,
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             *rule.FromPort,
			ToPort:               *rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            req.AccountID,
			Region:               req.Region,
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

	return list
}

// getAwsSGRuleByCid
func (g *securityGroup) getAwsSGRuleByCid(cts *rest.Contexts, cID string, sgID string) (*corecloud.
	AwsSecurityGroupRule, error) {

	listReq := &protocloud.AwsSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, id: %s, err: %v, rid: %s", cID, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}

// diffAwsSGRuleSyncDelete delete aws security group rule.
func (g *securityGroup) diffAwsSGRuleSyncDelete(cts *rest.Contexts, deleteCloudIDs []string,
	dsMap map[string]*proto.SecurityGroupSyncDS) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AwsSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_security_group_id", id),
		}
		err := g.dataCli.Aws.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, dsMap[id].HcSecurityGroup.ID)
		if err != nil {
			logs.Errorf("dataservice delete aws security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// SyncAwsSGRule sync aws security group rules.
func (g *securityGroup) SyncAwsSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := g.dataCli.Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &types.AwsSGRuleListOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(rules) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*AwsSGRuleSync)
	cloudIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		sgRuleSync := new(AwsSGRuleSync)
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		cloudMap[*rule.SecurityGroupRuleId] = sgRuleSync
		cloudIDs = append(cloudIDs, *rule.SecurityGroupRuleId)
		cloudAllIDs[*rule.SecurityGroupRuleId] = true
	}

	updateIDs, err := g.getAwsSGRuleDSSync(cloudIDs, req, cts, sgID)
	if err != nil {
		logs.Errorf("request getAwsSGRuleDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := g.syncAwsSGRuleUpdate(updateIDs, cloudMap, sgID, cts, req)
		if err != nil {
			logs.Errorf("request syncPublicImageUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	addIDs := make([]string, 0)
	for _, id := range updateIDs {
		if _, ok := cloudMap[id]; ok {
			cloudMap[id].IsUpdate = true
		}
	}

	for k, v := range cloudMap {
		if !v.IsUpdate {
			addIDs = append(addIDs, k)
		}
	}

	if len(addIDs) > 0 {
		err := g.syncAwsSGRuleAdd(addIDs, cts, req, cloudMap, sgID)
		if err != nil {
			logs.Errorf("request syncAwsSGRuleAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := g.getAwsSGRuleAllDS(req, cts, sgID)
	if err != nil {
		logs.Errorf("request getAwsSGRuleAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		for _, id := range deleteIDs {
			realDeleteFlag := true
			for _, rule := range rules {
				if *rule.SecurityGroupRuleId == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		err = g.syncAwsSGRuleDelete(cts, realDeleteIDs, sgID)
		if err != nil {
			logs.Errorf("request syncAwsSGRuleDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (g *securityGroup) syncAwsSGRuleUpdate(updateIDs []string, cloudMap map[string]*AwsSGRuleSync, sgID string,
	cts *rest.Contexts, req *proto.SecurityGroupSyncReq) error {

	rules := make([]*ec2.SecurityGroupRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	list := g.genAwsUpdateRulesList(rules, req, sgID, cts)
	updateReq := &protocloud.AwsSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := g.dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncAwsSGRuleAdd(addIDs []string, cts *rest.Contexts, req *proto.SecurityGroupSyncReq,
	cloudMap map[string]*AwsSGRuleSync, sgID string) error {

	rules := make([]*ec2.SecurityGroupRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := g.dataCli.Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := genAwsRulesList(rules, req, sg.ID)
	createReq := &protocloud.AwsSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := g.dataCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncAwsSGRuleDelete(cts *rest.Contexts, deleteCloudIDs []string, sgID string) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AwsSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := g.dataCli.Aws.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete aws security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (g *securityGroup) getAwsSGRuleAllDS(req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &protocloud.AwsSGRuleListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "security_group_id",
						Op:    filter.Equal.Factory(),
						Value: sgID,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := g.dataCli.Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}
	return dsIDs, nil
}

func (g *securityGroup) getAwsSGRuleDSSync(cloudIDs []string, req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	updateIDs := make([]string, 0)

	start := 0
	for {

		dataReq := &protocloud.AwsSGRuleListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "cloud_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "security_group_id",
						Op:    filter.Equal.Factory(),
						Value: sgID,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := g.dataCli.Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, nil
}
