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

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

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
