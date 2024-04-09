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

// Package loadbalancer ...
package loadbalancer

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/cvm"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initialize the clb service.
func InitService(c *capability.Capability) {
	svc := &lbSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}")
	// clb apis in biz
	h.Add("ListLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListLoadBalancer)
	h.Add("BatchCreateLB", http.MethodPost, "/load_balancers/create", svc.BatchCreateLB)
	h.Add("AssignLbToBiz", http.MethodPost, "/load_balancers/assign/bizs", svc.AssignLbToBiz)
	h.Add("GetLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetLoadBalancer)
	h.Add("TCloudDescribeResources", http.MethodPost, "/vendors/tcloud/load_balancers/resources/describe",
		svc.TCloudDescribeResources)
	h.Add("BatchDeleteBizLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteLoadBalancer)

	bizH.Add("UpdateBizTCloudLoadBalancer", http.MethodPatch, "/vendors/tcloud/load_balancers/{id}",
		svc.UpdateBizTCloudLoadBalancer)
	bizH.Add("ListBizLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListBizLoadBalancer)
	bizH.Add("GetBizLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetBizLoadBalancer)
	bizH.Add("BatchDeleteBizLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteBizLoadBalancer)

	bizH.Add("ListBizListener", http.MethodPost, "/load_balancers/{lb_id}/listeners/list", svc.ListBizListener)
	bizH.Add("GetBizListener", http.MethodGet, "/listeners/{id}", svc.GetBizListener)
	bizH.Add("ListBizListenerDomains", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/domains/list", svc.ListBizListenerDomains)

	bizH.Add("ListBizTargetsByTGID", http.MethodPost,
		"/target_groups/{target_group_id}/targets/list", svc.ListBizTargetsByTGID)
	bizH.Add("AssociateBizTargetGroupListenerRel", http.MethodPost, "/listeners/associate/target_group",
		svc.AssociateBizTargetGroupListenerRel)

	bizH.Add("CreateBizTargetGroup", http.MethodPost, "/target_groups/create", svc.CreateBizTargetGroup)
	bizH.Add("UpdateBizTargetGroup", http.MethodPatch, "/target_groups/{id}", svc.UpdateBizTargetGroup)
	bizH.Add("DeleteBizTargetGroup", http.MethodDelete, "/target_groups/batch", svc.DeleteBizTargetGroup)
	bizH.Add("ListBizTargetGroup", http.MethodPost, "/target_groups/list", svc.ListBizTargetGroup)
	bizH.Add("GetBizTargetGroup", http.MethodGet, "/target_groups/{id}", svc.GetBizTargetGroup)
	bizH.Add("BatchAddBizTargets", http.MethodPost, "/target_groups/{target_group_id}/targets/create",
		svc.BatchAddBizTargets)
	bizH.Add("BatchRemoveBizTargets", http.MethodDelete, "/target_groups/{target_group_id}/targets/batch",
		svc.BatchRemoveBizTargets)
	bizH.Add("BatchModifyBizTargetPort", http.MethodPatch, "/target_groups/{target_group_id}/targets/port",
		svc.BatchModifyBizTargetsPort)
	bizH.Add("BatchModifyBizTargetsWeight", http.MethodPatch, "/target_groups/{target_group_id}/targets/weight",
		svc.BatchModifyBizTargetsWeight)

	// 监听器
	bizH.Add("CreateBizListener", http.MethodPost, "/load_balancers/{lb_id}/listeners/create", svc.CreateBizListener)
	bizH.Add("UpdateBizListener", http.MethodPatch, "/listeners/{id}", svc.UpdateBizListener)
	bizH.Add("DeleteBizListener", http.MethodDelete, "/listeners/batch", svc.DeleteBizListener)
	bizH.Add("UpdateBizDomainAttr", http.MethodPatch, "/listeners/{lbl_id}/domains", svc.UpdateBizDomainAttr)

	// 规则
	bizH.Add("GetBizTCloudUrlRule", http.MethodGet,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.GetBizTCloudUrlRule)
	bizH.Add("ListBizUrlRulesByListener", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/list", svc.ListBizUrlRulesByListener)
	bizH.Add("ListBizTCloudRuleByTG", http.MethodPost,
		"/vendors/tcloud/target_groups/{target_group_id}/rules/list", svc.ListBizTCloudRuleByTG)
	bizH.Add("CreateBizTCloudUrlRule", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/create", svc.CreateBizTCloudUrlRule)
	bizH.Add("UpdateBizTCloudUrlRule", http.MethodPatch,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.UpdateBizTCloudUrlRule)
	bizH.Add("BatchDeleteBizTCloudUrlRule", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/batch", svc.BatchDeleteBizTCloudUrlRule)
	bizH.Add("BatchDeleteBizTCloudUrlRuleByDomain", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/by/domains/batch", svc.BatchDeleteBizTCloudUrlRuleByDomain)

	h.Load(c.WebService)
	bizH.Load(c.WebService)
}

type lbSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disk.Interface
	cvmLgc     cvm.Interface
	eipLgc     eip.Interface
}
