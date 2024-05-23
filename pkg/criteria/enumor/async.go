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

const (
	// TaskInit task state is init
	TaskInit TaskState = "init"
	// TaskPending task state is pending
	TaskPending TaskState = "pending"
	// TaskRunning task state is running
	TaskRunning TaskState = "running"
	// TaskRollback task state is rollback.
	TaskRollback TaskState = "rollback"
	// TaskCancel task state is cancel.
	TaskCancel TaskState = "canceled"
	// TaskSuccess task state is success
	TaskSuccess TaskState = "success"
	// TaskFailed task state is failed
	TaskFailed TaskState = "failed"
)

// FlowState is flow state.
type FlowState string

const (
	// FlowInit flow state is init（该状态不参与调度）
	FlowInit FlowState = "init"
	// FlowPending flow state is pending
	FlowPending FlowState = "pending"
	// FlowScheduled flow state is scheduled
	FlowScheduled FlowState = "scheduled"
	// FlowRunning flow state is running
	FlowRunning FlowState = "running"
	// FlowCancel flow state is cancel
	FlowCancel FlowState = "canceled"
	// FlowSuccess flow state is success
	FlowSuccess FlowState = "success"
	// FlowFailed flow state is failed
	FlowFailed FlowState = "failed"
)

// BackendType is backend type.
type BackendType string

// Validate BackendType.
func (v BackendType) Validate() error {
	switch v {
	case BackendMysql:
	default:
		return fmt.Errorf("unsupported backend type: %s", v)
	}

	return nil
}

const (
	// BackendMysql mysql backend
	BackendMysql BackendType = "mysql"
)
