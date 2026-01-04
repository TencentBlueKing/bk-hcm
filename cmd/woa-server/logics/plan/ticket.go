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
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpdaotypes "hcm/pkg/dal/dao/types/resource-plan"
	rpst "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/classifier"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// CreateResPlanTicket create resource plan ticket.
func (c *Controller) CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket request, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// construct resource plan ticket.
	ticket, err := c.constructResPlanTicket(kt, req, kt.User)
	if err != nil {
		logs.Errorf("failed to construct resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketID, err := c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ticketIDs, err := c.dao.ResPlanTicket().CreateWithTx(kt, txn, []rpt.ResPlanTicketTable{*ticket})
		if err != nil {
			logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		if len(ticketIDs) != 1 {
			logs.Errorf("create resource plan ticket, but len ticketIDs != 1, rid: %s", kt.Rid)
			return "", errors.New("create resource plan ticket, but len ticketIDs != 1")
		}

		ticketID := ticketIDs[0]

		// create resource plan ticket status.
		statuses := []rpts.ResPlanTicketStatusTable{{
			TicketID: ticketID,
			Status:   enumor.RPTicketStatusInit,
		}}
		if err = c.dao.ResPlanTicketStatus().CreateWithTx(kt, txn, statuses); err != nil {
			logs.Errorf("create resource plan ticket status failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		return ticketID, nil
	})

	if err != nil {
		logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketIDStr, ok := ticketID.(string)
	if !ok {
		logs.Errorf("convert resource plan ticket id %v from interface to string failed, err: %v, rid: %s",
			ticketID, err, kt.Rid)
		return "", fmt.Errorf("convert resource plan ticket id %v from interface to string failed", ticketID)
	}

	return ticketIDStr, nil
}

// constructResPlanTicket construct resource plan ticket.
func (c *Controller) constructResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq, applicant string) (
	*rpt.ResPlanTicketTable, error) {

	var originalOs, updatedOs decimal.Decimal
	var originalCpuCore, originalMemory, originalDiskSize int64
	var updatedCpuCore, updatedMemory, updatedDiskSize int64
	for _, demand := range req.Demands {
		if demand.Original != nil {
			originalOs = originalOs.Add((*demand.Original).Cvm.Os.Decimal)
			originalCpuCore += (*demand.Original).Cvm.CpuCore
			originalMemory += (*demand.Original).Cvm.Memory
			originalDiskSize += (*demand.Original).Cbs.DiskSize
		}

		if demand.Updated != nil {
			// 期望交付时间的预测需求月和其自然月必须一致，否则需要选择该周的其他时间
			et, err := times.ParseDay(demand.Updated.ExpectTime)
			if err != nil {
				logs.Errorf("failed to parse expect time, err: %v, expect_time: %s, rid: %s", err,
					demand.Updated.ExpectTime, kt.Rid)
				return nil, err
			}
			isCross, err := c.demandTime.IsDayCrossMonth(kt, et)
			if err != nil {
				logs.Errorf("failed to check if expect time is cross month, err: %v, expect_time: %s, rid: %s",
					err, et.String(), kt.Rid)
				return nil, err
			}
			if isCross {
				return nil, fmt.Errorf("expect_time should not be cross month, expect_time: %s",
					demand.Updated.ExpectTime)
			}
			updatedOs = updatedOs.Add((*demand.Updated).Cvm.Os.Decimal)
			updatedCpuCore += (*demand.Updated).Cvm.CpuCore
			updatedMemory += (*demand.Updated).Cvm.Memory
			updatedDiskSize += (*demand.Updated).Cbs.DiskSize
		}
	}

	demandsJson, err := tabletypes.NewJsonField(req.Demands)
	if err != nil {
		return nil, err
	}

	result := &rpt.ResPlanTicketTable{
		Type:             req.TicketType,
		Demands:          demandsJson,
		Applicant:        applicant,
		BkBizID:          req.BizOrgRel.BkBizID,
		BkBizName:        req.BizOrgRel.BkBizName,
		OpProductID:      req.BizOrgRel.OpProductID,
		OpProductName:    req.BizOrgRel.OpProductName,
		PlanProductID:    req.BizOrgRel.PlanProductID,
		PlanProductName:  req.BizOrgRel.PlanProductName,
		VirtualDeptID:    req.BizOrgRel.VirtualDeptID,
		VirtualDeptName:  req.BizOrgRel.VirtualDeptName,
		DemandClass:      req.DemandClass,
		OriginalOS:       originalOs.InexactFloat64(),
		OriginalCpuCore:  originalCpuCore,
		OriginalMemory:   originalMemory,
		OriginalDiskSize: originalDiskSize,
		UpdatedOS:        updatedOs.InexactFloat64(),
		UpdatedCpuCore:   updatedCpuCore,
		UpdatedMemory:    updatedMemory,
		UpdatedDiskSize:  updatedDiskSize,
		Remark:           req.Remark,
		Creator:          applicant,
		Reviser:          applicant,
		SubmittedAt:      time.Now().Format(constant.DateTimeLayout),
	}

	return result, nil
}

// GetResPlanTicketAudit get resource plan ticket audit.
func (c *Controller) GetResPlanTicketAudit(kt *kit.Kit, ticketID string, bkBizID int64) (
	*ptypes.GetResPlanTicketAuditResp, error) {

	return c.resFetcher.GetResPlanTicketAudit(kt, ticketID, bkBizID)
}

// GetResPlanTicketStatusInfo get resource plan ticket status information.
func (c *Controller) GetResPlanTicketStatusInfo(kt *kit.Kit, ticketID string) (
	*ptypes.GetRPTicketStatusInfo, error) {

	return c.resFetcher.GetResPlanTicketStatusInfo(kt, ticketID)
}

// GetCrpCurrentApprove get crp current approve.
func (c *Controller) GetCrpCurrentApprove(kt *kit.Kit, bkBizID int64, orderID string) (
	[]*ptypes.CrpAuditStep, error) {

	return c.resFetcher.GetCrpCurrentApprove(kt, bkBizID, orderID)
}

// ListAllResPlanTicket list all res plan ticket with status (join query).
func (c *Controller) ListAllResPlanTicket(kt *kit.Kit, listFilter *filter.Expression) (
	[]rpdaotypes.RPTicketWithStatus, error) {

	return c.resFetcher.ListAllResPlanTicket(kt, listFilter)
}

// ListResPlanTicketWithRes list res plan ticket with res.
func (c *Controller) ListResPlanTicketWithRes(kt *kit.Kit, req *core.ListReq) (*ptypes.RPTicketWithStatusAndResListRst,
	error) {

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rst, err := c.dao.ResPlanTicket().ListWithStatus(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list biz resource plan ticket with status, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if len(rst.Details) == 0 {
		return &ptypes.RPTicketWithStatusAndResListRst{Count: rst.Count}, nil
	}

	details, err := c.appendFieldToListResPlanTickets(kt, rst.Details)
	if err != nil {
		logs.Errorf("failed to append field to list res plan tickets, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &ptypes.RPTicketWithStatusAndResListRst{Count: 0, Details: details}, nil
}

func (c *Controller) appendFieldToListResPlanTickets(kt *kit.Kit, details []rpdaotypes.RPTicketWithStatus) (
	[]ptypes.RPTicketWithStatusAndRes, error) {

	// 获取子单通过情况
	ticketIDs := slice.Map(details, func(detail rpdaotypes.RPTicketWithStatus) string {
		return detail.ID
	})
	listOpt := &rpproto.ResPlanSubTicketListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("ticket_id", ticketIDs),
				tools.RuleEqual("status", enumor.RPSubTicketStatusDone),
			),
			Page: core.NewDefaultBasePage(),
			Fields: []string{"id", "ticket_id", "status", "sub_original_cpu_core", "sub_updated_cpu_core",
				"sub_original_memory", "sub_original_disk_size", "sub_updated_memory", "sub_updated_disk_size"},
		},
	}
	subTickets := make([]rpst.ResPlanSubTicketTable, 0)
	for {
		listRst, err := c.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, listOpt)
		if err != nil {
			logs.Errorf("failed to list resource plan sub ticket, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		subTickets = append(subTickets, listRst.Details...)

		if len(listRst.Details) < int(listOpt.Page.Limit) {
			break
		}
		listOpt.Page.Start += uint32(listOpt.Page.Limit)
	}

	subTicketMap := classifier.ClassifySlice(subTickets, func(item rpst.ResPlanSubTicketTable) string {
		return item.TicketID
	})

	ticketWithRes := make([]ptypes.RPTicketWithStatusAndRes, 0, len(details))
	for _, detail := range details {
		item := ptypes.RPTicketWithStatusAndRes{
			RPTicketWithStatus: detail,
			TicketTypeName:     detail.Type.Name(),
		}

		// set status name.
		item.StatusName = detail.Status.Name()
		// 资源需求报备数量
		switch detail.Type {
		case enumor.RPTicketTypeAdd:
			item.OriginalInfo = ptypes.NewNullResourceInfo()
			item.UpdatedInfo = ptypes.NewResourceInfo(detail.UpdatedCpuCore, detail.UpdatedMemory,
				detail.UpdatedDiskSize)
		case enumor.RPTicketTypeAdjust:
			item.OriginalInfo = ptypes.NewResourceInfo(detail.OriginalCpuCore, detail.OriginalMemory,
				detail.OriginalDiskSize)
			item.UpdatedInfo = ptypes.NewResourceInfo(detail.UpdatedCpuCore, detail.UpdatedMemory,
				detail.UpdatedDiskSize)
		case enumor.RPTicketTypeDelete:
			item.OriginalInfo = ptypes.NewResourceInfo(detail.OriginalCpuCore, detail.OriginalMemory,
				detail.OriginalDiskSize)
			item.UpdatedInfo = ptypes.NewNullResourceInfo()
		default:
			logs.Warnf("failed to append field to list res plan tickets: unsupported ticket type: %s, "+
				"ticket id: %s, rid: %s", detail.Type, detail.ID, kt.Rid)
		}

		// 已审批数
		if subs, ok := subTicketMap[detail.ID]; ok {
			item.AuditedOriginalInfo, item.AuditedUpdatedInfo = calcSubTicketsApprovedResources(subs)
		}

		ticketWithRes = append(ticketWithRes, item)
	}
	return ticketWithRes, nil
}

func calcSubTicketsApprovedResources(subTickets []rpst.ResPlanSubTicketTable) (
	ptypes.RPTicketResourceInfo, ptypes.RPTicketResourceInfo) {

	originalInfo := ptypes.NewNullResourceInfo()
	updatedInfo := ptypes.NewNullResourceInfo()
	for _, item := range subTickets {
		if cvt.PtrToVal(item.SubOriginalCPUCore) > 0 {
			originalInfo.Append(
				cvt.PtrToVal(item.SubOriginalCPUCore),
				cvt.PtrToVal(item.SubOriginalMemory),
				cvt.PtrToVal(item.SubOriginalDiskSize),
			)
		}
		if cvt.PtrToVal(item.SubUpdatedCPUCore) > 0 {
			updatedInfo.Append(
				cvt.PtrToVal(item.SubUpdatedCPUCore),
				cvt.PtrToVal(item.SubUpdatedMemory),
				cvt.PtrToVal(item.SubUpdatedDiskSize),
			)
		}
	}
	return originalInfo, updatedInfo
}

// ApproveResPlanSubTicketAdmin 审批资源预测子单 - 管理员审批阶段
func (c *Controller) ApproveResPlanSubTicketAdmin(kt *kit.Kit, subTicketID string, bizID int64,
	req *ptypes.AuditResPlanTicketAdminReq) error {

	// 校验审批人
	processors := c.resFetcher.GetAdminAuditors()
	if !slices.Contains(processors, kt.User) {
		logs.Errorf("not authorized to approve res plan sub ticket, user: %s, processors: %v, rid: %s",
			kt.User, processors, kt.Rid)
		return errors.New("not authorized to approve res plan sub ticket")
	}

	// 查询数据
	subTicket, err := c.getSubTicketByBiz(kt, subTicketID, bizID)
	if err != nil {
		logs.Errorf("failed to get sub_ticket status info, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 校验状态
	if !subTicket.IsInAdminAuditing() {
		logs.Errorf("sub ticket: %s is not in admin auditing, now status: %s, stage: %s, rid: %s",
			subTicket.ID, subTicket.Status.Name(), subTicket.Stage, kt.Rid)
	}

	// 审批，改变子单状态
	err = c.approveResPlanTicketAdmin(kt, subTicket, req)
	if err != nil {
		logs.Errorf("failed to approve res plan ticket admin, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (c *Controller) approveResPlanTicketAdmin(kt *kit.Kit, subTicket *rpst.ResPlanSubTicketTable,
	req *ptypes.AuditResPlanTicketAdminReq) error {

	var demandsStruct rpt.ResPlanDemands
	if err := json.Unmarshal([]byte(subTicket.SubDemands), &demandsStruct); err != nil {
		logs.Errorf("failed to unmarshal sub ticket demands, err: %v, id: %s, rid: %s", err,
			subTicket.ID, kt.Rid)
		return err
	}
	if len(demandsStruct) == 0 {
		logs.Errorf("sub ticket demands is empty, id: %s, rid: %s", subTicket.ID, kt.Rid)
		return fmt.Errorf("sub ticket %s demands is empty", subTicket.ID)
	}

	var subTicketType enumor.RPTicketType
	ticketStatus := enumor.RPSubTicketStatusRejected
	adminAuditStatus := enumor.RPAdminAuditStatusRejected
	if cvt.PtrToVal(req.Approval) {
		// 审批通过
		ticketStatus = enumor.RPSubTicketStatusAuditing
		adminAuditStatus = enumor.RPAdminAuditStatusDone
		if !cvt.PtrToVal(req.UseTransferPool) {
			// 不使用中转池，根据单据需求情况修改子单类型
			subTicketType = enumor.RPTicketTypeAdjust
			if demandsStruct[0].Original == nil {
				subTicketType = enumor.RPTicketTypeAdd
			}
			if demandsStruct[0].Updated == nil {
				subTicketType = enumor.RPTicketTypeDelete
			}
		}
	}

	updateItem := rpproto.ResPlanSubTicketUpdateReq{
		ID:                 subTicket.ID,
		SubType:            subTicketType,
		Status:             ticketStatus,
		AdminAuditStatus:   adminAuditStatus,
		AdminAuditOperator: kt.User,
		AdminAuditAt:       time.Now().Format(constant.DateTimeLayout),
	}
	updateReq := &rpproto.ResPlanSubTicketBatchUpdateReq{
		SubTickets: []rpproto.ResPlanSubTicketUpdateReq{updateItem},
	}
	return c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanSubTicket(kt, updateReq)
}

func (c *Controller) getSubTicketByBiz(kt *kit.Kit, subTicketID string, bizID int64) (
	*rpst.ResPlanSubTicketTable, error) {

	rules := []*filter.AtomRule{tools.RuleEqual("id", subTicketID)}
	if bizID != constant.AttachedAllBiz {
		rules = append(rules, tools.RuleEqual("bk_biz_id", bizID))
	}
	listReq := &rpproto.ResPlanSubTicketListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(rules...),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "sub_demands", "status", "stage", "admin_audit_status", "crp_sn", "crp_url"},
		},
	}
	rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list sub ticket, err: %v, id: %s, bk_biz_id: %d, rid: %s", err, subTicketID,
			bizID, kt.Rid)
		return nil, err
	}

	if len(rst.Details) < 1 {
		logs.Errorf("get sub ticket by biz, but len details != 1, id: %s, bk_biz_id: %d, rid: %s",
			subTicketID, bizID, kt.Rid)
		return nil, fmt.Errorf("cannot found sub_ticket: %s for biz %d", subTicketID, bizID)
	}

	ticketInfo := rst.Details[0]
	return &ticketInfo, nil
}

// GetResPlanTicketStatusByBiz 检查ticket是否存在，并校验业务id是否正确， bizID 为-1 表示不限制业务条件
func (c *Controller) GetResPlanTicketStatusByBiz(kt *kit.Kit, ticketID string, bizID int64) (
	*ptypes.GetRPTicketStatusInfo, error) {

	// 1. 检查ticket是否存在以及业务是否匹配
	rules := []*filter.AtomRule{tools.RuleEqual("id", ticketID)}
	if bizID != constant.AttachedAllBiz {
		rules = append(rules, tools.RuleEqual("bk_biz_id", bizID))
	}
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewCountPage(),
	}

	ticketRst, err := c.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket(%s,%d), err: %v, rid: %s", ticketID, bizID, err, kt.Rid)
		return nil, err
	}

	if ticketRst.Count < 1 {
		logs.Errorf("list resource plan ticket got %d != 1, rid: %s", ticketRst.Count, kt.Rid)
		return nil, fmt.Errorf("list resource plan ticket %s by biz %d failed", ticketID, bizID)
	}

	// 2. 查询对应状态单号
	statusOpt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	statusRst, err := c.dao.ResPlanTicketStatus().List(kt, statusOpt)
	if err != nil {
		logs.Errorf("failed to list status of resource plan ticket(%s), err: %v, rid: %s", ticketID, err, kt.Rid)
		return nil, err
	}

	if len(statusRst.Details) != 1 {
		logs.Errorf("list status of resource plan ticket got %d != 1, ticket_id: %s, rid: %s",
			len(statusRst.Details), ticketID, kt.Rid)
		return nil, errors.New("list status of resource plan ticket, but len != 1")
	}

	detail := statusRst.Details[0]
	result := &ptypes.GetRPTicketStatusInfo{
		Status:     detail.Status,
		StatusName: detail.Status.Name(),
		ItsmSN:     detail.ItsmSN,
		ItsmURL:    detail.ItsmURL,
		CrpSN:      detail.CrpSN,
		CrpURL:     detail.CrpURL,
		Message:    detail.Message,
	}

	return result, nil
}

// TerminateResPlanFailedTicket 终止失败的预测单据
func (c *Controller) TerminateResPlanFailedTicket(kt *kit.Kit, ticketID string) error {
	// 1. 获取主单信息
	ticket, err := c.resFetcher.GetTicketInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
		return err
	}

	// 2. 仅失败、部分失败的单据可以终止
	switch ticket.Status {
	case enumor.RPTicketStatusFailed, enumor.RPTicketStatusPartialFailed:
	default:
		logs.Errorf("ticket status is %s, can't terminate, ticket id: %s, rid: %s", ticket.Status, ticketID,
			kt.Rid)
		return fmt.Errorf("ticket status is %s, can't terminate", ticket.Status)
	}

	// 3. 子单汇总生效
	if err := c.dispatcher.FinishAuditFlow(kt, ticket); err != nil {
		logs.Errorf("failed to finish audit flow, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	// 4. 单据状态置为 RPTicketStatusTerminated
	err = c.updateTicketStatus(kt, &rpts.ResPlanTicketStatusTable{
		TicketID: ticketID,
		Status:   enumor.RPTicketStatusTerminated,
	})
	if err != nil {
		logs.Errorf("failed to update res plan ticket status, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
		return err
	}
	return nil
}
