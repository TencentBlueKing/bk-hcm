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

// Package resplandemandgpusuborder ...
package resplandemandgpusuborder

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanDemandGpuSubOrderColumns defines all the res_plan_demand_gpu_suborder table's columns.
var ResPlanDemandGpuSubOrderColumns = utils.MergeColumns(nil, ResPlanDemandGpuSubOrderColumnDescriptor)

// ResPlanDemandGpuSubOrderColumnDescriptor is ResPlanDemandGpuSubOrderTable's column descriptors.
var ResPlanDemandGpuSubOrderColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "order_id", NamedC: "order_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "demand_type", NamedC: "demand_type", Type: enumor.String},
	{Column: "demand_year", NamedC: "demand_year", Type: enumor.Numeric},
	{Column: "demand_month", NamedC: "demand_month", Type: enumor.Numeric},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "template_id", NamedC: "template_id", Type: enumor.String},
	{Column: "gpu_num", NamedC: "gpu_num", Type: enumor.Numeric},
	{Column: "qpm_max", NamedC: "qpm_max", Type: enumor.Numeric},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "comment", NamedC: "comment", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanDemandGpuSubOrderTable is used to save resource plan demand gpu sub order information.
type ResPlanDemandGpuSubOrderTable struct {
	// ID 需求子单ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// OrderID 需求主单ID
	OrderID string `db:"order_id" json:"order_id" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// OpProductName 运营产品名称
	OpProductName string `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	// TemplateID 模版ID
	TemplateID string `db:"template_id" json:"template_id" validate:"lte=32"`
	// DemandType 需求分类
	DemandType string `db:"demand_type" json:"demand_type" validate:"lte=64"`
	// DemandYear 需求年
	DemandYear int64 `db:"demand_year" json:"demand_year"`
	// DemandMonth 需求月
	DemandMonth int64 `db:"demand_month" json:"demand_month"`
	// GPUNum GPU预算卡数
	GPUNum int64 `db:"gpu_num" json:"gpu_num"`
	// QpmMax 峰值QPM
	QpmMax int64 `db:"qpm_max" json:"qpm_max"`
	// Status 状态
	Status enumor.RPDemandGPUSubOrderStatus `db:"status" json:"status" validate:"lte=32"`
	// Comment 评审意见
	Comment *types.JsonField `db:"comment" json:"comment"`
	// Extension 扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
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

// TableName is the ResPlanDemandGpuSubOrderTable's database table name.
func (r ResPlanDemandGpuSubOrderTable) TableName() table.Name {
	return table.ResPlanDemandGpuSubOrderTable
}

// InsertValidate validate resource plan demand gpu sub order on insertion.
func (r ResPlanDemandGpuSubOrderTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 || len(r.OrderID) == 0 {
		return errors.New("id and order_id can not be empty")
	}

	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if r.OpProductID <= 0 {
		return errors.New("op_product_id should be > 0")
	}

	if len(r.OpProductName) == 0 {
		return errors.New("op_product_name can not be empty")
	}

	if len(r.TemplateID) == 0 {
		return errors.New("template_id can not be empty")
	}

	if len(r.DemandType) == 0 {
		return errors.New("demand_type can not be empty")
	}

	if r.DemandYear < 0 {
		return errors.New("demand_year should be >= 0")
	}

	if r.DemandMonth < 0 || r.DemandMonth > 12 {
		return errors.New("demand_month should be >= 0 and <= 12")
	}

	if r.GPUNum < 0 {
		return errors.New("gpu_num should be >= 0")
	}

	if r.QpmMax < 0 {
		return errors.New("qpm_max should be >= 0")
	}

	if err := r.Status.Validate(); err != nil {
		return err
	}

	if r.Extension.IsEmpty() {
		return errors.New("extension can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate resource plan demand gpu sub order on update.
func (r ResPlanDemandGpuSubOrderTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.OrderID) != 0 {
		return errors.New("order_id can not update")
	}

	if r.BkBizID < 0 {
		return errors.New("bk biz id should be >= 0")
	}

	if r.OpProductID < 0 {
		return errors.New("op_product_id should be >= 0")
	}

	if r.DemandYear < 0 {
		return errors.New("demand_year should be >= 0")
	}

	if r.DemandMonth < 0 || r.DemandMonth > 12 {
		return errors.New("demand_month should be >= 0 and <= 12")
	}

	if r.GPUNum < 0 {
		return errors.New("gpu_num should be >= 0")
	}

	if r.QpmMax < 0 {
		return errors.New("qpm_max should be >= 0")
	}

	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
