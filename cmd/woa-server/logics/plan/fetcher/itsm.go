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

package fetcher

import (
	"fmt"
	"strings"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// getItsmAndCrpAuditStatus get itsm and crp audit status.
func (f *ResPlanFetcher) getItsmAndCrpAuditStatus(kt *kit.Kit, bkBizID int64,
	ticketStatus *ptypes.GetRPTicketStatusInfo) (*ptypes.GetRPTicketItsmAudit, *ptypes.GetRPTicketCrpAudit, error) {

	if ticketStatus.ItsmSN == constant.ResPlanItsmAuditSkip {
		return nil, nil, nil
	}

	itsmAudit := &ptypes.GetRPTicketItsmAudit{
		ItsmSN:  ticketStatus.ItsmSN,
		ItsmURL: ticketStatus.ItsmURL,
	}
	// 审批未开始
	if ticketStatus.ItsmSN == "" {
		itsmAudit.Status = enumor.RPTicketStatusInit
		itsmAudit.StatusName = itsmAudit.Status.Name()
		return itsmAudit, nil, nil
	}

	// 获取ITSM审批记录和当前审批节点
	itsmStatus, err := f.itsmCli.GetTicketStatus(kt, ticketStatus.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm audit status, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSN, kt.Rid)
		return nil, nil, err
	}
	itsmLogs, err := f.itsmCli.GetTicketLog(kt, ticketStatus.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm audit log, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSN, kt.Rid)
		return nil, nil, err
	}
	if itsmLogs.Data == nil {
		logs.Errorf("itsm audit log is empty, sn: %s, rid: %s", ticketStatus.ItsmSN, kt.Rid)
		return nil, nil, fmt.Errorf("itsm audit log is empty, sn: %s", ticketStatus.ItsmSN)
	}

	itsmAudit, err = f.setItsmAuditDetails(kt, bkBizID, itsmAudit, itsmStatus, itsmLogs.Data)
	if err != nil {
		logs.Errorf("failed to set itsm audit details, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSN, kt.Rid)
		return nil, nil, err
	}

	// ITSM审批中或审批终止在itsm阶段
	if ticketStatus.CrpSN == "" {
		// ITSM流程没有正常结束，将单据审批状态作为ITSM流程的当前状态
		if itsmAudit.Status != enumor.RPTicketStatusDone {
			itsmAudit.Status = ticketStatus.Status
			itsmAudit.StatusName = itsmAudit.Status.Name()
			itsmAudit.Message = ticketStatus.Message
			return itsmAudit, nil, nil
		}
		// ITSM流程正常结束，主单审批流即完结
		// 为兼容旧版本数据，可能存在 crp_sn 不为空的数据，需在下文逻辑中保留兼容
		return itsmAudit, nil, nil
	}
	// itsm审批流已结束
	itsmAudit.Status = enumor.RPTicketStatusDone
	itsmAudit.StatusName = itsmAudit.Status.Name()

	// 流程走到CRP步骤，获取CRP审批记录和当前审批节点
	crpCurrentSteps, err := f.GetCrpCurrentApprove(kt, bkBizID, ticketStatus.CrpSN)
	if err != nil {
		logs.Errorf("failed to get crp current approve, err: %v, sn: %s, rid: %s", err, ticketStatus.CrpSN, kt.Rid)
		return nil, nil, err
	}
	crpApproveLogs, err := f.GetCrpApproveLogs(kt, ticketStatus.CrpSN)
	if err != nil {
		logs.Errorf("failed to get crp approve logs, err: %v, sn: %s, rid: %s", err, ticketStatus.CrpSN, kt.Rid)
		return nil, nil, err
	}

	// CRP审批状态赋值
	crpAudit := &ptypes.GetRPTicketCrpAudit{
		CrpSN:        ticketStatus.CrpSN,
		CrpURL:       ticketStatus.CrpURL,
		Status:       ticketStatus.Status,
		StatusName:   ticketStatus.Status.Name(),
		Message:      ticketStatus.Message,
		CurrentSteps: crpCurrentSteps,
		Logs:         crpApproveLogs,
	}
	return itsmAudit, crpAudit, nil
}

// setItsmAuditDetails set itsm audit details
func (f *ResPlanFetcher) setItsmAuditDetails(kt *kit.Kit, bkBizID int64, itsmAudit *ptypes.GetRPTicketItsmAudit,
	current *itsm.GetTicketStatusResp, logData *itsm.GetTicketLogRst) (*ptypes.GetRPTicketItsmAudit, error) {

	// current steps
	itsmAudit.CurrentSteps = make([]*ptypes.ItsmAuditStep, len(current.Data.CurrentSteps))
	for i, step := range current.Data.CurrentSteps {
		// 校验审批人是否有该业务的访问权限
		processors := strings.Split(step.Processors, ",")
		processorAuth := make(map[string]bool)
		var err error
		if bkBizID > 0 && len(processors) > 0 {
			processorAuth, err = f.bizLogics.BatchCheckUserBizAccessAuth(kt, bkBizID, processors)
			if err != nil {
				return nil, err
			}
		}

		// 校验审批人是否有该业务的访问权限
		itsmAudit.CurrentSteps[i] = &ptypes.ItsmAuditStep{
			StateID:        step.StateId,
			Name:           step.Name,
			Processors:     processors,
			ProcessorsAuth: processorAuth,
		}
	}

	// logs
	itsmAudit.Logs = make([]*ptypes.ItsmAuditLog, 0, len(logData.Logs))
	for _, log := range logData.Logs {
		// 流程开始、结束、CRP审批 不展示
		if log.Message == itsm.AuditNodeStart || log.Message == itsm.AuditNodeEnd ||
			log.Operator == enumor.TicketOperatorNameCrpAudit {
			continue
		}

		itsmAudit.Logs = append(itsmAudit.Logs, &ptypes.ItsmAuditLog{
			Operator:  log.Operator,
			OperateAt: log.OperateAt,
			Message:   log.Message,
		})
	}

	// 如果itsm审批流已经到了CRP阶段，需要赋值为结束状态
	if len(current.Data.CurrentSteps) > 0 && current.Data.CurrentSteps[0].StateId == f.crpAuditNode.ID {
		itsmAudit.Status = enumor.RPTicketStatusDone
		itsmAudit.StatusName = itsmAudit.Status.Name()
		itsmAudit.CurrentSteps = itsmAudit.CurrentSteps[:0]
	}
	// 如果itsm审批流已结束，需要赋值为结束状态
	if current.Data.CurrentStatus == string(itsm.StatusFinished) {
		itsmAudit.Status = enumor.RPTicketStatusDone
		itsmAudit.StatusName = itsmAudit.Status.Name()
	}

	return itsmAudit, nil
}
