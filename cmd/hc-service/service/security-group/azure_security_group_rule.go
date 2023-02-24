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
	"strings"

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
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

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

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}

	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygrouprule.AzureCreateOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
	}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]securitygrouprule.AzureCreate, 0, len(req.EgressRuleSet))

		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, securitygrouprule.AzureCreate{
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
		opt.IngressRuleSet = make([]securitygrouprule.AzureCreate, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, securitygrouprule.AzureCreate{
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

	list := make([]protocloud.AzureSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		spec := protocloud.AzureSGRuleBatchCreate{
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
			CloudSecurityGroupID:       sg.CloudID,
			AccountID:                  req.AccountID,
			Region:                     sg.Region,
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
	result, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		createReq, sgID)
	if err != nil {
		return nil, err
	}

	return result, nil
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
			CloudID:                          rule.CloudID,
			Name:                             req.Name,
			Description:                      req.Memo,
			DestinationAddressPrefix:         req.DestinationAddressPrefix,
			DestinationAddressPrefixes:       req.DestinationAddressPrefixes,
			CloudDestinationSecurityGroupIDs: req.CloudDestinationSecurityGroupIDs,
			DestinationPortRange:             req.DestinationPortRange,
			DestinationPortRanges:            req.DestinationPortRanges,
			Protocol:                         req.Protocol,
			SourceAddressPrefix:              req.SourceAddressPrefix,
			SourceAddressPrefixes:            req.SourceAddressPrefixes,
			CloudSourceSecurityGroupIDs:      req.CloudSourceSecurityGroupIDs,
			SourcePortRange:                  req.SourcePortRange,
			SourcePortRanges:                 req.SourcePortRanges,
			Priority:                         req.Priority,
			Access:                           req.Access,
		},
	}
	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update azure security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	index := strings.LastIndex(rule.CloudID, rule.Name)
	cloudID := rule.CloudID[0:index] + req.Name
	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: []protocloud.AzureSGRuleUpdate{
			{
				ID:                         id,
				CloudID:                    cloudID,
				Name:                       req.Name,
				Memo:                       req.Memo,
				DestinationAddressPrefix:   req.DestinationAddressPrefix,
				DestinationAddressPrefixes: req.DestinationAddressPrefixes,
				DestinationPortRange:       req.DestinationPortRange,
				DestinationPortRanges:      req.DestinationPortRanges,
				Protocol:                   req.Protocol,
				SourceAddressPrefix:        req.SourceAddressPrefix,
				SourceAddressPrefixes:      req.SourceAddressPrefixes,
				SourcePortRange:            req.SourcePortRange,
				SourcePortRanges:           req.SourcePortRanges,
				Priority:                   req.Priority,
				Access:                     req.Access,
				CloudSecurityGroupID:       sg.CloudID,
				AccountID:                  sg.AccountID,
				Region:                     sg.Region,
				SecurityGroupID:            sg.ID,
			},
		},
	}
	err = g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *securityGroup) getAzureSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	AzureSecurityGroupRule, error) {

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, id: %s, err: %v, rid: %s", id, err,
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

	deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = g.dataCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// diffAzureSGRuleSyncAdd add azure security group rule.
func (g *securityGroup) diffAzureSGRuleSyncAdd(cts *rest.Contexts, ids []string,
	req *proto.SecurityGroupSyncReq) error {

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			logs.Errorf("request dataservice get azure security group failed, id: %s, err: %v, rid: %s", id, err,
				cts.Kit.Rid)
			return err
		}

		opt := &securitygrouprule.AzureListOption{
			Region:               req.Region,
			ResourceGroupName:    req.ResourceGroupName,
			CloudSecurityGroupID: sg.CloudID,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := genAzureRulesList(rules, sg.CloudID, id, req)
		createReq := &protocloud.AzureSGRuleCreateReq{
			Rules: list,
		}
		if len(createReq.Rules) <= 0 {
			continue
		}
		_, err = g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			createReq, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// genAzureRulesList gen AzureSGRuleBatchCreate list
func genAzureRulesList(rules []*armnetwork.SecurityRule, sgCloudID string,
	id string, req *proto.SecurityGroupSyncReq) []protocloud.AzureSGRuleBatchCreate {

	list := make([]protocloud.AzureSGRuleBatchCreate, 0, len(rules))

	for _, rule := range rules {
		spec := protocloud.AzureSGRuleBatchCreate{
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
			CloudSecurityGroupID:       sgCloudID,
			AccountID:                  req.AccountID,
			Region:                     req.Region,
			SecurityGroupID:            id,
		}
		switch *rule.Properties.Direction {
		case armnetwork.SecurityRuleDirectionInbound:
			spec.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			spec.Type = enumor.Egress
		default:
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

	return list
}

// diffAzureSGRuleSyncUpdate update huawei security group rule.
func (g *securityGroup) diffAzureSGRuleSyncUpdate(cts *rest.Contexts, updateCloudIDs []string,
	req *proto.SecurityGroupSyncReq, dsMap map[string]*proto.SecurityGroupSyncDS) error {

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	for _, id := range updateCloudIDs {
		sgID := dsMap[id].HcSecurityGroup.ID
		opt := &securitygrouprule.AzureListOption{
			Region:               req.Region,
			ResourceGroupName:    req.ResourceGroupName,
			CloudSecurityGroupID: id,
		}
		rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}

		list := g.genAzureUpdateRulesList(rules, sgID, cts, id, req)
		createReq := &protocloud.AzureSGRuleBatchUpdateReq{
			Rules: list,
		}

		if len(createReq.Rules) <= 0 {
			continue
		}
		err = g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

// genHuaWeiUpdateRulesList gen AzureUpdate list
func (g *securityGroup) genAzureUpdateRulesList(rules []*armnetwork.SecurityRule, sgID string,
	cts *rest.Contexts, id string, req *proto.SecurityGroupSyncReq) []protocloud.AzureSGRuleUpdate {

	list := make([]protocloud.AzureSGRuleUpdate, 0, len(rules))

	for _, rule := range rules {
		one, err := g.getAzureSGRuleByCid(cts, *rule.ID, sgID)
		if err != nil || one == nil {
			logs.Errorf("azure gen update RulesList getAzureSGRuleByCid failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}
		spec := protocloud.AzureSGRuleUpdate{
			ID:                         one.ID,
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
			CloudSecurityGroupID:       id,
			AccountID:                  req.AccountID,
			Region:                     req.Region,
			SecurityGroupID:            sgID,
		}
		switch *rule.Properties.Direction {
		case armnetwork.SecurityRuleDirectionInbound:
			spec.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			spec.Type = enumor.Egress
		default:
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

	return list
}

// getAzureSGRuleByCid
func (g *securityGroup) getAzureSGRuleByCid(cts *rest.Contexts, cID string, sgID string) (*corecloud.
	AzureSecurityGroupRule, error) {

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, id: %s, err: %v, rid: %s", cID, err,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}

// diffAzureSGRuleSyncDelete delete azure security group rule.
func (g *securityGroup) diffAzureSGRuleSyncDelete(cts *rest.Contexts, deleteCloudIDs []string,
	dsMap map[string]*proto.SecurityGroupSyncDS) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_security_group_id", id),
		}
		err := g.dataCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, dsMap[id].HcSecurityGroup.ID)
		if err != nil {
			logs.Errorf("dataservice delete azure security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// SyncAzureSGRule sync zure security group rules.
func (g *securityGroup) SyncAzureSGRule(cts *rest.Contexts) (interface{}, error) {

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req, err := g.decodeSecurityGroupSyncReq(cts)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &securitygrouprule.AzureListOption{
		Region:               req.Region,
		ResourceGroupName:    req.ResourceGroupName,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(rules) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*AzureSGRuleSync)
	cloudIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		sgRuleSync := new(AzureSGRuleSync)
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		cloudMap[*rule.ID] = sgRuleSync
		cloudIDs = append(cloudIDs, *rule.ID)
		cloudAllIDs[*rule.ID] = true
	}

	updateIDs, err := g.getAzureSGRuleDSSync(cloudIDs, req, cts, sgID)
	if err != nil {
		logs.Errorf("request getAzureSGRuleDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := g.syncAzureSGRuleUpdate(updateIDs, cloudMap, sgID, cts, req)
		if err != nil {
			logs.Errorf("request syncAzureSGRuleUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		err := g.syncAzureSGRuleAdd(addIDs, cts, req, cloudMap, sgID)
		if err != nil {
			logs.Errorf("request syncAzureSGRuleAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := g.getAzureSGRuleAllDS(req, cts, sgID)
	if err != nil {
		logs.Errorf("request getAzureSGRuleAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
				if *rule.ID == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		err = g.syncAzureSGRuleDelete(cts, realDeleteIDs, sgID)
		if err != nil {
			logs.Errorf("request syncAzureSGRuleDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (g *securityGroup) syncAzureSGRuleUpdate(updateIDs []string, cloudMap map[string]*AzureSGRuleSync, sgID string,
	cts *rest.Contexts, req *proto.SecurityGroupSyncReq) error {

	rules := make([]*armnetwork.SecurityRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := g.genAzureUpdateRulesList(rules, sgID, cts, sg.CloudID, req)
	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncAzureSGRuleAdd(addIDs []string, cts *rest.Contexts, req *proto.SecurityGroupSyncReq,
	cloudMap map[string]*AzureSGRuleSync, sgID string) error {

	rules := make([]*armnetwork.SecurityRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	list := genAzureRulesList(rules, sg.CloudID, sgID, req)
	createReq := &protocloud.AzureSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *securityGroup) syncAzureSGRuleDelete(cts *rest.Contexts, deleteCloudIDs []string, sgID string) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := g.dataCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete azure security group rules failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (g *securityGroup) getAzureSGRuleAllDS(req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &protocloud.AzureSGRuleListReq{
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

		results, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
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

func (g *securityGroup) getAzureSGRuleDSSync(cloudIDs []string, req *proto.SecurityGroupSyncReq,
	cts *rest.Contexts, sgID string) ([]string, error) {

	updateIDs := make([]string, 0)

	start := 0
	for {

		dataReq := &protocloud.AzureSGRuleListReq{
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

		results, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), dataReq, sgID)
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
