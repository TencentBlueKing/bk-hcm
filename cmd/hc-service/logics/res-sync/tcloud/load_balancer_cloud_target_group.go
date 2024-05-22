/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"hcm/cmd/hc-service/logics/res-sync/common"
	typescore "hcm/pkg/adaptor/types/core"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// CloudTargetGroup 同步指定负载均衡云目标组
func (cli *client) CloudTargetGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudTgs, err := cli.listTargetGroupFromCloud(kt, params)
	if err != nil {
		logs.Errorf("fail to list target group from cloud for sync, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dbTgs, err := cli.listTargetGroupFromDB(kt, params)
	if err != nil {
		logs.Errorf("fail to list target group from database for sync, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudTgs) == 0 && len(dbTgs) == 0 {
		return new(SyncResult), nil
	}

	// 比较基本信息
	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TargetGroup, corelb.TCloudTargetGroup](cloudTgs, dbTgs,
		isTGChange)

	// 删除云上已经删除的负载均衡实例
	if err = cli.deleteTargetGroup(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
		return nil, err
	}

	// 创建云上新增负载均衡实例
	_, err = cli.createTargetGroup(kt, params.AccountID, params.Region, addSlice)
	if err != nil {
		return nil, err
	}
	// 更新变更负载均衡
	if err = cli.updateTargetGroup(kt, updateMap); err != nil {
		return nil, err
	}

	// 同步目标组下的rs

	if err = cli.batchCloudTargetGroupRS(kt, params, opt); err != nil {
		logs.Errorf("fail to sync cloud target group rs, err:%v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

// RemoveTargetGroupDeleteFromCloud ...
func (cli *client) RemoveTargetGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("region", region),
			tools.RuleEqual("target_group_type", enumor.CloudTargetGroupType),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		tgFromDB, err := cli.dbCli.Global.LoadBalancer.ListTargetGroup(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list target group failed, err: %v, req: %v, rid: %s",
				enumor.TCloud, err, req, kt.Rid)
			return err
		}

		cloudIDs := slice.Map(tgFromDB.Details, func(tg corelb.BaseTargetGroup) string { return tg.CloudID })
		if len(cloudIDs) == 0 {
			break
		}

		var delCloudIDs []string
		params := &SyncBaseParams{AccountID: accountID, Region: region, CloudIDs: cloudIDs}
		delCloudIDs, err = cli.listRemovedTargetGroupID(kt, params)
		if err != nil {
			return err
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteTargetGroup(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(tgFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}
	return nil
}

func (cli *client) listRemovedTargetGroupID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	batchParam := &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
	}
	tgMap := cvt.StringSliceToMap(params.CloudIDs)

	for _, batchCloudID := range slice.Split(params.CloudIDs, constant.TCLBDescribeMax) {
		batchParam.CloudIDs = batchCloudID
		found, err := cli.listTargetGroupFromCloud(kt, batchParam)
		if err != nil {
			return nil, err
		}
		for _, tg := range found {
			delete(tgMap, tg.GetCloudID())
		}
	}

	return cvt.MapKeyToSlice(tgMap), nil
}

func (cli *client) listTargetGroupFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typeslb.TargetGroup, error) {
	opt := &typeslb.ListTargetGroupOption{
		Region:         params.Region,
		TargetGroupIds: params.CloudIDs,
	}
	return cli.cloudCli.ListTargetGroup(kt, opt)

}

func (cli *client) listTargetGroupFromDB(kt *kit.Kit, params *SyncBaseParams) ([]corelb.TCloudTargetGroup, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", params.AccountID),
			tools.RuleEqual("target_group_type", enumor.CloudTargetGroupType),
			tools.RuleEqual("region", params.Region)),
		Page: core.NewDefaultBasePage(),
	}
	if len(params.CloudIDs) > 0 {
		listReq.Filter.Rules = append(listReq.Filter.Rules, tools.RuleIn("cloud_id", params.CloudIDs))
	}
	groupsResp, err := cli.dbCli.TCloud.LoadBalancer.ListTargetGroup(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list target group for sync, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return groupsResp.Details, nil
}

