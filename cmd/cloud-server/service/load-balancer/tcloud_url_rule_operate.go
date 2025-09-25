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
	"hcm/pkg/tools/slice"
)

// CreateBizUrlRule 业务下新建url规则 TODO: 改成一次只创建一个规则
func (svc *lbSvc) CreateBizUrlRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, err
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is required")
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

	lblInfo, err := svc.validateListenerCertAndAuth(cts, vendor, bizID, lblID, req.Certificate)
	if err != nil {
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

	_, err = svc.applyTargetToRule(cts.Kit, tg.ID, createResp.SuccessCloudIDs[0], lblInfo, bizID)
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
func (svc *lbSvc) applyTargetToRule(kt *kit.Kit, tgID, ruleCloudID string, lblInfo *corelb.BaseListener,
	bkBizID int64) (string, error) {

	lb, err := svc.getLoadBalancerByID(kt, lblInfo.LbID)
	if err != nil {
		logs.Errorf("fail to get load balancer by id, id: %s, err: %v, rid: %s", lblInfo.LbID, err, kt.Rid)
		return "", err
	}
	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, lb.Vendor, lb.AccountID,
		enumor.TaskManagementSourceAPI, enumor.TaskListenerAddTarget)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	var taskDetails []*taskManagementDetail
	defer func() {
		if err == nil {
			return
		}
		// update task management state to failed
		if err := svc.updateTaskManagementState(kt, taskManagementID, enumor.TaskManagementFailed); err != nil {
			logs.Errorf("update task management state to failed failed, err: %v, taskManagementID: %s, rid: %s",
				err, taskManagementID, kt.Rid)
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

	tasks, taskDetails, err := svc.buildRuleAddTargetTasks(kt, tgID, ruleCloudID, taskManagementID, lblInfo, bkBizID)
	if err != nil {
		logs.Errorf("fail to build rule add target tasks, err: %v, tgID: %s, ruleCloudID: %s, taskManagementID: %s, "+
			"lblInfo: %+v, bkBizID: %d, rid: %s", err, tgID, ruleCloudID, taskManagementID, lblInfo, bkBizID, kt.Rid)
		return "", err
	}

	if len(taskDetails) == 0 {
		if err = svc.updateTaskManagementState(kt, taskManagementID, enumor.TaskManagementSuccess); err != nil {
			logs.Errorf("update task management state to success failed, err: %v, taskManagementID: %s, rid: %s",
				err, taskManagementID, kt.Rid)
		}
		req := &cloud.TGListenerRelStatusUpdateReq{BindingStatus: enumor.SuccessBindingStatus}
		err = svc.client.DataService().Global.LoadBalancer.BatchUpdateListenerRuleRelStatusByTGID(kt, tgID, req)
		if err != nil {
			logs.Errorf("fail to update listener rule rel status by tgID, err: %v, tgID: %s, rid: %s",
				err, tgID, kt.Rid)
			return "", err
		}
		return "", nil
	}

	if err := svc.createApplyTGFlow(kt, tgID, taskManagementID, lblInfo, tasks, taskDetails); err != nil {
		logs.Errorf("fail to create apply target group flow, err: %v, tgID: %s, taskManagementID: %s, tasks: %+v, rid: %s",
			err, tgID, taskManagementID, tasks, kt.Rid)
		return "", err
	}
	return taskManagementID, nil
}

func (svc *lbSvc) buildRuleAddTargetTasks(kt *kit.Kit, tgID, ruleCloudID, taskManagementID string,
	lblInfo *corelb.BaseListener, bkBizID int64) ([]apits.CustomFlowTask, []*taskManagementDetail, error) {

	listRsReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchAddRSCloudMaxLimit,
		},
	}
	// 判断规则类型
	var ruleType enumor.RuleType
	if lblInfo.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
	} else {
		ruleType = enumor.Layer4RuleType
	}
	// Build Task
	getNextID := counter.NewNumberCounterWithPrev(1, 10)
	taskDetails := make([]*taskManagementDetail, 0)
	updateTask, err := svc.buildUpdateUrlRuleHealthCheckTask(kt, lblInfo.ID, ruleCloudID, tgID, lblInfo.Vendor, getNextID)
	if err != nil {
		return nil, nil, err
	}
	tasks := []apits.CustomFlowTask{updateTask}
	for {
		rsResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listRsReq)
		if err != nil {
			logs.Errorf("fail to list target, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
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
			rsReq.Targets = append(rsReq.Targets, buildRegisterTarget(target))
		}
		details, err := svc.createListenerAddRsTaskDetails(kt, taskManagementID, bkBizID, rsReq.Targets)
		if err != nil {
			return nil, nil, err
		}

		cur, prev := getNextID()
		actionID := action.ActIDType(cur)
		tmp := apits.CustomFlowTask{
			ActionID:   actionID,
			ActionName: enumor.ActionListenerRuleAddTarget,
			Params: actionlb.ListenerRuleAddTargetOption{
				LoadBalancerID:               lblInfo.LbID,
				BatchRegisterTCloudTargetReq: rsReq,
				ManagementDetailIDs: slice.Map(details, func(item *taskManagementDetail) string {
					return item.taskDetailID
				}),
			},
			Retry: tableasync.NewRetryWithPolicy(10, 100, 500),
		}
		if prev != "" {
			tmp.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		tasks = append(tasks, tmp)
		for _, detail := range details {
			detail.actionID = string(actionID)
		}
		taskDetails = append(taskDetails, details...)
		if len(rsResp.Details) < constant.BatchAddRSCloudMaxLimit {
			break
		}
		listRsReq.Page.Start += constant.BatchAddRSCloudMaxLimit
	}
	// 绑定rs完成后，同步clb
	tasks = append(tasks,
		buildSyncClbFlowTask(lblInfo.Vendor, lblInfo.CloudLbID, lblInfo.AccountID, lblInfo.Region, getNextID))
	return tasks, taskDetails, nil
}

