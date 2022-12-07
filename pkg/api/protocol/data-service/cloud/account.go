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

// Package cloud 包提供各类云资源的请求与返回序列化器
package cloud

import (
	"time"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/models/cloud"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateAccountReq ...
type CreateAccountReq struct {
	Name         string                 `json:"name" validate:"required"`
	Vendor       string                 `json:"vendor" validate:"required"`
	DepartmentID int                    `json:"department_id" validate:"required,gt=0"`
	Type         string                 `json:"type" validate:"required"`
	Managers     []string               `json:"managers" validate:"required,gt=0,dive,required"`
	Extension    map[string]interface{} `json:"extension" validate:"required"`
	Memo         string                 `json:"memo"`
}

// Validate ...
func (c *CreateAccountReq) Validate() error {
	return validator.Validate.Struct(c)
}

// UpdateAccountsReq ...
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

// Validate ...
func (u *UpdateAccountsReq) Validate() error {
	return validator.Validate.Struct(u)
}

// ListAccountsReq ...
type ListAccountsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

// Validate ...
func (l *ListAccountsReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ToListOption ...
func (l *ListAccountsReq) ToListOption() *types.ListOption {
	return &types.ListOption{
		FilterExpr: &l.FilterExpr,
		Fields:     table.ListTableFields(new(tablecloud.AccountTable)),
	}
}

// AccountResp ...
type AccountResp struct {
	ID        uint64      `json:"id"`
	Name      string      `json:"name"`
	Vendor    string      `json:"vendor"`
	Managers  []string    `json:"managers"`
	Price     string      `json:"price"`
	PriceUnit string      `json:"price_unit"`
	Extension interface{} `json:"extension"`
	Creator   string      `json:"creator"`
	Reviser   string      `json:"reviser"`
	CreatedAt *time.Time  `json:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at"`
	Memo      string      `json:"memo"`
}

// NewAccountResp ...
func NewAccountResp(m *cloud.Account) *AccountResp {
	return &AccountResp{
		ID:        m.ID,
		Name:      m.Name,
		Vendor:    m.Vendor,
		Managers:  m.Managers,
		Price:     m.Price,
		PriceUnit: m.PriceUnit,
		Extension: m.Extension,
		Creator:   m.Creator,
		Reviser:   m.Reviser,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Memo:      m.Memo,
	}
}

// ListAccountsResult defines list instances for iam pull resource callback result.
type ListAccountsResult struct {
	Details []AccountResp `json:"details"`
}

// DeleteAccountsReq ...
type DeleteAccountsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

// Validate ...
func (d *DeleteAccountsReq) Validate() error {
	return validator.Validate.Struct(d)
}

// ListAccountsResp ...
type ListAccountsResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListAccountsResult `json:"data"`
}
