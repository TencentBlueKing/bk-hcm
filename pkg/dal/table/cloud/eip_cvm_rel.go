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

// EipCvmRelTableName 表名
var EipCvmRelTableName table.Name = "eip_cvm_rel"

// EipCvmRelColumns ...
var EipCvmRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, EipCvmRelColumnDescriptor)

// EipCvmRelColumnDescriptor ...
var EipCvmRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "eip_id", NamedC: "eip_id", Type: enumor.String},
	{Column: "cvm_id", NamedC: "cvm_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// EipCvmRelTable Eip 主机关联表
type EipCvmRelTable struct {
	model *EipCvmRelModel
}

// TableName ...
func (t *EipCvmRelTable) TableName() table.Name {
	return EipCvmRelTableName
}

// EipCvmRelModel Eip 主机关联数据模型
type EipCvmRelModel struct {
	ID        uint64     `db:"id" validate:"required" json:"id"`
	EipID     string     `db:"eip_id" validate:"required,lte=64" json:"eip_id"`
	CvmID     string     `db:"cvm_id" validate:"required,lte=64" json:"cvm_id"`
	Creator   string     `db:"creator" validate:"required,lte=64" json:"creator"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}

// InsertValidate ...
func (m *EipCvmRelModel) InsertValidate() error {
	return validator.Validate.Struct(m)
}
