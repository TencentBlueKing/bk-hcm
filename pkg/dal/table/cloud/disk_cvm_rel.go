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

// DiskCvmRelTableName 表名
var DiskCvmRelTableName table.Name = "disk_cvm_rel"

// DiskCvmRelColumns ...
var DiskCvmRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, DiskCvmRelColumnDescriptor)

// DiskCvmRelColumnDescriptor ...
var DiskCvmRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "disk_id", NamedC: "disk_id", Type: enumor.String},
	{Column: "cvm_id", NamedC: "cvm_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// DiskCvmRelTable 云盘主机关联表
type DiskCvmRelTable struct {
	model *DiskCvmRelModel
}

// TableName ...
func (t *DiskCvmRelTable) TableName() table.Name {
	return DiskCvmRelTableName
}

// DiskCvmRelModel 云盘主机关联数据模型
type DiskCvmRelModel struct {
	ID        uint64     `db:"id" json:"id"`
	DiskID    string     `db:"disk_id" validate:"required,lte=64" json:"disk_id"`
	CvmID     string     `db:"cvm_id" validate:"required,lte=64" json:"cvm_id"`
	Creator   string     `db:"creator" validate:"required,lte=64" json:"creator"`
	CreatedAt types.Time `db:"created_at" json:"created_at"`
}

// InsertValidate ...
func (m *DiskCvmRelModel) InsertValidate() error {
	return validator.Validate.Struct(m)
}
