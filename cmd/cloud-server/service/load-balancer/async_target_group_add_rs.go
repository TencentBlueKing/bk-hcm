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
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
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
		return svc.buildAddTCloudTarget(cts.Kit, bizID, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildAddTCloudTarget(kt *kit.Kit, bkBizID int64, body json.RawMessage, accountID string) (
	interface{}, error) {

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

	return svc.buildAddTCloudTargetTasks(kt, bkBizID, accountID, lbIDs[0], targetGroupRsListMap)
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
			AccountID:         accountID,
			TargetGroupID:     item.TargetGroupID,
			InstType:          item.InstType,
			CloudInstID:       item.CloudInstID,
			IP:                item.IP,
			Port:              item.Port,
			Weight:            item.Weight,
			TargetGroupRegion: item.TargetGroupRegion,
		})
	}
	return svc.client.DataService().Global.LoadBalancer.BatchCreateTCloudTarget(kt, rsReq)
}

func (svc *lbSvc) buildAddTCloudTargetTasks(kt *kit.Kit, bkBizID int64, accountID, lbID string,
	tgMap map[string][]*dataproto.TargetBaseReq) (*core.FlowStateResult, error) {

	// 预检测
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowAddTargetByLbID(kt, accountID, lbID, bkBizID, tgMap)
	if err != nil {
		logs.Errorf("init flow batch add target failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

func (svc *lbSvc) initFlowAddTargetByLbID(kt *kit.Kit, accountID, lbID string, bkBizID int64,
	tgMap map[string][]*dataproto.TargetBaseReq) (string, error) {

	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, enumor.TCloud, accountID,
		enumor.TaskManagementSourceAPI, enumor.TaskTargetGroupAddRS)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	var taskDetails []*taskManagementDetail
	defer func() {
		if err == nil {
			return
		}
		// update task management state to failed
		if err := svc.updateTaskManagementState(kt, taskManagementID, enumor.TaskManagementFailed); err != nil {
			logs.Errorf("update task management state to failed failed, err: %v, taskManagementID: %s, rid: %s",
				err, taskManagementID, kt.Rid)
		}
		// update task details state to failed
		taskDetailIDs := slice.Map(taskDetails, func(item *taskManagementDetail) string {
			return item.taskDetailID
		})
		if err := svc.updateTaskDetailState(kt, enumor.TaskDetailFailed, taskDetailIDs, err.Error()); err != nil {
			logs.Errorf("update task details state to failed failed, err: %v, taskDetails: %+v, rid: %s")
		}
	}()

	tasks, taskDetails, err := svc.buildAddTargetTasks(kt, accountID, lbID, taskManagementID, bkBizID, tgMap)
	if err != nil {
		logs.Errorf("build add target tasks failed, err: %v, accountID: %s, lbID: %s, bkBizID: %d, rid: %s", err,
			accountID, lbID, bkBizID, kt.Rid)
		return "", err
	}

	shareData := tableasync.NewShareData(map[string]string{
		"lb_id": lbID,
	})
	flowID, err := svc.buildFlow(kt, enumor.FlowTargetGroupAddRS, shareData, tasks)
	if err != nil {
		return "", err
	}
	for _, detail := range taskDetails {
		detail.flowID = flowID
	}

	if err = svc.updateTaskDetails(kt, taskDetails); err != nil {
		logs.Errorf("update task details failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return "", err
	}
	if err = svc.updateTaskManagement(kt, taskManagementID, flowID); err != nil {
		logs.Errorf("update task management failed, err: %v, taskManagementID: %s, rid: %s",
			err, taskManagementID, kt.Rid)
		return "", err
	}
	tgIDs := make([]string, 0, len(tgMap))
	for tgID := range tgMap {
		tgIDs = append(tgIDs, tgID)
	}
	if err = svc.buildSubFlow(kt, flowID, lbID, tgIDs, enumor.TargetGroupCloudResType,
		enumor.AddRSTaskType); err != nil {
		return "", err
	}
	return flowID, nil
}

func (svc *lbSvc) buildAddTargetTasks(kt *kit.Kit, accountID, lbID, taskManagementID string, bkBizID int64,
	tgMap map[string][]*dataproto.TargetBaseReq) ([]ts.CustomFlowTask, []*taskManagementDetail, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	getActionID := counter.NewNumStringCounter(1, 10)
	var lastActionID action.ActIDType
	taskDetails := make([]*taskManagementDetail, 0)
	for tgID, rsList := range tgMap {
		for _, parts := range slice.Split(rsList, constant.BatchAddRSCloudMaxLimit) {
			addRsParams, err := svc.convTCloudAddTargetReq(kt, parts, lbID, tgID, accountID)
			if err != nil {
				logs.Errorf("add target build tcloud request failed, err: %v, tgID: %s, parts: %+v, rid: %s",
					err, tgID, parts, kt.Rid)
				return nil, nil, err
			}

			details, err := svc.createTargetGroupAddRsTaskDetails(kt, taskManagementID, bkBizID, addRsParams)
			if err != nil {
				return nil, nil, err
			}

			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupAddRS,
				Params: &actionlb.OperateRsOption{
					Vendor: enumor.TCloud,
					ManagementDetailIDs: slice.Map(details, func(item *taskManagementDetail) string {
						return item.taskDetailID
					}),
					TCloudBatchOperateTargetReq: *addRsParams,
				},
				Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
			}
			if len(lastActionID) > 0 {
				tmpTask.DependOn = []action.ActIDType{lastActionID}
			}
			tasks = append(tasks, tmpTask)
			lastActionID = actionID

			for _, detail := range details {
				detail.actionID = string(actionID)
			}
			taskDetails = append(taskDetails, details...)
		}
	}

	return tasks, taskDetails, nil
}

func (svc *lbSvc) createTargetGroupAddRsTaskDetails(kt *kit.Kit, taskManagementID string, bkBizID int64,
	addRsParams *hcproto.TCloudBatchOperateTargetReq) ([]*taskManagementDetail, error) {

	details := make([]*taskManagementDetail, 0)
	for _, one := range addRsParams.RsList {
		details = append(details, &taskManagementDetail{
			param: one,
		})
	}
	if err := svc.createTaskDetails(kt, taskManagementID, bkBizID,
		enumor.TaskTargetGroupAddRS, details); err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
		return nil, err
	}
	return details, nil
}

// convTCloudAddTargetReq conv tcloud add target req.
func (svc *lbSvc) convTCloudAddTargetReq(kt *kit.Kit, targets []*dataproto.TargetBaseReq, lbID, targetGroupID,
	accountID string) (*hcproto.TCloudBatchOperateTargetReq, error) {

	instMap, err := svc.getInstWithTargetMap(kt, targets)
	if err != nil {
		return nil, err
	}

	tgReq := &core.ListReq{
		Filter: tools.EqualExpression("id", targetGroupID),
		Fields: []string{"region"},
		Page:   core.NewDefaultBasePage(),
	}
	tgResult, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, tgReq)
	if err != nil {
		logs.Errorf("fail to get target group, err: %v, target group id: %s, rid: %s", err, targetGroupID, kt.Rid)
		return nil, err
	}
	if len(tgResult.Details) != 1 {
		logs.Errorf("can not find target group, target group id: %s, rid: %s", targetGroupID, kt.Rid)
		return nil, fmt.Errorf("can not find target group, target group id: %s", targetGroupID)
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
		item.TargetGroupRegion = tgResult.Details[0].Region
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
