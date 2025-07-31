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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
)

// SyncSGRuleOption ...
type SyncSGRuleOption struct {
}

// Validate ...
func (opt SyncSGRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SecurityGroupRule ...
func (cli *client) SecurityGroupRule(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGRuleOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgFromDB, err := cli.listSGFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	sgMap := make(map[string]string)
	for _, one := range sgFromDB {
		sgMap[one.CloudID] = one.ID
	}

	if len(sgMap) != len(params.CloudIDs) {
		return nil, fmt.Errorf("sg num is not match")
	}

	err = concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncSGRuleOption{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudSGID: param,
			SGMap:     sgMap,
		}
		if _, err := cli.securityGroupRule(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s sg: %s sync sgRule failed, err: %v, rid: %s",
				enumor.HuaWei, params.AccountID, param, err, kt.Rid)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

type syncSGRuleOption struct {
	AccountID string            `json:"account_id" validate:"required"`
	Region    string            `json:"region" validate:"required"`
	CloudSGID string            `json:"cloud_sgid" validate:"required"`
	SGMap     map[string]string `json:"sg_map" validate:"required"`
}

// Validate ...
func (opt syncSGRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// securityGroupRule 同步安全组规则
func (cli *client) securityGroupRule(kt *kit.Kit, opt *syncSGRuleOption) (*SyncResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgRuleFromDB, err := cli.listSGRuleFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	sgRuleFromCloud, err := cli.listSGRuleFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(sgRuleFromCloud) == 0 && len(sgRuleFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[securitygrouprule.HuaWeiSGRule,
		corecloud.HuaWeiSecurityGroupRule](sgRuleFromCloud, sgRuleFromDB, isSGRuleChange)

	if len(delCloudIDs) > 0 {
		err := cli.deleteSGRule(kt, opt, delCloudIDs)
		if err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		err := cli.createSGRule(kt, opt, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		err := cli.updateSGRule(kt, opt, updateMap)
		if err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// createSGRule create security group rule
func (cli *client) createSGRule(kt *kit.Kit, opt *syncSGRuleOption,
	addSlice []securitygrouprule.HuaWeiSGRule) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("sgRule addSlice is <= 0, not create")
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil
	}

	list, err := cli.genAddRuleList(addSlice, opt)
	if err != nil {
		return err
	}

	createReq := &protocloud.HuaWeiSGRuleCreateReq{
		Rules: list,
	}
	_, err = cli.dbCli.HuaWei.SecurityGroup.BatchCreateSecurityGroupRule(
		kt.Ctx, kt.Header(), createReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to create sgRule success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

// genAddRuleList generate security group rule create list
func (cli *client) genAddRuleList(rules []securitygrouprule.HuaWeiSGRule,
	opt *syncSGRuleOption) ([]protocloud.HuaWeiSGRuleBatchCreate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.HuaWeiSGRuleBatchCreate, 0, len(rules))

	for _, sgRule := range rules {
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
			CloudSecurityGroupID:      opt.CloudSGID,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 opt.AccountID,
			Region:                    opt.Region,
			SecurityGroupID:           opt.SGMap[opt.CloudSGID],
		}
		list = append(list, rule)
	}

	return list, nil
}

// updateSGRule update security group rule
func (cli *client) updateSGRule(kt *kit.Kit, opt *syncSGRuleOption,
	updateMap map[string]securitygrouprule.HuaWeiSGRule) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("sgRule updateMap is <= 0, not update")
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list, err := cli.genUpdateRulesList(updateMap, opt)
	if err != nil {
		return err
	}

	updateReq := &protocloud.HuaWeiSGRuleBatchUpdateReq{
		Rules: list,
	}
	err = cli.dbCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroupRule(
		kt.Ctx, kt.Header(), updateReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to update sgRule success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

// genUpdateRulesList generate security group rule update list
func (cli *client) genUpdateRulesList(updateMap map[string]securitygrouprule.HuaWeiSGRule,
	opt *syncSGRuleOption) ([]protocloud.HuaWeiSGRuleBatchUpdate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.HuaWeiSGRuleBatchUpdate, 0)

	for id, sgRule := range updateMap {
		rule := protocloud.HuaWeiSGRuleBatchUpdate{
			ID:                        id,
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
			CloudSecurityGroupID:      opt.CloudSGID,
			CloudProjectID:            sgRule.ProjectId,
			AccountID:                 opt.AccountID,
			Region:                    opt.Region,
			SecurityGroupID:           opt.SGMap[opt.CloudSGID],
		}
		list = append(list, rule)
	}

	return list, nil
}

// deleteSGRule delete security group rule
func (cli *client) deleteSGRule(kt *kit.Kit, opt *syncSGRuleOption, delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("sgRule delCloudIDs is <= 0, not delete")
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	delSGRuleFromCloud, err := cli.listSGRuleFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudSGRuleIDMap := make(map[string]bool)
	for _, one := range delSGRuleFromCloud {
		delCloudSGRuleIDMap[one.Id] = false
	}

	canNotDelete := false
	for _, id := range delCloudIDs {
		if _, exsit := delCloudSGRuleIDMap[id]; exsit {
			canNotDelete = true
			break
		}
	}

	if canNotDelete {
		logs.Errorf("[%s] validate sgRule not exist failed, before delete, failed_count: %d, rid: %s",
			enumor.HuaWei, len(delSGRuleFromCloud), kt.Rid)
		return fmt.Errorf("validate sgRule not exist failed, before delete")
	}

	deleteReq := &protocloud.HuaWeiSGRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	err = cli.dbCli.HuaWei.SecurityGroup.BatchDeleteSecurityGroupRule(
		kt.Ctx, kt.Header(), deleteReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		logs.Errorf("[%s] dataservice delete huawei security group rules failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sgRule to delete sgRule success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listSGRuleFromCloud list security group rule from cloud
func (cli *client) listSGRuleFromCloud(kt *kit.Kit, opt *syncSGRuleOption) ([]securitygrouprule.HuaWeiSGRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgRuleopt := &securitygrouprule.HuaWeiListOption{
		Region:               opt.Region,
		CloudSecurityGroupID: opt.CloudSGID,
	}

	rules, err := cli.cloudCli.ListSecurityGroupRule(kt, sgRuleopt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list huawei security group rule failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	return rules, nil
}

// listSGFromDB list security group from database
func (cli *client) listSGRuleFromDB(kt *kit.Kit, opt *syncSGRuleOption) ([]corecloud.HuaWeiSecurityGroupRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	listReq := &protocloud.HuaWeiSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", opt.SGMap[opt.CloudSGID]),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	rules := make([]corecloud.HuaWeiSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := cli.dbCli.HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
			opt.SGMap[opt.CloudSGID])
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

// isSGRuleChange check if security group rule has changed
func isSGRuleChange(cloud securitygrouprule.HuaWeiSGRule,
	db corecloud.HuaWeiSecurityGroupRule) bool {

	if converter.PtrToVal(db.Memo) != cloud.Description {
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
