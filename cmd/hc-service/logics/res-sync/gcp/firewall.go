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

package gcp

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncFirewallOption ...
type SyncFirewallOption struct {
}

// Validate ...
func (opt SyncFirewallOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Firewall ...
func (cli *client) Firewall(kt *kit.Kit, params *SyncBaseParams, opt *SyncFirewallOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	firewallFromCloud, err := cli.listFirewallFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	firewallFromDB, err := cli.listFirewallFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(firewallFromCloud) == 0 && len(firewallFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[firewallrule.GcpFirewall, cloudcore.GcpFirewallRule](
		firewallFromCloud, firewallFromDB, isFirewallChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteFirewall(kt, params.AccountID, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createFirewall(kt, params.AccountID, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateFirewall(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createFirewall(kt *kit.Kit, accountID string, addSlice []firewallrule.GcpFirewall) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("firewall addSlice is <= 0, not create")
	}

	rulesCreate := make([]cloud.GcpFirewallRuleBatchCreate, 0)

	vpcSelfLinks := make([]string, 0)
	for _, one := range addSlice {
		vpcSelfLinks = append(vpcSelfLinks, one.Network)
	}

	opt := &QueryVpcsAndSyncOption{
		AccountID: accountID,
		SelfLink:  vpcSelfLinks,
	}
	vpcMap, err := cli.queryVpcsAndSync(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request QueryVpcIDsAndSync failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	for _, one := range addSlice {
		rule := cloud.GcpFirewallRuleBatchCreate{
			CloudID:               fmt.Sprint(one.Id),
			AccountID:             accountID,
			Name:                  one.Name,
			Priority:              one.Priority,
			Memo:                  one.Description,
			CloudVpcID:            vpcMap[one.Network].VpcCloudID,
			VpcSelfLink:           one.Network,
			VpcId:                 vpcMap[one.Network].VpcID,
			SourceRanges:          one.SourceRanges,
			BkBizID:               constant.UnassignedBiz,
			DestinationRanges:     one.DestinationRanges,
			SourceTags:            one.SourceTags,
			TargetTags:            one.TargetTags,
			SourceServiceAccounts: one.SourceServiceAccounts,
			TargetServiceAccounts: one.TargetServiceAccounts,
			Type:                  one.Direction,
			LogEnable:             one.LogConfig.Enable,
			Disabled:              one.Disabled,
			SelfLink:              one.SelfLink,
		}

		if len(one.Denied) != 0 {
			rule.Denied = firewallrule.ConvGcpDeniedProtoSets(one.Denied)
		}

		if len(one.Allowed) != 0 {
			rule.Allowed = firewallrule.ConvGcpAllowedProtoSets(one.Allowed)
		}

		rulesCreate = append(rulesCreate, rule)
	}

	batchCreateReq := &cloud.GcpFirewallRuleBatchCreateReq{
		FirewallRules: rulesCreate,
	}
	_, err = cli.dbCli.Gcp.Firewall.BatchCreateFirewallRule(kt.Ctx, kt.Header(), batchCreateReq)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync firewall to create firewall success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateFirewall(kt *kit.Kit, accountID string,
	updateMap map[string]firewallrule.GcpFirewall) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("firewall updateMap is <= 0, not update")
	}

	rulesUpdate := make([]cloud.GcpFirewallRuleBatchUpdate, 0)

	vpcSelfLinks := make([]string, 0)
	for _, one := range updateMap {
		vpcSelfLinks = append(vpcSelfLinks, one.Network)
	}

	opt := &QueryVpcsAndSyncOption{
		AccountID: accountID,
		SelfLink:  vpcSelfLinks,
	}
	vpcMap, err := cli.queryVpcsAndSync(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request QueryVpcIDsAndSync failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	for id, one := range updateMap {
		rule := cloud.GcpFirewallRuleBatchUpdate{
			ID:                    id,
			CloudID:               fmt.Sprint(one.Id),
			AccountID:             accountID,
			Name:                  one.Name,
			Priority:              one.Priority,
			Memo:                  one.Description,
			VpcSelfLink:           one.Network,
			CloudVpcID:            vpcMap[one.Network].VpcCloudID,
			VpcId:                 vpcMap[one.Network].VpcID,
			SourceRanges:          one.SourceRanges,
			BkBizID:               constant.UnassignedBiz,
			DestinationRanges:     one.DestinationRanges,
			SourceTags:            one.SourceTags,
			TargetTags:            one.TargetTags,
			SourceServiceAccounts: one.SourceServiceAccounts,
			TargetServiceAccounts: one.TargetServiceAccounts,
			Type:                  one.Direction,
			LogEnable:             one.LogConfig.Enable,
			Disabled:              one.Disabled,
			SelfLink:              one.SelfLink,
		}

		if len(one.Denied) != 0 {
			rule.Denied = firewallrule.ConvGcpDeniedProtoSets(one.Denied)
		}

		if len(one.Allowed) != 0 {
			rule.Allowed = firewallrule.ConvGcpAllowedProtoSets(one.Allowed)
		}

		rulesUpdate = append(rulesUpdate, rule)
	}

	batchCreateReq := &cloud.GcpFirewallRuleBatchUpdateReq{
		FirewallRules: rulesUpdate,
	}
	err = cli.dbCli.Gcp.Firewall.BatchUpdateFirewallRule(kt.Ctx, kt.Header(), batchCreateReq)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync firewall to update firewall success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteFirewall(kt *kit.Kit, accountID string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("firewall delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delFirewallFromCloud, err := cli.listFirewallFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFirewallFromCloud) > 0 {
		logs.Errorf("[%s] validate firewall not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delFirewallFromCloud), kt.Rid)
		return fmt.Errorf("validate firewall not exist failed, before delete")
	}

	deleteReq := &cloud.GcpFirewallRuleBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Gcp.Firewall.BatchDeleteFirewallRule(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete firewall failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync firewall to delete firewall success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listFirewallFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]firewallrule.GcpFirewall, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &firewallrule.ListOption{
		CloudIDs: converter.StringSliceToUint64Slice(params.CloudIDs),
	}
	result, _, err := cli.cloudCli.ListFirewallRule(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list firewall from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listFirewallFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.GcpFirewallRule, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &cloud.GcpFirewallRuleListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list firewall from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveFirewallDeleteFromCloud remove delete from cloud
func (cli *client) RemoveFirewallDeleteFromCloud(kt *kit.Kit, accountID string) error {
	req := &cloud.GcpFirewallRuleListReq{
		Field: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list firewall failed, err: %v, req: %v, rid: %s", enumor.Gcp,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listFirewallFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, fmt.Sprint(one.Id))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteFirewall(kt, accountID, cloudIDs); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func isFirewallChange(cloud firewallrule.GcpFirewall, db cloudcore.GcpFirewallRule) bool {

	if db.Name != cloud.Name {
		return true
	}

	if db.CloudID != fmt.Sprint(cloud.Id) {
		return true
	}

	if db.VpcSelfLink != cloud.Network {
		return true
	}

	if db.Priority != cloud.Priority {
		return true
	}

	if db.Memo != cloud.Description {
		return true
	}

	if !assert.IsStringSliceEqual(db.SourceRanges, cloud.SourceRanges) {
		return true
	}

	if !assert.IsStringSliceEqual(db.DestinationRanges, cloud.DestinationRanges) {
		return true
	}

	if !assert.IsStringSliceEqual(db.SourceTags, cloud.SourceTags) {
		return true
	}

	if !assert.IsStringSliceEqual(db.TargetTags, cloud.TargetTags) {
		return true
	}

	if !assert.IsStringSliceEqual(db.SourceServiceAccounts, cloud.SourceServiceAccounts) {
		return true
	}

	if !assert.IsStringSliceEqual(db.TargetServiceAccounts, cloud.TargetServiceAccounts) {
		return true
	}

	if db.Type != cloud.Direction {
		return true
	}

	if db.LogEnable != cloud.LogConfig.Enable {
		return true
	}

	if db.Disabled != cloud.Disabled {
		return true
	}

	if db.SelfLink != cloud.SelfLink {
		return true
	}

	return false
}

func (cli *client) queryVpcsAndSync(kt *kit.Kit, opt *QueryVpcsAndSyncOption) (map[string]*common.VpcDB, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcSelfLinks := slice.Unique(opt.SelfLink)

	listParams := &ListBySelfLinkOption{
		AccountID: opt.AccountID,
		SelfLink:  vpcSelfLinks,
	}
	result, err := cli.listVpcFromDBBySelfLink(kt, listParams)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, selfLinks: %v, rid: %s", err, vpcSelfLinks, kt.Rid)
		return nil, err
	}

	// 如果相等，则Vpc全部同步到了db
	if len(result) == len(vpcSelfLinks) {
		return convVpcMap(result), nil
	}

	existVpcSLMap := convVpcSLMap(result)

	notExistSelfLink := make([]string, 0)
	for _, selfLink := range vpcSelfLinks {
		if _, exist := existVpcSLMap[selfLink]; !exist {
			notExistSelfLink = append(notExistSelfLink, selfLink)
		}
	}

	listVpc := &ListBySelfLinkOption{
		AccountID: opt.AccountID,
		SelfLink:  notExistSelfLink,
	}
	vpcs, err := cli.listVpcFromCloudBySelfLink(kt, listVpc)
	if err != nil {
		logs.Errorf("list vpc from cloud by self link failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(vpcs))
	for _, one := range vpcs {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	params := &SyncBaseParams{
		AccountID: opt.AccountID,
		CloudIDs:  cloudIDs,
	}
	if _, err = cli.Vpc(kt, params, new(SyncVpcOption)); err != nil {
		return nil, err
	}

	// 同步完，二次查询
	listCloudIDParams := &ListBySelfLinkOption{
		AccountID: opt.AccountID,
		SelfLink:  vpcSelfLinks,
	}
	secondResult, err := cli.listVpcFromDBBySelfLink(kt, listCloudIDParams)
	if err != nil {
		logs.Errorf("list vpc from db by self link failed, err: %v, cloudIDs: %v, rid: %s", err, cloudIDs, kt.Rid)
		return nil, err
	}

	if len(secondResult) != len(vpcSelfLinks) {
		return nil, fmt.Errorf("some vpc can not sync, self_links: %v", notExistSelfLink)
	}

	return convVpcMap(secondResult), nil
}

func convVpcSLMap(result []cloudcore.Vpc[cloudcore.GcpVpcExtension]) map[string]string {
	m := make(map[string]string, len(result))
	for _, one := range result {
		m[one.Extension.SelfLink] = one.ID
	}
	return m
}

func convVpcMap(result []cloudcore.Vpc[cloudcore.GcpVpcExtension]) map[string]*common.VpcDB {
	m := make(map[string]*common.VpcDB)
	for _, one := range result {
		m[one.Extension.SelfLink] = &common.VpcDB{
			VpcCloudID: one.CloudID,
			VpcID:      one.ID,
		}
	}
	return m
}
