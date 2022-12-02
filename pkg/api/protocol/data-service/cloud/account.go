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
// cloud 包提供各类云资源的请求与返回序列化器
package cloud

import (
	"encoding/json"

	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/runtime/filter"
)

type CreateAccountReq struct {
	Name         string                 `json:"name" validate:"required"`
	Vendor       string                 `json:"vendor" validate:"required"`
	DepartmentID int                    `json:"department_id" validate:"required,gt=0"`
	Managers     []string               `json:"managers" validate:"required,gt=0,dive,required"`
	Extension    map[string]interface{} `json:"extension" validate:"required"`
}

// DepartmentID int `db:"department_id" json:"department_id"`
// // 账号类型(资源账号|登记账号)
// Type string `db:"type" json:"type"`
// // 账号资源同步状态
// SyncStatus string `db:"sync_status" json:"sync_status"`
// // 账号余额数值
// Price string `db:"price" json:"price"`
// // 账号余额单位
// PriceUnit string `db:"price_unit" json:"price_unit"`
// // 云厂商账号差异扩展字段
// Extension table.JsonField `db:"extension" json:"extension" unmarshal_type:"map"`
// // 创建者
// Creator string `db:"creator" json:"creator"`
// // 更新者
// Reviser string `db:"reviser" json:"reviser"`
// // 创建时间
// CreatedAt *time.Time `db:"created_at" json:"created_at"`
// // 更新时间
// UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
// // 账号信息备注
// Memo string `db:"memo" json:"memo"`

func (c *CreateAccountReq) ToModel() *tablecloud.AccountModel {
	managers, _ := json.Marshal(c.Managers)
	ext, _ := json.Marshal(c.Extension)
	return &tablecloud.AccountModel{
		Name:         c.Name,
		Vendor:       c.Vendor,
		DepartmentID: c.DepartmentID,
		Managers:     table.JsonField(managers),
		Extension:    table.JsonField(ext),
		ModelManager: &table.ModelManager{},
	}
}

type UpdateAccountReq struct {
	Managers   []string               `json:"managers" validate:"required,gt=0,dive,required"`
	Extension  map[string]interface{} `json:"extension" validate:"required"`
	FilterExpr filter.Expression      `json:"filter_expr" validate:"required"`
}

// ToModel ...
func (u *UpdateAccountReq) ToModel() *tablecloud.AccountModel {
	managers, _ := json.Marshal(u.Managers)
	ext, _ := json.Marshal(u.Extension)
	return &tablecloud.AccountModel{
		Managers:     table.JsonField(managers),
		Extension:    table.JsonField(ext),
		ModelManager: &table.ModelManager{UpdateFields: []string{"managers", "extension"}},
	}
}

type AccountResp struct {
	Name     string   `json:"name"`
	Vendor   string   `json:"vendor"`
	Managers []string `json:"managers"`
}
