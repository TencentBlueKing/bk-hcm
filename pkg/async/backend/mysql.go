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

package backend

import (
	"errors"
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	"hcm/pkg/async/action"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/async"
	tableasync "hcm/pkg/dal/table/async"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// NewMysql create mysql instance
func NewMysql(dao dao.Set) Backend {
	return &mysql{
		dao: dao,
	}
}

// mysql mysql mysql
type mysql struct {
	dao dao.Set
}

// BatchUpdateFlowStateByCAS CAS批量更新流状态
func (db *mysql) BatchUpdateFlowStateByCAS(kt *kit.Kit, infos []UpdateFlowInfo) error {

	for _, one := range infos {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	_, err := db.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range infos {
			info := &typesasync.UpdateFlowInfo{
				ID:     one.ID,
				Source: one.Source,
				Target: one.Target,
				Reason: one.Reason,
				Worker: one.Worker,
			}
			if err := db.dao.AsyncFlow().UpdateStateByCAS(kt, txn, info); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RetryTask 重试任务
func (db *mysql) RetryTask(kt *kit.Kit, flowID, taskID string) error {

	if err := db.checkFlowTaskForRetry(kt, flowID, taskID); err != nil {
		return err
	}

	flowUpdate := &typesasync.UpdateFlowInfo{
		ID:     flowID,
		Source: enumor.FlowFailed,
		Target: enumor.FlowPending,
		Reason: &tableasync.Reason{Message: "retry task " + taskID},
	}
	taskUpdate := &typesasync.UpdateTaskInfo{
		ID:     taskID,
		Source: enumor.TaskFailed,
		Target: enumor.TaskPending,
		Reason: &tableasync.Reason{Message: "retry task " + taskID},
	}

	_, err := db.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		if err := db.dao.AsyncFlowTask().UpdateStateByCAS(kt, txn, taskUpdate); err != nil {
			logs.Errorf("fail to update task status for retry, err: %v, task id: %s, rid: %s",
				err, taskID, kt.Rid)
			return nil, err
		}
		if err := db.dao.AsyncFlow().UpdateStateByCAS(kt, txn, flowUpdate); err != nil {
			logs.Errorf("fail to update flow status for retry, err: %v, flow id: %s, rid: %s",
				err, flowID, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *mysql) checkFlowTaskForRetry(kt *kit.Kit, flowID string, taskID string) error {
	if len(flowID) == 0 || len(taskID) == 0 {
		return errors.New("empty flow id or task id")
	}

	listOpt := &types.ListOption{
		Filter: tools.EqualExpression("id", flowID),
		Page:   core.NewDefaultBasePage(),
	}
	flowResp, err := db.dao.AsyncFlow().List(kt, listOpt)
	if err != nil {
		return err
	}
	if len(flowResp.Details) == 0 {
		return fmt.Errorf("flow %s not found", flowID)
	}
	if flowResp.Details[0].State != enumor.FlowFailed {
		return fmt.Errorf("flow(%s) state(%s) wrong, only `failed` allowed for retry",
			flowID, flowResp.Details[0].State)
	}

	listOpt = &types.ListOption{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", taskID),
			tools.RuleEqual("flow_id", flowID)),
		Page: core.NewDefaultBasePage(),
	}
	taskResp, err := db.dao.AsyncFlowTask().List(kt, listOpt)
	if err != nil {
		return err
	}
	if len(taskResp.Details) == 0 {
		return fmt.Errorf("task(%s) of flow(%s) not found", taskID, flowID)
	}
	if taskResp.Details[0].State != enumor.TaskFailed {
		return fmt.Errorf("task(%s) state(%s) wrong, only `failed` allowed for retry",
			taskID, taskResp.Details[0].State)
	}
	return nil
}

// UpdateTaskStateByCAS CAS更新任务状态
func (db *mysql) UpdateTaskStateByCAS(kt *kit.Kit, info *UpdateTaskInfo) error {
	update := &typesasync.UpdateTaskInfo{
		ID:     info.ID,
		Source: info.Source,
		Target: info.Target,
		Reason: info.Reason,
	}
	_, err := db.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := db.dao.AsyncFlowTask().UpdateStateByCAS(kt, txn, update)
		if err != nil {
			logs.Errorf("fail to update task state cas, err: %v, info: %+v, rid: %s", err, info, kt.Rid)
		}
		return nil, err
	})
	return err
}

var _ Backend = new(mysql)

// CreateFlow 创建任务流
func (db *mysql) CreateFlow(kt *kit.Kit, flow *model.Flow) (string, error) {

	flowState := enumor.FlowPending
	if flow.State == enumor.FlowInit {
		flowState = flow.State
	}

	result, err := db.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// 创建任务流
		md := &tableasync.AsyncFlowTable{
			Name:      flow.Name,
			State:     flowState,
			Reason:    new(tableasync.Reason),
			ShareData: flow.ShareData,
			Memo:      flow.Memo,
			Worker:    converter.ValToPtr(""),
			Creator:   kt.User,
			Reviser:   kt.User,
		}
		flowID, err := db.dao.AsyncFlow().Create(kt, txn, md)
		if err != nil {
			return nil, err
		}

		// 创建任务
		tasks := flow.Tasks
		mds := make([]tableasync.AsyncFlowTaskTable, 0, len(tasks))
		for _, one := range tasks {
			taskState := enumor.TaskPending
			if one.State == enumor.TaskInit {
				taskState = one.State
			}

			mds = append(mds, tableasync.AsyncFlowTaskTable{
				FlowID:     flowID,
				FlowName:   one.FlowName,
				ActionID:   string(one.ActionID),
				ActionName: one.ActionName,
				Params:     one.Params,
				Retry:      one.Retry,
				DependOn:   dependOnToStringArray(one.DependOn),
				State:      taskState,
				Reason:     new(tableasync.Reason),
				Creator:    kt.User,
				Reviser:    kt.User,
			})
		}
		if _, err = db.dao.AsyncFlowTask().BatchCreateWithTx(kt, txn, mds); err != nil {
			return nil, err
		}

		return flowID, nil
	})
	if err != nil {
		return "", err
	}

	flowID, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("return result not string type, type: %s", reflect.TypeOf(result).String())
	}

	return flowID, nil
}

