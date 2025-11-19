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
	"context"
	"fmt"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	resplan "hcm/pkg/dal/dao/types/resource-plan"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/slice"
)

// autoTransferNearExpireDemand 自动转移预测
func (c *Controller) autoTransferNearExpireDemand(ctx context.Context, loc *time.Location) {
	now := time.Now().In(loc)

	nextRunTime := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, loc)
	if now.After(nextRunTime) {
		nextRunTime = nextRunTime.Add(time.Hour * 24)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 等待到下一个检查时间
		time.Sleep(time.Until(nextRunTime))
		runTime := nextRunTime
		logs.Infof("auto transfer near expired demand start, run_time: %s",
			runTime.Format(constant.DateTimeLayout))

		kt := core.NewBackendKit()
		isNeed, err := c.needToTransferDemand(kt, runTime)
		if err != nil {
			logs.Errorf("%s: failed to check if need to push expire notifications, err: %v, time: %s, rid: %s",
				constant.ResPlanNearExpiredDemandTransferFailed, err, runTime.Format(constant.DateTimeLayout), kt.Rid)
			continue
		}
		if !isNeed {
			logs.Infof("auto transfer near expired demand finished, run_time: %s, status: skip",
				runTime.Format(constant.DateTimeLayout))
			// 计算下一个检查时间
			nextRunTime = runTime.Add(time.Hour * 24)
			continue
		}
		err = c.autoTransferResPlanDemands(kt, runTime)
		if err != nil {
			logs.Errorf("%s: failed to renew res plan demands, err: %v, time: %s, rid: %s",
				constant.ResPlanNearExpiredDemandTransferFailed, err, runTime.Format(constant.DateTimeLayout), kt.Rid)
			continue
		}

		logs.Infof("auto transfer near expired demand finished, run_time: %s, status: success",
			runTime.Format(constant.DateTimeLayout))

		// 计算下一个检查时间
		nextRunTime = runTime.Add(time.Hour * 24)
	}
}

// needToTransferDemand 判断该日期是否需要转移需求, 截止日当天需要进行预测转移
func (c *Controller) needToTransferDemand(kt *kit.Kit, t time.Time) (bool, error) {
	monthRange, err := c.demandTime.GetDemandDateRangeInMonth(kt, t)
	if err != nil {
		return false, err
	}
	monthEnd, err := time.Parse(constant.DateLayout, monthRange.End)
	if err != nil {
		logs.Errorf("failed to parse month range end, err: %v, month_range: %v, rid: %s", err, monthRange, kt.Rid)
		return false, err
	}
	return monthEnd.Day() == t.Day(), nil
}

// listNearExpiredDemands 获取即将过期的预测且需求类型为 CVM
func (c *Controller) listNearExpiredDemands(kt *kit.Kit, t time.Time) ([]*ptypes.ListResPlanDemandItem, error) {
	monthRange, err := c.demandTime.GetDemandDateRangeInMonth(kt, t)
	if err != nil {
		return nil, err
	}
	demandDetails := make([]*ptypes.ListResPlanDemandItem, 0)
	listReq := &ptypes.ListResPlanDemandReq{
		ExpiringOnly:    true,
		ExpectTimeRange: &monthRange,
		Page:            core.NewDefaultBasePage(),
	}
	for {
		rst, err := c.ListResPlanDemandAndOverview(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list res plan demand and overview, err: %v, expect_range: %+v, rid: %s", err,
				monthRange, kt.Rid)
			return nil, err
		}

		demandDetails = append(demandDetails, rst.Details...)
		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}
	return demandDetails, nil
}

