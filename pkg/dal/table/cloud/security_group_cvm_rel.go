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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// SecurityGroupCvmRelColumns defines all the security group cvm rel table's columns.
var SecurityGroupCvmRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, SecurityGroupCvmRelColumnDescriptor)

// SecurityGroupCvmRelColumnDescriptor is vpc security group rel table column descriptors.
var SecurityGroupCvmRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "cvm_id", NamedC: "cvm_id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// SecurityGroupCvmRelTable define security group cvm rel table.
type SecurityGroupCvmRelTable struct {
	ID              uint64     `db:"id" validate:"required" json:"id"`
	CvmID           string     `db:"cvm_id" validate:"required,lte=64" json:"cvm_id"`
	SecurityGroupID string     `db:"security_group_id" validate:"required,lte=64" json:"security_group_id"`
	Creator         string     `db:"creator" validate:"required,lte=64" json:"creator"`
	CreatedAt       types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
}

// TableName return cvm and security group rel table name.
func (t SecurityGroupCvmRelTable) TableName() table.Name {
	return table.SecurityGroupCvmTable
}

// InsertValidate security group and cvm rel table when insert.
func (t SecurityGroupCvmRelTable) InsertValidate() error {
	return validator.Validate.Struct(t)
}
