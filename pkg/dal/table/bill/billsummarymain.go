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

// AccountBillSummaryMainColumns defines account_bill_summary_main's columns.
var AccountBillSummaryMainColumns = utils.MergeColumns(nil, AccountBillSummaryMainColumnDescriptor)

// AccountBillSummaryMainColumnDescriptor is AwsBill's column descriptors.
var AccountBillSummaryMainColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "root_account_id", NamedC: "root_account_id", Type: enumor.String},
	{Column: "root_account_name", NamedC: "root_account_name", Type: enumor.String},
	{Column: "main_account_id", NamedC: "main_account_id", Type: enumor.String},
	{Column: "main_account_name", NamedC: "main_account_name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "product_name", NamedC: "product_name", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "last_synced_version", NamedC: "last_synced_version", Type: enumor.Numeric},
	{Column: "current_version", NamedC: "current_version", Type: enumor.Numeric},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "last_month_cost_synced", NamedC: "last_month_cost_synced", Type: enumor.Numeric},
	{Column: "last_month_rmb_cost_synced", NamedC: "last_month_rmb_cost_synced", Type: enumor.Numeric},
	{Column: "current_month_cost_synced", NamedC: "current_month_cost_synced", Type: enumor.Numeric},
	{Column: "current_month_rmb_cost_synced", NamedC: "current_month_rmb_cost_synced", Type: enumor.Numeric},
	{Column: "month_on_month_value", NamedC: "month_on_month_value", Type: enumor.Numeric},
	{Column: "current_month_cost", NamedC: "current_month_cost", Type: enumor.Numeric},
	{Column: "current_month_rmb_cost", NamedC: "current_month_rmb_cost", Type: enumor.Numeric},
	{Column: "rate", NamedC: "rate", Type: enumor.Numeric},
	{Column: "adjustment_cost", NamedC: "adjustment_cost", Type: enumor.Numeric},
	{Column: "adjustment_rmb_cost", NamedC: "adjustment_rmb_cost", Type: enumor.Numeric},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillSummaryMain account_bill_summary_main表，存储月度汇总账单
type AccountBillSummaryMain struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// RootAccountID 一级账号ID
	RootAccountID string `db:"root_account_id" json:"root_account_id"`
	// RootAccountName 一级账号名称
	RootAccountName string `db:"root_account_name" json:"root_account_name"`
	// MainAccountID 账号ID
	MainAccountID string `db:"main_account_id" json:"main_account_id"`
	// MainAccountName 一级账号名称
	MainAccountName string `db:"main_account_name" json:"main_account_name"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// ProductID 运营产品ID
	ProductID int64 `db:"product_id" json:"product_id"`
	// ProductName 运营产品名称
	ProductName string `db:"product_name" json:"product_name"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BkBizName 业务名称
	BkBizName string `db:"bk_biz_name" json:"bk_biz_name"`
	// BillYear 账单年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 账单月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// LastSyncedVersion 最后同步的账单版本
	LastSyncedVersion int `db:"last_synced_version" json:"last_synced_version"`
	// CurrentVersion 当前账单版本
	CurrentVersion int `db:"current_version" json:"current_version"`
	// Currency 币种
	Currency enumor.CurrencyCode `db:"currency" json:"currency"`
	// LastMonthCostSynced 上月已同步账单
	LastMonthCostSynced *types.Decimal `db:"last_month_cost_synced" json:"last_month_cost_synced"`
	// LastMonthRMBCostSynced 上月已同步人民币账单
	LastMonthRMBCostSynced *types.Decimal `db:"last_month_rmb_cost_synced" json:"last_month_rmb_cost_synced"`
	// CurrentMonthCostSynced 本月已同步账单
	CurrentMonthCostSynced *types.Decimal `db:"current_month_cost_synced" json:"current_month_cost_synced"`
	// CurrentMonthRMBCostSynced 本月已同步人民币账单
	CurrentMonthRMBCostSynced *types.Decimal `db:"current_month_rmb_cost_synced" json:"current_month_rmb_cost_synced"`
	// MonthOnMonthValue 本月已同步账单环比
	MonthOnMonthValue float64 `db:"month_on_month_value" json:"month_on_month_value"`
	// CurrentMonthCost 实时账单
	CurrentMonthCost *types.Decimal `db:"current_month_cost" json:"current_month_cost"`
	// CurrentMonthRMBCost 实时人民币账单
	CurrentMonthRMBCost *types.Decimal `db:"current_month_rmb_cost" json:"current_month_rmb_cost"`
	// Rate 汇率
	Rate float64 `db:"rate" json:"rate"`
	// AdjustmentCost 实时调账账单
	AdjustmentCost *types.Decimal `db:"adjustment_cost" json:"adjustment_cost"`
	// AdjustmentRMBCost 实时人民币调账账单
	AdjustmentRMBCost *types.Decimal `db:"adjustment_rmb_cost" json:"adjustment_rmb_cost"`
	// State 状态
	State enumor.MainBillSummaryState `db:"state" json:"state"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回月度汇总账单表名
func (abs *AccountBillSummaryMain) TableName() table.Name {
	return table.AccountBillSummaryMainTable
}

// InsertValidate validate account bill summary on insert
func (abs *AccountBillSummaryMain) InsertValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if len(abs.Vendor) == 0 {
		return errors.New("vendor is required")
	}
	if len(abs.RootAccountID) == 0 {
		return errors.New("root_account_id is required")
	}
	if len(abs.MainAccountID) == 0 {
		return errors.New("main_account_id is required")
	}
	if abs.BkBizID == 0 && abs.ProductID == 0 {
		return errors.New("bk_biz_id or product_id is required")
	}
	if abs.BillYear == 0 {
		return errors.New("bill_year is required")
	}
	if abs.BillMonth == 0 {
		return errors.New("bill_month is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill summary on update
func (abs *AccountBillSummaryMain) UpdateValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}
