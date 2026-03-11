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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	rpst "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// ListResPlanSubTicket list resource plan sub_ticket.
func (f *ResPlanFetcher) ListResPlanSubTicket(kt *kit.Kit, req *ptypes.ListResPlanSubTicketReq) (
	*ptypes.ListResPlanSubTicketResp, error) {

	listOpt := &rpproto.ResPlanSubTicketListReq{
		ListReq: req.GenListOption(),
	}

	listRst, err := f.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan sub ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return genResPlanSubTicketListResp(listRst), nil
}

func genResPlanSubTicketListResp(listRst *rpproto.ResPlanSubTicketListResult) *ptypes.ListResPlanSubTicketResp {
	details := make([]ptypes.ListResPlanSubTicketItem, 0, len(listRst.Details))
	for _, item := range listRst.Details {
		listItem := ptypes.ListResPlanSubTicketItem{
			ID:             item.ID,
			Status:         item.Status,
			StatusName:     item.Status.Name(),
			Message:        cvt.PtrToVal(item.Message),
			SubDemands:     item.SubDemands,
			Stage:          item.Stage,
			SubTicketType:  item.SubType,
			TicketTypeName: item.SubType.Name(),
			CrpSN:          item.CrpSN,
			CrpURL:         item.CrpURL,
			SubmittedAt:    item.SubmittedAt,
			CreatedAt:      item.CreatedAt.String(),
			UpdatedAt:      item.UpdatedAt.String(),
		}
		if cvt.PtrToVal(item.SubOriginalOS) > 0 {
			listItem.OriginalInfo = ptypes.NewResourceInfo(
				cvt.PtrToVal(item.SubOriginalCPUCore),
				cvt.PtrToVal(item.SubOriginalMemory),
				cvt.PtrToVal(item.SubOriginalDiskSize))
		} else {
			listItem.OriginalInfo = ptypes.NewNullResourceInfo()
		}
		if cvt.PtrToVal(item.SubUpdatedOS) > 0 {
			listItem.UpdatedInfo = ptypes.NewResourceInfo(
				cvt.PtrToVal(item.SubUpdatedCPUCore),
				cvt.PtrToVal(item.SubUpdatedMemory),
				cvt.PtrToVal(item.SubUpdatedDiskSize))
		} else {
			listItem.UpdatedInfo = ptypes.NewNullResourceInfo()
		}

		details = append(details, listItem)
	}

	return &ptypes.ListResPlanSubTicketResp{
		Count:   listRst.Count,
		Details: details,
	}
}

// GetResPlanSubTicketDetail get resource plan sub_ticket detail.
func (f *ResPlanFetcher) GetResPlanSubTicketDetail(kt *kit.Kit, subTicketID string) (*ptypes.GetSubTicketDetailResp,
	string, error) {

	// 获取子单详情
	detail, err := f.getSubTicketDetail(kt, subTicketID)
	if err != nil {
		logs.Errorf("failed to get res plan sub ticket detail, err: %v, id: %s, rid: %s", err, subTicketID,
			kt.Rid)
		return nil, "", err
	}

	// 从父单据获取申请人
	parentDetail, err := f.getTicketBaseInfo(kt, detail.TicketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v, id: %s, rid: %s", err, detail.TicketID, kt.Rid)
		return nil, "", err
	}

	subTicketDetail, err := genSubTicketDetailGetResp(kt, detail)
	if err != nil {
		logs.Errorf("failed to gen sub ticket detail get resp, err: %v, id: %s, rid: %s", err, subTicketID,
			kt.Rid)
		return nil, "", err
	}

	return subTicketDetail, parentDetail.Applicant, nil
}

// getSubTicketDetail get res plan sub_ticket detail
func (f *ResPlanFetcher) getSubTicketDetail(kt *kit.Kit, subTicketID string) (*tablers.ResPlanSubTicketTable, error) {

	getOpt := &rpproto.ResPlanSubTicketListReq{
		ListReq: core.ListReq{
			Filter: tools.EqualExpression("id", subTicketID),
			Page:   core.NewDefaultBasePage(),
		},
	}

	getRst, err := f.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, getOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan sub ticket, err: %v, id: %s, rid: %s", err, subTicketID,
			kt.Rid)
		return nil, err
	}

	if len(getRst.Details) <= 0 {
		logs.Errorf("no resource plan sub ticket found, id: %s, rid: %s", subTicketID, kt.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("sub ticket %s not found", subTicketID))
	}

	return &getRst.Details[0], nil
}

