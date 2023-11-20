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

// Package cloud 描述云资源表
package cloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

/*
account.go:
- 账号表结构定义：AccountTable
- 账号字段描述：AccountColumnDescriptor
- 账号表插入校验
- 账号表更新校验
- `hcm/pkg/api/core/account` 转账号表结构函数
- 账号表结构转 `hcm/pkg/api/core/account` 函数
*/

// AccountColumns defines all the account table's columns.
var AccountColumns = utils.MergeColumns(nil, AccountColumnDescriptor)

// AccountColumnDescriptor is AccountID's column descriptors.
var AccountColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "managers", NamedC: "managers", Type: enumor.Json},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "site", NamedC: "site", Type: enumor.String},
	{Column: "price", NamedC: "price", Type: enumor.String},
	{Column: "price_unit", NamedC: "price_unit", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "recycle_reserve_time", NamedC: "recycle_reserve_time", Type: enumor.Numeric},
}

// AccountTable 云账号表
type AccountTable struct {
	// ID 账号 ID
	ID string `db:"id" json:"id"`
	// Name 账号名称
	Name string `db:"name" json:"name"`
	// Vendor 云厂商
	Vendor string `db:"vendor" json:"vendor"`
	// Managers 责任人
	Managers types.StringArray `db:"managers" json:"managers"`
	// Type 账号类型(资源账号|登记账号)
	Type string `db:"type" json:"type"`
	// Site 站点(中国站｜国际站)
	Site string `db:"site"`
	// Price 账号余额数值
	Price string `db:"price" json:"price"`
	// PriceUnit 账号余额单位
	PriceUnit string `db:"price_unit" json:"price_unit"`
	// Extension 云厂商账号差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
	// Memo 账号信息备注
	Memo *string `db:"memo" json:"memo"`
	// RecycleReserveTime 回收站保留时长，单位: 小时
	RecycleReserveTime int `db:"recycle_reserve_time" json:"recycle_reserve_time"`
}

// TableName return account table name.
func (a AccountTable) TableName() table.Name {
	return table.AccountTable
}

// InsertValidate validate account table on insert.
func (a AccountTable) InsertValidate() error {
	if len(a.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(a.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	// TODO: 添加账号其他信息正则和长度校验。

	return nil
}

// UpdateValidate validate account table on update.
func (a AccountTable) UpdateValidate() error {
	if len(a.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	// TODO: 添加账号无法更新字段的校验和可以更新字段的正则及长度校验。

	return nil
}
