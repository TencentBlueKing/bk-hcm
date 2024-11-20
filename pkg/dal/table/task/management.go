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

// ManagementColumns defines all the task management table's columns.
var ManagementColumns = utils.MergeColumns(nil, ManagementColumnDescriptor)

// ManagementColumnDescriptor is task management column descriptors.
var ManagementColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "vendors", NamedC: "vendors", Type: enumor.Json},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "account_ids", NamedC: "account_ids", Type: enumor.Json},
	{Column: "resource", NamedC: "resource", Type: enumor.String},
	{Column: "operations", NamedC: "operations", Type: enumor.Json},
	{Column: "flow_ids", NamedC: "flow_ids", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ManagementTable is used to save task management information.
type ManagementTable struct {
	ID         string                        `db:"id" json:"id"`
	BkBizID    int64                         `db:"bk_biz_id" json:"bk_biz_id"`
	Source     enumor.TaskManagementSource   `db:"source" validate:"lte=16" json:"source"`
	Vendors    types.StringArray             `db:"vendors" json:"vendors"`
	State      enumor.TaskManagementState    `db:"state" validate:"lte=16" json:"state"`
	AccountIDs types.StringArray             `db:"account_ids" json:"account_ids"`
	Resource   enumor.TaskManagementResource `db:"resource" validate:"lte=16" json:"resource"`
	Operations types.StringArray             `db:"operations"  json:"operations"`
	FlowIDs    types.StringArray             `db:"flow_ids"  json:"flow_ids"`
	Extension  types.JsonField               `db:"extension" json:"extension"`
	Creator    string                        `db:"creator" validate:"max=64" json:"creator"`
	Reviser    string                        `db:"reviser" validate:"max=64" json:"reviser"`
	CreatedAt  types.Time                    `db:"created_at" validate:"isdefault" json:"created_at"`
	UpdatedAt  types.Time                    `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the task management database table name.
func (m ManagementTable) TableName() table.Name {
	return table.TaskManagementTable
}

// InsertValidate validate task management on insertion.
func (m ManagementTable) InsertValidate() error {
	if err := validator.Validate.Struct(m); err != nil {
		return err
	}

	if len(m.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(m.Operations) == 0 {
		return errors.New("operations can not be empty")
	}

	if len(m.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate task management on update.
func (m ManagementTable) UpdateValidate() error {
	if err := validator.Validate.Struct(m); err != nil {
		return err
	}

	if len(m.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(m.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
