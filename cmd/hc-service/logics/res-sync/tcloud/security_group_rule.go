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

package tcloud

import (
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// SyncSGRuleOption ...
type SyncSGRuleOption struct {
}

// Validate ...
func (opt SyncSGRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SecurityGroupRule 同步安全组规则唯一指定方法
func (cli *client) SecurityGroupRule(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGRuleOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var syncResult *SyncResult
	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncSGRuleOption{
			AccountID: params.AccountID,
			Region:    params.Region,
			SGID:      param,
		}
		var err error
		if syncResult, err = cli.securityGroupRule(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s sg: %s sync sgRule failed, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, param, err, kt.Rid)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return syncResult, nil
}

type syncSGRuleOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	SGID      string `json:"sgid" validate:"required"`
}

// Validate ...
func (opt syncSGRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) securityGroupRule(kt *kit.Kit, opt *syncSGRuleOption) (*SyncResult, error) {

	sg, err := cli.dbCli.TCloud.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), opt.SGID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get TCloud security group failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return nil, err
	}

	version, egressRuleMaps, ingressRuleMaps, argsTplCloudIDs, err := cli.listSGRuleFromCloud(kt, sg.Region, sg.CloudID)
	if err != nil {
		return nil, err
	}

	// 获取模版参数列表
	argsTplMap, err := cli.listArgumentTemplateMapFromDB(kt, argsTplCloudIDs)
	if err != nil {
		logs.Errorf("[%s] request dataservice get TCloud argument template failed, cloudIDs: %v, err: %v, rid: %s",
			enumor.TCloud, argsTplCloudIDs, err, kt.Rid)
		return nil, err
	}

	rulesFromDB, err := cli.listSGRuleFromDB(kt, sg.ID)
	if err != nil {
		return nil, err
	}

	updateRules := make(map[string]*corecloud.TCloudSecurityGroupRule)
	deleteRuleIDs := make([]string, 0)
	for _, one := range rulesFromDB {
		var ruleMap map[int64]*vpc.SecurityGroupPolicy
		switch one.Type {
		case enumor.Egress:
			ruleMap = egressRuleMaps
		case enumor.Ingress:
			ruleMap = ingressRuleMaps
		default:
			logs.Errorf("[%s] unknown security group rule type: %s, skip handle, rid: %s", enumor.TCloud,
				one.Type, kt.Rid)
			continue
		}
		policy, exist := ruleMap[one.CloudPolicyIndex]
		if !exist {
			deleteRuleIDs = append(deleteRuleIDs, one.ID)
			continue
		}
		delete(ruleMap, one.CloudPolicyIndex)
		if isSGRuleChange(version, policy, one) {
			updateRules[one.ID] = convTCloudRule(policy, &sg.BaseSecurityGroup, version, one.Type, argsTplMap)
		}
	}

	createRules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for _, policy := range egressRuleMaps {
		rule := convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Egress, argsTplMap)
		createRules = append(createRules, *rule)
	}

	for _, policy := range ingressRuleMaps {
		rule := convTCloudRule(policy, &sg.BaseSecurityGroup, version, enumor.Ingress, argsTplMap)
		createRules = append(createRules, *rule)
	}

	if len(deleteRuleIDs) != 0 {
		if err = cli.deleteSGRule(kt, sg.ID, deleteRuleIDs); err != nil {
			return nil, err
		}
	}

	if len(updateRules) != 0 {
		if err = cli.updateSGRule(kt, sg.ID, updateRules); err != nil {
			return nil, err
		}
	}

	syncResult := &SyncResult{}
	if len(createRules) != 0 {
		syncResult.CreatedIds, err = cli.createSGRule(kt, sg.ID, createRules)
		if err != nil {
			return nil, err
		}
	}

	return syncResult, nil
}

