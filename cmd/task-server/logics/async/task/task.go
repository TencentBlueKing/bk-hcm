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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/task-server/logics/async/backends/iface"
	"hcm/pkg/api/core/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
)

// Task define task
type Task struct {
	// 任务唯一标识（递增字符串）
	ID string `json:"id" validate:"required"`
	// 对应任务流唯一标识（递增字符串）
	FlowID string `json:"flow_id" validate:"required"`
	// 任务流名称
	FlowName string `json:"flow_name" validate:"required"`
	// 任务执行体名称
	ActionName string `json:"action_name" validate:"required"`
	// 任务参数定义可扩展
	Params json.RawMessage `json:"params" validate:"omitempty"`
	// 任务重试次数
	RetryCount int `json:"retry_count" validate:"omitempty"`
	// 任务超时时间
	TimeoutSecs int `json:"timeout_secs" validate:"required"`
	// 依赖任务的ID列表
	DependOn []string `json:"depend_on" validate:"required"`
	// 任务状态
	State enumor.TaskState `json:"state" validate:"required"`
	// 任务描述信息
	Memo string `json:"memo" validate:"omitempty"`
	// task kit
	kt *kit.Kit
	// backend
	backend iface.Backend
}

// Validate Task.
func (t *Task) Validate() error {

	if t.TimeoutSecs <= 0 {
		return errors.New("task need max run time ")
	}

	return validator.Validate.Struct(t)
}

