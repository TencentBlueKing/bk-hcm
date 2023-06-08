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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

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
		Page: core.NewDefaultBasePage(),
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
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, ids: %s, rid: %s", err, rules, kt.Rid)
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(result.Details))
	for _, info := range result.Details {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule,
			Action: meta.Assign, ResourceID: info.AccountID}, BizID: info.BkBizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}
