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

import (
	"fmt"
)

// TaskManagementState is task management state.
type TaskManagementState string

const (
	// TaskManagementRunning is a state indicating that task management running.
	TaskManagementRunning TaskManagementState = "running"
	// TaskManagementSuccess is a state indicating that task management success.
	TaskManagementSuccess TaskManagementState = "success"
	// TaskManagementFailed is a state indicating that task management failed.
	TaskManagementFailed TaskManagementState = "failed"
	// TaskManagementDeliverPartial is a state indicating that task management deliver partial.
	TaskManagementDeliverPartial TaskManagementState = "deliver_partial"
	// TaskManagementCancel is a state indicating that task management cancel.
	TaskManagementCancel TaskManagementState = "cancel"
)

// TaskManagementSource is task management source.
type TaskManagementSource string

// Validate ...
func (t TaskManagementSource) Validate() error {
	switch t {
	case TaskManagementSourceSops, TaskManagementSourceExcel:
		return nil
	default:
		return fmt.Errorf("invalid task management source: %s", t)
	}
}

const (
	// TaskManagementSourceSops is a source indicating that sops.
	TaskManagementSourceSops TaskManagementSource = "sops"
	// TaskManagementSourceExcel is a source indicating that excel.
	TaskManagementSourceExcel TaskManagementSource = "excel"
)

// TaskManagementResource is task management resource.
type TaskManagementResource string

const (
	// TaskManagementResClb is a resource indicating that clb.
	TaskManagementResClb TaskManagementResource = "clb"
)

// TaskDetailState is task detail state.
type TaskDetailState string

const (
	// TaskDetailInit is a state indicating that task detail init.
	TaskDetailInit TaskDetailState = "init"
	// TaskDetailRunning is a state indicating that task detail running.
	TaskDetailRunning TaskDetailState = "running"
	// TaskDetailSuccess is a state indicating that task detail success.
	TaskDetailSuccess TaskDetailState = "success"
	// TaskDetailFailed is a state indicating that task detail failed.
	TaskDetailFailed TaskDetailState = "failed"
	// TaskDetailCancel is a state indicating that task detail cancel.
	TaskDetailCancel TaskDetailState = "cancel"
)

// TaskOperation is task detail Operation.
type TaskOperation string

const (
	// TaskCreateLayer4Listener is a task indicating that create layer4 listener.
	TaskCreateLayer4Listener TaskOperation = "create_layer4_listener"

	// TaskCreateLayer7Listener is a task indicating that create layer7 listener.
	TaskCreateLayer7Listener TaskOperation = "create_layer7_listener"

	// TaskBindingLayer7RS is a task indicating that binding layer7 rs.
	TaskBindingLayer7RS TaskOperation = "binding_layer7_rs"

	// TaskBindingLayer4RS is a task indicating that binding layer4 rs.
	TaskBindingLayer4RS TaskOperation = "binding_layer4_rs"

	// TaskCreateLayer7Rule is a task indicating that create layer7 rule.
	TaskCreateLayer7Rule TaskOperation = "create_layer7_rule"
)
