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

package cloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// SecurityGroupCommonRelColumns defines all the security group common rel table's columns.
var SecurityGroupCommonRelColumns = utils.MergeColumns(
	utils.InsertWithoutPrimaryID, SecurityGroupCommonRelColumnDescriptor)

// SecurityGroupCommonRelColumnDescriptor is security group common rel table column descriptors.
var SecurityGroupCommonRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "res_vendor", NamedC: "res_vendor", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "priority", NamedC: "priority", Type: enumor.Numeric},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// SecurityGroupCommonRelTable define security group common rel table.
type SecurityGroupCommonRelTable struct {
	ID              uint64                   `db:"id" json:"id"`
	ResVendor       enumor.Vendor            `db:"res_vendor" validate:"lte=16" json:"res_vendor"`
	ResID           string                   `db:"res_id" validate:"lte=64" json:"res_id"`
	ResType         enumor.CloudResourceType `db:"res_type" validate:"required,lte=64" json:"res_type"`
	Priority        int64                    `db:"priority" validate:"min=0,max=65535" json:"priority"`
	SecurityGroupID string                   `db:"security_group_id" validate:"required,lte=64" json:"security_group_id"`
	Creator         string                   `db:"creator" validate:"lte=64" json:"creator"`
	Reviser         string                   `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt       types.Time               `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt       types.Time               `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return security group and common rel table name.
func (t SecurityGroupCommonRelTable) TableName() table.Name {
	return table.SecurityGroupCommonRelTable
}

// InsertValidate security group and common rel table when insert.
func (t SecurityGroupCommonRelTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ResVendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.ResID) == 0 {
		return errors.New("res_id is required")
	}

	if len(t.ResType) == 0 {
		return errors.New("res_type is required")
	}

	if len(t.SecurityGroupID) == 0 {
		return errors.New("security_group_id is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate load_balancer table when update.
func (t SecurityGroupCommonRelTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