func buildRegisterTarget(target corelb.BaseTarget) *hcproto.RegisterTarget {
	return &hcproto.RegisterTarget{
		CloudInstID:      target.CloudInstID,
		TargetType:       target.InstType,
		Port:             target.Port,
		Weight:           target.Weight,
		Zone:             target.Zone,
		InstName:         target.InstName,
		PrivateIPAddress: target.PrivateIPAddress,
		PublicIPAddress:  target.PublicIPAddress,
	}
}

func (svc *lbSvc) createListenerAddRsTaskDetails(kt *kit.Kit, taskManagementID string, bkBizID int64,
	targets []*hcproto.RegisterTarget) ([]*taskManagementDetail, error) {

	details := make([]*taskManagementDetail, 0)
	for _, param := range targets {
		detail := &taskManagementDetail{
			param: param,
		}
		details = append(details, detail)
	}
	var err error
	details, err = svc.createTaskDetails(kt, taskManagementID, bkBizID,
		enumor.TaskListenerAddTarget, details)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, taskManagementID: %s, bkBizID: %d, rid: %s", err,
			taskManagementID, bkBizID, kt.Rid)
		return nil, err
	}
	return details, nil
}

func (svc *lbSvc) createApplyTGFlow(kt *kit.Kit, tgID, taskManagementID string, lblInfo *corelb.BaseListener,
	tasks []apits.CustomFlowTask, taskDetails []*taskManagementDetail) error {

	flowID, err := svc.buildFlow(kt, enumor.FlowApplyTargetGroupToListenerRule, nil, tasks)
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	for _, detail := range taskDetails {
		detail.flowID = flowID
	}

	if err = svc.updateTaskDetails(kt, taskDetails); err != nil {
		logs.Errorf("update task details failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return err
	}
	if err = svc.updateTaskManagement(kt, taskManagementID, flowID); err != nil {
		logs.Errorf("update task management failed, err: %v, taskManagementID: %s, rid: %s",
			err, taskManagementID, kt.Rid)
		return err
	}

	if err = svc.buildSubFlow(kt, flowID, lblInfo.LbID, []string{tgID}, enumor.TargetGroupCloudResType,
		enumor.ApplyTargetGroupType); err != nil {
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
	if err := vendor.Validate(); err != nil {
		return nil, err
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
	if err := vendor.Validate(); err != nil {
		return nil, err
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
	if err := vendor.Validate(); err != nil {
		return nil, err
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

// BizUrlRuleBindTargetGroup ...
func (svc *lbSvc) BizUrlRuleBindTargetGroup(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, err
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is required")
	}

	req := new(cslb.TCloudRuleBindTargetGroup)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err = svc.targetGroupBindCheck(cts.Kit, bizID, req.TargetGroupID)
	if err != nil {
		logs.Errorf("fail to check target group bind, err: %v, bizID: %d, targetGroupID: %s, rid: %s", err, bizID,
			req.TargetGroupID, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		taskManagementID, err := svc.tcloudUrlBindTargetGroup(cts, bizID, req)
		if err != nil {
			return nil, err
		}
		return taskManagementID, nil
	default:
		return nil, fmt.Errorf("unsupport vendor for bind target group: %s", vendor)
	}
}

func (svc *lbSvc) tcloudUrlBindTargetGroup(cts *rest.Contexts, bizID int64, req *cslb.TCloudRuleBindTargetGroup) (
	string, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", req.UrlRuleID)),
		Page:   core.NewDefaultBasePage(),
	}
	resp, err := svc.listRuleWithCondition(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to list url rule with condition, err: %v, req: %+v, rid: %s", err, listReq, cts.Kit.Rid)
		return "", err
	}
	if len(resp.Details) == 0 {
		logs.Errorf("url rule not found, id: %s, req: %+v, rid: %s", req.UrlRuleID, listReq, cts.Kit.Rid)
		return "", errf.Newf(errf.RecordNotFound, "url rule(%s) not found", req.UrlRuleID)
	}
	rule := resp.Details[0]
	if rule.RuleType != enumor.Layer7RuleType {
		logs.Errorf("url rule is not layer7 rule, id: %s, ruleType: %s, rid: %s", req.UrlRuleID, rule.RuleType, cts.Kit.Rid)
		return "", errf.Newf(errf.InvalidParameter, "url rule(%s) is not layer7 rule", req.UrlRuleID)
	}

	lblInfo, lblBasicInfo, err := svc.getListenerByIDAndBiz(cts.Kit, enumor.TCloud, bizID, rule.LblID)
	if err != nil {
		logs.Errorf("fail to get listener info, bizID: %d, listenerID: %s, err: %v, rid: %s",
			bizID, rule.LblID, err, cts.Kit.Rid)
		return "", err
	}

	// 业务校验、鉴权
	valOpt := &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Create,
		BasicInfo:  lblBasicInfo,
	}
	if err = handler.BizOperateAuth(cts, valOpt); err != nil {
		return "", err
	}

	// 预检测-是否有执行中的负载均衡
	_, err = svc.checkResFlowRel(cts.Kit, lblInfo.LbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return "", err
	}

	taskManagementID, err := svc.applyTargetToRule(cts.Kit, req.TargetGroupID, rule.CloudID, lblInfo, bizID)
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", err
	}

	return taskManagementID, nil
}

// CreateBizUrlRuleWithoutBinding 业务下新建url规则, 区别于 CreateBizUrlRule 的地方在于不绑定目标组
func (svc *lbSvc) CreateBizUrlRuleWithoutBinding(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, err
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is required")
	}

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	req := new(cslb.TCloudRuleCreateWithoutBinding)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.validateListenerCertAndAuth(cts, vendor, bizID, lblID, req.Certificate)
	if err != nil {
		return nil, err
	}

	// 预检测-是否有执行中的负载均衡
	_, err = svc.checkResFlowRel(cts.Kit, lblInfo.LbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	createResp, err := svc.batchCreateUrlRuleWithoutTG(cts.Kit, vendor, lblID, req)
	if err != nil {
		logs.Errorf("fail to create url rule, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return createResp, nil
}

func (svc *lbSvc) validateListenerCertAndAuth(cts *rest.Contexts, vendor enumor.Vendor, bizID int64, lblID string,
	certificate *corelb.TCloudCertificateInfo) (*corelb.BaseListener, error) {

	lblInfo, lblBasicInfo, err := svc.getListenerByIDAndBiz(cts.Kit, vendor, bizID, lblID)
	if err != nil {
		logs.Errorf("fail to get listener info, bizID: %d, listenerID: %s, err: %v, rid: %s",
			bizID, lblID, err, cts.Kit.Rid)
		return nil, err
	}

	// if SNI Switch is off, certificates can only be set in listener not its rule
	if lblInfo.SniSwitch == enumor.SniTypeClose && certificate != nil {
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
	return lblInfo, nil
}

// batchCreateUrlRuleWithoutTG 仅批量创建url规则, 不创建目标组
func (svc *lbSvc) batchCreateUrlRuleWithoutTG(kt *kit.Kit, vendor enumor.Vendor, lblID string,
	req *cslb.TCloudRuleCreateWithoutBinding) (*hcproto.BatchCreateResult, error) {

	ruleCreate := hcproto.TCloudRuleCreate{
		Url:               req.Url,
		Domains:           req.Domains,
		SessionExpireTime: req.SessionExpireTime,
		Scheduler:         req.Scheduler,
		ForwardType:       req.ForwardType,
		DefaultServer:     req.DefaultServer,
		Http2:             req.Http2,
		TargetType:        req.TargetType,
		Quic:              req.Quic,
		TrpcFunc:          req.TrpcFunc,
		TrpcCallee:        req.TrpcCallee,
		Certificates:      req.Certificate,
		Memo:              req.Memo,
	}
	hcReq := &hcproto.TCloudRuleBatchCreateReq{Rules: []hcproto.TCloudRuleCreate{ruleCreate}}
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

func (svc *lbSvc) buildUpdateUrlRuleHealthCheckTask(kt *kit.Kit, lblID, ruleCloudID, tgID string, vendor enumor.Vendor,
	getNextID func() (cur string, prev string)) (apits.CustomFlowTask, error) {

	tg, err := svc.getTargetGroupByID(kt, tgID)
	if err != nil {
		return apits.CustomFlowTask{}, nil
	}

	cur, _ := getNextID()
	actionID := action.ActIDType(cur)
	tmp := apits.CustomFlowTask{
		ActionID:   actionID,
		ActionName: enumor.ActionListenerRuleUpdateHealthCheck,
		Params: actionlb.ListenerRuleUpdateHealthCheckOption{
			ListenerID:  lblID,
			CloudRuleID: ruleCloudID,
			Vendor:      vendor,
			HealthCheck: tg.HealthCheck,
		},
		Retry: tableasync.NewRetryWithPolicy(10, 100, 500),
	}
	return tmp, nil
}
