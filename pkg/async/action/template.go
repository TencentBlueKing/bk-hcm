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

// FlowTemplate define template.
type FlowTemplate struct {
	Name      enumor.FlowTplName    `json:"name" validate:"required"`
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

// TaskTemplate define task template.
type TaskTemplate struct {
	ActionID   string            `json:"action_id" validate:"required"`
	ActionName enumor.ActionName `json:"action_name" validate:"required"`
	NeedParam  bool              `json:"need_param" validate:"omitempty"`
	ParamType  interface{}       `json:"param_type" validate:"omitempty"`
	CanRetry   bool              `json:"can_retry" validate:"omitempty"`
	DependOn   []string          `json:"depend_on" validate:"omitempty"`
}

// Validate TaskTemplate.
func (tpl *TaskTemplate) Validate() error {
	return validator.Validate.Struct(tpl)
}
