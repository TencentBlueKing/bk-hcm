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

package splitter

import (
	"errors"
	"time"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

// createSubTicket create sub ticket from adjSplitGroupDemands
func (s *SubTicketSplitter) createSubTicket(kt *kit.Kit, ticketID string, allDemands rpt.ResPlanDemands,
	defaultType enumor.RPTicketType) error {

	// 查询管理员审批下限核数
	quotaCfg, err := s.resFetcher.GetPlanTransferQuotaConfigs(kt)
	if err != nil {
		logs.Errorf("failed to get plan transfer quota config, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 查询父单据详情
	ticket, err := s.getTicketBaseInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v, ticketID: %s, rid: %s", err, ticketID, kt.Rid)
		return err
	}

	subTickets := make([]rpproto.ResPlanSubTicketCreateReq, 0)
	for subTicketType, demands := range s.adjSplitGroupDemands {
		if len(demands) == 0 {
			continue
		}
		for _, d := range demands {
			logs.Infof("sub_ticket type: %s, item: %s, ticket id: %s, rid: %s", subTicketType, d.String(),
				ticketID, kt.Rid)
		}
		demandsJson, err := tabletypes.NewJsonField(demands)
		if err != nil {
			logs.Errorf("failed to create json field, err: %v, demands: %+v, rid: %s", err,
				cvt.PtrToSlice(demands), kt.Rid)
			return err
		}

		subTicket := constructSubTicketCreateReq(ticket, quotaCfg.AuditQuota, subTicketType, demands, demandsJson)
		subTickets = append(subTickets, subTicket)
	}
	// 没有发生拆分，可能是没有可转移的预测，此时按照完整的需求列表创建一个子单即可
	if len(subTickets) == 0 {
		demandsJson, err := tabletypes.NewJsonField(allDemands)
		if err != nil {
			logs.Errorf("failed to create json field, err: %v, demands: %+v, rid: %s", err,
				allDemands, kt.Rid)
			return err
		}
		subTicket := constructSubTicketCreateReq(ticket, quotaCfg.AuditQuota, defaultType, cvt.SliceToPtr(allDemands),
			demandsJson)
		subTickets = append(subTickets, subTicket)
	}

	createReq := &rpproto.ResPlanSubTicketBatchCreateReq{
		SubTickets: subTickets,
	}

	// 后台任务，使用提单人作为子单的创建人
	kt.User = ticket.Applicant
	_, err = s.client.DataService().Global.ResourcePlan.BatchCreateResPlanSubTicket(kt, createReq)
	if err != nil {
		logs.Errorf("create sub ticket failed, err: %v, req: %+v, rid: %s", err, createReq, kt.Rid)
		return err
	}
	return nil
}

// constructSubTicketCreateReq 构造子单据创建请求
func constructSubTicketCreateReq(ticket *rpt.ResPlanTicketTable, auditQuota int64, subTicketType enumor.RPTicketType,
	demands []*rpt.ResPlanDemand, demandsJson tabletypes.JsonField) rpproto.ResPlanSubTicketCreateReq {

	// transfer_in 和 transfer_out 在子单中均表现为 transfer 类型，通过 demand 内容决定转移方向
	if subTicketType == enumor.RPTicketTypeTransferIN || subTicketType == enumor.RPTicketTypeTransferOUT {
		subTicketType = enumor.RPTicketTypeTransfer
	}
	// transfer_out_exempt 类型在子单中表现为 transfer_exempt 类型
	if subTicketType == enumor.RPTicketTypeTransferOUTExempt {
		subTicketType = enumor.RPTicketTypeTransferExempt
	}

	var originalOs, updatedOs decimal.Decimal
	var originalCpuCore, originalMemory, originalDiskSize int64
	var updatedCpuCore, updatedMemory, updatedDiskSize int64
	for _, demand := range demands {
		if demand.Original != nil {
			originalOs = originalOs.Add(demand.Original.Cvm.Os.Decimal)
			originalCpuCore += demand.Original.Cvm.CpuCore
			originalMemory += demand.Original.Cvm.Memory
			originalDiskSize += demand.Original.Cbs.DiskSize
		}

		if demand.Updated != nil {
			updatedOs = updatedOs.Add(demand.Updated.Cvm.Os.Decimal)
			updatedCpuCore += demand.Updated.Cvm.CpuCore
			updatedMemory += demand.Updated.Cvm.Memory
			updatedDiskSize += demand.Updated.Cbs.DiskSize
		}
	}

	subTicket := rpproto.ResPlanSubTicketCreateReq{
		TicketID:        ticket.ID,
		BkBizID:         ticket.BkBizID,
		BkBizName:       ticket.BkBizName,
		OpProductID:     ticket.OpProductID,
		OpProductName:   ticket.OpProductName,
		PlanProductID:   ticket.PlanProductID,
		PlanProductName: ticket.PlanProductName,
		VirtualDeptID:   ticket.VirtualDeptID,
		VirtualDeptName: ticket.VirtualDeptName,
		SubType:         subTicketType,
		SubDemands:      demandsJson,
		// 默认等待类似子单合并
		Status:              enumor.RPSubTicketStatusWaiting,
		Stage:               enumor.RPSubTicketStageAdminAudit,
		AdminAuditStatus:    enumor.RPAdminAuditStatusAuditing,
		SubOriginalOS:       cvt.ValToPtr(originalOs.InexactFloat64()),
		SubOriginalCPUCore:  cvt.ValToPtr(originalCpuCore),
		SubOriginalMemory:   cvt.ValToPtr(originalMemory),
		SubOriginalDiskSize: cvt.ValToPtr(originalDiskSize),
		SubUpdatedOS:        cvt.ValToPtr(updatedOs.InexactFloat64()),
		SubUpdatedCPUCore:   cvt.ValToPtr(updatedCpuCore),
		SubUpdatedMemory:    cvt.ValToPtr(updatedMemory),
		SubUpdatedDiskSize:  cvt.ValToPtr(updatedDiskSize),
		SubmittedAt:         time.Now().Format(constant.DateTimeLayout),
	}
	// 调减单、自动延期单、非转移单跳过管理员审批
	if ticket.Type == enumor.RPTicketTypeDelete || ticket.Type == enumor.RPTicketTypeAutomaticTransfer ||
		subTicket.SubType != enumor.RPTicketTypeTransfer {

		subTicket.AdminAuditStatus = enumor.RPAdminAuditStatusSkip
	}
	if subTicket.SubType == enumor.RPTicketTypeTransfer || subTicket.SubType == enumor.RPTicketTypeTransferExempt {
		// 转移单核数小于审批下限，跳过管理员审批
		if updatedCpuCore <= auditQuota {
			subTicket.AdminAuditStatus = enumor.RPAdminAuditStatusSkip
		}
		// 转移单不等待合并
		subTicket.Status = enumor.RPSubTicketStatusAuditing
	}

	return subTicket
}

// getTicketBaseInfo 查询父单据详情
func (s *SubTicketSplitter) getTicketBaseInfo(kt *kit.Kit, ticketID string) (*rpt.ResPlanTicketTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := s.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket, but len details != 1")
	}

	return &rst.Details[0], nil
}
