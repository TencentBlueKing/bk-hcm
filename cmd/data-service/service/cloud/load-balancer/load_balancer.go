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

	h.Add("GetLoadBalancer", http.MethodGet, "/vendors/{vendor}/load_balancers/{id}", svc.GetLoadBalancer)
	h.Add("ListLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListLoadBalancer)
	h.Add("ListLoadBalancerExt", http.MethodPost, "/vendors/{vendor}/load_balancers/list", svc.ListLoadBalancerExt)
	h.Add("BatchCreateLoadBalancer", http.MethodPost, "/vendors/{vendor}/load_balancers/batch/create",
		svc.BatchCreateLoadBalancer)

	h.Add("BatchUpdateLoadBalancer",
		http.MethodPatch, "/vendors/{vendor}/load_balancers/batch/update", svc.BatchUpdateLoadBalancer)
	h.Add("BatchUpdateClbBizInfo", http.MethodPatch, "/load_balancers/biz/batch/update", svc.BatchUpdateClbBizInfo)
	h.Add("GetListener", http.MethodGet, "/vendors/{vendor}/listeners/{id}", svc.GetListener)
	h.Add("ListListener", http.MethodPost, "/load_balancers/listeners/list", svc.ListListener)
	h.Add("ListTCloudUrlRule", http.MethodPost, "/vendors/tcloud/load_balancers/url_rules/list", svc.ListTCloudUrlRule)
	h.Add("ListTarget", http.MethodPost, "/load_balancers/targets/list", svc.ListTarget)
	h.Add("ListTargetGroup", http.MethodPost, "/load_balancers/target_groups/list", svc.ListTargetGroup)
	h.Add("BatchDeleteLoadBalancer", http.MethodDelete, "/load_balancers/batch", svc.BatchDeleteLoadBalancer)
	h.Add("ListTargetGroupListenerRel", http.MethodPost, "/target_group_listener_rels/list",
		svc.ListTargetGroupListenerRel)

	h.Add("BatchCreateTargetGroup", http.MethodPost, "/vendors/{vendor}/target_groups/batch/create",
		svc.BatchCreateTargetGroup)
	h.Add("UpdateTargetGroup", http.MethodPatch, "/vendors/{vendor}/target_groups", svc.UpdateTargetGroup)
	h.Add("BatchDeleteTargetGroup", http.MethodDelete, "/target_groups/batch", svc.BatchDeleteTargetGroup)
	h.Add("GetTargetGroup", http.MethodGet, "/vendors/{vendor}/target_groups/{id}", svc.GetTargetGroup)
	h.Add("CreateTargetGroupListenerRel", http.MethodPost, "/target_group_listener_rels/create",
		svc.CreateTargetGroupListenerRel)

	// url规则
	h.Add("BatchCreateTCloudUrlRule",
		http.MethodPost, "/vendors/tcloud/url_rules/batch/create", svc.BatchCreateTCloudUrlRule)
	h.Add("BatchUpdateTCloudUrlRule",
		http.MethodPatch, "/vendors/tcloud/url_rules/batch/update", svc.BatchUpdateTCloudUrlRule)
	h.Add("BatchDeleteTCloudUrlRule",
		http.MethodDelete, "/vendors/tcloud/url_rules/batch", svc.BatchDeleteTCloudUrlRule)

	// 监听器
	h.Add("BatchCreateListener", http.MethodPost, "/vendors/{vendor}/listeners/batch/create", svc.BatchCreateListener)
	h.Add("BatchCreateListenerWithRule", http.MethodPost, "/vendors/{vendor}/listeners/rules/batch/create",
		svc.BatchCreateListenerWithRule)
	h.Add("BatchUpdateListener", http.MethodPatch, "/vendors/{vendor}/listeners/batch/update", svc.BatchUpdateListener)
	h.Add("BatchDeleteListener", http.MethodDelete, "/listeners/batch", svc.BatchDeleteListener)

	h.Load(cap.WebService)
}

type lbSvc struct {
	dao dao.Set
}
