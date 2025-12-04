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
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/times"
)

// listAndWatchTickets list and watch tickets
func (d *Dispatcher) listAndWatchTickets() error {
	logs.Infof("ready to list and watch tickets")
	if !d.sd.IsMaster() {
		// pop all pending orders
		d.ticketQueue.Clear()
		return nil
	}

	// list pending orders
	kt := core.NewBackendKit()
	pendingTkIDs, err := d.listAllPendingTickets(kt)
	if err != nil {
		logs.Errorf("failed to list pending resource plan tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// enqueue pending orders
	for _, tkID := range pendingTkIDs {
		d.ticketQueue.Enqueue(tkID)
	}

	return nil
}

// listAllPendingTickets list all pending tickets
func (d *Dispatcher) listAllPendingTickets(kt *kit.Kit) ([]string, error) {
	// list tickets of recent 7 days.
	dr := &times.DateRange{
		Start: time.Now().AddDate(0, 0, -enumor.PendingTicketTraceDay).Format(constant.DateLayout),
		End:   time.Now().Format(constant.DateLayout),
	}

	drOpt, err := tools.DateRangeExpression("submitted_at", dr)
	if err != nil {
		return nil, err
	}

	// TODO: 当单据数量超过500时，可能会漏单据。这里改为分页查询
	recentOpt := &types.ListOption{
		Fields: []string{"id"},
		Filter: drOpt,
		Page:   core.NewDefaultBasePage(),
	}

	tkRst, err := d.dao.ResPlanTicket().List(kt, recentOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	recentTkIDs := make([]string, 0)
	for _, ticket := range tkRst.Details {
		recentTkIDs = append(recentTkIDs, ticket.ID)
	}

	// list tickets with auditing status
	auditOpt := &types.ListOption{
		Fields: []string{"ticket_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("ticket_id", recentTkIDs),
			tools.RuleEqual("status", enumor.RPTicketStatusAuditing),
		),
		Page: core.NewDefaultBasePage(),
	}

	statusRst, err := d.dao.ResPlanTicketStatus().List(kt, auditOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pendingTkIDs := make([]string, 0)
	for _, ticket := range statusRst.Details {
		pendingTkIDs = append(pendingTkIDs, ticket.TicketID)
	}

	return pendingTkIDs, nil
}

// dealTicket deal a ticket
func (d *Dispatcher) dealTicket() error {
	logs.Infof("ready to deal ticket")

	// only master node handle plan tickets.
	if !d.sd.IsMaster() {
		return nil
	}

	// get one ticket from the work queue
	tkID, ok := d.ticketQueue.Pop()
	if !ok {
		return nil
	}

	// check the status of the ticket
	kt := core.NewBackendKit()
	logs.Infof("ready to handle ticket %s, rid: %s", tkID, kt.Rid)
	tkInfo, err := d.resFetcher.GetTicketInfo(kt, tkID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
		return err
	}

	if tkInfo.Status != enumor.RPTicketStatusAuditing {
		logs.Warnf("need not handle ticket for its status %s != %s, id: %s, rid: %s", tkInfo.Status,
			enumor.RPTicketStatusAuditing, tkID, kt.Rid)
		return nil
	}

	if tkInfo.ItsmSN == "" {
		logs.Errorf("failed to handle ticket for itsm sn is empty, id: %s, rid: %s", tkID, kt.Rid)
		return errors.New("failed to handle ticket for itsm sn is empty")
	}

	checkSubTicket := false
	if tkInfo.ItsmSN == constant.ResPlanItsmAuditSkip {
		if err = d.createSubTicket(kt, tkInfo); err != nil {
			logs.Errorf("failed to create sub ticket, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
			return err
		}
		checkSubTicket = true
	} else {
		checkSubTicket, err = d.checkItsmTicket(kt, tkInfo)
		if err != nil {
			logs.Errorf("failed to check itsm ticket, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
			return err
		}
	}

	if checkSubTicket {
		// 子单拆分完成，更新所有waiting子单的状态到auditing
		err = d.startWaitingSubTicket(kt, tkID, tkInfo.Applicant)
		if err != nil {
			logs.Errorf("failed to start waiting sub ticket, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
			return err
		}

		// 检查子单状态，更新主单状态
		return d.checkSubTicket(kt, tkInfo)
	}
	return nil
}

func (d *Dispatcher) checkItsmTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) (bool, error) {
	logs.Infof("ready to check itsm flow, sn: %s, id: %s, rid: %s", ticket.ItsmSN, ticket.ID, kt.Rid)
	checkSubTicket := false
	resp, err := d.itsmCli.GetTicketStatus(kt, ticket.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return false, err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusRejected,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
	}

	switch resp.Data.CurrentStatus {
	case string(itsm.StatusTerminated):
		// rejected
		update.Status = enumor.RPTicketStatusRejected
	case string(itsm.StatusRevoked):
		// revoked
		update.Status = enumor.RPTicketStatusRevoked
	case string(itsm.StatusRunning):
		// check if CRP audit state
		if len(resp.Data.CurrentSteps) == 0 {
			return false, d.checkTicketTimeout(kt, ticket)
		}

		if resp.Data.CurrentSteps[0].StateId != d.crpAuditNode.ID {
			return false, d.checkTicketTimeout(kt, ticket)
		}

		// CRP audit state, create CRP ticket
		checkSubTicket = true
		return checkSubTicket, d.createSubTicket(kt, ticket)
	case string(itsm.StatusFinished):
		// ITSM单正常完结时，依然需要确认子单的状态
		checkSubTicket = true
		return checkSubTicket, d.createSubTicket(kt, ticket)
	default:
		return checkSubTicket, d.checkTicketTimeout(kt, ticket)
	}

	if err = d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return checkSubTicket, err
	}

	if update.Status != enumor.RPTicketStatusRejected && update.Status != enumor.RPTicketStatusRevoked {
		return checkSubTicket, nil
	}
	// 单据被拒需要释放资源
	return checkSubTicket, d.unlockTicketOriginalDemands(kt, ticket.Demands)
}

// checkTicketTimeout check ticket timeout
func (d *Dispatcher) checkTicketTimeout(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	submitTime, err := time.Parse(constant.TimeStdFormat, ticket.SubmittedAt)
	if err != nil {
		logs.Errorf("failed to parse ticket submit time %s, err: %v, rid: %s", ticket.SubmittedAt, err, kt.Rid)
		return err
	}

	// set timeout as 5 days
	if time.Now().Before(submitTime.AddDate(0, 0, enumor.AuditFlowTimeoutDay)) {
		return nil
	}

	return d.updateTicketStatusFailed(kt, ticket, "audit flow timeout")
}

// FinishAuditFlow 单据的所有子单均已结单，汇总成功的子单并应用到本地
func (d *Dispatcher) FinishAuditFlow(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	itsmStatus, err := d.itsmCli.GetTicketStatus(kt, ticket.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	// 将itsm单结单，如果itsm单已结单，继续走后续流程
	if len(itsmStatus.Data.CurrentSteps) > 0 && itsmStatus.Data.CurrentSteps[0].StateId == d.crpAuditNode.ID {
		approveReq := &itsm.ApproveReq{
			Sn:       ticket.ItsmSN,
			StateID:  int(d.crpAuditNode.ID),
			Approver: d.crpAuditNode.Approver,
			Action:   "true",
			Remark:   "",
		}
		if err := d.itsmCli.Approve(kt, approveReq); err != nil {
			logs.Errorf("request itsm ticket approve failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	// crp单据通过后更新本地数据表
	if err := d.queryAndApplyResPlanDemandChange(kt, ticket); err != nil {
		logs.Errorf("%s: failed to upsert crp demand, err: %v, rid: %s", constant.DemandChangeAppliedFailed,
			err, kt.Rid)
		return err
	}
	return nil
}

// updateTicketStatus update ticket status.
func (d *Dispatcher) updateTicketStatus(kt *kit.Kit, ticket *rpts.ResPlanTicketStatusTable) error {
	expr := tools.EqualExpression("ticket_id", ticket.TicketID)
	if err := d.dao.ResPlanTicketStatus().Update(kt, expr, ticket); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// UpdateTicketStatusFailed update ticket status to failed.
func (d *Dispatcher) updateTicketStatusFailed(kt *kit.Kit, ticket *ptypes.TicketInfo, msg string) error {
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusFailed,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
		CrpSN:    ticket.CrpSN,
		CrpURL:   ticket.CrpURL,
		Message:  msg,
	}

	if err := d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 失败需要释放资源
	allDemandIDs := make([]string, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allDemandIDs = append(allDemandIDs, (*demand.Original).DemandID)
		}
	}
	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(allDemandIDs, 0)
	if err := d.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
