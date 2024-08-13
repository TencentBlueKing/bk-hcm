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

package loadbalancer

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// ListBatchOperationAsyncFlowRel ...
func (svc *lbSvc) ListBatchOperationAsyncFlowRel(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
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
	result, err := svc.dao.BatchOperationAsyncFlowRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list batch task rel failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list batch task rel failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.BatchOperationAsyncFlowRelListResult{Count: result.Count}, nil
	}

	details := make([]*corelb.BatchOperationAsyncFlowRel, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseBatchOperationAsyncFlowRel(one)
		details = append(details, tmpOne)
	}

	return &protocloud.BatchOperationAsyncFlowRelListResult{Details: details}, nil
}

func convTableToBaseBatchOperationAsyncFlowRel(table *tablelb.BatchOperationAsyncFlowRelTable) *corelb.BatchOperationAsyncFlowRel {
	return &corelb.BatchOperationAsyncFlowRel{
		ID:               table.ID,
		BatchOperationID: table.BatchOperationID,
		AuditID:          *table.AuditID,
		FlowID:           table.FlowID,
	}
}

// BatchCreateBatchOperationAsyncFlowRel ...
func (svc *lbSvc) BatchCreateBatchOperationAsyncFlowRel(cts *rest.Contexts) (any, error) {
	req := new(protocloud.BatchOperationAsyncFlowRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.BatchOperationAsyncFlowRelTable, 0, len(req.Rels))
		for _, item := range req.Rels {
			models = append(models, &tablelb.BatchOperationAsyncFlowRelTable{
				BatchOperationID: item.BatchOperationID,
				AuditID:          item.AuditID,
				FlowID:           item.FlowID,
				Creator:          cts.Kit.User,
			})
		}
		ids, err := svc.dao.BatchOperationAsyncFlowRel().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("fail to batch create batch task rel, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create batch task rel failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create batch task rel failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create batch task but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
