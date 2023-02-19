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

package firewall

import (
	"fmt"
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// InitFirewallService initial the security group service
func InitFirewallService(cap *capability.Capability) {
	svc := &firewallSvc{
		client:     cap.ApiClient,
		authorizer: cap.Authorizer,
		audit:      cap.Audit,
	}

	h := rest.NewHandler()
	h.Add("BatchDeleteGcpFirewallRule", http.MethodDelete, "/vendors/gcp/firewalls/rules/batch",
		svc.BatchDeleteGcpFirewallRule)
	h.Add("UpdateGcpFirewallRule", http.MethodPut, "/vendors/gcp/firewalls/rules/{id}", svc.UpdateGcpFirewallRule)
	h.Add("ListGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/list", svc.ListGcpFirewallRule)
	h.Add("GetGcpFirewallRule", http.MethodGet, "/vendors/gcp/firewalls/rules/{id}", svc.GetGcpFirewallRule)
	h.Add("AssignGcpFirewallRuleToBiz", http.MethodPost, "/vendors/gcp/firewalls/rules/assign/bizs",
		svc.AssignGcpFirewallRuleToBiz)

	h.Load(cap.WebService)
}

type firewallSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// BatchDeleteGcpFirewallRule batch delete gcp firewallSvc rule.
func (svc *firewallSvc) BatchDeleteGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpFirewallRuleBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.GcpFirewallRuleListReq{
		Field:  []string{"account_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule,
			Action: meta.Delete, ResourceID: one.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// 已分配业务的资源，不允许操作
	flt := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err = svc.checkGcpFirewallRulesInBiz(cts.Kit, flt, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.GcpFirewallRuleAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	successIDs := make([]string, 0)
	for _, id := range req.IDs {
		if err := svc.client.HCService().Gcp.Firewall.DeleteFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), id); err != nil {
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		successIDs = append(successIDs, id)
	}

	return nil, nil
}

// UpdateGcpFirewallRule update gcp firewallSvc rule.
func (svc *firewallSvc) UpdateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.GcpFirewallRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID, err := svc.queryAccountID(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule, Action: meta.Update,
		ResourceID: accountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// 已分配业务的资源，不允许操作
	flt := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: id}
	err = svc.checkGcpFirewallRulesInBiz(cts.Kit, flt, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.GcpFirewallRuleAuditResType, id, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.GcpFirewallRuleUpdateReq{
		Memo:              req.Memo,
		Priority:          req.Priority,
		SourceTags:        req.SourceTags,
		TargetTags:        req.TargetTags,
		Denied:            req.Denied,
		Allowed:           req.Allowed,
		SourceRanges:      req.SourceRanges,
		DestinationRanges: req.DestinationRanges,
		Disabled:          req.Disabled,
	}
	err = svc.client.HCService().Gcp.Firewall.UpdateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	if err != nil {
		logs.Errorf("update firewall rule failed, err: %v, id: %s, req: %v, rid: %s", err, id, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *firewallSvc) queryAccountID(kt *kit.Kit, ruleID string) (string, error) {
	listReq := &dataproto.GcpFirewallRuleListReq{
		Field:  []string{"account_id"},
		Filter: tools.EqualExpression("id", ruleID),
		Page:   core.DefaultBasePage,
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, id: %s, rid: %s", err, ruleID, kt.Rid)
		return "", err
	}

	if len(result.Details) == 0 {
		return "", errf.Newf(errf.RecordNotFound, "gcp firewall rule: %s not found", ruleID)
	}

	return result.Details[0].AccountID, nil
}

// ListGcpFirewallRule list gcp firewall rule.
func (svc *firewallSvc) ListGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpFirewallRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	authOpt := &meta.ListAuthResInput{Type: meta.GcpFirewallRule, Action: meta.Find}
	expr, noPermFlag, err := svc.authorizer.ListAuthInstWithFilter(cts.Kit, authOpt, req.Filter, "account_id")
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	listReq := &dataproto.GcpFirewallRuleListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// AssignGcpFirewallRuleToBiz assign gcp firewall rule to biz.
func (svc *firewallSvc) AssignGcpFirewallRuleToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignGcpFirewallRuleToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.assignAuth(cts.Kit, req.FirewallRuleIDs); err != nil {
		return nil, err
	}

	listReq := &dataproto.GcpFirewallRuleListReq{
		Field: []string{"id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: req.FirewallRuleIDs,
				},
				&filter.AtomRule{
					Field: "bk_biz_id",
					Op:    filter.NotEqual.Factory(),
					Value: constant.UnassignedBiz,
				},
			},
		},
		Page: core.DefaultBasePage,
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx,
		cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("ListFirewallRule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) != 0 {
		ids := make([]string, len(result.Details))
		for index, one := range result.Details {
			ids[index] = one.ID
		}
		return nil, fmt.Errorf("gcp firewall rule(ids=%v) already assigned", ids)
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.GcpFirewallRuleAuditResType, req.FirewallRuleIDs, req.BkBizID)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rule := make([]dataproto.GcpFirewallRuleBatchUpdate, 0, len(req.FirewallRuleIDs))
	for _, id := range req.FirewallRuleIDs {
		rule = append(rule, dataproto.GcpFirewallRuleBatchUpdate{
			ID:      id,
			BkBizID: req.BkBizID,
		})
	}
	update := &dataproto.GcpFirewallRuleBatchUpdateReq{
		FirewallRules: rule,
	}
	if err := svc.client.DataService().Gcp.Firewall.BatchUpdateFirewallRule(cts.Kit.Ctx,
		cts.Kit.Header(), update); err != nil {

		logs.Errorf("BatchUpdateFirewallRule failed, err: %v, req: %v, rid: %s", err, update,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *firewallSvc) assignAuth(kt *kit.Kit, rules []string) error {
	listReq := &dataproto.GcpFirewallRuleListReq{
		Field:  []string{"account_id"},
		Filter: tools.ContainersExpression("id", rules),
		Page:   core.DefaultBasePage,
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, ids: %s, rid: %s", err, rules, kt.Rid)
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(result.Details))
	for _, info := range result.Details {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule,
			Action: meta.Assign, ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}

// GetGcpFirewallRule get gcp firewall rule.
func (svc *firewallSvc) GetGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	accountID, err := svc.queryAccountID(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule, Action: meta.Find,
		ResourceID: accountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.GcpFirewallRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.DefaultBasePage,
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "gcp firewall rule: %s not found", id)
	}

	return result.Details[0], nil
}

// checkGcpFirewallRulesInBiz check if gcp firewall rules are in the specified biz.
func (svc *firewallSvc) checkGcpFirewallRulesInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
	req := &dataproto.GcpFirewallRuleListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count firewall rules that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d firewall rules are already assigned", result.Count)
	}

	return nil
}
