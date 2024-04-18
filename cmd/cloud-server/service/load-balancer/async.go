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
	"errors"
	"time"

	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/producer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BizTerminateFlow 终止flow
func (svc *lbSvc) BizTerminateFlow(cts *rest.Contexts) (any, error) {
	return svc.terminateFlow(cts, handler.BizOperateAuth)
}

// BizRetryTask ...
func (svc *lbSvc) BizRetryTask(cts *rest.Contexts) (any, error) {
	return svc.retryTask(cts, handler.BizOperateAuth)
}

// BizCloneFlow ....
func (svc *lbSvc) BizCloneFlow(cts *rest.Contexts) (any, error) {
	return svc.cloneFlow(cts, handler.BizOperateAuth)
}

// BizGetResultAfterTerminate ...
func (svc *lbSvc) BizGetResultAfterTerminate(cts *rest.Contexts) (any, error) {
	return svc.getResultAfterTerminate(cts, handler.BizOperateAuth)
}

// CancelFlow 终止flow
// 1. 检查负载均衡操作权限
// 2. 检查对应res_flow_rel状态，终止应该是处于 executing
// 3. 调用task server 终止
func (svc *lbSvc) terminateFlow(cts *rest.Contexts,
	operateAuth handler.ValidWithAuthHandler) (any, error) {

	// check lb operation perm first
	lbInfo, err := svc.getAndCheckLBPerm(cts, operateAuth)
	if err != nil {
		return nil, err
	}

	req := new(cslb.AsyncFlowTerminateReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rel, err := svc.getLoadBalancerFlowRel(cts.Kit, lbInfo.ID, req.FlowID)
	if err != nil {
		return nil, err
	}
	// 需要对应flow没有被逻辑终止

	if rel.Status != enumor.ExecutingResFlowStatus {
		return nil, errf.Newf(errf.InvalidParameter, "given flow status incorrect: %s", rel.Status)
	}
	// 从-flow 检查到任务结束会自动解锁
	err = svc.client.TaskServer().CancelFlow(cts.Kit, req.FlowID)
	if err != nil {
		logs.Errorf("fail to call task server to terminate flow(%s), err: %s, rid: %s", req.FlowID, err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// RetryTask 重试子任务 要求对应的rel为 executing, 且资源可以访问
func (svc *lbSvc) retryTask(cts *rest.Contexts,
	operateAuth handler.ValidWithAuthHandler) (any, error) {

	// check lb operate perm
	lbInfo, err := svc.getAndCheckLBPerm(cts, operateAuth)
	if err != nil {
		return nil, err
	}

	req := new(cslb.AsyncTaskRetryReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 对应的rel表 是否已经取消或失败
	rel, err := svc.getLoadBalancerFlowRel(cts.Kit, lbInfo.ID, req.FlowID)
	if err != nil {
		return nil, err
	}
	if rel.Status != enumor.ExecutingResFlowStatus {
		return nil, errf.Newf(errf.InvalidParameter, "given flow status incorrect: %s", rel.Status)
	}
	err = svc.client.TaskServer().RetryTask(cts.Kit, req.FlowID, req.TaskID)
	if err != nil {
		logs.Errorf("fail to call task server to retry flow(%s),task(%s), err: %s, rid: %s",
			req.FlowID, req.TaskID, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CloneFlow 重新发起
func (svc *lbSvc) cloneFlow(cts *rest.Contexts, operateAuth handler.ValidWithAuthHandler) (any, error) {

	// check lb operate perm
	lbInfo, err := svc.getAndCheckLBPerm(cts, operateAuth)
	if err != nil {
		return nil, err
	}

	req := new(cslb.AsyncFlowCloneReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rel, err := svc.getLoadBalancerFlowRel(cts.Kit, lbInfo.ID, req.FlowID)
	if err != nil {
		return nil, err
	}
	// 只能是终态 才能重新发起
	if rel.Status != enumor.CancelResFlowStatus &&
		rel.Status != enumor.TimeoutResFlowStatus &&
		rel.Status != enumor.SuccessResFlowStatus {
		return nil, errf.Newf(errf.InvalidParameter, "given flow status incorrect: %s", rel.Status)
	}

	cloneReq := &producer.CloneFlowOption{
		Memo:        "cloned for " + req.FlowID,
		IsInitState: true,
	}
	flowRet, err := svc.client.TaskServer().CloneFlow(cts.Kit, req.FlowID, cloneReq)
	if err != nil {
		logs.Errorf("fail to call task server to clone flow(%s), err: %s, rid: %s", req.FlowID, err, cts.Kit.Rid)
		return nil, err
	}
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowRet.ID,
				ResID:      lbInfo.ID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  []string{lbInfo.ID},
				SubResType: enumor.LoadBalancerCloudResType,
				TaskType:   rel.TaskType,
			},
		}},
	}

	_, err = svc.client.TaskServer().CreateTemplateFlow(cts.Kit, flowWatchReq)
	if err != nil {
		logs.Errorf("call task server to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, req.FlowID, cts.Kit.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(cts.Kit, lbInfo.ID, enumor.LoadBalancerCloudResType, req.FlowID,
		rel.TaskType)
	if err != nil {
		return nil, err
	}

	return flowRet, nil
}

// GetResultAfterTerminate 获取结束后的result
func (svc *lbSvc) getResultAfterTerminate(cts *rest.Contexts, operateAuth handler.ValidWithAuthHandler) (any, error) {
	// TODO
	return nil, nil
}

func (svc *lbSvc) getAndCheckLBPerm(cts *rest.Contexts,
	operateAuth handler.ValidWithAuthHandler) (*types.CloudResourceBasicInfo, error) {

	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errors.New("lb_id is required")
	}

	// 获取操作记录详情
	lbInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.LoadBalancerCloudResType, lbID)
	if err != nil {
		logs.Errorf("get load balancer basic info failed, id: %d, err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	err = operateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  lbInfo,
	})
	if err != nil {
		return nil, err
	}
	return lbInfo, err
}

func (svc *lbSvc) getLoadBalancerFlowRel(kt *kit.Kit, lbID, flowID string) (*corelb.BaseResFlowRel, error) {
	aWeekAgo := time.Now().Add(-time.Hour * 24 * constant.ResFlowLockExpireDays)
	// 检查任务是否 已终止
	relListReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", lbID),
			tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
			tools.RuleEqual("flow_id", flowID),
			tools.RuleGreaterThan("create_at", aWeekAgo.Format(constant.DateTimeLayout)),
		),
		Page: core.NewDefaultBasePage(),
	}
	relResp, err := svc.client.DataService().Global.LoadBalancer.ListResFlowRel(kt, relListReq)
	if err != nil {
		return nil, err
	}
	if len(relResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "relation of flow(%s) not found", flowID)
	}

	return &relResp.Details[0], nil
}
