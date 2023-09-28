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
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
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

	opt := &securitygrouprule.HuaWeiCreateOption{Region: sg.Region, CloudSecurityGroupID: sg.CloudID}
	if req.EgressRule != nil {
		opt.Rule = convertHuaweiCreateReq(req.IngressRule, enumor.Egress)
	}
	if req.IngressRule != nil {
		opt.Rule = convertHuaweiCreateReq(req.IngressRule, enumor.Ingress)
	}
	rule, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create huawei security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.HuaWeiSGRuleCreateReq{Rules: []protocloud.HuaWeiSGRuleBatchCreate{{
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
	}}}
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

func convertHuaweiCreateReq(sgRuleCreate *hcservice.HuaWeiSGRuleCreate,
	ruleType enumor.SecurityGroupRuleType) *securitygrouprule.HuaWeiCreate {

	return &securitygrouprule.HuaWeiCreate{
		Description:        sgRuleCreate.Memo,
		Ethertype:          sgRuleCreate.Ethertype,
		Protocol:           sgRuleCreate.Protocol,
		RemoteIPPrefix:     sgRuleCreate.RemoteIPPrefix,
		CloudRemoteGroupID: sgRuleCreate.CloudRemoteGroupID,
		Port:               sgRuleCreate.Port,
		Action:             sgRuleCreate.Action,
		Priority:           converter.ValToPtr(strconv.Itoa(int(sgRuleCreate.Priority))),
		Type:               ruleType,
	}
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

	opt := &securitygrouprule.HuaWeiDeleteOption{
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
		Page:   core.NewDefaultBasePage(),
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

// HuaWeiSGCount count huawei sg.
func (g *securityGroup) HuaWeiSGCount(cts *rest.Contexts) (interface{}, error) {
	req := new(corecloud.HuaWeiSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.Adaptor().HuaWei(&types.BaseSecret{
		CloudSecretID:  req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey,
	})
	if err != nil {
		return nil, err
	}

	return client.CountAllResources(cts.Kit, enumor.HuaWeiSGProviderType)
}
