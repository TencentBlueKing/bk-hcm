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
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	apicore "hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SyncAzureSGRule sync azure security group rules.
func SyncAzureSGRule(kt *kit.Kit, req *SyncAzureSecurityGroupOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client, sgID string) (interface{}, error) {

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := dataCli.Azure.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, kt.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &securitygrouprule.AzureListOption{
		Region:               req.Region,
		ResourceGroupName:    req.ResourceGroupName,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure security group rule failed, err: %v, rid: %s", err, kt.Rid)
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

	updateIDs, err := getAzureSGRuleDSSync(kt, cloudIDs, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("request getAzureSGRuleDSSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := syncAzureSGRuleUpdate(kt, updateIDs, cloudMap, sgID, req, dataCli)
		if err != nil {
			logs.Errorf("request syncAzureSGRuleUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
		err := syncAzureSGRuleAdd(kt, addIDs, req, cloudMap, sgID, dataCli)
		if err != nil {
			logs.Errorf("request syncAzureSGRuleAdd failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	dsIDs, err := getAzureSGRuleAllDS(kt, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("request getAzureSGRuleAllDS failed, err: %v, rid: %s", err, kt.Rid)
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
		rules, err := client.ListSecurityGroupRule(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, kt.Rid)
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

		if len(realDeleteIDs) > 0 {
			err := syncAzureSGRuleDelete(kt, realDeleteIDs, sgID, dataCli)
			if err != nil {
				logs.Errorf("request syncAzureSGRuleDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func syncAzureSGRuleUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*AzureSGRuleSync,
	sgID string, req *SyncAzureSecurityGroupOption, dataCli *dataservice.Client) error {

	rules := make([]*armnetwork.SecurityRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := dataCli.Azure.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	list := genAzureUpdateRulesList(kt, rules, sgID, sg.CloudID, req, dataCli)
	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncAzureSGRuleAdd(kt *kit.Kit, addIDs []string, req *SyncAzureSecurityGroupOption,
	cloudMap map[string]*AzureSGRuleSync, sgID string, dataCli *dataservice.Client) error {

	rules := make([]*armnetwork.SecurityRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := dataCli.Azure.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	list := genAzureAddRulesList(rules, sg.CloudID, sgID, req)
	createReq := &protocloud.AzureSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := dataCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncAzureSGRuleDelete(kt *kit.Kit, deleteCloudIDs []string,
	sgID string, dataCli *dataservice.Client) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := dataCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete azure security group rules failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getAzureSGRuleAllDS(kt *kit.Kit, req *SyncAzureSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) ([]string, error) {

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

		results, err := dataCli.Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, kt.Rid)
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

func getAzureSGRuleDSSync(kt *kit.Kit, cloudIDs []string, req *SyncAzureSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) ([]string, error) {

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

		results, err := dataCli.Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("from data-service list sg rule failed, err: %v, rid: %s", err, kt.Rid)
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

func genAzureUpdateRulesList(kt *kit.Kit, rules []*armnetwork.SecurityRule, sgID string,
	id string, req *SyncAzureSecurityGroupOption, dataCli *dataservice.Client) []protocloud.AzureSGRuleUpdate {

	list := make([]protocloud.AzureSGRuleUpdate, 0, len(rules))

	for _, rule := range rules {
		one, err := getAzureSGRuleByCid(kt, *rule.ID, sgID, dataCli)
		if err != nil || one == nil {
			logs.Errorf("azure gen update RulesList getAzureSGRuleByCid failed, err: %v, rid: %s", err, kt.Rid)
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

func getAzureSGRuleByCid(kt *kit.Kit, cID string, sgID string,
	dataCli *dataservice.Client) (*corecloud.AzureSecurityGroupRule, error) {

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := dataCli.Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, id: %s, err: %v, rid: %s", cID, err, kt.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}

func genAzureAddRulesList(rules []*armnetwork.SecurityRule, sgCloudID string,
	id string, req *SyncAzureSecurityGroupOption) []protocloud.AzureSGRuleBatchCreate {

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
