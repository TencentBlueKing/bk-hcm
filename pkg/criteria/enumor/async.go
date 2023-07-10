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

package enumor

import "fmt"

// AsyncTaskType is async task type.
type AsyncTaskType string

// Validate async task type is allowed.
func (v AsyncTaskType) Validate() error {
	switch v {
	case SingleAsyncTask:
	case GroupAsyncTask:
	case ChordAsyncTask:
	case ChainAsyncTask:
	case CronAsyncTask:
	default:
		return fmt.Errorf("unsupported async task: %s", v)
	}

	return nil
}

const (
	// single task
	SingleAsyncTask AsyncTaskType = "task"
	// group task
	GroupAsyncTask AsyncTaskType = "group"
	// chord task
	ChordAsyncTask AsyncTaskType = "chord"
	// chain task
	ChainAsyncTask AsyncTaskType = "chain"
	// cron task
	CronAsyncTask AsyncTaskType = "cron"
)

// AsyncTaskPriority is async task priority.
type AsyncTaskPriority uint8

// Validate async task priority is allowed.
func (v AsyncTaskPriority) Validate() error {
	switch v {
	case HighAsyncTaskPriority:
	case MiddleAsyncTaskPriority:
	case LowAsyncTaskPriority:
	default:
		return fmt.Errorf("unsupported async priority: %d", v)
	}

	return nil
}

const (
	// high priority
	HighAsyncTaskPriority AsyncTaskPriority = 1
	// middle priority
	MiddleAsyncTaskPriority AsyncTaskPriority = 3
	// low priority
	LowAsyncTaskPriority AsyncTaskPriority = 5
)
