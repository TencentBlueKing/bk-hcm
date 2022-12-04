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
	"time"

	"hcm/pkg/api/protocol/data-service/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/runtime/filter"
)

type CreateAccountReq struct {
	Name         string                 `json:"name" validate:"required"`
	Vendor       string                 `json:"vendor" validate:"required"`
	DepartmentID int                    `json:"department_id" validate:"required,gt=0"`
	Type         string                 `json:"type" validate:"required"`
	Managers     []string               `json:"managers" validate:"required,gt=0,dive,required"`
	Extension    map[string]interface{} `json:"extension" validate:"required"`
	Memo         string                 `json:"memo"`
}

func (c *CreateAccountReq) Validate() error {
	return validator.Validate.Struct(c)
}

func (c *CreateAccountReq) ToModel(creator string) *tablecloud.AccountModel {
	managers, _ := json.Marshal(c.Managers)
	ext, _ := json.Marshal(c.Extension)
	// TODO 反射机制让创建过程更加"动态"?
	return &tablecloud.AccountModel{
		Name:         c.Name,
		Vendor:       c.Vendor,
		DepartmentID: c.DepartmentID,
		Managers:     table.JsonField(managers),
		Extension:    table.JsonField(ext),
		Creator:      creator,
		Reviser:      creator,
		SyncStatus:   "", // 账号初始状态设置
		ModelManager: &table.ModelManager{},
	}
}

// TODO 增加值有效时进行校验的逻辑
// int 和 string 等基础类型, 可通过指针方式表示是否传递
type UpdateAccountsReq struct {
	Name         string                 `json:"name"`
	Managers     []string               `json:"managers"`
	Price        string                 `json:"price"`
	PriceUnit    string                 `json:"price_unit"`
	DepartmentID int                    `json:"department_id"`
	Extension    map[string]interface{} `json:"extension"`
	Memo         *string                `json:"memo"`
	FilterExpr   filter.Expression      `json:"filter_expr" validate:"required"`
}

func (u *UpdateAccountsReq) Validate() error {
	return validator.Validate.Struct(u)
}

// ToModel ...
func (u *UpdateAccountsReq) ToModel(reviser string) *tablecloud.AccountModel {
	managers, _ := json.Marshal(u.Managers)
	ext, _ := json.Marshal(u.Extension)

	var memo string
	if u.Memo != nil {
		memo = *u.Memo
	}

	// TODO 反射机制让创建过程更加"动态"?
	return &tablecloud.AccountModel{
		Name:         u.Name,
		Managers:     table.JsonField(managers),
		Price:        u.Price,
		PriceUnit:    u.PriceUnit,
		DepartmentID: u.DepartmentID,
		Extension:    table.JsonField(ext),
		Memo:         memo,
		Reviser:      reviser,
		ModelManager: &table.ModelManager{UpdateFields: validator.ExtractValidFields(u)},
	}
}

type ListAccountsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

func (l *ListAccountsReq) Validate() error {
	return validator.Validate.Struct(l)
}

func (l *ListAccountsReq) ToListOption() *types.ListOption {
	return &types.ListOption{
		FilterExpr: &l.FilterExpr,
		Fields:     table.ListModelFields(new(AccountData)),
	}
}

type AccountData struct {
	ID        uint64                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Vendor    string                 `json:"vendor" db:"vendor"`
	Managers  []string               `json:"managers" db:"managers"`
	Price     string                 `json:"price" db:"price"`
	PriceUnit string                 `json:"price_unit" db:"price_unit"`
	Extension map[string]interface{} `json:"extension" db:"extension"`
	Creator   string                 `json:"creator" db:"creator"`
	Reviser   string                 `json:"reviser" db:"reviser"`
	CreatedAt *time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time             `json:"updated_at" db:"updated_at"`
}

// NewAccountData ...
func NewAccountData(m *tablecloud.AccountModel) *AccountData {
	managers := make([]string, 0)
	json.Unmarshal([]byte(m.Managers), &managers)

	ext := make(map[string]interface{}, 0)
	json.Unmarshal([]byte(m.Extension), &ext)

	// TODO 反射机制让创建过程更加"动态"?
	return &AccountData{
		ID:        m.ID,
		Name:      m.Name,
		Vendor:    m.Vendor,
		Managers:  managers,
		Price:     m.Price,
		PriceUnit: m.PriceUnit,
		Extension: ext,
		Creator:   m.Creator,
		Reviser:   m.Reviser,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ListAccountsResult defines list instances for iam pull resource callback result.
type ListAccountsResult struct {
	Details []AccountData `json:"details"`
}

type DeleteAccountsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

func (d *DeleteAccountsReq) Validate() error {
	return validator.Validate.Struct(d)
}
