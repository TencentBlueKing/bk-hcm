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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	dt "hcm/pkg/api/core/cloud/device-type"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	rtypes "hcm/pkg/dal/dao/types/resource-plan"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// queryAndApplyResPlanDemandChange query and apply res plan demand change.
func (d *Dispatcher) queryAndApplyResPlanDemandChange(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	// call crp api to get plan order change info.
	changeDemandsOri, err := d.QueryCrpOrderChangeInfosByTicketID(kt, ticket.ID)
	if err != nil {
		logs.Errorf("failed to query crp order change info, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	// 所有子单全部未通过
	if len(changeDemandsOri) == 0 {
		// 解锁单据关联的原始预测
		unlockErr := d.unlockTicketOriginalDemands(kt, ticket.Demands)
		if unlockErr != nil {
			logs.Warnf("failed to unlock ticket original demands, err: %v, id: %s, rid: %s", unlockErr,
				ticket.ID, kt.Rid)
		}
		logs.Infof("no passed sub ticket, ticket id: %s, rid: %s", ticket.ID, kt.Rid)
		return nil
	}

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
	return d.ApplyResPlanDemandChange(kt, ticketCtx, changeDemandsOri)
}

// ApplyTicketCtx a context for ApplyResPlanDemandChange
type ApplyTicketCtx struct {
	ID              string             `json:"id"`
	Applicant       string             `json:"applicant"`
	BkBizID         int64              `json:"bk_biz_id"`
	BkBizName       string             `json:"bk_biz_name"`
	OpProductID     int64              `json:"op_product_id"`
	OpProductName   string             `json:"op_product_name"`
	PlanProductID   int64              `json:"plan_product_id"`
	PlanProductName string             `json:"plan_product_name"`
	VirtualDeptID   int64              `json:"virtual_dept_id"`
	VirtualDeptName string             `json:"virtual_dept_name"`
	Remark          string             `json:"remark"`
	Demands         rpt.ResPlanDemands `json:"demands"`
	CrpSN           string             `json:"crp_sn"`
	DemandClass     enumor.DemandClass `json:"demand_class"`
}

// ApplyResPlanDemandChange apply res plan demand change.
func (d *Dispatcher) ApplyResPlanDemandChange(kt *kit.Kit, ticketCtx *ApplyTicketCtx,
	changeInfos []*ptypes.CrpOrderChangeInfo) error {

	// 需要先把key相同的预测数据聚合，避免过大的扣减在数据库中不存在
	changeDemandsMap, err := d.aggregateDemandChangeInfo(kt, changeInfos, ticketCtx)
	if err != nil {
		logs.Errorf("failed to aggregate demand change info, err: %v, crp_sn: %s, rid: %s", err, ticketCtx.CrpSN,
			kt.Rid)
		return err
	}
	logs.Infof("aggregate demand change info start, ticketID: %s, crpSn: %s, changeDemandsMap: %+v, rid: %s",
		ticketCtx.ID, ticketCtx.CrpSN, cvt.PtrToSlice(maps.Values(changeDemandsMap)), kt.Rid)

	// changeDemand可能会在扣减时模糊匹配到同一个需求，因此需要在扣减操作生效前记录扣减量，最后统一执行
	upsertReq, updatedIDs, createLogReq, updateLogReq, err := d.prepareResPlanDemandChangeReq(kt, ticketCtx,
		changeDemandsMap)
	if err != nil {
		logs.Errorf("failed to prepare res plan demand change req, err: %v, ticket: %s, rid: %s", err,
			ticketCtx.ID, kt.Rid)
		return err
	}
	if len(upsertReq.CreateDemands) == 0 && len(upsertReq.UpdateDemands) == 0 {
		// 解锁单据关联的原始预测
		unlockErr := d.unlockTicketOriginalDemands(kt, ticketCtx.Demands)
		if unlockErr != nil {
			logs.Warnf("failed to unlock ticket original demands, err: %v, id: %s, rid: %s", unlockErr,
				ticketCtx.ID, kt.Rid)
		}
		logs.Infof("no need to update res plan demand, ticket id: %s, rid: %s", ticketCtx.ID, kt.Rid)
		return nil
	}

	createdIDs, err := d.BatchUpsertResPlanDemand(kt, upsertReq, updatedIDs)
	if err != nil {
		logs.Errorf("failed to batch upsert res plan demand, err: %v, ticket: %s, rid: %s", err, ticketCtx.ID,
			kt.Rid)
		return err
	}

	// 单据关联的原始预测也需要解锁，避免部分预测死锁
	unlockErr := d.unlockTicketOriginalDemands(kt, ticketCtx.Demands)
	if unlockErr != nil {
		logs.Warnf("failed to unlock ticket original demands, err: %v, id: %s, rid: %s", unlockErr,
			ticketCtx.ID, kt.Rid)
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

	logs.Infof("aggregate demand change info end, ticketID: %s, crpSn: %s, createdIDs: %v, updatedIDs: %v, rid: %s",
		ticketCtx.ID, ticketCtx.CrpSN, createdIDs, updatedIDs, kt.Rid)
	return nil
}

// CreateResPlanChangelog create resource plan demand changelog.
// If demandIDs is provided, will reset the changelog's demand id.
func (d *Dispatcher) CreateResPlanChangelog(kt *kit.Kit, changelogReqs []rpproto.DemandChangelogCreate,
	demandIDs []string) error {
	if len(changelogReqs) == 0 {
		return nil
	}

	if len(demandIDs) > 0 && len(demandIDs) != len(changelogReqs) {
		logs.Errorf("demand ids and changelog create reqs length not equal, demand ids: %v, "+
			"changelog create reqs: %v, rid: %s", demandIDs, changelogReqs, kt.Rid)
		return fmt.Errorf("demand ids and changelog create reqs length not equal")
	}

	createDemandIDs := make([]string, len(changelogReqs))
	for idx := range changelogReqs {
		if len(demandIDs) > 0 {
			changelogReqs[idx].DemandID = demandIDs[idx]
		}
		createDemandIDs[idx] = changelogReqs[idx].DemandID
	}

	changelogReq := &rpproto.DemandChangelogCreateReq{
		Changelogs: changelogReqs,
	}

	_, err := d.client.DataService().Global.ResourcePlan.BatchCreateDemandChangelog(kt, changelogReq)
	if err != nil {
		logs.Errorf("failed to create plan demand changelog, demand ids: %v, err: %v, rid: %s", createDemandIDs,
			err, kt.Rid)
		return err
	}

	return nil
}

// prepareResPlanDemandChangeReq 准备更新资源预测表的参数
// 返回值：更新参数，被更新的资源ID，新增预测的变更日志，修改预测的变更日志
// 因新增预测的变更日志需要在更新完成后追加预测ID，因此需要和修改的变更日志分开返回
func (d *Dispatcher) prepareResPlanDemandChangeReq(kt *kit.Kit, ticket *ApplyTicketCtx,
	changeDemandsMap map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo) (
	*rpproto.ResPlanDemandBatchUpsertReq, []string, []rpproto.DemandChangelogCreate, []rpproto.DemandChangelogCreate,
	error) {

	// needUpdateDemands记录全部扣减完成后的最终结果
	needUpdateDemands := make(map[string]*rpd.ResPlanDemandTable)
	// demandUpdatedResource记录变化量，用于记录变更日志
	demandUpdatedResource := make(map[string]*ptypes.DemandResource)

	// 从 woa_zone 获取城市/地区的中英文对照
	deviceTypeMap, err := d.getAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, err
	}

	batchCreateReq := make([]rpproto.ResPlanDemandCreateReq, 0)
	batchCreateLogReq := make([]rpproto.DemandChangelogCreate, 0)
	// 最后再调整disk_type为空的
	changeDemandsSlice := maps.PartitionSortKeys(changeDemandsMap, func(key ptypes.ResPlanDemandAggregateKey) bool {
		return key.DiskType != enumor.DiskUnknown
	})
	for _, aggregateKey := range changeDemandsSlice {
		changeDemand := changeDemandsMap[aggregateKey]

		localDemands, err := d.resFetcher.ListResPlanDemandByAggregateKey(kt, aggregateKey)
		if err != nil {
			logs.Errorf("failed to get local res plan demands, err: %v, aggregate key: %+v, change demand: %+v, "+
				"rid: %s", err, aggregateKey, *changeDemand, kt.Rid)
			return nil, nil, nil, nil, err
		}

		// 根据匹配情况决定如何调整数据库中的现有预测需求
		createReq, createLogReq, err := d.applyResPlanDemandChangeAggregate(kt, ticket, localDemands,
			changeDemand, needUpdateDemands, demandUpdatedResource)
		if err != nil {
			logs.Errorf("failed to apply res plan demand change aggregate, err: %v, change demand: %+v, rid: %s",
				err, *changeDemand, kt.Rid)
			return nil, nil, nil, nil, err
		}
		if createReq != nil {
			batchCreateReq = append(batchCreateReq, cvt.PtrToVal(createReq))
			batchCreateLogReq = append(batchCreateLogReq, cvt.PtrToVal(createLogReq))
		}
	}

	// 准备更新语句
	updatedIDs, batchUpdateReq, updateChangelogReqs, err := convUpdateResPlanDemandReqs(kt, ticket, needUpdateDemands,
		demandUpdatedResource, deviceTypeMap)
	if err != nil {
		logs.Errorf("failed to convert update res plan demand reqs, err: %v, ticket: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return nil, nil, nil, nil, err
	}

	upsertReq := &rpproto.ResPlanDemandBatchUpsertReq{
		CreateDemands: batchCreateReq,
		UpdateDemands: batchUpdateReq,
	}

	return upsertReq, updatedIDs, batchCreateLogReq, updateChangelogReqs, nil
}

func convUpdateResPlanDemandReqs(kt *kit.Kit, ticket *ApplyTicketCtx,
	updateDemands map[string]*rpd.ResPlanDemandTable, demandChangelog map[string]*ptypes.DemandResource,
	deviceTypeMap map[string]dt.DistinctDeviceType) ([]string, []rpproto.ResPlanDemandUpdateReq,
	[]rpproto.DemandChangelogCreate, error) {

	updatedDemandIDs := make([]string, 0)
	batchUpdateReq := make([]rpproto.ResPlanDemandUpdateReq, 0)
	batchCreateChangelogReq := make([]rpproto.DemandChangelogCreate, 0)

	for demandID, updated := range updateDemands {
		deviceInfo, ok := deviceTypeMap[updated.DeviceType]
		if !ok {
			logs.Errorf("device_type: %s, not found in device_type_map, rid: %s", updated.DeviceType, kt.Rid)
			return nil, nil, nil, fmt.Errorf("device_type: %s is not found", updated.DeviceType)
		}

		updatedDemandIDs = append(updatedDemandIDs, demandID)

		// OS and memory should be calculated by cpu core
		updatedOS := decimal.NewFromInt(cvt.PtrToVal(updated.CpuCore)).Div(decimal.NewFromInt(deviceInfo.CpuCore))
		updatedMemory := updatedOS.Mul(decimal.NewFromInt(deviceInfo.Memory)).IntPart()

		updateReq := rpproto.ResPlanDemandUpdateReq{
			ID:       demandID,
			OS:       &updatedOS,
			CpuCore:  updated.CpuCore,
			Memory:   &updatedMemory,
			DiskSize: updated.DiskSize,
		}
		if kt.User == constant.BackendOperationUserKey {
			updateReq.Reviser = ticket.Applicant
		}
		batchUpdateReq = append(batchUpdateReq, updateReq)

		// 变更日志
		demandChangeRes, ok := demandChangelog[demandID]
		if !ok {
			// 更新日志无法创建不阻断整个预测的更新流程，因此此处仅记录warning，不返回error
			logs.Warnf("failed to get demand change res, demand_id: %s, rid: %s", demandID, kt.Rid)
			continue
		}

		expectTimeStr, err := times.TransTimeStrWithLayout(strconv.Itoa(updated.ExpectTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Warnf("failed to convert expect time to string, err: %v, expect time: %d, demand_id: %s, rid: %s",
				err, updated.ExpectTime, demandID, kt.Rid)
			continue
		}

		changeCpuCore := demandChangeRes.CpuCore
		changeDiskSize := demandChangeRes.DiskSize
		changeOS := decimal.NewFromInt(changeCpuCore).Div(decimal.NewFromInt(deviceInfo.CpuCore))
		changeMemory := changeOS.Mul(decimal.NewFromInt(deviceInfo.Memory)).IntPart()

		logCreateReq := rpproto.DemandChangelogCreate{
			DemandID:       demandID,
			TicketID:       ticket.ID,
			CrpOrderID:     ticket.CrpSN,
			SuborderID:     "",
			Type:           enumor.DemandChangelogTypeAdjust,
			ExpectTime:     expectTimeStr,
			ObsProject:     updated.ObsProject,
			RegionName:     updated.RegionName,
			ZoneName:       updated.ZoneName,
			DeviceType:     updated.DeviceType,
			OSChange:       &changeOS,
			CpuCoreChange:  &changeCpuCore,
			MemoryChange:   &changeMemory,
			DiskSizeChange: &changeDiskSize,
			Remark:         ticket.Remark,
		}
		batchCreateChangelogReq = append(batchCreateChangelogReq, logCreateReq)
	}

	return updatedDemandIDs, batchUpdateReq, batchCreateChangelogReq, nil
}

// aggregateDemandChangeInfo 聚合预测变更信息
func (d *Dispatcher) aggregateDemandChangeInfo(kt *kit.Kit, changeDemands []*ptypes.CrpOrderChangeInfo,
	ticket *ApplyTicketCtx) (map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo, error) {

	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, _, err := d.resFetcher.GetMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	zoneNameMap, regionNameMap := d.resFetcher.GetMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	deviceTypeMap, err := d.getAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	changeDemandMap := make(map[ptypes.ResPlanDemandAggregateKey]*ptypes.CrpOrderChangeInfo)
	for _, changeDemand := range changeDemands {
		// 剔除非本业务的变更，这些是转移到中转池的变更，不需要体现在HCM
		if changeDemand.OpProductName != ticket.OpProductName {
			logs.Infof("op product name: %s, not match ticket op product name: %s, ticket id: %s, rid: %s",
				changeDemand.OpProductName, ticket.OpProductName, ticket.ID, kt.Rid)
			continue
		}

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

		// 因为CRP的预测和本地的不一定一致，这里需要根据聚合key获取一批通配的预测需求，在范围内调整，只保证通配范围内的总数和CRP对齐
		aggregateKey, err := changeDemand.GetAggregateKey(ticket.BkBizID, deviceTypeMap, expectTimeRange)
		if err != nil {
			logs.Errorf("failed to get aggregate key, err: %v, change demand: %+v, rid: %s", err, *changeDemand,
				kt.Rid)
			return nil, err
		}
		finalChange, ok := changeDemandMap[aggregateKey]
		if !ok {
			changeDemandMap[aggregateKey] = changeDemand
			continue
		}

		// 按第一条的机型，将通配机型的OS数转换后合并
		deviceInfo, ok := deviceTypeMap[finalChange.DeviceType]
		if !ok {
			logs.Errorf("device_type: %s, not found in device_type_map, rid: %s", finalChange.DeviceType, kt.Rid)
			return nil, fmt.Errorf("device_type: %s is not found", finalChange.DeviceType)
		}
		changeOS := decimal.NewFromInt(changeDemand.ChangeCpuCore).Div(decimal.NewFromInt(deviceInfo.CpuCore))

		finalChange.ChangeOs = finalChange.ChangeOs.Add(changeOS)
		finalChange.ChangeCpuCore = finalChange.ChangeCpuCore + changeDemand.ChangeCpuCore
		finalChange.ChangeMemory = finalChange.ChangeMemory + changeDemand.ChangeMemory

		finalChange.ChangeDiskSize = finalChange.ChangeDiskSize + changeDemand.ChangeDiskSize
	}

	return changeDemandMap, nil
}

// applyResPlanDemandChangeAggregate apply changes according to aggregation rules.
// return the requests to create, and the needUpdateDemands map to update with a transaction.
func (d *Dispatcher) applyResPlanDemandChangeAggregate(kt *kit.Kit, ticket *ApplyTicketCtx,
	localDemands []rpd.ResPlanDemandTable, changeDemand *ptypes.CrpOrderChangeInfo,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource) (
	*rpproto.ResPlanDemandCreateReq, *rpproto.DemandChangelogCreate, error) {

	// 追加预测，不需要模糊匹配
	if changeDemand.ChangeCpuCore > 0 {
		// 优先追加到完全匹配的预测需求上，如果找不到，直接新增
		demandKey := changeDemand.GetKey(ticket.BkBizID, ticket.DemandClass)
		matchDemands := matchDemandTableByDemandKey(kt, localDemands, demandKey)

		// 没有完全匹配的，直接新增
		if len(matchDemands) == 0 {
			createReq, createLogReq, err := convCreateResPlanDemandReqs(kt, ticket, changeDemand)
			if err != nil {
				logs.Errorf("failed to convert create res plan demand reqs, err: %v, bk_biz_id: %d, "+
					"changeDemand: %+v, rid: %s", err, ticket.BkBizID, changeDemand, kt.Rid)
				return nil, nil, err
			}

			return &createReq, &createLogReq, nil
		}

		// 以demandKey唯一键匹配到的数据，最多只可能有1条，追加
		localDemand := matchDemands[0]
		// 对已有数据的变更只记录变更量，全部变更处理完成后再统一拼装更新请求
		if _, ok := needUpdateDemands[localDemand.ID]; !ok {
			needUpdateDemands[localDemand.ID] = &localDemand
			demandUpdateRes[localDemand.ID] = &ptypes.DemandResource{
				DeviceType: localDemand.DeviceType,
				CpuCore:    0,
				DiskSize:   0,
			}
		}
		remainedCpuCore := cvt.PtrToVal(needUpdateDemands[localDemand.ID].CpuCore) + changeDemand.ChangeCpuCore
		remainedDiskSize := cvt.PtrToVal(needUpdateDemands[localDemand.ID].DiskSize) + changeDemand.ChangeDiskSize

		needUpdateDemands[localDemand.ID].CpuCore = &remainedCpuCore
		needUpdateDemands[localDemand.ID].DiskSize = &remainedDiskSize
		demandUpdateRes[localDemand.ID].CpuCore += changeDemand.ChangeCpuCore
		demandUpdateRes[localDemand.ID].DiskSize += changeDemand.ChangeDiskSize

		return nil, nil, nil
	}

	// 调减预测，在范围内随机模糊调减，优先调减加锁的
	err := d.DeductResPlanDemandAggregate(kt, localDemands, needUpdateDemands, demandUpdateRes, changeDemand)
	if err != nil {
		logs.Errorf("failed to update res plan demand aggregate, err: %v, bk_biz_id: %d, changeDemand: %+v, rid: %s",
			err, ticket.BkBizID, changeDemand, kt.Rid)
		return nil, nil, err
	}

	return nil, nil, nil
}

// QueryCrpOrderChangeInfosByTicketID query crp order change info by ticket id.
func (d *Dispatcher) QueryCrpOrderChangeInfosByTicketID(kt *kit.Kit, ticketID string) ([]*ptypes.CrpOrderChangeInfo,
	error) {

	subTickets := make([]ptypes.ListResPlanSubTicketItem, 0)
	// 获取单据下所有成功的子单
	listReq := &ptypes.ListResPlanSubTicketReq{
		TicketID: ticketID,
		Statuses: []enumor.RPSubTicketStatus{enumor.RPSubTicketStatusDone},
		Page:     core.NewDefaultBasePage(),
	}
	for {
		rst, err := d.resFetcher.ListResPlanSubTicket(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list res plan sub ticket, err: %v, id: %s, rid: %s", err, ticketID, kt.Rid)
			return nil, err
		}
		subTickets = append(subTickets, rst.Details...)

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	crpChangeInfos := make([]*ptypes.CrpOrderChangeInfo, 0)
	for _, subTicket := range subTickets {
		changeInfo, err := d.QueryCrpOrderChangeInfo(kt, subTicket.CrpSN)
		if err != nil {
			logs.Errorf("failed to query crp order change info, err: %v, sub_ticket_id: %s, sn: %s, rid: %s",
				err, subTicket.ID, subTicket.CrpSN, kt.Rid)
			return nil, err
		}

		crpChangeInfos = append(crpChangeInfos, changeInfo...)
	}
	return crpChangeInfos, nil
}

// QueryCrpOrderChangeInfo query crp order change info.
func (d *Dispatcher) QueryCrpOrderChangeInfo(kt *kit.Kit, orderID string) ([]*ptypes.CrpOrderChangeInfo,
	error) {

	// init request parameter.
	queryReq := &cvmapi.PlanOrderChangeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderChangeMethod,
		},
		Params: &cvmapi.PlanOrderChangeParam{
			Page: &cvmapi.Page{
				Start: 0,
				Size:  int(core.DefaultMaxPageLimit),
			},
			OrderId: []string{orderID},
			BgName:  []string{cvmapi.CvmCbsPlanQueryBgName},
		},
	}

	// query all demands.
	result := make([]*ptypes.CrpOrderChangeInfo, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		queryReq.Params.Page.Start = start
		rst, err := d.crpCli.QueryPlanOrderChange(kt.Ctx, kt.Header(), queryReq)
		if err != nil {
			logs.Errorf("failed to query crp order change info, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if rst.Error.Code != 0 {
			logs.Errorf("failed to query crp order change info, err: %v, crp_trace: %s, rid: %s",
				rst.Error.Message, rst.TraceId, kt.Rid)
			return nil, fmt.Errorf("failed to query crp order change info, err: %s", rst.Error.Message)
		}

		for _, ele := range rst.Result.Data {
			one, err := convOrderChangeInfoFromCrpRespItem(kt, orderID, ele)
			if err != nil {
				logs.Warnf("failed to convert crp order change info, err: %v, crp_result: %+v, rid: %s", err,
					*ele, kt.Rid)
				continue
			}

			result = append(result, one)
		}

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}

// convOrderChangeInfoFromCrpRespItem convert order change info from crp response item.
func convOrderChangeInfoFromCrpRespItem(kt *kit.Kit, orderID string, item *cvmapi.PlanOrderChangeItem) (
	*ptypes.CrpOrderChangeInfo, error) {

	diskType, err := enumor.GetDiskTypeFromCrpName(item.DiskTypeName)
	if err != nil {
		logs.Errorf("failed to get disk type from crp diskTypeName, err: %v, diskTypeName: %s, rid: %s",
			err, item.DiskTypeName, kt.Rid)
		return nil, err
	}

	if err := item.PlanType.Validate(); err != nil {
		logs.Errorf("crp plan type invalid, err: %v, planType: %s, rid: %s", err, item.PlanType, kt.Rid)
		return nil, err
	}

	if err := item.ProjectName.ValidateResPlan(); err != nil {
		logs.Errorf("crp project name invalid, err: %v, project name: %s, rid: %s", err, item.ProjectName, kt.Rid)
		return nil, err
	}

	resModeCode, err := item.ResourceMode.Code()
	if err != nil {
		logs.Errorf("crp resource mode invalid, err: %v, resource mode: %s, rid: %s", err, item.ResourceMode, kt.Rid)
		return nil, err
	}

	resType := enumor.DemandResTypeCVM
	// 机型为空时为CBS类型变更，TODO 暂不支持CBS类型的变更，没有机型无法和本地数据进行匹配
	if item.InstanceModel == "" {
		resType = enumor.DemandResTypeCBS
		return nil, errors.New("cbs type is not support")
	}

	return &ptypes.CrpOrderChangeInfo{
		OrderID:         orderID,
		VirtualDeptName: item.DeptName,
		PlanProductName: item.ProductName,
		OpProductName:   item.ProductName,
		ExpectTime:      item.UseTime,
		ReturnPlanTime:  item.ReturnPlanTime,
		ObsProject:      item.ProjectName,
		DemandResType:   resType,
		ResMode:         resModeCode,
		PlanType:        item.PlanType.GetCode(),
		RegionName:      item.CityName,
		ZoneName:        item.ZoneName,
		TechnicalClass:  item.TechnicalClass,
		DeviceFamily:    item.InstanceFamily,
		DeviceClass:     item.InstanceType,
		DeviceType:      item.InstanceModel,
		CoreType:        item.CoreTypeName,
		DiskType:        diskType,
		DiskTypeName:    item.DiskTypeName,
		DiskIO:          item.InstanceIO,
		ChangeOs:        item.ChangeCvmAmount,
		ChangeCpuCore:   item.ChangeCoreAmount,
		ChangeMemory:    item.ChangeRamAmount,
		ChangeDiskSize:  item.ChangedDiskAmount,
	}, nil
}

// matchDemandTableByDemandKey match demand table by demand key. expect only one.
func matchDemandTableByDemandKey(kt *kit.Kit, demands []rpd.ResPlanDemandTable,
	key ptypes.ResPlanDemandKey) []rpd.ResPlanDemandTable {

	res := make([]rpd.ResPlanDemandTable, 0)

	for _, demand := range demands {
		expectDateStr, err := times.TransTimeStrWithLayout(strconv.Itoa(demand.ExpectTime), constant.DateLayoutCompact,
			constant.DateLayout)
		if err != nil {
			logs.Warnf("failed to parse demand expect time, err: %v, expect_time: %d, rid: %s", err,
				demand.ExpectTime, kt.Rid)
			continue
		}

		demandKey := ptypes.ResPlanDemandKey{
			BkBizID:       demand.BkBizID,
			DemandClass:   demand.DemandClass,
			DemandResType: demand.DemandResType,
			ResMode:       demand.ResMode,
			ObsProject:    demand.ObsProject,
			ExpectTime:    expectDateStr,
			PlanType:      demand.PlanType,
			RegionID:      demand.RegionID,
			ZoneID:        demand.ZoneID,
			DeviceType:    demand.DeviceType,
			DiskType:      demand.DiskType,
			DiskIO:        demand.DiskIO,
		}

		if demandKey == key {
			res = append(res, demand)
			break
		}
	}

	return res
}

// DeductResPlanDemandAggregate 根据聚合Key筛选一组res plan demand调减
// 只有 demandChange.ChangeCpuCore <= 0 时才可以使用这个函数，> 0 时直接走新增
// 不考虑demand key完全一致，优先选择加锁的
func (d *Dispatcher) DeductResPlanDemandAggregate(kt *kit.Kit, aggregateDemands []rpd.ResPlanDemandTable,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource,
	demandChange *ptypes.CrpOrderChangeInfo) error {

	if demandChange.ChangeCpuCore > 0 {
		return fmt.Errorf("demand deduct cpu core should be <= 0, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
	}

	// 优先选择加锁的
	unlockedDemands := make([]rpd.ResPlanDemandTable, 0, len(aggregateDemands))
	for _, demand := range aggregateDemands {
		if demandChange.ChangeCpuCore == 0 {
			break
		}
		if cvt.PtrToVal(demand.Locked) != enumor.CrpDemandLocked {
			unlockedDemands = append(unlockedDemands, demand)
			continue
		}

		deductCpuNum := prepareDeductResPlanDemand(demand, demandChange.ChangeCpuCore, needUpdateDemands,
			demandUpdateRes)
		// deductCpuNum是正数，demandChange.ChangeCpuCore是负数
		demandChange.ChangeCpuCore += deductCpuNum
	}

	// 如果还没有扣减完，继续扣减未加锁的，但是这是预期外的，需要给一个警告
	for _, demand := range unlockedDemands {
		if demandChange.ChangeCpuCore == 0 {
			break
		}
		logs.Warnf("demand update occur on unlocked data and may result in unexpected results, update demand: %s, rid: %s",
			demand.ID, kt.Rid)

		deductCpuNum := prepareDeductResPlanDemand(demand, demandChange.ChangeCpuCore, needUpdateDemands,
			demandUpdateRes)
		// deductCpuNum是正数，demandChange.ChangeCpuCore是负数
		demandChange.ChangeCpuCore += deductCpuNum
	}

	// 现有的量不够调减
	if demandChange.ChangeCpuCore < 0 {
		logs.Errorf("failed to update res plan demand, remained demand is not enough to deduct, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
		return errors.New("remained demand is not enough to deduct")
	}

	return nil
}

// prepareDeductResPlanDemand prepare the needUpdateDemands map to deduct res plan demand.
// return is the cpu core will be deducted, always > 0.
func prepareDeductResPlanDemand(updateDemand rpd.ResPlanDemandTable, changeCpuCore int64,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource) int64 {

	if _, ok := needUpdateDemands[updateDemand.ID]; !ok {
		needUpdateDemands[updateDemand.ID] = &updateDemand
		demandUpdateRes[updateDemand.ID] = &ptypes.DemandResource{
			DeviceType: updateDemand.DeviceType,
			CpuCore:    0,
			DiskSize:   0,
		}
	}
	demandCpuCore := cvt.PtrToVal(needUpdateDemands[updateDemand.ID].CpuCore)

	changeCpuCoreAbs := math.Abs(float64(changeCpuCore))
	deductCpuNum := int64(math.Min(changeCpuCoreAbs, float64(demandCpuCore)))

	// 对已有数据的变更只记录变更量，全部变更处理完成后再统一拼装更新请求
	// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
	remainedCpuCore := demandCpuCore - deductCpuNum
	needUpdateDemands[updateDemand.ID].CpuCore = &remainedCpuCore
	demandUpdateRes[updateDemand.ID].CpuCore -= deductCpuNum

	return deductCpuNum
}

func newResPlanDemandCreateReq(ticket *ApplyTicketCtx, demand *ptypes.CrpOrderChangeInfo,
	expectTime time.Time) rpproto.ResPlanDemandCreateReq {

	osChange := demand.ChangeOs
	cpuCoreChange := demand.ChangeCpuCore
	memoryChange := demand.ChangeMemory
	diskSizeChange := demand.ChangeDiskSize
	return rpproto.ResPlanDemandCreateReq{
		BkBizID:         ticket.BkBizID,
		BkBizName:       ticket.BkBizName,
		OpProductID:     ticket.OpProductID,
		OpProductName:   ticket.OpProductName,
		PlanProductID:   ticket.PlanProductID,
		PlanProductName: ticket.PlanProductName,
		VirtualDeptID:   ticket.VirtualDeptID,
		VirtualDeptName: ticket.VirtualDeptName,
		DemandClass:     ticket.DemandClass,
		DemandResType:   demand.DemandResType,
		ResMode:         demand.ResMode,
		ObsProject:      demand.ObsProject,
		ExpectTime:      expectTime.Format(constant.DateLayout),
		PlanType:        demand.PlanType,
		AreaName:        demand.AreaName,
		RegionID:        demand.RegionID,
		RegionName:      demand.RegionName,
		ZoneID:          demand.ZoneID,
		ZoneName:        demand.ZoneName,
		TechnicalClass:  demand.TechnicalClass,
		DeviceFamily:    demand.DeviceFamily,
		DeviceClass:     demand.DeviceClass,
		DeviceType:      demand.DeviceType,
		CoreType:        demand.CoreType,
		DiskType:        demand.DiskType,
		DiskTypeName:    demand.DiskTypeName,
		OS:              &osChange,
		CpuCore:         &cpuCoreChange,
		Memory:          &memoryChange,
		DiskSize:        &diskSizeChange,
		DiskIO:          demand.DiskIO,
	}
}

func convCreateResPlanDemandReqs(kt *kit.Kit, ticket *ApplyTicketCtx, demand *ptypes.CrpOrderChangeInfo) (
	rpproto.ResPlanDemandCreateReq, rpproto.DemandChangelogCreate, error) {

	expectTimeFormat, err := time.Parse(constant.DateLayout, demand.ExpectTime)
	if err != nil {
		logs.Errorf("failed to parse expect time, err: %v, expect_time: %s, rid: %s", err, demand.ExpectTime,
			kt.Rid)
		return rpproto.ResPlanDemandCreateReq{}, rpproto.DemandChangelogCreate{}, err
	}

	osChange := demand.ChangeOs
	cpuCoreChange := demand.ChangeCpuCore
	memoryChange := demand.ChangeMemory
	diskSizeChange := demand.ChangeDiskSize
	createReq := newResPlanDemandCreateReq(ticket, demand, expectTimeFormat)

	if demand.ObsProject == enumor.ObsProjectShortLease {
		returnPlanTimeFormat, err := time.Parse(constant.DateLayout, demand.ReturnPlanTime)
		if err != nil {
			logs.Errorf("failed to parse return plan time, err: %v, return_plan_time: %s, rid: %s", err,
				demand.ReturnPlanTime, kt.Rid)
			return rpproto.ResPlanDemandCreateReq{}, rpproto.DemandChangelogCreate{}, err
		}
		createReq.ReturnPlanTime = returnPlanTimeFormat.Format(constant.DateLayout)
	}
	createReq.DiskType = createReq.DiskType.GetWithDefault()
	createReq.DiskTypeName = createReq.DiskType.Name()
	if kt.User == constant.BackendOperationUserKey {
		createReq.Creator = ticket.Applicant
	}

	// 更新日志
	logCreateReq := rpproto.DemandChangelogCreate{
		// DemandID 需要在demand创建后补充
		DemandID:       "",
		TicketID:       ticket.ID,
		CrpOrderID:     ticket.CrpSN,
		SuborderID:     "",
		Type:           enumor.DemandChangelogTypeAppend,
		ExpectTime:     expectTimeFormat.Format(constant.DateLayout),
		ObsProject:     demand.ObsProject,
		RegionName:     demand.RegionName,
		ZoneName:       demand.ZoneName,
		DeviceType:     demand.DeviceType,
		OSChange:       &osChange,
		CpuCoreChange:  &cpuCoreChange,
		MemoryChange:   &memoryChange,
		DiskSizeChange: &diskSizeChange,
		Remark:         ticket.Remark,
	}

	return createReq, logCreateReq, nil
}

// ApplyResPlanDemandChangeFromRPTickets 从给定的预测单据中将CRP的预测变更生效，用于从历史单据批量修复历史数据
func (d *Dispatcher) ApplyResPlanDemandChangeFromRPTickets(kt *kit.Kit, tickets []rtypes.RPTicketWithStatus) error {
	for _, ticket := range tickets {
		// 只看已通过的订单
		if ticket.Status != enumor.RPTicketStatusDone {
			continue
		}

		var demands rpt.ResPlanDemands
		if err := json.Unmarshal([]byte(ticket.Demands), &demands); err != nil {
			logs.Errorf("failed to unmarshal demands, err: %v, rid: %s", err, kt.Rid)
			return err

		}

		ticketInfo := &ptypes.TicketInfo{
			ID:               ticket.ID,
			Type:             ticket.Type,
			Applicant:        ticket.Applicant,
			BkBizID:          ticket.BkBizID,
			BkBizName:        ticket.BkBizName,
			OpProductID:      ticket.OpProductID,
			OpProductName:    ticket.OpProductName,
			PlanProductID:    ticket.PlanProductID,
			PlanProductName:  ticket.PlanProductName,
			VirtualDeptID:    ticket.VirtualDeptID,
			VirtualDeptName:  ticket.VirtualDeptName,
			DemandClass:      ticket.DemandClass,
			OriginalCpuCore:  ticket.OriginalCpuCore,
			OriginalMemory:   ticket.OriginalMemory,
			OriginalDiskSize: ticket.OriginalDiskSize,
			UpdatedCpuCore:   ticket.UpdatedCpuCore,
			UpdatedMemory:    ticket.UpdatedMemory,
			UpdatedDiskSize:  ticket.UpdatedDiskSize,
			Remark:           ticket.Remark,
			Demands:          demands,
			SubmittedAt:      ticket.SubmittedAt,
			Status:           ticket.Status,
			ItsmSN:           ticket.ItsmSN,
			CrpSN:            ticket.CrpSN,
		}

		if err := d.queryAndApplyResPlanDemandChange(kt, ticketInfo); err != nil {
			logs.Errorf("failed to apply res plan demand change, err: %v, ticket_info: %+v, rid: %s", err,
				*ticketInfo, kt.Rid)
			return err
		}

		logs.Infof("apply res plan demand change from ticket, bk_biz_id: %d, ticket_id: %s, rid: %s",
			ticket.BkBizID, ticket.ID, kt.Rid)
	}

	return nil
}
