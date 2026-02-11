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

package shortrental

import (
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	srtypes "hcm/cmd/woa-server/types/short-rental"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
)

// ListShortRentalReturnPlan 根据业务查询短租项目本月的退回计划，并将计划退回的主机分组返回
// 退回计划需按照运营产品 + 物理机机型族 + 城市分组
func (l *logics) ListShortRentalReturnPlan(kt *kit.Kit, planProductName string, opProductName string,
	hosts []*table.RecycleHost, deviceToPhysFamilyMap map[string]string) (
	map[srtypes.RecycleGroupKey][]*cvmapi.ReturnPlanItem, map[srtypes.RecycleGroupKey][]*table.RecycleHost, error) {

	// 1. 根据城市/地区ID列表获取城市/地区中文名
	regionMapPtr, err := l.client.DataService().Global.Meta.GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("failed to get region area map: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	if regionMapPtr == nil {
		logs.Errorf("failed to get region area map, regionMap is nil, rid: %s", kt.Rid)
		return nil, nil, fmt.Errorf("failed to get region area map, regionMap is nil")
	}
	regionMap := cvt.PtrToVal(regionMapPtr)

	// 2. 获取物理机机型族 + 城市分组
	groupHosts := make(map[srtypes.RecycleGroupKey][]*table.RecycleHost)
	physFamilies := make([]string, 0)
	cities := make([]string, 0)
	for _, host := range hosts {
		regionInfo, ok := regionMap[host.CloudRegionID]
		if !ok {
			logs.Errorf("cannot found region: %s, rid: %s", host.CloudRegionID, kt.Rid)
			return nil, nil, fmt.Errorf("cannot found region: %s", host.CloudRegionID)
		}
		cities = append(cities, regionInfo.RegionName)

		physFamily, ok := deviceToPhysFamilyMap[host.DeviceType]
		if !ok {
			logs.Errorf("cannot found device type: %s physical family, rid: %s", host.DeviceType, kt.Rid)
			return nil, nil, fmt.Errorf("cannot found device type: %s physical family", host.DeviceType)
		}
		physFamilies = append(physFamilies, physFamily)

		groupKey := srtypes.RecycleGroupKey{
			PlanProductName: planProductName,
			OpProductName:   opProductName,
			RegionName:      regionInfo.RegionName,
			PhysFamily:      physFamily,
		}
		groupHosts[groupKey] = append(groupHosts[groupKey], host)
	}

	// 3. 查询CRP
	returnPlans, err := l.listAllShortRentalReturnPlan(kt, []string{planProductName}, []string{opProductName},
		physFamilies, cities)
	if err != nil {
		logs.Errorf("failed to list all short rental return plan: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// 4. 结果分组返回
	groupPlans := make(map[srtypes.RecycleGroupKey][]*cvmapi.ReturnPlanItem)
	for _, returnPlan := range returnPlans {
		groupKey := srtypes.RecycleGroupKey{
			PlanProductName: returnPlan.PlanProductName,
			OpProductName:   returnPlan.ProductName,
			RegionName:      returnPlan.CityName,
			PhysFamily:      returnPlan.DeviceFamilyName,
		}
		groupPlans[groupKey] = append(groupPlans[groupKey], returnPlan)
	}

	return groupPlans, groupHosts, nil
}

// listAllShortRentalReturnPlan 从CRP查询运营产品 + 物理机机型族 + 城市分组下的所有退回计划
func (l *logics) listAllShortRentalReturnPlan(kt *kit.Kit, planProductNames, opProductNames, physFamilies,
	cities []string) ([]*cvmapi.ReturnPlanItem, error) {

	// 获取节假月的起止时间
	timeRange, err := l.demandTime.GetDemandDateRangeInMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand date range in month: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	queryParam := &cvmapi.QueryReturnPlanParam{
		// TODO CRP该接口需要用户有查询退回计划数据的权限，目前暂时用管理员来代替；需确认普通退回用户是否也有查询业务下退回计划的权限
		UserName:         strings.Split(constant.AdminHandler, ";")[0],
		StartDate:        timeRange.Start,
		EndDate:          timeRange.End,
		DeptName:         []string{cvmapi.CvmLaunchDeptName},
		PlanProductName:  planProductNames,
		ProductName:      opProductNames,
		ProjectName:      []enumor.ObsProject{enumor.ObsProjectShortLease},
		CityName:         cities,
		DeviceFamilyName: physFamilies,
		Page: &cvmapi.Page{
			Start: 0,
			Size:  int(core.DefaultMaxPageLimit),
		},
	}

	result := make([]*cvmapi.ReturnPlanItem, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		queryParam.Page.Start = start
		rst, err := l.thirdCli.CVM.QueryReturnPlan(kt.Ctx, kt.Header(), cvmapi.NewQueryReturnPlanReq(queryParam))
		if err != nil {
			logs.Errorf("failed to query return plan: %v, req: %+v, rid: %s", err, queryParam, kt.Rid)
			return nil, err
		}

		if rst.Result == nil {
			logs.Errorf("failed to query return plan, result is nil, trace_id: %s, req: %+v, rid: %s", rst.TraceId,
				cvt.PtrToVal(queryParam), kt.Rid)
			return nil, fmt.Errorf("failed to query return plan, result is nil, trace_id: %s, req: %+v", rst.TraceId,
				cvt.PtrToVal(queryParam))
		}

		result = append(result, rst.Result.Data...)

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}

// ListExecutedPlanCores 查询短租计划本月的退回执行量
// 退回计划的执行量需按照运营产品 + 物理机机型族 + 城市分组
func (l *logics) ListExecutedPlanCores(kt *kit.Kit, opProductID int64, recycleGroupKeys []srtypes.RecycleGroupKey) (
	map[srtypes.RecycleGroupKey]int64, error) {

	// 获取节假月的起止时间
	year, month, err := l.demandTime.GetDemandYearMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand year month: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	physicalFamiliesMap := make(map[string]interface{})
	regionNamesMap := make(map[string]interface{})
	for _, key := range recycleGroupKeys {
		physicalFamiliesMap[key.PhysFamily] = nil
		regionNamesMap[key.RegionName] = nil
	}

	sumReq := rpproto.ShortRentalReturnedRecordSumReq{
		PhysicalDeviceFamilies: maps.Keys(physicalFamiliesMap),
		RegionNames:            maps.Keys(regionNamesMap),
		OpProductIDs:           []int64{opProductID},
		Year:                   int64(year),
		Month:                  int64(month),
	}
	sumRsp, err := l.client.DataService().Global.ResourcePlan.SumReturnedCore(kt, &sumReq)
	if err != nil {
		logs.Errorf("failed to sum returned core: %v, req: %+v, rid: %s", err, sumReq, kt.Rid)
		return nil, err
	}

	result := make(map[srtypes.RecycleGroupKey]int64)
	for _, item := range sumRsp.Records {
		groupKey := srtypes.RecycleGroupKey{
			PlanProductName: item.PlanProductName,
			OpProductName:   item.OpProductName,
			RegionName:      item.RegionName,
			PhysFamily:      item.PhysicalDeviceFamily,
		}
		result[groupKey] = item.SumReturnedCore
	}

	return result, nil
}

// CalSplitRecycleHosts 根据短租退回计划的余量，计算并将host匹配到短租退回计划上
func (l *logics) CalSplitRecycleHosts(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost,
	recycleTypeSeq []table.RecycleType, allReturnedCPUCore, allPlanCPUCore decimal.Decimal) (
	[]*table.RecycleHost, int64, error) {

	// 1.统计所有待回收Host的CPU核数
	var err error
	hosts, err = l.getHostCPUCore(kt, hosts)
	if err != nil {
		logs.Errorf("failed to get host cpu core, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	// 2.匹配短租退回计划
	currentReturnedCore := int64(0)
	remainPlanCore := max(allPlanCPUCore.Sub(allReturnedCPUCore).IntPart(), 0)
	for _, hostItem := range hosts {
		// 物理机不需要参与
		if cmdb.IsPhysicalMachine(hostItem.SvrSourceTypeID) {
			continue
		}

		// 只能向下取整（退回数不可以超过计划数）
		if remainPlanCore < hostItem.CpuCore {
			continue
		}

		// 如果主机当前的回收类型优先级高于短租回收类型，那么该主机不能用于短租回收
		if !hostItem.RecycleType.CanUpdateRecycleType(recycleTypeSeq, table.RecycleTypeShortRental) {
			continue
		}

		// 匹配短租退回计划
		hostItem.IsMatchShortRental = true
		// 设置该Host的退还方式
		hostItem.RecycleType = table.RecycleTypeShortRental
		logs.Infof("check recycle host belongs to short rental project, bkBizID: %d, hostIP: %s, "+
			"subOrderID: %s, deviceType: %s, cpuCore: %d, planCore: %d, returnedCore: %d, currentCore: %d, rid: %s",
			bkBizID, hostItem.IP, hostItem.SuborderID, hostItem.DeviceType, hostItem.CpuCore,
			allPlanCPUCore, allReturnedCPUCore, currentReturnedCore, kt.Rid)
		// 核数累加
		currentReturnedCore += hostItem.CpuCore
		remainPlanCore = max(remainPlanCore-hostItem.CpuCore, 0)
	}

	return hosts, currentReturnedCore, nil
}

func (l *logics) getHostCPUCore(kt *kit.Kit, hosts []*table.RecycleHost) ([]*table.RecycleHost, error) {
	deviceTypes := make([]string, 0)
	for _, host := range hosts {
		deviceTypes = append(deviceTypes, host.DeviceType)
	}
	deviceTypes = slice.Unique(deviceTypes)

	// 根据设备列表获取设备机型CPU核数
	deviceTypeCpuCores := make(map[string]int64)
	for _, batch := range slice.Split(deviceTypes, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.DistinctDeviceTypeListReq{
			ListReq: core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
					tools.RuleIn("device_type", batch)),
				Page: core.NewDefaultBasePage(),
			},
		}
		rst, err := l.client.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(kt, listReq)
		if err != nil {
			logs.Errorf("list woa device type failed, err: %v, deviceTypes: %v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}

		for _, item := range rst.Details {
			deviceTypeCpuCores[item.DeviceType] = item.CpuCore
		}
	}

	// 2.计算每个回收Host对应的CPU核心数
	for _, hostItem := range hosts {
		// 物理机不需要参与计算
		if cmdb.IsPhysicalMachine(hostItem.SvrSourceTypeID) {
			continue
		}
		cpuCore, ok := deviceTypeCpuCores[hostItem.DeviceType]
		if !ok {
			logs.Errorf("device type not found, hostIP: %s, deviceType: %s, rid: %s", hostItem.IP,
				hostItem.DeviceType, kt.Rid)
			return nil, errf.Newf(errf.RecordNotFound, "device type not found, hostIP: %s, deviceType: %s",
				hostItem.IP, hostItem.DeviceType)
		}
		hostItem.CpuCore = cpuCore
	}

	return hosts, nil
}
