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

// TCloudZiyanRegionColumns defines all the TCloud region table's columns.
var TCloudZiyanRegionColumns = utils.MergeColumns(nil, TCloudZiyanRegionColumnDescriptor)

// TCloudZiyanRegionColumnDescriptor is TCloudZiyanRegion's column descriptors.
var TCloudZiyanRegionColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "area_name", NamedC: "area_name", Type: enumor.String},
	{Column: "city_name", NamedC: "city_name", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TCloudZiyanRegionTable tcloud_region表
type TCloudZiyanRegionTable struct {
	// ID 自增ID
	ID string `db:"id" validate:"len=0"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-"`
	// RegionID 地区ID
	RegionID string `db:"region_id" validate:"max=32"`
	// RegionName 地区名称，例如：华东地区(上海)
	RegionName string `db:"region_name" validate:"max=64"`
	// AreaName 地域名称，例如：华东地区
	AreaName string `db:"area_name" validate:"lte=64"`
	// CityName 城市名称，例如：上海
	CityName string `db:"city_name" validate:"lte=64"`
	// Status 地区状态(AVAILABLE:可用)
	Status string `db:"status" validate:"max=32"`
	// Source 数据来源：sync-同步，manually-手动添加
	Source enumor.RegionSource `db:"source" validate:"-"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless"`
}

// TableName return region table name.
func (v TCloudZiyanRegionTable) TableName() table.Name {
	return table.TCloudZiyanRegionTable
}

// InsertValidate validate region table on insert.
func (v TCloudZiyanRegionTable) InsertValidate() error {
	if err := v.Vendor.Validate(); err != nil {
		return err
	}

	if len(v.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(v.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(v.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	if len(v.CityName) == 0 {
		return errors.New("city name can not be empty")
	}

	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if err := v.Source.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate region table on update.
func (v TCloudZiyanRegionTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	// vendor, account id can not update
	if len(v.Vendor) > 0 {
		return errors.New("vendor can not update")
	}

	if len(v.Source) > 0 {
		if err := v.Source.Validate(); err != nil {
			return err
		}
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
