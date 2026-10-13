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
	"encoding/json"
	"fmt"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListExclusiveCluster list resource load balancer exclusive cluster.
func (svc *lbSvc) ListExclusiveCluster(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := handler.ListResourceAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list exclusive cluster auth failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag, err,
			cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &cslb.ListExclusiveClusterResult{Count: 0,
			Details: make([]corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension], 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.ListExclusiveCluster(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list exclusive cluster failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if result == nil {
		logs.Errorf("list exclusive cluster got empty response, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list exclusive cluster got empty response, rid: %s", cts.Kit.Rid)
	}

	if req.Page.Count {
		return &cslb.ListExclusiveClusterResult{Count: result.Count}, nil
	}

	details := make([]corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension], 0, len(result.Details))
	for _, one := range result.Details {
		ext := new(corelb.TCloudExclusiveClusterExtension)
		if err := json.Unmarshal(one.Extension, ext); err != nil {
			logs.Errorf("unmarshal exclusive cluster extension failed, id: %s, err: %v, rid: %s", one.ID, err,
				cts.Kit.Rid)
			return nil, fmt.Errorf("unmarshal exclusive cluster(id=%s) extension failed, err: %v", one.ID, err)
		}

		details = append(details, corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]{
			BaseExclusiveCluster: one.BaseExclusiveCluster,
			Extension:            ext,
		})
	}

	return &cslb.ListExclusiveClusterResult{Details: details}, nil
}
