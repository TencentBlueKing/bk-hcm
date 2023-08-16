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

// TaskState is task state.
type TaskState string

// Validate TaskState.
func (v TaskState) Validate() error {
	switch v {
	case TaskPending:
	case TaskRunning:
	case TaskBeforeSuccess:
	case TaskSuccess:
	case TaskBeforeFailed:
	case TaskFailed:
	default:
		return fmt.Errorf("unsupported task state type: %s", v)
	}

	return nil
}

// ValidateBeforeState validate before state.
func (v TaskState) ValidateBeforeState(beforeState TaskState) error {
	switch v {
	case TaskPending:
		if beforeState != TaskPending &&
			beforeState != TaskRunning &&
			beforeState != TaskBeforeSuccess {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case TaskRunning:
		if beforeState != TaskPending {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case TaskBeforeSuccess:
		if beforeState != TaskRunning {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case TaskSuccess:
		if beforeState != TaskBeforeSuccess {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case TaskBeforeFailed:
		if beforeState != TaskPending &&
			beforeState != TaskRunning &&
			beforeState != TaskBeforeSuccess {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case TaskFailed:
		if beforeState != TaskBeforeFailed {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	default:
		return fmt.Errorf("unsupported task state type: %s", v)
	}

	return nil
}

const (
	// TaskPending task state is pending
	TaskPending TaskState = "pending"
	// TaskRunning task state is running
	TaskRunning TaskState = "running"
	// TaskBeforeSuccess task state is before-success
	TaskBeforeSuccess TaskState = "before-success"
	// TaskSuccess task state is success
	TaskSuccess TaskState = "success"
	// TaskBeforeFailed task state is before-failed
	TaskBeforeFailed TaskState = "before-failed"
	// TaskFailed task state is failed
	TaskFailed TaskState = "failed"
)

// FlowState is flow state.
type FlowState string

// Validate FlowState.
func (v FlowState) Validate() error {
	switch v {
	case FlowPending:
	case FlowRunning:
	case FlowSuccess:
	case FlowFailed:
	default:
		return fmt.Errorf("unsupported flow state type: %s", v)
	}

	return nil
}

// ValidateBeforeState validate before state.
func (v FlowState) ValidateBeforeState(beforeState FlowState) error {
	switch v {
	case FlowPending:
	case FlowRunning:
		if beforeState != FlowPending {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case FlowSuccess:
		if beforeState != FlowRunning {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	case FlowFailed:
		if beforeState != FlowRunning {
			return fmt.Errorf("state type %s check before state %s failed", v, beforeState)
		}
	default:
		return fmt.Errorf("unsupported flow state type: %s", v)
	}

	return nil
}

const (
	// FlowPending flow state is pending
	FlowPending FlowState = "pending"
	// FlowRunning flow state is running
	FlowRunning FlowState = "running"
	// FlowSuccess flow state is success
	FlowSuccess FlowState = "success"
	// FlowFailed flow state is failed
	FlowFailed FlowState = "failed"
)
