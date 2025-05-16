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

package model

import (
	"errors"

	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/dal/table/types"
)

// Task define task struct.
type Task struct {
	ID         string             `json:"id"`
	FlowID     string             `json:"flow_id"`
	FlowName   enumor.FlowName    `json:"flow_name"`
	ActionID   action.ActIDType   `json:"action_id"`
	ActionName enumor.ActionName  `json:"action_name"`
	Params     types.JsonField    `json:"params"`
	Retry      *tableasync.Retry  `json:"can_retry"`
	DependOn   []action.ActIDType `json:"depend_on"`
	State      enumor.TaskState   `json:"state"`
	Reason     *tableasync.Reason `json:"reason"`
	Result     types.JsonField    `json:"result"`
	TenantID   string             `json:"tenant_id"`
	Creator    string             `json:"creator"`
	Reviser    string             `json:"reviser"`
	CreatedAt  string             `json:"created_at"`
	UpdatedAt  string             `json:"updated_at"`
}

// CreateValidate Task create validate.
func (t *Task) CreateValidate() error {
	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.FlowID) == 0 {
		return errors.New("flow_id is required")
	}

	if len(t.FlowName) == 0 {
		return errors.New("flow_task is required")
	}

	if len(t.ActionID) == 0 {
		return errors.New("action_id is required")
	}

	if len(t.ActionName) == 0 {
		return errors.New("action_name is required")
	}

	if t.Reason != nil {
		return errors.New("reason can not set")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(t.CreatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	if len(t.UpdatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate Task update validate.
func (t *Task) UpdateValidate() error {

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.FlowID) != 0 {
		return errors.New("flow_id can not set")
	}

	if len(t.FlowName) != 0 {
		return errors.New("flow_task can not set")
	}

	if len(t.ActionID) != 0 {
		return errors.New("action_id can not set")
	}

	if len(t.ActionName) != 0 {
		return errors.New("action_name can not set")
	}

	if len(t.Params) != 0 {
		return errors.New("params can not set")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not set")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	if len(t.UpdatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}
