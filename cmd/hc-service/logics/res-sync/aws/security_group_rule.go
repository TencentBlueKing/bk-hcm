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

package aws

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
	"hcm/pkg/tools/assert"
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
				enumor.Aws, params.AccountID, param, err, kt.Rid)
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

	addSlice, updateMap, delCloudIDs := common.Diff[securitygrouprule.AwsSGRule,
		corecloud.AwsSecurityGroupRule](sgRuleFromCloud, sgRuleFromDB, isSGRuleChange)

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

func (cli *client) createSGRule(kt *kit.Kit, opt *syncSGRuleOption,
	addSlice []securitygrouprule.AwsSGRule) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("sgRule addSlice is <= 0, not create")
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list, err := cli.genAddRuleList(addSlice, opt)
	if err != nil {
		return err
	}

	createReq := &protocloud.AwsSGRuleCreateReq{
		Rules: list,
	}
	_, err = cli.dbCli.Aws.SecurityGroup.BatchCreateSecurityGroupRule(
		kt.Ctx, kt.Header(), createReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to create sgRule success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) genAddRuleList(rules []securitygrouprule.AwsSGRule,
	opt *syncSGRuleOption) ([]protocloud.AwsSGRuleBatchCreate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.AwsSGRuleBatchCreate, 0, len(rules))

	for _, rule := range rules {
		one := protocloud.AwsSGRuleBatchCreate{
			CloudID:              converter.PtrToVal(rule.SecurityGroupRuleId),
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             rule.FromPort,
			ToPort:               rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: converter.PtrToVal(rule.GroupId),
			CloudGroupOwnerID:    converter.PtrToVal(rule.GroupOwnerId),
			AccountID:            opt.AccountID,
			Region:               opt.Region,
			SecurityGroupID:      opt.SGMap[opt.CloudSGID],
		}

		if converter.PtrToVal(rule.IsEgress) {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}

		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}

		list = append(list, one)
	}

	return list, nil
}

func (cli *client) updateSGRule(kt *kit.Kit, opt *syncSGRuleOption,
	updateMap map[string]securitygrouprule.AwsSGRule) error {

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

	updateReq := &protocloud.AwsSGRuleBatchUpdateReq{
		Rules: list,
	}
	err = cli.dbCli.Aws.SecurityGroup.BatchUpdateSecurityGroupRule(
		kt.Ctx, kt.Header(), updateReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to update sgRule success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) genUpdateRulesList(updateMap map[string]securitygrouprule.AwsSGRule,
	opt *syncSGRuleOption) ([]protocloud.AwsSGRuleUpdate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.AwsSGRuleUpdate, 0)

	for id, rule := range updateMap {
		one := protocloud.AwsSGRuleUpdate{
			ID:                   id,
			CloudID:              converter.PtrToVal(rule.SecurityGroupRuleId),
			IPv4Cidr:             rule.CidrIpv4,
			IPv6Cidr:             rule.CidrIpv6,
			Memo:                 rule.Description,
			FromPort:             rule.FromPort,
			ToPort:               rule.ToPort,
			Protocol:             rule.IpProtocol,
			CloudPrefixListID:    rule.PrefixListId,
			CloudSecurityGroupID: converter.PtrToVal(rule.GroupId),
			CloudGroupOwnerID:    converter.PtrToVal(rule.GroupOwnerId),
			AccountID:            opt.AccountID,
			Region:               opt.Region,
			SecurityGroupID:      opt.SGMap[opt.CloudSGID],
		}

		if converter.PtrToVal(rule.IsEgress) {
			one.Type = enumor.Egress
		} else {
			one.Type = enumor.Ingress
		}

		if rule.ReferencedGroupInfo != nil {
			one.CloudTargetSecurityGroupID = rule.ReferencedGroupInfo.GroupId
		}

		list = append(list, one)
	}

	return list, nil
}

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
		delCloudSGRuleIDMap[converter.PtrToVal(one.SecurityGroupRuleId)] = false
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
			enumor.Aws, len(delSGRuleFromCloud), kt.Rid)
		return fmt.Errorf("validate sgRule not exist failed, before delete")
	}

	deleteReq := &protocloud.AwsSGRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	err = cli.dbCli.Aws.SecurityGroup.BatchDeleteSecurityGroupRule(
		kt.Ctx, kt.Header(), deleteReq, opt.SGMap[opt.CloudSGID])
	if err != nil {
		logs.Errorf("[%s] dataservice delete aws security group rules failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sgRule to delete sgRule success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGRuleFromCloud(kt *kit.Kit, opt *syncSGRuleOption) ([]securitygrouprule.AwsSGRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgRuleopt := &securitygrouprule.AwsListOption{
		Region:               opt.Region,
		CloudSecurityGroupID: opt.CloudSGID,
	}

	rules, err := cli.cloudCli.ListSecurityGroupRule(kt, sgRuleopt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list aws security group rule failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return nil, err
	}

	return rules, nil
}

func (cli *client) listSGRuleFromDB(kt *kit.Kit, opt *syncSGRuleOption) ([]corecloud.AwsSecurityGroupRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	listReq := &protocloud.AwsSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", opt.SGMap[opt.CloudSGID]),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	rules := make([]corecloud.AwsSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := cli.dbCli.Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
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

func isSGRuleChange(cloud securitygrouprule.AwsSGRule,
	db corecloud.AwsSecurityGroupRule) bool {

	if db.CloudID != converter.PtrToVal(cloud.SecurityGroupRuleId) {
		return true
	}

	if !assert.IsPtrStringEqual(db.IPv4Cidr, cloud.CidrIpv4) {
		return true
	}

	if !assert.IsPtrStringEqual(db.IPv6Cidr, cloud.CidrIpv6) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Memo, cloud.Description) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.FromPort, cloud.FromPort) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.ToPort, cloud.ToPort) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Protocol, cloud.IpProtocol) {
		return true
	}

	if !assert.IsPtrStringEqual(db.CloudPrefixListID, cloud.PrefixListId) {
		return true
	}

	if db.CloudSecurityGroupID != *cloud.GroupId {
		return true
	}

	if db.CloudGroupOwnerID != *cloud.GroupOwnerId {
		return true
	}

	if cloud.ReferencedGroupInfo != nil {
		if !assert.IsPtrStringEqual(db.CloudTargetSecurityGroupID, cloud.ReferencedGroupInfo.GroupId) {
			return true
		}
	}

	return false
}
