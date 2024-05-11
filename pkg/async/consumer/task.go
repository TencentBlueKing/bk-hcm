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

package consumer

import (
	"errors"
	"fmt"
	"reflect"

	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
)

// Task 异步任务执行体，包含了任务运行流程、回滚流程。
type Task struct {
	model.Task `json:",inline"`

	Kit *kit.Kit `json:"-"`

	ExecuteKit run.ExecuteKit `json:"-"`
	Patch      func(taskKit *kit.Kit, task *model.Task) error
	Flow       *Flow
}

// ValidateBeforeExec task validate before execute.
func (task *Task) ValidateBeforeExec(act action.Action) error {
	switch task.State {
	case enumor.TaskPending, enumor.TaskRollback:
	default:
		return fmt.Errorf("task can not run，state: %s", task.State)
	}

	if len(task.Params) != 0 && task.Params != "{}" {
		if _, ok := act.(action.ParameterAction); !ok {
			return errors.New("task has params, but not have ParameterAction")
		}
	}

	if task.Retry.IsEnable() {
		if _, ok := act.(action.RollbackAction); !ok {
			return errors.New("task can retry, but not have RollbackAction")
		}
	}

	return nil
}

// ValidateBeforeRollback task validate before rollback.
func (task *Task) ValidateBeforeRollback(act action.Action) error {
	switch task.State {
	case enumor.TaskRunning, enumor.TaskRollback:
	default:
		return fmt.Errorf("task can not run，state: %s", task.State)
	}

	if len(task.Params) != 0 && task.Params != "{}" {
		if _, ok := act.(action.ParameterAction); !ok {
			return errors.New("task has params, but not have ParameterAction")
		}
	}

	if task.Retry.IsEnable() {
		if _, ok := act.(action.RollbackAction); !ok {
			return errors.New("task can retry, but not have RollbackAction")
		}
	}

	return nil
}

// InitDep init task for exec.
func (task *Task) InitDep(kt run.ExecuteKit, patch func(kt *kit.Kit, task *model.Task) error, flow *Flow) {
	task.ExecuteKit = kt
	task.Patch = patch
	task.Flow = flow
}

// Rollback 任务强制回滚，从Running或者Rollback状态
func (task *Task) Rollback() error {

	act, exist := action.GetAction(task.ActionName)
	if !exist {
		return fmt.Errorf("action: %s not found", task.ActionName)
	}

	if err := task.ValidateBeforeRollback(act); err != nil {
		return err
	}

	if len(task.Params) == 0 {
		return task.rollback(nil, act)
	}

	paramAct, ok := act.(action.ParameterAction)
	if !ok {
		return task.rollback(nil, act)
	}

	p := paramAct.ParameterNew()
	if p == nil {
		return task.rollback(nil, act)
	}

	if err := action.Decode(task.Params, p); err != nil {
		logs.Errorf("task decode params failed, params: %s, type: %s, rid: %s", task.Params,
			reflect.TypeOf(p).String(), task.ExecuteKit.Kit().Rid)
		return fmt.Errorf("task decode params failed, err: %v", err)
	}

	return task.rollback(p, act)
}

func (task *Task) prepareParams(act action.Action) (params any, err error) {
	if len(task.Params) == 0 {
		return nil, nil
	}

	paramAct, ok := act.(action.ParameterAction)
	if !ok {
		return nil, nil
	}

	p := paramAct.ParameterNew()
	if p == nil {
		return nil, nil
	}
	if err = action.Decode(task.Params, p); err != nil {
		logs.Errorf("task decode params failed, params: %s, type: %s, rid: %s", task.Params,
			reflect.TypeOf(p).String(), task.ExecuteKit.Kit().Rid)
		return nil, fmt.Errorf("task decode params failed, err: %v", err)
	}
	return p, nil
}

// rollback 任务强制回滚，从Running或者Rollback状态
func (task *Task) rollback(params interface{}, act action.Action) error {
	rollbackAct, ok := act.(action.RollbackAction)
	if !ok {
		return fmt.Errorf("action: %s not has RollbackAction", act.Name())
	}

	if err := rollbackAct.Rollback(task.ExecuteKit, params); err != nil {
		return fmt.Errorf("rollback failed, err: %v", err)
	}

	if err := task.UpdateState(enumor.TaskPending); err != nil {
		return err
	}

	return nil
}

// UpdateState update task state.
func (task *Task) UpdateState(state enumor.TaskState) error {
	return task.UpdateTask(state, "", nil)
}

// UpdateStateResult update task state and result.
func (task *Task) UpdateStateResult(state enumor.TaskState, result interface{}) error {
	return task.UpdateTask(state, "", result)
}

// UpdateTask update task.
func (task *Task) UpdateTask(state enumor.TaskState, reason string, result interface{}) error {
	md, err := task.buildTaskUpdateModel(task.ExecuteKit.Kit(), state, reason, result)
	if err != nil {
		return err
	}
	rty := retry.NewRetryPolicy(DefRetryCount, DefRetryRangeMS)
	err = rty.BaseExec(task.ExecuteKit.Kit(), func() error {
		return task.Patch(task.ExecuteKit.Kit(), md)
	})
	if err != nil {
		logs.Errorf("task update state failed, err: %v, retryCount: %d, id: %s, state: %s, reason: %s, rid: %s",
			err, DefRetryCount, task.ID, state, reason, task.ExecuteKit.Kit().Rid)
		return err
	}

	task.State = state

	return nil
}

func (task *Task) buildTaskUpdateModel(kt *kit.Kit, state enumor.TaskState, reason string,
	result interface{}) (*model.Task, error) {

	md := &model.Task{
		ID:    task.ID,
		State: state,
		Reason: &tableasync.Reason{
			Message:       task.Reason.Message,
			RollbackCount: task.Reason.RollbackCount,
			PreState:      string(task.State),
		},
	}
	if reason != "" {
		md.Reason.Message = reason
	}

	// 更新为rollback，记录rollback次数
	if state == enumor.TaskRollback {
		md.Reason.RollbackCount = task.Reason.RollbackCount + 1
	}
	if result != nil {
		field, err := types.NewJsonField(result)
		if err != nil {
			logs.Errorf("update task marshal result failed, err: %v, result: %v, rid: %s",
				err, result, kt.Rid)
			return nil, err
		}
		md.Result = field
	}
	return md, nil
}
