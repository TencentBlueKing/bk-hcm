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

// Package flow 异步任务流
package flow

import (
	"hcm/pkg/async/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// Flow 任务流
type Flow struct {
	// 任务流ID
	ID string `json:"id" validate:"required"`
	// 任务流名称
	Name string `json:"name" validate:"required"`
	// 任务流状态
	State enumor.FlowState `json:"state" validate:"required"`
	// 任务集合
	Tasks []task.Task `json:"tasks" validate:"required"`
	// 任务流描述
	Memo string `json:"memo" validate:"omitempty"`
}

// Validate Flow
func (t *Flow) Validate() error {
	return validator.Validate.Struct(t)
}
