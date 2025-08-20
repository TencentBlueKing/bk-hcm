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
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchModifyBizTargetsWeight batch modify biz targets weight.
func (svc *lbSvc) BatchModifyBizTargetsWeight(cts *rest.Contexts) (any, error) {
	return svc.batchModifyTargetWeight(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchModifyTargetWeight(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(cslb.BatchModifyTargetWeightReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch modify target weight request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	targetIDs := slice.Unique(req.TargetIDs)
	targets, err := svc.listTargetsByIDs(cts.Kit, targetIDs)
	if err != nil {
		return nil, err
	}
	if len(targets) != len(targetIDs) {
		return nil, fmt.Errorf("list target failed, expected: %d, actual: %d", len(targetIDs), len(targets))
	}
	// validate targets
	for _, target := range targets {
		if target.AccountID != req.AccountID {
			return nil, fmt.Errorf("target account_id: %s not match req account_id: %s", target.AccountID, req.AccountID)
		}
	}

	if err = svc.authBatchModifyTargetWeight(cts, targets, authHandler); err != nil {
		return nil, err
	}

	taskManagementID, err := svc.buildBatchModifyTargetWeightTask(cts.Kit, bizID, req, targets)
	if err != nil {
		return nil, err
	}
	return struct {
		TaskManagementID string `json:"task_management_id"`
	}{
		TaskManagementID: taskManagementID,
	}, nil
}

// buildBatchModifyTargetWeightTask 构建批量修改rs权重的任务管理
func (svc *lbSvc) buildBatchModifyTargetWeightTask(kt *kit.Kit, bkBizID int64, req *cslb.BatchModifyTargetWeightReq,
	targets []loadbalancer.BaseTarget) (string, error) {

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		kt, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, accountInfo.Vendor, req.AccountID,
		enumor.TaskManagementSourceAPI, enumor.TaskTargetGroupModifyWeight)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	tgToTargetsMap := classifier.ClassifySlice(targets, func(v loadbalancer.BaseTarget) string {
		return v.TargetGroupID
	})

	tgIDs := slice.Unique(slice.Map(targets, func(target loadbalancer.BaseTarget) string {
		return target.TargetGroupID
	}))
	relsMap, err := svc.listTGListenerRuleRelMapByTGIDs(kt, tgIDs)
	if err != nil {
		return "", err
	}
	for tgID, targetList := range tgToTargetsMap {
		_, ok := relsMap[tgID]
		if !ok {
			err = svc.batchUpdateTargetWeightDb(kt, taskManagementID, bkBizID, req.NewWeight, targetList)
			if err != nil {
				logs.Errorf("batch update target weight db failed, err: %v, tgID: %s, rid: %s", err, tgID, kt.Rid)
				return "", err
			}
		}
	}

	lbToRelsMap := classifier.ClassifyMap(relsMap, func(rel loadbalancer.BaseTargetListenerRuleRel) string {
		return rel.LbID
	})

	flowIDs := make([]string, 0, len(lbToRelsMap))
	for lbID, targetGroups := range lbToRelsMap {
		// 一个clb一个flow
		tgMap := make(map[string][]loadbalancer.BaseTarget)
		for _, tg := range targetGroups {
			tgMap[tg.TargetGroupID] = append(tgMap[tg.TargetGroupID], tgToTargetsMap[tg.TargetGroupID]...)
		}
		flowID, err := svc.buildModifyTargetWeightFlow(kt, lbID, req.AccountID, taskManagementID, accountInfo.Vendor,
			bkBizID, req.NewWeight, tgMap)
		if err != nil {
			logs.Errorf("build modify tcloud target tasks weight failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}
		flowIDs = append(flowIDs, flowID)
	}

	err = svc.updateTaskManagement(kt, taskManagementID, flowIDs...)
	if err != nil {
		return "", err
	}
	return taskManagementID, nil
}

func (svc *lbSvc) authBatchModifyTargetWeight(cts *rest.Contexts, targets []loadbalancer.BaseTarget,
	authHandler handler.ValidWithAuthHandler) error {

	tgIDs := slice.Map(targets, func(target loadbalancer.BaseTarget) string {
		return target.TargetGroupID
	})
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.TargetGroupCloudResType,
		IDs:          tgIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list target group basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return err
	}
	// authorized instances
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfos: basicInfoMap})
	if err != nil {
		logs.Errorf("batch modify target weight auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}
	return nil
}

func (svc *lbSvc) batchUpdateTargetWeightDb(kt *kit.Kit, taskManagementID string, bkBizID int64, newWeight *int64,
	targets []loadbalancer.BaseTarget) error {

	details := make([]*taskManagementDetail, 0)
	for _, one := range targets {
		details = append(details, &taskManagementDetail{
			param: one.ID,
		})
	}
	details, err := svc.createTaskDetails(kt, taskManagementID, bkBizID,
		enumor.TaskTargetGroupModifyWeight, details)
	if err != nil {
		return err
	}
	defer func() {
		detailIDs := slice.Map(details, func(d *taskManagementDetail) string { return d.taskDetailID })
		state := enumor.TaskDetailSuccess
		reason := ""
		if err != nil {
			state = enumor.TaskDetailFailed
			reason = err.Error()
		}
		if err := svc.updateTaskDetailState(kt, state, detailIDs, reason); err != nil {
			logs.Errorf("update task details state failed, err: %v, taskDetails: %+v, rid: %s", err, details, kt.Rid)
		}
	}()
	instExistsMap := make(map[string]struct{}, 0)
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: []*dataproto.TargetUpdate{}}
	for _, item := range targets {
		// 批量修改端口时，需要校验重复的实例ID的问题，否则云端接口也会报错
		if _, ok := instExistsMap[item.CloudInstID]; ok {
			return errf.Newf(errf.RecordDuplicated, "duplicate modify same inst(%s) to new_port", item.CloudInstID)
		}

		instExistsMap[item.CloudInstID] = struct{}{}
		updateReq.Targets = append(updateReq.Targets, &dataproto.TargetUpdate{
			ID:     item.ID,
			Weight: newWeight,
		})
	}

	if err = svc.client.DataService().Global.LoadBalancer.BatchUpdateTarget(kt, updateReq); err != nil {
		return err
	}
	return nil
}

func (svc *lbSvc) buildModifyTargetWeightFlow(kt *kit.Kit, lbID, accountID, taskManagementID string,
	vendor enumor.Vendor, bkBizID int64, newWeight *int64, tgToTargetsMap map[string][]loadbalancer.BaseTarget) (
	string, error) {

	// 预检测
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowModifyTargetWeight(kt, lbID, taskManagementID, accountID, vendor, bkBizID, newWeight,
		tgToTargetsMap)
	if err != nil {
		logs.Errorf("init flow batch modify target weigh failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyWeightTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (svc *lbSvc) initFlowModifyTargetWeight(kt *kit.Kit, lbID, taskManagementID, accountID string,
	vendor enumor.Vendor, bkBizID int64, newWeight *int64, tgToTargetsMap map[string][]loadbalancer.BaseTarget) (
	string, error) {

	var taskDetails []*taskManagementDetail
	var err error
	defer func() {
		if err == nil {
			return
		}
		// update task details state to failed
		taskDetailIDs := slice.Map(taskDetails, func(item *taskManagementDetail) string {
			return item.taskDetailID
		})
		if err := svc.updateTaskDetailState(kt, enumor.TaskDetailFailed, taskDetailIDs, err.Error()); err != nil {
			logs.Errorf("update task details state to failed failed, err: %v, taskDetails: %+v, rid: %s")
		}
	}()

	tasks, taskDetails, err := svc.buildModifyWeightFlowTasks(kt, lbID, accountID, taskManagementID, vendor,
		bkBizID, newWeight, tgToTargetsMap)
	if err != nil {
		return "", err
	}

	shareData := tableasync.NewShareData(map[string]string{
		"lb_id": lbID,
	})
	flowID, err := svc.buildFlow(kt, enumor.FlowTargetGroupModifyWeight, shareData, tasks)
	if err != nil {
		return "", err
	}
	for _, detail := range taskDetails {
		detail.flowID = flowID
	}

	// 下面的的代码如果执行出现error，不需要修改taskDetail的状态, 目前flow已经创建完毕，taskDetail由flowTask维护
	if err := svc.updateTaskDetails(kt, taskDetails); err != nil {
		logs.Errorf("update task details failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return "", err
	}
	if err := svc.buildSubFlow(kt, flowID, lbID, converter.MapKeyToSlice(tgToTargetsMap), enumor.TargetGroupCloudResType,
		enumor.ModifyWeightTaskType); err != nil {
		return "", err
	}
	return flowID, nil
}

func (svc *lbSvc) buildModifyWeightFlowTasks(kt *kit.Kit, lbID, accountID, taskManagementID string,
	vendor enumor.Vendor, bkBizID int64, newWeight *int64, tgToTargetsMap map[string][]loadbalancer.BaseTarget) (
	[]ts.CustomFlowTask, []*taskManagementDetail, error) {

	tasks := make([]ts.CustomFlowTask, 0)

	getActionID := counter.NewNumStringCounter(1, 10)
	var lastActionID action.ActIDType
	taskDetails := make([]*taskManagementDetail, 0)

	for tgID, targets := range tgToTargetsMap {
		elems := slice.Split(targets, constant.BatchModifyTargetWeightCloudMaxLimit)
		for _, parts := range elems {
			targetsIDs := slice.Map(parts, func(item loadbalancer.BaseTarget) string {
				return item.ID
			})
			rsWeightParams, err := svc.convTCloudOperateTargetReq(kt, targetsIDs, lbID, tgID, accountID, nil, newWeight)
			if err != nil {
				logs.Errorf("convert tcloud operate target req failed, err: %v, targetsIDs: %v, lbID: %s,"+
					" tgID: %s, accountID: %s, rid: %s", err, targetsIDs, lbID, tgID, accountID, kt.Rid)
				return nil, nil, err
			}
			details, err := svc.createTargetGroupModifyWeightTaskDetails(kt, taskManagementID, bkBizID, rsWeightParams)
			if err != nil {
				return nil, nil, err
			}

			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupModifyWeight,
				Params: &actionlb.OperateRsOption{
					Vendor: vendor,
					ManagementDetailIDs: slice.Map(details, func(item *taskManagementDetail) string {
						return item.taskDetailID
					}),
					TCloudBatchOperateTargetReq: *rsWeightParams,
				},
				Retry: &tableasync.Retry{
					Enable: true,
					Policy: &tableasync.RetryPolicy{
						Count:        500,
						SleepRangeMS: [2]uint{100, 200},
					},
				},
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

func (svc *lbSvc) createTargetGroupModifyWeightTaskDetails(kt *kit.Kit, taskManagementID string, bkBizID int64,
	addRsParams *hcproto.TCloudBatchOperateTargetReq) ([]*taskManagementDetail, error) {

	details := make([]*taskManagementDetail, 0)
	for _, one := range addRsParams.RsList {
		details = append(details, &taskManagementDetail{
			param: one,
		})
	}
	details, err := svc.createTaskDetails(kt, taskManagementID, bkBizID,
		enumor.TaskTargetGroupModifyWeight, details)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
		return nil, err
	}
	return details, nil
}
