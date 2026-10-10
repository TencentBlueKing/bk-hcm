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
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// LoadBalancerExclusiveClusterAssignAuditBuild load balancer exclusive cluster assign audit build. Exclusive
// cluster only supports being assigned to a biz, not delivered, other assigned resource types are rejected.
func (c *LoadBalancer) LoadBalancerExclusiveClusterAssignAuditBuild(kt *kit.Kit,
	assigns []protoaudit.CloudResourceAssignInfo) ([]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListExclusiveCluster(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		clusterInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		if one.AssignedResType != enumor.BizAuditAssignedResType {
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clusterInfo.CloudID,
			ResName:    clusterInfo.Name,
			ResType:    enumor.LoadBalancerExclusiveClusterAuditResType,
			Action:     enumor.Assign,
			BkBizID:    clusterInfo.BkBizID,
			Vendor:     clusterInfo.Vendor,
			AccountID:  clusterInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		})
	}

	return audits, nil
}

// ListExclusiveCluster list load balancer exclusive cluster.
func ListExclusiveCluster(kt *kit.Kit, dao dao.Set, ids []string) (
	map[string]tablelb.LoadBalancerExclusiveClusterTable, error) {

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerExclusiveCluster().List(kt, opt)
	if err != nil {
		logs.Errorf("list exclusive cluster failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.LoadBalancerExclusiveClusterTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
