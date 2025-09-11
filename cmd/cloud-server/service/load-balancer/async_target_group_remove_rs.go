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
	"hcm/pkg/api/cloud-server/task"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
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
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchRemoveBizTargets batch remove biz targets.
func (svc *lbSvc) BatchRemoveBizTargets(cts *rest.Contexts) (any, error) {
	return svc.batchRemoveBizTarget(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchRemoveBizTarget(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	any, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	req := new(cslb.BatchRemoveTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch remove target request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(); err != nil {
		return nil, err
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch remove target auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	targetIDs := slice.Unique(req.TargetIDs)
	targets, err := svc.listTargetsByIDs(cts.Kit, targetIDs)
	if err != nil {
		logs.Errorf("list targets by ids failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(targets) != len(targetIDs) {
		return nil, fmt.Errorf("list target failed, expected: %d, actual: %d", len(targetIDs), len(targets))
	}
	// validate targets
	for _, target := range targets {
		if target.AccountID != req.AccountID {
			return nil, fmt.Errorf("target account_id: %s not match req account_id: %s",
				target.AccountID, req.AccountID)
		}
	}

	taskManagementID, err := svc.buildRemoveTargetManagement(cts.Kit, accountInfo.Vendor, accountInfo.AccountID, bizID,
		targets)
	if err != nil {
		return nil, err
	}
	return task.CreateTaskManagementResp{TaskManagementID: taskManagementID}, nil
}

// buildRemoveTargetManagement builds the task management for removing targets.
func (svc *lbSvc) buildRemoveTargetManagement(kt *kit.Kit, vendor enumor.Vendor, accountID string, bkBizID int64,
	targets []corelb.BaseTarget) (string, error) {

	tgToTargetsMap := classifier.ClassifySlice(targets, corelb.BaseTarget.GetTargetGroupID)
	tgIDs := cvt.MapKeyToSlice(tgToTargetsMap)
	relsMap, err := svc.listTGListenerRuleRelMapByTGIDs(kt, tgIDs)
	if err != nil {
		return "", err
	}

	lbToRelsMap := classifier.ClassifyMap(relsMap, corelb.BaseTargetListenerRuleRel.GetLbID)
	for lbID := range lbToRelsMap {
		// 预检测
		_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
		if err != nil {
			logs.Errorf("check resource flow relation failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
			return "", err
		}
	}
	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, vendor, accountID,
		enumor.TaskManagementSourceAPI, enumor.TaskTargetGroupRemoveRS)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	tgRelatedInfo, err := svc.listTGRelatedInfoByRels(kt, vendor, cvt.MapValueToSlice(relsMap))
	if err != nil {
		logs.Errorf("list target group related info by rels failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	for tgID, targetList := range tgToTargetsMap {
		_, ok := relsMap[tgID]
		if !ok {
			err := svc.batchDeleteTargetDb(kt, accountID, tgID, taskManagementID, bkBizID, targetList,
				tgRelatedInfo[tgID])
			if err != nil {
				return "", err
			}
		}
	}
	flowIDs := make([]string, 0, len(lbToRelsMap))
	for lbID, tgRuleRels := range lbToRelsMap {
		// 一个clb一个flow
		tgMap := make(map[string][]corelb.BaseTarget)
		for _, rel := range tgRuleRels {
			tgMap[rel.TargetGroupID] = append(tgMap[rel.TargetGroupID], tgToTargetsMap[rel.TargetGroupID]...)
		}
		flowID, err := svc.buildRemoveTCloudTargetTasks(kt, accountID, lbID, taskManagementID, vendor, bkBizID,
			tgMap, tgRelatedInfo)
		if err != nil {
			logs.Errorf("build remove tcloud target tasks failed, err: %v, rid: %s", err, kt.Rid)
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

func (svc *lbSvc) batchDeleteTargetDb(kt *kit.Kit, accountID, tgID, taskManagementID string, bkBizID int64,
	targets []corelb.BaseTarget, info TGRelatedInfo) error {

	rsIDs := make([]string, 0, len(targets))
	details := make([]*taskManagementDetail, 0)
	for _, one := range targets {
		param := struct {
			TGRelatedInfo     `json:",inline"`
			corelb.BaseTarget `json:",inline"`
		}{
			TGRelatedInfo: info,
			BaseTarget:    one,
		}
		details = append(details, &taskManagementDetail{
			param: param,
		})
		rsIDs = append(rsIDs, one.ID)
	}
	details, err := svc.createTaskDetails(kt, taskManagementID, bkBizID, enumor.TaskTargetGroupRemoveRS, details)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
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

	for _, batch := range slice.Split(rsIDs, int(core.DefaultMaxPageLimit)) {
		delReq := &dataproto.LoadBalancerBatchDeleteReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", batch),
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("target_group_id", tgID),
			),
		}
		if err = svc.client.DataService().Global.LoadBalancer.BatchDeleteTarget(kt, delReq); err != nil {
			return err
		}
	}
	return nil
}

func (svc *lbSvc) buildRemoveTCloudTargetTasks(kt *kit.Kit, accountID, lbID, taskManagementID string,
	vendor enumor.Vendor, bkBizID int64, tgMap map[string][]corelb.BaseTarget, tgRelatedInfo map[string]TGRelatedInfo) (
	string, error) {

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowRemoveTargetByLbID(kt, accountID, lbID, taskManagementID, vendor, bkBizID, tgMap,
		tgRelatedInfo)
	if err != nil {
		logs.Errorf("init flow batch remove target failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.RemoveRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (svc *lbSvc) initFlowRemoveTargetByLbID(kt *kit.Kit, accountID, lbID, taskManagementID string,
	vendor enumor.Vendor, bkBizID int64, tgMap map[string][]corelb.BaseTarget, tgRelatedInfo map[string]TGRelatedInfo) (
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
			logs.Errorf("update task details state to failed, err: %v, taskDetails: %+v, rid: %s", err,
				taskDetails, kt.Rid)
		}
	}()

	tasks, taskDetails, err := svc.buildRemoveRSTasks(kt, accountID, lbID, taskManagementID, vendor, bkBizID,
		tgMap, tgRelatedInfo)
	if err != nil {
		logs.Errorf("build remove target tasks failed, err: %v, accountID: %s, lbID: %s, bkBizID: %d, rid: %s", err,
			accountID, lbID, bkBizID, kt.Rid)
		return "", err
	}

	shareData := tableasync.NewShareData(map[string]string{
		"lb_id": lbID,
	})
	flowID, err := svc.buildFlow(kt, enumor.FlowTargetGroupRemoveRS, shareData, tasks)
	if err != nil {
		logs.Errorf("build flow failed, err: %v, accountID: %s, lbID: %s, bkBizID: %d, rid: %s", err, accountID,
			lbID, bkBizID, kt.Rid)
		return "", err
	}
	for _, detail := range taskDetails {
		detail.flowID = flowID
	}

	if err := svc.updateTaskDetails(kt, taskDetails); err != nil {
		logs.Errorf("update task details failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return "", err
	}

	tgIDs := cvt.MapKeyToSlice(tgMap)
	if err := svc.buildSubFlow(kt, flowID, lbID, tgIDs, enumor.TargetGroupCloudResType,
		enumor.RemoveRSTaskType); err != nil {
		logs.Errorf("build sub flow failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return "", err
	}
	return flowID, nil
}

func (svc *lbSvc) buildRemoveRSTasks(kt *kit.Kit, accountID, lbID, taskManagementID string, vendor enumor.Vendor,
	bkBizID int64, tgMap map[string][]corelb.BaseTarget, tgRelatedInfo map[string]TGRelatedInfo) (
	[]ts.CustomFlowTask, []*taskManagementDetail, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	getActionID := counter.NewNumStringCounter(1, 10)
	var lastActionID action.ActIDType
	taskDetails := make([]*taskManagementDetail, 0)
	for tgID, rsList := range tgMap {
		for _, parts := range slice.Split(rsList, constant.BatchRemoveRSCloudMaxLimit) {
			removeRsParams, err := svc.convTCloudOperateTargetReq(parts, lbID, tgID, accountID, nil, nil)
			if err != nil {
				return nil, nil, err
			}
			details, err := svc.createTargetGroupRemoveRsTaskDetails(kt, taskManagementID, bkBizID, removeRsParams,
				tgRelatedInfo[tgID])
			if err != nil {
				return nil, nil, err
			}
			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupRemoveRS,
				Params: &actionlb.OperateRsOption{
					Vendor: vendor,
					ManagementDetailIDs: slice.Map(details, func(item *taskManagementDetail) string {
						return item.taskDetailID
					}),
					TCloudBatchOperateTargetReq: *removeRsParams,
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

type tgRemoveRSTaskDetailParam struct {
	TGRelatedInfo            `json:",inline"`
	*dataproto.TargetBaseReq `json:",inline"`
}

func (svc *lbSvc) createTargetGroupRemoveRsTaskDetails(kt *kit.Kit, taskManagementID string, bkBizID int64,
	addRsParams *hcproto.TCloudBatchOperateTargetReq, info TGRelatedInfo) ([]*taskManagementDetail, error) {

	details := make([]*taskManagementDetail, 0)
	for _, one := range addRsParams.RsList {
		details = append(details, &taskManagementDetail{
			param: tgRemoveRSTaskDetailParam{
				TGRelatedInfo: info,
				TargetBaseReq: one,
			},
		})
	}
	details, err := svc.createTaskDetails(kt, taskManagementID, bkBizID, enumor.TaskTargetGroupRemoveRS, details)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
		return nil, err
	}
	return details, nil
}

// convTCloudOperateTargetReq conv tcloud operate target req.
func (svc *lbSvc) convTCloudOperateTargetReq(targets []corelb.BaseTarget, lbID, targetGroupID, accountID string,
	newPort, newWeight *int64) (*hcproto.TCloudBatchOperateTargetReq, error) {

	instExistsMap := make(map[string]struct{})
	rsReq := &hcproto.TCloudBatchOperateTargetReq{TargetGroupID: targetGroupID, LbID: lbID}
	for _, item := range targets {
		// 批量修改端口时，需要校验重复的实例ID的问题，否则云端接口也会报错
		if cvt.PtrToVal(newPort) > 0 {
			if _, ok := instExistsMap[item.CloudInstID]; ok {
				return nil, errf.Newf(errf.RecordDuplicated, "duplicate modify same inst(%s) to new_port: %d",
					item.CloudInstID, cvt.PtrToVal(newPort))
			}
			instExistsMap[item.CloudInstID] = struct{}{}
		}

		rsReq.RsList = append(rsReq.RsList, &dataproto.TargetBaseReq{
			ID:                item.ID,
			IP:                item.IP,
			InstType:          item.InstType,
			CloudInstID:       item.CloudInstID,
			Port:              item.Port,
			Weight:            item.Weight,
			AccountID:         accountID,
			TargetGroupID:     targetGroupID,
			InstName:          item.InstName,
			PrivateIPAddress:  item.PrivateIPAddress,
			PublicIPAddress:   item.PublicIPAddress,
			CloudVpcIDs:       item.CloudVpcIDs,
			Zone:              item.Zone,
			NewPort:           newPort,
			NewWeight:         newWeight,
			TargetGroupRegion: item.TargetGroupRegion,
		})
	}
	return rsReq, nil
}
