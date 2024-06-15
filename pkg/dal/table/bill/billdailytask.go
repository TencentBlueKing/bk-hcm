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
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AccountBillDailyPullTaskColumns defines account_bill_daily_pull_task's columns
var AccountBillDailyPullTaskColumns = utils.MergeColumns(nil, AccountBillDailyPullTaskDescriptor)

// AccountBillDailyPullTaskDescriptor is AccountBillDailyPullTask's column descriptors
var AccountBillDailyPullTaskDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "root_account_id", NamedC: "root_account_id", Type: enumor.String},
	{Column: "main_account_id", NamedC: "main_account_id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "bill_day", NamedC: "bill_day", Type: enumor.Numeric},
	{Column: "version_id", NamedC: "version_id", Type: enumor.Numeric},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "count", NamedC: "count", Type: enumor.Numeric},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "flow_id", NamedC: "flow_id", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillDailyPullTask account_bill_daily_pull_task表，存储每日账单拉取状态
type AccountBillDailyPullTask struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// RootAccountID 一级账号ID
	RootAccountID string `db:"root_account_id" json:"root_account_id"`
	// MainAccountID 账号ID
	MainAccountID string `db:"main_account_id" json:"main_account_id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// ProductID 运营产品ID
	ProductID int64 `db:"product_id" json:"product_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BillYear 账单年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 账单月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// BillMonth 账单月份 YYYY-MM
	BillDay int `db:"bill_day" json:"bill_day"`
	// VersionID 版本号
	VersionID int `db:"version_id" json:"version_id"`
	// State 状态
	State string `db:"state" json:"state"`
	// Count 账单条目数量
	Count int64 `db:"count" json:"count"`
	// Currency 币种
	Currency string `db:"currency" json:"currency"`
	// Cost 金额，单位：元
	Cost *types.Decimal `db:"cost" json:"cost"`
	// FlowID task id
	FlowID string `db:"flow_id" json:"flow_id"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回表名
func (abdpt *AccountBillDailyPullTask) TableName() table.Name {
	return table.AccountBillDailyPullTaskTable
}

// InsertValidate validate account bill summary on insert
func (abdpt *AccountBillDailyPullTask) InsertValidate() error {
	if len(abdpt.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abdpt.RootAccountID) == 0 {
		return errors.New("root_account_id is required")
	}
	if len(abdpt.MainAccountID) == 0 {
		return errors.New("main_account_id is required")
	}
	if abdpt.BkBizID == 0 && abdpt.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if abdpt.BillYear == 0 {
		return errors.New("bill_year is required")
	}
	if abdpt.BillMonth == 0 {
		return errors.New("bill_month is required")
	}
	if abdpt.BillDay == 0 {
		return errors.New("bill_day is required")
	}
	if abdpt.VersionID < 0 {
		return fmt.Errorf("version_id %d is invalid", abdpt.VersionID)
	}
	if err := validator.Validate.Struct(abdpt); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill summary on update
func (abdpt *AccountBillDailyPullTask) UpdateValidate() error {
	if len(abdpt.ID) == 0 {
		return errors.New("id is required")
	}
	if err := validator.Validate.Struct(abdpt); err != nil {
		return err
	}
	return nil
}
