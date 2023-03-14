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

	"github.com/aws/aws-sdk-go/service/ec2"
)

// SyncAwsSGRule sync aws security group rules.
func SyncAwsSGRule(kt *kit.Kit, req *SyncAwsSecurityGroupOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client, sgID string) (interface{}, error) {

	client, err := ad.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	sg, err := dataCli.Aws.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &securitygrouprule.AwsListOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rules) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*AwsSGRuleSync)
	cloudIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		sgRuleSync := new(AwsSGRuleSync)
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		cloudMap[*rule.SecurityGroupRuleId] = sgRuleSync
		cloudIDs = append(cloudIDs, *rule.SecurityGroupRuleId)
		cloudAllIDs[*rule.SecurityGroupRuleId] = true
	}

	updateIDs, err := getAwsSGRuleDSSync(kt, cloudIDs, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("request getAwsSGRuleDSSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := syncAwsSGRuleUpdate(kt, updateIDs, cloudMap, sgID, req, dataCli)
		if err != nil {
			logs.Errorf("request syncAwsSGRuleUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
		err := syncAwsSGRuleAdd(kt, addIDs, req, cloudMap, sgID, dataCli)
		if err != nil {
			logs.Errorf("request syncAwsSGRuleAdd failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	dsIDs, err := getAwsSGRuleAllDS(kt, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("request getAwsSGRuleAllDS failed, err: %v, rid: %s", err, kt.Rid)
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
				if *rule.SecurityGroupRuleId == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		if len(realDeleteIDs) > 0 {
			err := syncAwsSGRuleDelete(kt, realDeleteIDs, sgID, dataCli)
			if err != nil {
				logs.Errorf("request syncAwsSGRuleDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func syncAwsSGRuleUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*AwsSGRuleSync, sgID string,
	req *SyncAwsSecurityGroupOption, dataCli *dataservice.Client) error {

	rules := make([]*ec2.SecurityGroupRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	list := genAwsUpdateRulesList(kt, rules, req, sgID, dataCli)
	updateReq := &protocloud.AwsSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncAwsSGRuleAdd(kt *kit.Kit, addIDs []string, req *SyncAwsSecurityGroupOption,
	cloudMap map[string]*AwsSGRuleSync, sgID string, dataCli *dataservice.Client) error {

	rules := make([]*ec2.SecurityGroupRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}

	sg, err := dataCli.Aws.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	list := genAwsAddRulesList(rules, req, sg.ID)
	createReq := &protocloud.AwsSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := dataCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncAwsSGRuleDelete(kt *kit.Kit, deleteCloudIDs []string, sgID string,
	dataCli *dataservice.Client) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.AwsSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := dataCli.Aws.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("dataservice delete aws security group rules failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getAwsSGRuleAllDS(kt *kit.Kit, req *SyncAwsSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &protocloud.AwsSGRuleListReq{
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

		results, err := dataCli.Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
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

func getAwsSGRuleDSSync(kt *kit.Kit, cloudIDs []string, req *SyncAwsSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) ([]string, error) {

	updateIDs := make([]string, 0)

	start := 0
	for {

		dataReq := &protocloud.AwsSGRuleListReq{
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

		results, err := dataCli.Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
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

func isAwsSGRuleChange(db *corecloud.AwsSecurityGroupRule, cloud *ec2.SecurityGroupRule) bool {

	if db.CloudID != *cloud.SecurityGroupRuleId {
		return true
	}

	if db.IPv4Cidr != cloud.CidrIpv4 {
		return true
	}

	if db.IPv6Cidr != cloud.CidrIpv6 {
		return true
	}

	if db.Memo != cloud.Description {
		return true
	}

	if db.FromPort != *cloud.FromPort {
		return true
	}

	if db.ToPort != *cloud.ToPort {
		return true
	}

	if db.Protocol != cloud.IpProtocol {
		return true
	}

	if db.CloudPrefixListID != cloud.PrefixListId {
		return true
	}

	if db.CloudSecurityGroupID != *cloud.GroupId {
		return true
	}

	if db.CloudGroupOwnerID != *cloud.GroupOwnerId {
		return true
	}

	if cloud.ReferencedGroupInfo != nil {
		if db.CloudTargetSecurityGroupID != cloud.ReferencedGroupInfo.GroupId {
			return true
		}
	}

	return false
}

func genAwsUpdateRulesList(kt *kit.Kit, rules []*ec2.SecurityGroupRule, req *SyncAwsSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) []protocloud.AwsSGRuleUpdate {

	list := make([]protocloud.AwsSGRuleUpdate, 0, len(rules))

	for _, rule := range rules {
		cOne, err := getAwsSGRuleByCid(kt, *rule.SecurityGroupRuleId, sgID, dataCli)
		if err != nil || cOne == nil {
			logs.Errorf("aws gen update RulesList getAwsSGRuleByCid failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		if !isAwsSGRuleChange(cOne, rule) {
			continue
		}

		one := protocloud.AwsSGRuleUpdate{
			ID:                   cOne.ID,
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             *rule.FromPort,
			ToPort:               *rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            req.AccountID,
			Region:               req.Region,
			SecurityGroupID:      sgID,
		}

		if *rule.IsEgress {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}

		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}

		list = append(list, one)
	}

	return list
}

func getAwsSGRuleByCid(kt *kit.Kit, cID string, sgID string,
	dataCli *dataservice.Client) (*corecloud.AwsSecurityGroupRule, error) {

	listReq := &protocloud.AwsSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := dataCli.Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, id: %s, err: %v, rid: %s", cID, err, kt.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}

func genAwsAddRulesList(rules []*ec2.SecurityGroupRule, req *SyncAwsSecurityGroupOption,
	id string) []protocloud.AwsSGRuleBatchCreate {

	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))

	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              *rule.SecurityGroupRuleId,
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             *rule.FromPort,
			ToPort:               *rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: *rule.GroupId,
			CloudGroupOwnerID:    *rule.GroupOwnerId,
			AccountID:            req.AccountID,
			Region:               req.Region,
			SecurityGroupID:      id,
		}

		if *rule.IsEgress {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}

		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}

		list = append(list, one)
	}

	return list
}