// 删除目标组
func (cli *client) deleteTargetGroup(kt *kit.Kit, accountID string, region string, cloudIDs []string) error {

	if len(cloudIDs) == 0 {
		return nil
	}
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("region", region),
			tools.RuleIn("cloud_id", cloudIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	return cli.dbCli.Global.LoadBalancer.DeleteTargetGroup(kt, listReq)
}

// 创建目标组
func (cli *client) createTargetGroup(kt *kit.Kit, accountID string, region string,
	addSlice []typeslb.TargetGroup) ([]string, error) {

	if len(addSlice) == 0 {
		return []string{}, nil
	}
	groups := make([]dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension], 0, len(addSlice))
	for _, cloud := range addSlice {
		groups = append(groups, dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			Name:      cvt.PtrToVal(cloud.TargetGroupName),
			CloudID:   cvt.PtrToVal(cloud.TargetGroupId),
			Vendor:    enumor.TCloud,
			AccountID: accountID,
			BkBizID:   constant.UnassignedBiz,
			Region:    region,
			// 云上目标组没有协议
			Protocol:        enumor.HttpProtocol,
			Port:            int64(cvt.PtrToVal(cloud.Port)),
			CloudVpcID:      cvt.PtrToVal(cloud.VpcId),
			TargetGroupType: enumor.CloudTargetGroupType,
		})
	}
	createReq := &dataproto.TCloudTargetGroupCreateReq{TargetGroups: groups}
	createResult, err := cli.dbCli.TCloud.LoadBalancer.BatchCreateTCloudTargetGroup(kt, createReq)
	if err != nil {
		logs.Errorf("fail to create cloud target group, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return createResult.IDs, nil
}

// 更新目标组
func (cli *client) updateTargetGroup(kt *kit.Kit, updateMap map[string]typeslb.TargetGroup) error {

	if len(updateMap) == 0 {
		return nil
	}
	updates := make([]*dataproto.TargetGroupExtUpdateReq[corelb.TCloudTargetGroupExtension], 0, len(updateMap))
	for localID, cloudInfo := range updateMap {
		updates = append(updates, &dataproto.TargetGroupExtUpdateReq[corelb.TCloudTargetGroupExtension]{
			ID:         localID,
			Name:       cvt.PtrToVal(cloudInfo.TargetGroupName),
			CloudVpcID: cvt.PtrToVal(cloudInfo.VpcId),
			Port:       int64(cvt.PtrToVal(cloudInfo.Port)),
		})
	}
	updateReq := &dataproto.TCloudTargetGroupBatchUpdateReq{TargetGroups: updates}
	return cli.dbCli.TCloud.LoadBalancer.BatchUpdateTargetGroup(kt, updateReq)
}

func (cli *client) batchCloudTargetGroupRS(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) error {
	dbTgs, err := cli.listTargetGroupFromDB(kt, params)
	if err != nil {
		logs.Errorf("fail to list target group from database for sync rs, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, tg := range dbTgs {
		cloudInstList, err := cli.listInstancesFromCloud(kt, params.Region, tg.CloudID)
		if err != nil {
			logs.Errorf("fail to list instance from cloud for sync ,err: %v, rid: %s", err, kt.Rid)
			return err
		}
		dbTargetList, err := cli.listInstancesFromDB(kt, tg.ID)
		if err != nil {
			logs.Errorf("fail to list rs from db for sync, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TargetGroupBackend, corelb.BaseTarget](cloudInstList,
			dbTargetList, isCloudRSChange)

		if err = cli.deleteCloudRs(kt, delCloudIDs); err != nil {
			logs.Errorf("fail to delete cloud rs for sync, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if _, err := cli.createCloudRs(kt, params.AccountID, tg.ID, addSlice); err != nil {
			logs.Errorf("fail to create cloud rs for sync, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if err := cli.updateCloudRs(kt, updateMap); err != nil {
			logs.Errorf("fail to update cloud rs for sync, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// 按cloudInstID 删除目标组中的rs
func (cli *client) deleteCloudRs(kt *kit.Kit, cloudIDs []string) error {
	if len(cloudIDs) == 0 {
		return nil
	}

	delReq := &dataproto.LoadBalancerBatchDeleteReq{Filter: tools.ContainersExpression("cloud_inst_id", cloudIDs)}
	err := cli.dbCli.Global.LoadBalancer.BatchDeleteTarget(kt, delReq)
	if err != nil {
		logs.Errorf("fail to delete cloud rs (ids=%v), err: %v, rid: %s", cloudIDs, err, kt.Rid)
		return err
	}

	return nil
}
func (cli *client) createCloudRs(kt *kit.Kit, accountID, tgId string, addSlice []typeslb.TargetGroupBackend) (
	[]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}

	var targets []*dataproto.TargetBaseReq
	for _, backend := range addSlice {
		dbVal := &dataproto.TargetBaseReq{
			InstType:           cvt.PtrToVal((*enumor.InstType)(backend.Type)),
			CloudInstID:        cvt.PtrToVal(backend.InstanceId),
			Port:               int64(cvt.PtrToVal(backend.Port)),
			AccountID:          accountID,
			TargetGroupID:      tgId,
			CloudTargetGroupID: cvt.PtrToVal(backend.TargetGroupId),
		}
		if backend.Weight != nil {
			dbVal.Weight = cvt.ValToPtr((int64)(cvt.PtrToVal(backend.Weight)))
		}
		targets = append(targets, dbVal)

	}

	created, err := cli.dbCli.Global.LoadBalancer.BatchCreateTCloudTarget(kt,
		&dataproto.TargetBatchCreateReq{Targets: targets})
	if err != nil {
		logs.Errorf("fail to create target for target group syncing, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return created.IDs, nil
}

// 更新rs中的信息
func (cli *client) updateCloudRs(kt *kit.Kit, updateMap map[string]typeslb.TargetGroupBackend) (err error) {

	if len(updateMap) == 0 {
		return nil
	}
	updates := make([]*dataproto.TargetUpdate, 0, len(updateMap))
	for id, backend := range updateMap {
		updates = append(updates, &dataproto.TargetUpdate{
			ID:               id,
			Port:             int64(cvt.PtrToVal(backend.Port)),
			Weight:           cvt.ValToPtr((int64)(cvt.PtrToVal(backend.Weight))),
			PrivateIPAddress: cvt.PtrToSlice(backend.PrivateIpAddresses),
			PublicIPAddress:  cvt.PtrToSlice(backend.PublicIpAddresses),
			InstName:         cvt.PtrToVal(backend.InstanceName),
		})
	}
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: updates}
	if err = cli.dbCli.Global.LoadBalancer.BatchUpdateTarget(kt, updateReq); err != nil {
		logs.Errorf("fail to update targets while syncing, err: %v, rid:%s", err, kt.Rid)
	}

	return err
}

func (cli *client) listInstancesFromDB(kt *kit.Kit, tgID string) ([]corelb.BaseTarget, error) {
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	var allTargets []corelb.BaseTarget
	for {
		targetResp, err := cli.dbCli.Global.LoadBalancer.ListTarget(kt, listReq)
		if err != nil {
			return nil, err
		}
		if len(targetResp.Details) == 0 {
			break
		}
		allTargets = append(allTargets, targetResp.Details...)
		listReq.Page.Start += uint32(len(targetResp.Details))
	}
	return allTargets, nil
}

func (cli *client) listInstancesFromCloud(kt *kit.Kit, region, tgCloudID string) ([]typeslb.TargetGroupBackend, error) {
	req := &typeslb.ListTargetGroupInstanceOption{
		Region:         region,
		TargetGroupIds: []string{tgCloudID},
		Page: &typescore.TCloudPage{
			Offset: 0,
			Limit:  typescore.TCloudQueryLimit,
		},
	}
	allInstList := make([]typeslb.TargetGroupBackend, 0)
	for {
		instList, err := cli.cloudCli.ListTargetGroupInstance(kt, req)
		if err != nil {
			logs.Errorf("fail to list target group instance form cloud, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(instList) == 0 {
			break
		}
		req.Page.Offset += uint64(len(instList))
		allInstList = append(allInstList, instList...)
	}
	return allInstList, nil
}

func isTGChange(cloudTG typeslb.TargetGroup, localTG corelb.TCloudTargetGroup) bool {

	if cvt.PtrToVal(cloudTG.TargetGroupName) != localTG.Name {
		return true
	}

	if cvt.PtrToVal(cloudTG.VpcId) != localTG.CloudVpcID {
		return true
	}

	if cvt.PtrToVal(cloudTG.Port) != uint64(localTG.Port) {
		return true
	}

	return false
}

// 判断rs信息是否变化
func isCloudRSChange(cloud typeslb.TargetGroupBackend, db corelb.BaseTarget) bool {
	if cvt.PtrToVal(cloud.Port) != uint64(db.Port) {
		return true
	}

	if cvt.PtrToVal(cloud.Weight) != uint64(cvt.PtrToVal(db.Weight)) {
		return true
	}
	if cvt.PtrToVal(cloud.InstanceName) != db.InstName {
		return true
	}

	if !assert.IsStringSliceEqual(cvt.PtrToSlice(cloud.PrivateIpAddresses), db.PrivateIPAddress) {
		return true
	}

	if !assert.IsStringSliceEqual(cvt.PtrToSlice(cloud.PublicIpAddresses), db.PublicIPAddress) {
		return true
	}
	return false
}
