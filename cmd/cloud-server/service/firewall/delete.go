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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchDeleteGcpFirewallRule batch delete gcp firewall rule.
func (svc *firewallSvc) BatchDeleteGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteGcpFirewallRule(cts, handler.ResValidWithAuth)
}

// BatchDeleteBizGcpFirewallRule batch delete biz gcp firewall rule.
func (svc *firewallSvc) BatchDeleteBizGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteGcpFirewallRule(cts, handler.BizValidWithAuth)
}

func (svc *firewallSvc) batchDeleteGcpFirewallRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.GcpFirewallRuleBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoMap, err := svc.listBasicInfo(cts.Kit, req.IDs)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.GcpFirewallRule,
		Action: meta.Delete, BasicInfos: basicInfoMap})
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
