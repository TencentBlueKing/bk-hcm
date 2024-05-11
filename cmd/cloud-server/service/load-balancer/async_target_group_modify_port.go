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
	dataproto "hcm/pkg/api/data-service/cloud"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchModifyBizTargetsPort batch modify biz targets port.
func (svc *lbSvc) BatchModifyBizTargetsPort(cts *rest.Contexts) (any, error) {
	return svc.batchModifyTargetPort(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchModifyTargetPort(cts *rest.Contexts,
	authHandler handler.ValidWithAuthHandler) (any, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch modify target port request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		logs.Errorf("get target group resource info failed, id: %s, err: %s, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// authorized instances
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: baseInfo})
	if err != nil {
		logs.Errorf("batch modify target port auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.buildModifyTCloudTargetPort(cts.Kit, req.Data, tgID, baseInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

func (svc *lbSvc) buildModifyTCloudTargetPort(kt *kit.Kit, body json.RawMessage,
	tgID, accountID string) (interface{}, error) {

	req := new(cslb.TCloudBatchModifyTargetPortReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

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
		if err = svc.batchUpdateTargetPortDb(kt, req); err != nil {
			return nil, err
		}
		return &core.FlowStateResult{State: enumor.FlowSuccess}, nil
	}

	return svc.buildModifyTCloudTargetTasksPort(kt, req, ruleRelList.Details[0].LbID, tgID, accountID)
}

func (svc *lbSvc) batchUpdateTargetPortDb(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetPortReq) error {
	tgReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", req.TargetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
	if err != nil {
		return err
	}
	if len(rsList.Details) == 0 {
		return errf.Newf(errf.RecordNotFound, "target_ids: %v is not found", req.TargetIDs)
	}

	instExistsMap := make(map[string]struct{}, 0)
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: []*dataproto.TargetUpdate{}}
	for _, item := range rsList.Details {
		// 批量修改端口时，需要校验重复的实例ID的问题，否则云端接口也会报错
		if _, ok := instExistsMap[item.CloudInstID]; ok {
			return errf.Newf(errf.RecordDuplicated, "duplicate modify same inst(%s) to new_port", item.CloudInstID)
		}

		instExistsMap[item.CloudInstID] = struct{}{}
		updateReq.Targets = append(updateReq.Targets, &dataproto.TargetUpdate{
			ID:   item.ID,
			Port: req.NewPort,
		})
	}

	return svc.client.DataService().Global.LoadBalancer.BatchUpdateTarget(kt, updateReq)
}

func (svc *lbSvc) buildModifyTCloudTargetTasksPort(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetPortReq, lbID, tgID,
	accountID string) (interface{}, error) {

	// 预检测
	_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowTargetPort(kt, req, lbID, tgID, accountID)
	if err != nil {
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyPortTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

func (svc *lbSvc) initFlowTargetPort(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetPortReq,
	lbID, tgID, accountID string) (string, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	elems := slice.Split(req.TargetIDs, constant.BatchModifyTargetPortCloudMaxLimit)
	getActionID := counter.NewNumStringCounter(1, 10)
	var lastActionID action.ActIDType
	for _, parts := range elems {
		rsPortParams, err := svc.convTCloudOperateTargetReq(kt, parts, lbID, tgID, accountID,
			cvt.ValToPtr(req.NewPort), nil)
		if err != nil {
			return "", err
		}
		actionID := action.ActIDType(getActionID())
		tmpTask := ts.CustomFlowTask{
			ActionID:   actionID,
			ActionName: enumor.ActionTargetGroupModifyPort,
			Params: &actionlb.OperateRsOption{
				Vendor:                      enumor.TCloud,
				TCloudBatchOperateTargetReq: *rsPortParams,
			},
			Retry: &tableasync.Retry{
				Enable: true,
				Policy: &tableasync.RetryPolicy{
					Count:        constant.FlowRetryMaxLimit,
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
	portReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowTargetGroupModifyPort,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, portReq)
	if err != nil {
		logs.Errorf("call taskserver to batch modify target port custom flow failed, err: %v, rid: %s", err, kt.Rid)
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
				SubResIDs:  []string{tgID},
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ModifyPortTaskType,
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
