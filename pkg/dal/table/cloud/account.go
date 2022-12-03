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

// cloud 包描述云资源 Model
package cloud

import (
	"time"

	"hcm/pkg/dal/table"
	"hcm/pkg/runtime/filter"
)

type AccountModel struct {
	// 账号自增 ID
	ID uint64 `db:"id"`
	// 账号名称
	Name string `db:"name"`
	// 云厂商
	Vendor string `db:"vendor"`
	// 责任人
	Managers table.JsonField `db:"managers" unmarshal_type:"stringslice"`
	// 部门 ID
	DepartmentID int `db:"department_id"`
	// 账号类型(资源账号|登记账号)
	Type string `db:"type"`
	// 账号资源同步状态
	SyncStatus string `db:"sync_status"`
	// 账号余额数值
	Price string `db:"price"`
	// 账号余额单位
	PriceUnit string `db:"price_unit"`
	// 云厂商账号差异扩展字段
	Extension table.JsonField `db:"extension" unmarshal_type:"map"`
	// 创建者
	Creator string `db:"creator"`
	// 更新者
	Reviser string `db:"reviser"`
	// 创建时间
	CreatedAt *time.Time `db:"created_at"`
	// 更新时间
	UpdatedAt *time.Time `db:"updated_at"`
	// 账号信息备注
	Memo string `db:"memo"`
	// model manager
	ModelManager *table.ModelManager
}

var _ table.Model = new(AccountModel)

// TableName is the account's database table name.
func (a *AccountModel) TableName() string {
	return "account"
}

// GenerateInsertSQL ...
func (a *AccountModel) GenerateInsertSQL() string {
	return a.ModelManager.GenerateInsertSQL(a)
}

// GenerateInsertSQL ...
func (a *AccountModel) GenerateUpdateSQL(expr *filter.Expression) (string, error) {
	return a.ModelManager.GenerateUpdateSQL(a, expr)
}

func (a *AccountModel) GenerateUpdateFieldKV() map[string]interface{} {
	return a.ModelManager.GenerateUpdateFieldKV(a)
}

// // ColumnNamesForInsert 生成 insert sql 中的 (column1, column2, column3, ...)
// func (a *AccountModel) ColumnNamesForInsert() string {
// 	return a.ModelManager.ColumnNamesForInsert(a)
// }
//
// // ColumnValuesForInsert 生成 insert sql 中的 (value1, value2, value3, ...)
// func (a *AccountModel) ColumnValuesForInsert() string {
// 	return a.ModelManager.ColumnValuesForInsert(a)
// }
