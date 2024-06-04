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

// Package bill ...
package bill

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AccountBillPullerColumns defines account_bill_puller's columns
var AccountBillPullerColumns = utils.MergeColumns(nil, AccountBillPullerDescriptor)

// AccountBillPullerDescriptor is AccountBillPuller's column descriptors
var AccountBillPullerDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "first_account_id", NamedC: "first_account_id", Type: enumor.String},
	{Column: "second_account_id", NamedC: "second_account_id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "pull_mode", NamedC: "pull_mode", Type: enumor.String},
	{Column: "sync_period", NamedC: "sync_period", Type: enumor.String},
	{Column: "bill_delay", NamedC: "bill_delay", Type: enumor.String},
	{Column: "final_bill_calendar_date", NamedC: "final_bill_calendar_date", Type: enumor.Numeric},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillPuller account_bill_puller表，存储云账单拉取器状态
type AccountBillPuller struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// FirstAccountID 一级账号ID
	FirstAccountID string `db:"first_account_id" json:"first_account_id"`
	// SecondAccountID 账号ID
	SecondAccountID string `db:"second_account_id" json:"second_account_id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// ProductID 运营产品ID
	ProductID int64 `db:"product_id" json:"product_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// PullMode 账单同步模式
	PullMode string `db:"pull_mode" json:"pull_mode"`
	// SyncPeriod 账单同步周期
	SyncPeriod string `db:"sync_period" json:"sync_period"`
	// BillDelay 账单延迟查询时间 单位：小时
	BillDelay string `db:"bill_delay" json:"bill_delay"`
	// FinalBillCalendarDate 出账日期 单位：日 （1～31）
	FinalBillCalendarDate int `db:"final_bill_calendar_date" validate:"lte=31" json:"final_bill_calendar_date"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回账单拉起器状态表表名
func (abs *AccountBillPuller) TableName() table.Name {
	return table.AccountBillPullerTable
}

// InsertValidate validate account bill summary on insert
func (abs *AccountBillPuller) InsertValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if abs.BkBizID == 0 && abs.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill summary on update
func (abs *AccountBillPuller) UpdateValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.FirstAccountID) == 0 {
		return errors.New("first_account_id is required")
	}
	if len(abs.SecondAccountID) == 0 {
		return errors.New("second_account_id is required")
	}
	if abs.BkBizID == 0 && abs.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}
