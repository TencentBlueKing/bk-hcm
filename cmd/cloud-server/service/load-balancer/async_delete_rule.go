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
	corelb "hcm/pkg/api/core/cloud/load-balancer"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchDeleteBizRule batch delete biz rule
func (svc *lbSvc) BatchDeleteBizRule(cts *rest.Contexts) (any, error) {
	return svc.batchDeleteBizRule(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchDeleteBizRule(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete rule request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.UrlRuleAuditResType,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch delete rule auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch delete rule auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		return svc.buildDeleteTCloudRule(cts.Kit, req.Data, enumor.TCloud)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildDeleteTCloudRule(kt *kit.Kit, body json.RawMessage, vendor enumor.Vendor) (interface{}, error) {
	req := new(cslb.TcloudBatchDeleteRuleReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var lbUrlRuleMap, lbListenerMap map[string][]string
	var err error
	if len(req.URLRuleIDs) != 0 {
		lbUrlRuleMap, err = svc.checkURLRuleExistsAndGroupByLb(kt, req.URLRuleIDs)
		if err != nil {
			logs.Errorf("delete rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("delete rule failed, err: %v, rid: %s", err, kt.Rid)
		}
	}
	if len(req.ListenerIDs) != 0 {
		lbListenerMap, err = svc.checkListenerExistsAndGroupByLb(kt, req.ListenerIDs)
		if err != nil {
			logs.Errorf("delete rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("delete rule failed, err: %v, rid: %s", err, kt.Rid)
		}
	}

	lbRuleMap := make(map[string]cslb.TcloudBatchDeleteRuleIDs)
	for lbID, urlRuleIDs := range lbUrlRuleMap {
		ruleIDs := cslb.TcloudBatchDeleteRuleIDs{
			URLRuleIDs: urlRuleIDs,
		}
		lbRuleMap[lbID] = ruleIDs
	}
	for lbID, listenerIDs := range lbListenerMap {
		ruleIDs, exists := lbRuleMap[lbID]
		if !exists {
			ruleIDs = cslb.TcloudBatchDeleteRuleIDs{}
		}
		ruleIDs.ListenerIDs = listenerIDs
		lbRuleMap[lbID] = ruleIDs
	}

	return svc.buildDeleteRuleTasks(kt, lbRuleMap, vendor)
}

func (svc *lbSvc) checkURLRuleExistsAndGroupByLb(kt *kit.Kit, urlRuleIDs []string) (map[string][]string, error) {
	// 查询URLRule是否存在，并以负载均衡为粒度进行分组
	lbUrlRuleMap := make(map[string][]string)

	urlRuleLbMap := make(map[string]string)
	urlRuleBaseInfoReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", urlRuleIDs)),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "lb_id"},
	}
	for {
		urlRuleBaseInfoResult, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleBaseInfoReq)
		if err != nil {
			logs.Errorf("list url rule failed, req: %v, err: %v, rid: %s", urlRuleBaseInfoReq, err, kt.Rid)
			return nil, err
		}
		for _, urlRuleBaseInfo := range urlRuleBaseInfoResult.Details {
			urlRuleLbMap[urlRuleBaseInfo.ID] = urlRuleBaseInfo.LbID
		}

		if uint(len(urlRuleBaseInfoResult.Details)) < core.DefaultMaxPageLimit {
			break
		}
		urlRuleBaseInfoReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	for _, urlRuleID := range urlRuleIDs {
		lbID, exit := urlRuleLbMap[urlRuleID]
		if !exit {
			return nil, fmt.Errorf("url rule: %s has not exists", urlRuleID)
		}
		if _, exists := lbUrlRuleMap[lbID]; !exists {
			lbUrlRuleMap[lbID] = make([]string, 0)
		}
		lbUrlRuleMap[lbID] = append(lbUrlRuleMap[lbID], urlRuleID)
	}

	return lbUrlRuleMap, nil
}

func (svc *lbSvc) checkListenerExistsAndGroupByLb(kt *kit.Kit, listenerIDs []string) (map[string][]string, error) {
	// 查询Listener是否存在，并以负载均衡为粒度进行分组
	listenerList := make([]corelb.Listener[corelb.TCloudListenerExtension], 0)
	listListenerReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", listenerIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		listenerResult, err := svc.client.DataService().TCloud.LoadBalancer.ListListener(kt, listListenerReq)
		if err != nil {
			logs.Errorf("list listener failed, req: %v, err: %v, rid: %s", listListenerReq, err, kt.Rid)
			return nil, fmt.Errorf("find listener failed, req: %v, err: %v, rid: %s", listListenerReq, err, kt.Rid)
		}
		listenerList = append(listenerList, listenerResult.Details...)

		if uint(len(listenerResult.Details)) < core.DefaultMaxPageLimit {
			break
		}

		listListenerReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	lbListenerMap := make(map[string][]string)
	for _, listener := range listenerList {
		if _, exists := lbListenerMap[listener.LbID]; !exists {
			lbListenerMap[listener.LbID] = make([]string, 0)
		}

		lbListenerMap[listener.LbID] = append(lbListenerMap[listener.LbID], listener.ID)
	}

	return lbListenerMap, nil
}

func (svc *lbSvc) buildDeleteRuleTasks(kt *kit.Kit, lbRuleMap map[string]cslb.TcloudBatchDeleteRuleIDs,
	vendor enumor.Vendor) ([]*core.FlowStateResult, error) {
	flowStateResults := make([]*core.FlowStateResult, 0)

	for lbID, ruleIDs := range lbRuleMap {
		// 预检测
		_, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
		if err != nil {
			logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		// 创建Flow跟Task的初始化数据
		flowID, err := svc.initFlowDeleteRule(kt, lbID, ruleIDs, vendor)
		if err != nil {
			logs.Errorf("init flow batch delete rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		// 锁定资源跟Flow的状态
		err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.DeleteRuleTaskType)
		if err != nil {
			logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		flowStateResults = append(flowStateResults, &core.FlowStateResult{FlowID: flowID})
	}

	return flowStateResults, nil
}

func (svc *lbSvc) initFlowDeleteRule(kt *kit.Kit, lbID string, ruleIDs cslb.TcloudBatchDeleteRuleIDs,
	vendor enumor.Vendor) (string, error) {

	getActionID := counter.NewNumberCounterWithPrev(1, 10)
	tasks := buildDeleteUrlRuleTasks(vendor, lbID, getActionID, ruleIDs.URLRuleIDs)
	deleteListenerTasks := buildDeleteListenerTasks(vendor, lbID, getActionID, ruleIDs.ListenerIDs)
	tasks = append(tasks, deleteListenerTasks...)

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowLoadBalancerDeleteRule,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch delete load balancer rule custom flow failed, err: %v, req: %+v, rid: %s",
			err, converter.PtrToVal(addReq), kt.Rid)
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
				SubResIDs:  nil,
				SubResType: enumor.ListenerCloudResType,
				TaskType:   enumor.DeleteRuleTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, lbID: %s,"+
			" tastType: %s, flowID: %s, rid: %s", err, lbID, enumor.DeleteRuleTaskType, flowID, kt.Rid)
		return "", err
	}
	return flowID, nil
}

func buildDeleteListenerTasks(vendor enumor.Vendor, lbID string, getActionID func() (cur string, prev string),
	listenerIDs []string) []ts.CustomFlowTask {

	tasks := make([]ts.CustomFlowTask, 0)
	for _, parts := range slice.Split(listenerIDs, constant.BatchDeleteListenerCloudMaxLimit) {
		cur, prev := getActionID()
		actionID := action.ActIDType(cur)
		tmpTask := ts.CustomFlowTask{
			ActionID:   actionID,
			ActionName: enumor.ActionLoadBalancerDeleteListener,
			Params: actionlb.DeleteListenerOption{
				Vendor:      vendor,
				LbID:        lbID,
				ListenerIDs: parts,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if len(prev) > 0 {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		tasks = append(tasks, tmpTask)
	}
	return tasks
}

func buildDeleteUrlRuleTasks(vendor enumor.Vendor, lbID string, getActionID func() (cur string, prev string),
	urlRuleIDs []string) []ts.CustomFlowTask {

	tasks := make([]ts.CustomFlowTask, 0)
	for _, parts := range slice.Split(urlRuleIDs, constant.BatchDeleteUrlRuleCloudMaxLimit) {
		cur, prev := getActionID()
		actionID := action.ActIDType(cur)
		tmpTask := ts.CustomFlowTask{
			ActionID:   actionID,
			ActionName: enumor.ActionLoadBalancerDeleteUrlRule,
			Params: &actionlb.DeleteURLRuleOption{
				Vendor:     vendor,
				LbID:       lbID,
				URLRuleIDs: parts,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if len(prev) > 0 {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		tasks = append(tasks, tmpTask)
	}
	return tasks
}
