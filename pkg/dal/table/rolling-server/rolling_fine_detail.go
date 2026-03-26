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

// Package rollingserver ...
package rollingserver

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"

	"github.com/shopspring/decimal"
)

// RollingFineDetailColumns defines all the rolling_fine_detail table's columns.
var RollingFineDetailColumns = utils.MergeColumns(nil, RollingFineDetailColumnDescriptor)

// RollingFineDetailColumnDescriptor is RollingFineDetailTable's column descriptors.
var RollingFineDetailColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "applied_record_id", NamedC: "applied_record_id", Type: enumor.String},
	{Column: "order_id", NamedC: "order_id", Type: enumor.Numeric},
	{Column: "suborder_id", NamedC: "suborder_id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "day", NamedC: "day", Type: enumor.Numeric},
	{Column: "roll_date", NamedC: "roll_date", Type: enumor.Numeric},
	{Column: "delivered_core", NamedC: "delivered_core", Type: enumor.Numeric},
	{Column: "returned_core", NamedC: "returned_core", Type: enumor.Numeric},
	{Column: "exempted_returned_core", NamedC: "exempted_returned_core", Type: enumor.Numeric},
	{Column: "fine", NamedC: "fine", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// RollingFineDetailTable is used to save rolling_fine_detail table.
type RollingFineDetailTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// AppliedRecordID 滚服申请记录信息的唯一标识
	AppliedRecordID string `db:"applied_record_id" json:"applied_record_id"`
	// OrderID 订单号
	OrderID uint64 `db:"order_id" json:"order_id"`
	// SuborderID 子订单号
	SubOrderID string `db:"suborder_id" json:"suborder_id"`
	// Year 子单号记录罚金的年份
	Year int `db:"year" json:"year"`
	// Month 子单号记录罚金的月份
	Month int `db:"month" json:"month"`
	// Day 子单号记录罚金的天
	Day int `db:"day" json:"day"`
	// RollDate 子单号记录罚金的年月日
	RollDate int `db:"roll_date" json:"roll_date"`
	// DeliveredCore 已交付核心数
	DeliveredCore uint64 `db:"delivered_core" json:"delivered_core"`
	// ReturnedCore 已退还核心数
	ReturnedCore uint64 `db:"returned_core" json:"returned_core"`
	// ExemptedReturnedCore 减免退还核心数
	ExemptedReturnedCore uint64 `db:"exempted_returned_core" json:"exempted_returned_core"`
	// Fine 超时退还罚金
	Fine decimal.Decimal `db:"fine" json:"fine"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
}

// TableName is the rolling_fine_detail table's name.
func (r RollingFineDetailTable) TableName() table.Name {
	return table.RollingFineDetail
}

// InsertValidate validate rolling_fine_detail on insert.
func (r RollingFineDetailTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.BkBizID <= 0 {
		return errors.New("bk_biz_id should be > 0")
	}

	if len(r.AppliedRecordID) == 0 {
		return errors.New("applied_record_id can not be empty")
	}

	if r.OrderID == 0 {
		return errors.New("order_id can not be empty")
	}

	if len(r.SubOrderID) == 0 {
		return errors.New("suborder_id can not be empty")
	}

	if r.Year <= 0 {
		return errors.New("year is required")
	}

	if r.Month <= 0 {
		return errors.New("month is required")
	}
	if r.Month > 12 {
		return errors.New("month should be <= 12")
	}

	if r.Day <= 0 {
		return errors.New("day is required")
	}

	if r.Day > 31 {
		return errors.New("day should be <= 31")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}
