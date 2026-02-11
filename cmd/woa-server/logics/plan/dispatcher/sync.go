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

package dispatcher

import (
	"fmt"
	"slices"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	dt "hcm/pkg/api/core/cloud/device-type"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	dmtypes "hcm/pkg/dal/dao/types/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
)

// SyncDemandFromCRPOrder 从crp订单同步预测需求，入参中的priorBizIDs允许当一个运营产品对应多个业务时，优先选择某些业务分配预测
func (d *Dispatcher) SyncDemandFromCRPOrder(kt *kit.Kit, crpSN string, priorBizIDs []int64,
	opProdToBizID map[string]int64) error {

	// 1. 根据crpSN查询并汇总crp变更记录
	changeDemandsOri, err := d.QueryCrpOrderChangeInfo(kt, crpSN)
	if err != nil {
		logs.Errorf("failed to query crp order change info, err: %v, crp_sn: %s, rid: %s", err, crpSN, kt.Rid)
		return err
	}

	// 2. 根据运营产品列表获取所有业务
	opProductNames := slice.Map(changeDemandsOri, func(item *ptypes.CrpOrderChangeInfo) string {
		return item.OpProductName
	})
	productNameBizIDMap, err := d.bizLogics.GetBkBizIDsByOpProductName(kt, opProductNames)
	if err != nil {
		logs.Errorf("failed to get bk biz id by op product name, err: %v, op product names: %v, rid: %s", err,
			opProductNames, kt.Rid)
		return err
	}

	// 3. 把key相同的预测数据聚合，避免过大的扣减在数据库中不存在
	// 这里预测没有业务信息，需要挑选运营产品中第一个业务作为预测业务，如果提供了 priorBizIDs，则优先使用该列表中的业务
	changeDemandsBizMap, err := d.aggregateDemandChangeInfoByProductName(kt, changeDemandsOri, productNameBizIDMap,
		priorBizIDs, opProdToBizID)
	if err != nil {
		logs.Errorf("failed to aggregate demand change info, err: %v, crp_sn: %s, rid: %s", err, crpSN,
			kt.Rid)
		return err
	}

	for bizID, changeDemandsMap := range changeDemandsBizMap {
		ticket, err := d.mockTicket(kt, bizID, crpSN)
		if err != nil {
			logs.Errorf("failed to mock ticket, err: %v, crp_sn: %s, biz_id: %d, rid: %s", err, crpSN, bizID,
				kt.Rid)
			return err
		}

		// changeDemand可能会在扣减时模糊匹配到同一个需求，因此需要在扣减操作生效前记录扣减量，最后统一执行
		ticketCtx := &ApplyTicketCtx{
			ID:              ticket.ID,
			Applicant:       ticket.Applicant,
			BkBizID:         ticket.BkBizID,
			BkBizName:       ticket.BkBizName,
			OpProductID:     ticket.OpProductID,
			OpProductName:   ticket.OpProductName,
			PlanProductID:   ticket.PlanProductID,
			PlanProductName: ticket.PlanProductName,
			VirtualDeptID:   ticket.VirtualDeptID,
			VirtualDeptName: ticket.VirtualDeptName,
			Remark:          ticket.Remark,
			Demands:         ticket.Demands,
			CrpSN:           ticket.CrpSN,
			DemandClass:     ticket.DemandClass,
		}
		upsertReq, updatedIDs, createLogReq, updateLogReq, err := d.prepareResPlanDemandChangeReq(kt, ticketCtx,
			changeDemandsMap)
		if err != nil {
			logs.Errorf("failed to prepare res plan demand change req, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		createdIDs, err := d.BatchUpsertResPlanDemand(kt, upsertReq, updatedIDs)
		if err != nil {
			logs.Errorf("failed to batch upsert res plan demand, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		// 创建预测的日志
		err = d.CreateResPlanChangelog(kt, createLogReq, createdIDs)
		if err != nil {
			// 变更记录创建失败不阻断整个预测的更新流程（且此时预测数据已变更，上游不应感知到error），因此此处仅记录warning，不返回error
			logs.Warnf("failed to create res plan demand changelog, err: %v, created demands: %v, rid: %s", err,
				createdIDs, kt.Rid)
		}
		// 更新预测的日志
		err = d.CreateResPlanChangelog(kt, updateLogReq, []string{})
		if err != nil {
			// 变更记录创建失败不阻断整个预测的更新流程（且此时预测数据已变更，上游不应感知到error），因此此处仅记录warning，不返回error
			logs.Warnf("failed to create res plan demand changelog, err: %v, updated demands: %v, rid: %s", err,
				updatedIDs, kt.Rid)
		}

		logs.Infof("aggregate demand change info end, crpSN: %s, bizID: %d, createdIDs: %v, updatedIDs: %v, rid: %s",
			crpSN, bizID, createdIDs, updatedIDs, kt.Rid)
	}
	return nil
}

func (d *Dispatcher) mockTicket(kt *kit.Kit, bkBizID int64, crpSN string) (*ptypes.TicketInfo, error) {
	bizOrgRel, err := d.bizLogics.GetBizOrgRel(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, bk biz id: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, err
	}

	return &ptypes.TicketInfo{
		// 目前创建操作日志时必须提供HCM单据ID，但是这个操作不是HCM发起的，需要mock一个ID
		ID:              "00000000",
		Applicant:       constant.BackendOperationUserKey,
		BkBizID:         bkBizID,
		BkBizName:       bizOrgRel.BkBizName,
		OpProductID:     bizOrgRel.OpProductID,
		OpProductName:   bizOrgRel.OpProductName,
		PlanProductID:   bizOrgRel.PlanProductID,
		PlanProductName: bizOrgRel.PlanProductName,
		VirtualDeptID:   bizOrgRel.VirtualDeptID,
		VirtualDeptName: bizOrgRel.VirtualDeptName,
		DemandClass:     enumor.DemandClassCVM,
		CrpSN:           crpSN,
	}, nil
}

func getBizIDByOpProductName(productName string, bizIDMap map[string][]int64, priorBizIDs []int64,
	opProdToBizID map[string]int64) int64 {

	bizID, ok := opProdToBizID[productName]
	if ok {
		return bizID
	}
	if bizList, ok := bizIDMap[productName]; ok {
		if len(bizList) > 0 {
			bizID = bizList[0]
		}
		for _, id := range bizList {
			if slices.Contains(priorBizIDs, id) {
				bizID = id
				break
			}
		}
	}
	return bizID
}

func (d *Dispatcher) aggregateDemandChangeInfoByProductName(kt *kit.Kit, changeDemands []*ptypes.CrpOrderChangeInfo,
	bizIDMap map[string][]int64, priorBizIDs []int64, opProdToBizID map[string]int64) (
	map[int64]map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo, error) {

	zoneNameMap, regionNameMap, deviceTypeMap, err := d.getMetas(kt)
	if err != nil {
		logs.Errorf("failed to get meta maps, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	changeDemandBizMap := make(map[int64]map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo)
	for _, changeDemand := range changeDemands {
		err := changeDemand.SetRegionAreaAndZoneID(zoneNameMap, regionNameMap)
		if err != nil {
			logs.Errorf("failed to set region area and zone id, err: %v, change demand: %+v, rid: %s", err,
				*changeDemand, kt.Rid)
			return nil, err
		}

		// 根据changeDemand的expectTime获取期望交付的时间范围
		expectTimeT, err := time.Parse(constant.DateLayout, changeDemand.ExpectTime)
		if err != nil {
			logs.Errorf("failed to parse expect time, err: %v, change demand: %+v, rid: %s", err, *changeDemand,
				kt.Rid)
			return nil, err
		}
		expectTimeRange, err := d.demandTime.GetDemandDateRangeInMonth(kt, expectTimeT)
		if err != nil {
			logs.Errorf("failed to get demand date range in month, err: %v, change demand: %+v, rid: %s", err,
				*changeDemand, kt.Rid)
			return nil, err
		}

		bizID := getBizIDByOpProductName(changeDemand.OpProductName, bizIDMap, priorBizIDs, opProdToBizID)
		if bizID == 0 {
			logs.Errorf("failed to get biz id by op product name, op product name: %s, rid: %s",
				changeDemand.OpProductName, kt.Rid)
			continue
		}
		// 因为CRP的预测和本地的不一定一致，这里需要根据聚合key获取一批通配的预测需求，在范围内调整，只保证通配范围内的总数和CRP对齐
		aggregateKey, err := changeDemand.GetAggregateKey(bizID, deviceTypeMap, expectTimeRange)
		if err != nil {
			logs.Errorf("failed to get aggregate key, err: %v, change demand: %+v, rid: %s", err, *changeDemand,
				kt.Rid)
			return nil, err
		}
		if _, ok := changeDemandBizMap[bizID]; !ok {
			changeDemandBizMap[bizID] = make(map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo)
		}
		finalChange, ok := changeDemandBizMap[bizID][aggregateKey]
		if !ok {
			changeDemandBizMap[bizID][aggregateKey] = changeDemand
			continue
		}

		// 按第一条的机型，将通配机型的OS数转换后合并
		deviceInfo, ok := deviceTypeMap[finalChange.DeviceType]
		if !ok {
			logs.Errorf("device_type: %s, not found in device_type_map, rid: %s", finalChange.DeviceType, kt.Rid)
			return nil, fmt.Errorf("device_type: %s is not found", finalChange.DeviceType)
		}
		if deviceInfo.CpuCore == 0 {
			logs.Errorf("device_type: %s, cpu_core is 0, rid: %s", finalChange.DeviceType, kt.Rid)
			return nil, fmt.Errorf("device_type: %s cpu_core is 0", finalChange.DeviceType)
		}
		changeOS := decimal.NewFromInt(changeDemand.ChangeCpuCore).Div(decimal.NewFromInt(deviceInfo.CpuCore))

		finalChange.ChangeOs = finalChange.ChangeOs.Add(changeOS)
		finalChange.ChangeCpuCore = finalChange.ChangeCpuCore + changeDemand.ChangeCpuCore
		finalChange.ChangeMemory = finalChange.ChangeMemory + changeDemand.ChangeMemory

		finalChange.ChangeDiskSize = finalChange.ChangeDiskSize + changeDemand.ChangeDiskSize
	}

	return changeDemandBizMap, nil
}

func (d *Dispatcher) getMetas(kt *kit.Kit) (map[string]string, map[string]dmtypes.RegionArea,
	map[string]dt.DistinctDeviceType, error) {

	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, deviceTypeMap, err := d.resFetcher.GetMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	zoneNameMap, regionNameMap := d.resFetcher.GetMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	return zoneNameMap, regionNameMap, deviceTypeMap, nil
}
