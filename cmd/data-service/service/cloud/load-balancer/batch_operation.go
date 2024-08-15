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
	"hcm/pkg/api/core/audit"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	auditDao "hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// ListBatchOperation ...
func (svc *lbSvc) ListBatchOperation(cts *rest.Contexts) (any, error) {
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
	result, err := svc.dao.BatchOperation().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list batch task failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list batch task failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.BatchOperationListResult{Count: result.Count}, nil
	}

	details := make([]*corelb.BatchOperation, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseBatchOperation(one)
		details = append(details, tmpOne)
	}

	return &protocloud.BatchOperationListResult{Details: details}, nil
}

func convTableToBaseBatchOperation(table *tablelb.BatchOperationTable) *corelb.BatchOperation {
	return &corelb.BatchOperation{
		ID:      table.ID,
		BkBizID: table.BkBizID,
		AuditID: table.AuditID,
		Detail:  table.Detail,
	}
}

// BatchCreateBatchOperation ...
func (svc *lbSvc) BatchCreateBatchOperation(cts *rest.Contexts) (any, error) {
	req := new(protocloud.BatchOperationBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablelb.BatchOperationTable, 0, len(req.Tasks))
		for _, item := range req.Tasks {
			models = append(models, &tablelb.BatchOperationTable{
				BkBizID: item.BkBizID,
				AuditID: item.AuditID,
				Detail:  item.Detail,
				Creator: cts.Kit.User,
			})
		}
		ids, err := svc.dao.BatchOperation().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create batch task, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create batch task failed, err: %v", err)
		}

		if err = batchCreateAudit(cts.Kit, txn, svc.dao.Audit(), models, req.Tasks, req.AccountID); err != nil {
			logs.Errorf("batch create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		// 更新batch_task的audit_id
		expr, err := tools.And(
			tools.ContainersExpression("res_id", ids),
			tools.RuleEqual("res_type", enumor.LoadBalancerAuditResType),
			tools.RuleEqual("action", enumor.BatchOperation),
		)
		if err != nil {
			return nil, err
		}
		auditList, err := svc.dao.Audit().ListWithTx(cts.Kit, txn, &types.ListOption{
			Filter: expr,
			Page:   core.NewDefaultBasePage(),
		})
		if err != nil {
			logs.Errorf("list audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		taskAuditMap := make(map[string]tableaudit.AuditTable)
		for _, tmp := range auditList.Details {
			taskAuditMap[tmp.ResID] = tmp
		}

		for _, model := range models {
			models[0].AuditID = int64(taskAuditMap[model.ID].ID)
			err := svc.dao.BatchOperation().UpdateByIDWithTx(cts.Kit, txn, model.ID, models[0])
			if err != nil {
				logs.Errorf("update batch task audit_id failed, model(%s), err: %v, rid: %s", model.ID, err, cts.Kit.Rid)
				return nil, err
			}
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create batch task failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create batch task but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}
	return &core.BatchCreateResult{IDs: ids}, nil
}

func batchCreateAudit(kt *kit.Kit, txn *sqlx.Tx, auditInterface auditDao.Interface,
	models []*tablelb.BatchOperationTable, tasks []*protocloud.BatchOperationCreateReq, accountID string) error {

	audits := make([]*tableaudit.AuditTable, 0)
	for i := 0; i < len(models); i++ {
		// record audit
		auditData := &audit.BatchOperationAuditDetail{
			BatchOperationID:   models[i].ID,
			BatchOperationType: tasks[i].Type,
		}
		audits = append(audits, &tableaudit.AuditTable{
			ResID:     models[i].ID,
			ResName:   fmt.Sprintf("batch-operation-%s", models[i].ID),
			ResType:   enumor.LoadBalancerAuditResType,
			BkBizID:   models[i].BkBizID,
			AccountID: accountID,
			Action:    enumor.BatchOperation,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: auditData,
			},
		})
	}
	if err := auditInterface.BatchCreateWithTx(kt, txn, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
