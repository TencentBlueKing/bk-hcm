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

	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// InitFirewallService initial the security group service
func InitFirewallService(cap *capability.Capability) {
	svc := &firewallSvc{
		client:     cap.ApiClient,
		authorizer: cap.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("DeleteGcpFirewallRule", http.MethodDelete, "/vendors/gcp/firewalls/rules/{id}", svc.DeleteGcpFirewallRule)
	h.Add("UpdateGcpFirewallRule", http.MethodPut, "/vendors/gcp/firewalls/rules/{id}", svc.UpdateGcpFirewallRule)
	h.Add("ListGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/list", svc.ListGcpFirewallRule)
	h.Add("AssignGcpFirewallRuleToBiz", http.MethodPost, "/vendors/gcp/firewalls/rules/assign/bizs",
		svc.AssignGcpFirewallRuleToBiz)

	h.Load(cap.WebService)
}

type firewallSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// DeleteGcpFirewallRule delete gcp firewallSvc rule.
func (svc *firewallSvc) DeleteGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	if err := svc.client.HCService().Gcp.Firewall.DeleteFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), id); err != nil {
		logs.Errorf("delete firewall rule failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
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
	err := svc.client.HCService().Gcp.Firewall.UpdateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	if err != nil {
		logs.Errorf("update firewall rule failed, err: %v, id: %s, req: %v, rid: %s", err, id, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
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
		Page: types.DefaultBasePage,
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
