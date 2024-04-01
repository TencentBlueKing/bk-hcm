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
	"hcm/pkg/api/data-service/cloud"
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
	// 1. 获取监听器详情
	cloudListenerTargets, err := cli.listTargetsFromCloud(kt, param, opt)
	if err != nil {
		return err
	}

	// 获取db中的目标组关系和rs列表
	relMap, tgRsMap, err := cli.listTargetsFromDB(kt, param, opt)
	if err != nil {
		return err
	}

	lbResp, err := cli.listLBFromDB(kt, &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{opt.CloudLBID},
	})
	if err != nil {
		logs.Errorf("fail to list lb from db for sync tg, err: %v, lb_id: %s, rid: %s", err, opt.CloudLBID, kt.Rid)
		return err
	}
	if len(lbResp) == 0 {
		logs.Errorf("can not find lb for sync tg, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return errf.Newf(errf.RecordNotFound, "lb not found: %s", opt.CloudLBID)
	}
	lb := lbResp[0]

	// TODO: 一个目标组只处理一次
	// 遍历云上的监听器、规则
	for _, listener := range cloudListenerTargets {
		if listener.GetProtocol().IsLayer7Protocol() {
			// 对比7层规则变化
			for _, rule := range listener.Rules {
				_ = rule

			}
			continue
		}
		err := cli.compareL4Listener(kt, lb, listener, opt, relMap, tgRsMap)
		if err != nil {
			logs.Errorf("fail to compare L4 listener rs change")
			return err
		}

	}
	return nil
}

func (cli *client) compareL4Listener(kt *kit.Kit, lb corelb.TCloudLoadBalancer, listener typeslb.TCloudListenerTarget,
	opt *SyncListenerOfSingleLBOption, relMap map[string]*corelb.BaseTargetListenerRuleRel,
	tgRsMap map[string][]corelb.BaseTarget) error {

	// 对比四层监听器
	rel, exists := relMap[cvt.PtrToVal(listener.ListenerId)]
	if !exists {
		// 云上监听器、规则中有RS，但是没有对应目标组，则在同步时自动创建目标组，并将RS加入目标组。
		if len(listener.Targets) > 0 {
			// TODO： 创建对应目标组
			return cli.createLocalTargetGroupL4(kt, lb, listener, opt)
		}
		// 只要本地没有目标组就跳过RS同步
		return nil
	}
	// 处理目标组中RS变化
	dbTargets := tgRsMap[rel.TargetGroupID]
	// 增加包裹类型
	cloudTargets := slice.Map(listener.Targets, func(rs *tclb.Backend) typeslb.Backend {
		return typeslb.Backend{Backend: rs}
	})
	// 比较对应的关系
	err := cli.handleRSChange(kt, rel.TargetGroupID, cloudTargets, dbTargets)
	if err != nil {
		logs.Errorf("fail to handle rs change for layer 4 listener, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (cli *client) handleRSChange(kt *kit.Kit, tgID string, cloudRsList []typeslb.Backend,
	dbRsList []corelb.BaseTarget) (err error) {

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.Backend, corelb.BaseTarget](cloudRsList, dbRsList,
		isRsChange)

	if err = cli.deleteRs(kt, delCloudIDs); err != nil {
		return err
	}

	if err = cli.updateRs(kt, updateMap); err != nil {
		return err
	}
	if _, err = cli.createRs(kt, addSlice); err != nil {
		return err
	}
	return nil
}

// 创建本地目标组以及关系
func (cli *client) createLocalTargetGroupL4(kt *kit.Kit, lb corelb.TCloudLoadBalancer,
	listener typeslb.TCloudListenerTarget, opt *SyncListenerOfSingleLBOption) error {

	lbl, rule, err := cli.listListenerWithRule(kt, cvt.PtrToVal(listener.ListenerId))
	if err != nil {
		return err
	}

	healthcheck, err := json.MarshalToString(rule)
	if err != nil {
		logs.Errorf("fail to marshal rule health check to string, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	tgCreate := cloud.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{
		TargetGroup: cloud.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			Name:            genAutoCreatedTargetGroupName(lbl),
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
		},
		ListenerRuleID:      lbl.ID,
		CloudListenerRuleID: lbl.CloudID,
		ListenerRuleType:    enumor.Layer4RuleType,
		LbID:                lb.ID,
		CloudLbID:           lb.CloudID,
		LblID:               lbl.ID,
		CloudLblID:          lbl.CloudLbID,
		BindingStatus:       enumor.SuccessBindingStatus,
	}
	for _, cloudTarget := range listener.Targets {
		tgCreate.TargetGroup.RsList = append(tgCreate.TargetGroup.RsList, &cloud.TargetBaseReq{
			InstType:    cvt.PtrToVal((*enumor.InstType)(cloudTarget.Type)),
			CloudInstID: cvt.PtrToVal(cloudTarget.InstanceId),
			Port:        cvt.PtrToVal(cloudTarget.Port),
			Weight:      cvt.PtrToVal(cloudTarget.Weight),
			// TODO: 缺少可用区
		})
	}
	tgCreateReq := &cloud.TCloudBatchCreateTgWithRelReq{
		TargetGroups: []cloud.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{tgCreate},
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
		return nil, nil, fmt.Errorf("listener can not be found by id(%s) while target group sync", listenerCloudID)
	}
	lbl := lblResp.Details[0]
	// 获取对应规则
	listReq.Filter = tools.ExpressionAnd(
		tools.RuleEqual("cloud_id", lbl.CloudID),
		tools.RuleEqual("lbl_id", lbl.ID))
	ruleResp, err := cli.dbCli.TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list rule of l4 listener, err: %v, lbl_id: %s, lbl_cloud_id: %s, rid: %s",
			lbl.ID, lbl.CloudID, kt.Rid)
		return nil, nil, err
	}
	if len(ruleResp.Details) == 0 {
		logs.Errorf("rule of listener can not be found by id(%s), lbl_id: %s, lbl_cloud_id: %s, rid: %s ",
			lbl.ID, lbl.CloudID, kt.Rid)
		return nil, nil, fmt.Errorf("rule of listener  can not be found by id(%s) while target group sync",
			listenerCloudID)
	}
	return cvt.ValToPtr(lbl), cvt.ValToPtr(ruleResp.Details[0]), nil
}

