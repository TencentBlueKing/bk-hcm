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

package rollingserver

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// RollingAppliedRecordColumns defines rolling_applied_record's columns.
var RollingAppliedRecordColumns = utils.MergeColumns(nil, RollingAppliedRecordColumnDescriptor)

// RollingAppliedRecordColumnDescriptor is column descriptors.
var RollingAppliedRecordColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "require_type", NamedC: "require_type", Type: enumor.Numeric},
	{Column: "applied_type", NamedC: "applied_type", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "order_id", NamedC: "order_id", Type: enumor.Numeric},
	{Column: "suborder_id", NamedC: "suborder_id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "day", NamedC: "day", Type: enumor.Numeric},
	{Column: "roll_date", NamedC: "roll_date", Type: enumor.Numeric},
	{Column: "applied_core", NamedC: "applied_core", Type: enumor.Numeric},
	{Column: "delivered_core", NamedC: "delivered_core", Type: enumor.Numeric},
	{Column: "exempted_returned_core", NamedC: "exempted_returned_core", Type: enumor.Numeric},
	{Column: "instance_group", NamedC: "instance_group", Type: enumor.String},
	{Column: "core_type", NamedC: "core_type", Type: enumor.String},
	{Column: "not_notice", NamedC: "not_notice", Type: enumor.Boolean},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// RollingAppliedRecord rolling applied record
type RollingAppliedRecord struct {
	// ID 自增ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// RequireType 需求类型
	RequireType enumor.RequireType `db:"require_type" json:"require_type"`
	// AppliedType 申请类型(枚举值：normal-普通申请、resource_pool-资源池申请、cvm_product-管理员cvm生产)
	AppliedType enumor.AppliedType `db:"applied_type" json:"applied_type" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// OrderID 主机申请的订单号
	OrderID uint64 `db:"order_id" json:"order_id"`
	// SubOrderID 主机申请的子订单号
	SubOrderID string `db:"suborder_id" json:"suborder_id" validate:"max=64"`
	// Year 申请时间年份
	Year int `db:"year" json:"year" validate:"max=9999"`
	// Month 申请时间月份
	Month int `db:"month" json:"month" validate:"max=12"`
	// Day 申请时间天
	Day int `db:"day" json:"day" validate:"max=31"`
	// RollDate 申请时间年月日
	RollDate int `db:"roll_date" json:"roll_date"`
	// AppliedCore cpu申请核心数
	AppliedCore *int64 `db:"applied_core" json:"applied_core"`
	// DeliveredCore cpu交付核心数
	DeliveredCore *int64 `db:"delivered_core" json:"delivered_core"`
	// ExemptedReturnedCore 减免退还核心数
	ExemptedReturnedCore *int64 `db:"exempted_returned_core" json:"exempted_returned_core"`
	// InstanceGroup 机型族
	InstanceGroup string `db:"instance_group" json:"instance_group"`
	// CoreType 机型核心类型
	CoreType enumor.CoreType `db:"core_type" json:"core_type"`
	// NotNotice 是否不通知
	NotNotice *bool `db:"not_notice" json:"not_notice"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName 表名
func (rar *RollingAppliedRecord) TableName() table.Name {
	return table.RollingAppliedRecordTable
}

// InsertValidate validate insert
func (rar *RollingAppliedRecord) InsertValidate() error {
	if len(rar.ID) == 0 {
		return errors.New("id is required")
	}
	if err := rar.RequireType.Validate(); err != nil {
		return err
	}
	if len(rar.AppliedType) == 0 {
		return errors.New("applied_type is required")
	}
	if rar.AppliedType.Validate() != nil {
		return rar.AppliedType.Validate()
	}
	if rar.BkBizID <= 0 {
		return errors.New("bk_biz_id is required")
	}
	if rar.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if len(rar.SubOrderID) == 0 {
		return errors.New("suborder_id is required")
	}
	if rar.Year <= 0 {
		return errors.New("year is required")
	}
	if rar.Month <= 0 {
		return errors.New("month is required")
	}
	if rar.Day <= 0 {
		return errors.New("day is required")
	}
	return validator.Validate.Struct(rar)
}

// UpdateValidate validate update
func (rar *RollingAppliedRecord) UpdateValidate() error {
	if len(rar.ID) == 0 {
		return errors.New("id is required")
	}
	return validator.Validate.Struct(rar)
}
