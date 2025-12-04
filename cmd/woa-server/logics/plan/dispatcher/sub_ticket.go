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
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/logics/plan/splitter"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
)

// listAndWatchTickets list and watch tickets
func (d *Dispatcher) listAndWatchSubTickets() error {
	logs.Infof("ready to list and watch sub tickets")
	if !d.sd.IsMaster() {
		// pop all pending orders
		d.subTicketQueue.Clear()
		return nil
	}

	// list pending orders
	kt := core.NewBackendKit()
	pendingTkIDs, err := d.listAllPendingSubTickets(kt)
	if err != nil {
		logs.Errorf("failed to list pending resource plan tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// enqueue pending orders
	for _, tkID := range pendingTkIDs {
		d.subTicketQueue.Enqueue(tkID)
	}

	return nil
}

// listAllPendingSubTickets list all pending sub tickets
func (d *Dispatcher) listAllPendingSubTickets(kt *kit.Kit) ([]string, error) {
	listOpt := &rpproto.ResPlanSubTicketListReq{
		ListReq: core.ListReq{
			Fields: []string{"id"},
			Filter: tools.EqualExpression("status", enumor.RPSubTicketStatusAuditing),
			Page:   core.NewDefaultBasePage(),
		},
	}

	subTicketIDs := make([]string, 0)
	for {
		tkRst, err := d.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, listOpt)
		if err != nil {
			logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, ticket := range tkRst.Details {
			subTicketIDs = append(subTicketIDs, ticket.ID)
		}

		if len(tkRst.Details) < int(listOpt.Page.Limit) {
			break
		}
		listOpt.Page.Start += uint32(listOpt.Page.Limit)
	}

	return subTicketIDs, nil
}

func (d *Dispatcher) dealSubTicket() error {
	// only master node handle plan tickets.
	if !d.sd.IsMaster() {
		return nil
	}

	// get one ticket from the work queue
	tkID, ok := d.subTicketQueue.Pop()
	if !ok {
		return nil
	}

	// check the status of the ticket
	kt := core.NewBackendKit()
	logs.Infof("ready to handle sub ticket %s, rid: %s", tkID, kt.Rid)
	tkInfo, err := d.resFetcher.GetSubTicketInfo(kt, tkID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
		return err
	}

	// Only tickets in the auditing status are processed
	// (`init` status is unused, `waiting` tickets will be handled in ticket flow)
	if tkInfo.Status != enumor.RPSubTicketStatusAuditing {
		logs.Warnf("need not handle sub ticket for its status %s is finished, id: %s, rid: %s", tkInfo.Status,
			tkID, kt.Rid)
		return nil
	}

	switch tkInfo.Stage {
	case enumor.RPSubTicketStageInit:
		// RPSubTicketStageInit is unused
		logs.Errorf("unsupported sub ticket stage %s, id: %s, rid: %s", tkInfo.Stage, tkID, kt.Rid)
		return fmt.Errorf("unsupported sub ticket stage(%s)", tkInfo.Stage)
	case enumor.RPSubTicketStageAdminAudit:
		// 确认管理员审批状态，审批通过时尝试创建CRP单据，创建成功时进入CRP审批阶段
		err = d.checkAdminAuditStatus(kt, tkInfo)
		if err != nil {
			logs.Errorf("%s: failed to handle sub ticket for its stage is admin audit, err: %v, id: %s, rid: %s",
				constant.ResPlanTicketWatchFailed, err, tkID, kt.Rid)
			return err
		}
	case enumor.RPSubTicketStageCRPAudit:
		return d.checkCrpTicket(kt, tkInfo)
	default:
		logs.Errorf("failed to handle ticket for its stage %s is invalid, id: %s, rid: %s", tkInfo.Stage,
			tkID, kt.Rid)
		return fmt.Errorf("failed to handle sub ticket(%s) for its stage(%s) is invalid", tkID, tkInfo.Stage)
	}

	return nil
}

// checkCrpTicket check crp ticket status
func (d *Dispatcher) checkCrpTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) error {
	logs.Infof("ready to check crp flow, sn: %s, id: %s, rid: %s", subTicket.CrpSN, subTicket.ID, kt.Rid)

	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderQueryMethod,
		},
		Params: &cvmapi.QueryPlanOrderParam{
			OrderIds: []string{subTicket.CrpSN},
		},
	}

	var planItem *cvmapi.QueryPlanOrderRst
	rangeMS := [2]uint{CreateCrpTicketDefaultRetryDelayMinMS, CreateCrpTicketDefaultRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for {
		resp, err := d.crpCli.QueryPlanOrder(kt.Ctx, nil, req)
		if err != nil {
			logs.Errorf("failed to add cvm & cbs plan order, err: %v, sub_ticket_id: %s, rid: %s", err, subTicket.ID,
				kt.Rid)
			return err
		}

		if resp.Error.Code != 0 {
			logs.Errorf("%s: failed to query crp plan order, code: %d, msg: %s, crp_sn: %s, crp_trace: %s, rid: %s",
				constant.ResPlanTicketWatchFailed, resp.Error.Code, resp.Error.Message, subTicket.CrpSN, resp.TraceId, kt.Rid)
			return fmt.Errorf("failed to query crp plan order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		}
		if resp.Result == nil {
			logs.Errorf("%s: failed to query crp plan order, for result is empty, crp_sn: %s, rid: %s",
				constant.ResPlanTicketWatchFailed, subTicket.CrpSN, kt.Rid)
			return errors.New("failed to query crp plan order, for result is empty")
		}
		var ok bool
		planItem, ok = resp.Result[subTicket.CrpSN]
		if !ok {
			logs.Errorf("%s: query crp plan order return no result by sn: %s, rid: %s",
				constant.ResPlanTicketWatchFailed, subTicket.CrpSN, kt.Rid)
			return fmt.Errorf("query crp plan order return no result by sn: %s", subTicket.CrpSN)
		}
		// CRP返回状态码为： 1 追加单， 2 调整单， 3 订单不存在， 4 其它错误（只有1 和 2 是正确的）
		if planItem.Code == 4 {
			// 进行重试
			if policy.RetryCount()+1 < CreateCrpTicketDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Warnf(
					"call crp rate limit, will sleep for retry, retry count: %d, err: %v, crp_trace: %s, rid: %s",
					policy.RetryCount(), resp.Error, resp.TraceId, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		if planItem.Code != 1 && planItem.Code != 2 && planItem.Code != 6 {
			logs.Errorf("%s: failed to query crp plan order, order status is incorrect, code: %d, data: %+v,"+
				" crp_trace: %s, rid: %s",
				constant.ResPlanTicketWatchFailed, planItem.Code, planItem.Data, resp.TraceId, kt.Rid)
			return fmt.Errorf("crp plan order status is incorrect, code: %d, sn: %s", planItem.Code, subTicket.CrpSN)
		}
		// 其他情况都跳过
		break
	}

	update := &rpproto.ResPlanSubTicketUpdateReq{
		ID:     subTicket.ID,
		Status: enumor.RPSubTicketStatusAuditing,
		CrpSN:  subTicket.CrpSN,
		CrpURL: subTicket.CrpURL,
	}

	switch planItem.Data.BaseInfo.Status {
	case cvmapi.PlanOrderStatusRejected:
		update.Status = enumor.RPSubTicketStatusRejected
	case cvmapi.PlanOrderStatusApproved:
		// 更新子单状态到成功，等待其他子单进入终态
		update.Status = enumor.RPSubTicketStatusDone
	default:
		return d.checkSubTicketTimeout(kt, subTicket)
	}
	if err := d.updateSubTicket(kt, subTicket, update); err != nil {
		logs.Errorf("failed to update resource plan sub ticket, err: %v, id: %s, update: %+v, rid: %s", err,
			subTicket.ID, update, kt.Rid)
		return err
	}
	return nil
}

// checkCrpTicket check crp ticket to update ticket status
func (d *Dispatcher) checkAdminAuditStatus(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) error {

	switch subTicket.AdminAuditStatus {
	case enumor.RPAdminAuditStatusSkip, enumor.RPAdminAuditStatusDone:
	case enumor.RPAdminAuditStatusAuditing:
		logs.Infof("sub ticket is in admin auditing, id: %s, rid: %s", subTicket.ID, kt.Rid)
		return nil
	case enumor.RPAdminAuditStatusRejected:
		// 理论上当admin审批状态为reject时，ticket已经处于终态，不应进入到 checkAdminAuditStatus 中
		logs.Errorf("invalid sub ticket status, admin audit status is rejected but still in auditing, id: %s, "+
			"ticket status: %s, rid: %s", subTicket.ID, subTicket.Status, kt.Rid)
		return errors.New("admin audit status is rejected but still in auditing")
	default:
		logs.Errorf("invalid admin audit status: %s, id: %s, rid: %s", subTicket.AdminAuditStatus, subTicket.ID,
			kt.Rid)
		return fmt.Errorf("invalid admin audit status: %s", subTicket.AdminAuditStatus)
	}

	// 创建CRP单据
	err := d.createCrpTicket(kt, subTicket)
	if err != nil {
		logs.Errorf("failed to create crp ticket, err: %v, id: %s, rid: %s", err, subTicket.ID, kt.Rid)
		return err
	}
	return nil
}

// checkSubTicket check sub ticket to update ticket status
func (d *Dispatcher) checkSubTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	logs.Infof("ready to check sub ticket status, id: %s, rid: %s", ticket.ID, kt.Rid)

	subTickets := make([]ptypes.ListResPlanSubTicketItem, 0)
	listReq := &ptypes.ListResPlanSubTicketReq{
		TicketID: ticket.ID,
		Page:     core.NewDefaultBasePage(),
	}
	for {
		rst, err := d.resFetcher.ListResPlanSubTicket(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list res plan sub ticket, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
			return err
		}

		subTickets = append(subTickets, rst.Details...)
		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	// 统计子单状态
	allDone, allFailed, hasFailed, allRejected, hasRejected := true, true, false, true, false
	for _, subTicket := range subTickets {
		// 有子单处于非终态，继续等待
		if subTicket.Status.IsUnfinished() {
			logs.Infof("sub ticket %s is not finished, id: %s, rid: %s", subTicket.ID, ticket.ID, kt.Rid)
			return nil
		}
		if subTicket.Status == enumor.RPSubTicketStatusInvalid {
			continue
		}

		// 拒绝状态为最低优先级，只有全部子单均为审批拒绝时才更新主单状态为审批拒绝
		allRejected = allRejected && subTicket.Status == enumor.RPSubTicketStatusRejected
		hasRejected = hasRejected || subTicket.Status == enumor.RPSubTicketStatusRejected
		// 审批拒绝不认为是失败（失败为非终态，不会触发结果汇总生效）
		allDone = allDone && subTicket.Status == enumor.RPSubTicketStatusDone
		allFailed = allFailed && subTicket.Status == enumor.RPSubTicketStatusFailed
		hasFailed = hasFailed || subTicket.Status == enumor.RPSubTicketStatusFailed
	}
	// 根据统计结果更新主单状态(失败优先级高于审批拒绝，当同时存在失败和审批拒绝时，以失败为最终结果)
	ticketStatus := enumor.RPTicketStatusAuditing
	switch {
	case allFailed:
		ticketStatus = enumor.RPTicketStatusFailed
	case hasFailed:
		ticketStatus = enumor.RPTicketStatusPartialFailed
	case allRejected:
		ticketStatus = enumor.RPTicketStatusRejected
	case hasRejected:
		ticketStatus = enumor.RPTicketStatusPartialRejected
	case allDone:
		ticketStatus = enumor.RPTicketStatusDone
	default:
		logs.Warnf("invalid sub ticket status, allDone: %t, allFailed: %t, hasFailed: %t, "+
			"allRejected: %t, hasRejected: %t, id: %s, rid: %s", allDone, allFailed, hasFailed, allRejected,
			hasRejected, ticket.ID, kt.Rid)
	}

	// 单据成功完结，且不存在失败的子单，需要将所有子单的结果汇总生效
	if !hasFailed {
		if err := d.FinishAuditFlow(kt, ticket); err != nil {
			logs.Errorf("failed to finish audit flow, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
			return err
		}
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   ticketStatus,
	}
	if err := d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// checkSubTicketTimeout check sub ticket timeout
func (d *Dispatcher) checkSubTicketTimeout(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) error {
	submitTime, err := time.Parse(constant.TimeStdFormat, subTicket.SubmittedAt)
	if err != nil {
		logs.Errorf("failed to parse sub ticket submit time %s, err: %v, rid: %s", subTicket.SubmittedAt,
			err, kt.Rid)
		return err
	}

	// set timeout as 28 days
	if time.Now().Before(submitTime.AddDate(0, 0, enumor.AuditFlowTimeoutDay)) {
		return nil
	}

	return d.updateSubTicketStatusFailed(kt, subTicket, "audit flow timeout")
}

// startWaitingSubTicket update waiting sub ticket status to auditing.
func (d *Dispatcher) startWaitingSubTicket(kt *kit.Kit, ticketID string, applicant string) error {
	updateReq := &rpproto.ResPlanSubTicketStatusUpdateReq{
		TicketID: ticketID,
		Source:   enumor.RPSubTicketStatusWaiting,
		Target:   enumor.RPSubTicketStatusAuditing,
	}
	// 后台任务，使用提单人作为子单的更新人
	kt.User = applicant
	err := d.client.DataService().Global.ResourcePlan.UpdateResPlanSubTicketStatusCAS(kt, updateReq)
	if err != nil {
		logs.Errorf("failed to update res plan sub ticket status %s to %s, err: %v, ticket id: %s, rid: %s",
			updateReq.Source, updateReq.Target, err, ticketID, kt.Rid)
		return err
	}
	return nil
}

// createSubTicket create res plan sub ticket.
func (d *Dispatcher) createSubTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	if ticket == nil {
		logs.Errorf("failed to create sub ticket, ticket is nil, rid: %s", kt.Rid)
		return errors.New("ticket is nil")
	}

	// 已存在子单时不再尝试创建
	// TODO 这种方式后台只会拆分一次子单，如果需要重试，需要通过接口同步实现
	listReq := &ptypes.ListResPlanSubTicketReq{
		TicketID: ticket.ID,
		Page:     core.NewCountPage(),
	}
	subTickets, err := d.resFetcher.ListResPlanSubTicket(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list res plan sub ticket, err: %v, ticket id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}
	if subTickets.Count > 0 {
		logs.Infof("sub ticket already exists, ticket id: %s, sub ticket number: %d, rid: %s", ticket.ID,
			subTickets.Count, kt.Rid)
		return nil
	}

	splitHelper := splitter.New(d.dao, d.client, d.crpCli, d.resFetcher, d.deviceTypesMap)
	switch ticket.Type {
	case enumor.RPTicketTypeDelete, enumor.RPTicketTypeAutomaticTransfer:
		return splitHelper.SplitDeleteTicket(kt, ticket.ID, ticket.Demands, ticket.PlanProductName,
			ticket.OpProductName)
	case enumor.RPTicketTypeAdd:
		return splitHelper.SplitAddTicket(kt, ticket.ID, ticket.Demands)
	case enumor.RPTicketTypeAdjust:
		return splitHelper.SplitAdjustTicket(kt, ticket.ID, ticket.Demands, ticket.PlanProductName,
			ticket.OpProductName)
	default:
		logs.Errorf("unsupported res plan ticket type, type: %s, rid: %s", ticket.Type, kt.Rid)
		return fmt.Errorf("unsupported res plan ticket type, type: %s", ticket.Type)
	}
}

// updateSubTicket update res plan sub ticket.
func (d *Dispatcher) updateSubTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	update *rpproto.ResPlanSubTicketUpdateReq) error {

	// 后台任务，使用提单人作为子单的更新人
	kt.User = subTicket.Applicant
	updateReq := &rpproto.ResPlanSubTicketBatchUpdateReq{
		SubTickets: []rpproto.ResPlanSubTicketUpdateReq{cvt.PtrToVal(update)},
	}
	err := d.client.DataService().Global.ResourcePlan.BatchUpdateResPlanSubTicket(kt, updateReq)
	if err != nil {
		logs.Errorf("failed to update res plan sub ticket, err: %v, sub ticket id: %s, rid: %s", err, subTicket.ID,
			kt.Rid)
		return err
	}
	return nil
}

// updateSubTicketStatus update ticket status.
func (d *Dispatcher) updateSubTicketStatus(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	target enumor.RPSubTicketStatus, msg *string) error {

	updateReq := &rpproto.ResPlanSubTicketStatusUpdateReq{
		IDs:      []string{subTicket.ID},
		TicketID: subTicket.ParentTicketID,
		Source:   subTicket.Status,
		Target:   target,
		Message:  msg,
	}
	// 后台任务，使用提单人作为子单的更新人
	kt.User = subTicket.Applicant
	err := d.client.DataService().Global.ResourcePlan.UpdateResPlanSubTicketStatusCAS(kt, updateReq)
	if err != nil {
		logs.Errorf("failed to update res plan sub ticket status %s to %s, err: %v, sub ticket id: %s, rid: %s",
			updateReq.Source, updateReq.Target, err, subTicket.ID, kt.Rid)
		return err
	}
	return nil
}

// updateSubTicketStatusFailed update ticket status to failed.
func (d *Dispatcher) updateSubTicketStatusFailed(kt *kit.Kit, subTicket *ptypes.SubTicketInfo, msg string) error {
	if err := d.updateSubTicketStatus(kt, subTicket, enumor.RPSubTicketStatusFailed, &msg); err != nil {
		logs.Errorf("failed to update resource plan sub ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
