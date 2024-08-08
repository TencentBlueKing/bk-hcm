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

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ModifyWeight 批量修改RS权重接口
func (svc *lbSvc) ModifyWeight(cts *rest.Contexts) (interface{}, error) {

	req := new(cloud.BatchOperationReq[*lblogic.ModifyWeightRecord])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch operation modify weight auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.modifyWeight(cts, req)
}

func (svc *lbSvc) modifyWeight(cts *rest.Contexts,
	req *cloud.BatchOperationReq[*lblogic.ModifyWeightRecord]) (interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	flows := make([]string, 0)
	flowAuditMap := make(map[string]uint64)
	for _, tmp := range req.Data {
		// 一个CLB一个异步任务
		lb, err := svc.getLoadBalancersByID(cts.Kit, bizID, tmp.ClbID)
		if err != nil {
			logs.Errorf("get load balancer failed, lbID: %s, err: %v, rid: %s", tmp.ClbID, err, cts.Kit.Rid)
			return nil, err
		}
		flowID, err := buildAsyncFlow(cts.Kit, svc, tmp.Listeners, lb, svc.initBatchModifyWeightTask)
		if err != nil {
			logs.Errorf("build async flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		auditRecord, err := svc.getAuditByLoadBalanceID(cts.Kit, lb.ID)
		if err != nil {
			logs.Errorf("get audit failed, lbID: %s, err: %v, rid: %s", lb.ID, err, cts.Kit.Rid)
			return nil, err
		}
		flows = append(flows, flowID)
		flowAuditMap[flowID] = auditRecord.ID
	}

	detail, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	batchOperationID, err := svc.saveBatchOperationRecord(cts, string(detail), flowAuditMap, req.AccountID)
	if err != nil {
		logs.Errorf("save batch operation record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return batchOperationID, nil
}

func (svc *lbSvc) initBatchModifyWeightTask(kt *kit.Kit, listenerList []*lblogic.ModifyWeightRecord,
	lb *corelb.BaseLoadBalancer) (string, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	tgIDs := make([]string, 0)
	for _, listener := range listenerList {
		// 获取每个监听器下的rs，构造异步任务
		targets, err := listener.GetTargets(kt, svc.client.DataService(), lb)
		if err != nil {
			return "", err
		}

		tmpTasks := buildBatchModifyWeightTask(listener.TargetGroupID, lb.ID, lb.Vendor, targets)
		tasks = append(tasks, tmpTasks...)
		tgIDs = append(tgIDs, listener.TargetGroupID)
	}

	// 对所有task进行排序并指定actionID
	getActionID := counter.NewNumStringCounter(0, 10)
	var lastActionID action.ActIDType
	for i := 0; i < len(tasks); i++ {
		actionID := action.ActIDType(getActionID())
		tasks[i].ActionID = actionID
		if len(lastActionID) > 0 {
			tasks[i].DependOn = []action.ActIDType{lastActionID}
		}
		lastActionID = actionID
	}

	flowID, err := svc.buildBatchOperationFlow(kt, lb.ID, enumor.FlowTargetGroupModifyWeight, tasks, tgIDs)
	if err != nil {
		logs.Errorf("build batch operation flow failed, err: %v, rid: %s, tasks: %v", err, kt.Rid, tasks)
		return "", err
	}
	return flowID, nil
}

func buildBatchModifyWeightTask(tgID, lbID string, vendor enumor.Vendor,
	targets []*cloud.TargetBaseReq) []ts.CustomFlowTask {

	tasks := make([]ts.CustomFlowTask, 0)
	elems := slice.Split(targets, constant.BatchModifyTargetWeightCloudMaxLimit)

	for _, parts := range elems {
		rsWeightParams := &hcproto.TCloudBatchOperateTargetReq{
			TargetGroupID: tgID,
			LbID:          lbID,
			RsList:        parts,
		}

		tmpTask := ts.CustomFlowTask{
			ActionName: enumor.ActionTargetGroupModifyWeight,
			Params: &actionlb.OperateRsOption{
				Vendor:                      vendor,
				TCloudBatchOperateTargetReq: *rsWeightParams,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		tasks = append(tasks, tmpTask)
	}

	return tasks
}
