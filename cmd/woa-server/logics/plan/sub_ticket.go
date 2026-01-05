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
	"fmt"

	"hcm/cmd/woa-server/logics/plan/splitter"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListResPlanSubTicket list resource plan sub_ticket.
func (c *Controller) ListResPlanSubTicket(kt *kit.Kit, req *ptypes.ListResPlanSubTicketReq) (
	*ptypes.ListResPlanSubTicketResp, error) {

	return c.resFetcher.ListResPlanSubTicket(kt, req)
}

// GetResPlanSubTicketDetail get resource plan sub_ticket detail.
func (c *Controller) GetResPlanSubTicketDetail(kt *kit.Kit, subTicketID string) (*ptypes.GetSubTicketDetailResp, string,
	error) {

	return c.resFetcher.GetResPlanSubTicketDetail(kt, subTicketID)
}

// GetResPlanSubTicketAudit get res plan sub ticket audit
func (c *Controller) GetResPlanSubTicketAudit(kt *kit.Kit, bizID int64, subTicketID string) (
	*ptypes.GetSubTicketAuditResp, string, error) {

	return c.resFetcher.GetResPlanSubTicketAudit(kt, bizID, subTicketID)
}

// RetryResPlanFailedSubTickets 重试失败的子单
func (c *Controller) RetryResPlanFailedSubTickets(kt *kit.Kit, ticketID string) error {
	// TODO 先用同步方式验证速度，拆单速度太慢的话需要转异步；转异步时需要根据kt创建新的子kit
	// 1. 获取主单信息
	ticket, err := c.resFetcher.GetTicketInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
		return err
	}

	// 2. 已终止的主单不可重试
	if ticket.Status == enumor.RPTicketStatusTerminated {
		logs.Errorf("ticket already terminated, cannot retry, ticket id: %s, rid: %s", ticketID, kt.Rid)
		return fmt.Errorf("ticket already terminated")
	}

	// 3. 处理失败的子单，更改状态，并汇总后重新拆分子单
	if err := c.retrySplitResPlanTickets(kt, ticketID, ticket); err != nil {
		logs.Errorf("failed to retry split res plan tickets, err: %v, ticket id: %s, rid: %s", err, ticketID,
			kt.Rid)
		return err
	}

	// 4. 为防止重试过程中所有子单进入终态触发主单结单，这里将主单重新启动
	err = c.updateTicketStatus(kt, &rpts.ResPlanTicketStatusTable{
		TicketID: ticketID,
		Status:   enumor.RPTicketStatusAuditing,
	})
	if err != nil {
		logs.Errorf("failed to update res plan ticket status, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
		return err
	}
	return nil
}

// retrySplitResPlanTickets 重试，将失败的子单置为已失效状态，并汇总后重新拆分子单
func (c *Controller) retrySplitResPlanTickets(kt *kit.Kit, ticketID string, ticket *ptypes.TicketInfo) error {
	// 1. 获取失败的子单列表
	failedIDs, failedDemands, err := c.getFailedSubTicketsByTicketID(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get failed sub tickets by ticket id %s, err: %v, rid: %s", ticketID, err, kt.Rid)
		return err
	}
	if len(failedIDs) == 0 {
		logs.Warnf("no failed sub tickets found, ticket id: %s, rid: %s", ticketID, kt.Rid)
		return nil
	}

	// 2. 将失败的子单置为已失效状态
	updateReq := &rpproto.ResPlanSubTicketStatusUpdateReq{
		IDs:      failedIDs,
		TicketID: ticketID,
		Source:   enumor.RPSubTicketStatusFailed,
		Target:   enumor.RPSubTicketStatusInvalid,
	}
	err = c.client.DataService().Global.ResourcePlan.UpdateResPlanSubTicketStatusCAS(kt, updateReq)
	if err != nil {
		logs.Errorf("failed to update res plan sub ticket status %s to %s, err: %v, ticket id: %s, rid: %s",
			updateReq.Source, updateReq.Target, err, ticketID, kt.Rid)
		return err
	}

	// 当重试失败时，需要将子单状态回滚至失败
	var splitErr error
	defer func() {
		if splitErr != nil {
			updateReq.Source = enumor.RPSubTicketStatusInvalid
			updateReq.Target = enumor.RPSubTicketStatusFailed
			subErr := c.client.DataService().Global.ResourcePlan.UpdateResPlanSubTicketStatusCAS(kt, updateReq)
			if subErr != nil {
				logs.Errorf("failed to update res plan sub ticket status %s to %s, err: %v, ticket id: %s, rid: %s",
					updateReq.Source, updateReq.Target, subErr, ticketID, kt.Rid)
			}
		}
	}()

	// 3. 汇总失败的单据重新拆分子单
	splitHelper := splitter.New(c.dao, c.client, c.crpCli, c.resFetcher, c.deviceTypesMap)
	switch ticket.Type {
	case enumor.RPTicketTypeDelete, enumor.RPTicketTypeAutomaticTransfer:
		splitErr = splitHelper.SplitDeleteTicket(kt, ticket.ID, ticket.Type, failedDemands, ticket.PlanProductName,
			ticket.OpProductName)
	case enumor.RPTicketTypeAdd:
		splitErr = splitHelper.SplitAddTicket(kt, ticket.ID, failedDemands)
	case enumor.RPTicketTypeAdjust:
		splitErr = splitHelper.SplitAdjustTicket(kt, ticket.ID, failedDemands, ticket.PlanProductName,
			ticket.OpProductName)
	default:
		splitErr = fmt.Errorf("unsupported res plan ticket type, type: %s", ticket.Type)
	}
	if splitErr != nil {
		logs.Errorf("failed to split res plan ticket, err: %v, ticket id: %s, rid: %s", splitErr, ticket.ID, kt.Rid)
		return splitErr
	}

	return nil
}

// getFailedSubTicketsByTicketID 获取失败的子单列表
func (c *Controller) getFailedSubTicketsByTicketID(kt *kit.Kit, ticketID string) ([]string, rpt.ResPlanDemands, error) {
	// 获取失败的子单列表
	failedIDs := make([]string, 0)
	failedDemands := make([]rpt.ResPlanDemand, 0)
	listReq := &ptypes.ListResPlanSubTicketReq{
		TicketID: ticketID,
		Statuses: []enumor.RPSubTicketStatus{enumor.RPSubTicketStatusFailed},
		Page:     core.NewDefaultBasePage(),
	}
	for {
		rst, err := c.resFetcher.ListResPlanSubTicket(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list res plan sub ticket, err: %v, ticket id: %s, rid: %s", err, ticketID,
				kt.Rid)
			return nil, nil, err
		}

		for _, item := range rst.Details {
			var demands rpt.ResPlanDemands
			if err = json.Unmarshal([]byte(item.SubDemands), &demands); err != nil {
				logs.Errorf("failed to unmarshal demands, err: %v, sub ticket id: %s, rid: %s", err, item.ID,
					kt.Rid)
				return nil, nil, err
			}

			failedIDs = append(failedIDs, item.ID)
			failedDemands = append(failedDemands, demands...)
		}

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return failedIDs, failedDemands, nil
}
