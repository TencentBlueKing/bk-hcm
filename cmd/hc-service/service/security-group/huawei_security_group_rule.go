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
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
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

	sg, err := g.dataCli.SecurityGroup().GetHuaWeiSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.Spec.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSGRuleCreateOption{
		Region:               sg.Spec.Region,
		CloudSecurityGroupID: sg.Spec.CloudID,
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
		Rules: []corecloud.HuaWeiSecurityGroupRuleSpec{
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
				Type:                      opt.Rule.Type,
				CloudSecurityGroupID:      sg.Spec.CloudID,
				CloudProjectID:            rule.ProjectId,
				AccountID:                 req.AccountID,
				Region:                    sg.Spec.Region,
				SecurityGroupID:           sg.ID,
			},
		},
	}
	ids, err := g.dataCli.SecurityGroup().BatchCreateHuaWeiSGRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
	if err != nil {
		return nil, err
	}

	return ids, nil
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

	client, err := g.ad.HuaWei(cts.Kit, rule.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSGRuleDeleteOption{
		Region:      rule.Spec.Region,
		CloudRuleID: rule.Spec.CloudID,
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	deleteReq := &protocloud.HuaWeiSGRuleDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = g.dataCli.SecurityGroup().DeleteHuaWeiSGRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getHuaWeiSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	HuaWeiSecurityGroupRule, error) {

	listReq := &protocloud.HuaWeiSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page: &daotypes.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	listResp, err := g.dataCli.SecurityGroup().ListHuaWeiSGRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
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
