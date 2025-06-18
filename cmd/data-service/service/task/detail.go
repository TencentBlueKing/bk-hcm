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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tabletask "hcm/pkg/dal/table/task"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// CreateTaskDetail create task detail.
func (svc *service) CreateTaskDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(task.CreateDetailReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	detailIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tabletask.DetailTable, 0, len(req.Items))
		for _, item := range req.Items {
			param, err := json.MarshalToString(item.Param)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			model := tabletask.DetailTable{
				BkBizID:          item.BkBizID,
				TaskManagementID: item.TaskManagementID,
				FlowID:           item.FlowID,
				TaskActionIDs:    item.TaskActionIDs,
				Operation:        item.Operation,
				Param:            tabletype.JsonField(param),
				State:            item.State,
				Reason:           item.Reason,
				Creator:          cts.Kit.User,
				Reviser:          cts.Kit.User,
			}
			if item.Extension != nil {
				extension, err := json.MarshalToString(item.Extension)
				if err != nil {
					return nil, errf.NewFromErr(errf.InvalidParameter, err)
				}
				model.Extension = tabletype.JsonField(extension)
			}
			if item.Result != nil {
				result, err := json.MarshalToString(item.Result)
				if err != nil {
					return nil, errf.NewFromErr(errf.InvalidParameter, err)
				}
				model.Result = tabletype.JsonField(result)
			}

			models = append(models, model)
		}
		ids, err := svc.dao.TaskDetail().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create task detail failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create task detail commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := detailIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create task detail but return id type not string, id type: %v",
			reflect.TypeOf(detailIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteTaskDetail delete task detail.
func (svc *service) DeleteTaskDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(task.DeleteDetailReq)
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
	listResp, err := svc.dao.TaskDetail().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list task detail failed, err: %v", err)
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
		if err := svc.dao.TaskDetail().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete task detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateTaskDetail update task detail.
func (svc *service) UpdateTaskDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(task.UpdateDetailReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateIDs := slice.Map(req.Items, func(i task.UpdateTaskDetailField) string { return i.ID })
	existMap, err := svc.getExistingTaskDetails(cts, updateIDs)
	if err != nil {
		logs.Errorf("fail to check detail exists, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {

			existData, exist := existMap[one.ID]
			if !exist {
				continue
			}

			detail := &tabletask.DetailTable{
				BkBizID:          one.BkBizID,
				TaskManagementID: one.TaskManagementID,
				FlowID:           one.FlowID,
				TaskActionIDs:    one.TaskActionIDs,
				Operation:        one.Operation,
				State:            one.State,
				Reason:           one.Reason,
				Reviser:          cts.Kit.User,
			}
			if one.Param != nil {
				merge, err := json.UpdateMerge(one.Param, string(existData.Param))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge param failed, err: %v", err)
				}
				detail.Param = tabletype.JsonField(merge)
			}
			if one.Result != nil {
				merge, err := json.UpdateMerge(one.Result, string(existData.Result))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge param failed, err: %v", err)
				}
				detail.Result = tabletype.JsonField(merge)
			}

			if one.Extension != nil {
				merge, err := json.UpdateMerge(one.Extension, string(existData.Extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				detail.Extension = tabletype.JsonField(merge)
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.TaskDetail().UpdateWithTx(cts.Kit, txn, flt, detail); err != nil {
				logs.Errorf("update task detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update task detail failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *service) getExistingTaskDetails(cts *rest.Contexts, ids []string) (map[string]tabletask.DetailTable, error) {

	existMap := make(map[string]tabletask.DetailTable, len(ids))

	for _, idBatch := range slice.Split(ids, int(filter.DefaultMaxInLimit)) {
		opt := &types.ListOption{
			Filter: tools.ContainersExpression("id", idBatch),
			Page:   core.NewDefaultBasePage(),
		}
		list, err := svc.dao.TaskDetail().List(cts.Kit, opt)
		if err != nil {
			logs.Errorf("list task detail failed, err: %v, ids: %v, rid: %s", err, idBatch, cts.Kit.Rid)
			return nil, err
		}
		for _, one := range list.Details {
			existMap[one.ID] = one
		}
	}
	return existMap, nil
}

// ListTaskDetail list task detail.
func (svc *service) ListTaskDetail(cts *rest.Contexts) (interface{}, error) {
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
	res, err := svc.dao.TaskDetail().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list task detail failed, err: %v", err)
	}
	if req.Page.Count {
		return &task.ListDetailResult{Count: res.Count}, nil
	}

	details := make([]coretask.Detail, 0, len(res.Details))
	for _, one := range res.Details {
		extension := new(coretask.DetailExt)
		if len(one.Extension) != 0 {
			if err = json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("UnmarshalFromString json extension failed, err: %v", err)
			}
		}
		details = append(details, coretask.Detail{
			ID:               one.ID,
			BkBizID:          one.BkBizID,
			TaskManagementID: one.TaskManagementID,
			FlowID:           one.FlowID,
			TaskActionIDs:    one.TaskActionIDs,
			Operation:        one.Operation,
			Param:            one.Param,
			Result:           one.Result,
			State:            one.State,
			Reason:           one.Reason,
			Extension:        extension,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &task.ListDetailResult{Details: details}, nil
}
