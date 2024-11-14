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

package task

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tabletask "hcm/pkg/dal/table/task"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// CreateTaskManagement create task management.
func (svc *service) CreateTaskManagement(cts *rest.Contexts) (interface{}, error) {
	req := new(task.CreateManagementReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	managementIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tabletask.ManagementTable, 0, len(req.Items))
		for _, item := range req.Items {
			vendors := make(tabletype.StringArray, 0, len(item.Vendors))
			for _, vendor := range item.Vendors {
				vendors = append(vendors, string(vendor))
			}
			operations := make([]string, len(item.Operations))
			for i, operation := range item.Operations {
				operations[i] = string(operation)
			}
			model := tabletask.ManagementTable{
				BkBizID:    item.BkBizID,
				Source:     item.Source,
				Vendors:    vendors,
				State:      item.State,
				AccountIDs: item.AccountIDs,
				Resource:   item.Resource,
				Operations: operations,
				FlowIDs:    item.FlowIDs,
				Creator:    cts.Kit.User,
				Reviser:    cts.Kit.User,
			}
			if item.Extension != nil {
				extension, err := json.MarshalToString(item.Extension)
				if err != nil {
					return nil, errf.NewFromErr(errf.InvalidParameter, err)
				}
				model.Extension = tabletype.JsonField(extension)
			}

			models = append(models, model)
		}
		ids, err := svc.dao.TaskManagement().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create task Management failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create task management commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := managementIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create task management but return id type not string, id type: %v",
			reflect.TypeOf(managementIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteTaskManagement delete task management.
func (svc *service) DeleteTaskManagement(cts *rest.Contexts) (interface{}, error) {
	req := new(task.DeleteManagementReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.TaskManagement().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list task Management failed, err: %v", err)
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
		if err := svc.dao.TaskManagement().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateTaskManagement update task management.
func (svc *service) UpdateTaskManagement(cts *rest.Contexts) (interface{}, error) {
	req := new(task.UpdateManagementReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Items))
	for _, one := range req.Items {
		ids = append(ids, one.ID)
	}
	opt := &types.ListOption{Filter: tools.ContainersExpression("id", ids), Page: core.NewDefaultBasePage()}
	list, err := svc.dao.TaskManagement().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	existMap := make(map[string]tabletask.ManagementTable, len(list.Details))
	for _, one := range list.Details {
		existMap[one.ID] = one
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {
			existData, exist := existMap[one.ID]
			if !exist {
				continue
			}

			management := &tabletask.ManagementTable{
				BkBizID:    one.BkBizID,
				Source:     one.Source,
				State:      one.State,
				AccountIDs: one.AccountIDs,
				Resource:   one.Resource,
				FlowIDs:    one.FlowIDs,
				Reviser:    cts.Kit.User,
			}
			if len(one.Vendors) != 0 {
				vendors := make(tabletype.StringArray, 0, len(one.Vendors))
				for _, vendor := range one.Vendors {
					vendors = append(vendors, string(vendor))
				}
				management.Vendors = vendors
			}

			if len(one.Operations) != 0 {
				operations := make([]string, len(one.Operations))
				for i, operation := range one.Operations {
					operations[i] = string(operation)
				}
				management.Operations = operations
			}

			if one.Extension != nil {
				merge, err := json.UpdateMerge(one.Extension, string(existData.Extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				management.Extension = tabletype.JsonField(merge)
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.TaskManagement().UpdateWithTx(cts.Kit, txn, flt, management); err != nil {
				logs.Errorf("update task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update task management failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListTaskManagement list task management.
func (svc *service) ListTaskManagement(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	res, err := svc.dao.TaskManagement().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list task management failed, err: %v", err)
	}
	if req.Page.Count {
		return &task.ListManagementResult{Count: res.Count}, nil
	}

	managements := make([]coretask.Management, 0, len(res.Details))
	for _, one := range res.Details {
		extension := new(coretask.ManagementExt)
		if len(one.Extension) != 0 {
			if err = json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("UnmarshalFromString json extension failed, err: %v", err)
			}
		}
		operations := make([]enumor.TaskOperation, len(one.Operations))
		for i, operation := range one.Operations {
			operations[i] = enumor.TaskOperation(operation)
		}

		managements = append(managements, coretask.Management{
			ID:         one.ID,
			BkBizID:    one.BkBizID,
			Source:     one.Source,
			Vendors:    one.Vendors.ToVendors(),
			State:      one.State,
			AccountIDs: one.AccountIDs,
			Resource:   one.Resource,
			Operations: operations,
			FlowIDs:    one.FlowIDs,
			Extension:  extension,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &task.ListManagementResult{Details: managements}, nil
}

// CancelTaskManagement cancel task management.
func (svc *service) CancelTaskManagement(cts *rest.Contexts) (interface{}, error) {
	req := new(task.CancelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		managementFilter := tools.ContainersExpression("id", req.IDs)
		managementUpdate := &tabletask.ManagementTable{State: enumor.TaskManagementCancel}
		if err := svc.dao.TaskManagement().UpdateWithTx(cts.Kit, txn, managementFilter, managementUpdate); err != nil {
			logs.Errorf("update task management failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, fmt.Errorf("cancel task management failed, err: %v", err)
		}

		taskFilter := tools.ExpressionAnd(tools.RuleIn("task_management_id", req.IDs),
			tools.RuleEqual("state", enumor.TaskDetailInit))
		taskUpdate := &tabletask.DetailTable{State: enumor.TaskDetailCancel}
		if err := svc.dao.TaskDetail().UpdateWithTx(cts.Kit, txn, taskFilter, taskUpdate); err != nil {
			logs.Errorf("update task detail failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, fmt.Errorf("cancel task management failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("cancel task management failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
