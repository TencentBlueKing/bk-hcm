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
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitFirewallService initial the security group service
func InitFirewallService(cap *capability.Capability) {
	svc := &firewallSvc{
		client:     cap.ApiClient,
		authorizer: cap.Authorizer,
		audit:      cap.Audit,
	}

	h := rest.NewHandler()
	// 资源下相关接口
	h.Add("CreateGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/create",
		svc.CreateGcpFirewallRule)
	h.Add("BatchDeleteGcpFirewallRule", http.MethodDelete, "/vendors/gcp/firewalls/rules/batch",
		svc.BatchDeleteGcpFirewallRule)
	h.Add("UpdateGcpFirewallRule", http.MethodPut, "/vendors/gcp/firewalls/rules/{id}", svc.UpdateGcpFirewallRule)
	h.Add("ListGcpFirewallRule", http.MethodPost, "/vendors/gcp/firewalls/rules/list", svc.ListGcpFirewallRule)
	h.Add("GetGcpFirewallRule", http.MethodGet, "/vendors/gcp/firewalls/rules/{id}", svc.GetGcpFirewallRule)
	h.Add("AssignGcpFirewallRuleToBiz", http.MethodPost, "/vendors/gcp/firewalls/rules/assign/bizs",
		svc.AssignGcpFirewallRuleToBiz)

	// 业务下相关接口
	h.Add("CreateBizGcpFirewallRule", http.MethodPost, "/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/create",
		svc.CreateBizGcpFirewallRule)
	h.Add("BatchDeleteBizGcpFirewallRule", http.MethodDelete, "/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/batch",
		svc.BatchDeleteBizGcpFirewallRule)
	h.Add("UpdateBizGcpFirewallRule", http.MethodPut, "/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/{id}",
		svc.UpdateBizGcpFirewallRule)
	h.Add("ListBizGcpFirewallRule", http.MethodPost, "/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/list",
		svc.ListBizGcpFirewallRule)
	h.Add("GetBizGcpFirewallRule", http.MethodGet, "/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/{id}",
		svc.GetBizGcpFirewallRule)

	h.Load(cap.WebService)
}

type firewallSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
