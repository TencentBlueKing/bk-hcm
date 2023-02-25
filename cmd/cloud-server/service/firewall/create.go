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
	"strconv"

	proto "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateGcpFirewallRule ...
func (svc *firewallSvc) CreateGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(proto.GcpFirewallRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GcpFirewallRule, Action: meta.Create,
		ResourceID: strconv.FormatInt(bizID, 10)}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	createReq := &hcproto.GcpFirewallRuleCreateReq{
		BkBizID:           bizID,
		AccountID:         req.AccountID,
		CloudVpcID:        req.CloudVpcID,
		Name:              req.Name,
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
	result, err := svc.client.HCService().Gcp.Firewall.CreateFirewallRule(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("create gcp firewall rule failed, err: %v, req: %+v, rid: %s", err, createReq, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
