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
	if !task.Retry.IsEnable() {
		_, runErr = task.runOnce(act)
	} else {
		runErr = task.Retry.Run(func() error {
			needRetry, err := task.runOnce(act)
			if err == nil {
				return nil
			}

			if !needRetry {
				return err
			}

			// 允许重试，将Task状态由 running -> rollback，进行回滚
			if patchErr := task.UpdateState(enumor.TaskRollback, err.Error()); patchErr != nil {
				return fmt.Errorf("task set rollback state failed, after runAction failed, err: %v, patchErr: %v",
					err, patchErr)
			}

			return nil
		})
	}
	if runErr != nil {
		logs.Errorf("task run failed, err: %v, task: %+v, rid: %s", runErr, task, task.ExecuteKit.Kit().Rid)

		if patchErr := task.UpdateState(enumor.TaskFailed, runErr.Error()); patchErr != nil {
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

func (task *Task) runOnce(act action.Action) (needRetry bool, err error) {
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
		return false, fmt.Errorf("task decode params failed, err: %v", err)
	}

	return task.runAction(p, act)
}

func (task *Task) runActionWithRetry(params interface{}, act action.Action) (err error) {

	rty := retry.NewRetryPolicy(defRetryCount, defRetryRangeMS)
	for rty.RetryCount() < uint32(defRetryCount) {
		canRetry, err := task.runAction(params, act)
		if err != nil {
			logs.Errorf("task: %s run action failed, err: %v, canRetry: %v, rid: %s", task.ID, err, canRetry,
				task.ExecuteKit.Kit().Rid)

			// 如果不允许重试，直接返回
			if !canRetry {
				return err
			}

			// 允许重试，将Task状态由 running -> rollback，进行回滚
			if patchErr := task.UpdateState(enumor.TaskRollback, err.Error()); patchErr != nil {
				return fmt.Errorf("task set rollback state failed, after runAction failed, err: %v, patchErr: %v",
					err, patchErr)
			}
		}

		return nil
	}

	return err
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

	if err := task.UpdateState(enumor.TaskPending, ""); err != nil {
		return err
	}

	return nil
}

// runAction 执行Action，且只有执行Action运行逻辑失败才会允许重试，更改状态失败不进行重试。
func (task *Task) runAction(params interface{}, act action.Action) (retry bool, err error) {

	if task.State == enumor.TaskRollback {
		rollbackAct, ok := act.(action.RollbackAction)
		if !ok {
			return false, fmt.Errorf("action: %s not has RollbackAction", act.Name())
		}

		if err = rollbackAct.Rollback(task.ExecuteKit, params); err != nil {
			return true, fmt.Errorf("rollback failed, err: %v", err)
		}

		if err = task.UpdateState(enumor.TaskPending, ""); err != nil {
			return false, err
		}
	}

	if task.State == enumor.TaskPending {
		if err = task.UpdateState(enumor.TaskRunning, ""); err != nil {
			return false, err
		}

		if err = act.Run(task.ExecuteKit, params); err != nil {
			return true, fmt.Errorf("run failed, err: %v", err)
		}

		if err = task.UpdateState(enumor.TaskSuccess, ""); err != nil {
			return false, err
		}
	}

	return false, nil
}

// UpdateState update task state.
func (task *Task) UpdateState(state enumor.TaskState, reason string) error {
	md := &model.Task{
		ID:    task.ID,
		State: state,
		Reason: &tableasync.Reason{
			Message: reason,
		},
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
