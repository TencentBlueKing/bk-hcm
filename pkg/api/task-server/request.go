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

package taskserver

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// AsyncTask async task
type AsyncTask struct {
	TaskType     enumor.AsyncTaskType `json:"task_type" validate:"required"`
	CallBackTask *Step                `json:"call_back_task" validate:"omitempty"`
	Steps        []*Step              `json:"steps" validate:"min=1,required"`
}

// Validate ...
func (a *AsyncTask) Validate() error {
	if a.TaskType.Validate() != nil {
		return errors.New("async task type is not allowed")
	}

	if a.TaskType == enumor.SingleAsyncTask && len(a.Steps) > 1 {
		return errors.New("async steps is not match single task type")
	}

	if a.TaskType == enumor.ChordAsyncTask && a.CallBackTask == nil {
		return errors.New("async steps is not match chord task type")
	}

	return validator.Validate.Struct(a)
}

// Step step is task
type Step struct {
	TaskName       string                   `json:"task_name" validate:"required"`
	TaskPriority   enumor.AsyncTaskPriority `json:"task_priority" validate:"required"`
	Args           []Arg                    `json:"args" validate:"omitempty"`
	RetryCount     int                      `json:"retry_count" validate:"omitempty"`
	UUID           string                   `json:"uuid" validate:"omitempty"`
	GroupUUID      string                   `json:"group_uuid" validate:"omitempty"`
	GroupTaskCount int                      `json:"group_task_count" validate:"omitempty"`
	OnSuccess      []*Step                  `json:"on_success" validate:"omitempty"`
	OnError        []*Step                  `json:"on_error" validate:"omitempty"`
	// set Immutable false means send A task result to B task
	Immutable bool `json:"immutable" validate:"omitempty"`
}

// Validate ...
func (step *Step) Validate() error {
	if step.TaskPriority.Validate() != nil {
		return errors.New("async task priority is not allowed")
	}

	return validator.Validate.Struct(step)
}

// Arg func arg
type Arg struct {
	Type  string      `json:"type" validate:"required"`
	Value interface{} `json:"value" validate:"required"`
}

// Validate ...
func (arg *Arg) Validate() error {
	return validator.Validate.Struct(arg)
}