func (t *Task) DoTask(kt *kit.Kit, backend iface.Backend) error {

	if err := t.Validate(); err != nil {
		return err
	}

	// 任务状态为成功或者失败直接返回
	if t.State == enumor.TaskSuccess || t.State == enumor.TaskFailed {
		return nil
	}

	action, isExist := ActionManagerInstance.GetAction(t.ActionName)
	if !isExist {
		return fmt.Errorf("action: %s can not find", t.ActionName)
	}

	t.kt = kt
	t.backend = backend

	if t.State == enumor.TaskBeforeFailed {
		return t.doRunBeforeFailed(action)
	}

	doneTaskChan := make(chan struct{}, 1)
	taskRetChan := make(chan error, 1)

	go func() {
		defer func() {
			doneTaskChan <- struct{}{}
		}()

		switch t.State {
		case enumor.TaskPending:
			if err := t.doRunBefore(action); err != nil {
				taskRetChan <- err
				return
			}
		case enumor.TaskRunning:
			if err := t.doRun(action); err != nil {
				taskRetChan <- err
				return
			}
		case enumor.TaskBeforeSuccess:
			if err := t.doRunBeforeSuccess(action); err != nil {
				taskRetChan <- err
				return
			}
		}

		taskRetChan <- nil
	}()

	select {
	case <-doneTaskChan:
		err := <-taskRetChan
		if err != nil {
			if err := t.doRetry(action); err != nil {
				return err
			}
		}
	case <-time.After(time.Duration(t.TimeoutSecs) * time.Second):
		if err := t.doRetry(action); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) doRunBefore(action Action) error {
	logs.V(3).Infof("[async] do run before start with rid %s", t.kt.Rid)

	err := action.RunBefore(t.kt, t.Params)
	if err != nil {
		logs.Errorf("[async] run before func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	//执行成功设置任务状态为TaskRunning
	if err := t.ChangeTaskState(enumor.TaskRunning, constant.AsyncDefaultJson); err != nil {
		logs.Errorf("[async] change task state failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 执行Run
	if err := t.doRun(action); err != nil {
		logs.Errorf("[async] do run func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] do run before end with rid %s", t.kt.Rid)
	return nil
}

func (t *Task) doRun(action Action) error {
	logs.V(3).Infof("[async] do run start with rid %s", t.kt.Rid)

	err := action.Run(t.kt, t.Params)
	if err != nil {
		logs.Errorf("[async] run func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 执行成功设置任务状态为TaskBeforeSuccess
	if err := t.ChangeTaskState(enumor.TaskBeforeSuccess, constant.AsyncDefaultJson); err != nil {
		logs.Errorf("[async] change task state failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 执行RunBeforeSuccess
	if err := t.doRunBeforeSuccess(action); err != nil {
		logs.Errorf("[async] do run before success func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] do run end with rid %s", t.kt.Rid)
	return nil
}

func (t *Task) doRunBeforeSuccess(action Action) error {
	logs.V(3).Infof("[async] do run before success start with rid %s", t.kt.Rid)

	err := action.RunBeforeSuccess(t.kt, t.Params)
	if err != nil {
		logs.Errorf("[async] run before success func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 执行成功设置任务状态为TaskSuccess
	if err := t.ChangeTaskState(enumor.TaskSuccess, constant.AsyncDefaultJson); err != nil {
		logs.Errorf("[async] change task state failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] do run before success end with rid %s", t.kt.Rid)
	return nil
}

func (t *Task) doRunBeforeFailed(action Action) error {
	logs.V(3).Infof("[async] do run before failed start with rid %s", t.kt.Rid)

	err := action.RunBeforeFailed(t.kt, t.Params)
	if err != nil {
		logs.Errorf("[async] run before failed func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 执行成功设置任务状态为TaskFailed
	if err := t.ChangeTaskState(enumor.TaskFailed, err.Error()); err != nil {
		logs.Errorf("[async] change task state failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] do run before failed end with rid %s", t.kt.Rid)
	return nil
}

func (t *Task) doRetryBefore(action Action) error {
	logs.V(3).Infof("[async] do run retry before start with rid %s", t.kt.Rid)

	err := action.RetryBefore(t.kt, t.Params)
	if err != nil {
		logs.Errorf("[async] retry before func failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 重试次数减1
	t.RetryCount = t.RetryCount - 1

	// 执行成功设置任务状态为TaskPending
	if err := t.ChangeTaskState(enumor.TaskPending, constant.AsyncDefaultJson); err != nil {
		logs.Errorf("[async] change task state failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	logs.V(3).Infof("[async] do run retry before end with rid %s", t.kt.Rid)
	return nil
}

// doRetry 判断是否可以重试，如果可以则执行RetryBefore等待继续调度，否则执行RunBeforeFailed，
func (t *Task) doRetry(action Action) error {
	if t.RetryCount != 0 {
		return t.doRetryBefore(action)
	}
	return t.doRunBeforeFailed(action)
}

// ChangeTaskState change task state
func (t *Task) ChangeTaskState(state enumor.TaskState, reason string) error {
	// 校验状态值
	if err := state.ValidateBeforeState(t.State); err != nil {
		logs.Errorf("[async] task validate failed, err: %v, rid: %s", err, t.kt.Rid)
		return err
	}

	// 更新任务状态操作重试
	maxRetryCount := uint32(3)
	r := retry.NewRetryPolicy(uint(maxRetryCount), [2]uint{1000, 15000})
	var lastError error
	lastError = nil
	for r.RetryCount() < maxRetryCount {
		if err := t.backend.SetTaskStateWithReason(t.ID, state, reason); err != nil {
			logs.Errorf("[async] set task state with reason failed times %d, err: %v, rid: %s",
				r.RetryCount(), err, t.kt.Rid)
			lastError = err
			r.Sleep()
		}

		if lastError == nil {
			t.State = state
			break
		}
	}

	return lastError
}

// ConvTaskResultToTask conv taskresult to task
func ConvTaskResultToTask(taskResult []task.AsyncFlowTask) []Task {
	if len(taskResult) <= 0 {
		return nil
	}

	tasks := make([]Task, 0, len(taskResult))
	for _, one := range taskResult {
		task := Task{
			ID:          one.ID,
			FlowID:      one.FlowID,
			FlowName:    one.FlowName,
			ActionName:  one.ActionName,
			Params:      json.RawMessage(one.Params),
			RetryCount:  one.RetryCount,
			TimeoutSecs: one.TimeoutSecs,
			DependOn:    one.DependOn,
			State:       one.State,
			Memo:        one.Memo,
		}
		tasks = append(tasks, task)
	}

	return tasks
}
