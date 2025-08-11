/*
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
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	apits "hcm/pkg/api/task-server"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
)

// CreateBizUrlRule 业务下新建url规则 TODO: 改成一次只创建一个规则
func (svc *lbSvc) CreateBizUrlRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID < 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id id is required")
	}

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	// 限制一次只能创建一条规则
	req := new(cslb.TCloudRuleCreate)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, lblBasicInfo, err := svc.getListenerByIDAndBiz(cts.Kit, vendor, bizID, lblID)
	if err != nil {
		logs.Errorf("fail to get listener info, bizID: %d, listenerID: %s, err: %v, rid: %s",
			bizID, lblID, err, cts.Kit.Rid)
		return nil, err
	}

	// if SNI Switch is off, certificates can only be set in listener not its rule
	if lblInfo.SniSwitch == enumor.SniTypeClose && req.Certificate != nil {
		return nil, errf.New(errf.InvalidParameter, "can not set certificate on rule of sni_switch off listener")
	}

	// 业务校验、鉴权
	valOpt := &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Create,
		BasicInfo:  lblBasicInfo,
	}
	if err = handler.BizOperateAuth(cts, valOpt); err != nil {
		return nil, err
	}

	tg, err := svc.targetGroupBindCheck(cts.Kit, bizID, req.TargetGroupID)
	if err != nil {
		return nil, err
	}

	// 预检测-是否有执行中的负载均衡
	_, err = svc.checkResFlowRel(cts.Kit, lblInfo.LbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	createResp, err := svc.batchCreateUrlRule(cts.Kit, vendor, lblID, req, tg)
	if err != nil {
		logs.Errorf("fail to create url rule, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = svc.applyTargetToRule(cts.Kit, tg.ID, createResp.SuccessCloudIDs[0], lblInfo)
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return createResp, nil
}

// batchCreateUrlRule 批量创建url规则
func (svc *lbSvc) batchCreateUrlRule(kt *kit.Kit, vendor enumor.Vendor, lblID string, req *cslb.TCloudRuleCreate,
	tg *corelb.BaseTargetGroup) (*hcproto.BatchCreateResult, error) {

	hcReq := &hcproto.TCloudRuleBatchCreateReq{Rules: []hcproto.TCloudRuleCreate{convRuleCreate(req, tg)}}
	var createResp *hcproto.BatchCreateResult
	var err error
	switch vendor {
	case enumor.TCloud:
		createResp, err = svc.client.HCService().TCloud.Clb.BatchCreateUrlRule(kt, lblID, hcReq)
		if err != nil {
			logs.Errorf("fail to create tcloud url rule, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupport vendor for create rule: %s", vendor)
	}
	if len(createResp.SuccessCloudIDs) == 0 {
		logs.Errorf("no rule have been created, lblID: %s, req: %+v, rid: %s", lblID, hcReq, kt.Rid)
		return nil, errors.New("create failed, reason: unknown")
	}
	return createResp, nil
}

// 构建异步任务将目标组中的RS绑定到对应规则上
func (svc *lbSvc) applyTargetToRule(kt *kit.Kit, tgID, ruleCloudID string, lblInfo *corelb.BaseListener) error {

	// 查找目标组中的rs
	listRsReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.BatchAddRSCloudMaxLimit,
		},
	}
	// Build Task
	tasks := make([]apits.CustomFlowTask, 0)
	getNextID := counter.NewNumStringCounter(1, 10)
	// 判断规则类型
	var ruleType enumor.RuleType
	if lblInfo.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
	} else {
		ruleType = enumor.Layer4RuleType
	}
	// 按目标组数量拆分任务批次
	for {
		rsResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listRsReq)
		if err != nil {
			logs.Errorf("fail to list target, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		if len(rsResp.Details) == 0 {
			break
		}

		rsReq := &hcproto.BatchRegisterTCloudTargetReq{
			CloudListenerID: lblInfo.CloudID,
			CloudRuleID:     ruleCloudID,
			TargetGroupID:   tgID,
			RuleType:        ruleType,
			Targets:         make([]*hcproto.RegisterTarget, 0, len(rsResp.Details)),
		}
		for _, target := range rsResp.Details {
			rsReq.Targets = append(rsReq.Targets, &hcproto.RegisterTarget{
				CloudInstID:      target.CloudInstID,
				TargetType:       target.InstType,
				Port:             target.Port,
				Weight:           target.Weight,
				Zone:             target.Zone,
				InstName:         target.InstName,
				PrivateIPAddress: target.PrivateIPAddress,
				PublicIPAddress:  target.PublicIPAddress,
			})
		}
		tasks = append(tasks, apits.CustomFlowTask{
			ActionID:   action.ActIDType(getNextID()),
			ActionName: enumor.ActionListenerRuleAddTarget,
			Params: actionlb.ListenerRuleAddTargetOption{
				LoadBalancerID:               lblInfo.LbID,
				BatchRegisterTCloudTargetReq: rsReq,
			},
			DependOn: nil,
			Retry:    tableasync.NewRetryWithPolicy(10, 100, 500),
		})

		if len(rsResp.Details) < constant.BatchAddRSCloudMaxLimit {
			break
		}
		listRsReq.Page.Start += constant.BatchAddRSCloudMaxLimit
	}

	if len(tasks) == 0 {
		req := &cloud.TGListenerRelStatusUpdateReq{BindingStatus: enumor.SuccessBindingStatus}
		err := svc.client.DataService().Global.LoadBalancer.BatchUpdateListenerRuleRelStatusByTGID(kt, tgID, req)
		if err != nil {
			logs.Errorf("fail to update listener rule rel status by tgID, err: %v, tgID: %s, rid: %s",
				err, tgID, kt.Rid)
			return err
		}
		return nil
	}
	return svc.createApplyTGFlow(kt, tgID, lblInfo, tasks)
}

// createApplyTGFlow create a custom flow to apply target group to listener rule
func (svc *lbSvc) createApplyTGFlow(kt *kit.Kit, tgID string, lblInfo *corelb.BaseListener,
	tasks []apits.CustomFlowTask) error {

	mainFlowResult, err := svc.client.TaskServer().CreateCustomFlow(kt, &apits.AddCustomFlowReq{
		Name:        enumor.FlowApplyTargetGroupToListenerRule,
		IsInitState: true,
		Tasks:       tasks,
	})
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	flowID := mainFlowResult.ID
	// 创建从任务并加锁
	flowWatchReq := &apits.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []apits.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lblInfo.LbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  []string{tgID},
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ApplyTargetGroupType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call task server to create res flow status watch flow failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return err
	}

	// 锁定负载均衡跟Flow的状态
	err = svc.lockResFlowStatus(kt, lblInfo.LbID, enumor.LoadBalancerCloudResType, flowID, enumor.ApplyTargetGroupType)
	if err != nil {
		logs.Errorf("fail to lock load balancer(%s) for flow(%s), err: %v, rid: %s",
			lblInfo.LbID, flowID, err, kt.Rid)
		return err
	}
	return nil
}

// convRuleCreate convert TCloudRuleCreate to hcproto.TCloudRuleCreate
func convRuleCreate(rule *cslb.TCloudRuleCreate, tg *corelb.BaseTargetGroup) hcproto.TCloudRuleCreate {
	return hcproto.TCloudRuleCreate{
		Url:                rule.Url,
		TargetGroupID:      rule.TargetGroupID,
		CloudTargetGroupID: tg.CloudID,
		Domains:            rule.Domains,
		SessionExpireTime:  rule.SessionExpireTime,
		Scheduler:          rule.Scheduler,
		ForwardType:        rule.ForwardType,
		DefaultServer:      rule.DefaultServer,
		Http2:              rule.Http2,
		TargetType:         rule.TargetType,
		Quic:               rule.Quic,
		TrpcFunc:           rule.TrpcFunc,
		TrpcCallee:         rule.TrpcCallee,
		HealthCheck:        tg.HealthCheck,
		Certificates:       rule.Certificate,
		Memo:               rule.Memo,
	}
}

// targetGroupBindCheck 目标组绑定检查，检查成功返回目标组id为索引的map
func (svc *lbSvc) targetGroupBindCheck(kt *kit.Kit, bizID int64, tgId string) (*corelb.BaseTargetGroup, error) {

	// 检查目标组是否存在
	tgResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bizID),
			tools.RuleEqual("id", tgId),
		),
		Page: core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to query target group(id:%s) info, err: %v, rid: %s", tgId, err, kt.Rid)
		return nil, err
	}

	if len(tgResp.Details) == 0 {
		logs.Errorf("target group can not be found, id: %s, rid: %s", tgId, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "target group(%s) can not be found", tgId)
	}
	tg := &tgResp.Details[0]
	// 检查对应的目标组是否被绑定
	relResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgId),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return nil, err
	}
	if len(relResp.Details) > 0 {
		rel := relResp.Details[0]
		return nil, fmt.Errorf("target group(%s) already been bound to rule or listener(%s)",
			rel.TargetGroupID, rel.CloudListenerRuleID)
	}
	return tg, nil
}

// UpdateBizUrlRule 更新规则
func (svc *lbSvc) UpdateBizUrlRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	ruleID := cts.PathParameter("rule_id").String()
	if len(ruleID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "rule id is required")
	}

	req := new(hcproto.TCloudRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts,
		&handler.ValidWithAuthOption{
			Authorizer: svc.authorizer,
			ResType:    meta.UrlRuleAuditResType,
			Action:     meta.Update,
			BasicInfo:  lblInfo,
		})
	if err != nil {
		return nil, err
	}

	// 更新审计
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert rule update request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.UrlRuleAuditResType, lblInfo.ID, ruleID, updateFields)
	if err != nil {
		logs.Errorf("create update rule audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Clb.UpdateUrlRule(cts.Kit, lblID, ruleID, req)
	default:
		return nil, fmt.Errorf("unsupport vendor for update rule: %s", vendor)
	}
}

// BatchDeleteBizUrlRule 批量删除规则
func (svc *lbSvc) BatchDeleteBizUrlRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(hcproto.TCloudRuleDeleteByIDReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Delete,
		BasicInfo:  lblInfo,
	})
	if err != nil {
		return nil, err
	}

	// 按规则删除审计
	err = svc.audit.ChildResDeleteAudit(cts.Kit, enumor.UrlRuleAuditResType, lblID, req.RuleIDs)
	if err != nil {
		logs.Errorf("create url rule delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	switch vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Clb.BatchDeleteUrlRule(cts.Kit, lblID, req)
	default:
		return nil, fmt.Errorf("unsupport vendor for delete rule: %s", vendor)
	}
}

// BatchDeleteBizUrlRuleByDomain 批量按域名删除规则
func (svc *lbSvc) BatchDeleteBizUrlRuleByDomain(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(hcproto.TCloudRuleDeleteByDomainReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Delete,
		BasicInfo:  lblInfo,
	})
	if err != nil {
		return nil, err
	}

	// 按域名删除审计
	err = svc.audit.ChildResDeleteAudit(cts.Kit, enumor.UrlRuleDomainAuditResType, lblID, req.Domains)
	if err != nil {
		logs.Errorf("create url rule delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Clb.BatchDeleteUrlRuleByDomain(cts.Kit, lblID, req)
	default:
		return nil, fmt.Errorf("unsupport vendor for delete rule: %s", vendor)
	}
}