func genAutoCreatedTargetGroupName(lbl *corelb.Listener[corelb.TCloudListenerExtension]) string {
	return "auto-" + lbl.CloudID
}

// 按cloudInstID 删除目标组中的rs
func (cli *client) deleteRs(kt *kit.Kit, cloudIDs []string) error {
	if len(cloudIDs) == 0 {
		return nil
	}
	// TODO
	fmt.Printf("删除目标组中的rs: %+v \n", cloudIDs)

	return nil
}

func (cli *client) updateRs(kt *kit.Kit, updateMap map[string]typeslb.Backend) error {

	if len(updateMap) == 0 {
		return nil
	}
	// TODO 更新rs中的信息
	fmt.Printf("更新目标组中的rs: %+v \n", updateMap)

	return nil
}

func (cli *client) createRs(kt *kit.Kit, addSlice []typeslb.Backend) ([]string, error) {

	if len(addSlice) == 0 {
		return nil, nil
	}
	// TODO 添加rs
	fmt.Printf("添加目标组中的rs: %+v \n", addSlice)

	return nil, nil
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
		relMap[rel.CloudListenerRuleID] = &relResp.Details[i]
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

	if cvt.PtrToVal(cloud.Weight) != db.Weight {
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

// SyncTargetGroupOption ...
type SyncTargetGroupOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	BizID     int64  `json:"biz_id" validate:"required"`
	// 对应的负载均衡
	LBID      string `json:"lbid" validate:"required"`
	CloudLBID string `json:"cloud_lbid" validate:"required"`
}
