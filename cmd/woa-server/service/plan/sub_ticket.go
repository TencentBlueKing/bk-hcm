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
	"errors"
	"fmt"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListResPlanSubTicket list resource plan sub ticket
func (s *service) ListResPlanSubTicket(cts *rest.Contexts) (interface{}, error) {
	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	return s.listResPlanSubTicket(cts, constant.AttachedAllBiz, authorized)
}

// ListBizResPlanSubTicket list biz resource plan sub ticket
func (s *service) ListBizResPlanSubTicket(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanSubTicket(cts, bkBizID, true)
}

func (s *service) listResPlanSubTicket(cts *rest.Contexts, bkBizID int64, authorized bool) (interface{}, error) {
	req := new(ptypes.ListResPlanSubTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan sub ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource sub ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	req.BizID = bkBizID

	// 没有单据管理权限的只能查询自己的单据
	if !authorized {
		opt := &types.ListOption{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("id", req.TicketID),
				tools.RuleEqual("applicant", cts.Kit.User),
			),
			Page: core.NewCountPage(),
		}
		rst, err := s.dao.ResPlanTicket().List(cts.Kit, opt)
		if err != nil {
			logs.Errorf("failed to list resource plan ticket, err: %v, id: %s, applicant: %s, rid: %s", err,
				req.TicketID, cts.Kit.User, cts.Kit.Rid)
			return nil, err
		}
		if rst.Count == 0 {
			return nil, errf.NewFromErr(errf.PermissionDenied,
				fmt.Errorf("no permission to list sub_ticket for ticket: %s", req.TicketID))
		}
	}

	rst, err := s.planController.Fetch().ListResPlanSubTicket(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list resource plan sub ticket, err: %v, id: %s, applicant: %s, rid: %s", err,
			req.TicketID, cts.Kit.User, cts.Kit.Rid)
		return nil, err
	}

	for i := range rst.Details {
		// 屏蔽延期类型，将其改为调整类型
		if rst.Details[i].SubTicketType == enumor.RPTicketTypeDelay {
			rst.Details[i].SubTicketType = enumor.RPTicketTypeAdjust
			rst.Details[i].TicketTypeName = rst.Details[i].SubTicketType.Name()
		}
		// 屏蔽免审转移类型，将其改为转移类型
		if rst.Details[i].SubTicketType == enumor.RPTicketTypeTransferExempt {
			rst.Details[i].SubTicketType = enumor.RPTicketTypeTransfer
			rst.Details[i].TicketTypeName = rst.Details[i].SubTicketType.Name()
		}

		// RPSubTicketStatusWaiting 仅用于内部状态流转，对外统一展示为待审批
		if rst.Details[i].Status == enumor.RPSubTicketStatusWaiting {
			rst.Details[i].Status = enumor.RPSubTicketStatusInit
			rst.Details[i].StatusName = rst.Details[i].Status.Name()
		}
	}
	return rst, nil
}

// GetResPlanSubTicketDetail get resource plan sub ticket detail
func (s *service) GetResPlanSubTicketDetail(cts *rest.Contexts) (interface{}, error) {
	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	subTicketID := cts.PathParameter("sub_ticket_id").String()
	if len(subTicketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub_ticket_id can not be empty"))
	}

	detail, err := s.getResPlanSubTicketDetail(cts.Kit, constant.AttachedAllBiz, subTicketID, authorized)
	if err != nil {
		logs.Errorf("failed to get resource plan sub ticket detail, err: %v, id: %s, rid: %s", err, subTicketID,
			cts.Kit.Rid)
		return nil, err
	}
	return detail, nil
}

// GetBizResPlanSubTicketDetail get biz resource plan sub ticket detail
func (s *service) GetBizResPlanSubTicketDetail(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	subTicketID := cts.PathParameter("sub_ticket_id").String()
	if len(subTicketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub_ticket_id can not be empty"))
	}

	detail, err := s.getResPlanSubTicketDetail(cts.Kit, bkBizID, subTicketID, true)
	if err != nil {
		logs.Errorf("failed to get resource plan sub ticket detail, err: %v, id: %s, rid: %s", err, subTicketID,
			cts.Kit.Rid)
		return nil, err
	}
	return detail, nil
}

