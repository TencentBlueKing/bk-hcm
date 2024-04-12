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

// Package loadbalancer ...
package loadbalancer

import (
	"encoding/json"
	"errors"
	"fmt"

	"hcm/cmd/cloud-server/logics/async"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/cmd/task-server/logics/flow"
	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
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
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// DeleteBizTargetGroup delete biz target group.
func (svc *lbSvc) DeleteBizTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteTargetGroup(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) deleteTargetGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.TargetGroupCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list target group basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.TargetGroupAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit target group failed, ids: %v, err: %v, rid: %s",
			basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	// delete tcloud cloud target group
	err = svc.client.DataService().Global.LoadBalancer.DeleteTargetGroup(cts.Kit, &core.ListReq{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("[%s] request dataservice to delete target group failed, ids: %s, err: %v, rid: %s",
			enumor.TCloud, req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteBizListener delete biz listener.
func (svc *lbSvc) DeleteBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteListener(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) deleteListener(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.BatchDeleteReq)
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

	// delete tcloud cloud listener
	err = svc.client.HCService().TCloud.Clb.DeleteListener(cts.Kit, req)
	if err != nil {
		logs.Errorf("[%s] request hcservice to delete listener failed, ids: %s, err: %v, rid: %s",
			enumor.TCloud, req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRemoveBizTargets batch remove biz targets.
func (svc *lbSvc) BatchRemoveBizTargets(cts *rest.Contexts) (any, error) {
	return svc.batchRemoveBizTarget(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) batchRemoveBizTarget(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch remove target request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
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

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.buildRemoveTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildRemoveTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string) (interface{},
	error) {

	req := new(cslb.TCloudTargetBatchRemoveReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	targetIDs := make([]string, 0)
	lbIDs := make([]string, 0)
	targetGroupIDMap := make(map[string][]string, 0)
	targetGroupRuleRelMap := make(map[string][]corelb.BaseTargetListenerRuleRel, 0)
	for _, item := range req.TargetGroups {
		if _, ok := targetGroupRuleRelMap[item.TargetGroupID]; !ok {
			// 根据目标组ID，获取目标组绑定的监听器、规则列表
			ruleRelReq := &core.ListReq{
				Filter: tools.EqualExpression("target_group_id", item.TargetGroupID),
				Page:   core.NewDefaultBasePage(),
			}
			ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
			if err != nil {
				logs.Errorf("list tcloud listener url rule failed, tgItem: %+v, err: %v, rid: %s", item, err, kt.Rid)
				return nil, err
			}
			targetGroupRuleRelMap[item.TargetGroupID] = ruleRelList.Details
		}

		// 该目标组尚未绑定监听器及规则，不需要云端操作
		ruleRelList := targetGroupRuleRelMap[item.TargetGroupID]
		if len(ruleRelList) == 0 {
			err := svc.batchDeleteTargetDb(kt, accountID, item.TargetGroupID, item.TargetIDs)
			if err != nil {
				return nil, err
			}
			targetIDs = append(targetIDs, item.TargetIDs...)
		} else {
			lbIDs = slice.Unique(slice.Map(ruleRelList, func(rel corelb.BaseTargetListenerRuleRel) string {
				return rel.LbID
			}))
			targetGroupIDMap[item.TargetGroupID] = append(targetGroupIDMap[item.TargetGroupID], item.TargetIDs...)
		}
	}

	// 都是未绑定监听器的目标组，不需要云端操作
	if len(targetGroupIDMap) == 0 {
		return &corelb.TargetOperateResult{TargetIDs: targetIDs}, nil
	}

	// 目标组需要属于同一个负载均衡
	if len(lbIDs) > 1 {
		return nil, errf.New(errf.InvalidParameter, "target group need belong to the same load balancer")
	}

	return svc.buildRemoveTCloudTargetTasks(kt, accountID, lbIDs[0], targetGroupIDMap)
}

func (svc *lbSvc) batchDeleteTargetDb(kt *kit.Kit, accountID, tgID string, targetIDs []string) error {
	rsIDs := make([]string, 0)
	tgReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", targetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
	if err != nil {
		return err
	}
	if len(rsList.Details) > 0 {
		rsIDs = append(rsIDs, rsList.Details[0].ID)
	}

	if len(rsIDs) == 0 {
		return nil
	}

	delReq := &dataproto.LoadBalancerBatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", rsIDs),
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("target_group_id", tgID),
		),
	}
	return svc.client.DataService().Global.LoadBalancer.BatchDeleteTarget(kt, delReq)
}

func (svc *lbSvc) buildRemoveTCloudTargetTasks(kt *kit.Kit, accountID, lbID string, tgMap map[string][]string) (
	*core.FlowStateResult, error) {

	// 预检测
	err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	flowID, err := svc.initFlowRemoveTargetByLbID(kt, accountID, lbID, tgMap)
	if err != nil {
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.RemoveRSTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

func (svc *lbSvc) initFlowRemoveTargetByLbID(kt *kit.Kit, accountID string, lbID string, tgMap map[string][]string) (
	string, error) {

	tasks := make([]ts.CustomFlowTask, 0)
	getActionID := counter.NewNumStringCounter(1, 10)
	var tgIDs []string
	var lastActionID action.ActIDType
	for tgID, rsList := range tgMap {
		tgIDs = append(tgIDs, tgID)
		elems := slice.Split(rsList, constant.BatchRemoveRSCloudMaxLimit)
		for _, parts := range elems {
			removeRsParams, err := svc.convTCloudOperateTargetReq(kt, parts, tgID, accountID, nil, nil)
			if err != nil {
				return "", err
			}
			actionID := action.ActIDType(getActionID())
			tmpTask := ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionTargetGroupRemoveRS,
				Params: &actionlb.OperateRsOption{
					Vendor:                      enumor.TCloud,
					TCloudBatchOperateTargetReq: *removeRsParams,
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
	}
	removeReq := &ts.AddCustomFlowReq{Name: enumor.FlowTargetGroupRemoveRS, Tasks: tasks, IsInitState: true}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, removeReq)
	if err != nil {
		logs.Errorf("call taskserver to batch remove rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
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
				TaskType:   enumor.RemoveRSTaskType,
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

// convTCloudOperateTargetReq conv tcloud operate target req.
func (svc *lbSvc) convTCloudOperateTargetReq(kt *kit.Kit, targetIDs []string, targetGroupID,
	accountID string, newPort, newWeight *int64) (*hcproto.TCloudBatchOperateTargetReq, error) {

	targetReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", targetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	targetList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
	if err != nil {
		logs.Errorf("failed to list target by id, targetIDs: %v, err: %v, rid: %s", targetIDs, err, kt.Rid)
		return nil, err
	}
	if len(targetList.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target_ids: %v is not found", targetIDs)
	}

	instExistsMap := make(map[string]struct{}, 0)
	rsReq := &hcproto.TCloudBatchOperateTargetReq{TargetGroupID: targetGroupID}
	for _, item := range targetList.Details {
		// 批量修改端口时，需要校验重复的实例ID的问题，否则云端接口也会报错
		if cvt.PtrToVal(newPort) > 0 {
			if _, ok := instExistsMap[item.CloudInstID]; ok {
				return nil, errf.Newf(errf.RecordDuplicated, "duplicate modify same inst(%s) to new_port: %d",
					item.CloudInstID, cvt.PtrToVal(newPort))
			}
			instExistsMap[item.CloudInstID] = struct{}{}
		}

		rsReq.RsList = append(rsReq.RsList, &dataproto.TargetBaseReq{
			ID:               item.ID,
			InstType:         item.InstType,
			CloudInstID:      item.CloudInstID,
			Port:             item.Port,
			Weight:           item.Weight,
			AccountID:        accountID,
			TargetGroupID:    targetGroupID,
			InstName:         item.InstName,
			PrivateIPAddress: item.PrivateIPAddress,
			PublicIPAddress:  item.PublicIPAddress,
			CloudVpcIDs:      item.CloudVpcIDs,
			Zone:             item.Zone,
			NewPort:          newPort,
			NewWeight:        newWeight,
		})
	}
	return rsReq, nil
}

// BatchDeleteLoadBalancer 批量删除负载均衡
func (svc *lbSvc) BatchDeleteLoadBalancer(cts *rest.Contexts) (any, error) {
	return svc.batchDeleteLoadBalancer(cts, handler.ResOperateAuth)
}

// BatchDeleteBizLoadBalancer 业务下批量删除负载均衡
func (svc *lbSvc) BatchDeleteBizLoadBalancer(cts *rest.Contexts) (any, error) {
	return svc.batchDeleteLoadBalancer(cts, handler.BizOperateAuth)
}

// batchDeleteLoadBalancer 批量删除负载均衡
func (svc *lbSvc) batchDeleteLoadBalancer(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (any, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	infoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.LoadBalancerCloudResType,
		IDs:          req.IDs,
		Fields:       append(types.CommonBasicInfoFields, "region"),
	}
	lbInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, infoReq)
	if err != nil {
		return nil, err
	}
	for _, lbID := range req.IDs {
		info, exist := lbInfoMap[lbID]
		if !exist {
			return nil, fmt.Errorf("load balancer(%s) not found", lbID)
		}
		if info.Vendor != enumor.TCloud {
			return nil, errors.New("only supports tcloud")
		}

	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Delete,
		BasicInfos: lbInfoMap,
	})
	if err != nil {
		return nil, err
	}

	if err = svc.loadBalancerDeleteCheck(cts.Kit, req.IDs); err != nil {
		return nil, err
	}
	// 按规则删除审计
	err = svc.audit.ResDeleteAudit(cts.Kit, enumor.LoadBalancerAuditResType, req.IDs)
	if err != nil {
		logs.Errorf("create load balancer delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 按账号+地域分列表
	tasks := buildTCloudLBDeletionTasks(lbInfoMap)
	flowReq := &ts.AddCustomFlowReq{
		Name:        enumor.FlowDeleteLoadBalancer,
		ShareData:   nil,
		Tasks:       tasks,
		IsInitState: false,
	}
	flowResp, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		return nil, err
	}
	return nil, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), flowResp.ID)
}

// 负载均衡删除检查
func (svc *lbSvc) loadBalancerDeleteCheck(kt *kit.Kit, lbIDs []string) error {
	// 检查是否存在监听器
	lblListReq := &core.ListReq{
		Filter: tools.ContainersExpression("lb_id", lbIDs),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblListReq)
	if err != nil {
		logs.Errorf("fail to query listener for delete load balancers, err: %v, lb ids: %v, rid: %s",
			err, lbIDs, kt.Rid)
		return nil
	}
	if len(listenerResp.Details) != 0 {
		lbl := listenerResp.Details[0]
		return fmt.Errorf("load balancer(%s) with listener(%s:%s) can not be deleted",
			lbl.CloudLbID, lbl.CloudID, lbl.Name)
	}
	return nil
}

func buildTCloudLBDeletionTasks(infoMap map[string]types.CloudResourceBasicInfo) (tasks []ts.CustomFlowTask) {

	tcloudReqMap := make(map[string]*hcproto.TCloudBatchDeleteLoadbalancerReq, len(infoMap))
	// TODO: 后期支持多vendor
	for id, info := range infoMap {
		key := genAccountRegionKey(info)
		if tcloudReqMap[key] == nil {
			tcloudReqMap[key] = &hcproto.TCloudBatchDeleteLoadbalancerReq{
				AccountID: info.AccountID,
				Region:    info.Region,
				IDs:       []string{},
			}

		}
		req := tcloudReqMap[key]
		req.IDs = append(req.IDs, id)
	}
	getNextID := counter.NewNumStringCounter(1, 10)
	for _, req := range tcloudReqMap {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(getNextID()),
			ActionName: enumor.ActionDeleteLoadBalancer,
			Params: actionlb.DeleteLoadBalancerOption{
				Vendor:                           enumor.TCloud,
				TCloudBatchDeleteLoadbalancerReq: cvt.PtrToVal(req),
			},
			Retry: tableasync.NewRetryWithPolicy(3, 1000, 5000),
		})

	}
	return tasks
}

func genAccountRegionKey(info types.CloudResourceBasicInfo) string {
	return info.AccountID + "_" + info.Region
}
