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
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// BatchCreateAzureSGRule batch create azure security group rule.
func (g *securityGroup) BatchCreateAzureSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.AzureSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.SecurityGroup().GetAzureSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.Spec.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.Azure(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSGRuleCreateOption{
		Region:               sg.Spec.Region,
		CloudSecurityGroupID: sg.Spec.CloudID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]types.AzureSGRuleCreate, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, types.AzureSGRuleCreate{
				Name:                             rule.Name,
				Description:                      rule.Memo,
				DestinationAddressPrefix:         rule.DestinationAddressPrefix,
				DestinationAddressPrefixes:       rule.DestinationAddressPrefixes,
				CloudDestinationSecurityGroupIDs: rule.CloudDestinationSecurityGroupIDs,
				DestinationPortRange:             rule.DestinationPortRange,
				DestinationPortRanges:            rule.DestinationPortRanges,
				Protocol:                         rule.Protocol,
				SourceAddressPrefix:              rule.SourceAddressPrefix,
				SourceAddressPrefixes:            rule.SourceAddressPrefixes,
				CloudSourceSecurityGroupIDs:      rule.CloudSourceSecurityGroupIDs,
				SourcePortRange:                  rule.SourcePortRange,
				SourcePortRanges:                 rule.SourcePortRanges,
				Priority:                         rule.Priority,
				Type:                             rule.Type,
				Access:                           rule.Access,
			})
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]types.AzureSGRuleCreate, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, types.AzureSGRuleCreate{
				Name:                             rule.Name,
				Description:                      rule.Memo,
				DestinationAddressPrefix:         rule.DestinationAddressPrefix,
				DestinationAddressPrefixes:       rule.DestinationAddressPrefixes,
				CloudDestinationSecurityGroupIDs: rule.CloudDestinationSecurityGroupIDs,
				DestinationPortRange:             rule.DestinationPortRange,
				DestinationPortRanges:            rule.DestinationPortRanges,
				Protocol:                         rule.Protocol,
				SourceAddressPrefix:              rule.SourceAddressPrefix,
				SourceAddressPrefixes:            rule.SourceAddressPrefixes,
				CloudSourceSecurityGroupIDs:      rule.CloudSourceSecurityGroupIDs,
				SourcePortRange:                  rule.SourcePortRange,
				SourcePortRanges:                 rule.SourcePortRanges,
				Priority:                         rule.Priority,
				Type:                             rule.Type,
				Access:                           rule.Access,
			})
		}
	}
	rules, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	list := make([]corecloud.AzureSecurityGroupRuleSpec, 0, len(rules))
	for _, rule := range rules {
		spec := corecloud.AzureSecurityGroupRuleSpec{
			CloudID:                    *rule.ID,
			Etag:                       rule.Etag,
			Name:                       *rule.Name,
			Memo:                       rule.Properties.Description,
			DestinationAddressPrefix:   rule.Properties.DestinationAddressPrefix,
			DestinationAddressPrefixes: rule.Properties.DestinationAddressPrefixes,
			DestinationPortRange:       rule.Properties.DestinationPortRange,
			DestinationPortRanges:      rule.Properties.DestinationPortRanges,
			Protocol:                   string(*rule.Properties.Protocol),
			ProvisioningState:          string(*rule.Properties.ProvisioningState),
			SourceAddressPrefix:        rule.Properties.SourceAddressPrefix,
			SourceAddressPrefixes:      rule.Properties.SourceAddressPrefixes,
			SourcePortRange:            rule.Properties.SourcePortRange,
			SourcePortRanges:           rule.Properties.SourcePortRanges,
			Priority:                   *rule.Properties.Priority,
			Access:                     string(*rule.Properties.Access),
			CloudSecurityGroupID:       sg.Spec.CloudID,
			AccountID:                  req.AccountID,
			Region:                     sg.Spec.Region,
			SecurityGroupID:            sg.ID,
		}

		switch *rule.Properties.Direction {
		case armnetwork.SecurityRuleDirectionInbound:
			spec.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			spec.Type = enumor.Egress
		default:
			return nil, fmt.Errorf("unknown security group rule direction: %s", *rule.Properties.Direction)
		}

		if len(rule.Properties.DestinationApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(rule.Properties.DestinationApplicationSecurityGroups))
			for _, one := range rule.Properties.DestinationApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			spec.CloudDestinationSecurityGroupIDs = ids
		}

		if len(rule.Properties.SourceApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(rule.Properties.SourceApplicationSecurityGroups))
			for _, one := range rule.Properties.SourceApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			spec.CloudSourceSecurityGroupIDs = ids
		}

		list = append(list, spec)
	}

	createReq := &protocloud.AzureSGRuleCreateReq{
		Rules: list,
	}
	ids, err := g.dataCli.SecurityGroup().BatchCreateAzureSGRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

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

	sg, err := g.dataCli.SecurityGroup().GetAzureSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	rule, err := g.getAzureSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, rule.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSGRuleUpdateOption{
		Region:               rule.Spec.Region,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		Rule: &types.AzureSGRuleUpdate{
			CloudID:                          rule.Spec.CloudID,
			Name:                             req.Spec.Name,
			Description:                      req.Spec.Memo,
			DestinationAddressPrefix:         req.Spec.DestinationAddressPrefix,
			DestinationAddressPrefixes:       req.Spec.DestinationAddressPrefixes,
			CloudDestinationSecurityGroupIDs: req.Spec.CloudDestinationSecurityGroupIDs,
			DestinationPortRange:             req.Spec.DestinationPortRange,
			DestinationPortRanges:            req.Spec.DestinationPortRanges,
			Protocol:                         req.Spec.Protocol,
			SourceAddressPrefix:              req.Spec.SourceAddressPrefix,
			SourceAddressPrefixes:            req.Spec.SourceAddressPrefixes,
			CloudSourceSecurityGroupIDs:      req.Spec.CloudSourceSecurityGroupIDs,
			SourcePortRange:                  req.Spec.SourcePortRange,
			SourcePortRanges:                 req.Spec.SourcePortRanges,
			Priority:                         req.Spec.Priority,
			Access:                           req.Spec.Access,
		},
	}
	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update azure security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: []protocloud.AzureSGRuleUpdate{
			{
				ID: id,
				Spec: &corecloud.AzureSecurityGroupRuleSpec{
					CloudID:                    rule.ID,
					Etag:                       rule.Spec.Etag,
					Name:                       req.Spec.Name,
					Memo:                       req.Spec.Memo,
					DestinationAddressPrefix:   req.Spec.DestinationAddressPrefix,
					DestinationAddressPrefixes: req.Spec.DestinationAddressPrefixes,
					DestinationPortRange:       req.Spec.DestinationPortRange,
					DestinationPortRanges:      req.Spec.DestinationPortRanges,
					Protocol:                   req.Spec.Protocol,
					SourceAddressPrefix:        req.Spec.SourceAddressPrefix,
					SourceAddressPrefixes:      req.Spec.SourceAddressPrefixes,
					SourcePortRange:            req.Spec.SourcePortRange,
					SourcePortRanges:           req.Spec.SourcePortRanges,
					Priority:                   req.Spec.Priority,
					Access:                     req.Spec.Access,
					CloudSecurityGroupID:       sg.Spec.CloudID,
					AccountID:                  sg.Spec.AccountID,
					Region:                     sg.Spec.Region,
					SecurityGroupID:            sg.ID,
				},
			},
		},
	}
	err = g.dataCli.SecurityGroup().BatchUpdateAzureSGRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getAzureSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	AzureSecurityGroupRule, error) {

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page: &daotypes.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	listResp, err := g.dataCli.SecurityGroup().ListAzureSGRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
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

	sg, err := g.dataCli.SecurityGroup().GetAzureSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	rule, err := g.getAzureSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, rule.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSGRuleDeleteOption{
		Region:               rule.Spec.Region,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		CloudSecurityGroupID: rule.Spec.CloudSecurityGroupID,
		CloudRuleID:          rule.Spec.CloudID,
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	deleteReq := &protocloud.AzureSGRuleDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = g.dataCli.SecurityGroup().DeleteAzureSGRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID); err != nil {
		return nil, err
	}

	return nil, nil
}
