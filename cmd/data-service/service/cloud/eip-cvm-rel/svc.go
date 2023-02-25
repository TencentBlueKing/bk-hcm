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

package eipcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	reldao "hcm/pkg/dal/dao/cloud/eip-cvm-rel"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

type relSvc struct {
	dao.Set
	objectDao *reldao.EipCvmRelDao
}

// Init ...
func (svc *relSvc) Init() {
	d := &reldao.EipCvmRelDao{}
	registeredDao := svc.GetObjectDao(d.Name())
	if registeredDao == nil {
		d.ObjectDaoManager = new(dao.ObjectDaoManager)
		svc.RegisterObjectDao(d)
	}

	svc.objectDao = svc.GetObjectDao(d.Name()).(*reldao.EipCvmRelDao)
}

// BatchCreate ...
func (svc *relSvc) BatchCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipCvmRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		rels := make([]*tablecloud.EipCvmRelModel, len(req.Rels))
		for idx, relReq := range req.Rels {
			rels[idx] = &tablecloud.EipCvmRelModel{
				CvmID:   relReq.CvmID,
				EipID:   relReq.EipID,
				Creator: cts.Kit.User,
			}
		}

		return nil, svc.objectDao.BatchCreateWithTx(cts.Kit, txn, rels)
	})

	return nil, err
}

// List ...
func (svc *relSvc) List(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipCvmRelListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list eip cvm rels failed, err: %v", err)
	}

	if req.Page.Count {
		return &dataproto.EipCvmRelListResult{Count: data.Count}, nil
	}

	details := make([]*dataproto.EipCvmRelResult, len(data.Details))
	for idx, r := range data.Details {
		details[idx] = &dataproto.EipCvmRelResult{
			ID:        r.ID,
			EipID:     r.EipID,
			CvmID:     r.CvmID,
			Creator:   r.Creator,
			CreatedAt: r.CreatedAt,
		}
	}

	return &dataproto.EipCvmRelListResult{Details: details}, nil
}

// BatchDelete ...
func (svc *relSvc) BatchDelete(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipCvmRelDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page:   core.DefaultBasePage,
	}

	relResult, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list eip cvm rels failed, err: %v", err)
	}

	if len(relResult.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]uint64, len(relResult.Details))
	for idx, rel := range relResult.Details {
		delIDs[idx] = rel.ID
	}

	_, err = svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return nil, svc.objectDao.DeleteWithTx(cts.Kit, txn, tools.ContainersExpression("id", delIDs))
	})
	return nil, err
}

// ListWithEip ...
func (svc *relSvc) ListWithEip(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipCvmRelWithEipListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	data, err := svc.objectDao.ListJoinEip(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}

	eips := make([]*dataproto.EipWithCvmID, len(data.Details))
	for idx, d := range data.Details {
		eips[idx] = &dataproto.EipWithCvmID{
			EipResult: eip.EipResult{
				ID:        d.ID,
				Vendor:    d.Vendor,
				CloudID:   d.CloudID,
				AccountID: d.AccountID,
				Name:      d.Name,
				BkBizID:   d.BkBizID,
				Region:    d.Region,
				Status:    d.Status,
				PublicIp:  d.PublicIp,
				PrivateIp: d.PrivateIp,
				Creator:   d.Creator,
				Reviser:   d.Reviser,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			},
			CvmID:        d.CvmID,
			RelCreator:   d.RelCreator,
			RelCreatedAt: d.RelCreatedAt,
		}
	}

	return eips, nil
}
