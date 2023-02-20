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
	"strconv"

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

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// CreateHuaWeiSGRule create huawei security group rule.
func (g *securityGroup) CreateHuaWeiSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.HuaWeiSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSGRuleCreateOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
	}
	if req.EgressRule != nil {
		priority := strconv.Itoa(int(req.EgressRule.Priority))
		opt.Rule = &types.HuaWeiSGRuleCreate{
			Description:        req.EgressRule.Memo,
			Ethertype:          req.EgressRule.Ethertype,
			Protocol:           req.EgressRule.Protocol,
			RemoteIPPrefix:     req.EgressRule.RemoteIPPrefix,
			CloudRemoteGroupID: req.EgressRule.CloudRemoteGroupID,
			Port:               req.EgressRule.Port,
			Action:             req.EgressRule.Action,
			Priority:           &priority,
			Type:               enumor.Egress,
		}
	}

	if req.IngressRule != nil {
		priority := strconv.Itoa(int(req.IngressRule.Priority))
		opt.Rule = &types.HuaWeiSGRuleCreate{
			Description:        req.IngressRule.Memo,
			Ethertype:          req.IngressRule.Ethertype,
			Protocol:           req.IngressRule.Protocol,
			RemoteIPPrefix:     req.IngressRule.RemoteIPPrefix,
			CloudRemoteGroupID: req.IngressRule.CloudRemoteGroupID,
			Port:               req.IngressRule.Port,
			Action:             req.IngressRule.Action,
			Priority:           &priority,
			Type:               enumor.Ingress,
		}
	}
	rule, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.HuaWeiSGRuleCreateReq{
		Rules: []protocloud.HuaWeiSGRuleBatchCreate{
			{
				CloudID:                   rule.Id,
				Memo:                      &rule.Description,
				Protocol:                  rule.Protocol,
				Ethertype:                 rule.Ethertype,
				CloudRemoteGroupID:        rule.RemoteGroupId,
				RemoteIPPrefix:            rule.RemoteIpPrefix,
				CloudRemoteAddressGroupID: rule.RemoteAddressGroupId,
				Port:                      rule.Multiport,
				Priority:                  int64(rule.Priority),
				Action:                    rule.Action,
				Type:                      opt.Rule.Type,
				CloudSecurityGroupID:      sg.CloudID,
				CloudProjectID:            rule.ProjectId,
				AccountID:                 req.AccountID,
				Region:                    sg.Region,
				SecurityGroupID:           sg.ID,
			},
		},
	}
	result, err := g.dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		createReq, sgID)
	if err != nil {
		return nil, err
	}

	if len(result.IDs) != 1 {
		logs.Errorf("batch create security group rule success, but return id count: %d not right, rid: %s",
			len(result.IDs), cts.Kit.Rid)

		return nil, fmt.Errorf("batch create security group rule success, but return id count: %d not right",
			len(result.IDs))
	}

	return &core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteHuaWeiSGRule delete huawei security group rule.
