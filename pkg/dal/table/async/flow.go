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
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AsyncFlowColumns defines all the async_flow table's columns.
var AsyncFlowColumns = utils.MergeColumns(nil, AsyncFlowTableColumnDescriptor)

// AsyncFlowTableColumnDescriptor is async_flow's column descriptors.
var AsyncFlowTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "reason", NamedC: "reason", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AsyncFlowTable define async_flow table.
type AsyncFlowTable struct {
	ID        string           `db:"id" json:"id" validate:"lte=64"`
	Name      string           `db:"name" json:"name"`
	State     enumor.FlowState `db:"state" json:"state"`
	Reason    types.JsonField  `db:"reason" json:"reason"`
	Memo      string           `db:"memo" json:"memo"`
	Creator   string           `db:"creator" json:"creator" validate:"lte=64"`
	Reviser   string           `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt time.Time        `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt time.Time        `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// AsyncFlowReason define async flow reason
type AsyncFlowReason struct {
	Message string `json:"message"`
}

// TableName return async_flow table name.
func (a AsyncFlowTable) TableName() table.Name {
	return table.AsyncFlowTable
}

// InsertValidate async_flow table when insert.
func (a AsyncFlowTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) == 0 {
		return errors.New("id is required")
	}

	if len(a.Name) == 0 {
		return errors.New("name is required")
	}

	if len(a.State) == 0 {
		return errors.New("state is required")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate async_flow table when update.
func (a AsyncFlowTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Name) != 0 {
		return errors.New("name can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
