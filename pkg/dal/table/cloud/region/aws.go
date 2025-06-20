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

// AwsRegionColumns defines all the Aws region table's columns.
var AwsRegionColumns = utils.MergeColumns(nil, AwsRegionColumnDescriptor)

// AwsRegionColumnDescriptor is AwsRegion's column descriptors.
var AwsRegionColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "endpoint", NamedC: "endpoint", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AwsRegionTable aws_region表
type AwsRegionTable struct {
	// ID 自增ID
	ID string `db:"id" validate:"len=0"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-"`
	// 云账号id
	AccountID string `db:"account_id" validate:"lte=64"`
	// RegionID 地区ID
	RegionID string `db:"region_id" validate:"max=32"`
	// RegionName 地区名称
	RegionName string `db:"region_name" validate:"max=64"`
	// Status 地区状态(opt-in-not-required、opted-in、not-opted-in)
	Status string `db:"status" validate:"max=32"`
	// Endpoint Aws的Endpoint
	Endpoint string `db:"endpoint" validate:"max=64"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return region table name.
func (v AwsRegionTable) TableName() table.Name {
	return table.AwsRegionTable
}

// InsertValidate validate region table on insert.
func (v AwsRegionTable) InsertValidate() error {
	if err := v.Vendor.Validate(); err != nil {
		return err
	}

	if len(v.Vendor) == 0 {
		return errors.New("vendor can not be empty")
	}

	if len(v.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}
	if len(v.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if len(v.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate region table on update.
func (v AwsRegionTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(v)
}
