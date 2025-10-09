/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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
	apits "hcm/pkg/api/task-server"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// DeleteBizListener delete biz listener.
func (svc *lbSvc) DeleteBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteListener(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) deleteListener(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(cslb.DeleteListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ListenerCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list listener basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}
	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.ListenerAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit listener failed, ids: %v, err: %v, rid: %s",
			basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	if err = svc.validateListenerTargetWeight(cts.Kit, req.IDs); err != nil {
		logs.Errorf("validate listener target weight failed, ids: %s, err: %v, rid: %s", req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	taskManagementID, err := svc.createTaskManagementForDelLbl(cts.Kit, bizID, req)
	if err != nil {
		return nil, err
	}
	return task.CreateTaskManagementResp{TaskManagementID: taskManagementID}, nil
}

func (svc *lbSvc) buildDelLblTaskManagement(kt *kit.Kit, bkBizID int64, accountID string) (
	string, enumor.Vendor, error) {

	// create task management
	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(kt, enumor.AccountCloudResType, accountID)
	if err != nil {
		logs.Errorf("get account info failed, accountID: %s, err: %v, rid: %s", accountID, err, kt.Rid)
		return "", "", err
	}

	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, accountInfo.Vendor, accountID,
		enumor.TaskManagementSourceAPI, enumor.TaskDeleteListener)
	if err != nil {
		logs.Errorf("create task management failed, bkBizID: %d, accountID: %s, err: %v, rid: %s",
			bkBizID, accountID, err, kt.Rid)
		return "", "", err
	}
	return taskManagementID, accountInfo.Vendor, nil

}

func (svc *lbSvc) createTaskManagementForDelLbl(kt *kit.Kit, bkBizID int64, req *cslb.DeleteListenerReq) (
	string, error) {

	// list listener
	listeners, err := svc.listListenersByIDs(kt, req.IDs)
	if err != nil {
		logs.Errorf("list listeners failed, ids: %v, err: %v, rid: %s", req.IDs, err, kt.Rid)
		return "", err
	}
	clbIDLblMap := classifier.ClassifySlice(listeners, func(item corelb.BaseListener) string {
		return item.LbID
	})
	for lbID := range clbIDLblMap {
		// 预检测-是否有执行中的负载均衡
		_, err = svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
		if err != nil {
			logs.Errorf("check res flow rel failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
			return "", err
		}
	}

	// create task management
	taskManagementID, vendor, err := svc.buildDelLblTaskManagement(kt, bkBizID, req.AccountID)
	if err != nil {
		logs.Errorf("build delete lbl task management failed, err: %v, bkBizID: %d, accountID: %s, rid: %s",
			err, bkBizID, req.AccountID, kt.Rid)
		return "", err
	}
	flowIDs := make([]string, 0)
	for lbID, listeners := range clbIDLblMap {
		// 一个clb 对应一个flow
		flowID, err := svc.buildDeleteListenerTask(kt, vendor, lbID, taskManagementID, bkBizID, listeners)
		if err != nil {
			logs.Errorf("build delete listener task failed, lbID: %s, err: %v, rid: %s",
				lbID, err, kt.Rid)
			return "", err
		}
		flowIDs = append(flowIDs, flowID)
	}
	if err = svc.updateTaskManagement(kt, taskManagementID, flowIDs...); err != nil {
		logs.Errorf("update task management failed, err: %v, taskManagementID: %s, flowIDs: %v, rid: %s",
			err, taskManagementID, flowIDs, kt.Rid)
		return "", err
	}

	return taskManagementID, nil
}

func (svc *lbSvc) buildDeleteListenerTask(kt *kit.Kit, vendor enumor.Vendor, lbID, taskManagementID string,
	bkBizID int64, listeners []corelb.BaseListener) (string, error) {

	var taskDetails []*taskManagementDetail
	var err error
	defer func() {
		if err == nil {
			return
		}
		taskDetailIDs := slice.Map(taskDetails, func(item *taskManagementDetail) string {
			return item.taskDetailID
		})
		if err := svc.updateTaskDetailState(kt, enumor.TaskDetailFailed, taskDetailIDs, err.Error()); err != nil {
			logs.Errorf("update task details state to failed failed, err: %v, taskDetails: %+v, rid: %s",
				err, taskDetails, kt.Rid)
		}
	}()

	tasks, taskDetails, err := svc.generateFlowTasks(kt, listeners, vendor, lbID, taskManagementID, bkBizID)
	if err != nil {
		logs.Errorf("generate flow tasks failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}
	flowID, err := svc.buildFlow(kt, enumor.FlowBatchTaskDeleteListener, tableasync.NewShareData(map[string]string{
		"lb_id": lbID,
	}), tasks)
	if err != nil {
		logs.Errorf("build flow failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}
	for _, detail := range taskDetails {
		detail.flowID = flowID
	}
	if err = svc.updateTaskDetails(kt, taskDetails); err != nil {
		logs.Errorf("update task details failed, err: %v, taskDetails: %+v, rid: %s", err, taskDetails, kt.Rid)
		return "", err
	}

	// 下面的的代码如果执行出现error，不需要修改taskDetail的状态, 目前flow已经创建完毕，taskDetail由flowTask维护
	if buildErr := svc.buildSubFlow(kt, flowID, lbID, nil, enumor.ListenerCloudResType,
		enumor.DeleteListenerTaskType); buildErr != nil {
		logs.Errorf("build sub flow failed, err: %v, lbID: %s, rid: %s", buildErr, lbID, kt.Rid)
		return "", buildErr
	}
	// 锁定负载均衡跟Flow的状态
	if lockErr := svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID,
		enumor.ApplyTargetGroupType); lockErr != nil {
		logs.Errorf("lock res flow status failed, err: %v, lbID: %s, rid: %s", lockErr, lbID, kt.Rid)
		return "", lockErr
	}
	return flowID, nil
}

// generateFlowTasks ...
func (svc *lbSvc) generateFlowTasks(kt *kit.Kit, listeners []corelb.BaseListener, vendor enumor.Vendor, lbID,
	taskManagementID string, bkBizID int64) ([]apits.CustomFlowTask, []*taskManagementDetail, error) {

	var tasks []apits.CustomFlowTask
	var taskDetails []*taskManagementDetail
	getNextID := counter.NewNumberCounterWithPrev(1, 10)
	for _, batch := range slice.Split(listeners, constant.BatchDeleteListenerCloudMaxLimit) {
		details, err := svc.createDeleteListenerTaskDetails(kt, taskManagementID, bkBizID, batch)
		if err != nil {
			logs.Errorf("create delete listener task details failed, err: %v, lbID: %s, rid: %s",
				err, lbID, kt.Rid)
			return nil, nil, err
		}
		cur, prev := getNextID()
		tmpTask := apits.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskDeleteListener,
			Params: &actionlb.BatchTaskDeleteListenerOption{
				Vendor:         vendor,
				LoadBalancerID: lbID,
				ManagementDetailIDs: slice.Map(details, func(item *taskManagementDetail) string {
					return item.taskDetailID
				}),
				BatchDeleteReq: &core.BatchDeleteReq{
					IDs: slice.Map(batch, func(item corelb.BaseListener) string {
						return item.ID
					}),
				},
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		for _, detail := range details {
			detail.actionID = cur
		}
		tasks = append(tasks, tmpTask)
		taskDetails = append(taskDetails, details...)
	}
	return tasks, taskDetails, nil
}

func (svc *lbSvc) createDeleteListenerTaskDetails(kt *kit.Kit, taskManagementID string, bkBizID int64,
	listeners []corelb.BaseListener) ([]*taskManagementDetail, error) {

	details := make([]*taskManagementDetail, 0)
	for _, listener := range listeners {
		detail := &taskManagementDetail{
			param: listener,
		}
		details = append(details, detail)
	}
	if _, err := svc.createTaskDetails(kt, taskManagementID, bkBizID,
		enumor.TaskDeleteListener, details); err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
		return nil, err
	}
	return details, nil
}

// validateListenerTargetWeight 校验监听器绑定的所有rs权重是否为0
func (svc *lbSvc) validateListenerTargetWeight(kt *kit.Kit, ids []string) error {
	for _, id := range ids {
		stat, err := svc.getListenerTargetWeightStat(kt, id)
		if err != nil {
			return err
		}
		if stat.NonZeroWeightCount > 0 {
			return fmt.Errorf("listener %s has targets with non-zero weight", id)
		}
	}
	return nil
}

// getTGIDsByListenerID 七层监听器会对应多个目标组，四层监听器只有一个目标组
func (svc *lbSvc) getTGIDsByListenerID(kt *kit.Kit, listenerID, loadBalancerID string) ([]string, error) {
	targetGroupIDs := make([]string, 0)
	listTGReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lbl_id", listenerID),
			tools.RuleEqual("lb_id", loadBalancerID),
		),
		Page: core.NewDefaultBasePage(),
	}
	for {
		rels, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, listTGReq)
		if err != nil {
			logs.Errorf("list target group listener rel failed, req: %+v, err: %v, rid: %s", listTGReq, err, kt.Rid)
			return nil, err
		}
		for _, detail := range rels.Details {
			targetGroupIDs = append(targetGroupIDs, detail.TargetGroupID)
		}
		if len(rels.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		listTGReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return targetGroupIDs, nil
}

func (svc *lbSvc) getListenerTargetWeightStat(kt *kit.Kit, listenerID string) (*cslb.ListenerTargetsStat, error) {

	lbl, err := svc.getListenerByID(kt, listenerID)
	if err != nil {
		logs.Errorf("get listener failed, listenerID: %s, err: %v, rid: %s",
			listenerID, err, kt.Rid)
		return nil, err
	}

	targetGroupIDs, err := svc.getTGIDsByListenerID(kt, listenerID, lbl.LbID)
	if err != nil {
		logs.Errorf("get target group ids by listener id failed, listenerID: %s, err: %v, rid: %s",
			listenerID, err, kt.Rid)
		return nil, err
	}
	result := &cslb.ListenerTargetsStat{}
	for _, batch := range slice.Split(targetGroupIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("target_group_id", batch),
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			targets, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listReq)
			if err != nil {
				logs.Errorf("list target failed, req: %+v, err: %v, rid: %s", listReq, err, kt.Rid)
				return nil, err
			}
			for _, detail := range targets.Details {
				if converter.PtrToVal(detail.Weight) == 0 {
					result.ZeroWeightCount++
				}
			}
			result.TotalCount += len(targets.Details)
			if len(targets.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	result.NonZeroWeightCount = result.TotalCount - result.ZeroWeightCount
	return result, nil
}

// ListBizListenerTargetWeightStat list biz listener rs weight stat.
func (svc *lbSvc) ListBizListenerTargetWeightStat(cts *rest.Contexts) (interface{}, error) {
	req := new(cslb.ListListenerTargetsStatReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ListenerCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list listener basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}
	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	result := make(map[string]*cslb.ListenerTargetsStat)
	for _, id := range req.IDs {
		stat, err := svc.getListenerTargetWeightStat(cts.Kit, id)
		if err != nil {
			logs.Errorf("get listener target weight stat failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
			return nil, err
		}
		result[id] = stat
	}

	return result, nil
}
