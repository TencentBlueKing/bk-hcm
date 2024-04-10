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

package tablelb

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResourceFlowLockColumns defines all the resource_flow_lock table's columns.
var ResourceFlowLockColumns = utils.MergeColumns(nil, ResourceFlowLockColumnsDescriptor)

// ResourceFlowLockColumnsDescriptor is resource_flow_lock column descriptors.
var ResourceFlowLockColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "owner", NamedC: "owner", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResourceFlowLockTable 资源与Flow的锁定表
type ResourceFlowLockTable struct {
	ResID   string                   `db:"res_id" validate:"lte=64" json:"res_id"`
	ResType enumor.CloudResourceType `db:"res_type" validate:"lte=64" json:"res_type"`
	Owner   string                   `db:"owner" validate:"lte=64" json:"owner"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return table name.
func (cfl ResourceFlowLockTable) TableName() table.Name {
	return table.ResourceFlowLockTable
}

// InsertValidate validate table when insert.
func (cfl ResourceFlowLockTable) InsertValidate() error {
	if err := validator.Validate.Struct(cfl); err != nil {
		return err
	}

	if len(cfl.ResID) == 0 {
		return errors.New("res_id is required")
	}

	if len(cfl.ResType) == 0 {
		return errors.New("res_type is required")
	}

	if len(cfl.Owner) == 0 {
		return errors.New("owner is required")
	}

	if len(cfl.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate validate table when update.
func (cfl ResourceFlowLockTable) UpdateValidate() error {
	if err := validator.Validate.Struct(cfl); err != nil {
		return err
	}

	if len(cfl.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(cfl.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
