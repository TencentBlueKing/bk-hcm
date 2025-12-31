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

package fetcher

import (
	"fmt"
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"
)

// GetResPlanDemandDetail get res plan demand detail
func (f *ResPlanFetcher) GetResPlanDemandDetail(kt *kit.Kit, demandID string, bkBizIDs []int64) (
	*ptypes.GetPlanDemandDetailResp, error) {

	listRules := make([]*filter.AtomRule, 0)
	listRules = append(listRules, tools.RuleEqual("id", demandID))
	if len(bkBizIDs) > 0 {
		listRules = append(listRules, tools.RuleIn("bk_biz_id", bkBizIDs))
	}

	listReq := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(listRules...),
			Page:   core.NewDefaultBasePage(),
		},
	}

	rst, err := f.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) == 0 {
		return nil, fmt.Errorf("demand %s not found in bk_biz_id: %v", demandID, bkBizIDs)
	}
	detail := rst.Details[0]

	expectDateStr, err := times.TransTimeStrWithLayout(strconv.Itoa(detail.ExpectTime),
		constant.DateLayoutCompact, constant.DateLayout)
	if err != nil {
		logs.Errorf("failed to parse demand expect time, err: %v, expect_time: %d, rid: %s", err,
			detail.ExpectTime, kt.Rid)
		return nil, err
	}
	// 短租项目需要提供预期退回时间字段
	returnTimePtr := new(string)
	if detail.ObsProject == enumor.ObsProjectShortLease {
		returnTime, err := times.TransTimeStrWithLayout(strconv.Itoa(detail.ReturnPlanTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Warnf("failed to parse demand return plan time, err: %v, return_plan_time: %d, rid: %s", err,
				detail.ReturnPlanTime, kt.Rid)
		} else {
			returnTimePtr = &returnTime
		}
	}

	result := &ptypes.GetPlanDemandDetailResp{
		DemandID:        detail.ID,
		ExpectTime:      expectDateStr,
		ReturnPlanTime:  returnTimePtr,
		BkBizID:         detail.BkBizID,
		BkBizName:       detail.BkBizName,
		DeptID:          detail.VirtualDeptID,
		DeptName:        detail.VirtualDeptName,
		PlanProductID:   detail.PlanProductID,
		PlanProductName: detail.PlanProductName,
		OpProductID:     detail.OpProductID,
		OpProductName:   detail.OpProductName,
		ObsProject:      detail.ObsProject,
		AreaName:        detail.AreaName,
		RegionID:        detail.RegionID,
		RegionName:      detail.RegionName,
		ZoneID:          detail.ZoneID,
		ZoneName:        detail.ZoneName,
		PlanType:        detail.PlanType.Name(),
		CoreType:        detail.CoreType,
		DeviceFamily:    detail.DeviceFamily,
		DeviceClass:     detail.DeviceClass,
		DeviceType:      detail.DeviceType,
		OS:              detail.OS.Decimal,
		Memory:          cvt.PtrToVal(detail.Memory),
		CpuCore:         cvt.PtrToVal(detail.CpuCore),
		DiskSize:        cvt.PtrToVal(detail.DiskSize),
		DiskIO:          detail.DiskIO,
		DiskType:        detail.DiskType,
		DiskTypeName:    detail.DiskType.Name(),
		ResMode:         detail.ResMode.Name(),
	}
	return result, nil
}

// ListResPlanDemandByAggregateKey list res plan demand by key.
func (f *ResPlanFetcher) ListResPlanDemandByAggregateKey(kt *kit.Kit, demandKey ptypes.ResPlanDemandAggregateKey) (
	[]rpd.ResPlanDemandTable, error) {

	listRules := make([]*filter.AtomRule, 0)

	startExpTime, err := times.ConvStrTimeToInt(demandKey.ExpectTimeRange.Start, constant.DateLayout)
	if err != nil {
		logs.Errorf("failed to parse month range, err: %v, month_range: %v, rid: %s", err,
			demandKey.ExpectTimeRange, kt.Rid)
		return nil, err
	}
	endExpTime, err := times.ConvStrTimeToInt(demandKey.ExpectTimeRange.End, constant.DateLayout)
	if err != nil {
		logs.Errorf("failed to parse month range, err: %v, month_range: %v, rid: %s", err,
			demandKey.ExpectTimeRange, kt.Rid)
		return nil, err
	}

	listRules = append(listRules, tools.RuleEqual("bk_biz_id", demandKey.BkBizID))
	listRules = append(listRules, tools.RuleEqual("plan_type", demandKey.PlanType))
	listRules = append(listRules, tools.RuleEqual("obs_project", demandKey.ObsProject))

	listRules = append(listRules, tools.RuleGreaterThanEqual("expect_time", startExpTime))
	listRules = append(listRules, tools.RuleLessThanEqual("expect_time", endExpTime))

	listRules = append(listRules, tools.RuleEqual("region_id", demandKey.RegionID))
	listRules = append(listRules, tools.RuleEqual("device_family", demandKey.DeviceFamily))
	listRules = append(listRules, tools.RuleEqual("core_type", demandKey.CoreType))
	listRules = append(listRules, tools.RuleEqual("demand_res_type", demandKey.ResType))
	// DiskType == DiskUnknown 时，允许匹配任何磁盘类型
	if demandKey.DiskType != enumor.DiskUnknown {
		listRules = append(listRules, tools.RuleEqual("disk_type", demandKey.DiskType))
	}

	listFilter := tools.ExpressionAnd(listRules...)

	listReq := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Filter: listFilter,
			Page: &core.BasePage{
				Start: 0,
				Limit: core.DefaultMaxPageLimit,
			},
		},
	}

	result := make([]rpd.ResPlanDemandTable, 0)
	for {
		rst, err := f.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list local res plan demand, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result = append(result, rst.Details...)

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return result, nil
}
