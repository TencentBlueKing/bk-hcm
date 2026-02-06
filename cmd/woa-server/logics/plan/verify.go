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

package plan

import (
	"fmt"
	"sort"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	ttypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	dt "hcm/pkg/api/core/cloud/device-type"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

// VerifyResPlanDemandV2 verify resource plan demand for subOrders.
func (c *Controller) VerifyResPlanDemandV2(kt *kit.Kit, bkBizID int64, requireType enumor.RequireType,
	subOrders []ttypes.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

	obsProject := requireType.ToObsProject()
	// get all device type maps.
	deviceTypeMap, err := c.GetAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type map failed, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, err
	}

	result := make([]ptypes.VerifyResPlanDemandElem, len(subOrders))
	resultIndex := make(map[int]int)
	verifySlice := make([]VerifyResPlanElemV2, 0)

	for idx, subOrder := range subOrders {
		// if resource type is not cvm or upgrade_cvm, set verify result to not involved.
		if subOrder.ResourceType != ttypes.ResourceTypeCvm && subOrder.ResourceType != ttypes.ResourceTypeUpgradeCvm {
			result[idx] = ptypes.VerifyResPlanDemandElem{
				VerifyResult: enumor.VerifyResPlanRstNotInvolved,
			}
			continue
		}

		baseVerifySlice := c.getVerifySliceWithDeviceInfo(deviceTypeMap, subOrder)
		verifyElems, err := c.fillVerifyElems(kt, subOrder, bkBizID, obsProject, baseVerifySlice)
		if err != nil {
			logs.Errorf("failed to get verify res plan elem, err: %v, suborder spec: %+v, rid: %s", err,
				subOrder.Spec, kt.Rid)
			return nil, err
		}

		for _, elem := range verifyElems {
			resultIndex[len(verifySlice)] = idx
			verifySlice = append(verifySlice, elem)
		}
	}

	// call verify resource plan demands to verify each cvm demands.
	rst, err := c.VerifyProdDemandsV2(kt, bkBizID, requireType, verifySlice)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand v2, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if len(rst) != len(verifySlice) {
		logs.Errorf("verify resource plan demand v2 failed, rst len: %d, verifySlice len: %d, bkBizID: %d, rid: %s",
			len(rst), len(verifySlice), bkBizID, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// set result.
	for idx, ele := range rst {
		subOrderIdx := resultIndex[idx]
		result[subOrderIdx].NeedCPUCore += ele.NeedCPUCore
		result[subOrderIdx].ResPlanCore += ele.ResPlanCore
		if result[subOrderIdx].VerifyResult == enumor.VerifyResPlanRstFailed {
			continue
		}
		result[subOrderIdx].VerifyResult = ele.VerifyResult
		result[subOrderIdx].Reason = ele.Reason
	}

	for i, res := range result {
		// 同一子单可能拆分多组 VerifyResPlanElemV2 用于校验预测，在这里统一合并计算失败的具体总核数
		if res.NeedCPUCore != 0 && res.VerifyResult == enumor.VerifyResPlanRstFailed {
			result[i].Reason = fmt.Sprintf("可用预测量不足，需要%d核，余量%d核", res.NeedCPUCore, res.ResPlanCore)
		}
	}

	logs.Infof("verify res plan demand v2 end, bkBizID: %d, verifySlice: %+v, result: %+v, rid: %s",
		bkBizID, verifySlice, result, kt.Rid)
	return result, nil
}

// getVerifySliceWithDeviceInfo 获取用于校验预测的对象，仅从子单信息中填充机型、地域、核心数等和机型强相关的字段
// 其他字段在后续 fillVerifyElems 方法中补充
func (c *Controller) getVerifySliceWithDeviceInfo(deviceTypeMap map[string]dt.DistinctDeviceType,
	subOrder ttypes.Suborder) (verifySlice []VerifyResPlanElemV2) {

	switch subOrder.ResourceType {
	case ttypes.ResourceTypeUpgradeCvm:
		// 升降配不修改云盘，不考虑云盘预测的大小
		verifySlice = make([]VerifyResPlanElemV2, 0, len(subOrder.UpgradeCVMList))
		for _, device := range subOrder.UpgradeCVMList {
			var cpuCore int64
			if deviceInfo, ok := deviceTypeMap[device.TargetInstanceType]; ok {
				cpuCore = deviceInfo.CpuCore
			}
			verifySlice = append(verifySlice, VerifyResPlanElemV2{
				DeviceType: device.TargetInstanceType,
				RegionID:   device.RegionID,
				ZoneID:     device.ZoneID,
				CpuCore:    cpuCore,
			})
		}
	default:
		// 常规申领，每个子单仅需要计算一个总核心数\总云盘大小
		var cpuCore int64
		if deviceInfo, ok := deviceTypeMap[subOrder.Spec.DeviceType]; ok {
			cpuCore = deviceInfo.CpuCore
		}
		verifySlice = append(verifySlice, VerifyResPlanElemV2{
			DeviceType: subOrder.Spec.DeviceType,
			RegionID:   subOrder.Spec.Region,
			ZoneID:     subOrder.Spec.Zone,
			CpuCore:    int64(subOrder.Replicas) * cpuCore,
		})
	}

	return verifySlice
}

func (c *Controller) fillVerifyElems(kt *kit.Kit, subOrder ttypes.Suborder, bkBizID int64, obsProject enumor.ObsProject,
	verifyGroups []VerifyResPlanElemV2) ([]VerifyResPlanElemV2, error) {

	// 是否包年包月
	isPrePaid := true
	// 升降配需同时校验预测内和预测外，按照按量计费模式校验
	if subOrder.ResourceType == ttypes.ResourceTypeUpgradeCvm ||
		subOrder.Spec.ChargeType.GetWithDefault() != cvmapi.ChargeTypePrePaid {
		isPrePaid = false
	}

	nowDemandYear, nowDemandMonth, err := c.demandTime.GetDemandYearMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand year month, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	availableTime := NewAvailableTime(nowDemandYear, nowDemandMonth)

	createElem := func(deviceType, regionID, zoneID string, cpuCore int64) VerifyResPlanElemV2 {

		return VerifyResPlanElemV2{
			IsPrePaid:     isPrePaid,
			AvailableTime: availableTime,
			DeviceType:    deviceType,
			ObsProject:    obsProject,
			BkBizID:       bkBizID,
			DemandClass:   enumor.DemandClassCVM,
			RegionID:      regionID,
			ZoneID:        zoneID,
			CpuCore:       cpuCore,
		}
	}

	verifyElems := make([]VerifyResPlanElemV2, 0, len(verifyGroups))
	for _, elem := range verifyGroups {
		verifyElems = append(verifyElems, createElem(
			elem.DeviceType,
			elem.RegionID,
			elem.ZoneID,
			elem.CpuCore,
		))
	}
	return verifyElems, nil
}

// GetPlanTypeAvlDeviceTypesV2 get plan type available device types v2.
func (c *Controller) GetPlanTypeAvlDeviceTypesV2(kt *kit.Kit, planType enumor.PlanTypeCode,
	req *ptypes.GetCvmChargeTypeDeviceTypeReq, prodRemainMap map[ResPlanPoolKeyV2]map[string]int64) (
	[]ptypes.DeviceTypeAvailable, error) {

	// get region and zone all matched device types from mongodb.
	matchedDeviceTypes, err := c.getMatchedDeviceTypes(kt, req.Region, req.Zone)
	if err != nil {
		logs.Errorf("failed to get matched device types v2, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}

	// get available device type map.
	avlDeviceTypeMap, err := c.getProdRemainAvlDeviceTypeMap(kt, req, prodRemainMap, planType)
	if err != nil {
		return nil, err
	}

	// get available device type's matched device type.
	for deviceType, remainCore := range avlDeviceTypeMap {
		matched, err := c.IsDeviceMatched(kt, matchedDeviceTypes, deviceType)
		if err != nil {
			logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for idx, ok := range matched {
			if ok && idx < len(matchedDeviceTypes) {
				avlDeviceTypeMap[matchedDeviceTypes[idx]] = remainCore
			}
		}
	}

	// convert device type map to result.
	result := make([]ptypes.DeviceTypeAvailable, len(matchedDeviceTypes))
	for idx, deviceType := range matchedDeviceTypes {
		available := false
		remainCore := int64(0)
		if deviceTypeCore, ok := avlDeviceTypeMap[deviceType]; ok {
			available = true
			remainCore = deviceTypeCore
		}

		result[idx] = ptypes.DeviceTypeAvailable{
			DeviceType: deviceType,
			Available:  available,
			RemainCore: remainCore,
		}
	}

	// sort result, put available of true to the head.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Available
	})

	return result, nil
}

func (c *Controller) getProdRemainAvlDeviceTypeMap(kt *kit.Kit, req *ptypes.GetCvmChargeTypeDeviceTypeReq,
	prodRemainMap map[ResPlanPoolKeyV2]map[string]int64, planType enumor.PlanTypeCode) (map[string]int64, error) {

	nowDemandYear, nowDemandMonth, err := c.demandTime.GetDemandYearMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand year month, err: %v, planType: %s, rid: %s", err, planType, kt.Rid)
		return nil, err
	}

	availableTime := NewAvailableTime(nowDemandYear, nowDemandMonth)
	obsProject := req.RequireType.ToObsProject()
	avlDeviceTypeMap := make(map[string]int64)

	for key, remainCoreMap := range prodRemainMap {
		//  机房裁撤需要忽略预测内、预测外 --story=121848852
		if req.RequireType == enumor.RequireTypeDissolve || key.PlanType == planType {
			avlDeviceTypeMap = getAvlDeviceTypeMap(req, key, remainCoreMap, availableTime, obsProject, avlDeviceTypeMap)
		}
	}
	return avlDeviceTypeMap, nil
}

func getAvlDeviceTypeMap(req *ptypes.GetCvmChargeTypeDeviceTypeReq, key ResPlanPoolKeyV2,
	remainCoreMap map[string]int64, availableTime AvailableTime, obsProject enumor.ObsProject,
	avlDeviceTypeMap map[string]int64) map[string]int64 {

	if key.AvailableTime == availableTime && key.ObsProject == obsProject &&
		key.BkBizID == req.BkBizID && key.RegionID == req.Region {
		for _, remain := range remainCoreMap {
			if remain > 0 {
				avlDeviceTypeMap[key.DeviceType] += remain
			}
		}
	}
	return avlDeviceTypeMap
}

// getMatchedDeviceTypes get matched device types.
func (c *Controller) getMatchedDeviceTypes(kt *kit.Kit, regionID, zoneID string) ([]string, error) {
	filters := []*filter.AtomRule{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
		tools.RuleEqual("region", regionID),
		tools.RuleEqual("disable", false),
	}
	// zone name may be empty, if it is not empty, supplement it into filter.
	if zoneID != "" && zoneID != cvmapi.CvmSeparateCampus {
		filters = append(filters, tools.RuleEqual("zone", zoneID))
	}

	matchedDeviceTypes := make([]string, 0)
	req := core.ListReq{
		Filter: tools.ExpressionAnd(filters...),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		resp, err := c.client.DataService().TCloudZiyan.DeviceType.ListDeviceType(kt, &protocloud.DeviceTypeListReq{
			ListReq: req})
		if err != nil {
			logs.Errorf("failed to list device type, err: %v, filters: %+v, rid: %s", err, filters, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			matchedDeviceTypes = append(matchedDeviceTypes, detail.DeviceType)
		}
		if len(resp.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return matchedDeviceTypes, nil
}
