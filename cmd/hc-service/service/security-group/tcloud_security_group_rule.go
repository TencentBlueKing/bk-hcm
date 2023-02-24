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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	apicore "hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

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

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		SecurityGroupID:      sg.ID,
		AccountID:            sg.AccountID,
	}

	opt := &securitygrouprule.TCloudCreateOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]securitygrouprule.TCloud, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, securitygrouprule.TCloud{
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
		opt.IngressRuleSet = make([]securitygrouprule.TCloud, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, securitygrouprule.TCloud{
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

		if _, syncErr := g.syncSecurityGroupRule(cts.Kit, client, syncOpt); syncErr != nil {
			logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", syncErr, syncOpt, cts.Kit.Rid)
		}

		return nil, err
	}

	ids, err := g.syncSecurityGroupRule(cts.Kit, client, syncOpt)
	if err != nil {
		logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", err, syncOpt, cts.Kit.Rid)
		return nil, err
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

type syncSecurityGroupRuleOption struct {
	Region               string
	CloudSecurityGroupID string
	SecurityGroupID      string
	AccountID            string
}

// syncSecurityGroupRule 进行云上和db中安全组规则的同步。
// Note: 腾讯云安全组规则 CloudPolicyIndex 是动态变化的，必须同时通过 Version + CloudPolicyIndex 才能唯一确定一个安全组规则，
// 所以每次安全组规则的变动，都需要进行同步。
func (g *securityGroup) syncSecurityGroupRule(kt *kit.Kit, client *tcloud.TCloud, opt *syncSecurityGroupRuleOption) (
	[]string, error) {

	listOpt := &securitygrouprule.TCloudListOption{
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
		Page:   core.DefaultBasePage,
	}
	start := uint32(0)
	dbRules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
			opt.SecurityGroupID)
		if err != nil {
			return nil, err
		}

		dbRules = append(dbRules, listResp.Details...)

		if len(listResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	updateRules := make(map[string]*corecloud.TCloudSecurityGroupRule)
	deleteRuleIDs := make([]string, 0)
	for _, one := range dbRules {
		switch one.Type {
		case enumor.Egress:
			policy, exist := egressRuleMaps[one.CloudPolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
				continue
			}

			rule := genSGRuleSpec(policy, *rules.Version, opt)
			rule.Type = enumor.Egress

			delete(egressRuleMaps, one.CloudPolicyIndex)
			updateRules[one.ID] = rule

		case enumor.Ingress:
			policy, exist := ingressRuleMaps[one.CloudPolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
				continue
			}

			rule := genSGRuleSpec(policy, *rules.Version, opt)
			rule.Type = enumor.Ingress

			delete(ingressRuleMaps, one.CloudPolicyIndex)
			updateRules[one.ID] = rule

		default:
			logs.Errorf("unknown security group rule type: %s, skip handle, rid: %s", one.Type, kt.Rid)
		}
	}

	createRules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for _, policy := range egressRuleMaps {
		rule := genSGRuleSpec(policy, *rules.Version, opt)
		rule.Type = enumor.Egress

		createRules = append(createRules, *rule)
	}

	for _, policy := range ingressRuleMaps {
		rule := genSGRuleSpec(policy, *rules.Version, opt)
		rule.Type = enumor.Ingress

		createRules = append(createRules, *rule)
	}

	if len(updateRules) != 0 {
		if err = g.updateSecurityGroupRule(kt, opt.SecurityGroupID, updateRules); err != nil {
			return nil, err
		}
	}

	if len(deleteRuleIDs) != 0 {
		if err = g.deleteSecurityGroupRule(kt, opt.SecurityGroupID, deleteRuleIDs); err != nil {
			return nil, err
		}
	}

	if len(createRules) != 0 {
		ids, err := g.createSecurityGroupRule(kt, opt.SecurityGroupID, createRules)
		if err != nil {
			return nil, err
		}

		return ids, nil
	}

	return make([]string, 0), nil
}

func (g *securityGroup) deleteSecurityGroupRule(kt *kit.Kit, sgID string, delIDs []string) error {
	req := &protocloud.TCloudSGRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("id", delIDs),
	}
	err := g.dataCli.TCloud.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("request dataservice to delete tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (g *securityGroup) createSecurityGroupRule(kt *kit.Kit, sgID string, rules []corecloud.
	TCloudSecurityGroupRule) ([]string, error) {

	ruleCreates := make([]protocloud.TCloudSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		ruleCreates = append(ruleCreates, protocloud.TCloudSGRuleBatchCreate{
			CloudPolicyIndex:           rule.CloudPolicyIndex,
			Version:                    rule.Version,
			Protocol:                   rule.Protocol,
			Port:                       rule.Port,
			CloudServiceID:             rule.CloudServiceID,
			CloudServiceGroupID:        rule.CloudServiceGroupID,
			IPv4Cidr:                   rule.IPv4Cidr,
			IPv6Cidr:                   rule.IPv6Cidr,
			CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			CloudAddressID:             rule.CloudAddressID,
			CloudAddressGroupID:        rule.CloudAddressGroupID,
			Action:                     rule.Action,
			Memo:                       rule.Memo,
			Type:                       rule.Type,
			CloudSecurityGroupID:       rule.CloudSecurityGroupID,
			SecurityGroupID:            rule.SecurityGroupID,
			Region:                     rule.Region,
			AccountID:                  rule.AccountID,
		})
	}
	req := &protocloud.TCloudSGRuleCreateReq{
		Rules: ruleCreates,
	}
	result, err := g.dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result.IDs, nil
}

func (g *securityGroup) updateSecurityGroupRule(kt *kit.Kit, sgID string, updateRules map[string]*corecloud.
	TCloudSecurityGroupRule) error {

	rules := make([]protocloud.TCloudSGRuleBatchUpdate, 0, len(updateRules))
	for id, rule := range updateRules {
		rules = append(rules, protocloud.TCloudSGRuleBatchUpdate{
			ID:                         id,
			CloudPolicyIndex:           rule.CloudPolicyIndex,
			Version:                    rule.Version,
			Protocol:                   rule.Protocol,
			Port:                       rule.Port,
			CloudServiceID:             rule.CloudServiceID,
			CloudServiceGroupID:        rule.CloudServiceGroupID,
			IPv4Cidr:                   rule.IPv4Cidr,
			IPv6Cidr:                   rule.IPv6Cidr,
			CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			CloudAddressID:             rule.CloudAddressID,
			CloudAddressGroupID:        rule.CloudAddressGroupID,
			Action:                     rule.Action,
			Memo:                       rule.Memo,
			Type:                       rule.Type,
			CloudSecurityGroupID:       rule.CloudSecurityGroupID,
			SecurityGroupID:            rule.SecurityGroupID,
			Region:                     rule.Region,
			AccountID:                  rule.AccountID,
		})
	}
	req := &protocloud.TCloudSGRuleBatchUpdateReq{
		Rules: rules,
	}
	if err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID); err != nil {
		logs.Errorf("request dataservice to batch update tcloud security group rule failed, err: %v, rid: %s", err,
			kt.Rid)
		return err
	}

	return nil
}

func genSGRuleSpec(policy *vpc.SecurityGroupPolicy, version string, opt *syncSecurityGroupRuleOption) *corecloud.
	TCloudSecurityGroupRule {

	spec := &corecloud.TCloudSecurityGroupRule{
		CloudPolicyIndex:           *policy.PolicyIndex,
		Version:                    version,
		Protocol:                   policy.Protocol,
		Port:                       policy.Port,
		IPv4Cidr:                   policy.CidrBlock,
		IPv6Cidr:                   policy.Ipv6CidrBlock,
		CloudTargetSecurityGroupID: policy.SecurityGroupId,
		Action:                     *policy.Action,
		Memo:                       policy.PolicyDescription,
		Type:                       enumor.Ingress,
		CloudSecurityGroupID:       opt.CloudSecurityGroupID,
		SecurityGroupID:            opt.SecurityGroupID,
		Region:                     opt.Region,
		AccountID:                  opt.AccountID,
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

	client, err := g.ad.TCloud(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		SecurityGroupID:      rule.SecurityGroupID,
		AccountID:            rule.AccountID,
	}

	opt := &securitygrouprule.TCloudUpdateOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		Version:              rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleSet = []securitygrouprule.TCloudUpdateSpec{
			{
				CloudPolicyIndex:           rule.CloudPolicyIndex,
				Protocol:                   req.Protocol,
				Port:                       req.Port,
				IPv4Cidr:                   req.IPv4Cidr,
				IPv6Cidr:                   req.IPv6Cidr,
				CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
				Action:                     req.Action,
				Description:                req.Memo,
			}}

	case enumor.Ingress:
		opt.IngressRuleSet = []securitygrouprule.TCloudUpdateSpec{
			{
				CloudPolicyIndex:           rule.CloudPolicyIndex,
				Protocol:                   req.Protocol,
				Port:                       req.Port,
				IPv4Cidr:                   req.IPv4Cidr,
				IPv6Cidr:                   req.IPv6Cidr,
				CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
				Action:                     req.Action,
				Description:                req.Memo,
			}}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}

	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)

		if _, syncErr := g.syncSecurityGroupRule(cts.Kit, client, syncOpt); syncErr != nil {
			logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", syncErr, syncOpt, cts.Kit.Rid)
		}

		return nil, err
	}

	if _, syncErr := g.syncSecurityGroupRule(cts.Kit, client, syncOpt); syncErr != nil {
		logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", syncErr, syncOpt, cts.Kit.Rid)
		return nil, syncErr
	}

	return nil, nil
}

