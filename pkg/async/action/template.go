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

package action

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	tableasync "hcm/pkg/dal/table/async"
)

// FlowTemplate 任务流模版定义，用于定义执行任务流模版，用户根据任务流 Name 创建任务流实例去创建异步任务。
type FlowTemplate struct {
	Name      enumor.FlowName       `json:"name" validate:"required"`
	ShareData *tableasync.ShareData `json:"share_data"`
	Tasks     []TaskTemplate        `json:"tasks" validate:"required,min=1"`
}

// Validate FlowTemplate.
func (tpl *FlowTemplate) Validate() error {
	if err := validator.Validate.Struct(tpl); err != nil {
		return err
	}

	for _, one := range tpl.Tasks {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ActIDType define action id type.
type ActIDType string

// TaskTemplate 任务模版定义，用于定义任务流模版中用到的任务的组织关系、请求参数、重试参数等.
type TaskTemplate struct {
	ActionID   ActIDType         `json:"action_id" validate:"required"`
	ActionName enumor.ActionName `json:"action_name" validate:"required"`
	DependOn   []ActIDType       `json:"depend_on" validate:"omitempty"`

	// Params 异步任务运行请求参数相关控制参数。
	Params *Params `json:"params" validate:"omitempty"`

	// Retry 任务运行重试相关配置参数，如果不设置，默认不允许进行重试。
	Retry *tableasync.Retry `json:"retry" validate:"omitempty"`
}

// Validate TaskTemplate.
func (tpl *TaskTemplate) Validate() error {
	if err := validator.Validate.Struct(tpl); err != nil {
		return err
	}

	if tpl.Retry != nil {
		if err := tpl.Retry.Validate(); err != nil {
			return err
		}
	} else {
		tpl.Retry = new(tableasync.Retry)
	}

	if tpl.Params != nil {
		if err := tpl.Params.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Params 异步任务参数相关控制参数
type Params struct {
	// Type 参数类型
	Type interface{} `json:"type" validate:"required"`
}

// Validate Params.
func (p Params) Validate() error {
	return validator.Validate.Struct(p)
}
