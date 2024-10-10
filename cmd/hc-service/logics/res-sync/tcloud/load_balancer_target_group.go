/*
 *
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
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// LocalTargetGroup 同步本地目标组
func (cli *client) LocalTargetGroup(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOption,
	cloudListeners []typeslb.TCloudListener) error {

	// 目前主要是同步健康检查
	healthMap := make(map[string]*tclb.HealthCheck, len(cloudListeners))
	cloudIDs := make([]string, 0, len(cloudListeners))
	// 收集云端健康检查
	for _, listener := range cloudListeners {
		if !listener.GetProtocol().IsLayer7Protocol() {
			// 四层监听器，直接获取健康检查
			healthMap[listener.GetCloudID()] = listener.HealthCheck
			cloudIDs = append(cloudIDs, listener.GetCloudID())
			continue
		}
		for _, rule := range listener.Rules {
			healthMap[cvt.PtrToVal(rule.LocationId)] = rule.HealthCheck
			cloudIDs = append(cloudIDs, cvt.PtrToVal(rule.LocationId))
		}
	}
	if len(cloudIDs) == 0 {
		// 只有7层监听器，且所有监听器都为空的情况会导致这里为空
		logs.Infof("no listener found for %v, rid: %s", cloudListeners, kt.Rid)
		return nil
	}
	tgCloudHealthMap, tgList, err := cli.getTargetGruop(kt, opt.LBID, cloudIDs, healthMap)
	if err != nil {
		logs.Errorf("fail to get local target group for sync, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	for _, tg := range tgList {

		if !isHealthCheckChange(tgCloudHealthMap[tg.CloudID], tg.HealthCheck, false) {
			continue
		}

		// 更新 健康检查
		updateReq := &dataproto.TargetGroupUpdateReq{
			IDs:         []string{tg.ID},
			HealthCheck: convHealthCheck(tgCloudHealthMap[tg.CloudID]),
		}
		err = cli.dbCli.TCloud.LoadBalancer.BatchUpdateTCloudTargetGroup(kt, updateReq)
		if err != nil {
			logs.Errorf("fail to update target group health check during sync, err: %v, rid: %s", err, kt.Rid)
			return err
		}

	}
	return nil
}

func (cli *client) getTargetGruop(kt *kit.Kit, lbId string, cloudIDs []string,
	healthMap map[string]*tclb.HealthCheck) (map[string]*tclb.HealthCheck, []corelb.BaseTargetGroup, error) {

	// 查找本地 目标组
	relReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbId),
			tools.RuleIn("cloud_listener_rule_id", cloudIDs)),
		Page: core.NewDefaultBasePage(),
	}
	relResp, err := cli.dbCli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, relReq)
	if err != nil {
		logs.Errorf("fail to get target group rel for sync, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	if len(relResp.Details) == 0 {
		return nil, nil, nil
	}

	tgIds := make([]string, 0, len(relResp.Details))
	tgCloudHealthMap := make(map[string]*tclb.HealthCheck, len(relResp.Details))
	for _, rel := range relResp.Details {
		tgIds = append(tgIds, rel.TargetGroupID)
		tgCloudHealthMap[rel.TargetGroupID] = healthMap[rel.CloudListenerRuleID]
	}
	// 查找目标组
	tgReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", tgIds),
		Page:   core.NewDefaultBasePage(),
	}
	tgResp, err := cli.dbCli.Global.LoadBalancer.ListTargetGroup(kt, tgReq)
	if err != nil {
		logs.Errorf("fail to get target group for sync, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	return tgCloudHealthMap, tgResp.Details, nil
}

// ListenerTargets 监听器下的target，用来更新目标组.
// SyncBaseParams 中的CloudID作为监听器id筛选，不传的话就是同步当前LB下的全部监听器
func (cli *client) ListenerTargets(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOption) error {

	cloudListenerTargets, relMap, tgRsMap, lb, err := cli.listTargetRelated(kt, param, opt)
	if err != nil {
		logs.Errorf("fail to list related res during targets syncing, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 一个目标组只处理一次
	isTGHandled := genExists[string]()
	compareWrapper := func(rel *corelb.BaseTargetListenerRuleRel, cloudTargets []*tclb.Backend) error {
		if rel.BindingStatus == enumor.BindingBindingStatus {
			return nil
		}
		tgId := rel.TargetGroupID
		if isTGHandled(tgId) {
			return nil
		}

		// 存在则比较
		return cli.compareTargetsChange(kt, param.AccountID, tgId, cloudTargets, tgRsMap[tgId])
	}
	// 遍历云上的监听器、规则
	for _, listener := range cloudListenerTargets {
		if !listener.GetProtocol().IsLayer7Protocol() {
			// ---- for layer 4 对比监听器变化 ----
			rel, exists := relMap[cvt.PtrToVal(listener.ListenerId)]
			if !exists {
				// 云上监听器、但是没有对应目标组，则在同步时自动创建目标组，并将RS加入目标组。
				if err := cli.createLocalTargetGroupL4(kt, opt, lb, listener); err != nil {
					logs.Errorf("fail to create local target group for layer 4 listener, rid: %s", kt.Rid)
					return err
				}
				// 只要本地没有目标组就跳过RS同步
				continue
			}
			if err := compareWrapper(rel, listener.Targets); err != nil {
				logs.Errorf("fail to compare L4 listener rs change, err: %v, rid:%s", err, kt.Rid)
				return err
			}
			continue
		}
		// ---- for layer 7 对比规则变化 ----
		for _, rule := range listener.Rules {
			rel, exists := relMap[cvt.PtrToVal(rule.LocationId)]
			if !exists {
				// 没有对应目标组关系，则在同步时自动创建目标组，并将RS加入目标组。
				if err := cli.createLocalTargetGroupL7(kt, opt, lb, listener, rule); err != nil {
					logs.Errorf("fail to create local target group for layer 7 rule, rid: %s", kt.Rid)
					return err
				}
				// 跳过比较
				continue
			}
			// 存在则比较
			if err := compareWrapper(rel, rule.Targets); err != nil {
				logs.Errorf("fail to compare L7 rule rs change, err: %v, rid:%s", err, kt.Rid)
				return err
			}
		}
	}
	return nil
}

// 获取同步rs所需关联资源
func (cli *client) listTargetRelated(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOption) (
	[]typeslb.TCloudListenerTarget, map[string]*corelb.BaseTargetListenerRuleRel,
	map[string][]corelb.BaseTarget, *corelb.TCloudLoadBalancer, error) {

	// 获取监听器详情
	cloudListenerTargets, err := cli.listTargetsFromCloud(kt, param, opt)
	if err != nil {
		logs.Errorf("fail to list target from cloud while syncing, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, err
	}

	// 获取db中的目标组关系和rs列表
	relMap, tgRsMap, err := cli.listTargetsFromDB(kt, param, opt)
	if err != nil {
		logs.Errorf("fail to list target from db while syncing, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, err
	}
	if opt.CachedLoadBalancer != nil && opt.CachedLoadBalancer.CloudID == opt.CloudLBID {
		return cloudListenerTargets, relMap, tgRsMap, opt.CachedLoadBalancer, nil
	}

	logs.Errorf("cache for target sync mismatch opt: %+v, rid:%s ", opt, kt.Rid)

	lbResp, err := cli.listLBFromDB(kt, &SyncBaseParams{
		AccountID: param.AccountID,
		Region:    param.Region,
		CloudIDs:  []string{opt.CloudLBID},
	})
	if err != nil {
		logs.Errorf("fail to list lb from db for sync tg, err: %v, lb_id: %s, rid: %s", err, opt.CloudLBID, kt.Rid)
		return nil, nil, nil, nil, err
	}
	if len(lbResp) == 0 {
		logs.Errorf("can not find lb for sync tg, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, nil, nil, nil, errf.Newf(errf.RecordNotFound, "lb not found: %s", opt.CloudLBID)
	}

	return cloudListenerTargets, relMap, tgRsMap, cvt.ValToPtr(lbResp[0]), nil
}

func (cli *client) compareTargetsChange(kt *kit.Kit, accountID, tgID string, cloudTargets []*tclb.Backend,
	dbRsList []corelb.BaseTarget) (err error) {

	// 增加包裹类型
	cloudRsList := slice.Map(cloudTargets, func(rs *tclb.Backend) typeslb.Backend {
		return typeslb.Backend{Backend: rs}
	})
	addSlice, updateMap, delLocalIDs := diff[typeslb.Backend, corelb.BaseTarget](cloudRsList, dbRsList, isRsChange)

	if err = cli.deleteRs(kt, delLocalIDs); err != nil {
		return err
	}

	if err = cli.updateRs(kt, updateMap); err != nil {
		return err
	}
	if _, err = cli.createRs(kt, accountID, tgID, addSlice); err != nil {
		return err
	}

	return nil
}

// 为rs创建目标组不跳过没有rs的规则
func (cli *client) createLocalTargetGroupL7(kt *kit.Kit, opt *SyncListenerOption,
	lb *corelb.TCloudLoadBalancer, listener typeslb.TCloudListenerTarget, cloudRule *tclb.RuleTargets) error {

	// 获取数据库中的规则
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_id", cloudRule.LocationId),
			tools.RuleEqual("cloud_lbl_id", listener.ListenerId)),
		Page: core.NewDefaultBasePage(),
	}

	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of l7 listener, err: %v, rule id: %s, lbl_cloud_id: %s, rid: %s",
			err, cloudRule.LocationId, listener.ListenerId, kt.Rid)
		return err
	}
	if len(ruleResp.Details) == 0 {
		logs.Errorf("rule of listener can not be found by id(%s),err: %v, lbl_cloud_id: %s, rid: %s ",
			err, cloudRule.LocationId, listener.ListenerId, kt.Rid)
		return fmt.Errorf("rule of listener can not be found by id(%+v)", cloudRule.LocationId)
	}
	dbRule := ruleResp.Details[0]
	healthcheck, err := json.MarshalToString(dbRule.HealthCheck)
	if err != nil {
		logs.Errorf("fail to marshal rule health check to string, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	tgCreate := dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{
		TargetGroup: dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			Name:            genTargetGroupNameL7(dbRule),
			Vendor:          enumor.TCloud,
			AccountID:       lb.AccountID,
			BkBizID:         opt.BizID,
			Region:          lb.Region,
			Protocol:        listener.GetProtocol(),
			Port:            cvt.PtrToVal(listener.Port),
			VpcID:           lb.VpcID,
			CloudVpcID:      lb.CloudVpcID,
			TargetGroupType: enumor.LocalTargetGroupType,
			Weight:          0,
			HealthCheck:     types.JsonField(healthcheck),
			Memo:            cvt.ValToPtr("auto created for rule " + cvt.PtrToVal(cloudRule.LocationId)),
			RsList:          slice.Map(cloudRule.Targets, convTarget(lb.AccountID)),
		},
		ListenerRuleID:      dbRule.ID,
		CloudListenerRuleID: dbRule.CloudID,
		ListenerRuleType:    enumor.Layer7RuleType,
		LbID:                dbRule.LbID,
		CloudLbID:           dbRule.CloudLbID,
		LblID:               dbRule.LblID,
		CloudLblID:          dbRule.CloudLBLID,
		BindingStatus:       enumor.SuccessBindingStatus,
	}

	tgCreateReq := &dataproto.TCloudBatchCreateTgWithRelReq{
		TargetGroups: []dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{tgCreate},
	}
	_, err = cli.dbCli.TCloud.LoadBalancer.BatchCreateTargetGroupWithRel(kt, tgCreateReq)
	if err != nil {
		logs.Errorf("fail to create tcloud target group with rel, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func convTarget(accountID string) func(cloudTarget *tclb.Backend) *dataproto.TargetBaseReq {
	return func(cloudTarget *tclb.Backend) *dataproto.TargetBaseReq {
		localTarget := typeslb.Backend{Backend: cloudTarget}
		target := &dataproto.TargetBaseReq{
			InstName:         cvt.PtrToVal(cloudTarget.InstanceName),
			InstType:         cvt.PtrToVal((*enumor.InstType)(cloudTarget.Type)),
			CloudInstID:      cvt.PtrToVal(cloudTarget.InstanceId),
			Port:             cvt.PtrToVal(cloudTarget.Port),
			Weight:           cloudTarget.Weight,
			AccountID:        accountID,
			PrivateIPAddress: cvt.PtrToSlice(cloudTarget.PrivateIpAddresses),
			PublicIPAddress:  cvt.PtrToSlice(cloudTarget.PublicIpAddresses),
			IP:               localTarget.GetIP(),
		}
		if enumor.InstType(cvt.PtrToVal(cloudTarget.Type)) == enumor.CcnInstType {
			target.CloudInstID = localTarget.GetCloudID()
		}
		if enumor.InstType(cvt.PtrToVal(cloudTarget.Type)) == enumor.EniInstType {
			target.CloudInstID = cvt.PtrToVal(localTarget.EniId)
		}
		return target
	}
}

// 创建本地目标组以及关系，不会跳过没有rs的监听器
func (cli *client) createLocalTargetGroupL4(kt *kit.Kit, opt *SyncListenerOption,
	lb *corelb.TCloudLoadBalancer, listener typeslb.TCloudListenerTarget) error {

	lbl, rule, err := cli.listListenerWithRule(kt, cvt.PtrToVal(listener.ListenerId))
	if err != nil {
		logs.Errorf("fail to list listener with rule, err: %v, rid:%s", err, kt.Rid)
		return err
	}

	healthcheck, err := json.MarshalToString(rule.HealthCheck)
	if err != nil {
		logs.Errorf("fail to marshal rule health check to string, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	tgCreate := dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{
		TargetGroup: dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			Name:            genTargetGroupNameL4(lbl),
			Vendor:          enumor.TCloud,
			AccountID:       lbl.AccountID,
			BkBizID:         lbl.BkBizID,
			Region:          lb.Region,
			Protocol:        lbl.Protocol,
			Port:            lbl.Port,
			VpcID:           lb.VpcID,
			CloudVpcID:      lb.CloudVpcID,
			TargetGroupType: enumor.LocalTargetGroupType,
			Weight:          0,
			HealthCheck:     types.JsonField(healthcheck),
			Memo:            cvt.ValToPtr("auto created for listener " + cvt.PtrToVal(listener.ListenerId)),
			RsList:          slice.Map(listener.Targets, convTarget(lb.AccountID)),
		},
		// 需要用4层对应的规则id
		ListenerRuleID:      rule.ID,
		CloudListenerRuleID: lbl.CloudID,
		ListenerRuleType:    enumor.Layer4RuleType,
		LbID:                lbl.LbID,
		CloudLbID:           lbl.CloudLbID,
		LblID:               lbl.ID,
		CloudLblID:          lbl.CloudID,
		BindingStatus:       enumor.SuccessBindingStatus,
	}

	tgCreateReq := &dataproto.TCloudBatchCreateTgWithRelReq{
		TargetGroups: []dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{tgCreate},
	}
	_, err = cli.dbCli.TCloud.LoadBalancer.BatchCreateTargetGroupWithRel(kt, tgCreateReq)
	if err != nil {
		logs.Errorf("fail to create tcloud target group with rel, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (cli *client) listListenerWithRule(kt *kit.Kit, listenerCloudID string) (
	*corelb.Listener[corelb.TCloudListenerExtension], *corelb.TCloudLbUrlRule, error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", listenerCloudID),
		Page:   core.NewDefaultBasePage(),
	}
	lblResp, err := cli.dbCli.TCloud.LoadBalancer.ListListener(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list listener of lb(%s) for create local target group, err: %v, rid: %s",
			listenerCloudID, err, kt.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) == 0 {
		// 出现云上新增的监听器，本地没有的，跳过, 等待下次同步
		logs.Errorf("listener can not be found by id(%s) while target group sync, rid: %s",
			listenerCloudID, kt.Rid)
		return nil, nil, fmt.Errorf("listener can not be found by id(%s) while target group syncing", listenerCloudID)
	}
	lbl := lblResp.Details[0]
	// 获取对应规则
	listReq.Filter = tools.ExpressionAnd(
		tools.RuleEqual("cloud_id", lbl.CloudID),
		tools.RuleEqual("lbl_id", lbl.ID))
	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of l4 listener, err: %v, lbl_id: %s, lbl_cloud_id: %s, rid: %s",
			err, lbl.ID, lbl.CloudID, kt.Rid)
		return nil, nil, err
	}
	if len(ruleResp.Details) == 0 {
		logs.Errorf("rule of listener can not be found by id(%s), lbl_id: %s, lbl_cloud_id: %s, rid: %s ",
			lbl.ID, lbl.CloudID, kt.Rid)
		return nil, nil, fmt.Errorf("rule of listener  can not be found by id(%s) while target group syncing",
			listenerCloudID)
	}
	return cvt.ValToPtr(lbl), cvt.ValToPtr(ruleResp.Details[0]), nil
}

// 按cloudInstID 删除目标组中的rs
func (cli *client) deleteRs(kt *kit.Kit, localIds []string) error {
	if len(localIds) == 0 {
		return nil
	}
	for _, idBatch := range slice.Split(localIds, constant.BatchOperationMaxLimit) {
		delReq := &dataproto.LoadBalancerBatchDeleteReq{Filter: tools.ContainersExpression("id", idBatch)}
		err := cli.dbCli.Global.LoadBalancer.BatchDeleteTarget(kt, delReq)
		if err != nil {
			logs.Errorf("fail to delete rs, err: %v, ids: %v, rid: %s", err, idBatch, kt.Rid)
			return err
		}
	}

	return nil
}

// 更新rs中的信息
func (cli *client) updateRs(kt *kit.Kit, updateMap map[string]typeslb.Backend) (err error) {

	if len(updateMap) == 0 {
		return nil
	}
	updates := make([]*dataproto.TargetUpdate, 0, len(updateMap))
	for id, backend := range updateMap {
		updates = append(updates, &dataproto.TargetUpdate{
			ID:               id,
			Port:             cvt.PtrToVal(backend.Port),
			Weight:           backend.Weight,
			PrivateIPAddress: cvt.PtrToSlice(backend.PrivateIpAddresses),
			PublicIPAddress:  cvt.PtrToSlice(backend.PublicIpAddresses),
			InstName:         cvt.PtrToVal(backend.InstanceName),
		})
	}

	for _, updateBatch := range slice.Split(updates, constant.BatchOperationMaxLimit) {
		updateReq := &dataproto.TargetBatchUpdateReq{Targets: updateBatch}
		if err = cli.dbCli.Global.LoadBalancer.BatchUpdateTarget(kt, updateReq); err != nil {
			logs.Errorf("fail to update targets while syncing, err: %v, rid:%s", err, kt.Rid)
		}
	}

	return err
}

func (cli *client) createRs(kt *kit.Kit, accountID, tgId string, addSlice []typeslb.Backend) ([]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}

	targets := make([]*dataproto.TargetBaseReq, 0, constant.BatchOperationMaxLimit)
	createIds := make([]string, 0, len(targets))

	for _, targetBatch := range slice.Split(addSlice, constant.BatchOperationMaxLimit) {
		targets = targets[:0]
		for _, backend := range targetBatch {
			rs := convTarget(accountID)(backend.Backend)
			rs.TargetGroupID = tgId
			targets = append(targets, rs)
		}
		created, err := cli.dbCli.Global.LoadBalancer.BatchCreateTCloudTarget(kt,
			&dataproto.TargetBatchCreateReq{Targets: targets})
		if err != nil {
			logs.Errorf("fail to create target for target group syncing, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		createIds = append(createIds, created.IDs...)
	}

	return createIds, nil
}

// 获取云上监听器列表
func (cli *client) listTargetsFromCloud(kt *kit.Kit, param *SyncBaseParams,
	opt *SyncListenerOption) ([]typeslb.TCloudListenerTarget, error) {

	listOpt := &typeslb.TCloudListTargetsOption{
		Region:         param.Region,
		LoadBalancerId: opt.CloudLBID,
		ListenerIds:    param.CloudIDs,
	}
	return cli.cloudCli.ListTargets(kt, listOpt)
}

// 获取云上监听器列表
func (cli *client) listTargetsFromDB(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOption) (
	relMap map[string]*corelb.BaseTargetListenerRuleRel, tgRsMap map[string][]corelb.BaseTarget, err error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("lb_id", opt.LBID),
		Page:   core.NewDefaultBasePage(),
	}

	if len(param.CloudIDs) > 0 {
		listReq.Filter.Rules = append(listReq.Filter.Rules, tools.RuleIn("cloud_lbl_id", param.CloudIDs))
	}
	// 获取关系
	relList := make([]corelb.BaseTargetListenerRuleRel, 0)
	for {
		relRespTemp, err := cli.dbCli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
		if err != nil {
			logs.Errorf("fail to list target group listener rel, err: %v, req: %v, cloud_lbl_id: %s, rid: %s", err,
				listReq, param.CloudIDs, kt.Rid)
			return nil, nil, err
		}
		relList = append(relList, relRespTemp.Details...)

		if uint(len(relRespTemp.Details)) < core.DefaultMaxPageLimit {
			break
		}

		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	relMap = make(map[string]*corelb.BaseTargetListenerRuleRel)
	tgRsMap = make(map[string][]corelb.BaseTarget)
	if len(relList) == 0 {
		return relMap, tgRsMap, nil
	}

	tgIDMap := make(map[string]struct{}, len(relList))

	for i, rel := range relList {
		tgIDMap[rel.TargetGroupID] = struct{}{}
		relMap[rel.CloudListenerRuleID] = cvt.ValToPtr(relList[i])
	}

	// 目标组ID 去重
	tgIDs := cvt.MapKeyToStringSlice(tgIDMap)
	// 查询对应的rs列表
	for _, tgID := range tgIDs {
		// 按目标组查询
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("target_group_id", tgID)),
			Page:   core.NewDefaultBasePage(),
		}
		for {
			// 查询对应的rs列表
			rsList, err := cli.dbCli.Global.LoadBalancer.ListTarget(kt, req)
			if err != nil {
				logs.Errorf("fail to list targets of target group(id=%s), err: %v, rid: %s", tgID, err, kt.Rid)
				return nil, nil, err
			}
			tgRsMap[tgID] = append(tgRsMap[tgID], rsList.Details...)

			if uint(len(rsList.Details)) < core.DefaultMaxPageLimit {
				break
			}

			req.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return relMap, tgRsMap, nil
}

// 判断rs信息是否变化
func isRsChange(cloud typeslb.Backend, db corelb.BaseTarget) bool {
	if cloud.GetIP() != db.IP {
		return true
	}
	if cvt.PtrToVal(cloud.Port) != db.Port {
		return true
	}

	if cvt.PtrToVal(cloud.Weight) != cvt.PtrToVal(db.Weight) {
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

func genTargetGroupNameL4(lbl *corelb.Listener[corelb.TCloudListenerExtension]) string {
	return "auto-" + lbl.CloudID
}

func genTargetGroupNameL7(rule corelb.TCloudLbUrlRule) string {
	return "auto-" + rule.CloudID
}

// SyncTargetGroupOption ...
type SyncTargetGroupOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	BizID     int64  `json:"biz_id" validate:"required"`
	// 对应的负载均衡
	LBID      string `json:"lbid" validate:"required"`
	CloudLBID string `json:"cloud_lbid" validate:"required"`
}

// diff 该diff 和common.Diff的区别在于该接口的delete返回本地id
func diff[CloudType common.CloudResType, DBType common.DBResType](dataFromCloud []CloudType, dataFromDB []DBType,
	isChange func(CloudType, DBType) bool) (newAddData []CloudType, updateMap map[string]CloudType,
	delLocalIDs []string) {

	dbMap := make(map[string]DBType, len(dataFromDB))
	for _, one := range dataFromDB {
		dbMap[one.GetCloudID()] = one
	}

	newAddData = make([]CloudType, 0)
	updateMap = make(map[string]CloudType, 0)
	for _, oneFromCloud := range dataFromCloud {
		oneFromDB, exist := dbMap[oneFromCloud.GetCloudID()]
		if !exist {
			newAddData = append(newAddData, oneFromCloud)
			continue
		}

		delete(dbMap, oneFromCloud.GetCloudID())
		if isChange(oneFromCloud, oneFromDB) {
			updateMap[oneFromDB.GetID()] = oneFromCloud
		}
	}

	// 返回本地id 而不是云上id
	delLocalIDs = make([]string, 0, len(dbMap))
	for _, item := range dbMap {
		delLocalIDs = append(delLocalIDs, item.GetID())
	}

	return newAddData, updateMap, delLocalIDs
}

func genExists[T comparable]() (exists func(T) bool) {
	existsMap := make(map[T]struct{})
	exists = func(k T) bool {
		if _, exist := existsMap[k]; exist {
			return exist
		}
		existsMap[k] = struct{}{}
		return false
	}
	return exists
}
