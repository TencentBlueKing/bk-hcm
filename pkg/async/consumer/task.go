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
	Patch      func(kt *kit.Kit, task *model.Task) error
	Flow       *Flow
}

// ValidateBeforeExec task validate before execute.
func (task *Task) ValidateBeforeExec(act action.Action) error {
	switch task.State {
	case enumor.TaskPending:
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

// Run 任务执行。
func (task *Task) Run() error {

	act, exist := action.GetAction(task.ActionName)
	if !exist {
		return fmt.Errorf("action: %s not found", task.ActionName)
	}

	if err := task.ValidateBeforeExec(act); err != nil {
		return err
	}

	var runErr error
	var failedResult interface{}
	if !task.Retry.IsEnable() {
		_, failedResult, runErr = task.runOnce(act)
	} else {
		failedResult, runErr = task.Retry.Run(func() (interface{}, error) {
			needRetry, failed, err := task.runOnce(act)
			if err == nil {
				return nil, nil
			}

			if !needRetry {
				return failed, err
			}

			// 允许重试，将Task状态由 running -> rollback，进行回滚
			if patchErr := task.UpdateTask(enumor.TaskRollback, err.Error(), failed); patchErr != nil {
				return failed, fmt.Errorf("task set rollback state failed, after runAction failed, err: %v, "+
					"patchErr: %v", err, patchErr)
			}

			return nil, nil
		})
	}
	if runErr != nil {
		logs.Errorf("task run failed, err: %v, task: %+v, result: %+v, rid: %s", runErr, task, failedResult,
			task.ExecuteKit.Kit().Rid)

		if patchErr := task.UpdateTask(enumor.TaskFailed, runErr.Error(), failedResult); patchErr != nil {
			logs.Errorf("task set failed state failed, after run failed, err: %v, patchErr: %v, rid: %s",
				runErr, patchErr, task.ExecuteKit.Kit().Rid)
			return fmt.Errorf("task set failed state failed, after run failed, err: %v, patchErr: %v",
				runErr, patchErr)
		}

		return runErr
	}

	return nil
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

func (task *Task) runOnce(act action.Action) (needRetry bool, failedResult interface{}, err error) {
	if len(task.Params) == 0 {
		return task.runAction(nil, act)
	}

	paramAct, ok := act.(action.ParameterAction)
	if !ok {
		return task.runAction(nil, act)
	}

	p := paramAct.ParameterNew()
	if p == nil {
		return task.runAction(nil, act)
	}

	if err = action.Decode(task.Params, p); err != nil {
		logs.Errorf("task decode params failed, params: %s, type: %s, rid: %s", task.Params,
			reflect.TypeOf(p).String(), task.ExecuteKit.Kit().Rid)
		return false, nil, fmt.Errorf("task decode params failed, err: %v", err)
	}

	return task.runAction(p, act)
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

// runAction 执行Action，且只有执行Action运行逻辑失败才会允许重试，更改状态失败不进行重试。
func (task *Task) runAction(params interface{}, act action.Action) (retry bool, failedResult interface{}, err error) {

	if task.State == enumor.TaskRollback {
		rollbackAct, ok := act.(action.RollbackAction)
		if !ok {
			return false, nil, fmt.Errorf("action: %s not has RollbackAction", act.Name())
		}

		if err = rollbackAct.Rollback(task.ExecuteKit, params); err != nil {
			return true, nil, fmt.Errorf("rollback failed, err: %v", err)
		}

		if err = task.UpdateState(enumor.TaskPending); err != nil {
			return false, nil, err
		}
	}

	if task.State == enumor.TaskPending {
		if err = task.UpdateState(enumor.TaskRunning); err != nil {
			return false, nil, err
		}

		result, err := act.Run(task.ExecuteKit, params)
		if err != nil {
			return true, result, fmt.Errorf("run failed, err: %v", err)
		}

		// 如果执行成功，返回 result 属于成功结果，设置成功状态时，同时设置成功结果。如果执行失败，
		// 结果属于失败结果，交与上层更新失败或回滚等操作，更新失败结果。
		if err = task.UpdateStateResult(enumor.TaskSuccess, result); err != nil {
			return false, result, err
		}
	}

	return false, nil, nil
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
	md := &model.Task{
		ID:    task.ID,
		State: state,
		Reason: &tableasync.Reason{
			Message: reason,
		},
	}

	if result != nil {
		field, err := types.NewJsonField(result)
		if err != nil {
			logs.Errorf("update task marshal result failed, err: %v, result: %v, rid: %s", err, result,
				task.ExecuteKit.Kit().Rid)
			return err
		}
		md.Result = field
	}

	rty := retry.NewRetryPolicy(defRetryCount, defRetryRangeMS)
	err := rty.BaseExec(task.ExecuteKit.Kit(), func() error {
		return task.Patch(task.ExecuteKit.Kit(), md)
	})
	if err != nil {
		logs.Errorf("task update state failed, err: %v, retryCount: %d, id: %s, state: %s, reason: %s, rid: %s",
			err, defRetryCount, task.ID, state, reason, task.ExecuteKit.Kit().Rid)
		return err
	}

	task.State = state

	return nil
}
