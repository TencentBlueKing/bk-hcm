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
	"strings"

	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateAzureSGRule update azure security group rule.
func (g *securityGroup) UpdateAzureSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcservice.AzureSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	rule, err := g.getAzureSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AzureUpdateOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		Rule: &securitygrouprule.AzureUpdate{
			CloudID:                             rule.CloudID,
			Name:                                req.Name,
			Description:                         req.Memo,
			DestinationAddressPrefix:            req.DestinationAddressPrefix,
			DestinationAddressPrefixes:          req.DestinationAddressPrefixes,
			CloudDestinationAppSecurityGroupIDs: nil,
			DestinationPortRange:                req.DestinationPortRange,
			DestinationPortRanges:               req.DestinationPortRanges,
			Protocol:                            req.Protocol,
			SourceAddressPrefix:                 req.SourceAddressPrefix,
			SourceAddressPrefixes:               req.SourceAddressPrefixes,
			CloudSourceAppSecurityGroupIDs:      nil,
			SourcePortRange:                     req.SourcePortRange,
			SourcePortRanges:                    req.SourcePortRanges,
			Priority:                            req.Priority,
			Access:                              req.Access,
		},
	}
	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update azure security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	if err = g.updateAzureDSRule(cts, rule, req, sg); err != nil {
		return nil, err
	}

	return nil, nil
}

// updateAzureDSRule update azure security group rule in data service.
func (g *securityGroup) updateAzureDSRule(cts *rest.Contexts, rule *corecloud.AzureSecurityGroupRule,
	req *hcservice.AzureSGRuleUpdateReq, sg *corecloud.SecurityGroup[corecloud.AzureSecurityGroupExtension]) error {

	index := strings.LastIndex(rule.CloudID, rule.Name)
	cloudID := rule.CloudID[0:index] + req.Name
	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: []protocloud.AzureSGRuleUpdate{
			{
				ID:                                  rule.ID,
				CloudID:                             cloudID,
				Name:                                req.Name,
				Memo:                                req.Memo,
				DestinationAddressPrefix:            req.DestinationAddressPrefix,
				DestinationAddressPrefixes:          req.DestinationAddressPrefixes,
				DestinationPortRange:                req.DestinationPortRange,
				DestinationPortRanges:               req.DestinationPortRanges,
				CloudSourceAppSecurityGroupIDs:      nil,
				CloudDestinationAppSecurityGroupIDs: nil,
				Protocol:                            req.Protocol,
				SourceAddressPrefix:                 req.SourceAddressPrefix,
				SourceAddressPrefixes:               req.SourceAddressPrefixes,
				SourcePortRange:                     req.SourcePortRange,
				SourcePortRanges:                    req.SourcePortRanges,
				Priority:                            req.Priority,
				Access:                              req.Access,
				CloudSecurityGroupID:                sg.CloudID,
				AccountID:                           sg.AccountID,
				Region:                              sg.Region,
				SecurityGroupID:                     sg.ID,
			},
		},
	}
	err := g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sg.ID)
	if err != nil {
		logs.Errorf("call dataservice to BatchUpdateSecurityGroupRule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

// DeleteAzureSGRule delete azure security group rule.
func (g *securityGroup) DeleteAzureSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	rule, err := g.getAzureSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AzureDeleteOption{
		Region:               rule.Region,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		CloudRuleID:          rule.CloudID,
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	// 云上删除成功后，删除data-service中的记录
	deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = g.dataCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