// listSGRuleFromCloud list tcloud security group rule from database
func (cli *client) listSGRuleFromDB(kt *kit.Kit, sgID string) (
	[]corecloud.TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", sgID),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	rules := make([]corecloud.TCloudSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := cli.dbCli.TCloud.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
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

// listSGRuleFromCloud list tcloud security group rule from cloud
func (cli *client) listSGRuleFromCloud(kt *kit.Kit, region, cloudSGID string) (string,
	map[int64]*vpc.SecurityGroupPolicy, map[int64]*vpc.SecurityGroupPolicy, []string, error) {
	listOpt := &securitygrouprule.TCloudListOption{
		Region:               region,
		CloudSecurityGroupID: cloudSGID,
	}
	rules, err := cli.cloudCli.ListSecurityGroupRule(kt, listOpt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list tcloud security group rule failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return "", nil, nil, nil, err
	}

	// 该安全规则绑定的参数模版的云端ID数组
	argsTplCloudIDs := make([]string, 0)
	egressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Egress))
	ingressRuleMaps := make(map[int64]*vpc.SecurityGroupPolicy, len(rules.Ingress))
	for _, egress := range rules.Egress {
		egressRuleMaps[*egress.PolicyIndex] = egress
		// 该安全组规则绑定的参数模版
		if egress.AddressTemplate != nil {
			tmpAddressID := converter.PtrToVal(egress.AddressTemplate.AddressId)
			if len(tmpAddressID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpAddressID)
			}
			tmpAddressGroupID := converter.PtrToVal(egress.AddressTemplate.AddressGroupId)
			if len(tmpAddressGroupID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpAddressGroupID)
			}
		}
		if egress.ServiceTemplate != nil {
			tmpServiceID := converter.PtrToVal(egress.ServiceTemplate.ServiceId)
			if len(tmpServiceID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpServiceID)
			}
			tmpServiceGroupID := converter.PtrToVal(egress.ServiceTemplate.ServiceGroupId)
			if len(tmpServiceGroupID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpServiceGroupID)
			}
		}
	}

	for _, ingress := range rules.Ingress {
		ingressRuleMaps[*ingress.PolicyIndex] = ingress
		// 该安全组规则绑定的参数模版
		if ingress.AddressTemplate != nil {
			tmpAddressID := converter.PtrToVal(ingress.AddressTemplate.AddressId)
			if len(tmpAddressID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpAddressID)
			}
			tmpAddressGroupID := converter.PtrToVal(ingress.AddressTemplate.AddressGroupId)
			if len(tmpAddressGroupID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpAddressGroupID)
			}
		}
		if ingress.ServiceTemplate != nil {
			tmpServiceID := converter.PtrToVal(ingress.ServiceTemplate.ServiceId)
			if len(tmpServiceID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpServiceID)
			}
			tmpServiceGroupID := converter.PtrToVal(ingress.ServiceTemplate.ServiceGroupId)
			if len(tmpServiceGroupID) != 0 {
				argsTplCloudIDs = append(argsTplCloudIDs, tmpServiceGroupID)
			}
		}
	}

	return converter.PtrToVal(rules.Version), egressRuleMaps, ingressRuleMaps, argsTplCloudIDs, nil
}

