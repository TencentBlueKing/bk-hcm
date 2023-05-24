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

	"hcm/pkg/adaptor/huawei"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	apicore "hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// SyncHuaWeiSecurityGroupOption define sync huawei sg and sg rule option.
type SyncHuaWeiSecurityGroupOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiSecurityGroupOption
func (opt SyncHuaWeiSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.SGBatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.SGBatchOperationMaxLimit)
	}

	return nil
}

// SyncHuaWeiSGRule sync huawei security group rules.
func SyncHuaWeiSGRule(kt *kit.Kit, req *SyncHuaWeiSecurityGroupOption,
	client *huawei.HuaWei, dataCli *dataservice.Client, sgID string) (interface{}, error) {

	sg, err := dataCli.HuaWei.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get huawei security group failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err, err
	}

	cloudAllIDs := make(map[string]bool)
	opt := &securitygrouprule.HuaWeiListOption{
		Region:               req.Region,
		CloudSecurityGroupID: sg.CloudID,
	}

	rules, err := client.ListSecurityGroupRule(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list huawei security group rule failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	if len(*rules.SecurityGroupRules) <= 0 {
		return nil, nil
	}

	cloudMap := make(map[string]*HuaWeiSGRuleSync)
	cloudIDs := make([]string, 0, len(*rules.SecurityGroupRules))
	for _, rule := range *rules.SecurityGroupRules {
		sgRuleSync := new(HuaWeiSGRuleSync)
		sgRuleSync.IsUpdate = false
		sgRuleSync.SGRule = rule
		cloudMap[rule.Id] = sgRuleSync
		cloudIDs = append(cloudIDs, rule.Id)
		cloudAllIDs[rule.Id] = true
	}

	updateIDs, err := getHuaWeiSGRuleDSSync(kt, cloudIDs, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("[%s] request getHuaWeiSGRuleDSSync failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := syncHuaWeiSGRuleUpdate(kt, updateIDs, cloudMap, sgID, req, dataCli)
		if err != nil {
			logs.Errorf("[%s] request syncHuaWeiSGRuleUpdate failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
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
		err := syncHuaWeiSGRuleAdd(kt, addIDs, req, cloudMap, sgID, dataCli)
		if err != nil {
			logs.Errorf("[%s] request syncHuaWeiSGRuleAdd failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
			return nil, err
		}
	}

	dsIDs, err := getHuaWeiSGRuleAllDS(kt, req, sgID, dataCli)
	if err != nil {
		logs.Errorf("[%s] request getHuaWeiSGRuleAllDS failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
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
			logs.Errorf("[%s] request adaptor to list aws security group rule failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
			return nil, err
		}

		for _, id := range deleteIDs {
			realDeleteFlag := true
			for _, rule := range *rules.SecurityGroupRules {
				if rule.Id == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		if len(realDeleteIDs) > 0 {
			err := syncHuaWeiSGRuleDelete(kt, realDeleteIDs, sgID, dataCli)
			if err != nil {
				logs.Errorf("[%s] request syncHuaWeiSGRuleDelete failed, err: %v, rid: %s", enumor.HuaWei,
					err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func syncHuaWeiSGRuleUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*HuaWeiSGRuleSync,
	sgID string, req *SyncHuaWeiSecurityGroupOption, dataCli *dataservice.Client) error {

	rulesResp := new(model.ListSecurityGroupRulesResponse)
	rules := make([]model.SecurityGroupRule, 0)
	for _, id := range updateIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}
	rulesResp.SecurityGroupRules = &rules

	sg, err := dataCli.HuaWei.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get huawei security group failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	list := genHuaWeiUpdateRulesList(kt, rulesResp, sgID, sg.CloudID, req, dataCli)
	updateReq := &protocloud.HuaWeiSGRuleBatchUpdateReq{
		Rules: list,
	}

	if len(updateReq.Rules) > 0 {
		for _, v := range updateReq.Rules {
			cloudMap[v.CloudID].IsRealUpdate = true
		}
		err := dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), updateReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncHuaWeiSGRuleAdd(kt *kit.Kit, addIDs []string, req *SyncHuaWeiSecurityGroupOption,
	cloudMap map[string]*HuaWeiSGRuleSync, sgID string, dataCli *dataservice.Client) error {

	rulesResp := new(model.ListSecurityGroupRulesResponse)
	rules := make([]model.SecurityGroupRule, 0)
	for _, id := range addIDs {
		if value, ok := cloudMap[id]; ok {
			rules = append(rules, value.SGRule)
		}
	}
	rulesResp.SecurityGroupRules = &rules

	sg, err := dataCli.HuaWei.SecurityGroup.GetSecurityGroup(kt.Ctx, kt.Header(), sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get huawei security group failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	list := genHuaWeiAddRuleList(rulesResp, req, sg.CloudID, sgID)
	createReq := &protocloud.HuaWeiSGRuleCreateReq{
		Rules: list,
	}

	if len(createReq.Rules) > 0 {
		_, err := dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), createReq, sgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncHuaWeiSGRuleDelete(kt *kit.Kit, deleteCloudIDs []string, sgID string,
	dataCli *dataservice.Client) error {

	for _, id := range deleteCloudIDs {
		deleteReq := &protocloud.HuaWeiSGRuleBatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		err := dataCli.HuaWei.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), deleteReq, sgID)
		if err != nil {
			logs.Errorf("[%s] dataservice delete huawei security group rules failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
			return err
		}
	}

	return nil
}

func getHuaWeiSGRuleAllDS(kt *kit.Kit, req *SyncHuaWeiSecurityGroupOption, sgID string,
	dataCli *dataservice.Client) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &protocloud.HuaWeiSGRuleListReq{
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

		results, err := dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("[%s] from data-service list sg rule failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
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

func getHuaWeiSGRuleDSSync(kt *kit.Kit, cloudIDs []string, req *SyncHuaWeiSecurityGroupOption,
	sgID string, dataCli *dataservice.Client) ([]string, error) {

	updateIDs := make([]string, 0)

	start := 0
	for {

		dataReq := &protocloud.HuaWeiSGRuleListReq{
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

		results, err := dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), dataReq, sgID)
		if err != nil {
			logs.Errorf("[%s] from data-service list sg rule failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
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

func genHuaWeiAddRuleList(rules *model.ListSecurityGroupRulesResponse, req *SyncHuaWeiSecurityGroupOption,
	sgCloudID string, id string) []protocloud.HuaWeiSGRuleBatchCreate {

	list := make([]protocloud.HuaWeiSGRuleBatchCreate, 0, len(*rules.SecurityGroupRules))

	for _, sgRule := range *rules.SecurityGroupRules {
		rule := protocloud.HuaWeiSGRuleBatchCreate{
			CloudID:                   sgRule.Id,
			Memo:                      &sgRule.Description,
			Protocol:                  sgRule.Protocol,
			Ethertype:                 sgRule.Ethertype,
			CloudRemoteGroupID:        sgRule.RemoteGroupId,
			RemoteIPPrefix:            sgRule.RemoteIpPrefix,
			CloudRemoteAddressGroupID: sgRule.RemoteAddressGroupId,
			Port:                      sgRule.Multiport,
			Priority:                  int64(sgRule.Priority),
			Action:                    sgRule.Action,
			Type:                      enumor.SecurityGroupRuleType(sgRule.Direction),
			CloudSecurityGroupID:      sgCloudID,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 req.AccountID,
			Region:                    req.Region,
			SecurityGroupID:           id,
		}
		list = append(list, rule)
	}

	return list
}

func isHuaWeiSGRuleChange(db *corecloud.HuaWeiSecurityGroupRule, cloud model.SecurityGroupRule) bool {

	if *db.Memo != cloud.Description {
		return true
	}

	if db.Protocol != cloud.Protocol {
		return true
	}

	if db.Ethertype != cloud.Ethertype {
		return true
	}

	if db.CloudRemoteGroupID != cloud.RemoteGroupId {
		return true
	}

	if db.RemoteIPPrefix != cloud.RemoteIpPrefix {
		return true
	}

	if db.CloudRemoteAddressGroupID != cloud.RemoteAddressGroupId {
		return true
	}

	if db.Port != cloud.Multiport {
		return true
	}

	if db.Priority != int64(cloud.Priority) {
		return true
	}

	if db.Action != cloud.Action {
		return true
	}

	if db.Type != enumor.SecurityGroupRuleType(cloud.Direction) {
		return true
	}

	if db.CloudProjectID != cloud.ProjectId {
		return true
	}

	return false
}

func genHuaWeiUpdateRulesList(kt *kit.Kit, rules *model.ListSecurityGroupRulesResponse,
	sgID string, id string, req *SyncHuaWeiSecurityGroupOption, dataCli *dataservice.Client) []protocloud.HuaWeiSGRuleBatchUpdate {

	list := make([]protocloud.HuaWeiSGRuleBatchUpdate, 0)

	for _, sgRule := range *rules.SecurityGroupRules {
		one, err := getHuaWeiSGRuleByCid(kt, sgRule.Id, sgID, dataCli)
		if err != nil || one == nil {
			logs.Errorf("[%s] gen update RulesList getHuaWeiSGRuleByCid failed, err: %v, rid: %s", enumor.HuaWei,
				err, kt.Rid)
			continue
		}

		if !isHuaWeiSGRuleChange(one, sgRule) {
			continue
		}

		rule := protocloud.HuaWeiSGRuleBatchUpdate{
			ID:                        one.ID,
			CloudID:                   sgRule.Id,
			Memo:                      &sgRule.Description,
			Protocol:                  sgRule.Protocol,
			Ethertype:                 sgRule.Ethertype,
			CloudRemoteGroupID:        sgRule.RemoteGroupId,
			RemoteIPPrefix:            sgRule.RemoteIpPrefix,
			CloudRemoteAddressGroupID: sgRule.RemoteAddressGroupId,
			Port:                      sgRule.Multiport,
			Priority:                  int64(sgRule.Priority),
			Action:                    sgRule.Action,
			Type:                      enumor.SecurityGroupRuleType(sgRule.Direction),
			CloudSecurityGroupID:      id,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 req.AccountID,
			Region:                    req.Region,
			SecurityGroupID:           sgID,
		}
		list = append(list, rule)
	}

	return list
}

func getHuaWeiSGRuleByCid(kt *kit.Kit, cID string, sgID string,
	dataCli *dataservice.Client) (*corecloud.HuaWeiSecurityGroupRule, error) {

	listReq := &protocloud.HuaWeiSGRuleListReq{
		Filter: tools.EqualExpression("cloud_id", cID),
		Page:   core.DefaultBasePage,
	}
	listResp, err := dataCli.HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("[%s] request dataservice get huawei security group failed, id: %s, err: %v, rid: %s", enumor.HuaWei,
			cID, err, kt.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", cID)
	}

	return &listResp.Details[0], nil
}
