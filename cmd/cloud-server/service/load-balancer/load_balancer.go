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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService initialize the clb service.
func InitService(c *capability.Capability) {
	svc := &lbSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		cvmLgc:     c.Logics.Cvm,
	}

	h := rest.NewHandler()

	// clb apis in biz
	h.Add("ListLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListLoadBalancer)
	h.Add("GetBizLoadBalancer", http.MethodGet, "/bizs/{bk_biz_id}/load_balancers/{id}", svc.GetBizLoadBalancer)
	h.Add("UpdateBizTCloudLoadBalancer", http.MethodPatch, "/bizs/{bk_biz_id}/vendors/tcloud/load_balancers/{id}",
		svc.UpdateBizTCloudLoadBalancer)
	h.Add("ListBizLoadBalancer", http.MethodPost, "/bizs/{bk_biz_id}/load_balancers/list", svc.ListBizLoadBalancer)

	h.Add("GetLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetLoadBalancer)
	h.Add("TCloudDescribeResources", http.MethodPost, "/vendors/tcloud/load_balancers/resources/describe",
		svc.TCloudDescribeResources)
	h.Add("BatchCreateLB", http.MethodPost, "/load_balancers/create", svc.BatchCreateLB)
	h.Add("AssignLbToBiz", http.MethodPost, "/load_balancers/assign/bizs", svc.AssignLbToBiz)
	h.Add("ListBizListener", http.MethodPost, "/bizs/{bk_biz_id}/load_balancers/{lb_id}/listeners/list",
		svc.ListBizListener)
	h.Add("ListBizLbUrlRule", http.MethodPost, "/bizs/{bk_biz_id}/target_groups/{target_group_id}/listeners/list",
		svc.ListBizLbUrlRule)
	h.Add("GetBizListener", http.MethodGet, "/bizs/{bk_biz_id}/listeners/{id}", svc.GetBizListener)

	h.Load(c.WebService)
}

type lbSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disk.Interface
	cvmLgc     cvm.Interface
	eipLgc     eip.Interface
}

func (svc *lbSvc) listLoadBalancerMap(kt *kit.Kit, clbIDs []string) (map[string]corelb.BaseLoadBalancer, error) {
	if len(clbIDs) == 0 {
		return nil, nil
	}

	clbReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", clbIDs),
		Page:   core.NewDefaultBasePage(),
	}
	clbList, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, clbReq)
	if err != nil {
		logs.Errorf("[clb] list load balancer failed, clbIDs: %v, err: %v, rid: %s", clbIDs, err, kt.Rid)
		return nil, err
	}

	clbMap := make(map[string]corelb.BaseLoadBalancer, len(clbList.Details))
	for _, clbItem := range clbList.Details {
		clbMap[clbItem.ID] = clbItem
	}

	return clbMap, nil
}
