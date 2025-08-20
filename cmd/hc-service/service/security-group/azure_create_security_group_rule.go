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
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
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
			opt.EgressRuleSet = append(opt.EgressRuleSet, convAzureRule(rule))
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]securitygrouprule.AzureCreate, 0, len(req.IngressRuleSet))

		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, convAzureRule(rule))
		}
	}
	rules, err := client.CreateSecurityGroupRule(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create azure security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 云上创建成功后，批量创建data-service中的记录
	createReq := convAzureDSReq(rules, sg, req)
	result, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		createReq, sgID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func convAzureDSReq(rules []*securitygrouprule.AzureSGRule,
	sg *corecloud.SecurityGroup[corecloud.AzureSecurityGroupExtension],
	req *hcservice.AzureSGRuleCreateReq) *protocloud.AzureSGRuleCreateReq {

	list := make([]protocloud.AzureSGRuleBatchCreate, 0, len(rules))
	for _, rule := range rules {
		spec := protocloud.AzureSGRuleBatchCreate{
			CloudID:                             *rule.ID,
			Etag:                                rule.Etag,
			Name:                                *rule.Name,
			Memo:                                rule.Description,
			DestinationAddressPrefix:            rule.DestinationAddressPrefix,
			DestinationAddressPrefixes:          rule.DestinationAddressPrefixes,
			DestinationPortRange:                rule.DestinationPortRange,
			CloudDestinationAppSecurityGroupIDs: nil,
			CloudSourceAppSecurityGroupIDs:      nil,
			DestinationPortRanges:               rule.DestinationPortRanges,
			Protocol:                            string(*rule.Protocol),
			ProvisioningState:                   string(*rule.ProvisioningState),
			SourceAddressPrefix:                 rule.SourceAddressPrefix,
			SourceAddressPrefixes:               rule.SourceAddressPrefixes,
			SourcePortRange:                     rule.SourcePortRange,
			SourcePortRanges:                    rule.SourcePortRanges,
			Priority:                            *rule.Priority,
			Access:                              string(*rule.Access),
			CloudSecurityGroupID:                sg.CloudID,
			AccountID:                           req.AccountID,
			Region:                              sg.Region,
			SecurityGroupID:                     sg.ID,
		}

		switch *rule.Direction {
		case armnetwork.SecurityRuleDirectionInbound:
			spec.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			spec.Type = enumor.Egress
		}

		if len(rule.DestinationApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(rule.DestinationApplicationSecurityGroups))
			for _, one := range rule.DestinationApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			spec.CloudDestinationAppSecurityGroupIDs = ids
		}

		if len(rule.SourceApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(rule.SourceApplicationSecurityGroups))
			for _, one := range rule.SourceApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			spec.CloudSourceAppSecurityGroupIDs = ids
		}

		list = append(list, spec)
	}

	createReq := &protocloud.AzureSGRuleCreateReq{
		Rules: list,
	}
	return createReq
}

func convAzureRule(rule hcservice.AzureSGRuleCreate) securitygrouprule.AzureCreate {
	return securitygrouprule.AzureCreate{
		Name:                                rule.Name,
		Description:                         rule.Memo,
		DestinationAddressPrefix:            rule.DestinationAddressPrefix,
		DestinationAddressPrefixes:          rule.DestinationAddressPrefixes,
		CloudDestinationAppSecurityGroupIDs: nil,
		DestinationPortRange:                rule.DestinationPortRange,
		DestinationPortRanges:               rule.DestinationPortRanges,
		Protocol:                            rule.Protocol,
		SourceAddressPrefix:                 rule.SourceAddressPrefix,
		SourceAddressPrefixes:               rule.SourceAddressPrefixes,
		CloudSourceAppSecurityGroupIDs:      nil,
		SourcePortRange:                     rule.SourcePortRange,
		SourcePortRanges:                    rule.SourcePortRanges,
		Priority:                            rule.Priority,
		Type:                                rule.Type,
		Access:                              rule.Access,
	}
}
