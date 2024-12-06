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

package task

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// DetailColumns defines all the task detail table's columns.
var DetailColumns = utils.MergeColumns(nil, DetailColumnDescriptor)

// DetailColumnDescriptor is task detail column descriptors.
var DetailColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "task_management_id", NamedC: "task_management_id", Type: enumor.String},
	{Column: "flow_id", NamedC: "flow_id", Type: enumor.String},
	{Column: "task_action_ids", NamedC: "task_action_ids", Type: enumor.Json},
	{Column: "operation", NamedC: "operation", Type: enumor.String},
	{Column: "param", NamedC: "param", Type: enumor.Json},
	{Column: "result", NamedC: "result", Type: enumor.Json},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "reason", NamedC: "reason", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DetailTable is used to save task detail information.
type DetailTable struct {
	ID               string                 `db:"id" json:"id"`
	BkBizID          int64                  `db:"bk_biz_id" json:"bk_biz_id"`
	TaskManagementID string                 `db:"task_management_id" validate:"lte=64" json:"task_management_id"`
	FlowID           string                 `db:"flow_id" validate:"lte=64" json:"flow_id"`
	TaskActionIDs    types.StringArray      `db:"task_action_ids" json:"task_action_ids"`
	Operation        enumor.TaskOperation   `db:"operation" validate:"lte=64" json:"operation"`
	Param            types.JsonField        `db:"param" json:"param"`
	Result           types.JsonField        `db:"result" json:"result"`
	State            enumor.TaskDetailState `db:"state" validate:"lte=16" json:"state"`
	Reason           string                 `db:"reason" validate:"lte=1024" json:"reason"`
	Extension        types.JsonField        `db:"extension" json:"extension"`
	Creator          string                 `db:"creator" validate:"max=64" json:"creator"`
	Reviser          string                 `db:"reviser" validate:"max=64" json:"reviser"`
	CreatedAt        types.Time             `db:"created_at" validate:"isdefault" json:"created_at"`
	UpdatedAt        types.Time             `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the task detail database table name.
func (d DetailTable) TableName() table.Name {
	return table.TaskDetailTable
}

// InsertValidate validate task detail on insertion.
func (d DetailTable) InsertValidate() error {
	if err := validator.Validate.Struct(d); err != nil {
		return err
	}

	if len(d.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(d.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate task detail on update.
func (d DetailTable) UpdateValidate() error {
	if err := validator.Validate.Struct(d); err != nil {
		return err
	}

	if len(d.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(d.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
