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

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {
	svc := &lbSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// clb apis in res
	h.Add("ListLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListLoadBalancer)
	h.Add("ListLoadBalancerWithDeleteProtection", http.MethodPost,
		"/load_balancers/with/delete_protection/list", svc.ListLoadBalancerWithDeleteProtect)
	h.Add("BatchCreateLB", http.MethodPost, "/load_balancers/create", svc.BatchCreateLB)
	h.Add("InquiryPriceLoadBalancer", http.MethodPost, "/load_balancers/prices/inquiry", svc.InquiryPriceLoadBalancer)
	h.Add("AssignLbToBiz", http.MethodPost, "/load_balancers/assign/bizs", svc.AssignLbToBiz)
	h.Add("GetLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetLoadBalancer)
	h.Add("TCloudDescribeResources", http.MethodPost,
		"/vendors/tcloud/load_balancers/resources/describe", svc.TCloudDescribeResources)
	h.Add("BatchDeleteLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteLoadBalancer)
	h.Add("ListListenerCountByLbIDs", http.MethodPost, "/load_balancers/listeners/count", svc.ListListenerCountByLbIDs)
	h.Add("GetLoadBalancerLockStatus", http.MethodGet,
		"/load_balancers/{id}/lock/status", svc.GetLoadBalancerLockStatus)
	h.Add("ListResLoadBalancerQuotas", http.MethodPost, "/load_balancers/quotas", svc.ListResLoadBalancerQuotas)

	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}")
	bizService(bizH, svc)
	bizURLRuleService(bizH, svc)
	bizSopService(bizH, svc)

	h.Load(c.WebService)
	bizH.Load(c.WebService)
}

func bizService(h *rest.Handler, svc *lbSvc) {
	// h.Add("BizBatchCreateLB", http.MethodPost, "/load_balancers/create", svc.BizBatchCreateLB)
	h.Add("UpdateBizTCloudLoadBalancer", http.MethodPatch,
		"/vendors/tcloud/load_balancers/{id}", svc.UpdateBizTCloudLoadBalancer)
	h.Add("ListBizLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListBizLoadBalancer)
	h.Add("ListLoadBalancerWithDeleteProtection", http.MethodPost,
		"/load_balancers/with/delete_protection/list", svc.ListBizLoadBalancerWithDeleteProtect)
	h.Add("GetBizLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetBizLoadBalancer)
	h.Add("BatchDeleteBizLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteBizLoadBalancer)

	h.Add("ListBizListener", http.MethodPost, "/load_balancers/{lb_id}/listeners/list", svc.ListBizListener)
	h.Add("GetBizListener", http.MethodGet, "/listeners/{id}", svc.GetBizListener)
	h.Add("ListBizListenerDomains", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/domains/list", svc.ListBizListenerDomains)
	h.Add("ListBizListenerCountByLbIDs", http.MethodPost, "/load_balancers/listeners/count",
		svc.ListBizListenerCountByLbIDs)
	h.Add("GetBizLoadBalancerLockStatus", http.MethodGet,
		"/load_balancers/{id}/lock/status", svc.GetBizLoadBalancerLockStatus)
	h.Add("ListBizLoadBalancerQuotas", http.MethodPost, "/load_balancers/quotas", svc.ListBizLoadBalancerQuotas)

	h.Add("TCloudCreateSnatIps", http.MethodPost,
		"/vendors/tcloud/load_balancers/{lb_id}/snat_ips/create", svc.TCloudCreateSnatIps)
	h.Add("TCloudDeleteSnatIps", http.MethodDelete,
		"/vendors/tcloud/load_balancers/{lb_id}/snat_ips", svc.TCloudDeleteSnatIps)

	// 目标组
	h.Add("ListBizTargetsByTGID", http.MethodPost,
		"/target_groups/{target_group_id}/targets/list", svc.ListBizTargetsByTGID)

	h.Add("StatBizTargetWeight", http.MethodPost,
		"/target_groups/targets/weight_stat", svc.StatBizTargetWeight)
	h.Add("AssociateBizTargetGroupListenerRel", http.MethodPost,
		"/listeners/associate/target_group", svc.AssociateBizTargetGroupListenerRel)

	h.Add("CreateBizTargetGroup", http.MethodPost, "/target_groups/create", svc.CreateBizTargetGroup)
	h.Add("UpdateBizTargetGroup", http.MethodPatch, "/target_groups/{id}", svc.UpdateBizTargetGroup)
	h.Add("UpdateBizTargetGroupHealth", http.MethodPatch,
		"/target_groups/{id}/health_check", svc.UpdateBizTargetGroupHealth)
	h.Add("DeleteBizTargetGroup", http.MethodDelete, "/target_groups/batch", svc.DeleteBizTargetGroup)
	h.Add("ListBizTargetGroup", http.MethodPost, "/target_groups/list", svc.ListBizTargetGroup)
	h.Add("GetBizTargetGroup", http.MethodGet, "/target_groups/{id}", svc.GetBizTargetGroup)
	// 与异步任务相关的操作
	h.Add("BatchAddBizTargets", http.MethodPost, "/target_groups/targets/create", svc.BatchAddBizTargets)
	h.Add("BatchRemoveBizTargets", http.MethodDelete, "/target_groups/targets/batch", svc.BatchRemoveBizTargets)
	h.Add("BatchModifyBizTargetPort",
		http.MethodPatch, "/target_groups/{target_group_id}/targets/port", svc.BatchModifyBizTargetsPort)
	h.Add("BatchModifyBizTargetsWeight", http.MethodPatch,
		"/target_groups/{target_group_id}/targets/weight", svc.BatchModifyBizTargetsWeight)
	h.Add("BatchDeleteBizRule", http.MethodDelete, "/rule/batch", svc.BatchDeleteBizRule)

	h.Add("CancelFlow", http.MethodPost, "/load_balancers/{lb_id}/async_flows/terminate", svc.BizTerminateFlow)
	h.Add("RetryTask", http.MethodPost, "/load_balancers/{lb_id}/async_tasks/retry", svc.BizRetryTask)
	h.Add("CloneFlow", http.MethodPost, "/load_balancers/{lb_id}/async_flows/clone", svc.BizCloneFlow)
	h.Add("GetResultAfterTerminate", http.MethodPost,
		"/load_balancers/{lb_id}/async_flows/result_after_terminate", svc.BizGetResultAfterTerminate)

	h.Add("ListBizTargetsHealthByTGID", http.MethodPost,
		"/target_groups/{target_group_id}/targets/health", svc.ListBizTargetsHealthByTGID)

	// 监听器
	h.Add("CreateBizListener", http.MethodPost, "/load_balancers/{lb_id}/listeners/create", svc.CreateBizListener)
	h.Add("UpdateBizListener", http.MethodPatch, "/listeners/{id}", svc.UpdateBizListener)
	h.Add("DeleteBizListener", http.MethodDelete, "/listeners/batch", svc.DeleteBizListener)
	h.Add("UpdateBizDomainAttr", http.MethodPatch, "/listeners/{lbl_id}/domains", svc.UpdateBizDomainAttr)

}

func bizURLRuleService(h *rest.Handler, svc *lbSvc) {
	// 规则
	h.Add("GetBizTCloudUrlRule", http.MethodGet,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.GetBizTCloudUrlRule)
	h.Add("ListBizUrlRulesByListener", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/list", svc.ListBizUrlRulesByListener)
	h.Add("ListBizTCloudRuleByTG", http.MethodPost,
		"/vendors/tcloud/target_groups/{target_group_id}/rules/list", svc.ListBizTCloudRuleByTG)
	h.Add("CreateBizTCloudUrlRule", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/create", svc.CreateBizTCloudUrlRule)
	h.Add("UpdateBizTCloudUrlRule", http.MethodPatch,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.UpdateBizTCloudUrlRule)
	h.Add("BatchDeleteBizTCloudUrlRule", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/batch", svc.BatchDeleteBizTCloudUrlRule)
	h.Add("BatchDeleteBizTCloudUrlRuleByDomain", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/by/domains/batch", svc.BatchDeleteBizTCloudUrlRuleByDomain)
}

func bizSopService(h *rest.Handler, svc *lbSvc) {
	// 标准运维
	h.Add("BatchBizAddTargetGroupRS", http.MethodPost,
		"/sops/target_groups/targets/create", svc.BatchBizAddTargetGroupRS)
	h.Add("BatchBizRemoveTargetGroupRS", http.MethodDelete,
		"/sops/target_groups/targets/batch", svc.BatchBizRemoveTargetGroupRS)
	h.Add("BatchBizModifyWeightTargetGroup", http.MethodPatch,
		"/sops/target_groups/targets/weight", svc.BatchBizModifyWeightTargetGroup)
	h.Add("BatchBizRuleOnline", http.MethodPost,
		"/sops/rule/online", svc.BatchBizRuleOnline)
	h.Add("BatchBizRuleOffline", http.MethodDelete,
		"/sops/rule/offline", svc.BatchBizRuleOffline)
}

type lbSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disk.Interface
	cvmLgc     cvm.Interface
	eipLgc     eip.Interface
}
