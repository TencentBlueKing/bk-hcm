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

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ListSecurityGroupRulesByCloudTargetSGID list security group rules by cloud target security group id.
// return map[cloudSecurityGroupID][]securityGroupRuleID
func ListSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client,
	vendor enumor.Vendor, sgID string) (map[string][]string, error) {

	var cloudSGToSgRulesMap map[string][]string
	var err error
	switch vendor {
	case enumor.TCloud:
		cloudSGToSgRulesMap, err = listTCloudSecurityGroupRulesByCloudTargetSGID(kt, cli, sgID)
	case enumor.Aws:
		cloudSGToSgRulesMap, err = listAwsSecurityGroupRulesByCloudTargetSGID(kt, cli, sgID)
	case enumor.Azure:
		cloudSGToSgRulesMap, err = listAzureSecurityGroupRulesByCloudTargetSGID(kt, cli, sgID)
	case enumor.HuaWei:
		cloudSGToSgRulesMap, err = listHuaweiSecurityGroupRulesByCloudTargetSGID(kt, cli, sgID)
	default:
		return nil, fmt.Errorf("unsupported vendor %s for validateSecurityGroupRuleRel", vendor)
	}
	if err != nil {
		logs.Errorf("list SecurityGroupRules failed, err: %v, sgID: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}

	return cloudSGToSgRulesMap, nil
}

// listTCloudSecurityGroupRulesByCloudTargetSGID 查询来源为安全组的 安全组规则,
// 返回 map[安全组ID]安全组规则ID列表
func listTCloudSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client, sgID string) (
	map[string][]string, error) {

	sgCloudIDToRuleIDs := make(map[string][]string)
	listReq := &dataproto.TCloudSGRuleListReq{
		Filter: tools.ExpressionAnd(tools.RuleNotEqual("cloud_target_security_group_id", "")),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		resp, err := cli.TCloud.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
		if err != nil {
			return nil, err
		}
		for _, rule := range resp.Details {
			if rule.CloudTargetSecurityGroupID != nil {
				cloudID := *rule.CloudTargetSecurityGroupID
				sgCloudIDToRuleIDs[cloudID] = append(sgCloudIDToRuleIDs[cloudID], rule.ID)
			}
		}
		if len(resp.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgCloudIDToRuleIDs, nil
}

// listAwsSecurityGroupRulesByCloudTargetSGID 查询来源为安全组的 安全组规则,
// 返回 map[安全组ID]安全组规则ID列表
func listAwsSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client, sgID string) (
	map[string][]string, error) {

	sgCloudIDToRuleIDs := make(map[string][]string)
	listReq := &dataproto.AwsSGRuleListReq{
		Filter: tools.ExpressionAnd(tools.RuleNotEqual("cloud_target_security_group_id", "")),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		resp, err := cli.Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
		if err != nil {
			return nil, err
		}
		for _, rule := range resp.Details {
			if rule.CloudTargetSecurityGroupID != nil {
				cloudID := *rule.CloudTargetSecurityGroupID
				sgCloudIDToRuleIDs[cloudID] = append(sgCloudIDToRuleIDs[cloudID], rule.ID)
			}
		}
		if len(resp.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgCloudIDToRuleIDs, nil
}

// listAzureSecurityGroupRulesByCloudTargetSGID 查询来源为安全组的 安全组规则,
// 返回 map[安全组ID]安全组规则ID列表
func listAzureSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client, sgID string) (
	map[string][]string, error) {

	sgCloudIDToRuleIDs := make(map[string][]string)
	listReq := &dataproto.AzureSGRuleListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleJsonLengthGreaterThan("cloud_source_app_security_group_ids", 0),
		),
		Page: core.NewDefaultBasePage(),
	}
	for {
		resp, err := cli.Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
		if err != nil {
			return nil, err
		}
		for _, rule := range resp.Details {
			for _, cloudID := range converter.PtrToSlice(rule.CloudSourceAppSecurityGroupIDs) {
				sgCloudIDToRuleIDs[cloudID] = append(sgCloudIDToRuleIDs[cloudID], rule.ID)
			}
		}
		if len(resp.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgCloudIDToRuleIDs, nil
}

// listHuaweiSecurityGroupRulesByCloudTargetSGID 查询来源为安全组的 安全组规则,
// 返回 map[安全组ID]安全组规则ID列表
func listHuaweiSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client, sgID string) (
	map[string][]string, error) {

	sgCloudIDToRuleIDs := make(map[string][]string)
	listReq := &dataproto.HuaWeiSGRuleListReq{
		Filter: tools.ExpressionAnd(tools.RuleNotEqual("cloud_remote_group_id", "")),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		resp, err := cli.HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
		if err != nil {
			return nil, err
		}
		for _, rule := range resp.Details {
			sgCloudIDToRuleIDs[rule.CloudRemoteGroupID] = append(sgCloudIDToRuleIDs[rule.CloudRemoteGroupID], rule.ID)
		}
		if len(resp.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgCloudIDToRuleIDs, nil
}
