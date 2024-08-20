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
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

func (svc *lbSvc) buildBatchModifyTCloudTargetWeight(kt *kit.Kit, tgTargetsMap map[string][]string,
	newWeight *int64, accountID string) ([]*core.FlowStateResult, error) {

	lbTgMap := make(map[string][]string)
	for tgID, targetIDs := range tgTargetsMap {
		// 根据目标组ID，获取目标组绑定的监听器、规则列表
		ruleRelReq := &core.ListReq{
			Filter: tools.EqualExpression("target_group_id", tgID),
			Page:   core.NewDefaultBasePage(),
		}
		ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
		if err != nil {
			logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
			return nil, err
		}

		// 该目标组尚未绑定监听器及规则，不需要云端操作
		if len(ruleRelList.Details) == 0 {
			err = svc.batchUpdateTargetWeightDb(kt, &cslb.TCloudBatchModifyTargetWeightReq{
				TargetIDs: targetIDs,
				NewWeight: newWeight,
			})
			if err != nil {
				logs.Errorf("batch update target weight in database failed, targetIDs: %v, newWeight: %d, err: %v, rid: %s",
					targetIDs, newWeight, err, kt.Rid)
				return nil, err
			}
			continue
		}

		// 需要云端操作的，以lbID分组记录信息
		lbID := ruleRelList.Details[0].LbID
		if _, exists := lbTgMap[lbID]; !exists {
			lbTgMap[lbID] = make([]string, 0)
		}
		lbTgMap[lbID] = append(lbTgMap[lbID], tgID)
	}

	// 一个lb对应一个流程
	flowStateResults := make([]*core.FlowStateResult, 0)
	for lbID, tgIDs := range lbTgMap {
		tgAndReqSlice := make([]cslb.TgIDAndTCloudBatchModifyTargetWeightReq, 0)
		for _, tgID := range tgIDs {
			tgAndReq := cslb.TgIDAndTCloudBatchModifyTargetWeightReq{
				TgID: tgID,
				Req: cslb.TCloudBatchModifyTargetWeightReq{
					TargetIDs: tgTargetsMap[tgID],
					NewWeight: newWeight,
				},
			}
			tgAndReqSlice = append(tgAndReqSlice, tgAndReq)
		}
		flowStateResult, err := svc.buildBatchModifyTCloudTargetTasksWeight(kt, accountID, lbID, tgAndReqSlice)
		if err != nil {
			logs.Errorf("build batch modify tcloud target weight tasks failed, err: %v, accountID: %s, lbID: %s, tgAndReqSlice: %+v, rid: %s",
				err, accountID, lbID, tgAndReqSlice, kt.Rid)
			return nil, err
		}
		flowStateResults = append(flowStateResults, flowStateResult)
	}

	return flowStateResults, nil
}

func (svc *lbSvc) buildBatchModifyTCloudTargetTasksWeight(kt *kit.Kit, accountID string,
	lbID string, tgAndReqSlice []cslb.TgIDAndTCloudBatchModifyTargetWeightReq) (*core.FlowStateResult, error) {

	// 预检测
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowBatchModifyTargetWeight(kt, accountID, lbID, tgAndReqSlice)
	if err != nil {
		logs.Errorf("init flow batch modify target weight failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyWeightTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

func (svc *lbSvc) initFlowBatchModifyTargetWeight(kt *kit.Kit, accountID string,
	lbID string, tgAndReqSlice []cslb.TgIDAndTCloudBatchModifyTargetWeightReq) (string, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	getActionID := counter.NewNumStringCounter(1, 10)
	var lastActionID action.ActIDType
	tgIDs := make([]string, 0, len(tgAndReqSlice))
	for _, tgAndReq := range tgAndReqSlice {
		tgIDs = append(tgIDs, tgAndReq.TgID)
		elems := slice.Split(tgAndReq.Req.TargetIDs, constant.BatchModifyTargetWeightCloudMaxLimit)
		for _, parts := range elems {
			rsWeightParams, err := svc.convTCloudOperateTargetReq(kt, parts, lbID, tgAndReq.TgID, accountID, nil, tgAndReq.Req.NewWeight)
			if err != nil {
				return "", err
			}
			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupModifyWeight,
				Params: &actionlb.OperateRsOption{
					Vendor:                      enumor.TCloud,
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
		}
	}
	rsWeightReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowTargetGroupModifyWeight,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, rsWeightReq)
	if err != nil {
		logs.Errorf("call taskserver to batch modify target weight custom flow failed, err: %v, rsWeightReq: %+v, rid: %s",
			err, converter.PtrToVal(rsWeightReq), kt.Rid)
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
				TaskType:   enumor.ModifyWeightTaskType,
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