func (g *securityGroup) getTCloudSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
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

	client, err := g.ad.TCloud(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncOpt := &syncSecurityGroupRuleOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		SecurityGroupID:      rule.SecurityGroupID,
		AccountID:            rule.AccountID,
	}

	opt := &securitygrouprule.TCloudDeleteOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		Version:              rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleIndexes = []int64{rule.CloudPolicyIndex}

	case enumor.Ingress:
		opt.IngressRuleIndexes = []int64{rule.CloudPolicyIndex}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)

		if _, syncErr := g.syncSecurityGroupRule(cts.Kit, client, syncOpt); syncErr != nil {
			logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", syncErr, syncOpt, cts.Kit.Rid)
		}

		return nil, err
	}

	if _, syncErr := g.syncSecurityGroupRule(cts.Kit, client, syncOpt); syncErr != nil {
		logs.Errorf("sync tcloud security group failed, err: %v, opt: %v, rid: %s", syncErr, syncOpt, cts.Kit.Rid)
		return nil, syncErr
	}

	return nil, nil
}

// diffTCloudSGRuleSyncAdd add tcloud security group rule.
func (g *securityGroup) diffTCloudSGRuleSyncAdd(cts *rest.Contexts, ids []string,
	req *proto.SecurityGroupSyncReq) error {

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range ids {

		sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			logs.Errorf("request dataservice get tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		listOpt := &securitygrouprule.TCloudListOption{
			Region:               req.Region,
			CloudSecurityGroupID: sg.CloudID,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
		if len(rules.Egress)+len(rules.Ingress) <= 0 {
			continue
		}

		createRules := []corecloud.TCloudSecurityGroupRule{}
		opt := &syncSecurityGroupRuleOption{
			Region:               req.Region,
			CloudSecurityGroupID: sg.CloudID,
			SecurityGroupID:      id,
			AccountID:            req.AccountID,
		}
		for _, eRule := range rules.Egress {
			createRules = append(createRules, *genTCloudSGRuleSpecByType(eRule, *rules.Version, enumor.Egress, opt))
		}
		for _, iRule := range rules.Ingress {
			createRules = append(createRules, *genTCloudSGRuleSpecByType(iRule, *rules.Version, enumor.Ingress, opt))
		}

		_, err = g.createSecurityGroupRule(cts.Kit, id, createRules)
		if err != nil {
			return err
		}
	}

	return nil
}

// genTCloudSGRuleSpecByType gen TCloudSecurityGroupRule struct
func genTCloudSGRuleSpecByType(policy *vpc.SecurityGroupPolicy, version string, typ enumor.SecurityGroupRuleType,
	opt *syncSecurityGroupRuleOption) *corecloud.TCloudSecurityGroupRule {

	spec := &corecloud.TCloudSecurityGroupRule{
		CloudPolicyIndex:           *policy.PolicyIndex,
		Version:                    version,
		Protocol:                   policy.Protocol,
		Port:                       policy.Port,
		IPv4Cidr:                   policy.CidrBlock,
		IPv6Cidr:                   policy.Ipv6CidrBlock,
		CloudTargetSecurityGroupID: policy.SecurityGroupId,
		Action:                     *policy.Action,
		Memo:                       policy.PolicyDescription,
		Type:                       typ,
		CloudSecurityGroupID:       opt.CloudSecurityGroupID,
		SecurityGroupID:            opt.SecurityGroupID,
		Region:                     opt.Region,
		AccountID:                  opt.AccountID,
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

// diffTCloudSGRuleSyncUpdate update tcloud security group rule.
func (g *securityGroup) diffTCloudSGRuleSyncUpdate(cts *rest.Contexts, updateCloudIDs []string,
	req *proto.SecurityGroupSyncReq, dsMap map[string]*proto.SecurityGroupSyncDS) error {

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range updateCloudIDs {

		sgID := dsMap[id].HcSecurityGroup.ID

		listOpt := &securitygrouprule.TCloudListOption{
			Region:               req.Region,
			CloudSecurityGroupID: id,
		}
		cloudRules, err := client.ListSecurityGroupRule(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := g.genTCloudRulesList(cloudRules, cts, sgID, id)
		req := &protocloud.TCloudSGRuleBatchUpdateReq{
			Rules: list,
		}
		if len(req.Rules) <= 0 {
			continue
		}
		if err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			req, sgID); err != nil {
			logs.Errorf("request dataservice to batch update tcloud security group rule failed, err: %v, rid: %s", err,
				cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// genTCloudRulesList gen protocloud.TCloudSGRuleBatchUpdate list
func (g *securityGroup) genTCloudRulesList(cloudRules *vpc.SecurityGroupPolicySet, cts *rest.Contexts,
	sgID string, id string) []protocloud.TCloudSGRuleBatchUpdate {
	list := make([]protocloud.TCloudSGRuleBatchUpdate, 0)
	for _, rule := range cloudRules.Egress {
		rID, err := g.getTCloudSGRuleBy(cts, sgID, *rule.PolicyIndex, id, enumor.Egress)
		if err != nil {
			logs.Errorf("tcloud gen RulesList getTCloudSGRuleBy failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}

		if *rID.Protocol == *rule.Protocol &&
			*rID.Port == *rule.Port &&
			*rID.IPv4Cidr == *rule.CidrBlock &&
			*rID.IPv6Cidr == *rule.Ipv6CidrBlock &&
			*rID.Memo == *rule.PolicyDescription {
			continue
		}

		list = append(list, protocloud.TCloudSGRuleBatchUpdate{
			ID:       rID.ID,
			Protocol: rule.Protocol,
			Port:     rule.Port,
			IPv4Cidr: rule.CidrBlock,
			IPv6Cidr: rule.Ipv6CidrBlock,
			Memo:     rule.PolicyDescription,
		})
	}

	for _, rule := range cloudRules.Ingress {
		rID, err := g.getTCloudSGRuleBy(cts, sgID, *rule.PolicyIndex, id, enumor.Ingress)
		if err != nil {
			continue
		}

		if *rID.Protocol == *rule.Protocol &&
			*rID.Port == *rule.Port &&
			*rID.IPv4Cidr == *rule.CidrBlock &&
			*rID.IPv6Cidr == *rule.Ipv6CidrBlock &&
			*rID.Memo == *rule.PolicyDescription {
			continue
		}

		list = append(list, protocloud.TCloudSGRuleBatchUpdate{
			ID:       rID.ID,
			Protocol: rule.Protocol,
			Port:     rule.Port,
			IPv4Cidr: rule.CidrBlock,
			IPv6Cidr: rule.Ipv6CidrBlock,
			Memo:     rule.PolicyDescription,
		})
	}

	return list
}

// getTCloudSGRuleBy
func (g *securityGroup) getTCloudSGRuleBy(cts *rest.Contexts, sgID string, cpId int64,
	cId string, typ enumor.SecurityGroupRuleType) (*corecloud.TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "cloud_policy_index", Op: filter.Equal.Factory(), Value: cpId},
				filter.AtomRule{Field: "cloud_security_group_id", Op: filter.Equal.Factory(), Value: cId},
				filter.AtomRule{Field: "security_group_id", Op: filter.Equal.Factory(), Value: sgID},
				filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: typ},
			},
		},
		Page: core.DefaultBasePage,
	}

	listResp, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, id: %s, err: %v, rid: %s", cpId, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cpId)
	}

	return &listResp.Details[0], nil
}

// diffTCloudSGRuleSyncDelete delete tcloud security group rule.
func (g *securityGroup) diffTCloudSGRuleSyncDelete(cts *rest.Contexts, deleteCloudIDs []string,
	dsMap map[string]*proto.SecurityGroupSyncDS) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.TCloudSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_security_group_id", id),
		}
		err := g.dataCli.TCloud.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, dsMap[id].HcSecurityGroup.ID)
		if err != nil {
			logs.Errorf("dataservice delete tcloud security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// SyncTCloudSGRule sync tcloud security group rules.
func (g *securityGroup) SyncTCloudSGRule(cts *rest.Contexts) (interface{}, error) {

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get TCloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err, err
	}

	opt := &securitygrouprule.TCloudListOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list TCloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(rules.Egress) <= 0 && len(rules.Ingress) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*TCloudSGRuleSync)
	for _, rule := range rules.Egress {
		sgRuleSync := new(TCloudSGRuleSync)
		sgRuleSync.Version = *rules.Version
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		sgRuleSync.Typ = enumor.Egress
		id := getTCloudSGRuleID(*rule.PolicyIndex, sg.CloudID, enumor.Egress)
		cloudMap[id] = sgRuleSync
	}

	for _, rule := range rules.Ingress {
		sgRuleSync := new(TCloudSGRuleSync)
		sgRuleSync.Version = *rules.Version
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		sgRuleSync.Typ = enumor.Ingress
		id := getTCloudSGRuleID(*rule.PolicyIndex, sg.CloudID, enumor.Ingress)
		cloudMap[id] = sgRuleSync
	}

	updateIDs, err := g.getTCloudSGRuleDSSync(cloudMap, req, cts, sgID)
	if err != nil {
		logs.Errorf("request getTCloudSGRuleDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := g.syncTCloudSGRuleUpdate(updateIDs, cloudMap, sgID, cts, req)
		if err != nil {
			logs.Errorf("request syncTCloudSGRuleUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		err := g.syncTCloudSGRuleAdd(addIDs, cts, req, cloudMap, sgID)
		if err != nil {
			logs.Errorf("request syncTCloudSGRuleAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsMap, err := g.getTCloudSGRuleAllDS(req, cts, sgID)
	if err != nil {
		logs.Errorf("request getTCloudSGRuleAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteMap := make(map[string]*TCloudSGRuleSync)
	for dsKey, dsValue := range dsMap {
		if _, ok := cloudMap[dsKey]; !ok {
			deleteMap[dsKey] = dsValue
		}
	}

	if len(deleteMap) > 0 {
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list TCloud security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		for _, rule := range rules.Egress {
			id := getTCloudSGRuleID(*rule.PolicyIndex, sg.CloudID, enumor.Egress)
			if _, ok := deleteMap[id]; ok {
				delete(deleteMap, id)
			}
		}

		for _, rule := range rules.Ingress {
			id := getTCloudSGRuleID(*rule.PolicyIndex, sg.CloudID, enumor.Ingress)
			if _, ok := deleteMap[id]; ok {
				delete(deleteMap, id)
			}
		}

		err = g.syncTCloudSGRuleDelete(cts, deleteMap, sgID, req)
		if err != nil {
			logs.Errorf("request syncTCloudSGRuleDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (g *securityGroup) syncTCloudSGRuleUpdate(updateIDs []string, cloudMap map[string]*TCloudSGRuleSync, sgID string,
	cts *rest.Contexts, req *proto.SecurityGroupSyncReq) error {

	rules := make([]*TCloudSGRuleSync, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value)
		}
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := g.genTCloudUpdateRulesList(rules, cts, sgID, sg.CloudID)
	updateReq := &protocloud.TCloudSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		err := g.dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncTCloudSGRuleAdd(addIDs []string, cts *rest.Contexts, req *proto.SecurityGroupSyncReq,
	cloudMap map[string]*TCloudSGRuleSync, sgID string) error {

	rules := make([]*TCloudSGRuleSync, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value)
		}
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	createRules := []corecloud.TCloudSecurityGroupRule{}
	opt := &syncSecurityGroupRuleOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
		SecurityGroupID:      sgID,
		AccountID:            req.AccountID,
	}
	for _, rule := range rules {
		createRules = append(createRules, *genTCloudSGRuleSpecByType(rule.SGRule, rule.Version, rule.Typ, opt))
	}

	_, err = g.createSecurityGroupRule(cts.Kit, sgID, createRules)
	if err != nil {
		return err
	}

	return nil
}

func (g *securityGroup) syncTCloudSGRuleDelete(cts *rest.Contexts, deleteMap map[string]*TCloudSGRuleSync,
	sgID string, req *proto.SecurityGroupSyncReq) error {

	for _, v := range deleteMap {
		deleteReq := &protocloud.TCloudSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("id", v.SGRuleID),
		}

		err := g.dataCli.TCloud.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete tcloud security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (g *securityGroup) getTCloudSGRuleAllDS(req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) (map[string]*TCloudSGRuleSync, error) {

	start := 0
	dsMap := make(map[string]*TCloudSGRuleSync)
	for {

		dataReq := &protocloud.TCloudSGRuleListReq{
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

		results, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				id := getTCloudSGRuleID(detail.CloudPolicyIndex, detail.CloudSecurityGroupID, detail.Type)
				tcloudSync := new(TCloudSGRuleSync)
				tcloudSync.SGRuleID = detail.ID
				dsMap[id] = tcloudSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}
	return dsMap, nil
}

func (g *securityGroup) getTCloudSGRuleDSSync(cloudMap map[string]*TCloudSGRuleSync, req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	updateIDs := make([]string, 0)

	for _, v := range cloudMap {

		dataReq := &protocloud.TCloudSGRuleListReq{
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
						Value: v.SGRule.SecurityGroupId,
					},
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: string(v.Typ),
					},
					&filter.AtomRule{
						Field: "cloud_policy_index",
						Op:    filter.Equal.Factory(),
						Value: *v.SGRule.PolicyIndex,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(0),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				id := getTCloudSGRuleID(detail.CloudPolicyIndex, detail.CloudSecurityGroupID, detail.Type)
				updateIDs = append(updateIDs, id)
			}
		}

	}

	return updateIDs, nil
}

func (g *securityGroup) genTCloudUpdateRulesList(sgRule []*TCloudSGRuleSync, cts *rest.Contexts,
	sgID string, id string) []protocloud.TCloudSGRuleBatchUpdate {

	list := make([]protocloud.TCloudSGRuleBatchUpdate, 0)

	for _, rule := range sgRule {
		rID, err := g.getTCloudSGRuleBy(cts, sgID, *rule.SGRule.PolicyIndex, id, rule.Typ)
		if err != nil {
			logs.Errorf("tcloud gen RulesList getTCloudSGRuleBy failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}

		if *rID.Protocol == *rule.SGRule.Protocol &&
			*rID.Port == *rule.SGRule.Port &&
			*rID.IPv4Cidr == *rule.SGRule.CidrBlock &&
			*rID.IPv6Cidr == *rule.SGRule.Ipv6CidrBlock &&
			*rID.Memo == *rule.SGRule.PolicyDescription {
			continue
		}

		list = append(list, protocloud.TCloudSGRuleBatchUpdate{
			ID:       rID.ID,
			Protocol: rule.SGRule.Protocol,
			Port:     rule.SGRule.Port,
			IPv4Cidr: rule.SGRule.CidrBlock,
			IPv6Cidr: rule.SGRule.Ipv6CidrBlock,
			Memo:     rule.SGRule.PolicyDescription,
		})
	}

	return list
}

func getTCloudSGRuleID(pIndex int64, sgID string, typ enumor.SecurityGroupRuleType) string {
	flag := strconv.FormatInt(pIndex, 10) + sgID + string(typ)
	h := md5.New()
	h.Write([]byte(flag))
	return hex.EncodeToString(h.Sum(nil))
}
