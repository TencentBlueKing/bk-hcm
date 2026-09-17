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
	hcprotolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListBizExclusiveClusterIdleVips 查询指定四层（TGW）独占集群当前闲置的VIP列表，实时查云、不落库，
// 供业务视角独占型负载均衡购买页"指定IP"下拉选项使用。
func (svc *lbSvc) ListBizExclusiveClusterIdleVips(cts *rest.Contexts) (any, error) {
	req := new(cslb.ListExclusiveClusterIdleVipsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// only check biz authorization here, ownership of the specific cluster is verified separately below.
	_, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list biz exclusive cluster idle vips auth failed, noPermFlag: %v, err: %v, rid: %s",
			noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return nil, errf.New(errf.PermissionDenied, "no permission to access load balancer exclusive cluster")
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := lblogic.CheckExclusiveClusterIdleVipQueryable(cts.Kit, svc.client.DataService(), bkBizID,
		req.CloudClusterID); err != nil {
		return nil, err
	}

	result, err := svc.client.HCService().TCloud.Clb.DescribeClusterIdleVips(cts.Kit,
		&hcprotolb.TCloudDescribeClusterIdleVipsReq{
			AccountID: req.AccountID,
			Region:    req.Region,
			ClusterID: req.CloudClusterID,
		})
	if err != nil {
		logs.Errorf("describe exclusive cluster idle vips failed, cloud_cluster_id: %s, err: %v, rid: %s",
			req.CloudClusterID, err, cts.Kit.Rid)
		return nil, err
	}

	return &cslb.ListExclusiveClusterIdleVipsResult{Count: result.Count, Details: result.Details}, nil
}
