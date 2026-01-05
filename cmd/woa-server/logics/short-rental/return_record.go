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
	"strconv"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	mtypes "hcm/cmd/woa-server/types/meta"
	srtypes "hcm/cmd/woa-server/types/short-rental"
	"hcm/pkg"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
)

// CreateReturnedHostRecord 根据回收子单ID创建短租退回记录
func (l *logics) CreateReturnedHostRecord(kt *kit.Kit, bkBizID int64, orderID uint64, subOrderID string,
	status enumor.ShortRentalStatus) error {

	// 1. 查询回收子单对应的回收主机列表
	hosts, err := l.getRecycleHosts(kt, subOrderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts, err: %v, subOrderId: %s, rid: %s", err, subOrderID, kt.Rid)
		return err
	}

	// 2. 查询业务对应的规划产品、运营产品
	bizsOrgRel, err := l.bizLogics.ListBizsOrgRel(kt, []int64{bkBizID})
	if err != nil {
		logs.Errorf("failed to list bizs org rel: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return err
	}
	if _, ok := bizsOrgRel[bkBizID]; !ok {
		logs.Errorf("failed to list bizs org rel, bkBizID: %d not found, rid: %s", bkBizID, kt.Rid)
		return fmt.Errorf("failed to list bizs org rel, bkBizID: %d not found", bkBizID)
	}

	// 3. 查询退回的CVM机型对应的物理机机型族映射
	deviceTypes := make([]string, 0, len(hosts))
	for _, host := range hosts {
		deviceTypes = append(deviceTypes, host.DeviceType)
	}
	deviceToPhysFamilyMap, err := l.ListDeviceTypeFamily(kt, deviceTypes)
	if err != nil {
		logs.Errorf("failed to list device type family: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 4. 根据城市/地区ID列表获取城市/地区中文名
	regionMapPtr, err := l.client.DataService().Global.Meta.GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("failed to get region area map: %v, rid: %s", err, kt.Rid)
		return err
	}
	if regionMapPtr == nil {
		logs.Errorf("failed to get region area map, regionMap is nil, rid: %s", kt.Rid)
		return fmt.Errorf("failed to get region area map, regionMap is nil")
	}
	regionMap := cvt.PtrToVal(regionMapPtr)

	// 5. 获取物理机机型族 + 城市分组
	groupHosts := make(map[srtypes.RecycleGroupKey][]*table.RecycleHost)
	for _, host := range hosts {
		regionInfo, ok := regionMap[host.CloudRegionID]
		if !ok {
			logs.Errorf("cannot found region: %s, rid: %s", host.CloudRegionID, kt.Rid)
			return fmt.Errorf("cannot found region: %s", host.CloudRegionID)
		}

		physFamily, ok := deviceToPhysFamilyMap[host.DeviceType]
		if !ok {
			logs.Errorf("cannot found device type: %s physical family, rid: %s", host.DeviceType, kt.Rid)
			return fmt.Errorf("cannot found device type: %s physical family", host.DeviceType)
		}

		groupKey := srtypes.RecycleGroupKey{
			RegionID:   regionInfo.RegionID,
			RegionName: regionInfo.RegionName,
			PhysFamily: physFamily,
		}
		groupHosts[groupKey] = append(groupHosts[groupKey], host)
	}

	// 6. 为所有分组创建短租退回记录
	if err := l.createReturnedRecords(kt, bizsOrgRel[bkBizID], orderID, subOrderID, groupHosts, status); err != nil {
		logs.Errorf("failed to create returned records, err: %v, subOrderID: %s, rid: %s", err, subOrderID,
			kt.Rid)
		return err
	}
	return nil
}

// createReturnedRecords 创建短租退回记录
func (l *logics) createReturnedRecords(kt *kit.Kit, bizOrgRel *mtypes.BizOrgRel, orderID uint64, subOrderID string,
	groupHosts map[srtypes.RecycleGroupKey][]*table.RecycleHost, status enumor.ShortRentalStatus) error {

	year, month, err := l.demandTime.GetDemandYearMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand year month: %v, rid: %s", err, kt.Rid)
		return err
	}

	strTime := time.Now().Format(constant.DateLayoutCompact)
	timeInt, err := strconv.Atoi(strTime)
	if err != nil {
		logs.Errorf("convert str time to int failed, strTime: %s, err: %v", strTime, err)
		return err
	}

	createRecords := make([]rpproto.ShortRentalReturnedRecordCreateReq, 0)
	for groupKey, hosts := range groupHosts {
		returnedCore := uint64(0)
		for _, host := range hosts {
			returnedCore += uint64(host.CpuCore)
		}

		createRecords = append(createRecords, rpproto.ShortRentalReturnedRecordCreateReq{
			BkBizID:              bizOrgRel.BkBizID,
			BkBizName:            bizOrgRel.BkBizName,
			OpProductID:          bizOrgRel.OpProductID,
			OpProductName:        bizOrgRel.OpProductName,
			PlanProductID:        bizOrgRel.PlanProductID,
			PlanProductName:      bizOrgRel.PlanProductName,
			VirtualDeptID:        bizOrgRel.VirtualDeptID,
			VirtualDeptName:      bizOrgRel.VirtualDeptName,
			OrderID:              int64(orderID),
			SuborderID:           subOrderID,
			Year:                 int64(year),
			Month:                int64(month),
			ReturnedDate:         int64(timeInt),
			PhysicalDeviceFamily: groupKey.PhysFamily,
			RegionID:             groupKey.RegionID,
			RegionName:           groupKey.RegionName,
			Status:               status,
			ReturnedCore:         &returnedCore,
		})
	}

	createReq := &rpproto.ShortRentalReturnedRecordBatchCreateReq{
		Records: createRecords,
	}
	_, err = l.client.DataService().Global.ResourcePlan.BatchCreateShortRentalReturnedRecord(kt, createReq)
	if err != nil {
		logs.Errorf("failed to create short rental returned records, err: %v, subOrderID: %s, rid: %s",
			err, subOrderID, kt.Rid)
		return err
	}
	return nil
}

// UpdateReturnedStatusBySubOrderID 根据回收子单ID更新短租退回记录
func (l *logics) UpdateReturnedStatusBySubOrderID(kt *kit.Kit, subOrderID string,
	updateTo enumor.ShortRentalStatus) error {

	// 根据回收子单ID查询回收记录ID，不会超过500，不需要分页查询
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("suborder_id", subOrderID),
		),
		Page: core.NewDefaultBasePage(),
	}
	rst, err := l.client.DataService().Global.ResourcePlan.ListShortRentalReturnedRecord(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list short rental returned records, err: %v, bk subOrderID: %s, rid: %s",
			err, subOrderID, kt.Rid)
		return err
	}

	if len(rst.Details) == 0 {
		return nil
	}

	// 准备更新语句
	updateRecords := make([]rpproto.ShortRentalReturnedRecordUpdateReq, 0)
	for _, record := range rst.Details {
		updateRecords = append(updateRecords, rpproto.ShortRentalReturnedRecordUpdateReq{
			ID:     record.ID,
			Status: updateTo,
		})
	}

	updateReq := &rpproto.ShortRentalReturnedRecordBatchUpdateReq{
		Records: updateRecords,
	}
	if err = l.client.DataService().Global.ResourcePlan.BatchUpdateShortRentalReturnedRecord(kt,
		updateReq); err != nil {
		logs.Errorf("failed to update short rental returned records, err: %v, subOrderID: %s, rid: %s",
			err, subOrderID, kt.Rid)
		return err
	}
	return nil
}

// getRecycleHosts get hosts by subOrderId
func (l *logics) getRecycleHosts(kt *kit.Kit, subOrderID string) ([]*table.RecycleHost, error) {
	filter := map[string]interface{}{
		"suborder_id": subOrderID,
	}
	recycleHosts := make([]*table.RecycleHost, 0)
	startIndex := 0
	for {
		page := metadata.BasePage{
			Start: startIndex,
			Limit: pkg.BKMaxInstanceLimit,
		}
		hosts, err := dao.Set().RecycleHost().FindManyRecycleHost(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get recycle hosts, err: %v, subOrderId: %s, rid: %s", err, subOrderID, kt.Rid)
			return nil, err
		}
		recycleHosts = append(recycleHosts, hosts...)
		if len(hosts) < pkg.BKMaxInstanceLimit {
			break
		}
		startIndex += pkg.BKMaxInstanceLimit
	}

	return recycleHosts, nil
}
