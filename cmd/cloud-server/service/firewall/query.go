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
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListGcpFirewallRule list gcp firewall rule.
func (svc *firewallSvc) ListGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listGcpFirewallRule(cts, handler.ListResourceAuthRes)
}

// ListBizGcpFirewallRule list biz gcp firewall rule.
func (svc *firewallSvc) ListBizGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listGcpFirewallRule(cts, handler.ListBizAuthRes)
}

func (svc *firewallSvc) listGcpFirewallRule(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(proto.GcpFirewallRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.GcpFirewallRule, Action: meta.Find, Filter: req.Filter})
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
	return svc.getGcpFirewallRule(cts, handler.ResValidWithAuth)
}

// GetBizGcpFirewallRule get biz gcp firewall rule.
func (svc *firewallSvc) GetBizGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	return svc.getGcpFirewallRule(cts, handler.BizValidWithAuth)
}

func (svc *firewallSvc) getGcpFirewallRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	id := cts.PathParameter("id").String()

	basicInfo, err := svc.getBasicInfo(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.GcpFirewallRule,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.GcpFirewallRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
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

func (svc *firewallSvc) listBasicInfo(kt *kit.Kit, ruleIDs []string) (map[string]types.CloudResourceBasicInfo, error) {
	listReq := &dataproto.GcpFirewallRuleListReq{
		Field:  []string{"account_id", "bk_biz_id"},
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, ids: %+v, rid: %s", err, ruleIDs, kt.Rid)
		return nil, err
	}

	basicInfoMap := make(map[string]types.CloudResourceBasicInfo)
	for _, rule := range result.Details {
		basicInfoMap[rule.ID] = types.CloudResourceBasicInfo{AccountID: rule.AccountID, BkBizID: rule.BkBizID}
	}

	return basicInfoMap, nil
}

func (svc *firewallSvc) getBasicInfo(kt *kit.Kit, ruleID string) (*types.CloudResourceBasicInfo, error) {
	listReq := &dataproto.GcpFirewallRuleListReq{
		Field:  []string{"account_id", "bk_biz_id"},
		Filter: tools.EqualExpression("id", ruleID),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Gcp.Firewall.ListFirewallRule(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list firewall rule failed, err: %v, id: %s, rid: %s", err, ruleID, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "gcp firewall rule: %s not found", ruleID)
	}

	rule := result.Details[0]
	return &types.CloudResourceBasicInfo{AccountID: rule.AccountID, BkBizID: rule.BkBizID}, nil
}
