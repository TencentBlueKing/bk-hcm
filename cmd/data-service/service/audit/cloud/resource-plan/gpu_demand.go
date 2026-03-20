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

package resourceplan

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	rpgpuorder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
	rpgpusuborder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewResourcePlan new resource plan.
func NewResourcePlan(dao dao.Set) *ResourcePlan {
	return &ResourcePlan{
		dao: dao,
	}
}

// ResourcePlan define resource plan audit.
type ResourcePlan struct {
	dao dao.Set
}

// GpuOrderCreateAudits builds create audit entries for GPU demand main orders.
func GpuOrderCreateAudits(kt *kit.Kit, models []rpgpuorder.ResPlanDemandGpuOrderTable) []*tableaudit.AuditTable {
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, m := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:    m.ID,
			ResType:  enumor.ResPlanGPUDemandsOrderAuditResType,
			Action:   enumor.Create,
			BkBizID:  m.BkBizID,
			Vendor:   enumor.Ziyan,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: m,
			},
		})
	}

	return audits
}

// ResPlanDemandGpuOrderUpdateAuditBuild resource plan demand gpu order update audit build.
func (r *ResourcePlan) ResPlanDemandGpuOrderUpdateAuditBuild(kt *kit.Kit,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}

	idMap, err := ListResPlanDemandGpuOrder(kt, r.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		info, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ResID,
			ResType:  enumor.ResPlanGPUDemandsOrderAuditResType,
			Action:   enumor.Update,
			BkBizID:  info.BkBizID,
			Vendor:   enumor.Ziyan,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    info,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// ResPlanDemandGpuSubOrderUpdateAuditBuild resource plan demand gpu sub order update audit build.
func (r *ResourcePlan) ResPlanDemandGpuSubOrderUpdateAuditBuild(kt *kit.Kit,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}

	idMap, err := ListResPlanDemandGpuSubOrder(kt, r.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		info, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ResID,
			ResType:  enumor.ResPlanGPUDemandsSuborderAuditResType,
			Action:   enumor.Update,
			BkBizID:  info.BkBizID,
			Vendor:   enumor.Ziyan,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    info,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// ListResPlanDemandGpuOrder list resource plan demand gpu order.
func ListResPlanDemandGpuOrder(kt *kit.Kit, dao dao.Set, ids []string) (
	map[string]rpgpuorder.ResPlanDemandGpuOrderTable, error) {

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}

	list, err := dao.ResPlanDemandGpuOrder().List(kt, opt)
	if err != nil {
		logs.Errorf("list resource plan demand gpu order failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]rpgpuorder.ResPlanDemandGpuOrderTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// ListResPlanDemandGpuSubOrder list resource plan demand gpu sub order.
func ListResPlanDemandGpuSubOrder(kt *kit.Kit, dao dao.Set, ids []string) (
	map[string]rpgpusuborder.ResPlanDemandGpuSubOrderTable, error) {

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}

	list, err := dao.ResPlanDemandGpuSubOrder().List(kt, opt)
	if err != nil {
		logs.Errorf("list resource plan demand gpu sub order failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]rpgpusuborder.ResPlanDemandGpuSubOrderTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
