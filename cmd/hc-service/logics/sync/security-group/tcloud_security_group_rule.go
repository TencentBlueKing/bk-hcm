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
	"hcm/pkg/adaptor/tcloud"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type syncSecurityGroupRuleOption struct {
	Region               string
	CloudSecurityGroupID string
	SecurityGroupID      string
	AccountID            string
}

// SyncTCloudSGRule sync tcloud security group rules.
func SyncTCloudSGRule(kt *kit.Kit, client *tcloud.TCloud, dataCli *dataservice.Client,
	sgID string) ([]string, error) {

	sg, err := dataCli.TCloud.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get TCloud security group failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return nil, err
	}

	version, egressRuleMaps, ingressRuleMaps, err := listRuleFromCloud(kt, client, sg.Region, sg.CloudID)
	if err != nil {
		return nil, err
	}

	rulesFromDB, err := listRuleFromDB(kt, dataCli, sg.ID)
	if err != nil {
		return nil, err
	}

	updateRules := make(map[string]*corecloud.TCloudSecurityGroupRule)
	deleteRuleIDs := make([]string, 0)
	for _, one := range rulesFromDB {
		switch one.Type {
		case enumor.Egress:
			policy, exist := egressRuleMaps[one.CloudPolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
				continue
			}

			delete(egressRuleMaps, one.CloudPolicyIndex)

			if isSecurityGroupRuleChange(version, policy, one) {
				updateRules[one.ID] = convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Egress)
			}

		case enumor.Ingress:
			policy, exist := ingressRuleMaps[one.CloudPolicyIndex]
			if !exist {
				deleteRuleIDs = append(deleteRuleIDs, one.ID)
				continue
			}

			delete(ingressRuleMaps, one.CloudPolicyIndex)

			if isSecurityGroupRuleChange(version, policy, one) {
				updateRules[one.ID] = convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Ingress)
			}

		default:
			logs.Errorf("[%s] unknown security group rule type: %s, skip handle, rid: %s", enumor.TCloud,
				one.Type, kt.Rid)
		}
	}

	createRules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for _, policy := range egressRuleMaps {
		rule := convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Egress)
		createRules = append(createRules, *rule)
	}

	for _, policy := range ingressRuleMaps {
		rule := convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Ingress)
		createRules = append(createRules, *rule)
	}

	if len(updateRules) != 0 {
		if err = updateSecurityGroupRule(kt, dataCli, sg.ID, updateRules); err != nil {
			return nil, err
		}
	}

	if len(deleteRuleIDs) != 0 {
		if err = deleteSecurityGroupRule(kt, dataCli, sg.ID, deleteRuleIDs); err != nil {
			return nil, err
		}
	}

	if len(createRules) != 0 {
		ids, err := createSecurityGroupRule(kt, dataCli, sg.ID, createRules)
		if err != nil {
			return nil, err
		}

		return ids, nil
	}

	return make([]string, 0), nil
}

func listRuleFromDB(kt *kit.Kit, dataCli *dataservice.Client, sgID string) (
	[]corecloud.TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", sgID),
		Page:   core.DefaultBasePage,
	}
	start := uint32(0)
	rules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
			sgID)
		if err != nil {
			return nil, err
		}

		rules = append(rules, listResp.Details...)

		if len(listResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return rules, nil
}

func listRuleFromCloud(kt *kit.Kit, client *tcloud.TCloud, region, cloudSGID string) (string,
	map[int64]*vpc.SecurityGroupPolicy, map[int64]*vpc.SecurityGroupPolicy, error) {
	listOpt := &securitygrouprule.TCloudListOption{
		Region:               region,
		CloudSecurityGroupID: cloudSGID,
	}
	rules, err := client.ListSecurityGroupRule(kt, listOpt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list tcloud security group rule failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return "", nil, nil, err
	}

	egressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Egress))
	ingressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Ingress))
	for _, egress := range rules.Egress {
		egressRuleMaps[*egress.PolicyIndex] = egress
	}

	for _, ingress := range rules.Ingress {
		ingressRuleMaps[*ingress.PolicyIndex] = ingress
	}

	return *rules.Version, egressRuleMaps, ingressRuleMaps, nil
}

func updateSecurityGroupRule(kt *kit.Kit, dataCli *dataservice.Client, sgID string, updateRules map[string]*corecloud.
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
	if err := dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID); err != nil {
		logs.Errorf("[%s] request dataservice to batch update tcloud security group rule failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	return nil
}

func convTCloudRule(policy *vpc.SecurityGroupPolicy, sg *corecloud.BaseSecurityGroup, version string,
	ruleType enumor.SecurityGroupRuleType) *corecloud.TCloudSecurityGroupRule {

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
		Type:                       ruleType,
		CloudSecurityGroupID:       sg.CloudID,
		SecurityGroupID:            sg.ID,
		Region:                     sg.Region,
		AccountID:                  sg.AccountID,
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

func deleteSecurityGroupRule(kt *kit.Kit, dataCli *dataservice.Client, sgID string, delIDs []string) error {
	req := &protocloud.TCloudSGRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("id", delIDs),
	}
	err := dataCli.TCloud.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice to delete tcloud security group rule failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	return nil
}

func createSecurityGroupRule(kt *kit.Kit, dataCli *dataservice.Client, sgID string, rules []corecloud.
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
	result, err := dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud security group rule failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return nil, err
	}

	return result.IDs, nil
}

func isSecurityGroupRuleChange(version string, cloud *vpc.SecurityGroupPolicy,
	db corecloud.TCloudSecurityGroupRule) bool {

	if version != db.Version {
		return true
	}

	if cloud.PolicyIndex == nil || *cloud.PolicyIndex != db.CloudPolicyIndex {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Protocol, db.Protocol) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Port, db.Port) {
		return true
	}

	if cloud.ServiceTemplate != nil && (db.CloudServiceID != nil || db.CloudServiceGroupID != nil) {
		if !assert.IsPtrStringEqual(cloud.ServiceTemplate.ServiceId, db.CloudServiceID) {
			return true
		}

		if !assert.IsPtrStringEqual(cloud.ServiceTemplate.ServiceGroupId, db.CloudServiceGroupID) {
			return true
		}
	}

	if cloud.ServiceTemplate == nil && (db.CloudServiceID != nil || db.CloudServiceGroupID != nil) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CidrBlock, db.IPv4Cidr) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Ipv6CidrBlock, db.IPv6Cidr) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SecurityGroupId, db.CloudTargetSecurityGroupID) {
		return true
	}

	if cloud.AddressTemplate != nil && (db.CloudAddressID != nil || db.CloudAddressGroupID != nil) {
		if !assert.IsPtrStringEqual(cloud.AddressTemplate.AddressId, db.CloudAddressID) {
			return true
		}

		if !assert.IsPtrStringEqual(cloud.AddressTemplate.AddressGroupId, db.CloudAddressGroupID) {
			return true
		}
	}

	if cloud.AddressTemplate == nil && (db.CloudAddressID != nil || db.CloudAddressGroupID != nil) {
		return true
	}

	if *cloud.Action != db.Action {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.PolicyDescription, db.Memo) {
		return true
	}

	return false
}