// BatchUpdateFlow 批量更新任务流
func (db *mysql) BatchUpdateFlow(kt *kit.Kit, flows []model.Flow) error {

	_, err := db.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range flows {
			md := &tableasync.AsyncFlowTable{
				State:     one.State,
				Reason:    one.Reason,
				ShareData: one.ShareData,
				Memo:      one.Memo,
				Worker:    one.Worker,
				Reviser:   one.Reviser,
			}

			if err := db.dao.AsyncFlow().UpdateByIDWithTx(kt, txn, one.ID, md); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ListFlow 查询任务流
func (db *mysql) ListFlow(kt *kit.Kit, input *ListInput) ([]model.Flow, error) {

	opt := &types.ListOption{
		Fields: input.Fields,
		Filter: input.Filter,
		Page:   input.Page,
	}
	list, err := db.dao.AsyncFlow().List(kt, opt)
	if err != nil {
		return nil, err
	}

	flows := make([]model.Flow, 0, len(list.Details))
	for _, one := range list.Details {
		flows = append(flows, model.Flow{
			ID:        one.ID,
			Name:      one.Name,
			State:     one.State,
			Reason:    one.Reason,
			ShareData: one.ShareData,
			Memo:      one.Memo,
			Worker:    one.Worker,
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		})
	}

	return flows, nil
}

// BatchCreateTask 批量创建任务
func (db *mysql) BatchCreateTask(kt *kit.Kit, tasks []model.Task) ([]string, error) {

	mds := make([]tableasync.AsyncFlowTaskTable, 0, len(tasks))
	for _, one := range tasks {
		mds = append(mds, tableasync.AsyncFlowTaskTable{
			FlowID:     one.FlowID,
			FlowName:   one.FlowName,
			ActionID:   string(one.ActionID),
			ActionName: one.ActionName,
			Params:     one.Params,
			Retry:      one.Retry,
			DependOn:   dependOnToStringArray(one.DependOn),
			State:      enumor.TaskPending,
			Reason:     one.Reason,
			Creator:    one.Creator,
			Reviser:    one.Reviser,
		})
	}

	return db.dao.AsyncFlowTask().BatchCreate(kt, mds)
}

// UpdateTask 更新任务
func (db *mysql) UpdateTask(kt *kit.Kit, task *model.Task) error {

	md := &tableasync.AsyncFlowTaskTable{
		Retry:   task.Retry,
		State:   task.State,
		Result:  task.Result,
		Reason:  task.Reason,
		Reviser: kt.User,
	}

	return db.dao.AsyncFlowTask().UpdateByID(kt, task.ID, md)
}

// ListTask 查询任务
func (db *mysql) ListTask(kt *kit.Kit, input *ListInput) ([]model.Task, error) {

	opt := &types.ListOption{
		Fields: input.Fields,
		Filter: input.Filter,
		Page:   input.Page,
	}
	list, err := db.dao.AsyncFlowTask().List(kt, opt)
	if err != nil {
		return nil, err
	}

	tasks := make([]model.Task, 0, len(list.Details))
	for _, one := range list.Details {
		tasks = append(tasks, model.Task{
			ID:         one.ID,
			FlowID:     one.FlowID,
			FlowName:   one.FlowName,
			ActionID:   action.ActIDType(one.ActionID),
			ActionName: one.ActionName,
			Params:     one.Params,
			Retry:      one.Retry,
			DependOn:   dependOnToActIDArray(one.DependOn),
			State:      one.State,
			Reason:     one.Reason,
			Result:     one.Result,
			Creator:    one.Creator,
			Reviser:    one.Reviser,
			CreatedAt:  one.CreatedAt.String(),
			UpdatedAt:  one.UpdatedAt.String(),
		})
	}

	return tasks, nil
}

func dependOnToStringArray(d []action.ActIDType) tabletypes.StringArray {
	result := make(tabletypes.StringArray, 0, len(d))
	for _, one := range d {
		result = append(result, string(one))
	}

	return result
}

func dependOnToActIDArray(d tabletypes.StringArray) []action.ActIDType {
	result := make([]action.ActIDType, 0, len(d))
	for _, one := range d {
		result = append(result, action.ActIDType(one))
	}

	return result
}
