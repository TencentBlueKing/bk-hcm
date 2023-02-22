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
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// VpcCvmRelColumns defines all the vpc cvm rel table's columns.
var VpcCvmRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, VpcCvmRelColumnDescriptor)

// VpcCvmRelColumnDescriptor is vpc cvm rel table column descriptors.
var VpcCvmRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "cvm_id", NamedC: "cvm_id", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// VpcCvmRelTable define vpc cvm rel table.
type VpcCvmRelTable struct {
	ID        uint64     `db:"id" validate:"required" json:"id"`
	CvmID     string     `db:"cvm_id" validate:"required,lte=64" json:"cvm_id"`
	VpcID     string     `db:"vpc_id" validate:"required,lte=64" json:"vpc_id"`
	Creator   string     `db:"creator" validate:"required,lte=64" json:"creator"`
	CreatedAt *time.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
}

// TableName return cvm and security group rel table name.
func (t VpcCvmRelTable) TableName() table.Name {
	return table.VpcCvmRelTable
}

// InsertValidate security group and cvm rel table when insert.
func (t VpcCvmRelTable) InsertValidate() error {
	return validator.Validate.Struct(t)
}
