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

// AccountBillMonthTaskColumns defines account_bill_month_task's columns
var AccountBillMonthTaskColumns = utils.MergeColumns(nil, AccountBillMonthTaskDescriptor)

// AccountBillMonthTaskDescriptor is AccountBillMonthTask's column descriptors
var AccountBillMonthTaskDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "root_account_id", NamedC: "root_account_id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "version_id", NamedC: "version_id", Type: enumor.Numeric},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "count", NamedC: "count", Type: enumor.Numeric},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "pull_index", NamedC: "pull_index", Type: enumor.Numeric},
	{Column: "pull_flow_id", NamedC: "pull_flow_id", Type: enumor.String},
	{Column: "split_index", NamedC: "split_index", Type: enumor.Numeric},
	{Column: "split_flow_id", NamedC: "split_flow_id", Type: enumor.String},
	{Column: "summary_flow_id", NamedC: "summary_flow_id", Type: enumor.String},
	{Column: "summary_detail", NamedC: "summary_detail", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AccountBillMonthTask account_bill_month_task表，存储云账单拉取器状态
type AccountBillMonthTask struct {
	// ID 自增ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// RootAccountID 一级账号ID
	RootAccountID string `db:"root_account_id" json:"root_account_id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// BillYear 账单年份
	BillYear int `db:"bill_year" json:"bill_year"`
	// BillMonth 账单月份
	BillMonth int `db:"bill_month" json:"bill_month"`
	// VersionID 版本号
	VersionID int `db:"version_id" json:"version_id"`
	// State 状态
	State enumor.RootAccountMonthBillTaskState `db:"state" json:"state"`
	// Count 账单条目数量
	Count uint64 `db:"count" json:"count"`
	// Currency 币种
	Currency enumor.CurrencyCode `db:"currency" json:"currency"`
	// Cost 金额，单位：元
	Cost *types.Decimal `db:"cost" json:"cost"`
	// PullIndex pull index
	PullIndex uint64 `db:"pull_index" json:"pull_index"`
	// PullFlowID task id
	PullFlowID string `db:"pull_flow_id" json:"pull_flow_id"`
	// SplitIndex split index
	SplitIndex uint64 `db:"split_index" json:"split_index"`
	// SplitFlowID split flow id
	SplitFlowID string `db:"split_flow_id" json:"split_flow_id"`
	// SummaryFlowID summary flow id
	SummaryFlowID string `db:"summary_flow_id" json:"summary_flow_id"`
	// SummaryDetail detail of summary
	SummaryDetail string `db:"summary_detail" json:"summary_detail"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回账单拉起器状态表表名
func (abs *AccountBillMonthTask) TableName() table.Name {
	return table.AccountBillMonthTaskTable
}

// InsertValidate validate account bill month task on insert
func (abs *AccountBillMonthTask) InsertValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}

// UpdateValidate validate account bill month task on update
func (abs *AccountBillMonthTask) UpdateValidate() error {
	if len(abs.ID) == 0 {
		return errors.New("id is required")
	}
	if err := validator.Validate.Struct(abs); err != nil {
		return err
	}
	return nil
}