// getResPlanSubTicketDetail get resource plan sub ticket detail
func (s *service) getResPlanSubTicketDetail(kt *kit.Kit, bizID int64, subTicketID string, authorized bool) (
	*ptypes.GetSubTicketDetailResp, error) {

	detail, applicant, err := s.planController.Fetch().GetResPlanSubTicketDetail(kt, subTicketID)
	if err != nil {
		logs.Errorf("failed to get resource plan sub ticket detail, err: %v, id: %s, rid: %s", err, subTicketID,
			kt.Rid)
		return nil, err
	}

	// 没有单据管理权限的只能查询自己的单据
	if !authorized {
		if applicant != kt.User {
			return nil, errf.NewFromErr(errf.PermissionDenied, errors.New("no permission to access this ticket"))
		}
	}

	// 有业务参数时，单据业务需匹配
	if bizID != constant.AttachedAllBiz && detail.BaseInfo.BkBizID != bizID {
		return nil, errf.NewFromErr(errf.PermissionDenied, errors.New("no permission to access this ticket"))
	}

	// 屏蔽延期类型，将其改为调整类型
	if detail.BaseInfo.Type == enumor.RPTicketTypeDelay {
		detail.BaseInfo.Type = enumor.RPTicketTypeAdjust
		detail.BaseInfo.TypeName = detail.BaseInfo.Type.Name()
	}
	// 屏蔽免审转移类型，将其改为转移类型
	if detail.BaseInfo.Type == enumor.RPTicketTypeTransferExempt {
		detail.BaseInfo.Type = enumor.RPTicketTypeTransfer
		detail.BaseInfo.TypeName = detail.BaseInfo.Type.Name()
	}

	// RPSubTicketStatusWaiting 仅用于内部状态流转，对外统一展示为待审批
	if detail.StatusInfo.Status == enumor.RPSubTicketStatusWaiting {
		detail.StatusInfo.Status = enumor.RPSubTicketStatusInit
		detail.StatusInfo.StatusName = detail.StatusInfo.Status.Name()
	}

	return detail, nil
}

// GetResPlanSubTicketAudit get resource plan sub ticket audit
func (s *service) GetResPlanSubTicketAudit(cts *rest.Contexts) (interface{}, error) {
	// authorize ticket resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Application, Action: meta.Find}}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	return s.getResPlanSubTicketAudit(cts, constant.AttachedAllBiz, authorized)
}

// GetBizResPlanSubTicketAudit get biz resource plan sub ticket audit
func (s *service) GetBizResPlanSubTicketAudit(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.getResPlanSubTicketAudit(cts, bkBizID, true)
}

func (s *service) getResPlanSubTicketAudit(cts *rest.Contexts, bkBizID int64, authorized bool) (interface{}, error) {
	subTicketID := cts.PathParameter("sub_ticket_id").String()
	if len(subTicketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub_ticket_id can not be empty"))
	}

	auditResp, applicant, err := s.planController.Fetch().GetResPlanSubTicketAudit(cts.Kit, bkBizID, subTicketID)
	if err != nil {
		logs.Errorf("failed to get resource plan sub ticket audit, err: %v, id: %s, rid: %s", err, subTicketID,
			cts.Kit.Rid)
		return nil, err
	}

	// 没有单据管理权限的只能查询自己的单据
	if !authorized {
		if applicant != cts.Kit.User {
			return nil, errf.NewFromErr(errf.PermissionDenied, errors.New("no permission to access this ticket"))
		}
	}

	return auditResp, nil
}

// ApproveBizResPlanSubTicketAdminNode 业务下 审批预测单-管理员审批阶段
func (s *service) ApproveBizResPlanSubTicketAdminNode(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	subTicketID := cts.PathParameter("sub_ticket_id").String()
	if len(subTicketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub_ticket_id can not be empty"))
	}

	req := new(ptypes.AuditResPlanTicketAdminReq)
	if err := cts.DecodeInto(req); err != nil {
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

	if err := s.planController.ApproveResPlanSubTicketAdmin(cts.Kit, subTicketID, bkBizID, req); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ApproveResPlanSubTicketAdminNode 审批预测单-管理员审批阶段
func (s *service) ApproveResPlanSubTicketAdminNode(cts *rest.Contexts) (any, error) {
	subTicketID := cts.PathParameter("sub_ticket_id").String()
	if len(subTicketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub_ticket_id can not be empty"))
	}

	req := new(ptypes.AuditResPlanTicketAdminReq)
	if err := cts.DecodeInto(req); err != nil {
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

	err := s.planController.ApproveResPlanSubTicketAdmin(cts.Kit, subTicketID, constant.AttachedAllBiz, req)
	if err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
