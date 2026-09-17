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

package loadbalancer

import (
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// ListBizExclusiveClusterTags 查询业务下已分配的公网独占集群，按 (cluster_tag, cluster_type) 分组聚合返回，
// 供业务视角独占型负载均衡购买页渲染标签/集群下拉选项使用。
func (svc *lbSvc) ListBizExclusiveClusterTags(cts *rest.Contexts) (any, error) {
	req := new(cslb.ListExclusiveClusterTagsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rules := []*filter.AtomRule{
		tools.RuleEqual("network", enumor.PublicClusterNetwork),
		tools.RuleEqual("account_id", req.AccountID),
		tools.RuleEqual("region", req.Region),
		tools.RuleEqual("isp", req.Isp),
	}
	if len(req.ClusterType) != 0 {
		rules = append(rules, tools.RuleEqual("cluster_type", req.ClusterType))
	}

	// list authorized instances, ListBizAuthRes ANDs the returned filter with a bk_biz_id(from path) condition.
	expr, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: tools.ExpressionAnd(rules...)})
	if err != nil {
		logs.Errorf("list biz exclusive cluster tags auth failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag,
			err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &cslb.ListExclusiveClusterTagsResult{Details: make([]cslb.ExclusiveClusterTagGroup, 0)}, nil
	}

	return lblogic.AggregateExclusiveClusterTags(cts.Kit, svc.client.DataService(), expr, req)
}
