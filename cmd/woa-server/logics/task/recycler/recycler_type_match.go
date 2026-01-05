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

package recycler

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/maps"

	"github.com/shopspring/decimal"
)

// matchRollingServer 匹配滚服项目回收类型
func (r *recycler) matchRollingServer(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost,
	recycleTypeSeq []table.RecycleType) ([]*table.RecycleHost, error) {

	// 查询当月所有业务总的回收CPU总核心数
	allBizReturnedCpuCore, err := r.rsLogic.GetCurrentMonthAllReturnedCpuCore(kt)
	if err != nil {
		logs.Errorf("query rolling recycle all returned cpu core failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 查询系统配置的全局总额度
	globalQuota, err := r.rsLogic.GetRollingGlobalQuota(kt)
	if err != nil {
		logs.Errorf("query rolling recycle global quota config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 对业务的主机Host列表，匹配归类为“滚服项目”
	hosts, allBizReturnedCpuCore, err = r.rsLogic.CalSplitRecycleHosts(kt, bkBizID, hosts,
		recycleTypeSeq, allBizReturnedCpuCore, globalQuota)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for check recycle quota bkBizID: %d, err: %v, rid: %s",
			bkBizID, err, kt.Rid)
		return nil, err
	}

	return hosts, nil
}

// matchShortRental 匹配短租项目回收类型
func (r *recycler) matchShortRental(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost,
	recycleTypeSeq []table.RecycleType) ([]*table.RecycleHost, error) {

	// 1. 查询业务对应的规划产品、运营产品
	bizsOrgRel, err := r.bizLogic.ListBizsOrgRel(kt, []int64{bkBizID})
	if err != nil {
		logs.Errorf("failed to list bizs org rel: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, err
	}
	if _, ok := bizsOrgRel[bkBizID]; !ok {
		logs.Errorf("failed to list bizs org rel, bkBizID: %d not found, rid: %s", bkBizID, kt.Rid)
		return nil, fmt.Errorf("failed to list bizs org rel, bkBizID: %d not found", bkBizID)
	}
	planProductName := bizsOrgRel[bkBizID].PlanProductName
	opProductName := bizsOrgRel[bkBizID].OpProductName

	// 2. 查询退回的CVM机型对应的物理机机型族映射
	deviceTypes := make([]string, 0, len(hosts))
	for _, host := range hosts {
		deviceTypes = append(deviceTypes, host.DeviceType)
	}
	deviceToPhysFamilyMap, err := r.srLogic.ListDeviceTypeFamily(kt, deviceTypes)
	if err != nil {
		logs.Errorf("failed to list device type family: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 3.根据退回CVM对应物理机机型族、大小核心、所属城市，查询运营产品下的退回计划
	returnPlans, returnHosts, err := r.srLogic.ListShortRentalReturnPlan(kt, planProductName, opProductName,
		hosts, deviceToPhysFamilyMap)
	if err != nil {
		logs.Errorf("failed to list short rental return plan: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 4. 查询本地短租退回计划已执行核心数，使用运营产品、物理机机型族、所属城市 + 退回年月汇总
	executedPlans, err := r.srLogic.ListExecutedPlanCores(kt, bizsOrgRel[bkBizID].OpProductID, maps.Keys(returnHosts))
	if err != nil {
		logs.Errorf("failed to list executed plan cores: %v, group_keys: %+v, rid: %s", err,
			maps.Keys(returnHosts), kt.Rid)
		return nil, err
	}

	// 5.对业务的主机Host列表，匹配为“短租项目”
	// TODO 需要查询机型对应代次的接口，CRP排期中
	rstHosts := make([]*table.RecycleHost, 0, len(hosts))
	for groupKey, groupHosts := range returnHosts {
		// 汇总计算该分组下退回计划的总核心数
		allPlanCpuCores := decimal.NewFromInt(0)
		for _, plan := range returnPlans[groupKey] {
			allPlanCpuCores = allPlanCpuCores.Add(plan.CoreAmount)
		}

		// 汇总计算该分组下退回计划的已执行核心数
		returnedCpuCores := decimal.NewFromInt(0)
		if _, ok := executedPlans[groupKey]; ok {
			returnedCpuCores = decimal.NewFromInt(executedPlans[groupKey])
		}

		// 根据退回计划的剩余核心数判断是否可以匹配短租退回
		tmpHosts, _, err := r.srLogic.CalSplitRecycleHosts(kt, bkBizID, groupHosts, recycleTypeSeq,
			returnedCpuCores, allPlanCpuCores)
		if err != nil {
			logs.Errorf("failed to cal split recycle hosts: %v, group_key: %+v, rid: %s", err, groupKey, kt.Rid)
			return nil, err
		}
		rstHosts = append(rstHosts, tmpHosts...)
	}

	return rstHosts, nil
}
