/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"hcm/cmd/woa-server/logics/plan"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	mtypes "hcm/pkg/dal/dao/types/meta"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	cvt "hcm/pkg/tools/converter"
)

// ListResPlanTicket list resource plan ticket.
func (s *service) ListResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListResPlanTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// convert request to list option.
	opt, err := req.GenListOption()
	if err != nil {
		logs.Errorf("failed to convert to list option, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if !authorized {
		// 没有单据管理权限的只能查询自己的单据
		opt.Filter.Rules = append(opt.Filter.Rules, tools.RuleEqual("applicant", cts.Kit.User))
	}

	return s.planController.ListResPlanTicketWithRes(cts.Kit, opt)
}

// ListBizResPlanTicket list biz resource plan ticket.
func (s *service) ListBizResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(ptypes.ListBizResPlanTicketReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list biz resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate list biz resource ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// convert request to list option.
	listResPlanTicketReq := &ptypes.ListResPlanTicketReq{
		BkBizIDs:        []int64{bkBizID},
		TicketIDs:       req.TicketIDs,
		Statuses:        req.Statuses,
		ObsProjects:     req.ObsProjects,
		TicketTypes:     req.TicketTypes,
		Applicants:      req.Applicants,
		SubmitTimeRange: req.SubmitTimeRange,
		Page:            req.Page,
	}
	opt, err := listResPlanTicketReq.GenListOption()
	if err != nil {
		logs.Errorf("failed to convert to list option, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return s.planController.ListResPlanTicketWithRes(cts.Kit, opt)
}

// CreateBizResPlanTicket create biz resource plan ticket.
func (s *service) CreateBizResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.CreateResPlanTicketReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Create}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// 验证预测提报参数
	if err = s.validateResPlanTicket(req, bkBizID); err != nil {
		return nil, err
	}

	ticketID, err := s.createResPlanTicket(cts.Kit, bkBizID, req)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// create resource plan ticket itsm audit flow.
	if err = s.planController.CreateAuditFlow(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil
}

func (s *service) validateResPlanTicket(req *ptypes.CreateResPlanTicketReq, bkBizID int64) error {
	for _, item := range req.Demands {
		// 只允许931业务，提报滚服项目的预测
		if item.ObsProject == enumor.ObsProjectRollServer && bkBizID != enumor.ResourcePlanRollServerBiz {
			return errf.Newf(errf.InvalidParameter, "this business does not support rolling server project")
		}
	}
	return nil
}

// createResPlanTicket create resource plan ticket.
func (s *service) createResPlanTicket(kt *kit.Kit, bkBizID int64, req *ptypes.CreateResPlanTicketReq) (string, error) {
	// get create resource plan ticket needed zoneMap, regionAreaMap and deviceTypeMap.
	zoneMap, regionAreaMap, deviceTypeMap, err := s.planController.Fetch().GetMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// convert request demands to demands defined in resource plan ticket table.
	demands := make(rpt.ResPlanDemands, len(req.Demands))
	for idx, demand := range req.Demands {
		demands[idx] = rpt.ResPlanDemand{
			DemandClass: req.DemandClass,
			Updated: &rpt.UpdatedRPDemandItem{
				ObsProject:     demand.ObsProject,
				ExpectTime:     demand.ExpectTime,
				ReturnPlanTime: demand.ReturnPlanTime,
				ZoneID:         demand.ZoneID,
				ZoneName:       zoneMap[demand.ZoneID],
				RegionID:       demand.RegionID,
				RegionName:     regionAreaMap[demand.RegionID].RegionName,
				AreaName:       regionAreaMap[demand.RegionID].AreaName,
				DemandSource:   demand.DemandSource,
				Remark:         demand.Remark,
			},
		}

		if slices.Contains(demand.DemandResTypes, enumor.DemandResTypeCVM) {
			deviceType := demand.Cvm.DeviceType
			demands[idx].Updated.Cvm = rpt.Cvm{
				ResMode:        demand.Cvm.ResMode,
				DeviceType:     deviceType,
				DeviceClass:    deviceTypeMap[deviceType].DeviceClass,
				DeviceFamily:   deviceTypeMap[deviceType].DeviceFamily,
				TechnicalClass: deviceTypeMap[deviceType].TechnicalClass,
				CoreType:       deviceTypeMap[deviceType].CoreType,
				Os:             tabletypes.Decimal{Decimal: cvt.PtrToVal(demand.Cvm.Os)},
				CpuCore:        cvt.PtrToVal(demand.Cvm.CpuCore),
				Memory:         cvt.PtrToVal(demand.Cvm.Memory),
			}
		}

		if slices.Contains(demand.DemandResTypes, enumor.DemandResTypeCBS) {
			demands[idx].Updated.Cbs = rpt.Cbs{
				DiskType:     demand.Cbs.DiskType,
				DiskTypeName: demand.Cbs.DiskType.Name(),
				DiskIo:       cvt.PtrToVal(demand.Cbs.DiskIo),
				DiskSize:     cvt.PtrToVal(demand.Cbs.DiskSize),
			}
		}
	}

	// get biz org relation.
	bizOrgRel, err := s.bizLogics.GetBizOrgRel(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	logicsReq := &plan.CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeAdd,
		DemandClass: req.DemandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     demands,
		Remark:      req.Remark,
	}

	ticketID, err := s.planController.CreateResPlanTicket(kt, logicsReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return ticketID, nil
}

// getMetaNameMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap. map key is name
func (s *service) getMetaNameMaps(kt *kit.Kit) (map[string]string, map[string]mtypes.RegionArea, error) {
	zoneMap, regionAreaMap, _, err := s.planController.Fetch().GetMetaMaps(kt)
	if err != nil {
		return nil, nil, err
	}

	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)
	return zoneNameMap, regionNameMap, nil
}

func getMetaNameMapsFromIDMap(zoneMap map[string]string, regionAreaMap map[string]mtypes.RegionArea) (
	map[string]string, map[string]mtypes.RegionArea) {

	zoneNameMap := make(map[string]string)
	for id, name := range zoneMap {
		zoneNameMap[name] = id
	}
	regionNameMap := make(map[string]mtypes.RegionArea)
	for _, item := range regionAreaMap {
		regionNameMap[item.RegionName] = item
	}
	return zoneNameMap, regionNameMap
}

// GetResPlanTicket get resource plan ticket detail.
func (s *service) GetResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	ticketID := cts.PathParameter("id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	resp := new(ptypes.GetResPlanTicketResp)
	resp.ID = ticketID

	// get base info and demands.
	baseInfo, demands, err := s.getRPTicketBaseInfoAndDemands(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket base info and demands failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.BaseInfo = baseInfo
	resp.Demands = demands

	if !authorized {
		if baseInfo.Applicant != cts.Kit.User {
			return new(ptypes.GetResPlanTicketResp), nil
		}
	}

	// get status info.
	statusInfo, err := s.planController.GetResPlanTicketStatusInfo(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket status info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.StatusInfo = statusInfo

	return resp, nil
}

// getRPTicketBaseInfoAndDemands get resource plan ticket base information and demands.
func (s *service) getRPTicketBaseInfoAndDemands(kt *kit.Kit, ticketID string) (*ptypes.GetRPTicketBaseInfo,
	[]ptypes.GetRPTicketDemand, error) {

	// search resource plan ticket table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, nil, errors.New("list resource plan ticket, but len details != 1")
	}

	detail := rst.Details[0]
	baseInfo := &ptypes.GetRPTicketBaseInfo{
		Type:            detail.Type,
		TypeName:        detail.Type.Name(),
		Applicant:       detail.Applicant,
		BkBizID:         detail.BkBizID,
		BkBizName:       detail.BkBizName,
		OpProductID:     detail.OpProductID,
		OpProductName:   detail.OpProductName,
		PlanProductID:   detail.PlanProductID,
		PlanProductName: detail.PlanProductName,
		VirtualDeptID:   detail.VirtualDeptID,
		VirtualDeptName: detail.VirtualDeptName,
		DemandClass:     detail.DemandClass,
		Remark:          detail.Remark,
		SubmittedAt:     detail.SubmittedAt,
	}

	var demandsStruct rpt.ResPlanDemands
	if err = json.Unmarshal([]byte(rst.Details[0].Demands), &demandsStruct); err != nil {
		logs.Errorf("failed to unmarshal demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	demands := make([]ptypes.GetRPTicketDemand, len(demandsStruct))
	for idx, demand := range demandsStruct {
		demands[idx] = ptypes.GetRPTicketDemand{
			DemandClass:  demand.DemandClass,
			OriginalInfo: demand.Original,
			UpdatedInfo:  demand.Updated,
		}
	}

	return baseInfo, demands, nil
}

// GetBizResPlanTicket get biz resource plan ticket detail.
func (s *service) GetBizResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	ticketID := cts.PathParameter("id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	resp := new(ptypes.GetResPlanTicketResp)
	resp.ID = ticketID

	// get base info and demands.
	baseInfo, demands, err := s.getRPTicketBaseInfoAndDemands(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket base info and demands failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.BaseInfo = baseInfo
	resp.Demands = demands

	if baseInfo.BkBizID != bkBizID {
		logs.Errorf("ticket: %s is not belongs to bk_biz_id: %d, rid: %s", ticketID, bkBizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("ticket is not belongs to bk_biz_id: %d",
			bkBizID))
	}

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: baseInfo.BkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// get status info.
	statusInfo, err := s.planController.GetResPlanTicketStatusInfo(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket status info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.StatusInfo = statusInfo

	return resp, nil
}

// GetResPlanTicketAudit get biz resource plan ticket audit.
func (s *service) GetResPlanTicketAudit(cts *rest.Contexts) (interface{}, error) {
	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.planController.GetResPlanTicketAudit(cts.Kit, ticketID, constant.AttachedAllBiz)
}

// GetBizResPlanTicketAudit get biz resource plan ticket audit.
func (s *service) GetBizResPlanTicketAudit(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.planController.GetResPlanTicketAudit(cts.Kit, ticketID, bkBizID)
}

// ApproveBizResPlanTicketITSMNode 业务下 审批预测单对应itsm单据
func (s *service) ApproveBizResPlanTicketITSMNode(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	req := new(ptypes.AuditResPlanTicketITSMReq)
	if err := cts.DecodeInto(&req); err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		return nil, err
	}

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.approveResPlanTicketITSMByBiz(cts.Kit, ticketID, bkBizID, req)
}

// ApproveResPlanTicketITSMNode 审批预测单对应itsm单据
func (s *service) ApproveResPlanTicketITSMNode(cts *rest.Contexts) (any, error) {
	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	req := new(ptypes.AuditResPlanTicketITSMReq)
	if err := cts.DecodeInto(&req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.approveResPlanTicketITSMByBiz(cts.Kit, ticketID, constant.AttachedAllBiz, req)
}

func (s *service) approveResPlanTicketITSMByBiz(kt *kit.Kit, ticketID string, bizID int64,
	req *ptypes.AuditResPlanTicketITSMReq) (any, error) {

	// 查询数据
	status, err := s.planController.GetResPlanTicketStatusByBiz(kt, ticketID, bizID)
	if err != nil {
		logs.Errorf("failed to get resource plan ticket status info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// 校验状态
	if status.Status != enumor.RPTicketStatusAuditing {
		return nil, fmt.Errorf("ticket %s is not in auditing status", ticketID)
	}
	if len(status.ItsmSN) == 0 {
		return nil, fmt.Errorf("ITSM SN of ticket %s can not be found", ticketID)
	}
	// 进行审批
	approveReq := &itsm.ApproveNodeOpt{
		SN:       status.ItsmSN,
		StateId:  req.StateId,
		Operator: kt.User,
		Approval: cvt.PtrToVal(req.Approval),
		Remark:   req.Remark,
	}
	if err := s.planController.ApproveTicketITSMByBiz(kt, ticketID, approveReq); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return nil, nil
}

// RetryResPlanTicket 重试资源预测单中失败的子单
func (s *service) RetryResPlanTicket(cts *rest.Contexts) (any, error) {
	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := s.planController.RetryResPlanFailedSubTickets(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to retry res plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// RetryBizResPlanTicket 业务下 重试资源预测单中失败的子单
func (s *service) RetryBizResPlanTicket(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := s.planController.RetryResPlanFailedSubTickets(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to retry res plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TerminateResPlanTicket 终止失败的资源预测单，审批中的单据不可终止
func (s *service) TerminateResPlanTicket(cts *rest.Contexts) (any, error) {
	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := s.planController.TerminateResPlanFailedTicket(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to terminate res plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TerminateBizResPlanTicket 业务下 终止失败的资源预测单，审批中的单据不可终止
func (s *service) TerminateBizResPlanTicket(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	ticketID := cts.PathParameter("ticket_id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := s.planController.TerminateResPlanFailedTicket(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to terminate res plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
