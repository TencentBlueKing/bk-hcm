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

package resplandemandgpuorder

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanDemandGpuOrderColumns defines all the res_plan_demand_gpu_order table's columns.
var ResPlanDemandGpuOrderColumns = utils.MergeColumns(nil, ResPlanDemandGpuOrderColumnDescriptor)

// ResPlanDemandGpuOrderColumnDescriptor is ResPlanDemandGpuOrderTable's column descriptors.
var ResPlanDemandGpuOrderColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "template_id", NamedC: "template_id", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanDemandGpuOrderTable is used to save GPU需求提报主单信息.
type ResPlanDemandGpuOrderTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// OpProductName 运营产品名称
	OpProductName string `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	// TemplateID 模版ID
	TemplateID string `db:"template_id" json:"template_id" validate:"lte=32"`
	// Status 状态
	Status enumor.ResPlanDemandGpuOrderStatus `db:"status" json:"status" validate:"lte=32"`
	// Remark 备注
	Remark string `db:"remark" json:"remark" validate:"lte=255"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser 修改者
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at" validate:"isdefault"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at" validate:"isdefault"`
}

// TableName is the res_plan_demand_gpu_order's database table name.
func (r ResPlanDemandGpuOrderTable) TableName() table.Name {
	return table.ResPlanDemandGpuOrderTable
}

// InsertValidate validate on insertion.
func (r ResPlanDemandGpuOrderTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.BkBizID <= 0 {
		return errors.New("bk_biz_id should be > 0")
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

	if err := r.Status.Validate(); err != nil {
		return err
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate on update.
func (r ResPlanDemandGpuOrderTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	return nil
}
