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

// Package clb ...
package clb

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/cvm"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// InitService initialize the clb service.
func InitService(c *capability.Capability) {
	svc := &clbSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		cvmLgc:     c.Logics.Cvm,
	}

	h := rest.NewHandler()

	// clb apis in biz
	h.Add("ListLoadBalancer", http.MethodPost, "/clbs/list", svc.ListLoadBalancer)
	h.Add("GetBizLoadBalancer", http.MethodGet, "/bizs/{bk_biz_id}/clbs/{id}", svc.GetBizLoadBalancer)
	h.Add("ListBizLoadBalancer", http.MethodPost, "/bizs/{bk_biz_id}/clbs/list", svc.ListBizLoadBalancer)

	h.Add("GetLoadBalancer", http.MethodGet, "/clbs/{id}", svc.GetLoadBalancer)
	h.Add("BatchCreateCLB", http.MethodPost, "/clbs/create", svc.BatchCreateCLB)
	h.Add("TCloudDescribeResources", http.MethodPost, "/vendors/tcloud/clbs/resources/describe",
		svc.TCloudDescribeResources)
	h.Add("BatchCreateCLB", http.MethodPost, "/clbs/create",
		svc.BatchCreateCLB)
	h.Add("ListBizListener", http.MethodPost, "/bizs/{bk_biz_id}/clbs/{clb_id}/listeners/list", svc.ListBizListener)
	h.Add("ListBizClbUrlRule", http.MethodPost, "/bizs/{bk_biz_id}/target_groups/{target_group_id}/listeners/list",
		svc.ListBizClbUrlRule)
	h.Add("GetBizListener", http.MethodGet, "/bizs/{bk_biz_id}/listeners/{id}", svc.GetBizListener)
	h.Add("GetListener", http.MethodGet, "/listeners/{id}", svc.GetListener)

	h.Load(c.WebService)
}

type clbSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disk.Interface
	cvmLgc     cvm.Interface
	eipLgc     eip.Interface
}

func (svc *clbSvc) listLoadBalancerMap(kt *kit.Kit, clbIDs []string) (map[string]coreclb.BaseClb, error) {
	if len(clbIDs) == 0 {
		return nil, nil
	}

	clbReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: clbIDs,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	clbList, err := svc.client.DataService().Global.LoadBalancer.ListClb(kt, clbReq)
	if err != nil {
		logs.Errorf("[clb] list load balancer failed, clbIDs: %v, err: %v, rid: %s", clbIDs, err, kt.Rid)
		return nil, err
	}

	clbMap := make(map[string]coreclb.BaseClb, len(clbList.Details))
	for _, clbItem := range clbList.Details {
		clbMap[clbItem.ID] = clbItem
	}

	return clbMap, nil
}
