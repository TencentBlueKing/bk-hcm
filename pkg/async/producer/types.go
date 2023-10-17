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

package producer

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
)

// AddFlowOption define add flow option.
type AddFlowOption struct {
	// Name 任务流模版名称
	Name enumor.FlowTplName `json:"name" validate:"required"`
	// Memo 备注
	Memo string `json:"memo" validate:"omitempty"`
	// Tasks 任务私有化参数设置
	Tasks []Task `json:"tasks" validate:"omitempty"`
}

// Validate AddFlowOption
func (opt *AddFlowOption) Validate() error {

	if err := opt.Name.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(opt)
}

// Task define task info.
type Task struct {
	// ActionID 任务在当前任务流模版中的唯一ID
	ActionID string `json:"action_id" validate:"required"`
	// Params 任务执行请求参数
	Params types.JsonField `json:"params" validate:"required"`
}

// Validate Task
func (task *Task) Validate() error {
	return validator.Validate.Struct(task)
}
