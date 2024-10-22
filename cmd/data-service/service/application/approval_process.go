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

package application

import (
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableapplication "hcm/pkg/dal/table/application"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// InitApprovalProcessService ...
func InitApprovalProcessService(cap *capability.Capability) {
	svc := &approvalProcessSvc{
		dao: cap.Dao,
	}
	h := rest.NewHandler()

	h.Add("CreateApprovalProcesses", "POST", "/approval_processes/create", svc.CreateApprovalProcesses)
	h.Add("UpdateApprovalProcesses", "PATCH", "/approval_processes/{approval_process_id}",
		svc.UpdateApprovalProcesses)
	h.Add("ListApprovalProcesses", "POST", "/approval_processes/list", svc.ListApprovalProcesses)

	h.Load(cap.WebService)
}

type approvalProcessSvc struct {
	dao dao.Set
}

// CreateApprovalProcesses ...
func (svc *approvalProcessSvc) CreateApprovalProcesses(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ApprovalProcessCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	process := &tableapplication.ApprovalProcessTable{
		ApplicationType: string(req.ApplicationType),
		ServiceID:       req.ServiceID,
		Creator:         cts.Kit.User,
		Reviser:         cts.Kit.User,
	}

	approvalProcessID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		approvalProcessID, err := svc.dao.ApprovalProcess().CreateWithTx(cts.Kit, txn, process)
		if err != nil {
			return nil, fmt.Errorf("create approval process failed, err: %v", err)
		}
		return approvalProcessID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := approvalProcessID.(string)
	if !ok {
		return nil, fmt.Errorf("create approval process but return id type not string, id type: %v",
			reflect.TypeOf(approvalProcessID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

// UpdateApprovalProcesses ...
func (svc *approvalProcessSvc) UpdateApprovalProcesses(cts *rest.Contexts) (interface{}, error) {
	approvalProcessID := cts.PathParameter("approval_process_id").String()

	req := new(proto.ApprovalProcessUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	approvalProcess := &tableapplication.ApprovalProcessTable{
		ServiceID: req.ServiceID,
	}

	err := svc.dao.ApprovalProcess().Update(cts.Kit, tools.EqualExpression("id", approvalProcessID), approvalProcess)
	if err != nil {
		logs.Errorf("update approval process failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("update approval process failed, err: %v", err)
	}

	return nil, nil
}

func (svc *approvalProcessSvc) convertToApprovalProcessResp(
	approvalProcess *tableapplication.ApprovalProcessTable,
) *proto.ApprovalProcessResp {
	return &proto.ApprovalProcessResp{
		ID:              approvalProcess.ID,
		ApplicationType: enumor.ApplicationType(approvalProcess.ApplicationType),
		ServiceID:       approvalProcess.ServiceID,
		Managers:        approvalProcess.Managers,
		Revision: core.Revision{
			Creator:   approvalProcess.Creator,
			Reviser:   approvalProcess.Reviser,
			CreatedAt: approvalProcess.CreatedAt.String(),
			UpdatedAt: approvalProcess.UpdatedAt.String(),
		},
	}
}

// ListApprovalProcesses ...
func (svc *approvalProcessSvc) ListApprovalProcesses(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ApprovalProcessListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}
	daoApprovalProcessResp, err := svc.dao.ApprovalProcess().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list approval process failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list approval process failed, err: %v", err)
	}
	if req.Page.Count {
		return &proto.ApprovalProcessListResult{Count: daoApprovalProcessResp.Count}, nil
	}

	details := make([]*proto.ApprovalProcessResp, 0, len(daoApprovalProcessResp.Details))
	for _, approvalProcess := range daoApprovalProcessResp.Details {
		details = append(details, svc.convertToApprovalProcessResp(approvalProcess))
	}

	return &proto.ApprovalProcessListResult{Details: details}, nil
}