func genSubTicketDetailGetResp(kt *kit.Kit, detail *tablers.ResPlanSubTicketTable) (
	*ptypes.GetSubTicketDetailResp, error) {

	baseInfo := ptypes.GetSubTicketBaseInfo{
		Type:          detail.SubType,
		TypeName:      detail.SubType.Name(),
		BkBizID:       detail.BkBizID,
		OpProductID:   detail.OpProductID,
		PlanProductID: detail.PlanProductID,
		VirtualDeptID: detail.VirtualDeptID,
		SubmittedAt:   detail.SubmittedAt,
	}

	statusInfo := ptypes.GetSubTicketStatusInfo{
		Status:           detail.Status,
		StatusName:       detail.Status.Name(),
		Stage:            detail.Stage,
		AdminAuditStatus: detail.AdminAuditStatus,
		CrpSN:            detail.CrpSN,
		CrpURL:           detail.CrpURL,
		Message:          cvt.PtrToVal(detail.Message),
	}

	var demandsStruct rpt.ResPlanDemands
	if err := json.Unmarshal([]byte(detail.SubDemands), &demandsStruct); err != nil {
		logs.Errorf("failed to unmarshal demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	demands := make([]ptypes.GetRPTicketDemand, len(demandsStruct))
	for idx, demand := range demandsStruct {
		demands[idx] = ptypes.GetRPTicketDemand{
			DemandClass:  demand.DemandClass,
			OriginalInfo: demand.Original,
			UpdatedInfo:  demand.Updated,
		}
	}

	return &ptypes.GetSubTicketDetailResp{
		ID:         detail.ID,
		BaseInfo:   baseInfo,
		StatusInfo: statusInfo,
		Demands:    demands,
	}, nil
}

// GetAdminAuditors get admin auditors
func (f *ResPlanFetcher) GetAdminAuditors() []string {
	// 管理员审批阶段审批人
	processors := f.resPlanCfg.AdminAuditor
	if len(processors) == 0 {
		processors = strings.Split(constant.AdminHandler, ";")
	}
	return processors
}

// GetResPlanSubTicketAudit get res plan sub ticket audit
func (f *ResPlanFetcher) GetResPlanSubTicketAudit(kt *kit.Kit, bizID int64, subTicketID string) (
	*ptypes.GetSubTicketAuditResp, string, error) {

	// 获取子单详情
	detail, err := f.getSubTicketDetail(kt, subTicketID)
	if err != nil {
		logs.Errorf("failed to get res plan sub ticket detail, err: %v, id: %s, rid: %s", err, subTicketID,
			kt.Rid)
		return nil, "", err
	}
	if bizID != constant.AttachedAllBiz && detail.BkBizID != bizID {
		return nil, "", errors.New("no permission to access this ticket")
	}

	// 从父单据获取申请人
	parentDetail, err := f.getTicketBaseInfo(kt, detail.TicketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v, id: %s, rid: %s", err, detail.TicketID, kt.Rid)
		return nil, "", err
	}

	processors := f.GetAdminAuditors()
	var crpAudit *ptypes.GetRPTicketCrpAudit
	// 流程走到CRP步骤，获取CRP审批记录和当前审批节点
	if detail.CrpSN != "" {
		crpAudit = new(ptypes.GetRPTicketCrpAudit)
		crpCurrentSteps, err := f.GetCrpCurrentApprove(kt, detail.BkBizID, detail.CrpSN)
		if err != nil {
			logs.Errorf("failed to get crp current approve, err: %v, sn: %s, rid: %s", err, detail.CrpSN, kt.Rid)
			return nil, "", err
		}
		crpApproveLogs, err := f.GetCrpApproveLogs(kt, detail.CrpSN)
		if err != nil {
			logs.Errorf("failed to get crp approve logs, err: %v, sn: %s, rid: %s", err, detail.CrpSN, kt.Rid)
			return nil, "", err
		}

		crpAudit.CrpSN = detail.CrpSN
		crpAudit.CrpURL = detail.CrpURL
		// CRP审批状态赋值
		crpAudit.Status = enumor.RPTicketStatusAuditing
		// 没有当前节点，即CRP单据结束（不一定成功）
		if len(crpCurrentSteps) == 0 {
			crpAudit.Status = enumor.RPTicketStatusDone
		}
		crpAudit.StatusName = crpAudit.Status.Name()
		crpAudit.Message = cvt.PtrToVal(detail.Message)
		crpAudit.CurrentSteps = crpCurrentSteps
		crpAudit.Logs = crpApproveLogs
	}

	adminAudit := &ptypes.GetRPTicketAdminAudit{
		Status: detail.AdminAuditStatus,
	}
	adminAudit.CurrentSteps = append(adminAudit.CurrentSteps, &ptypes.AdminAuditStep{
		Name:       constant.AdminAuditStepName,
		Processors: processors,
		ProcessorsAuth: slice.FuncToMap(processors, func(processor string) (string, bool) {
			return processor, true
		}),
	})
	if detail.AdminAuditStatus.IsFinished() {
		adminAudit.CurrentSteps = []*ptypes.AdminAuditStep{}
		adminAudit.Logs = append(adminAudit.Logs, &ptypes.AdminAuditLog{
			Name:        constant.AdminAuditStepName,
			Operator:    detail.AdminAuditOperator,
			OperateAt:   detail.AdminAuditAt,
			OperateInfo: cvt.PtrToVal(detail.OperateInfo),
		})
	}

	return &ptypes.GetSubTicketAuditResp{
		ID:         detail.ID,
		AdminAudit: adminAudit,
		CRPAudit:   crpAudit,
	}, parentDetail.Applicant, nil
}

// GetSubTicketInfo get sub ticket info
func (f *ResPlanFetcher) GetSubTicketInfo(kt *kit.Kit, subTicketID string) (*ptypes.SubTicketInfo, error) {
	info, err := f.getSubTicketByID(kt, subTicketID)
	if err != nil {
		logs.Errorf("failed to get sub ticket info, err: %v, sub ticket id: %s, rid: %s", err, subTicketID, kt.Rid)
		return nil, err
	}

	var demands rpt.ResPlanDemands
	if err = json.Unmarshal([]byte(info.SubDemands), &demands); err != nil {
		logs.Errorf("failed to unmarshal demands, err: %v, sub ticket id: %s, rid: %s", err, subTicketID, kt.Rid)
		return nil, err
	}

	parentTicket, err := f.getTicketBaseInfo(kt, info.TicketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v, ticket id: %s, rid: %s", err, info.TicketID, kt.Rid)
		return nil, err
	}

	brief := &ptypes.SubTicketInfo{
		ID:               subTicketID,
		ParentTicketID:   info.TicketID,
		BkBizID:          info.BkBizID,
		BkBizName:        info.BkBizName,
		OpProductID:      info.OpProductID,
		OpProductName:    info.OpProductName,
		PlanProductID:    info.PlanProductID,
		PlanProductName:  info.PlanProductName,
		VirtualDeptID:    info.VirtualDeptID,
		VirtualDeptName:  info.VirtualDeptName,
		DemandClass:      parentTicket.DemandClass,
		Applicant:        parentTicket.Applicant,
		Type:             info.SubType,
		OriginalCpuCore:  cvt.PtrToVal(info.SubOriginalCPUCore),
		OriginalMemory:   cvt.PtrToVal(info.SubOriginalMemory),
		OriginalDiskSize: cvt.PtrToVal(info.SubOriginalDiskSize),
		UpdatedCpuCore:   cvt.PtrToVal(info.SubUpdatedCPUCore),
		UpdatedMemory:    cvt.PtrToVal(info.SubUpdatedMemory),
		UpdatedDiskSize:  cvt.PtrToVal(info.SubUpdatedDiskSize),
		Demands:          demands,
		SubmittedAt:      info.SubmittedAt,
		Status:           info.Status,
		Stage:            info.Stage,
		AdminAuditStatus: info.AdminAuditStatus,
		CrpSN:            info.CrpSN,
		CrpURL:           info.CrpURL,
	}

	return brief, nil
}

// getSubTicketByID get sub ticket by id
func (f *ResPlanFetcher) getSubTicketByID(kt *kit.Kit, subTicketID string) (*rpst.ResPlanSubTicketTable, error) {
	listOpt := &rpproto.ResPlanSubTicketListReq{
		ListReq: core.ListReq{
			Filter: tools.EqualExpression("id", subTicketID),
			Page:   core.NewDefaultBasePage(),
		},
	}

	rst, err := f.client.DataService().Global.ResourcePlan.ListResPlanSubTicket(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan sub ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan sub ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan sub ticket, but len details != 1")
	}

	return &rst.Details[0], nil
}
