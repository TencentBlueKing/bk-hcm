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

// Package audit ...
package audit

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	coreaudit "hcm/pkg/api/core/audit"
	"hcm/pkg/api/data-service/audit"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// InitService initialize the audit service.
func InitService(c *capability.Capability) {
	svc := &svc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("GetAudit", http.MethodGet, "/audits/{id}", svc.GetAudit)
	h.Add("ListAudit", http.MethodPost, "/audits/list", svc.ListAudit)
	h.Add("ListAuditAsyncFlow", http.MethodPost, "/audits/async_flow/list", svc.ListAuditAsyncFlow)
	h.Add("ListAuditAsyncTask", http.MethodPost, "/audits/async_task/list", svc.ListAuditAsyncTask)

	// biz audit apis
	h.Add("GetBizAudit", http.MethodGet, "/bizs/{bk_biz_id}/audits/{id}", svc.GetBizAudit)
	h.Add("ListBizAudit", http.MethodPost, "/bizs/{bk_biz_id}/audits/list", svc.ListBizAudit)
	h.Add("ListBizAuditAsyncFlow", http.MethodPost, "/bizs/{bk_biz_id}/audits/async_flow/list",
		svc.ListBizAuditAsyncFlow)
	h.Add("ListBizAuditAsyncTask", http.MethodPost, "/bizs/{bk_biz_id}/audits/async_task/list",
		svc.ListBizAuditAsyncTask)

	h.Load(c.WebService)
}

type svc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// GetAudit get audit.
func (svc svc) GetAudit(cts *rest.Contexts) (interface{}, error) {
	return svc.getAudit(cts, handler.ResOperateAuth)
}

// GetBizAudit get biz audit.
func (svc svc) GetBizAudit(cts *rest.Contexts) (interface{}, error) {
	return svc.getAudit(cts, handler.BizOperateAuth)
}

func (svc svc) getAudit(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id, err := cts.PathParameter("id").Uint64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rawAudit, err := svc.client.DataService().Global.Audit.GetAuditRaw(cts.Kit, id)
	if err != nil {
		logs.Errorf("get audit failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Audit,
		Action:    meta.Find,
		BasicInfo: &types.CloudResourceBasicInfo{BkBizID: rawAudit.BkBizID, AccountID: rawAudit.AccountID}})
	if err != nil {
		return nil, err
	}

	return rawAudit, nil
}

// ListAudit list audit.
func (svc svc) ListAudit(cts *rest.Contexts) (interface{}, error) {
	return svc.listAudit(cts, handler.ListResourceAuthRes)
}

// ListBizAudit list biz audit.
func (svc svc) ListBizAudit(cts *rest.Contexts) (interface{}, error) {
	return svc.listAudit(cts, handler.ListBizAuthRes)
}

func (svc svc) listAudit(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.AuditListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Audit, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &audit.ListResult{Count: 0, Details: make([]coreaudit.Audit, 0)}, nil
	}
	req.Filter = expr

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.Audit.ListAudit(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}

// ListAuditAsyncFlow 查询资源下异步任务的操作记录详情.
func (svc svc) ListAuditAsyncFlow(cts *rest.Contexts) (interface{}, error) {
	return svc.listAuditAsyncFlow(cts, handler.ListResourceAuthRes)
}

// ListBizAuditAsyncFlow 查询业务下异步任务的操作记录详情.
func (svc svc) ListBizAuditAsyncFlow(cts *rest.Contexts) (interface{}, error) {
	return svc.listAuditAsyncFlow(cts, handler.ListBizAuthRes)
}

func (svc svc) listAuditAsyncFlow(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(proto.AuditAsyncFlowListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	idFilter := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{&filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: req.AuditID}},
	}
	// authorize
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Audit, Action: meta.Find, Filter: idFilter})
	if err != nil {
		return nil, err
	}

	result := &audit.GetAsyncTaskResp{}
	if noPermFlag {
		return result, nil
	}

	// 获取操作记录详情
	auditInfo, err := svc.client.DataService().Global.Audit.GetAudit(cts.Kit.Ctx, cts.Kit.Header(), req.AuditID)
	if err != nil {
		logs.Errorf("get audit by id failed, err: %v, id: %d, req: %+v, rid: %s", err, req.AuditID, req, cts.Kit.Rid)
		return nil, err
	}
	if auditInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "audit: %d not found", req.AuditID)
	}

	// 获取异步任务-Flow详情
	flow, err := svc.client.TaskServer().GetFlow(cts.Kit, req.FlowID)
	if err != nil {
		logs.Errorf("get flow by id failed, err: %v, auditID: %d, flowID: %s, rid: %s",
			err, req.AuditID, req.FlowID, cts.Kit.Rid)
		return nil, err
	}
	result.Flow = flow

	// 获取异步任务-子任务列表
	taskReq := &core.ListReq{
		Fields: []string{"id", "flow_id", "action_id", "action_name", "state", "reason", "created_at", "updated_at"},
		Filter: tools.EqualExpression("flow_id", req.FlowID),
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "action_id",
			Order: core.Ascending,
		},
	}
	taskList, err := svc.client.TaskServer().ListTask(cts.Kit, taskReq)
	if err != nil {
		logs.Errorf("get flow by id failed, flowID: %s, err: %v, rid: %s", req.FlowID, err, cts.Kit.Rid)
		return nil, err
	}
	result.Tasks = taskList.Details

	return result, nil
}

// ListAuditAsyncTask 查询资源下异步任务的操作记录指定子任务的详情.
func (svc svc) ListAuditAsyncTask(cts *rest.Contexts) (interface{}, error) {
	return svc.listAuditAsyncTask(cts, handler.ListResourceAuthRes)
}

// ListBizAuditAsyncTask 查询业务下异步任务的操作记录指定子任务的详情.
func (svc svc) ListBizAuditAsyncTask(cts *rest.Contexts) (interface{}, error) {
	return svc.listAuditAsyncTask(cts, handler.ListBizAuthRes)
}

func (svc svc) listAuditAsyncTask(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(proto.AuditAsyncTaskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	idFilter := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{&filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: req.AuditID}},
	}
	// authorize
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Audit, Action: meta.Find, Filter: idFilter})
	if err != nil {
		return nil, err
	}

	result := &audit.GetAsyncTaskResp{}
	if noPermFlag {
		return result, nil
	}

	// 获取操作记录详情
	auditInfo, err := svc.client.DataService().Global.Audit.GetAudit(cts.Kit.Ctx, cts.Kit.Header(), req.AuditID)
	if err != nil {
		logs.Errorf("get audit by id failed, id: %d, err: %v, rid: %s", req.AuditID, err, cts.Kit.Rid)
		return nil, err
	}
	if auditInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "audit: %d not found", req.AuditID)
	}

	// 获取异步任务-Flow详情
	flow, err := svc.client.TaskServer().GetFlow(cts.Kit, req.FlowID)
	if err != nil {
		logs.Errorf("get flow by id failed, flowID: %s, err: %v, rid: %s", req.FlowID, err, cts.Kit.Rid)
		return nil, err
	}
	result.Flow = flow

	// 获取异步任务-子任务列表
	taskReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("flow_id", req.FlowID),
			tools.RuleEqual("action_id", req.ActionID),
		),
		Page: core.NewDefaultBasePage(),
	}
	taskList, err := svc.client.TaskServer().ListTask(cts.Kit, taskReq)
	if err != nil {
		logs.Errorf("get flow by id failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}
	result.Tasks = taskList.Details

	return result, nil
}
