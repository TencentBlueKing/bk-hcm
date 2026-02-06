/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package task scheduler
package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	model "hcm/cmd/woa-server/model/task"
	gctypes "hcm/cmd/woa-server/types/green-channel"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	coredevicetype "hcm/pkg/api/core/cloud/device-type"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

// UpdateBizApplyTicket update biz apply order
func (s *service) UpdateBizApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update biz apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BkBizId = bkBizID

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to update biz apply ticket, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to save apply ticket, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	return s.updateApplyTicket(cts.Kit, input)
}

// UpdateApplyTicket update apply order
func (s *service) UpdateApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to save apply ticket, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	return s.updateApplyTicket(cts.Kit, input)
}

// updateApplyTicket create or update apply ticket
func (s *service) updateApplyTicket(kt *kit.Kit, input *types.ApplyReq) (any, error) {
	rst, err := s.logics.Scheduler().UpdateApplyTicket(kt, input)
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, bkBizID: %d, rid: %s", err, input.BkBizId, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyTicket get biz apply ticket
func (s *service) GetBizApplyTicket(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply ticket permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	input := new(types.GetApplyTicketReq)
	if err = cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	input.BkBizID = bkBizID

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyTicket get apply ticket
func (s *service) GetApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyTicketReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyAuditItsm get biz apply audit
func (s *service) GetBizApplyAuditItsm(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply audit permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	input := new(types.GetApplyAuditItsmReq)
	if err = cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	input.BkBizID = bkBizID

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	return s.getApplyAuditItsm(cts.Kit, input)
}

// GetBizApplyAuditCrp get biz apply audit
func (s *service) GetBizApplyAuditCrp(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply ticket crp audit permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	req := new(types.GetApplyAuditCrpReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to get biz apply ticket crp audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to get biz apply ticket crp audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	recordFilter := make(map[string]interface{})
	recordFilter["suborder_id"] = req.SuborderId
	recordFilter["task_id"] = req.CrpTicketId
	page := metadata.BasePage{Start: 0, Limit: 1}
	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(cts.Kit.Ctx, page, recordFilter)
	if err != nil {
		logs.Errorf("failed to list generate records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(records) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("generate record not found"))
	}

	applyOrders, err := s.listApplyOrders(cts.Kit, page, withSuborderIDs(records[0].SubOrderId), withBizID(bkBizID))
	if err != nil {
		logs.Errorf("failed to list apply orders, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(applyOrders) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("apply order not found"))
	}

	// 校验主机申请单的业务ID
	if applyOrders[0].BkBizId != bkBizID {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("apply order not found"))
	}

	return s.getApplyAuditCrp(cts.Kit, req, applyOrders[0].ResourceType)
}

// GetApplyAuditItsm get apply audit
func (s *service) GetApplyAuditItsm(cts *rest.Contexts) (any, error) {
	req := new(types.GetApplyAuditItsmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ticket, err := s.listApplyTicket(cts.Kit, withOrderID(int64(req.OrderId)))
	if err != nil {
		logs.Errorf("failed to list apply orders, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: ticket.BkBizId,
	})
	if err != nil {
		logs.Errorf("failed to check get apply audit itsm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.getApplyAuditItsm(cts.Kit, req)
}

// GetApplyAuditCrp ...
func (s *service) GetApplyAuditCrp(cts *rest.Contexts) (interface{}, error) {
	req := new(types.GetApplyAuditCrpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	recordFilter := make(map[string]interface{})
	recordFilter["suborder_id"] = req.SuborderId
	recordFilter["task_id"] = req.CrpTicketId
	page := metadata.BasePage{Start: 0, Limit: 1}
	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(cts.Kit.Ctx, page, recordFilter)
	if err != nil {
		logs.Errorf("failed to list generate records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(records) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("generate record not found"))
	}

	applyOrders, err := s.listApplyOrders(cts.Kit, page, withSuborderIDs(records[0].SubOrderId))
	if err != nil {
		logs.Errorf("failed to list apply orders, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(applyOrders) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("apply order not found"))
	}

	resAttr := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: applyOrders[0].BkBizId,
	}
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, resAttr)
	if err != nil {
		logs.Errorf("failed to check get apply audit crp, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.getApplyAuditCrp(cts.Kit, req, applyOrders[0].ResourceType)
}

// getApplyAuditItsm get apply ticket audit info
func (s *service) getApplyAuditItsm(kt *kit.Kit, req *types.GetApplyAuditItsmReq) (*types.GetApplyAuditItsmRst, error) {
	rst, err := s.logics.Scheduler().GetApplyAuditItsm(kt, req)
	if err != nil {
		logs.Errorf("failed to get apply ticket itsm audit info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// getApplyAuditCrp get apply ticket audit info
func (s *service) getApplyAuditCrp(kt *kit.Kit, req *types.GetApplyAuditCrpReq, resType types.ResourceType) (any,
	error) {

	rst, err := s.logics.Scheduler().GetApplyAuditCrp(kt, req, resType)
	if err != nil {
		logs.Errorf("failed to get apply ticket crp audit info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// AuditBizApplyTicket 业务下审批ITSM单据
func (s *service) AuditBizApplyTicket(cts *rest.Contexts) (any, error) {

	bkBizId, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(types.BizApplyAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	ticket, err := s.listApplyTicket(cts.Kit, withOrderID(int64(req.OrderId)), withBizID(bkBizId))
	if err != nil {
		logs.Errorf("failed to get apply order for audit biz apply ticket, err: %v, biz: %d, order: %d, rid: %s",
			err, bkBizId, req.OrderId, cts.Kit.Rid)
		return nil, err
	}
	// 业务访问权限
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: ticket.BkBizId,
	})
	if err != nil {
		logs.Errorf("failed to check apply ticket perm for audit biz apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	auditParam := &types.ApplyAuditReq{
		OrderId:      req.OrderId,
		ItsmTicketId: ticket.ItsmTicketId,
		StateId:      req.StateId,
		Operator:     cts.Kit.User,
		Approval:     req.Approval,
		Remark:       req.Remark,
	}
	if err := s.logics.Scheduler().AuditTicket(cts.Kit, auditParam); err != nil {
		logs.Errorf("failed to audit biz apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// AuditApplyTicket audit apply ticket
func (s *service) AuditApplyTicket(cts *rest.Contexts) (any, error) {
	req := new(types.ResApplyAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	ticket, err := s.listApplyTicket(cts.Kit, withOrderID(int64(req.OrderId)))
	if err != nil {
		logs.Errorf("failed to get apply order for audit ticket, err: %v, order: %d, rid: %s",
			err, req.OrderId, cts.Kit.Rid)
		return nil, err
	}

	// 资源下主机申领权限
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: ticket.BkBizId,
	})
	if err != nil {
		logs.Errorf("failed to check apply ticket perm for audit res apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	auditParam := &types.ApplyAuditReq{
		OrderId:      req.OrderId,
		ItsmTicketId: ticket.ItsmTicketId,
		StateId:      req.StateId,
		// 默认使用当前用户
		Operator: cts.Kit.User,
		Approval: req.Approval,
		Remark:   req.Remark,
	}
	if err := s.logics.Scheduler().AuditTicket(cts.Kit, auditParam); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AutoAuditApplyTicket system automatic audit apply ticket
func (s *service) AutoAuditApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyAutoAuditReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().AutoAuditTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// ApproveApplyTicket approve or reject apply ticket
func (s *service) ApproveApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApproveApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().ApproveTicket(cts.Kit, input); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateBizApplyOrder create biz apply order
func (s *service) CreateBizApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create biz apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BkBizId = bkBizID

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to create biz apply ticket, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to create apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	return s.createApplyOrder(cts.Kit, input)
}

// CreateApplyOrder creates apply order
func (s *service) CreateApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to create apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	return s.createApplyOrder(cts.Kit, input)
}

func (s *service) validateDeviceTypeForGreenAndRoll(kt *kit.Kit, input *types.ApplyReq) error {
	if input.RequireType != enumor.RequireTypeGreenChannel && input.RequireType != enumor.RequireTypeRollServer {
		return nil
	}

	deviceTypes := make([]string, 0)
	for _, suborder := range input.Suborders {
		if suborder.Spec != nil {
			deviceTypes = append(deviceTypes, suborder.Spec.DeviceType)
		}
	}
	deviceTypes = slice.Unique(deviceTypes)

	deviceInfos := make([]coredevicetype.DistinctDeviceType, 0)
	for _, batch := range slice.Split(deviceTypes, int(core.DefaultMaxPageLimit)) {
		deviceReq := &protocloud.DistinctDeviceTypeListReq{
			ListReq: core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
					tools.RuleIn("device_type", batch)),
				Page: core.NewDefaultBasePage(),
			},
		}
		resp, err := s.configLogics.Device().ListDistinctDeviceType(kt, deviceReq)
		if err != nil {
			logs.Errorf("failed to get device type info, err: %v, deviceTypes: %v, rid: %s", err, batch, kt.Rid)
			return err
		}
		deviceInfos = append(deviceInfos, resp.Details...)
	}

	// 获取小额绿通的配置
	cvmApplyConfigs := gctypes.CvmApplyConfig{}
	if input.RequireType == enumor.RequireTypeGreenChannel {
		gcConfigs, err := s.gcLogics.GetConfigs(kt)
		if err != nil {
			logs.Errorf("get green channel configs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		cvmApplyConfigs = gcConfigs.CvmApplyConfig
	}

	var unsupportedTypes []string
	for _, item := range deviceInfos {
		if item.DeviceTypeClass == cvmapi.SpecialType {
			unsupportedTypes = append(unsupportedTypes, item.DeviceType)
		}
		// 小额绿通只能申请[标准型]、[16核以下]的机型
		if !(input.RequireType == enumor.RequireTypeGreenChannel && cvmApplyConfigs.Enabled) {
			continue
		}
		if !(slices.Contains(cvmApplyConfigs.DeviceGroups, item.DeviceFamily) && item.CpuCore <=
			cvmApplyConfigs.CpuMaxLimit) {
			unsupportedTypes = append(unsupportedTypes, item.DeviceType)
		}
	}
	if len(unsupportedTypes) > 0 {
		return fmt.Errorf("device types %v are not supported for green channel or roll server apply",
			slice.Unique(unsupportedTypes))
	}
	return nil
}

// createApplyOrder creates apply order
func (s *service) createApplyOrder(kt *kit.Kit, input *types.ApplyReq) (any, error) {
	if err := s.verifyAccordingToRequireType(kt, input); err != nil {
		logs.Errorf("failed to verify according to require type, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.verifyResPlanDemand(kt, input); err != nil {
		logs.Errorf("failed to verify res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().CreateApplyOrder(kt, input)
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

func (s *service) verifyAccordingToRequireType(kt *kit.Kit, input *types.ApplyReq) error {
	switch input.RequireType {
	case enumor.RequireTypeRollServer:
		if err := s.validateDeviceTypeForGreenAndRoll(kt, input); err != nil {
			return err
		}
		if err := s.verifyProhibitDAPrefixDeviceType(kt, input); err != nil {
			return err
		}
		return nil
	case enumor.RequireTypeGreenChannel:
		return s.validateDeviceTypeForGreenAndRoll(kt, input)
	case enumor.RequireTypeDissolve:
		return s.verifyBizDissolveQuota(kt, input)
	default:
		return nil
	}
}

func (s *service) verifyBizDissolveQuota(kt *kit.Kit, input *types.ApplyReq) error {
	if input.RequireType != enumor.RequireTypeDissolve {
		return nil
	}

	// 查询业务机房裁撤可申请额度
	bizID := input.BkBizId
	bizSummaryMap, err := s.dissolveLogics.Table().ListBizCpuCoreSummary(kt, []int64{input.BkBizId})
	if err != nil {
		logs.Errorf("list biz dissolve cpu core summary failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}
	summary, ok := bizSummaryMap[bizID]
	if !ok {
		logs.Errorf("can not find biz dissolve cpu core summary, bizID: %d, rid: %s", bizID, kt.Rid)
		return fmt.Errorf("can not find biz dissolve cpu core summary, bizID: %d", bizID)
	}
	dissolveQuota := summary.TotalCore - summary.DeliveredCore

	// 计算当前单据申请的CPU核数
	var appliedCore int64
	for _, subOrder := range input.Suborders {
		appliedCore += int64(subOrder.AppliedCore)
	}

	// 判断申请的额度是否大于可申请额度
	if appliedCore > dissolveQuota {
		logs.Errorf("applied cpu core more than biz dissolve quota, applied: %d, quota: %d, bizID: %d, rid: %s",
			appliedCore, dissolveQuota, bizID, kt.Rid)
		return fmt.Errorf("申请的CPU核数超过机房裁撤可申请额度, 申请CPU核数: %d, 可申请额度: %d", appliedCore,
			dissolveQuota)
	}

	return nil
}

func (s *service) verifyProhibitDAPrefixDeviceType(kt *kit.Kit, input *types.ApplyReq) error {
	for _, subOrder := range input.Suborders {
		if subOrder.Spec == nil {
			logs.Errorf("suborder spec is nil, rid: %s", kt.Rid)
			return errors.New("suborder spec is nil")
		}

		if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(subOrder.Spec.DeviceType)), "DA") {
			logs.Errorf("prohibit to apply for DA prefix device type, rid: %s", kt.Rid)
			return errors.New("prohibit to apply for DA prefix device type")
		}
	}
	return nil
}

// verifyResPlanDemand 资源预测余量校验
func (s *service) verifyResPlanDemand(kt *kit.Kit, input *types.ApplyReq) error {
	if !input.RequireType.NeedVerifyResPlan() {
		return nil
	}

	verifyBizID := input.BkBizId
	if input.RequireType.IsUseManageBizPlan() {
		verifyBizID = enumor.ResourcePoolBiz
	}

	subOrders := make([]types.Suborder, 0, len(input.Suborders))
	for _, subPtr := range input.Suborders {
		subOrders = append(subOrders, cvt.PtrToVal(subPtr))
	}

	planRst, err := s.planLogics.VerifyResPlanDemandV2(kt, verifyBizID, input.RequireType, subOrders)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand, err: %v, bk_biz_id: %d, rid: %s", err,
			verifyBizID, kt.Rid)
		return errf.NewFromErr(errf.ResPlanVerifyFailed, err)
	}

	for idx, verifyEle := range planRst {
		if verifyEle.VerifyResult != enumor.VerifyResPlanRstFailed {
			continue
		}
		errOrder := "failed to verify res plan demand"
		if len(subOrders) > idx {
			errOrder = fmt.Sprintf("suborder %d failed the resource plan demand verify", idx+1)
		}
		return errf.New(errf.ResPlanVerifyFailed, fmt.Sprintf("%s, reason: %s", errOrder,
			verifyEle.Reason))
	}

	return nil
}

// GetApplyBizOrder get biz apply order
func (s *service) GetApplyBizOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BkBizID = []int64{bkBizID}

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	return s.getApplyOrder(cts.Kit, input)
}

// GetApplyOrder get apply order
func (s *service) GetApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if len(input.Source) == 0 {
		input.Source = enumor.GetAllApplyTicketSource()
	}

	return s.getApplyOrder(cts.Kit, input)
}

// getApplyOrder gets apply order info
func (s *service) getApplyOrder(kt *kit.Kit, input *types.GetApplyParam) (any, error) {
	// 主机申领-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bkBizID := range input.BkBizID {
		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bkBizID,
		})
	}
	err := s.authorizer.AuthorizeWithPerm(kt, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get apply order, inputBizIDs: %v, err: %v, rid: %s",
			input.BkBizID, err, kt.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(kt, input)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyOrder gets given business's apply order info
func (s *service) GetBizApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetBizApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// check permission
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizID,
	})
	if err != nil {
		logs.Errorf("no permission to get biz apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	param := &types.GetApplyParam{
		BkBizID: []int64{input.BkBizID},
		Start:   input.Start,
		End:     input.End,
		Page:    input.Page,
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(cts.Kit, param)
	if err != nil {
		logs.Errorf("failed to get biz apply order, param: %+v, err: %v, rid: %s", param, err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyStatus gets apply order status
func (s *service) GetApplyStatus(cts *rest.Contexts) (any, error) {
	orderId, err := strconv.Atoi(cts.Request.PathParameter("order_id"))
	if err != nil {
		logs.Errorf("failed to get apply order status, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if orderId <= 0 {
		logs.Errorf("failed to get apply order status, for invalid order id %d <= 0, rid: %s", orderId, cts.Kit.Rid)
		return nil, errf.Newf(pkg.CCErrCommParamsIsInvalid, "order_id")
	}

	input := &types.GetApplyParam{
		OrderID: []uint64{uint64(orderId)},
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order status, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyDetail get biz apply detail
func (s *service) GetBizApplyDetail(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply detail permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetApplyDetail(cts)
}

// GetApplyDetail gets apply order detail info
func (s *service) GetApplyDetail(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyDetailReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyGenerate get biz apply generate
func (s *service) GetBizApplyGenerate(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply generate permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetApplyGenerate(cts)
}

// GetApplyGenerate gets apply order generate records
func (s *service) GetApplyGenerate(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyGenerateReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyGenerate(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyInit get biz apply init
func (s *service) GetBizApplyInit(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply init permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetApplyInit(cts)
}

// GetApplyInit gets apply order init records
func (s *service) GetApplyInit(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyInitReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply init record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply init record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyInit(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply init record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDiskCheck gets apply order disk check records
func (s *service) GetApplyDiskCheck(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyInitReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyDiskCheck(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyDeliver get biz apply deliver
func (s *service) GetBizApplyDeliver(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz apply deliver permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetApplyDeliver(cts)
}

// GetApplyDeliver gets apply order deliver records
func (s *service) GetApplyDeliver(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyDeliverReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyDeliver(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyDevice create biz apply device
func (s *service) GetBizApplyDevice(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.getApplyDevice(cts, bkBizIDMap)
}

// GetApplyDevice get apply order delivered devices
func (s *service) GetApplyDevice(cts *rest.Contexts) (any, error) {
	return s.getApplyDevice(cts, make(map[int64]struct{}))
}

// getApplyDevice get apply order delivered devices
func (s *service) getApplyDevice(cts *rest.Contexts, bkBizIDMap map[int64]struct{}) (any, error) {
	input := new(types.GetApplyDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// 解析参数里的业务ID，用于鉴权，是必传参数
	bkBizIDs, err := s.parseInputForBkBizID(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to parse input for bizID, err: %+v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	// 主机申领-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bizID := range bkBizIDs {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizID]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizID, bkBizIDMap)
		}

		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bizID,
		})
	}
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get apply device, bizIDs: %v, err: %v, rid: %s", bkBizIDs, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

func (s *service) parseInputForBkBizID(kt *kit.Kit, input *types.GetApplyDeviceReq) ([]int64, error) {
	filterMap, err := input.GetFilter()
	if err != nil {
		logs.Errorf("failed to parse input filter, err: %v, input: %+v, rid: %s", err, input, kt.Rid)
		return nil, err
	}

	var bkBizIDs []int64
	paramMap, ok := filterMap["$and"].([]map[string]interface{})
	if !ok {
		return nil, errf.Newf(errf.InvalidParameter, "filter is illegal")
	}

	for _, paramItem := range paramMap {
		condMap, ok := paramItem["bk_biz_id"]
		if !ok {
			continue
		}
		// 如果找到了业务ID，但解析失败则break
		fieldMap, ok := condMap.(map[string]interface{})
		if !ok {
			break
		}
		numbers, ok := fieldMap["$in"].([]interface{})
		if !ok {
			logs.Errorf("bk_biz_id value is not []interface, fieldMap: %+v, rid: %s", fieldMap, kt.Rid)
			return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id is illegal")
		}

		for _, val := range numbers {
			number, ok := val.(json.Number)
			if !ok {
				logs.Errorf("bk_biz_id value is not json.Number, val: %+v, valType: %+v, rid: %s",
					val, reflect.TypeOf(val), kt.Rid)
				return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id value is not json.Number")
			}
			bkBizID, err := number.Int64()
			if err != nil {
				logs.Errorf("bk_biz_id value is not int64, number: %+v, valType: %+v, err: %v, rid: %s",
					number, reflect.TypeOf(number), err, kt.Rid)
				return nil, err
			}
			bkBizIDs = append(bkBizIDs, bkBizID)
		}
		break
	}

	if len(bkBizIDs) <= 0 {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id is required")
	}

	return bkBizIDs, nil
}

// GetDeliverDeviceByOrder get delivered devices by order id
func (s *service) GetDeliverDeviceByOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetDeliverDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply delivered device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules: []querybuilder.Rule{
			querybuilder.AtomRule{
				Field:    "order_id",
				Operator: querybuilder.OperatorEqual,
				Value:    input.OrderId,
			}},
	}
	if len(input.SuborderId) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "suborder_id",
			Operator: querybuilder.OperatorEqual,
			Value:    input.SuborderId,
		})
	}
	param := &types.GetApplyDeviceReq{
		Filter: &querybuilder.QueryFilter{
			Rule: rule,
		},
	}

	rst, err := s.logics.Scheduler().GetApplyDevice(cts.Kit, param)
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	type deviceBriefInfo struct {
		Ip      string `json:"ip" bson:"ip"`
		AssetId string `json:"asset_id" bson:"asset_id"`
	}
	type getDeviceBriefRst struct {
		Count int64              `json:"count"`
		Info  []*deviceBriefInfo `json:"info"`
	}

	briefRst := &getDeviceBriefRst{
		Count: int64(len(rst.Info)),
		Info:  make([]*deviceBriefInfo, 0),
	}
	for _, device := range rst.Info {
		briefRst.Info = append(briefRst.Info, &deviceBriefInfo{
			Ip:      device.Ip,
			AssetId: device.AssetId,
		})
	}

	return briefRst, nil
}

// ExportDeliverDevice export delivered devices
func (s *service) ExportDeliverDevice(cts *rest.Contexts) (any, error) {
	input := new(types.ExportDeliverDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.Scheduler().ExportDeliverDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizMatchDevice get biz match device
func (s *service) GetBizMatchDevice(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check get biz match device permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetMatchDevice(cts)
}

// GetMatchDevice get apply order match devices
func (s *service) GetMatchDevice(cts *rest.Contexts) (any, error) {
	input := new(types.GetMatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetMatchDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// MatchBizDevice match biz device
func (s *service) MatchBizDevice(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check match biz device permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.MatchDevice(cts)
}

// MatchDevice execute apply order match devices
func (s *service) MatchDevice(cts *rest.Contexts) (any, error) {
	input := new(types.MatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to match devices, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().MatchDevice(cts.Kit, input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// MatchBizPoolDevice match biz pool device
func (s *service) MatchBizPoolDevice(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check match biz pool device permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.MatchPoolDevice(cts)
}

// MatchPoolDevice execute apply order match devices from resource pool
func (s *service) MatchPoolDevice(cts *rest.Contexts) (any, error) {
	input := new(types.MatchPoolDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to match pool devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to match pool devices, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err = s.logics.Scheduler().MatchPoolDevice(cts.Kit, input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// PauseApplyOrder pauses apply order
func (s *service) PauseApplyOrder(_ *rest.Contexts) (any, error) {
	// TODO
	return nil, nil
}

// ResumeApplyOrder resumes apply order
func (s *service) ResumeApplyOrder(_ *rest.Contexts) (any, error) {
	// TODO
	return nil, nil
}

// StartBizApplyOrder start biz apply order
func (s *service) StartBizApplyOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.startApplyOrder(cts, bkBizIDMap, meta.Biz, meta.Create)
}

// StartApplyOrder 主机申请重试
func (s *service) StartApplyOrder(cts *rest.Contexts) (any, error) {
	return s.startApplyOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Create)
}

// startApplyOrder 主机申请重试
func (s *service) startApplyOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.StartApplyOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to start apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to start apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to start apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(pkg.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to start apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to start apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
				bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().StartApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to start recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TerminateBizApplyOrder terminate biz apply order
func (s *service) TerminateBizApplyOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.terminateApplyOrder(cts, bkBizIDMap, meta.Biz, meta.Create)
}

// TerminateApplyOrder terminate apply order
func (s *service) TerminateApplyOrder(cts *rest.Contexts) (any, error) {
	return s.terminateApplyOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Create)
}

// terminateApplyOrder terminate apply order
func (s *service) terminateApplyOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.TerminateApplyOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to terminate apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to terminate apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to terminate apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(pkg.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to terminate apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to terminate apply order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().TerminateApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to terminate recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ModifyBizApplyOrder modify biz apply order
func (s *service) ModifyBizApplyOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.modifyApplyOrder(cts, bkBizIDMap, meta.Biz, meta.Create)
}

// ModifyApplyOrder 修改需求重试
func (s *service) ModifyApplyOrder(cts *rest.Contexts) (any, error) {
	return s.modifyApplyOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Create)
}

// modifyApplyOrder modify apply order
func (s *service) modifyApplyOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.ModifyApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	suborderIDs := []string{input.SuborderID}
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, suborderIDs)
	if err != nil {
		logs.Errorf("failed to modify apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(pkg.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to modify apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
				bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().ModifyApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to modify recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetBizApplyModify get biz apply modify
func (s *service) GetBizApplyModify(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check get biz apply modify permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetApplyModify(cts)
}

// GetApplyModify get apply order modification records
func (s *service) GetApplyModify(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyModifyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyModify(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 转换旧版可用区到新版可用区，方便前端统一展示
	for _, modifyItem := range rst.Info {
		if modifyItem.Details == nil {
			continue
		}

		// 只有旧版可用区，才需要转换
		if modifyItem.Details.PreData != nil && len(modifyItem.Details.PreData.Zones) == 0 {
			modifyItem.Details.PreData.Zones = []string{modifyItem.Details.PreData.Zone}
			// 分Campus
			if modifyItem.Details.PreData.Zone == cvmapi.CvmSeparateCampus {
				modifyItem.Details.PreData.Zones = []string{cvmapi.CvmZoneAll}
				modifyItem.Details.PreData.ResAssign = enumor.CampusResAssign
			}
		}

		if modifyItem.Details.CurData != nil && len(modifyItem.Details.CurData.Zones) == 0 {
			modifyItem.Details.CurData.Zones = []string{modifyItem.Details.CurData.Zone}
			// 分Campus
			if modifyItem.Details.CurData.Zone == cvmapi.CvmSeparateCampus {
				modifyItem.Details.CurData.Zones = []string{cvmapi.CvmZoneAll}
				modifyItem.Details.CurData.ResAssign = enumor.CampusResAssign
			}
		}
	}

	return rst, nil
}

// getApplyOrderBizIds get apply order biz ids
func (s *service) getApplyOrderBizIds(kit *kit.Kit, suborderIds []string) ([]int64, error) {
	filter := map[string]interface{}{}

	if len(suborderIds) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: suborderIds,
		}
	}

	bizIds := make([]int64, 0)
	page := metadata.BasePage{
		Start: 0,
		Limit: 500,
	}
	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return bizIds, err
	}

	for _, inst := range insts {
		bizIds = append(bizIds, inst.BkBizId)
	}

	bizIds = util.IntArrayUnique(bizIds)

	return bizIds, nil
}

// filterOptFunc list apply orders
type filterOptFunc func(filter map[string]interface{})

// withOrderID filter apply order by order id
func withOrderID(orderID int64) filterOptFunc {
	return func(filter map[string]interface{}) {
		filter["order_id"] = orderID
	}
}

// withSuborderIDs filter apply order by suborder ids
func withSuborderIDs(suborderIDs ...string) filterOptFunc {
	return func(filter map[string]interface{}) {
		if len(suborderIDs) == 0 {
			return
		}

		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: suborderIDs,
		}
	}
}

// withBizID filter apply order by biz id
func withBizID(bizID int64) filterOptFunc {
	return func(filter map[string]interface{}) {
		filter["bk_biz_id"] = bizID
	}
}

// listApplyOrders list apply orders
func (s *service) listApplyOrders(kit *kit.Kit, page metadata.BasePage, filterOptFns ...filterOptFunc) (
	[]*types.ApplyOrder, error) {

	filter := map[string]interface{}{}
	for _, fn := range filterOptFns {
		fn(filter)
	}

	orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// listApplyOrders list apply orders
func (s *service) listApplyTicket(kit *kit.Kit, filterOptFns ...filterOptFunc) (*types.ApplyTicket, error) {
	filter := mapstr.MapStr{}
	for _, fn := range filterOptFns {
		fn(filter)
	}

	ticket, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// CheckRollingServerHost check rolling server host
func (s *service) CheckRollingServerHost(cts *rest.Contexts) (any, error) {
	input := new(types.CheckRollingServerHostReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("check rolling server host failed, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().CheckRollingServerHost(cts.Kit, input)
	if err != nil {
		logs.Errorf("check rolling server host failed, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CancelApplyTicketItsm ...
func (s *service) CancelApplyTicketItsm(cts *rest.Contexts) (interface{}, error) {
	req := new(types.CancelApplyTicketItsmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ticket, err := s.listApplyTicket(cts.Kit, withOrderID(req.OrderID))
	if err != nil {
		logs.Errorf("failed to list apply orders, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: ticket.BkBizId,
	})
	if err != nil {
		logs.Errorf("failed to check cancel apply ticket itsm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Scheduler().CancelApplyTicketItsm(cts.Kit, req); err != nil {
		logs.Errorf("failed to cancel apply ticket itsm, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// CancelBizApplyTicketItsm ...
func (s *service) CancelBizApplyTicketItsm(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check cancel biz apply ticket itsm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(types.CancelApplyTicketItsmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Scheduler().CancelApplyTicketItsm(cts.Kit, req); err != nil {
		logs.Errorf("failed to cancel biz apply ticket itsm, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// CancelApplyTicketCrp ...
func (s *service) CancelApplyTicketCrp(cts *rest.Contexts) (interface{}, error) {
	req := new(types.CancelApplyTicketCrpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}
	applyOrders, err := s.listApplyOrders(cts.Kit, page, withSuborderIDs(req.SubOrderID))
	if err != nil {
		logs.Errorf("failed to list apply orders, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	resAttrs := make([]meta.ResourceAttribute, 0)
	for _, applyOrder := range applyOrders {
		resAttrs = append(resAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: applyOrder.BkBizId,
		})
	}

	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, resAttrs...); err != nil {
		logs.Errorf("failed to check cancel apply ticket itsm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Scheduler().CancelApplyTicketCrp(cts.Kit, req); err != nil {
		logs.Errorf("failed to cancel apply ticket crp, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CancelBizApplyTicketCrp ...
func (s *service) CancelBizApplyTicketCrp(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check cancel biz apply ticket crp, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	req := new(types.CancelApplyTicketCrpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Scheduler().CancelApplyTicketCrp(cts.Kit, req); err != nil {
		logs.Errorf("failed to cancel apply ticket crp, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListApplyAuditInfo list apply audit info
func (s *service) ListApplyAuditInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(types.ListApplyAuditInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	details := make([]types.ListApplyAuditInfo, len(req.TicketIDs))
	for i, ticketID := range req.TicketIDs {
		auditInfo, err := s.getApplyAuditItsm(cts.Kit, &types.GetApplyAuditItsmReq{OrderId: ticketID})
		if err != nil {
			logs.Errorf("failed to get apply audit itsm, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		status := itsm.Status(auditInfo.Status)
		if err = status.Validate(); err != nil {
			logs.Errorf("status is invalid, err: %v, ticket id: %d, rid: %s", err, ticketID, cts.Kit.Rid)
			return nil, fmt.Errorf("status is invalid, err: %v, ticket id: %d", err, ticketID)
		}
		rst, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, &types.GetApplyTicketReq{OrderId: ticketID})
		if err != nil {
			logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		if rst == nil || rst.ApplyTicket == nil {
			logs.Errorf("can not find apply ticket, ticket id: %d, rid: %s", ticketID, cts.Kit.Rid)
			return nil, fmt.Errorf("can not find apply ticket, ticket id: %d", ticketID)
		}
		details[i] = types.ListApplyAuditInfo{
			TicketID:     ticketID,
			Status:       status,
			TicketInfo:   rst,
			CurrentSteps: auditInfo.CurrentSteps,
		}
		if !status.IsFinalState() {
			continue
		}
		if len(auditInfo.Logs) == 0 {
			logs.Errorf("audit logs is empty, ticket id: %d, rid: %s", ticketID, cts.Kit.Rid)
			return nil, fmt.Errorf("audit logs is empty, ticket id: %d", ticketID)
		}
		endAt, err := time.Parse(constant.DateTimeLayout, auditInfo.Logs[len(auditInfo.Logs)-1].OperateAt)
		if err != nil {
			logs.Errorf("failed to parse end at, err: %v, ticket id: %d, rid: %s", err, ticketID, cts.Kit.Rid)
			return nil, fmt.Errorf("failed to parse end at, err: %v, ticket id: %d", err, ticketID)
		}
		details[i].EndAt = cvt.ValToPtr(endAt)
	}

	return types.ListApplyAuditInfoResp{Details: details}, nil
}

// ApproveApplyTicketNode approve or reject apply ticket node
func (s *service) ApproveApplyTicketNode(cts *rest.Contexts) (any, error) {
	req := new(types.ApproveApplyTicketNodeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	auditInfo, err := s.getApplyAuditItsm(cts.Kit, &types.GetApplyAuditItsmReq{OrderId: req.TicketID})
	if err != nil {
		logs.Errorf("failed to get apply audit itsm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	status := itsm.Status(auditInfo.Status)
	if err = status.Validate(); err != nil {
		logs.Errorf("status is invalid, err: %v, ticket id: %d, rid: %s", err, req.TicketID, cts.Kit.Rid)
		return nil, fmt.Errorf("status is invalid, err: %v, ticket id: %d", err, req.TicketID)
	}
	if status.IsFinalState() {
		return nil, nil
	}

	param := itsm.ApproveNodeOpt{
		SN:       auditInfo.ItsmTicketId,
		StateId:  req.StateID,
		Operator: req.Operator,
		Approval: cvt.PtrToVal(req.Approval),
		Remark:   req.Remark,
	}
	if err = s.itsmClient.ApproveNode(cts.Kit, &param); err != nil {
		logs.Errorf("failed to approve itsm node, err: %v, param: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// FindApproveNodeResult find approve node result
func (s *service) FindApproveNodeResult(cts *rest.Contexts) (any, error) {
	req := new(types.FindApproveNodeResultReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, &types.GetApplyTicketReq{OrderId: req.TicketID})
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if rst == nil || rst.ApplyTicket == nil {
		logs.Errorf("can not find apply ticket, ticket id: %d, rid: %s", req.TicketID, cts.Kit.Rid)
		return nil, fmt.Errorf("can not find apply ticket, ticket id: %d", req.TicketID)
	}

	result, err := s.itsmClient.GetApproveNodeResult(cts.Kit, rst.ItsmTicketId, req.StateID)
	if err != nil {
		logs.Errorf("failed to find approve node result, err: %v, ticket id: %s, state id: %d, rid: %s",
			err, rst.ItsmTicketId, req.StateID, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListHostApplyItsmTicket list host apply itsm ticket
func (s *service) ListHostApplyItsmTicket(cts *rest.Contexts) (any, error) {
	req := new(types.ListHostApplyItsmTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	filter := mapstr.MapStr{
		"create_at": mapstr.MapStr{"$gte": req.CreateTime},
		"stage":     types.TicketStageAudit,
	}
	page := metadata.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit}
	tickets := make([]*types.ApplyTicket, 0)
	for {
		resp, err := model.Operation().ApplyTicket().FindManyApplyTicket(cts.Kit.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get apply ticket, err: %v, filter: %v, rid: %s", err, filter, cts.Kit.Rid)
			return nil, err
		}
		tickets = append(tickets, resp...)

		if len(resp) < page.Limit {
			break
		}
		page.Start += page.Limit
	}
	if len(tickets) == 0 {
		return types.ListHostApplyItsmTicketData{Tickets: make([]types.HostApplyItsmTicket, 0)}, nil
	}

	itsmTickets := make([]types.HostApplyItsmTicket, 0)
	for _, ticket := range tickets {
		auditInfo, err := s.getApplyAuditItsm(cts.Kit, &types.GetApplyAuditItsmReq{OrderId: ticket.OrderId})
		if err != nil {
			logs.Errorf("failed to get itsm audit, err: %v, ticket id: %d, rid: %s", err, ticket.OrderId, cts.Kit.Rid)
			return nil, err
		}
		status := itsm.Status(auditInfo.Status)
		if err = status.Validate(); err != nil {
			logs.Errorf("status is invalid, err: %v, ticket id: %d, rid: %s", err, ticket.OrderId, cts.Kit.Rid)
			return nil, fmt.Errorf("status is invalid, err: %v, ticket id: %d", err, ticket.OrderId)
		}
		if status.IsFinalState() || len(auditInfo.CurrentSteps) == 0 {
			continue
		}

		stepName := enumor.HostApplyItsmStepName(auditInfo.CurrentSteps[0].Name)
		if err = stepName.Validate(); err != nil {
			logs.Errorf("step name is invalid, err: %v, ticket id: %d, stepName: %s, rid: %s", err, ticket.OrderId,
				stepName, cts.Kit.Rid)
			continue
		}

		itsmTickets = append(itsmTickets, types.HostApplyItsmTicket{
			ID:            auditInfo.ItsmTicketId,
			Url:           auditInfo.ItsmTicketLink,
			User:          ticket.User,
			ApprovalState: stepName.GetApprovalState(),
			CreateTime:    ticket.CreateAt,
		})
	}

	return types.ListHostApplyItsmTicketData{Tickets: itsmTickets}, nil
}

// ListHostApplyCrpTicket list host apply crp ticket
func (s *service) ListHostApplyCrpTicket(cts *rest.Contexts) (any, error) {
	req := new(types.ListHostApplyCrpTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	subOrderMap, crpIDSubOrderIDMap, err := s.getRunningSubOrderInfo(cts.Kit, req.CreateTime)
	if err != nil {
		logs.Errorf("failed to get running sub order info, err: %v, create time: %v, rid: %s", err, req.CreateTime,
			cts.Kit.Rid)
		return nil, err
	}
	if len(subOrderMap) == 0 || len(crpIDSubOrderIDMap) == 0 {
		return types.ListHostApplyCrpTicketData{Tickets: make([]types.HostApplyCrpTicket, 0)}, nil
	}

	tickets := make([]types.HostApplyCrpTicket, 0)
	for crpID, subOrderID := range crpIDSubOrderIDMap {
		subOrder, ok := subOrderMap[subOrderID]
		if !ok {
			continue
		}
		crpReq := types.GetApplyAuditCrpReq{CrpTicketId: crpID, SuborderId: subOrderID}
		auditCrp, err := s.logics.Scheduler().GetApplyAuditCrp(cts.Kit, &crpReq, subOrder.ResourceType)
		if err != nil {
			logs.Errorf("failed to get apply ticket crp audit info, err: %v, req: %v, resourceType: %s, rid: %s", err,
				crpReq, subOrder.ResourceType, cts.Kit.Rid)
			continue
		}
		crpOrderStatus := enumor.CrpOrderStatus(auditCrp.CurrentStep.Status)
		if !crpOrderStatus.IsAdminApproval() {
			continue
		}

		tickets = append(tickets, types.HostApplyCrpTicket{
			ID:            auditCrp.CrpTicketId,
			Url:           auditCrp.CrpTicketLink,
			User:          subOrder.User,
			ApprovalState: crpOrderStatus.GetApprovalState(),
			CreateTime:    subOrder.CreateAt,
		})
	}

	return types.ListHostApplyCrpTicketData{Tickets: tickets}, nil
}

func (s *service) getRunningSubOrderInfo(kt *kit.Kit, createTime *time.Time) (map[string]*types.ApplyOrder,
	map[string]string, error) {

	subOrderFilter := mapstr.MapStr{
		"create_at": mapstr.MapStr{"$gte": createTime},
		"stage":     types.TicketStageRunning,
	}
	subOrderPage := metadata.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit}
	subOrderMap := make(map[string]*types.ApplyOrder)
	for {
		orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, subOrderPage, subOrderFilter)
		if err != nil {
			logs.Errorf("failed to find apply orders, err: %v, filter: %v, rid: %s", err, subOrderFilter, kt.Rid)
			return nil, nil, err
		}
		for _, order := range orders {
			subOrderMap[order.SubOrderId] = order
		}
		if len(orders) < subOrderPage.Limit {
			break
		}
		subOrderPage.Start += subOrderPage.Limit
	}
	if len(subOrderMap) == 0 {
		return make(map[string]*types.ApplyOrder), make(map[string]string), nil
	}

	genRecordFilter := mapstr.MapStr{
		"suborder_id": mapstr.MapStr{"$in": maps.Keys(subOrderMap)},
		"status":      types.GenerateStatusHandling,
	}
	genRecordPage := metadata.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit}
	crpIDSubOrderIDMap := make(map[string]string)
	for {
		genRecords, err := model.Operation().GenerateRecord().FindManyGenerateRecord(kt.Ctx, genRecordPage,
			genRecordFilter)
		if err != nil {
			logs.Errorf("failed to find generate records, err: %v, filter: %v, rid: %s", err, genRecordFilter,
				kt.Rid)
			return nil, nil, err
		}
		for _, genRecord := range genRecords {
			if genRecord.TaskId == "" {
				continue
			}
			crpIDSubOrderIDMap[genRecord.TaskId] = genRecord.SubOrderId
		}
		if len(genRecords) < genRecordPage.Limit {
			break
		}
		genRecordPage.Start += genRecordPage.Limit
	}

	return subOrderMap, crpIDSubOrderIDMap, nil
}

// UpdateApplyTicketDemand update apply ticket demand
func (s *service) UpdateApplyTicketDemand(cts *rest.Contexts) (any, error) {
	req := new(types.UpdateApplyTicketDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ticket, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, &types.GetApplyTicketReq{OrderId: req.TicketID})
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, ticket id: %d, rid: %s", err, req.TicketID, cts.Kit.Rid)
		return nil, err
	}
	if ticket == nil || ticket.ApplyTicket == nil {
		logs.Errorf("can not find apply ticket, ticket id: %d, rid: %s", req.TicketID, cts.Kit.Rid)
		return nil, fmt.Errorf("can not find apply ticket, ticket id: %d", req.TicketID)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: ticket.BkBizId,
	}); err != nil {
		logs.Errorf("no permission to update apply ticket demand, bizID: %d, err: %v, rid: %s", ticket.BkBizId,
			err, cts.Kit.Rid)
		return nil, err
	}

	if ticket.Stage != types.TicketStageAudit {
		logs.Errorf("failed to update apply ticket demand, for invalid stage: %s != %s, ticket id: %d, rid: %s",
			ticket.Stage, types.TicketStageAudit, req.TicketID, cts.Kit.Rid)
		return nil, fmt.Errorf("invalid ticket stage:%s != %s", ticket.Stage, types.TicketStageAudit)
	}

	for index, suborder := range req.Suborders {
		if suborder.ResourceType != types.ResourceTypeCvm {
			return nil, fmt.Errorf("resource type is not cvm, ticket id: %d, suborder index: %d", req.TicketID, index)
		}
	}

	// 如果oldSuborders为空，代表第一次更新需求，需要记录原始需求
	if ticket.OldSuborders == nil {
		ticket.OldSuborders = ticket.Suborders
	}
	ticket.Suborders = req.Suborders

	if err := s.logics.Scheduler().UpdateApplyTicketDemand(cts.Kit, ticket.ApplyTicket); err != nil {
		logs.Errorf("failed to update apply ticket demand, err: %v, ticket id: %d, rid: %s", err, req.TicketID,
			cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}
