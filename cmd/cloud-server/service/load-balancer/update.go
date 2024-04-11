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
	"hcm/cmd/task-server/logics/flow"
	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	hclbproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// UpdateBizTCloudLoadBalancer  业务下更新clb
func (svc *lbSvc) UpdateBizTCloudLoadBalancer(cts *rest.Contexts) (any, error) {

	lbID := cts.PathParameter("id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hclbproto.TCloudLBUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.LoadBalancerCloudResType,
		lbID)
	if err != nil {
		logs.Errorf("getLoadBalancer resource vendor failed, id: %s, err: %s, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  baseInfo})
	if err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.LoadBalancerAuditResType, lbID, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	switch baseInfo.Vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Clb.Update(cts.Kit, lbID, req)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}

}

// UpdateBizTargetGroup update biz target group.
func (svc *lbSvc) UpdateBizTargetGroup(cts *rest.Contexts) (any, error) {
	return svc.updateTargetGroup(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) updateTargetGroup(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.TargetGroupCloudResType, id)
	if err != nil {
		return nil, err
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType:   meta.TargetGroup,
		Action:    meta.Update,
		BasicInfo: baseInfo,
	})
	if err != nil {
		logs.Errorf("update target group basic info auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.batchUpdateTCloudTargetGroup(cts, id)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

// 更新目标组基本信息
func (svc *lbSvc) batchUpdateTCloudTargetGroup(cts *rest.Contexts, id string) (interface{}, error) {
	req := new(cslb.TargetGroupUpdateReq)
	if err := cts.DecodeInto(cts); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	dbReq := &dataproto.TargetGroupUpdateReq{
		IDs:        append(req.IDs, id),
		Name:       req.Name,
		VpcID:      req.VpcID,
		CloudVpcID: req.CloudVpcID,
		Region:     req.Region,
		Protocol:   req.Protocol,
		Port:       req.Port,
		Weight:     req.Weight,
	}
	err := svc.client.DataService().TCloud.LoadBalancer.BatchUpdateTCloudTargetGroup(cts.Kit, dbReq)
	if err != nil {
		logs.Errorf("update tcloud target group failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateBizTargetGroupHealth update biz target group health check
func (svc *lbSvc) UpdateBizTargetGroupHealth(cts *rest.Contexts) (any, error) {
	return svc.updateTargetGroupHealth(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) updateTargetGroupHealth(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	any, error) {

	tgID := cts.PathParameter("id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target group id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.
		GetResBasicInfo(cts.Kit, enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		return nil, err
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType:   meta.TargetGroup,
		Action:    meta.Update,
		BasicInfo: baseInfo,
	})
	if err != nil {
		logs.Errorf("update target group health check auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.updateTCloudTargetGroupHealthCheck(cts, tgID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

func (svc *lbSvc) updateTCloudTargetGroupHealthCheck(cts *rest.Contexts, tgID string) (any, error) {

	req := new(hclbproto.HealthCheckUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 更新云上监听器
	if err := svc.updateRelatedListenerHealthCheck(cts.Kit, tgID, req); err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(req.HealthCheck)
	if err != nil {
		return nil, err
	}
	// 3. 更新db
	dbReq := &dataproto.TargetGroupUpdateReq{
		IDs:         []string{tgID},
		HealthCheck: tabletype.JsonField(bytes),
	}

	err = svc.client.DataService().TCloud.LoadBalancer.BatchUpdateTCloudTargetGroup(cts.Kit, dbReq)
	if err != nil {
		logs.Errorf("update db tcloud target group failed, err: %v,  req: %+v, rid: %s", dbReq, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *lbSvc) updateRelatedListenerHealthCheck(kt *kit.Kit, tgID string,
	healthReq *hclbproto.HealthCheckUpdateReq) error {
	// 1. 获取目标组关联监听器
	relListReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	relResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, relListReq)
	if err != nil {
		return err
	}
	if len(relResp.Details) == 0 {
		// 无关联关系 直接返回
		return nil
	}

	// 本地目标组只有一个关联的规则或者监听器
	rel := relResp.Details[0]
	// 2. 更新云上监听器/规则
	switch rel.ListenerRuleType {
	case enumor.Layer7RuleType:
		// 仅更新规则的健康检查字段
		req := &hclbproto.TCloudRuleUpdateReq{HealthCheck: healthReq.HealthCheck}
		err := svc.client.HCService().TCloud.Clb.UpdateUrlRule(kt, rel.LblID, rel.ListenerRuleID, req)
		if err != nil {
			logs.Errorf("fail to update health check of rule, err: %v, listener id: %s, rule id: %s, rid: %s",
				err, rel.LblID, rel.ListenerRuleID, kt.Rid)
			return err
		}
	case enumor.Layer4RuleType:
		err := svc.client.HCService().TCloud.Clb.UpdateListenerHealthCheck(kt, rel.LblID, healthReq)
		if err != nil {
			logs.Errorf("fail to update health check of listener, err: %v, listener id: %s,  rid: %s",
				err, rel.LblID, kt.Rid)
			return err
		}
	}
	return nil
}

// UpdateBizListener update biz listener.
func (svc *lbSvc) UpdateBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.updateListener(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) updateListener(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update listener request decode failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("update listener auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, accID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.batchUpdateTCloudListener(cts.Kit, req.Data, id)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *lbSvc) batchUpdateTCloudListener(kt *kit.Kit, body json.RawMessage, id string) (interface{}, error) {
	req := new(hclbproto.ListenerWithRuleUpdateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.client.HCService().TCloud.Clb.UpdateListener(kt, id, req)
	if err != nil {
		logs.Errorf("update tcloud listener failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateBizDomainAttr update biz domain attr.
func (svc *lbSvc) UpdateBizDomainAttr(cts *rest.Contexts) (interface{}, error) {
	return svc.updateDomainAttr(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) updateDomainAttr(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lbl_id is required")
	}

	req := new(hclbproto.DomainAttrUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update listener request decode failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		logs.Errorf("get listener resource vendor failed, lblID: %s, err: %s, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}

	err = authHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  baseInfo,
	})
	if err != nil {
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.Clb.UpdateDomainAttr(cts.Kit, lblID, req)
		if err != nil {
			logs.Errorf("update tcloud listener url rule domain attr failed, lblID: %s, req: %+v, err: %v, rid: %s",
				lblID, req, err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

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
		err = svc.batchUpdateTargetPortDb(kt, req, tgID, accountID)
		if err != nil {
			return nil, err
		}
		return &core.FlowStateResult{State: enumor.FlowSuccess}, nil
	}

	return svc.buildModifyTCloudTargetTasksPort(kt, req, ruleRelList.Details[0].LbID, tgID, accountID)
}

func (svc *lbSvc) batchUpdateTargetPortDb(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetPortReq,
	accountID, tgID string) error {

	rsPortParams, err := svc.convTCloudOperateTargetReq(kt, req.TargetIDs, tgID, accountID,
		converter.ValToPtr(req.NewPort), nil)
	if err != nil {
		return err
	}

	// 检查RS是否已绑定该目标组
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: []*dataproto.TargetUpdate{}}
	for _, item := range rsPortParams.RsList {
		tgReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("target_group_id", tgID),
				tools.RuleEqual("cloud_inst_id", item.CloudInstID),
				tools.RuleEqual("port", item.Port),
			),
			Page: core.NewDefaultBasePage(),
		}
		tmpRsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
		if err != nil {
			return err
		}
		if len(tmpRsList.Details) > 0 {
			updateReq.Targets = append(updateReq.Targets, &dataproto.TargetUpdate{
				ID:   tmpRsList.Details[0].ID,
				Port: converter.PtrToVal(item.NewPort),
			})
		}
	}
	if len(updateReq.Targets) == 0 {
		return nil
	}

	return svc.client.DataService().Global.LoadBalancer.BatchUpdateTarget(kt, updateReq)
}

func (svc *lbSvc) buildModifyTCloudTargetTasksPort(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetPortReq, lbID, tgID,
	accountID string) (interface{}, error) {

	// 预检测
	err := svc.checkResFlowRel(kt, lbID, tgID, enumor.LoadBalancerCloudResType, enumor.TargetGroupCloudResType)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	tasks := make([]ts.CustomFlowTask, 0)
	elems := slice.Split(req.TargetIDs, constant.BatchModifyTargetPortCloudMaxLimit)
	getActionID := counter.NewNumStringCounter(1, 10)
	for _, parts := range elems {
		rsPortParams, err := svc.convTCloudOperateTargetReq(kt, parts, tgID, accountID,
			converter.ValToPtr(req.NewPort), nil)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(getActionID()),
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
		})
	}
	removeReq := &ts.AddCustomFlowReq{Name: enumor.FlowTargetGroupModifyPort, Tasks: tasks, IsInitState: true}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, removeReq)
	if err != nil {
		logs.Errorf("call taskserver to batch modify target port custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
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
				SubResID:   tgID,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ModifyPortTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyPortTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

// BatchModifyBizTargetsWeight batch modify biz targets weight.
func (svc *lbSvc) BatchModifyBizTargetsWeight(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchModifyTargetWeight(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) batchModifyTargetWeight(cts *rest.Contexts,
	authHandler handler.ValidWithAuthHandler, bkBizID int64) (any, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch modify target weight request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		logs.Errorf("batch modify target weight auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.buildModifyTCloudTargetWeight(cts.Kit, req.Data, tgID, baseInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

func (svc *lbSvc) buildModifyTCloudTargetWeight(kt *kit.Kit, body json.RawMessage,
	tgID, accountID string) (interface{}, error) {

	req := new(cslb.TCloudBatchModifyTargetWeightReq)
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
		err = svc.batchUpdateTargetWeightDb(kt, req, tgID, accountID)
		if err != nil {
			return nil, err
		}
		return &core.FlowStateResult{State: enumor.FlowSuccess}, nil
	}

	return svc.buildModifyTCloudTargetTasksWeight(kt, req, ruleRelList.Details[0].ID, tgID, accountID)
}

func (svc *lbSvc) batchUpdateTargetWeightDb(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetWeightReq,
	accountID, tgID string) error {

	rsPortParams, err := svc.convTCloudOperateTargetReq(kt, req.TargetIDs, tgID, accountID,
		nil, converter.ValToPtr(req.NewWeight))
	if err != nil {
		return err
	}

	// 检查RS是否已绑定该目标组
	updateReq := &dataproto.TargetBatchUpdateReq{Targets: []*dataproto.TargetUpdate{}}
	for _, item := range rsPortParams.RsList {
		tgReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("target_group_id", tgID),
				tools.RuleEqual("cloud_inst_id", item.CloudInstID),
				tools.RuleEqual("port", item.Port),
			),
			Page: core.NewDefaultBasePage(),
		}
		tmpRsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
		if err != nil {
			return err
		}
		if len(tmpRsList.Details) > 0 {
			updateReq.Targets = append(updateReq.Targets, &dataproto.TargetUpdate{
				ID:     tmpRsList.Details[0].ID,
				Weight: item.Weight,
			})
		}
	}
	if len(updateReq.Targets) == 0 {
		return nil
	}

	return svc.client.DataService().Global.LoadBalancer.BatchUpdateTarget(kt, updateReq)
}

func (svc *lbSvc) buildModifyTCloudTargetTasksWeight(kt *kit.Kit, req *cslb.TCloudBatchModifyTargetWeightReq,
	lbID, tgID, accountID string) (interface{}, error) {

	// 预检测
	err := svc.checkResFlowRel(kt, lbID, tgID, enumor.LoadBalancerCloudResType, enumor.TargetGroupCloudResType)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	tasks := make([]ts.CustomFlowTask, 0)
	elems := slice.Split(req.TargetIDs, constant.BatchModifyTargetWeightCloudMaxLimit)
	getActionID := counter.NewNumStringCounter(1, 10)
	for _, parts := range elems {
		rsWeightParams, err := svc.convTCloudOperateTargetReq(kt, parts, tgID, accountID,
			nil, converter.ValToPtr(req.NewWeight))
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(getActionID()),
			ActionName: enumor.ActionTargetGroupModifyWeight,
			Params: &actionlb.OperateRsOption{
				Vendor:                      enumor.TCloud,
				TCloudBatchOperateTargetReq: *rsWeightParams,
			},
			Retry: &tableasync.Retry{
				Enable: true,
				Policy: &tableasync.RetryPolicy{
					Count:        constant.FlowRetryMaxLimit,
					SleepRangeMS: [2]uint{100, 200},
				},
			},
		})
	}
	rsWeightReq := &ts.AddCustomFlowReq{Name: enumor.FlowTargetGroupModifyWeight, Tasks: tasks, IsInitState: true}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, rsWeightReq)
	if err != nil {
		logs.Errorf("call taskserver to batch modify target weight custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
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
				SubResID:   tgID,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ModifyWeightTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, lbID, enumor.LoadBalancerCloudResType, flowID, enumor.ModifyWeightTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}
