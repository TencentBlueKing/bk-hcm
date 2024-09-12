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

// Package loadbalancer 负载均衡的DB接口
package loadbalancer

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

var svc *lbSvc

// InitService initial the clb service
func InitService(cap *capability.Capability) {
	svc = &lbSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	// 负载均衡
	h.Add("GetLoadBalancer", http.MethodGet, "/vendors/{vendor}/load_balancers/{id}", svc.GetLoadBalancer)
	h.Add("ListLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListLoadBalancer)
	h.Add("ListLoadBalancerRaw", http.MethodPost, "/load_balancers/list_with_extension", svc.ListLoadBalancerRaw)
	h.Add("ListLoadBalancerExt", http.MethodPost, "/vendors/{vendor}/load_balancers/list", svc.ListLoadBalancerExt)
	h.Add("BatchCreateLoadBalancer", http.MethodPost, "/vendors/{vendor}/load_balancers/batch/create",
		svc.BatchCreateLoadBalancer)
	h.Add("BatchUpdateLoadBalancer",
		http.MethodPatch, "/vendors/{vendor}/load_balancers/batch/update", svc.BatchUpdateLoadBalancer)
	h.Add("BatchUpdateLbBizInfo", http.MethodPatch, "/load_balancers/bizs/batch/update", svc.BatchUpdateLbBizInfo)
	h.Add("BatchDeleteLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteLoadBalancer)

	// 监听器
	h.Add("GetListener", http.MethodGet, "/vendors/{vendor}/listeners/{id}", svc.GetListener)
	h.Add("ListListener", http.MethodPost, "/load_balancers/listeners/list", svc.ListListener)
	h.Add("ListListenerExt", http.MethodPost, "/vendors/tcloud/load_balancers/listeners/list", svc.ListListenerExt)
	h.Add("BatchCreateListener", http.MethodPost, "/vendors/{vendor}/listeners/batch/create", svc.BatchCreateListener)
	h.Add("BatchCreateListenerWithRule", http.MethodPost, "/vendors/{vendor}/listeners/rules/batch/create",
		svc.BatchCreateListenerWithRule)
	h.Add("BatchUpdateListener", http.MethodPatch, "/vendors/{vendor}/listeners/batch/update", svc.BatchUpdateListener)
	h.Add("BatchDeleteListener", http.MethodDelete, "/listeners/batch", svc.BatchDeleteListener)
	h.Add("CountListenerByLbIDs", http.MethodPost, "/load_balancers/listeners/count", svc.CountListenerByLbIDs)
	h.Add("BatchUpdateListenerBizInfo", http.MethodPatch,
		"/load_balancers/listeners/bizs/batch/update", svc.BatchUpdateListenerBizInfo)
	h.Add("ListListenerWithTargets", http.MethodPost, "/load_balancers/listeners/with/targets/list",
		svc.ListListenerWithTargets)
	h.Add("ListBatchListeners", http.MethodPost, "/load_balancers/listeners/batch/list", svc.ListBatchListeners)

	// url规则
	h.Add("BatchCreateTCloudUrlRule",
		http.MethodPost, "/vendors/tcloud/url_rules/batch/create", svc.BatchCreateTCloudUrlRule)
	h.Add("BatchUpdateTCloudUrlRule",
		http.MethodPatch, "/vendors/tcloud/url_rules/batch/update", svc.BatchUpdateTCloudUrlRule)
	h.Add("BatchDeleteTCloudUrlRule",
		http.MethodDelete, "/vendors/tcloud/url_rules/batch", svc.BatchDeleteTCloudUrlRule)
	h.Add("ListTCloudUrlRule", http.MethodPost, "/vendors/tcloud/load_balancers/url_rules/list", svc.ListTCloudUrlRule)

	// 目标组
	h.Add("BatchCreateTargetGroup", http.MethodPost,
		"/vendors/{vendor}/target_groups/batch/create", svc.BatchCreateTargetGroup)
	h.Add("BatchCreateTargetGroupWithRel", http.MethodPost,
		"/vendors/{vendor}/target_groups/with/rels/batch/create", svc.BatchCreateTargetGroupWithRel)
	h.Add("GetTargetGroup", http.MethodGet, "/vendors/{vendor}/target_groups/{id}", svc.GetTargetGroup)
	h.Add("ListTargetGroup", http.MethodPost, "/load_balancers/target_groups/list", svc.ListTargetGroup)
	h.Add("UpdateTargetGroup", http.MethodPatch, "/vendors/{vendor}/target_groups", svc.UpdateTargetGroup)
	h.Add("BatchDeleteTargetGroup", http.MethodDelete, "/target_groups/batch", svc.BatchDeleteTargetGroup)
	h.Add("BatchUpdateListenerBizInfo", http.MethodPatch,
		"/load_balancers/target_groups/bizs/batch/update", svc.BatchUpdateTargetGroupBizInfo)
	// RS
	h.Add("BatchDeleteTarget", http.MethodDelete, "/load_balancers/targets/batch", svc.BatchDeleteTarget)
	h.Add("BatchUpdateTarget", http.MethodPatch, "/load_balancers/targets/batch/update", svc.BatchUpdateTarget)
	h.Add("ListTarget", http.MethodPost, "/load_balancers/targets/list", svc.ListTarget)
	h.Add("BatchCreateTarget", http.MethodPost, "/targets/batch/create", svc.BatchCreateTarget)

	// 目标组 规则关联关系
	h.Add("CreateTargetGroupListenerRel", http.MethodPost,
		"/target_group_listener_rels/create", svc.CreateTargetGroupListenerRel)
	h.Add("ListTargetGroupListenerRel", http.MethodPost,
		"/target_group_listener_rels/list", svc.ListTargetGroupListenerRel)
	h.Add("BatchUpdateListenerRuleRelStatusByTGID", http.MethodPatch,
		"/target_group_listener_rels/target_groups/{tg_id}/update", svc.BatchUpdateListenerRuleRelStatusByTGID)

	// 资源与Flow相关的接口
	resFlowRel(h)

	h.Load(cap.WebService)
}

// resFlowRel 资源与Flow相关的接口
func resFlowRel(h *rest.Handler) {
	// 资源跟Flow锁定
	h.Add("CreateResFlowLock", http.MethodPost, "/res_flow_locks/create", svc.CreateResFlowLock)
	h.Add("DeleteResFlowLock", http.MethodDelete, "/res_flow_locks/batch", svc.DeleteResFlowLock)
	h.Add("ListResFlowLock", http.MethodPost, "/res_flow_locks/list", svc.ListResFlowLock)
	h.Add("ResFlowLock", http.MethodPost, "/res_flow_locks/lock", svc.ResFlowLock)
	h.Add("ResFlowUnLock", http.MethodPost, "/res_flow_locks/unlock", svc.ResFlowUnLock)

	// 资源跟Flow关联关系
	h.Add("BatchCreateResFlowRel", http.MethodPost, "/res_flow_rels/batch/create", svc.BatchCreateResFlowRel)
	h.Add("BatchUpdateResFlowRel", http.MethodPatch, "/res_flow_rels/batch/update", svc.BatchUpdateResFlowRel)
	h.Add("BatchDeleteResFlowRel", http.MethodDelete, "/res_flow_rels/batch", svc.BatchDeleteResFlowRel)
	h.Add("ListResFlowRel", http.MethodPost, "/res_flow_rels/list", svc.ListResFlowRel)
}

type lbSvc struct {
	dao dao.Set
}
