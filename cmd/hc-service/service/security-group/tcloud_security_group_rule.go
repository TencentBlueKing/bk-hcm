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

	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// BatchCreateTCloudSGRule batch create tcloud security group rule.
// 腾讯云安全组规则索引是一个动态的，所以每次创建需要将云上安全组规则计算一遍。
func (g *securityGroup) BatchCreateTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.TCloudSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.SecurityGroup().GetTCloudSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.Spec.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.TCloud(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.TCloudSGRuleCreateOption{
		Region:               sg.Spec.Region,
		CloudSecurityGroupID: sg.Spec.CloudID,
		EgressRuleSet:        nil,
		IngressRuleSet:       nil,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]types.TCloudSGRule, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, types.TCloudSGRule{
				Protocol:                   rule.Protocol,
				Port:                       rule.Port,
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
				Action:                     rule.Action,
				Description:                rule.Memo,
			})
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]types.TCloudSGRule, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, types.TCloudSGRule{
				Protocol:                   rule.Protocol,
				Port:                       rule.Port,
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
				Action:                     rule.Action,
				Description:                rule.Memo,
			})
		}
	}
	if err := client.CreateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to create tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               sg.Spec.Region,
		CloudSecurityGroupID: sg.Spec.CloudID,
		SecurityGroupID:      sg.ID,
		AccountID:            sg.Spec.AccountID,
	}
	ids, err := g.syncSecurityGroupRule(cts.Kit, client, syncOpt)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

type syncSecurityGroupRuleOption struct {
	Region               string
	CloudSecurityGroupID string
	SecurityGroupID      string
	AccountID            string
}

