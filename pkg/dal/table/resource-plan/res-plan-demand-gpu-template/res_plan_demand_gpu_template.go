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

// Package resplandemandgputemplate ...
package resplandemandgputemplate

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanDemandGpuTemplateColumns defines all the res_plan_demand_gpu_template table's columns.
var ResPlanDemandGpuTemplateColumns = utils.MergeColumns(nil, ResPlanDemandGpuTemplateColumnDescriptor)

// ResPlanDemandGpuTemplateColumnDescriptor is ResPlanDemandGpuTemplateTable's column descriptors.
var ResPlanDemandGpuTemplateColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "tpl_schema", NamedC: "tpl_schema", Type: enumor.Json},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanDemandGpuTemplateTable is used to save res_plan_demand_gpu_template's data.
type ResPlanDemandGpuTemplateTable struct {
	// ID 模版ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Schema 模版内容(一个Excel对应一条记录) schema是关键字，这里需要用tpl_schema(template schema)代替
	TplSchema types.JsonField `db:"tpl_schema" json:"tpl_schema"`
	// Remark 备注
	Remark string `db:"remark" json:"remark" validate:"lte=255"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator" validate:"lte=64"`
	// Reviser 修改人
	Reviser string `db:"reviser" json:"reviser" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the ResPlanDemandGpuTemplateTable's database table name.
func (r ResPlanDemandGpuTemplateTable) TableName() table.Name {
	return table.ResPlanDemandGpuTemplateTable
}

// InsertValidate validate res_plan_demand_gpu_template on insertion.
func (r ResPlanDemandGpuTemplateTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.TplSchema.IsEmpty() {
		return errors.New("schema can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate res_plan_demand_gpu_template on update.
func (r ResPlanDemandGpuTemplateTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	return nil
}
