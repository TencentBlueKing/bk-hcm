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

// Package application ...
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
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// InitApplicationService ...
func InitApplicationService(cap *capability.Capability) {
	svc := &applicationSvc{
		dao: cap.Dao,
	}
	h := rest.NewHandler()

	h.Add("Create", "POST", "/applications/create", svc.Create)
	h.Add("Update", "PATCH", "/applications/{application_id}", svc.Update)
	h.Add("Get", "GET", "/applications/{application_id}", svc.Get)
	h.Add("List", "POST", "/applications/list", svc.List)

	h.Load(cap.WebService)
}

type applicationSvc struct {
	dao dao.Set
}

func (svc *applicationSvc) Create(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ApplicationCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	application := &tableapplication.ApplicationTable{
		SN:             req.SN,
		Source:         string(req.Source),
		Type:           string(req.Type),
		Status:         string(req.Status),
		BkBizIDs:       req.BkBizIDs,
		Applicant:      cts.Kit.User,
		Content:        tabletype.JsonField(req.Content),
		DeliveryDetail: tabletype.JsonField(req.DeliveryDetail),
		Memo:           req.Memo,
		Creator:        cts.Kit.User,
		Reviser:        cts.Kit.User,
	}

	applicationID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		applicationID, err := svc.dao.Application().CreateWithTx(cts.Kit, txn, application)
		if err != nil {
			return nil, fmt.Errorf("create application failed, err: %v", err)
		}
		return applicationID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := applicationID.(string)
	if !ok {
		return nil, fmt.Errorf("create application but return id type not string, id type: %v",
			reflect.TypeOf(applicationID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

func (svc *applicationSvc) Update(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()

	req := new(proto.ApplicationUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	application := &tableapplication.ApplicationTable{
		Status: string(req.Status),
	}
	if req.DeliveryDetail != nil {
		application.DeliveryDetail = tabletype.JsonField(*req.DeliveryDetail)
	}

	err := svc.dao.Application().Update(cts.Kit, tools.EqualExpression("id", applicationID), application)
	if err != nil {
		logs.Errorf("update application failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("update application failed, err: %v", err)
	}

	return nil, nil
}

func (svc *applicationSvc) convertToApplicationResp(
	application *tableapplication.ApplicationTable,
) *proto.ApplicationResp {
	return &proto.ApplicationResp{
		ID:             application.ID,
		Source:         enumor.ApplicationSource(application.Source),
		SN:             application.SN,
		Type:           enumor.ApplicationType(application.Type),
		Status:         enumor.ApplicationStatus(application.Status),
		BkBizIDs:       application.BkBizIDs,
		Applicant:      application.Applicant,
		Content:        string(application.Content),
		DeliveryDetail: string(application.DeliveryDetail),
		Memo:           application.Memo,
		Revision: core.Revision{
			Creator:   application.Creator,
			Reviser:   application.Reviser,
			CreatedAt: application.CreatedAt.String(),
			UpdatedAt: application.UpdatedAt.String(),
		},
	}
}

func (svc *applicationSvc) List(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ApplicationListReq)
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
	daoApplicationResp, err := svc.dao.Application().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list application failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list application failed, err: %v", err)
	}
	if req.Page.Count {
		return &proto.ApplicationListResult{Count: daoApplicationResp.Count}, nil
	}

	details := make([]*proto.ApplicationResp, 0, len(daoApplicationResp.Details))
	for _, application := range daoApplicationResp.Details {
		details = append(details, svc.convertToApplicationResp(application))
	}

	return &proto.ApplicationListResult{Details: details}, nil
}

func (svc *applicationSvc) Get(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", applicationID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listApplicationDetails, err := svc.dao.Application().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list application failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list application failed, err: %v", err)
	}
	details := listApplicationDetails.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list application failed, application(id=%s) don't exist", applicationID)
	}

	return svc.convertToApplicationResp(details[0]), nil
}
