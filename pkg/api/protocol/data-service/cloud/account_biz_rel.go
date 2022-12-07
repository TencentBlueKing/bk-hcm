///*
// * TencentBlueKing is pleased to support the open source community by making
// * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
// * Copyright (C) 2022 THL A29 Limited,
// * a Tencent company. All rights reserved.
// * Licensed under the MIT License (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at http://opensource.org/licenses/MIT
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on
// * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// * either express or implied. See the License for the
// * specific language governing permissions and limitations under the License.
// *
// * We undertake not to change the open source license (MIT license) applicable
// *
// * to the current version of the project delivered to anyone in the future.
// */

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

// CreateAccountBizRelReq ...
type CreateAccountBizRelReq struct {
	BkBizID   int    `json:"bk_biz_id" validate:"required"`
	AccountID uint64 `json:"account_id" validate:"required"`
}

// Validate ...
func (c *CreateAccountBizRelReq) Validate() error {
	return validator.Validate.Struct(c)
}

// UpdateAccountBizRelsReq ...
type UpdateAccountBizRelsReq struct {
	BkBizID    int               `json:"bk_biz_id" validate:"required"`
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

// Validate ...
func (u *UpdateAccountBizRelsReq) Validate() error {
	return validator.Validate.Struct(u)
}

// ListAccountBizRelsReq ...
type ListAccountBizRelsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

// Validate ...
func (l *ListAccountBizRelsReq) Validate() error {
	return validator.Validate.Struct(l)
}

// ToListOption ...
func (l *ListAccountBizRelsReq) ToListOption() *types.ListOption {
	return &types.ListOption{
		FilterExpr: &l.FilterExpr,
		Fields:     table.ListTableFields(new(tablecloud.AccountBizRelTable)),
	}
}

// AccountBizRelResp ...
type AccountBizRelResp struct {
	ID        uint64     `json:"id" db:"id"`
	BkBizID   int        `json:"bk_biz_id" db:"bk_biz_id"`
	Creator   string     `json:"creator" db:"creator"`
	Reviser   string     `json:"reviser" db:"reviser"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

// NewAccountBizRelResp ...
func NewAccountBizRelResp(m *cloud.AccountBizRel) *AccountBizRelResp {
	// TODO 反射机制让创建过程更加"动态"?
	return &AccountBizRelResp{
		ID:        m.ID,
		BkBizID:   m.BkBizID,
		Creator:   m.Creator,
		Reviser:   m.Reviser,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ListAccountBizRelsResult ...
type ListAccountBizRelsResult struct {
	Details []AccountBizRelResp `json:"details"`
}

// DeleteAccountBizRelsReq ...
type DeleteAccountBizRelsReq struct {
	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
}

// Validate ...
func (d *DeleteAccountBizRelsReq) Validate() error {
	return validator.Validate.Struct(d)
}

// ListAccountBizRelsResp ...
type ListAccountBizRelsResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListAccountBizRelsResult `json:"data"`
}
