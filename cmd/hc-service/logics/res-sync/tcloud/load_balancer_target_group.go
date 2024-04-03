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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/classifier"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// ListenerTargets 监听器下的target，用来更新目标组.
// SyncBaseParams 中的CloudID作为监听器id筛选，不传的话就是同步当前LB下的全部监听器
func (cli *client) ListenerTargets(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOfSingleLBOption) error {

	cloudListenerTargets, relMap, tgRsMap, lb, err := cli.listTargetRelated(kt, param, opt)
	if err != nil {
		logs.Errorf("fail to list related res while syncing targets, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	processedTargetGroup := make(map[string]struct{})
	// TODO: 一个目标组只处理一次
	// 遍历云上的监听器、规则
	for _, listener := range cloudListenerTargets {
		if !listener.GetProtocol().IsLayer7Protocol() {
			// ---- for layer 4 对比监听器变化 ----
			rel, exists := relMap[cvt.PtrToVal(listener.ListenerId)]
			if !exists {
				// 云上监听器、规则中有RS，但是没有对应目标组，则在同步时自动创建目标组，并将RS加入目标组。
				if err := cli.createLocalTargetGroupL4(kt, opt, lb, listener); err != nil {
					logs.Errorf("fail to create local target group for layer 4 listener, rid: %s", kt.Rid)
					return err
				}
				// 只要本地没有目标组就跳过RS同步
				continue
			}
			tgId := rel.TargetGroupID
			if _, exists := processedTargetGroup[tgId]; exists {
				continue
			}
			processedTargetGroup[tgId] = struct{}{}
			// 存在则比较
			if err := cli.compareTargetsChange(kt, opt.AccountID, tgId, listener.Targets, tgRsMap[tgId]); err != nil {
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
			tgId := rel.TargetGroupID
			if _, exists := processedTargetGroup[tgId]; exists {
				continue
			}
			processedTargetGroup[tgId] = struct{}{}
			// 存在则比较
			if err := cli.compareTargetsChange(kt, opt.AccountID, tgId, rule.Targets, tgRsMap[tgId]); err != nil {
				logs.Errorf("fail to compare L7 rule rs change, err: %v, rid:%s", err, kt.Rid)
				return err
			}
		}
	}
	return nil
}

// 获取同步rs所需关联资源
func (cli *client) listTargetRelated(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOfSingleLBOption) (
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

	lbResp, err := cli.listLBFromDB(kt, &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
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
	dbRsList []corelb.BaseTarget) (
	err error) {

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

func (cli *client) createLocalTargetGroupL7(kt *kit.Kit, opt *SyncListenerOfSingleLBOption,
	lb *corelb.TCloudLoadBalancer, listener typeslb.TCloudListenerTarget, cloudRule *tclb.RuleTargets) error {

	// 跳过没有RS的规则
	if len(cloudRule.Targets) == 0 {
		return nil
	}

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
			AccountID:       opt.AccountID,
			BkBizID:         opt.BizID,
			Region:          opt.Region,
			Protocol:        listener.GetProtocol(),
			Port:            cvt.PtrToVal(listener.Port),
			VpcID:           lb.VpcID,
			CloudVpcID:      lb.CloudVpcID,
			TargetGroupType: enumor.LocalTargetGroupType,
			Weight:          0,
			HealthCheck:     types.JsonField(healthcheck),
			Memo:            cvt.ValToPtr("auto created for rule " + cvt.PtrToVal(cloudRule.LocationId)),
			RsList:          slice.Map(cloudRule.Targets, convTarget),
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

func convTarget(cloudTarget *tclb.Backend) *dataproto.TargetBaseReq {
	return &dataproto.TargetBaseReq{
		InstType:    cvt.PtrToVal((*enumor.InstType)(cloudTarget.Type)),
		CloudInstID: cvt.PtrToVal(cloudTarget.InstanceId),
		Port:        cvt.PtrToVal(cloudTarget.Port),
		Weight:      cloudTarget.Weight,
	}
}

// 创建本地目标组以及关系，会跳过没有rs的监听器
func (cli *client) createLocalTargetGroupL4(kt *kit.Kit, opt *SyncListenerOfSingleLBOption,
	lb *corelb.TCloudLoadBalancer, listener typeslb.TCloudListenerTarget) error {

	// 跳过没有rs的监听器
	if len(listener.Targets) == 0 {
		return nil
	}
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
			Region:          opt.Region,
			Protocol:        lbl.Protocol,
			Port:            lbl.Port,
			VpcID:           lb.VpcID,
			CloudVpcID:      lb.CloudVpcID,
			TargetGroupType: enumor.LocalTargetGroupType,
			Weight:          0,
			HealthCheck:     types.JsonField(healthcheck),
			Memo:            cvt.ValToPtr("auto created for listener " + cvt.PtrToVal(listener.ListenerId)),
			RsList:          slice.Map(listener.Targets, convTarget),
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

	delReq := &dataproto.LoadBalancerBatchDeleteReq{Filter: tools.ContainersExpression("id", localIds)}
	err := cli.dbCli.Global.LoadBalancer.BatchDeleteTarget(kt, delReq)
	if err != nil {
		logs.Errorf("fail to delete rs (ids=%v), err: %v, rid: %s", localIds, err, kt.Rid)
		return err
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
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: updates}
	if err = cli.dbCli.Global.LoadBalancer.BatchUpdateTarget(kt, updateReq); err != nil {
		logs.Errorf("fail to update targets while syncing, err: %v, rid:%s", err, kt.Rid)
	}

	return err
}

func (cli *client) createRs(kt *kit.Kit, accountID, tgId string, addSlice []typeslb.Backend) ([]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}

	var targets []*dataproto.TargetBaseReq
	for _, backend := range addSlice {
		targets = append(targets, &dataproto.TargetBaseReq{
			InstType:      cvt.PtrToVal((*enumor.InstType)(backend.Type)),
			CloudInstID:   cvt.PtrToVal(backend.InstanceId),
			Port:          cvt.PtrToVal(backend.Port),
			Weight:        backend.Weight,
			AccountID:     accountID,
			TargetGroupID: tgId,
		})
	}

	created, err := cli.dbCli.Global.LoadBalancer.BatchCreateTCloudTarget(kt,
		&dataproto.TargetBatchCreateReq{Targets: targets})
	if err != nil {
		logs.Errorf("fail to create target for target group syncing, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return created.IDs, nil
}

// 获取云上监听器列表
func (cli *client) listTargetsFromCloud(kt *kit.Kit, param *SyncBaseParams,
	opt *SyncListenerOfSingleLBOption) ([]typeslb.TCloudListenerTarget, error) {

	listOpt := &typeslb.TCloudListTargetsOption{
		Region:         opt.Region,
		LoadBalancerId: opt.CloudLBID,
		ListenerIds:    param.CloudIDs,
	}
	return cli.cloudCli.ListTargets(kt, listOpt)
}

// 获取云上监听器列表
func (cli *client) listTargetsFromDB(kt *kit.Kit, param *SyncBaseParams, opt *SyncListenerOfSingleLBOption) (
	relMap map[string]*corelb.BaseTargetListenerRuleRel, tgRsMap map[string][]corelb.BaseTarget, err error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("lb_id", opt.LBID),
		Page:   core.NewDefaultBasePage(),
	}

	if len(param.CloudIDs) > 0 {
		listReq.Filter.Rules = append(listReq.Filter.Rules, tools.RuleIn("cloud_lbl_id", param.CloudIDs))
	}
	// 获取关系
	relResp, err := cli.dbCli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
	if err != nil {
		logs.Errorf("fail to ListTargetGroupListenerRel, err: %v, rid: %s ", err, kt.Rid)
		return nil, nil, err
	}
	relMap = make(map[string]*corelb.BaseTargetListenerRuleRel)
	tgRsMap = make(map[string][]corelb.BaseTarget)
	if len(relResp.Details) == 0 {
		return relMap, tgRsMap, nil
	}

	tgIDMap := make(map[string]struct{}, len(relResp.Details))

	for i, rel := range relResp.Details {
		tgIDMap[rel.TargetGroupID] = struct{}{}
		relMap[rel.CloudListenerRuleID] = cvt.ValToPtr(relResp.Details[i])
	}
	relResp.Details = nil
	// 目标组ID 去重
	tgIDs := cvt.MapKeyToStringSlice(tgIDMap)

	// 查询对应的rs列表
	rsList, err := cli.dbCli.Global.LoadBalancer.ListTarget(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("target_group_id", tgIDs)),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list targets of target group(ids=%v), err: %v, rid: %s", tgIDs, err, kt.Rid)
		return nil, nil, err
	}
	// 按目标组分
	tgRsMap = classifier.ClassifySlice(rsList.Details, func(rs corelb.BaseTarget) string {
		return rs.TargetGroupID
	})

	return relMap, tgRsMap, nil
}

// 判断rs信息是否变化
func isRsChange(cloud typeslb.Backend, db corelb.BaseTarget) bool {
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
