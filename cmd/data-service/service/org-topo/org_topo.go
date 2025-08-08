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

// Package orgtopo org topo service
package orgtopo

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dataorgtopo "hcm/pkg/api/data-service/org_topo"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchCreateOrgTopo batch create org topo.
func (svc *service) BatchCreateOrgTopo(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.BatchCreateOrgTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch create org topo decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch create org topo validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// batch create
	recordIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ids, err := svc.dao.OrgTopo().BatchCreate(cts.Kit, txn, req.OrgTopos)
		if err != nil {
			logs.Errorf("batch create org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	ids, ok := recordIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create org topo but return id type is not string, id type: %v",
			reflect.TypeOf(recordIDs).String())
	}
	return &core.BatchCreateResult{IDs: ids}, nil
}

// ListOrgTopo list org topo.
func (svc *service) ListOrgTopo(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list org topo decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list org topo validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.OrgTopo().List(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	resp := &dataorgtopo.ListResp{
		Count:   result.Count,
		Details: result.Details,
	}

	return resp, nil
}

// ListOrgTopoByDeptIDs list org topo by dept ids.
func (svc *service) ListOrgTopoByDeptIDs(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.ListByDeptIDsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list org topo by deptids decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list org topo by deptids validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.OrgTopo().ListByDeptIDs(cts.Kit, req.DeptIDs)
	if err != nil {
		logs.Errorf("list org topo by deptids failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	resp := &dataorgtopo.ListResp{
		Details: result.Details,
	}

	return resp, nil
}

// BatchUpdateOrgTopo batch update org topo.
func (svc *service) BatchUpdateOrgTopo(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.BatchUpdateOrgTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch update org topo decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch update org topo validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if _, err := svc.dao.OrgTopo().BatchUpdate(cts.Kit, txn, req.OrgTopos); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// BatchDeleteOrgTopo batch delete org topo.
func (svc *service) BatchDeleteOrgTopo(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.BatchDeleteOrgTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete org topo decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete org topo validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.OrgTopo().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch delete list org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("delete list org topo failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if _, err = svc.dao.OrgTopo().BatchDelete(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch delete org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchUpsertOrgTopo batch upsert org topo.
func (svc *service) BatchUpsertOrgTopo(cts *rest.Contexts) (interface{}, error) {
	req := new(dataorgtopo.BatchUpsertOrgTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch upsert org topo decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch upsert org topo validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	recordIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ids := make([]string, 0)
		for _, batchIDs := range slice.Split(req.AddOrgTopos, constant.BatchOperationMaxLimit) {
			splitIDs, err := svc.dao.OrgTopo().BatchCreate(cts.Kit, txn, batchIDs)
			if err != nil {
				logs.Errorf("batch upsert of create org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
			ids = append(ids, splitIDs...)
		}
		for _, batchIDs := range slice.Split(req.UpdateOrgTopos, constant.BatchOperationMaxLimit) {
			_, err := svc.dao.OrgTopo().BatchUpdate(cts.Kit, txn, batchIDs)
			if err != nil {
				logs.Errorf("batch upsert of update org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch upsert org topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	ids, ok := recordIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch update org topo but return id type is not string, id type: %v",
			reflect.TypeOf(recordIDs).String())
	}
	return &core.BatchCreateResult{IDs: ids}, nil
}
