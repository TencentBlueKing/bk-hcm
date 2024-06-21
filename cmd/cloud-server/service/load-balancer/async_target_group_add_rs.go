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

package loadbalancer

import (
	"encoding/json"
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/producer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchAddBizTargets create add biz targets.
func (svc *lbSvc) BatchAddBizTargets(cts *rest.Contexts) (any, error) {
	return svc.batchAddBizTarget(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchAddBizTarget(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch add target request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch add target auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.buildAddTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildAddTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string) (interface{}, error) {
	req := new(cslb.TCloudTargetBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查传入的Target是否已存在
	if err := svc.checkTargetExists(kt, req, accountID); err != nil {
		return nil, err
	}

	targetIDs := make([]string, 0)
	lbIDs := make([]string, 0)
	targetGroupRsListMap := make(map[string][]*dataproto.TargetBaseReq, 0)

	groupIds := slice.Map(req.TargetGroups, func(tg *cslb.TCloudBatchAddTargetReq) string { return tg.TargetGroupID })

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", groupIds),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud listener url rule failed, tgIds: %+v, err: %v, rid: %s", groupIds, err, kt.Rid)
		return nil, err
	}
	// 按目标组id分类
	targetGroupRuleRelMap := classifier.ClassifySlice(ruleRelList.Details,
		func(r corelb.BaseTargetListenerRuleRel) string { return r.TargetGroupID })

	for _, group := range req.TargetGroups {
		// 该目标组尚未绑定监听器及规则，不需要云端操作
		tgId := group.TargetGroupID
		relList := targetGroupRuleRelMap[tgId]
		if len(relList) == 0 {
			rsIDs, err := svc.batchCreateTargetDb(kt, group.Targets, tgId, accountID)
			if err != nil {
				logs.Errorf("fail to insert target, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			targetIDs = append(targetIDs, rsIDs.IDs...)
		} else {
			for _, rel := range relList {
				lbIDs = append(lbIDs, rel.LbID)
			}
			targetGroupRsListMap[tgId] = append(targetGroupRsListMap[tgId], group.Targets...)
		}

	}
	// 都是未绑定监听器的目标组，不需要云端操作
	if len(targetGroupRsListMap) == 0 {
		return &corelb.TargetOperateResult{TargetIDs: targetIDs}, nil
	}

	// 目标组需要属于同一个负载均衡
	if len(slice.Unique(lbIDs)) > 1 {
		return nil, errf.New(errf.InvalidParameter, "target group need belong to the same load balancer")
	}

	return svc.buildAddTCloudTargetTasks(kt, accountID, lbIDs[0], targetGroupRsListMap)
}

func (svc *lbSvc) checkTargetExists(kt *kit.Kit, req *cslb.TCloudTargetBatchCreateReq, accountID string) error {
	for _, tgItem := range req.TargetGroups {
		cloudInstIDs := make([]string, 0)
		ports := make([]int64, 0)
		for _, item := range tgItem.Targets {
			cloudInstIDs = append(cloudInstIDs, item.CloudInstID)
			ports = append(ports, item.Port)
		}
		tgReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("target_group_id", tgItem.TargetGroupID),
				tools.RuleIn("cloud_inst_id", cloudInstIDs),
				tools.RuleIn("port", ports),
			),
			Page: core.NewDefaultBasePage(),
		}
		rsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
		if err != nil {
			return err
		}
		if len(rsList.Details) > 0 {
			tmpCloudInstIds := slice.Unique(slice.Map(rsList.Details, func(target corelb.BaseTarget) string {
				return target.CloudInstID
			}))
			return errf.Newf(errf.RecordDuplicated, "targetGroupID: %s, cloudInstIDs: %v has exist",
				tgItem.TargetGroupID, tmpCloudInstIds)
		}
	}
	return nil
}

func (svc *lbSvc) batchCreateTargetDb(kt *kit.Kit, targets []*dataproto.TargetBaseReq, tgID, accountID string) (
	*core.BatchCreateResult, error) {

	addRsParams, err := svc.convTCloudAddTargetReq(kt, targets, "", tgID, accountID)
	if err != nil {
		return nil, err
	}

	rsReq := &dataproto.TargetBatchCreateReq{}
	for _, item := range addRsParams.RsList {
		rsReq.Targets = append(rsReq.Targets, &dataproto.TargetBaseReq{
			AccountID:     accountID,
			TargetGroupID: item.TargetGroupID,
			InstType:      item.InstType,
			CloudInstID:   item.CloudInstID,
			Port:          item.Port,
			Weight:        item.Weight,
		})
	}
	return svc.client.DataService().Global.LoadBalancer.BatchCreateTCloudTarget(kt, rsReq)
}

func (svc *lbSvc) buildAddTCloudTargetTasks(kt *kit.Kit, accountID, lbID string,
	tgMap map[string][]*dataproto.TargetBaseReq) (*core.FlowStateResult, error) {

	// 预检测
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowAddTargetByLbID(kt, accountID, lbID, tgMap)
	if err != nil {
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

func (svc *lbSvc) initFlowAddTargetByLbID(kt *kit.Kit, accountID, lbID string,
	tgMap map[string][]*dataproto.TargetBaseReq) (string, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	getActionID := counter.NewNumStringCounter(1, 10)
	var tgIDs []string
	var lastActionID action.ActIDType
	for tgID, rsList := range tgMap {
		tgIDs = append(tgIDs, tgID)
		elems := slice.Split(rsList, constant.BatchAddRSCloudMaxLimit)
		for _, parts := range elems {
			addRsParams, err := svc.convTCloudAddTargetReq(kt, parts, lbID, tgID, accountID)
			if err != nil {
				logs.Errorf("add target build tcloud request failed, err: %v, tgID: %s, parts: %+v, rid: %s",
					err, tgID, parts, kt.Rid)
				return "", err
			}
			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupAddRS,
				Params: &actionlb.OperateRsOption{
					Vendor:                      enumor.TCloud,
					TCloudBatchOperateTargetReq: *addRsParams,
				},
				Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
			}
			if len(lastActionID) > 0 {
				tmpTask.DependOn = []action.ActIDType{lastActionID}
			}
			tasks = append(tasks, tmpTask)
			lastActionID = actionID
		}
	}
	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowTargetGroupAddRS,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch add rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flowID := result.ID
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  tgIDs,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.AddRSTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

// convTCloudAddTargetReq conv tcloud add target req.
func (svc *lbSvc) convTCloudAddTargetReq(kt *kit.Kit, targets []*dataproto.TargetBaseReq, lbID, targetGroupID,
	accountID string) (*hcproto.TCloudBatchOperateTargetReq, error) {

	instMap, err := svc.getInstWithTargetMap(kt, targets)
	if err != nil {
		return nil, err
	}

	rsReq := &hcproto.TCloudBatchOperateTargetReq{TargetGroupID: targetGroupID, LbID: lbID}
	for _, item := range targets {
		item.TargetGroupID = targetGroupID
		item.AccountID = accountID
		item.InstName = instMap[item.CloudInstID].Name
		item.PrivateIPAddress = instMap[item.CloudInstID].PrivateIPv4Addresses
		item.PublicIPAddress = instMap[item.CloudInstID].PublicIPv4Addresses
		item.CloudVpcIDs = instMap[item.CloudInstID].CloudVpcIDs
		item.Zone = instMap[item.CloudInstID].Zone
		rsReq.RsList = append(rsReq.RsList, item)
	}
	return rsReq, nil
}

func (svc *lbSvc) getInstWithTargetMap(kt *kit.Kit, targets []*dataproto.TargetBaseReq) (
	map[string]corecvm.BaseCvm, error) {

	cloudCvmIDs := make([]string, 0)
	for _, item := range targets {
		if item.InstType == enumor.CvmInstType {
			cloudCvmIDs = append(cloudCvmIDs, item.CloudInstID)
		}
	}

	// 查询Cvm信息
	cvmMap := make(map[string]corecvm.BaseCvm)
	if len(cloudCvmIDs) > 0 {
		cvmReq := &core.ListReq{
			Filter: tools.ContainersExpression("cloud_id", cloudCvmIDs),
			Page:   core.NewDefaultBasePage(),
		}
		cvmList, err := svc.client.DataService().Global.Cvm.ListCvm(kt, cvmReq)
		if err != nil {
			logs.Errorf("failed to list cvm by cloudIDs, cloudIDs: %v, err: %v, rid: %s", cloudCvmIDs, err, kt.Rid)
			return nil, err
		}

		for _, item := range cvmList.Details {
			cvmMap[item.CloudID] = item
		}
	}

	return cvmMap, nil
}

func (svc *lbSvc) lockResFlowStatus(kt *kit.Kit, resID string, resType enumor.CloudResourceType, flowID string,
	taskType enumor.TaskType) error {

	// 锁定资源跟Flow的状态
	opt := &dataproto.ResFlowLockReq{
		ResID:    resID,
		ResType:  resType,
		FlowID:   flowID,
		Status:   enumor.ExecutingResFlowStatus,
		TaskType: taskType,
	}
	err := svc.client.DataService().Global.LoadBalancer.ResFlowLock(kt, opt)
	if err != nil {
		logs.Errorf("call dataservice to lock res and flow failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return err
	}

	// 更新Flow状态为pending
	flowStateReq := &producer.UpdateCustomFlowStateOption{
		FlowInfos: []backend.UpdateFlowInfo{{
			ID:     flowID,
			Source: enumor.FlowInit,
			Target: enumor.FlowPending,
		}},
	}
	err = svc.client.TaskServer().UpdateCustomFlowState(kt, flowStateReq)
	if err != nil {
		logs.Errorf("call taskserver to update flow state failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return err
	}

	return nil
}

func (svc *lbSvc) checkResFlowRel(kt *kit.Kit, resID string, resType enumor.CloudResourceType) (
	*corelb.BaseResFlowLock, error) {

	// 预检测-当前资源是否有锁定中的数据
	lockReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType),
		),
		Page: core.NewDefaultBasePage(),
	}
	lockRet, err := svc.client.DataService().Global.LoadBalancer.ListResFlowLock(kt, lockReq)
	if err != nil {
		logs.Errorf("list res flow lock failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType,
			kt.Rid)
		return nil, err
	}
	if len(lockRet.Details) > 0 {
		return &lockRet.Details[0], errf.Newf(errf.LoadBalancerTaskExecuting, "resID: %s is processing", resID)
	}

	// 预检测-当前资源是否有未终态的状态
	flowRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType),
			tools.RuleEqual("status", enumor.ExecutingResFlowStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	flowRelRet, err := svc.client.DataService().Global.LoadBalancer.ListResFlowRel(kt, flowRelReq)
	if err != nil {
		logs.Errorf("list res flow rel failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType, kt.Rid)
		return nil, err
	}
	if len(flowRelRet.Details) > 0 {
		return &corelb.BaseResFlowLock{
				ResID:   resID,
				ResType: resType,
				Owner:   flowRelRet.Details[0].FlowID,
			},
			errf.Newf(errf.LoadBalancerTaskExecuting, "%s of resID: %s is processing", resType, resID)
	}

	return nil, nil
}
