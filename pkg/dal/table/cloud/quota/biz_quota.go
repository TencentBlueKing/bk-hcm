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

package tablequota

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// BizQuotaColumns defines all the biz quota table's columns.
var BizQuotaColumns = utils.MergeColumns(nil, BizQuotaTableColumnDescriptor)

// BizQuotaTableColumnDescriptor 业务配额表字段描述
var BizQuotaTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_quota_id", NamedC: "cloud_quota_id", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "levels", NamedC: "levels", Type: enumor.Json},
	{Column: "dimension", NamedC: "dimension", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.String},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.String},
}

// BizQuotaTable 业务配额表
type BizQuotaTable struct {
	// ID ID
	ID string `db:"id" validate:"max=64" json:"id"`
	// CloudQuotaID 云配额ID
	CloudQuotaID string `db:"cloud_quota_id" validate:"max=64" json:"cloud_quota_id"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"max=64" json:"account_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
	// ResType 资源类型
	ResType enumor.BizQuotaResType `db:"res_type" validate:"max=16" json:"res_type"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-" json:"vendor"`
	// Region 地域
	Region string `db:"region" validate:"max=255" json:"region"`
	// Zone 可用区
	Zone string `db:"zone" validate:"max=255" json:"zone"`
	// Levels 配额层级
	Levels Levels `db:"levels" validate:"-" json:"levels"`
	// Dimension 维度
	Dimension Dimensions `db:"dimension" validate:"-" json:"dimension"`
	// Memo 备注
	Memo *string `db:"memo" validate:"omitempty,max=255" json:"memo"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName return biz quota table name.
func (t BizQuotaTable) TableName() table.Name {
	return table.BizQuotaTable
}

// InsertValidate validate biz quota table on insert.
func (t BizQuotaTable) InsertValidate() error {
	if err := t.Vendor.Validate(); err != nil {
		return err
	}

	if len(t.CloudQuotaID) == 0 {
		return errors.New("cloud quota id can not be empty")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if t.BkBizID == 0 {
		return errors.New("biz id can not be empty")
	}

	if len(t.ResType) == 0 {
		return errors.New("res type can not be empty")
	}

	if len(t.Levels) == 0 {
		return errors.New("levels can not be empty")
	}

	if len(t.Dimension) == 0 {
		return errors.New("dimension can not be empty")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(t)
}

// UpdateValidate validate vpc table on update.
func (t BizQuotaTable) UpdateValidate() error {
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
