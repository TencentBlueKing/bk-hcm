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
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
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

	h.Add("CreateBizTargetGroup", http.MethodPost, "/bizs/{bk_biz_id}/target_groups/create", svc.CreateBizTargetGroup)
	h.Add("UpdateBizTargetGroup", http.MethodPatch, "/bizs/{bk_biz_id}/target_groups/{id}", svc.UpdateBizTargetGroup)
	h.Add("DeleteBizTargetGroup", http.MethodDelete, "/bizs/{bk_biz_id}/target_groups/batch", svc.DeleteBizTargetGroup)
	h.Add("ListBizTargetGroup", http.MethodPost, "/bizs/{bk_biz_id}/target_groups/list", svc.ListBizTargetGroup)
	h.Add("GetTargetGroup", http.MethodGet, "/target_groups/{id}", svc.GetTargetGroup)
	h.Add("AssociateBizTargetGroupListenerRel", http.MethodPost, "/bizs/{bk_biz_id}/listeners/associate/target_group",
		svc.AssociateBizTargetGroupListenerRel)

	h.Load(c.WebService)

	bizH.Add("UpdateBizTCloudLoadBalancer", http.MethodPatch, "/vendors/tcloud/load_balancers/{id}",
		svc.UpdateBizTCloudLoadBalancer)
	bizH.Add("ListBizLoadBalancer", http.MethodPost, "/load_balancers/list", svc.ListBizLoadBalancer)
	bizH.Add("GetBizLoadBalancer", http.MethodGet, "/load_balancers/{id}", svc.GetBizLoadBalancer)

	bizH.Add("ListBizListener", http.MethodPost, "/load_balancers/{lb_id}/listeners/list", svc.ListBizListener)
	bizH.Add("ListBizUrlRulesByListener", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/list", svc.ListBizUrlRulesByListener)
	bizH.Add("ListBizTCloudRuleByTG", http.MethodPost,
		"/vendors/tcloud/target_groups/{target_group_id}/rules/list", svc.ListBizTCloudRuleByTG)
	bizH.Add("GetBizListener", http.MethodGet, "/listeners/{id}", svc.GetBizListener)
	bizH.Add("GetBizListenerDomains", http.MethodGet,
		"/vendors/tcloud/listeners/{lbl_id}/domains", svc.GetBizListenerDomains)
	bizH.Add("GetBizTCloudUrlRule", http.MethodGet,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.GetBizTCloudUrlRule)

	bizH.Add("ListBizTargetsByTGID", http.MethodPost,
		"/vendors/tcloud/target_groups/{target_group_id}/targets/list", svc.ListBizTargetsByTGID)
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

func (svc *lbSvc) listLoadBalancerMap(kt *kit.Kit, lbIDs []string) (map[string]corelb.BaseLoadBalancer, error) {
	if len(lbIDs) == 0 {
		return nil, nil
	}

	clbReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lbIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, clbReq)
	if err != nil {
		logs.Errorf("list load balancer failed, lbIDs: %v, err: %v, rid: %s", lbIDs, err, kt.Rid)
		return nil, err
	}

	lbMap := make(map[string]corelb.BaseLoadBalancer, len(lbList.Details))
	for _, lbItem := range lbList.Details {
		lbMap[lbItem.ID] = lbItem
	}

	return lbMap, nil
}

func (svc *lbSvc) listListenerByID(kt *kit.Kit, lblID string, bkBizID int64) ([]corelb.BaseListener, error) {
	lblReq := &dataproto.ListListenerReq{
		ListReq: core.ListReq{
			Filter: tools.EqualExpression("id", lblID),
			Page:   core.NewDefaultBasePage(),
		},
	}
	if bkBizID > 0 {
		bizReq, err := tools.And(
			filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bkBizID}, lblReq.Filter,
		)
		if err != nil {
			return nil, err
		}
		lblReq.Filter = bizReq
	}
	lblList, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("list listener by id failed, lblID: %s, err: %v, rid: %s", lblID, err, kt.Rid)
		return nil, err
	}

	return lblList.Details, nil
}