// updateSGRule update security group rule
func (cli *client) updateSGRule(kt *kit.Kit, sgID string, updateRules map[string]*corecloud.
	TCloudSecurityGroupRule) error {

	// convert update rules map to rule slice
	ruleSlice := make([]protocloud.TCloudSGRuleBatchUpdate, 0, len(updateRules))
	for id, rule := range updateRules {
		//  override id by map key
		ruleSlice = append(ruleSlice, protocloud.TCloudSGRuleBatchUpdate{
			ID:               id,
			CloudPolicyIndex: rule.CloudPolicyIndex,
			Version:          rule.Version,
			// 如果云上该字段被更新为空，会置于null，但对于海垒来说，空值是 ""，null是不更新。
			Protocol:                   converter.ValToPtr(converter.PtrToVal(rule.Protocol)),
			Port:                       rule.Port,
			ServiceID:                  rule.ServiceID,
			CloudServiceID:             rule.CloudServiceID,
			ServiceGroupID:             rule.ServiceGroupID,
			CloudServiceGroupID:        rule.CloudServiceGroupID,
			IPv4Cidr:                   rule.IPv4Cidr,
			IPv6Cidr:                   rule.IPv6Cidr,
			CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
			AddressID:                  rule.AddressID,
			CloudAddressID:             rule.CloudAddressID,
			AddressGroupID:             rule.AddressGroupID,
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
	// split rules into batches to avoid reaching batch operation limit
	ruleBatches := slice.Split(ruleSlice, constant.BatchOperationMaxLimit)
	for batchIdx, updateRuleBatch := range ruleBatches {
		req := &protocloud.TCloudSGRuleBatchUpdateReq{Rules: updateRuleBatch}
		err := cli.dbCli.TCloud.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
		if err != nil {
			logs.Errorf("[%s] request dataservice to batch update tcloud security group rule failed, "+
				"err: %v, batch idx: %d, rid: %s", enumor.TCloud, err, batchIdx, kt.Rid)
			return err
		}
	}
	return nil
}

// deleteSGRule delete security group rule
func (cli *client) deleteSGRule(kt *kit.Kit, sgID string, delIDs []string) error {

	// split rules into batches to avoid reaching batch operation limit
	delIdBatches := slice.Split(delIDs, constant.BatchOperationMaxLimit)
	for batchIdx, delIdBatch := range delIdBatches {

		req := &protocloud.TCloudSGRuleBatchDeleteReq{
			Filter: tools.ContainersExpression("id", delIdBatch),
		}
		err := cli.dbCli.TCloud.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
		if err != nil {
			logs.Errorf("[%s] request dataservice to delete tcloud security group rule failed,"+
				" err: %v,batch idx：%d, rid: %s", enumor.TCloud, err, batchIdx, kt.Rid)
			return err
		}
	}

	return nil
}

// createSGRule crate security group rule
func (cli *client) createSGRule(kt *kit.Kit, sgID string, allRules []corecloud.
	TCloudSecurityGroupRule) ([]string, error) {

	// split all rules into batches to avoid reaching batch operation limit
	splitRuleBatches := slice.Split(allRules, constant.BatchOperationMaxLimit)
	resultIds := make([]string, 0, len(allRules))
	for batchIdx, ruleBatch := range splitRuleBatches {
		ruleCreates := make([]protocloud.TCloudSGRuleBatchCreate, 0, len(ruleBatch))
		for _, rule := range ruleBatch {
			ruleCreates = append(ruleCreates, protocloud.TCloudSGRuleBatchCreate{
				CloudPolicyIndex:           rule.CloudPolicyIndex,
				Version:                    rule.Version,
				Protocol:                   converter.ValToPtr(converter.PtrToVal(rule.Protocol)),
				Port:                       rule.Port,
				ServiceID:                  rule.ServiceID,
				CloudServiceID:             rule.CloudServiceID,
				ServiceGroupID:             rule.ServiceGroupID,
				CloudServiceGroupID:        rule.CloudServiceGroupID,
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
				AddressID:                  rule.AddressID,
				CloudAddressID:             rule.CloudAddressID,
				AddressGroupID:             rule.AddressGroupID,
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
		req := &protocloud.TCloudSGRuleCreateReq{Rules: ruleCreates}
		resultBatch, err := cli.dbCli.TCloud.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), req, sgID)
		if err != nil {
			logs.Errorf("[%s] request dataservice to create tcloud security group rule failed, "+
				"err: %v,batch idx: %d, rid: %s", enumor.TCloud, err, batchIdx, kt.Rid)
			return nil, err
		}
		resultIds = append(resultIds, resultBatch.IDs...)

	}

	return resultIds, nil
}

func convTCloudRule(policy *vpc.SecurityGroupPolicy, sg *corecloud.BaseSecurityGroup, version string,
	ruleType enumor.SecurityGroupRuleType,
	argsTplMap map[string]coreargstpl.BaseArgsTpl) *corecloud.TCloudSecurityGroupRule {

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

		// 该安全组规则绑定的[协议端口]参数模版的本地ID
		if argsTplInfo, ok := argsTplMap[converter.PtrToVal(spec.CloudServiceID)]; ok {
			spec.ServiceID = converter.ValToPtr(argsTplInfo.ID)
		} else {
			spec.ServiceID = converter.ValToPtr("")
		}

		if argsTplInfo, ok := argsTplMap[converter.PtrToVal(spec.CloudServiceGroupID)]; ok {
			spec.ServiceGroupID = converter.ValToPtr(argsTplInfo.ID)
		} else {
			spec.ServiceGroupID = converter.ValToPtr("")
		}
	}

	if policy.AddressTemplate != nil {
		spec.CloudAddressID = policy.AddressTemplate.AddressId
		spec.CloudAddressGroupID = policy.AddressTemplate.AddressGroupId

		// 该安全组规则绑定的[IP]参数模版的本地ID
		if argsTplInfo, ok := argsTplMap[converter.PtrToVal(spec.CloudAddressID)]; ok {
			spec.AddressID = converter.ValToPtr(argsTplInfo.ID)
		} else {
			spec.AddressID = converter.ValToPtr("")
		}

		if argsTplInfo, ok := argsTplMap[converter.PtrToVal(spec.CloudAddressGroupID)]; ok {
			spec.AddressGroupID = converter.ValToPtr(argsTplInfo.ID)
		} else {
			spec.AddressGroupID = converter.ValToPtr("")
		}
	}

	return spec
}

func isSGRuleChange(version string, cloud *vpc.SecurityGroupPolicy, db corecloud.TCloudSecurityGroupRule) bool {
	if version != db.Version {
		return true
	}
	if converter.PtrToVal(cloud.PolicyIndex) != db.CloudPolicyIndex {
		return true
	}
	if !assert.IsPtrStringEqual(converter.ValToPtr(converter.PtrToVal(cloud.Protocol)), db.Protocol) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.Port, db.Port) {
		return true
	}

	if cloud.ServiceTemplate != nil {
		// 参数模版-协议端口、协议端口组为空，则需要更新
		if (converter.PtrToVal(cloud.ServiceTemplate.ServiceId) != "" && db.ServiceID == nil) ||
			(converter.PtrToVal(cloud.ServiceTemplate.ServiceGroupId) != "" && db.ServiceGroupID == nil) {
			return true
		}

		if db.CloudServiceID != nil || db.CloudServiceGroupID != nil {
			if !assert.IsPtrStringEqual(cloud.ServiceTemplate.ServiceId, db.CloudServiceID) {
				return true
			}

			if !assert.IsPtrStringEqual(cloud.ServiceTemplate.ServiceGroupId, db.CloudServiceGroupID) {
				return true
			}
		}
	}

	if cloud.ServiceTemplate == nil && (converter.PtrToVal(db.ServiceID) != "" ||
		converter.PtrToVal(db.CloudServiceID) != "" || converter.PtrToVal(db.ServiceGroupID) != "" ||
		converter.PtrToVal(db.CloudServiceGroupID) != "") {
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

	if cloud.AddressTemplate != nil {
		// 参数模版-IP地址、IP地址组为空，则需要更新
		if (converter.PtrToVal(cloud.AddressTemplate.AddressId) != "" && db.AddressID == nil) ||
			(converter.PtrToVal(cloud.AddressTemplate.AddressGroupId) != "" && db.AddressGroupID == nil) {
			return true
		}

		if db.CloudAddressID != nil || db.CloudAddressGroupID != nil {
			if !assert.IsPtrStringEqual(cloud.AddressTemplate.AddressId, db.CloudAddressID) {
				return true
			}

			if !assert.IsPtrStringEqual(cloud.AddressTemplate.AddressGroupId, db.CloudAddressGroupID) {
				return true
			}
		}
	}

	if cloud.AddressTemplate == nil && (converter.PtrToVal(db.AddressID) != "" ||
		converter.PtrToVal(db.CloudAddressID) != "" || converter.PtrToVal(db.AddressGroupID) != "" ||
		converter.PtrToVal(db.CloudAddressGroupID) != "") {
		return true
	}

	if converter.PtrToVal(cloud.Action) != db.Action {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.PolicyDescription, db.Memo) {
		return true
	}
	return false
}

// listArgumentTemplateMapFromDB list tcloud argument template from database
func (cli *client) listArgumentTemplateMapFromDB(kt *kit.Kit, cloudIDs []string) (
	map[string]coreargstpl.BaseArgsTpl, error) {

	argsTplMap := make(map[string]coreargstpl.BaseArgsTpl, 0)
	if len(cloudIDs) == 0 {
		return argsTplMap, nil
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	for {
		listReq.Page.Start = start
		listResp, err := cli.dbCli.Global.ArgsTpl.ListArgsTpl(kt, listReq)
		if err != nil {
			return nil, err
		}

		for _, item := range listResp.Details {
			argsTplMap[item.CloudID] = item
		}

		if len(listResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return argsTplMap, nil
}
