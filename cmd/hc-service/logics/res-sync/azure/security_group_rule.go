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

package azure

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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SyncSGRuleOption ...
type SyncSGRuleOption struct {
}

// SGMapData ...
type SGMapData struct {
	ID     string
	Region string
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

	sgMap := make(map[string]*SGMapData)
	for _, one := range sgFromDB {
		sgMap[one.CloudID] = &SGMapData{
			ID:     one.ID,
			Region: one.Region,
		}
	}

	if len(sgMap) != len(params.CloudIDs) {
		return nil, fmt.Errorf("sg num is not match")
	}

	err = concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncSGRuleOption{
			AccountID:         params.AccountID,
			ResourceGroupName: params.ResourceGroupName,
			CloudSGID:         param,
			SGMap:             sgMap,
		}
		if _, err := cli.securityGroupRule(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s sg: %s sync sgRule failed, err: %v, rid: %s",
				enumor.Azure, params.AccountID, param, err, kt.Rid)
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
	AccountID         string                `json:"account_id" validate:"required"`
	ResourceGroupName string                `json:"resource_group_name" validate:"required"`
	CloudSGID         string                `json:"cloud_sgid" validate:"required"`
	SGMap             map[string]*SGMapData `json:"sg_map" validate:"required"`
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

	addSlice, updateMap, delCloudIDs := common.Diff[securitygrouprule.AzureSGRule,
		corecloud.AzureSecurityGroupRule](sgRuleFromCloud, sgRuleFromDB, isSGRuleChange)

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
	addSlice []securitygrouprule.AzureSGRule) error {

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

	createReq := &protocloud.AzureSGRuleCreateReq{
		Rules: list,
	}
	_, err = cli.dbCli.Azure.SecurityGroup.BatchCreateSecurityGroupRule(kt.Ctx, kt.Header(), createReq,
		opt.SGMap[opt.CloudSGID].ID)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to create sgRule success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) genAddRuleList(rules []securitygrouprule.AzureSGRule,
	opt *syncSGRuleOption) ([]protocloud.AzureSGRuleBatchCreate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.AzureSGRuleBatchCreate, 0, len(rules))

	for _, sgRule := range rules {
		rule := protocloud.AzureSGRuleBatchCreate{
			CloudID:                    converter.PtrToVal(sgRule.ID),
			Etag:                       sgRule.Etag,
			Name:                       converter.PtrToVal(sgRule.Name),
			Memo:                       sgRule.Description,
			DestinationAddressPrefix:   sgRule.DestinationAddressPrefix,
			DestinationAddressPrefixes: sgRule.DestinationAddressPrefixes,
			DestinationPortRange:       sgRule.DestinationPortRange,
			DestinationPortRanges:      sgRule.DestinationPortRanges,
			Protocol:                   string(converter.PtrToVal(sgRule.Protocol)),
			ProvisioningState:          string(converter.PtrToVal(sgRule.ProvisioningState)),
			SourceAddressPrefix:        sgRule.SourceAddressPrefix,
			SourceAddressPrefixes:      sgRule.SourceAddressPrefixes,
			SourcePortRange:            sgRule.SourcePortRange,
			SourcePortRanges:           sgRule.SourcePortRanges,
			Priority:                   converter.PtrToVal(sgRule.Priority),
			Access:                     string(converter.PtrToVal(sgRule.Access)),
			CloudSecurityGroupID:       opt.CloudSGID,
			Region:                     opt.SGMap[opt.CloudSGID].Region,
			AccountID:                  opt.AccountID,
			SecurityGroupID:            opt.SGMap[opt.CloudSGID].ID,
		}

		switch converter.PtrToVal(sgRule.Direction) {
		case armnetwork.SecurityRuleDirectionInbound:
			rule.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			rule.Type = enumor.Egress
		default:
		}

		if len(sgRule.DestinationApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(sgRule.DestinationApplicationSecurityGroups))
			for _, one := range sgRule.DestinationApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			rule.CloudDestinationAppSecurityGroupIDs = ids
		}

		if len(sgRule.SourceApplicationSecurityGroups) != 0 {
			ids := make([]*string, 0, len(sgRule.SourceApplicationSecurityGroups))
			for _, one := range sgRule.SourceApplicationSecurityGroups {
				ids = append(ids, one.ID)
			}
			rule.CloudSourceAppSecurityGroupIDs = ids
		}

		list = append(list, rule)
	}

	return list, nil
}