func (g *securityGroup) DeleteHuaWeiSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := g.getHuaWeiSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSGRuleDeleteOption{
		Region:      rule.Region,
		CloudRuleID: rule.CloudID,
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	deleteReq := &protocloud.HuaWeiSGRuleBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = g.dataCli.HuaWei.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getHuaWeiSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	HuaWeiSecurityGroupRule, error) {

	listReq := &protocloud.HuaWeiSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// diffHuaWeiSGRuleSyncAdd add huawei security group rule.
func (g *securityGroup) diffHuaWeiSGRuleSyncAdd(cts *rest.Contexts, ids []string,
	req *proto.SecurityGroupSyncReq) error {

	client, err := g.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			logs.Errorf("request dataservice get huawei security group failed, id: %s, err: %v, rid: %s", id, err,
				cts.Kit.Rid)
			return err
		}
		// TODO 分页逻辑
		opt := &types.HuaWeiSGRuleListOption{
			Region:               req.Region,
			CloudSecurityGroupID: sg.CloudID,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := genHuaWeiRuleList(rules, req, sg.CloudID, id)
		createReq := &protocloud.HuaWeiSGRuleCreateReq{
			Rules: list,
		}

		if len(createReq.Rules) <= 0 {
			continue
		}
		_, err = g.dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// genHuaWeiCreateReq gen protocloud.HuaWeiSGRuleBatchCreate list
func genHuaWeiRuleList(rules *model.ListSecurityGroupRulesResponse, req *proto.SecurityGroupSyncReq,
	sgCloudID string, id string) []protocloud.HuaWeiSGRuleBatchCreate {
	list := make([]protocloud.HuaWeiSGRuleBatchCreate, 0, len(*rules.SecurityGroupRules))
	for _, sgRule := range *rules.SecurityGroupRules {
		rule := protocloud.HuaWeiSGRuleBatchCreate{
			CloudID:                   sgRule.Id,
			Memo:                      &sgRule.Description,
			Protocol:                  sgRule.Protocol,
			Ethertype:                 sgRule.Ethertype,
			CloudRemoteGroupID:        sgRule.RemoteGroupId,
			RemoteIPPrefix:            sgRule.RemoteIpPrefix,
			CloudRemoteAddressGroupID: sgRule.RemoteAddressGroupId,
			Port:                      sgRule.Multiport,
			Priority:                  int64(sgRule.Priority),
			Action:                    sgRule.Action,
			Type:                      enumor.SecurityGroupRuleType(sgRule.Direction),
			CloudSecurityGroupID:      sgCloudID,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 req.AccountID,
			Region:                    req.Region,
			SecurityGroupID:           id,
		}
		list = append(list, rule)
	}
	return list
}

// diffHuaWeiSGRuleSyncUpdate update huawei security group rule.
func (g *securityGroup) diffHuaWeiSGRuleSyncUpdate(cts *rest.Contexts, updateCloudIDs []string,
	req *proto.SecurityGroupSyncReq, dsMap map[string]*proto.SecurityGroupSyncDS) error {

	client, err := g.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range updateCloudIDs {
		sgID := dsMap[id].HcSecurityGroup.ID
		// TODO 分页逻辑
		opt := &types.HuaWeiSGRuleListOption{
			Region:               req.Region,
			CloudSecurityGroupID: id,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := g.genHuaWeiUpdateRulesList(rules, cts, sgID, id, req)
		createReq := &protocloud.HuaWeiSGRuleBatchUpdateReq{
			Rules: list,
		}

		if len(createReq.Rules) <= 0 {
			continue
		}
		err = g.dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

// genHuaWeiUpdateRulesList gen HuaWeiSGRuleBatchUpdate list
func (g *securityGroup) genHuaWeiUpdateRulesList(rules *model.ListSecurityGroupRulesResponse, cts *rest.Contexts,
	sgID string, id string, req *proto.SecurityGroupSyncReq) []protocloud.HuaWeiSGRuleBatchUpdate {

	list := make([]protocloud.HuaWeiSGRuleBatchUpdate, 0)

	for _, sgRule := range *rules.SecurityGroupRules {
		one, err := g.getHuaWeiSGRuleByCid(cts, sgRule.Id, sgID)
		if err != nil || one == nil {
			logs.Errorf("huawei gen update RulesList getHuaWeiSGRuleByCid failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}
		if *one.Memo == sgRule.Description &&
			one.Protocol == sgRule.Protocol &&
			one.Ethertype == sgRule.Ethertype &&
			one.CloudRemoteGroupID == sgRule.RemoteGroupId &&
			one.RemoteIPPrefix == sgRule.RemoteIpPrefix &&
			one.CloudRemoteAddressGroupID == sgRule.RemoteAddressGroupId &&
			one.Port == sgRule.Multiport &&
			one.Priority == int64(sgRule.Priority) &&
			one.Action == sgRule.Action &&
			one.Type == enumor.SecurityGroupRuleType(sgRule.Direction) &&
			one.CloudProjectID == sgRule.ProjectId {
			continue
		}
		rule := protocloud.HuaWeiSGRuleBatchUpdate{
			ID:                        one.ID,
			CloudID:                   sgRule.Id,
			Memo:                      &sgRule.Description,
			Protocol:                  sgRule.Protocol,
			Ethertype:                 sgRule.Ethertype,
			CloudRemoteGroupID:        sgRule.RemoteGroupId,
			RemoteIPPrefix:            sgRule.RemoteIpPrefix,
			CloudRemoteAddressGroupID: sgRule.RemoteAddressGroupId,
			Port:                      sgRule.Multiport,
			Priority:                  int64(sgRule.Priority),
			Action:                    sgRule.Action,
			Type:                      enumor.SecurityGroupRuleType(sgRule.Direction),
			CloudSecurityGroupID:      id,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 req.AccountID,
			Region:                    req.Region,
			SecurityGroupID:           sgID,
		}
		list = append(list, rule)
	}

	return list
}

// getHuaWeiSGRuleByCid get HuaWeiSecurityGroupRule
func (g *securityGroup) getHuaWeiSGRuleByCid(cts *rest.Contexts, cID string, sgID string) (*corecloud.HuaWeiSecurityGroupRule, error) {

	listReq := &protocloud.HuaWeiSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, id: %s, err: %v, rid: %s", cID, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}

// diffHuaWeiSGRuleSyncDelete delete huawei security group rule.
func (g *securityGroup) diffHuaWeiSGRuleSyncDelete(cts *rest.Contexts, deleteCloudIDs []string,
	dsMap map[string]*proto.SecurityGroupSyncDS) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.HuaWeiSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_security_group_id", id),
		}
		err := g.dataCli.HuaWei.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, dsMap[id].HcSecurityGroup.ID)
		if err != nil {
			logs.Errorf("dataservice delete huawei security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// SyncHuaWeiSGRule sync huawei security group rules.
func (g *securityGroup) SyncHuaWeiSGRule(cts *rest.Contexts) (interface{}, error) {

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &types.HuaWeiSGRuleListOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(*rules.SecurityGroupRules) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*HuaWeiSGRuleSync)
	cloudIDs := make([]string, 0, len(*rules.SecurityGroupRules))
	for _, rule := range *rules.SecurityGroupRules {
		sgRuleSync := new(HuaWeiSGRuleSync)
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		cloudMap[rule.Id] = sgRuleSync
		cloudIDs = append(cloudIDs, rule.Id)
		cloudAllIDs[rule.Id] = true
	}

	updateIDs, err := g.getHuaWeiSGRuleDSSync(cloudIDs, req, cts, sgID)
	if err != nil {
		logs.Errorf("request getHuaWeiSGRuleDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := g.syncHuaWeiSGRuleUpdate(updateIDs, cloudMap, sgID, cts, req)
		if err != nil {
			logs.Errorf("request syncHuaWeiSGRuleUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		err := g.syncHuaWeiSGRuleAdd(addIDs, cts, req, cloudMap, sgID)
		if err != nil {
			logs.Errorf("request syncHuaWeiSGRuleAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := g.getHuaWeiSGRuleAllDS(req, cts, sgID)
	if err != nil {
		logs.Errorf("request getHuaWeiSGRuleAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
			for _, rule := range *rules.SecurityGroupRules {
				if rule.Id == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		err = g.syncHuaWeiSGRuleDelete(cts, realDeleteIDs, sgID)
		if err != nil {
			logs.Errorf("request syncHuaWeiSGRuleDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (g *securityGroup) syncHuaWeiSGRuleUpdate(updateIDs []string, cloudMap map[string]*HuaWeiSGRuleSync, sgID string,
	cts *rest.Contexts, req *proto.SecurityGroupSyncReq) error {

	rulesResp := new(model.ListSecurityGroupRulesResponse)
	rules := make([]model.SecurityGroupRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}
	rulesResp.SecurityGroupRules = &rules

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := g.genHuaWeiUpdateRulesList(rulesResp, cts, sgID, sg.CloudID, req)
	updateReq := &protocloud.HuaWeiSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := g.dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncHuaWeiSGRuleAdd(addIDs []string, cts *rest.Contexts, req *proto.SecurityGroupSyncReq,
	cloudMap map[string]*HuaWeiSGRuleSync, sgID string) error {

	rulesResp := new(model.ListSecurityGroupRulesResponse)
	rules := make([]model.SecurityGroupRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}
	rulesResp.SecurityGroupRules = &rules

	sg, err := g.dataCli.HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := genHuaWeiRuleList(rulesResp, req, sg.CloudID, sgID)
	createReq := &protocloud.HuaWeiSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := g.dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncHuaWeiSGRuleDelete(cts *rest.Contexts, deleteCloudIDs []string, sgID string) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.HuaWeiSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := g.dataCli.HuaWei.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete huawei security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (g *securityGroup) getHuaWeiSGRuleAllDS(req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &protocloud.HuaWeiSGRuleListReq{
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

		results, err := g.dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
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

func (g *securityGroup) getHuaWeiSGRuleDSSync(cloudIDs []string, req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	updateIDs := make([]string, 0)

	start := 0
	for {

		dataReq := &protocloud.HuaWeiSGRuleListReq{
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

		results, err := g.dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
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
