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

package tableasync

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AsyncFlowTaskColumns defines all the async_flow_task table's columns.
var AsyncFlowTaskColumns = utils.MergeColumns(nil, AsyncFlowTaskTableColumnDescriptor)

// AsyncFlowTaskTableColumnDescriptor is async_flow_task's column descriptors.
var AsyncFlowTaskTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "flow_id", NamedC: "flow_id", Type: enumor.String},
	{Column: "flow_name", NamedC: "flow_name", Type: enumor.String},
	{Column: "action_id", NamedC: "action_id", Type: enumor.String},
	{Column: "action_name", NamedC: "action_name", Type: enumor.String},
	{Column: "params", NamedC: "params", Type: enumor.Json},
	{Column: "retry", NamedC: "retry", Type: enumor.Json},
	{Column: "depend_on", NamedC: "depend_on", Type: enumor.Json},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "reason", NamedC: "reason", Type: enumor.Json},
	{Column: "result", NamedC: "result", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AsyncFlowTaskTable define async_flow_task table.
type AsyncFlowTaskTable struct {
	ID         string            `db:"id" json:"id" validate:"lte=64"`
	FlowID     string            `db:"flow_id" json:"flow_id"`
	FlowName   enumor.FlowName   `db:"flow_name" json:"flow_name"`
	ActionID   string            `db:"action_id" json:"action_id"`
	ActionName enumor.ActionName `db:"action_name" json:"action_name"`
	Params     types.JsonField   `db:"params" json:"params"`
	Retry      *Retry            `db:"retry" json:"retry"`
	DependOn   types.StringArray `db:"depend_on" json:"depend_on"`
	State      enumor.TaskState  `db:"state" json:"state"`
	Reason     *Reason           `db:"reason" json:"reason"`
	Result     types.JsonField   `db:"result" json:"result"`
	Creator    string            `db:"creator" json:"creator" validate:"lte=64"`
	Reviser    string            `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt  types.Time        `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt  types.Time        `db:"updated_at" json:"updated_at" validate:"excluded_unless"`

	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return async_flow_task table name.
func (a AsyncFlowTaskTable) TableName() table.Name {
	return table.AsyncFlowTaskTable
}

// InsertValidate async_flow_task table when insert.
func (a AsyncFlowTaskTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) == 0 {
		return errors.New("id is required")
	}

	if len(a.FlowID) == 0 {
		return errors.New("flow_id is required")
	}

	if len(a.FlowName) == 0 {
		return errors.New("flow_name is required")
	}

	if len(a.ActionName) == 0 {
		return errors.New("action_name is required")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate async_flow_task table when update.
func (a AsyncFlowTaskTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.FlowID) != 0 {
		return errors.New("flow_id can not update")
	}

	if len(a.FlowName) != 0 {
		return errors.New("flow_name can not update")
	}

	if len(a.ActionName) != 0 {
		return errors.New("action_name can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
