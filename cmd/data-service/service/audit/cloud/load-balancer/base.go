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

// NewLoadBalancer new clb.
func NewLoadBalancer(dao dao.Set) *LoadBalancer {
	return &LoadBalancer{
		dao: dao,
	}
}

// LoadBalancer define clb audit.
type LoadBalancer struct {
	dao dao.Set
}

// LoadBalancerUpdateAuditBuild clb update audit build.
func (c *LoadBalancer) LoadBalancerUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Update,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    clbInfo,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// LoadBalancerDeleteAuditBuild clb delete audit build.
func (c *LoadBalancer) LoadBalancerDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Delete,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: clbInfo,
			},
		})
	}

	return audits, nil
}

// LoadBalancerAssignAuditBuild clb assign audit build.
func (c *LoadBalancer) LoadBalancerAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audit := &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Assign,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		}

		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			audit.Action = enumor.Assign
		case enumor.DeliverAssignedResType:
			audit.Action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// ListLoadBalancer list clb.
func ListLoadBalancer(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablelb.LoadBalancerTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancer().List(kt, opt)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, ids: %v, rid: %f", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.LoadBalancerTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
