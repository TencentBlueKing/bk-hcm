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
	"time"

	"hcm/pkg/criteria/validator"
)

// Option defines consumer run option.
type Option struct {
	Scheduler  *SchedulerOption  `json:"scheduler" validate:"required"`
	Executor   *ExecutorOption   `json:"executor" validate:"required"`
	Dispatcher *DispatcherOption `json:"dispatcher" validate:"required"`
	WatchDog   *WatchDogOption   `json:"watch_dog" validate:"required"`
}

// Validate Option
func (opt Option) Validate() error {
	return validator.Validate.Struct(opt)
}

// SchedulerOption 公共组件，负责获取分配给当前节点的任务流，并解析成任务树后，派发当前要执行的任务给executor执行
type SchedulerOption struct {
	WatchIntervalSec                uint `json:"watch_interval_sec" validate:"required"`
	WorkerNumber                    uint `json:"worker_number" validate:"required"`
	ScheduledFlowFetcherConcurrency uint `json:"scheduled_flow_fetcher_concurrency" validate:"required"`
	CanceledFlowFetcherConcurrency  uint `json:"canceled_flow_fetcher_concurrency" validate:"required"`
}

// Validate SchedulerOption.
func (opt SchedulerOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ExecutorOption 公共组件，负责执行异步任务
type ExecutorOption struct {
	WorkerNumber       uint `json:"worker_number" validate:"required"`
	TaskExecTimeoutSec uint `json:"task_exec_timeout_sec" validate:"required"`
}

// Validate ExecutorOption
func (opt ExecutorOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// DispatcherOption 主节点组件，负责派发任务
type DispatcherOption struct {
	WatchIntervalSec              uint `json:"watch_interval_sec" validate:"required"`
	PendingFlowFetcherConcurrency uint `json:"pending_flow_fetcher_concurrency" validate:"required"`
}

// Validate DispatcherOption
func (opt DispatcherOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// WatchDogOption 主节点组件，负责异常任务修正（超时任务，任务处理节点已经挂掉的任务等）
type WatchDogOption struct {
	WatchIntervalSec    uint `json:"watch_interval_sec" validate:"required"`
	TaskRunTimeoutSec   uint `json:"task_run_timeout_sec" validate:"required"`
	ShutdownWaitTimeSec uint `json:"shutdown_wait_time_sec" validate:"required"`
	WorkerNumber        uint `json:"worker_number" validate:"required"`
}

// Validate WatchDogOption
func (opt WatchDogOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SleepPolicy defines the policy of loop interval with different scenario.
type SleepPolicy struct {
	baseInterval time.Duration
}

// ShortSleep sleep with a short time
func (sp SleepPolicy) ShortSleep() {
	time.Sleep(sp.baseInterval / 3)
}

// ExceptionSleep sleep when the exception occurs.
func (sp SleepPolicy) ExceptionSleep() {
	time.Sleep(sp.baseInterval + sp.baseInterval/2)
}

// NormalSleep sleep with the default interval.
func (sp SleepPolicy) NormalSleep() {
	time.Sleep(sp.baseInterval)
}
