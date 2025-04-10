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

package region

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// HuaWeiRegionColumns defines all the huawei region table's columns.
var HuaWeiRegionColumns = utils.MergeColumns(nil, HuaWeiRegionTableColumnDescriptor)

// HuaWeiRegionTableColumnDescriptor is HuaWeiRegion's column descriptors.
var HuaWeiRegionTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "service", NamedC: "service", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "locales_pt_br", NamedC: "locales_pt_br", Type: enumor.String},
	{Column: "locales_zh_cn", NamedC: "locales_zh_cn", Type: enumor.String},
	{Column: "locales_en_us", NamedC: "locales_en_us", Type: enumor.String},
	{Column: "locales_es_us", NamedC: "locales_es_us", Type: enumor.String},
	{Column: "locales_es_es", NamedC: "locales_es_es", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// HuaWeiRegionTable 华为地域表
type HuaWeiRegionTable struct {
	// ID 账号 ID
	ID string `db:"id"`
	// Service 资源类型
	Service string `db:"service"`
	// RegionID 地域ID
	RegionID string `db:"region_id"`
	// Type 地域类型
	Type string `db:"type"`
	// LocalesPtBr 地域国际化
	LocalesPtBr string `db:"locales_pt_br"`
	// LocalesZhCn 地域国际化
	LocalesZhCn string `db:"locales_zh_cn"`
	// LocalesEnUs 地域国际化
	LocalesEnUs string `db:"locales_en_us"`
	// LocalesEsUs 地域国际化
	LocalesEsUs string `db:"locales_es_us"`
	// LocalesEsEs 地域国际化
	LocalesEsEs string `db:"locales_es_es"`
	// Creator 创建者
	Creator string `db:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return huawei region table name.
func (h HuaWeiRegionTable) TableName() table.Name {
	return table.HuaWeiRegionTable
}

// InsertValidate huawei region table when insert.
func (t HuaWeiRegionTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.RegionID) == 0 {
		return errors.New("region_id is required")
	}

	if len(t.Type) == 0 {
		return errors.New("type is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate azure huawei region table when update.
func (t HuaWeiRegionTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