// syncSecurityGroupRule 进行云上和db中安全组规则的同步。
// Note: 腾讯云安全组规则 PolicyIndex 是动态变化的，必须同时通过 Version + PolicyIndex 才能唯一确定一个安全组规则，
// 所以每次安全组规则的变动，都需要进行同步。
func (g *securityGroup) syncSecurityGroupRule(kt *kit.Kit, client *tcloud.TCloud, opt *syncSecurityGroupRuleOption) (
	[]string, error) {

	listOpt := &types.TCloudSGRuleListOption{
		Region:               opt.Region,
		CloudSecurityGroupID: opt.CloudSecurityGroupID,
	}
	rules, err := client.ListSecurityGroupRule(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	egressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Egress))
	ingressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Ingress))
	for _, egress := range rules.Egress {
		egressRuleMaps[*egress.PolicyIndex] = egress
	}

	for _, ingress := range rules.Ingress {
		ingressRuleMaps[*ingress.PolicyIndex] = ingress
	}

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", opt.SecurityGroupID),
		Page: &daotypes.BasePage{
			Start: 0,
			Limit: daotypes.DefaultMaxPageLimit,
		},
	}
	start := uint32(0)
	dbRules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := g.dataCli.SecurityGroup().ListTCloudSGRule(kt.Ctx, kt.Header(), listReq, opt.SecurityGroupID)
		if err != nil {
			return nil, err
		}

		dbRules = append(dbRules, listResp.Details...)

		if len(listResp.Details) < int(daotypes.DefaultMaxPageLimit) {
			break
		}

		start += uint32(daotypes.DefaultMaxPageLimit)
	}

	updateRules := make(map[string]*corecloud.TCloudSecurityGroupRuleSpec)
	deleteRuleIDs := make([]string, 0)
	for _, one := range dbRules {
		switch one.Spec.Type {
		case enumor.Egress:
			policy, exist := egressRuleMaps[one.Spec.PolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
			}

			delete(egressRuleMaps, one.Spec.PolicyIndex)
			spec := genSGRuleSpec(policy, *rules.Version, opt.CloudSecurityGroupID, opt.SecurityGroupID, opt.AccountID)
			spec.Type = enumor.Egress

			updateRules[one.ID] = spec

		case enumor.Ingress:
			policy, exist := ingressRuleMaps[one.Spec.PolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
			}

			delete(ingressRuleMaps, one.Spec.PolicyIndex)
			spec := genSGRuleSpec(policy, *rules.Version, opt.CloudSecurityGroupID, opt.SecurityGroupID, opt.AccountID)
			spec.Type = enumor.Ingress

			updateRules[one.ID] = spec

		default:
			logs.Errorf("unknown security group rule type: %s, skip handle, rid: %s", one.Spec.Type, kt.Rid)
		}
	}

	createRules := make([]corecloud.TCloudSecurityGroupRuleSpec, 0)
	for _, policy := range egressRuleMaps {
		spec := genSGRuleSpec(policy, *rules.Version, opt.CloudSecurityGroupID, opt.SecurityGroupID, opt.AccountID)
		spec.Type = enumor.Egress

		createRules = append(createRules, *spec)
	}

	for _, policy := range ingressRuleMaps {
		spec := genSGRuleSpec(policy, *rules.Version, opt.CloudSecurityGroupID, opt.SecurityGroupID, opt.AccountID)
		spec.Type = enumor.Ingress

		createRules = append(createRules, *spec)
	}

	if err = g.updateSecurityGroupRule(kt, opt.SecurityGroupID, updateRules); err != nil {
		return nil, err
	}

	if err = g.deleteSecurityGroupRule(kt, opt.SecurityGroupID, deleteRuleIDs); err != nil {
		return nil, err
	}

	ids, err := g.createSecurityGroupRule(kt, opt.SecurityGroupID, createRules)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (g *securityGroup) deleteSecurityGroupRule(kt *kit.Kit, sgID string, delIDs []string) error {
	req := &protocloud.TCloudSGRuleDeleteReq{
		Filter: tools.ContainersExpression("id", delIDs),
	}
	err := g.dataCli.SecurityGroup().DeleteTCloudSGRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("request dataservice to delete tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (g *securityGroup) createSecurityGroupRule(kt *kit.Kit, sgID string, rules []corecloud.
	TCloudSecurityGroupRuleSpec) ([]string, error) {

	req := &protocloud.TCloudSGRuleCreateReq{
		Rules: rules,
	}
	result, err := g.dataCli.SecurityGroup().BatchCreateTCloudSGRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result.IDs, nil
}

func (g *securityGroup) updateSecurityGroupRule(kt *kit.Kit, sgID string, updateRules map[string]*corecloud.
	TCloudSecurityGroupRuleSpec) error {

	rules := make([]protocloud.TCloudSGRuleBatchUpdateOption, 0, len(updateRules))
	for id, spec := range updateRules {
		rules = append(rules, protocloud.TCloudSGRuleBatchUpdateOption{
			ID:   id,
			Spec: spec,
		})
	}
	req := &protocloud.TCloudSGRuleBatchUpdateReq{
		Rules: rules,
	}
	if err := g.dataCli.SecurityGroup().BatchUpdateTCloudSGRule(kt.Ctx, kt.Header(), req, sgID); err != nil {
		logs.Errorf("request dataservice to batch update tcloud security group rule failed, err: %v, rid: %s", err,
			kt.Rid)
		return err
	}

	return nil
}

func genSGRuleSpec(policy *vpc.SecurityGroupPolicy, version, cloudSGID, sgID, accountID string) *corecloud.
	TCloudSecurityGroupRuleSpec {

	spec := &corecloud.TCloudSecurityGroupRuleSpec{
		PolicyIndex:                *policy.PolicyIndex,
		Version:                    version,
		Protocol:                   policy.Protocol,
		Port:                       policy.Port,
		IPv4Cidr:                   policy.CidrBlock,
		IPv6Cidr:                   policy.Ipv6CidrBlock,
		CloudTargetSecurityGroupID: policy.SecurityGroupId,
		Action:                     *policy.Action,
		Memo:                       policy.PolicyDescription,
		Type:                       enumor.Ingress,
		CloudSecurityGroupID:       cloudSGID,
		SecurityGroupID:            sgID,
		AccountID:                  accountID,
	}

	if policy.ServiceTemplate != nil {
		spec.CloudServiceID = policy.ServiceTemplate.ServiceId
		spec.CloudServiceGroupID = policy.ServiceTemplate.ServiceGroupId
	}

	if policy.AddressTemplate != nil {
		spec.CloudAddressID = policy.AddressTemplate.AddressId
		spec.CloudAddressGroupID = policy.AddressTemplate.AddressGroupId
	}

	return spec
}

// UpdateTCloudSGRule update tcloud security group rule.
func (g *securityGroup) UpdateTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcservice.TCloudSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rule, err := g.getTCloudSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, rule.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.TCloudSGRuleUpdateOption{
		Region:               rule.Spec.Region,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		Version:              rule.Spec.Version,
	}
	switch rule.Spec.Type {
	case enumor.Egress:
		opt.EgressRuleSet = []types.TCloudSGRuleUpdateSpec{
			{
				PolicyIndex:                rule.Spec.PolicyIndex,
				Protocol:                   req.Spec.Protocol,
				Port:                       req.Spec.Port,
				IPv4Cidr:                   req.Spec.IPv4Cidr,
				IPv6Cidr:                   req.Spec.IPv6Cidr,
				CloudTargetSecurityGroupID: req.Spec.CloudTargetSecurityGroupID,
				Action:                     req.Spec.Action,
				Description:                req.Spec.Memo,
			}}

	case enumor.Ingress:
		opt.IngressRuleSet = []types.TCloudSGRuleUpdateSpec{
			{
				PolicyIndex:                rule.Spec.PolicyIndex,
				Protocol:                   req.Spec.Protocol,
				Port:                       req.Spec.Port,
				IPv4Cidr:                   req.Spec.IPv4Cidr,
				IPv6Cidr:                   req.Spec.IPv6Cidr,
				CloudTargetSecurityGroupID: req.Spec.CloudTargetSecurityGroupID,
				Action:                     req.Spec.Action,
				Description:                req.Spec.Memo,
			}}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Spec.Type)
	}

	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               rule.Spec.Region,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		SecurityGroupID:      rule.Spec.SecurityGroupID,
		AccountID:            rule.Spec.AccountID,
	}
	_, err = g.syncSecurityGroupRule(cts.Kit, client, syncOpt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getTCloudSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page: &daotypes.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	listResp, err := g.dataCli.SecurityGroup().ListTCloudSGRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// DeleteTCloudSGRule delete tcloud security group rule.
func (g *securityGroup) DeleteTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := g.getTCloudSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, rule.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.TCloudSGRuleDeleteOption{
		Region:               rule.Spec.Region,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		Version:              rule.Spec.Version,
	}
	switch rule.Spec.Type {
	case enumor.Egress:
		opt.EgressRuleIndexes = []int64{rule.Spec.PolicyIndex}

	case enumor.Ingress:
		opt.IngressRuleIndexes = []int64{rule.Spec.PolicyIndex}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Spec.Type)
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               rule.Spec.Region,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		SecurityGroupID:      rule.Spec.SecurityGroupID,
		AccountID:            rule.Spec.AccountID,
	}
	_, err = g.syncSecurityGroupRule(cts.Kit, client, syncOpt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
