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

package loadbalancer

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ClbFlowRelColumns defines all the clb_flow_rel table's columns.
var ClbFlowRelColumns = utils.MergeColumns(nil, ClbFlowRelColumnsDescriptor)

// ClbFlowRelColumnsDescriptor is clb_flow_rel's column descriptors.
var ClbFlowRelColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "flow_id", NamedC: "flow_id", Type: enumor.String},
	{Column: "task_type", NamedC: "task_type", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ClbFlowRelTable CLB与FlowID关系表
type ClbFlowRelTable struct {
	ID       string `db:"id" validate:"lte=64" json:"id"`
	ResID    string `db:"res_id" validate:"lte=64" json:"res_id"`
	FlowID   string `db:"flow_id" validate:"lte=64" json:"flow_id"`
	TaskType string `db:"task_type" validate:"lte=64" json:"task_type"`
	Status   string `db:"status" validate:"lte=64" json:"status"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return clb_flow_rel table name.
func (cft ClbFlowRelTable) TableName() table.Name {
	return table.ClbFlowRelTable
}

// InsertValidate clb_flow_rel table when insert.
func (cft ClbFlowRelTable) InsertValidate() error {
	if err := validator.Validate.Struct(cft); err != nil {
		return err
	}

	if len(cft.ResID) == 0 {
		return errors.New("res_id is required")
	}

	if len(cft.FlowID) == 0 {
		return errors.New("flow_id is required")
	}

	if len(cft.Status) == 0 {
		return errors.New("status is required")
	}

	if len(cft.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate clb_flow_rel table when update.
func (cft ClbFlowRelTable) UpdateValidate() error {
	if err := validator.Validate.Struct(cft); err != nil {
		return err
	}

	if len(cft.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(cft.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