func (c *Controller) autoTransferResPlanDemands(kt *kit.Kit, runTime time.Time) error {
	demands, err := c.listNearExpiredDemands(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to list near expired demands, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 过滤掉变更中
	demands = slice.Filter(demands, func(item *ptypes.ListResPlanDemandItem) bool {
		if item.Status != enumor.DemandStatusCanApply {
			logs.Infof("skip demand, status: %s, rid: %s", item.Status, kt.Rid)
			return false
		}
		return true
	})
	bizDemands := classifier.ClassifySlice(demands, func(item *ptypes.ListResPlanDemandItem) int64 {
		return item.BkBizID
	})

	ticketIDs := make([]string, 0)
	for bizID, bizDemandList := range bizDemands {
		demandsByClass := classifier.ClassifySlice(bizDemandList,
			func(item *ptypes.ListResPlanDemandItem) enumor.DemandClass {
				return item.DemandClass
			})
		for demandClass, items := range demandsByClass {
			ticketID, err := c.autoTransferBizResPlanDemands(kt, bizID, items, demandClass)
			if err != nil {
				logs.Errorf("failed to renew biz res plan demands, err: %v, biz_id: %d, demand_class: %s, demands: %+v, rid: %s",
					err, bizID, demandClass, items, kt.Rid)
				continue
			}
			ticketIDs = append(ticketIDs, ticketID)
		}
	}

	timer := time.NewTimer(time.Until(runTime.Add(time.Hour * 3)))
	defer timer.Stop()
	// 下面的循环，发生报错也不退出循环，由timer控制循环退出
	for {
		select {
		case <-timer.C:
			return nil
		default:
		}
		logs.Infof("auto transfer near expired demand list tickets by ids, ticket_ids: %+v, rid: %s", ticketIDs, kt.Rid)

		if len(ticketIDs) == 0 {
			return nil
		}
		tickets, err := c.listTicketsByIDs(kt, ticketIDs)
		if err != nil {
			logs.Errorf("failed to list tickets by ids, err: %v, ticket_ids: %+v, rid: %s", err, ticketIDs, kt.Rid)
			time.Sleep(time.Second * 5)
			continue
		}

		listenIDs := make([]string, 0)
		for _, ticket := range tickets {
			logs.Infof("auto transfer near expired demand process ticket, ticket_id: %s, status: %s, rid: %s",
				ticket.ID, ticket.Status, kt.Rid)
			switch ticket.Status {
			case enumor.RPTicketStatusDone:
				// 终态，直接忽略
				logs.Infof("[auto transfer] ticket is done, skip, ticket_id: %s, rid: %s", ticket.ID, kt.Rid)
				continue
			case enumor.RPTicketStatusFailed, enumor.RPTicketStatusPartialFailed:
				// 失败，直接忽略，手动重试
				logs.Errorf("[auto transfer] ticket is failed, ticket_id: %s, rid: %s", ticket.ID, kt.Rid)
				continue
			case enumor.RPTicketStatusAuditing:
			default:
				logs.Errorf("[auto transfer] unexpected ticket status: %s, ticket_id: %s, rid: %s",
					ticket.Status, ticket.ID, kt.Rid)
			}
			listenIDs = append(listenIDs, ticket.ID)
		}
		ticketIDs = listenIDs
		time.Sleep(time.Second * 5)
	}
}

func (c *Controller) listTicketsByIDs(kt *kit.Kit, ticketIDs []string) ([]resplan.RPTicketWithStatus, error) {
	listOpt := &types.ListOption{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", ticketIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := c.dao.ResPlanTicket().ListWithStatus(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list biz resource plan ticket with status, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	return rst.Details, nil
}

func (c *Controller) autoTransferBizResPlanDemands(kt *kit.Kit, bkBizID int64,
	demands []*ptypes.ListResPlanDemandItem, demandClass enumor.DemandClass) (string, error) {

	bizOrgRel, err := c.bizLogics.GetBizOrgRel(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	resPlanDemands, lockedItems, demandIDs := c.constructAutoTransferDemands(demands)
	createTicketReq := &CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeAutomaticTransfer,
		DemandClass: demandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     resPlanDemands,
		Remark:      "",
	}
	// create cancel resource plan ticket.
	ticketID, err := c.CreateResPlanTicket(kt, createTicketReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	// lock all resource plan demand.
	for i := range lockedItems {
		lockedItems[i].TicketID = ticketID
	}
	lockReq := &rpproto.ResPlanDemandLockOpReq{
		LockedItems: lockedItems,
	}
	if err = c.client.DataService().Global.ResourcePlan.LockResPlanDemand(kt, lockReq); err != nil {
		logs.Errorf("failed to lock all resource plan demand, err: %v, demandIDs: %v, rid: %s", err,
			demandIDs, kt.Rid)
		return "", err
	}

	// defer is used to unlock all resource plan demand when some errors occur.
	defer func() {
		if err != nil {
			if tmpErr := c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, lockReq); tmpErr != nil {
				logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", tmpErr, kt.Rid)
			}
		}
	}()
	// 修改ticket的状态至audit,正式进入后台dispatch流程
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticketID,
		ItsmSN:   constant.ResPlanItsmAuditSkip,
		Status:   enumor.RPTicketStatusAuditing,
	}
	if err = c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return ticketID, nil
}

func (c *Controller) constructAutoTransferDemands(demands []*ptypes.ListResPlanDemandItem) (rpt.ResPlanDemands,
	[]rpproto.ResPlanDemandLockOpItem, []string) {

	result := make(rpt.ResPlanDemands, 0, len(demands))
	lockedItems := make([]rpproto.ResPlanDemandLockOpItem, 0, len(demands))
	demandIDs := make([]string, 0, len(demands))
	for _, demand := range demands {
		demandIDs = append(demandIDs, demand.DemandID)
		lockedItems = append(lockedItems, rpproto.ResPlanDemandLockOpItem{
			ID:            demand.DemandID,
			LockedCPUCore: demand.RemainedCpuCore,
		})
		result = append(result, rpt.ResPlanDemand{
			DemandClass: demand.DemandClass,
			Original: &rpt.OriginalRPDemandItem{
				DemandID:   demand.DemandID,
				ObsProject: demand.ObsProject,
				ExpectTime: demand.ExpectTime,
				ZoneID:     demand.ZoneID,
				ZoneName:   demand.ZoneName,
				RegionID:   demand.RegionID,
				RegionName: demand.RegionName,
				AreaID:     demand.AreaID,
				AreaName:   demand.AreaName,
				Cvm: rpt.Cvm{
					ResMode:        demand.ResMode.Name(),
					DeviceType:     demand.DeviceType,
					DeviceClass:    demand.DeviceClass,
					DeviceFamily:   demand.DeviceFamily,
					TechnicalClass: demand.TechnicalClass,
					CoreType:       string(demand.CoreType),
					Os:             tabletype.Decimal{Decimal: demand.RemainedOS},
					CpuCore:        demand.RemainedCpuCore,
					Memory:         demand.RemainedMemory,
				},
				Cbs: rpt.Cbs{
					DiskType:     demand.DiskType,
					DiskTypeName: demand.DiskTypeName,
					DiskIo:       demand.DiskIO,
					DiskSize:     demand.RemainedDiskSize,
				},
			},
		})
	}

	return result, lockedItems, demandIDs
}

// AutoTransferBizResPlanDemandByID 根据业务ID和需求ID自动转移预测
func (c *Controller) AutoTransferBizResPlanDemandByID(kt *kit.Kit, bkBizID int64, demandIDs []string) ([]string, error) {
	if len(demandIDs) == 0 {
		return nil, fmt.Errorf("demand_ids is empty")
	}

	// 根据 demandIDs 获取 demand 信息
	listReq := &ptypes.ListResPlanDemandReq{
		BkBizIDs:  []int64{bkBizID},
		DemandIDs: demandIDs,
		Page:      core.NewDefaultBasePage(),
	}

	rst, err := c.ListResPlanDemandAndOverview(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list res plan demand, err: %v, demand_ids: %v, biz_id: %d, rid: %s",
			err, demandIDs, bkBizID, kt.Rid)
		return nil, err
	}

	if len(rst.Details) == 0 {
		logs.Errorf("demands not found, demand_ids: %v, biz_id: %d, rid: %s", demandIDs, bkBizID, kt.Rid)
		return nil, fmt.Errorf("demands %v not found in biz %d", demandIDs, bkBizID)
	}

	// 验证所有需求是否属于指定的 bizID 且状态正确
	validDemands := make([]*ptypes.ListResPlanDemandItem, 0, len(rst.Details))
	for _, demand := range rst.Details {
		// 验证 demand 是否属于指定的 bizID
		if demand.BkBizID != bkBizID {
			logs.Errorf("demand does not belong to biz, demand_id: %s, demand_biz_id: %d, biz_id: %d, rid: %s",
				demand.DemandID, demand.BkBizID, bkBizID, kt.Rid)
			return nil, fmt.Errorf("demand %s does not belong to biz %d", demand.DemandID, bkBizID)
		}

		// 验证 demand 状态
		if demand.Status != enumor.DemandStatusCanApply {
			logs.Errorf("demand status is not can apply, demand_id: %s, status: %s, rid: %s",
				demand.DemandID, demand.Status, kt.Rid)
			return nil, fmt.Errorf("demand %s status is %s, not can apply", demand.DemandID, demand.Status)
		}

		validDemands = append(validDemands, demand)
	}

	// 按 demandClass 分组处理
	demandsByClass := classifier.ClassifySlice(validDemands,
		func(item *ptypes.ListResPlanDemandItem) enumor.DemandClass {
			return item.DemandClass
		})

	ticketIDs := make([]string, 0)
	for demandClass, items := range demandsByClass {
		ticketID, err := c.autoTransferBizResPlanDemands(kt, bkBizID, items, demandClass)
		if err != nil {
			logs.Errorf("failed to auto transfer biz res plan demands, err: %v, biz_id: %d, demand_class: %s,"+
				" demand_ids: %v, rid: %s",
				err, bkBizID, demandClass, demandIDs, kt.Rid)
			return nil, err
		}
		ticketIDs = append(ticketIDs, ticketID)
	}

	logs.Infof("successfully auto transfer biz res plan demands, demand_ids: %v, biz_id: %d, ticket_ids: %v, rid: %s",
		demandIDs, bkBizID, ticketIDs, kt.Rid)

	return ticketIDs, nil
}