func (cli *client) updateSGRule(kt *kit.Kit, opt *syncSGRuleOption,
	updateMap map[string]securitygrouprule.AzureSGRule) error {

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

	updateReq := &protocloud.AzureSGRuleBatchUpdateReq{
		Rules: list,
	}
	err = cli.dbCli.Azure.SecurityGroup.BatchUpdateSecurityGroupRule(kt.Ctx, kt.Header(), updateReq,
		opt.SGMap[opt.CloudSGID].ID)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync sgRule to update sgRule success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) genUpdateRulesList(updateMap map[string]securitygrouprule.AzureSGRule,
	opt *syncSGRuleOption) ([]protocloud.AzureSGRuleUpdate, error) {

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	list := make([]protocloud.AzureSGRuleUpdate, 0)

	for id, rule := range updateMap {
		spec := protocloud.AzureSGRuleUpdate{
			ID:                         id,
			CloudID:                    converter.PtrToVal(rule.ID),
			Etag:                       rule.Etag,
			Name:                       converter.PtrToVal(rule.Name),
			Memo:                       rule.Description,
			DestinationAddressPrefix:   rule.DestinationAddressPrefix,
			DestinationAddressPrefixes: rule.DestinationAddressPrefixes,
			DestinationPortRange:       rule.DestinationPortRange,
			DestinationPortRanges:      rule.DestinationPortRanges,
			Protocol:                   string(converter.PtrToVal(rule.Protocol)),
			ProvisioningState:          string(converter.PtrToVal(rule.ProvisioningState)),
			SourceAddressPrefix:        rule.SourceAddressPrefix,
			SourceAddressPrefixes:      rule.SourceAddressPrefixes,
			SourcePortRange:            rule.SourcePortRange,
			SourcePortRanges:           rule.SourcePortRanges,
			Priority:                   converter.PtrToVal(rule.Priority),
			Access:                     string(converter.PtrToVal(rule.Access)),
			CloudSecurityGroupID:       opt.CloudSGID,
			Region:                     opt.SGMap[opt.CloudSGID].Region,
			AccountID:                  opt.AccountID,
			SecurityGroupID:            opt.SGMap[opt.CloudSGID].ID,
		}

		switch converter.PtrToVal(rule.Direction) {
		case armnetwork.SecurityRuleDirectionInbound:
			spec.Type = enumor.Ingress
		case armnetwork.SecurityRuleDirectionOutbound:
			spec.Type = enumor.Egress
		default:
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
		delCloudSGRuleIDMap[converter.PtrToVal(one.ID)] = false
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
			enumor.Azure, len(delSGRuleFromCloud), kt.Rid)
		return fmt.Errorf("validate sgRule not exist failed, before delete")
	}

	deleteReq := &protocloud.AzureSGRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	err = cli.dbCli.Azure.SecurityGroup.BatchDeleteSecurityGroupRule(kt.Ctx, kt.Header(), deleteReq,
		opt.SGMap[opt.CloudSGID].ID)
	if err != nil {
		logs.Errorf("[%s] dataservice delete azure security group rules failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sgRule to delete sgRule success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGRuleFromCloud(kt *kit.Kit, opt *syncSGRuleOption) ([]securitygrouprule.AzureSGRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgRuleopt := &securitygrouprule.AzureListOption{
		ResourceGroupName:    opt.ResourceGroupName,
		CloudSecurityGroupID: opt.CloudSGID,
	}

	rules, err := cli.cloudCli.ListSecurityGroupRule(kt, sgRuleopt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list azure security group rule failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return nil, err
	}

	results := make([]securitygrouprule.AzureSGRule, 0, len(rules))
	for _, one := range rules {
		results = append(results, converter.PtrToVal(one))
	}

	return results, nil
}

func (cli *client) listSGRuleFromDB(kt *kit.Kit, opt *syncSGRuleOption) ([]corecloud.AzureSecurityGroupRule, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if _, exsit := opt.SGMap[opt.CloudSGID]; !exsit {
		return nil, fmt.Errorf("cloud_sgid: %s can not find hcm sgid", opt.CloudSGID)
	}

	listReq := &protocloud.AzureSGRuleListReq{
		Filter: tools.EqualExpression("security_group_id", opt.SGMap[opt.CloudSGID].ID),
		Page:   core.NewDefaultBasePage(),
	}
	start := uint32(0)
	rules := make([]corecloud.AzureSecurityGroupRule, 0)
	for {
		listReq.Page.Start = start
		listResp, err := cli.dbCli.Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq,
			opt.SGMap[opt.CloudSGID].ID)
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

func isSGRuleChange(cloud securitygrouprule.AzureSGRule,
	db corecloud.AzureSecurityGroupRule) bool {

	if !assert.IsPtrStringEqual(db.Etag, cloud.Etag) {
		return true
	}

	if db.Name != converter.PtrToVal(cloud.Name) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Memo, cloud.Description) {
		return true
	}

	if db.Protocol != string(converter.PtrToVal(cloud.Protocol)) {
		return true
	}

	if db.ProvisioningState != string(converter.PtrToVal(cloud.ProvisioningState)) {
		return true
	}

	if db.Priority != converter.PtrToVal(cloud.Priority) {
		return true
	}

	if db.Access != string(converter.PtrToVal(cloud.Access)) {
		return true
	}

	destinationIDs := make([]*string, 0)
	if len(cloud.DestinationApplicationSecurityGroups) != 0 {
		for _, one := range cloud.DestinationApplicationSecurityGroups {
			destinationIDs = append(destinationIDs, one.ID)
		}
	}

	if !assert.IsPtrStringSliceEqual(db.CloudDestinationAppSecurityGroupIDs, destinationIDs) {
		return true
	}

	sourceIDs := make([]*string, 0)
	if len(cloud.SourceApplicationSecurityGroups) != 0 {
		for _, one := range cloud.SourceApplicationSecurityGroups {
			sourceIDs = append(sourceIDs, one.ID)
		}
	}

	if !assert.IsPtrStringSliceEqual(db.CloudSourceAppSecurityGroupIDs, sourceIDs) {
		return true
	}

	if isSGRuleSourceInfoChange(cloud, db) {
		return true
	}

	if isSGRuleDestinationInfoChange(cloud, db) {
		return true
	}

	return false
}

func isSGRuleSourceInfoChange(cloud securitygrouprule.AzureSGRule,
	db corecloud.AzureSecurityGroupRule) bool {

	if !assert.IsPtrStringEqual(db.SourceAddressPrefix, cloud.SourceAddressPrefix) {
		return true
	}

	if !assert.IsPtrStringSliceEqual(db.SourceAddressPrefixes, cloud.SourceAddressPrefixes) {
		return true
	}

	if !assert.IsPtrStringEqual(db.SourcePortRange, cloud.SourcePortRange) {
		return true
	}

	if !assert.IsPtrStringSliceEqual(db.SourcePortRanges, cloud.SourcePortRanges) {
		return true
	}
	return false
}

func isSGRuleDestinationInfoChange(cloud securitygrouprule.AzureSGRule,
	db corecloud.AzureSecurityGroupRule) bool {

	if !assert.IsPtrStringEqual(db.DestinationAddressPrefix, cloud.DestinationAddressPrefix) {
		return true
	}

	if !assert.IsPtrStringSliceEqual(db.DestinationAddressPrefixes, cloud.DestinationAddressPrefixes) {
		return true
	}

	if !assert.IsPtrStringEqual(db.DestinationPortRange, cloud.DestinationPortRange) {
		return true
	}

	if !assert.IsPtrStringSliceEqual(db.DestinationPortRanges, cloud.DestinationPortRanges) {
		return true
	}
	return false
}
