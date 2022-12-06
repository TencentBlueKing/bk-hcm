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

//
//import (
//	"time"
//
//	"hcm/pkg/api/protocol/data-service/validator"
//	"hcm/pkg/dal/dao/types"
//	"hcm/pkg/dal/table"
//	tablecloud "hcm/pkg/dal/table/cloud"
//	"hcm/pkg/runtime/filter"
//)
//
//type CreateAccountBizRelReq struct {
//	BkBizID   uint64 `json:"bk_biz_id" validate:"required"`
//	AccountID uint64 `json:"account_id" validate:"required"`
//}
//
//func (c *CreateAccountBizRelReq) Validate() error {
//	return validator.Validate.Struct(c)
//}
//
//func (c *CreateAccountBizRelReq) ToModel(creator string) *tablecloud.AccountBizRelModel {
//	return &tablecloud.AccountBizRelModel{
//		BkBizID:      c.BkBizID,
//		AccountID:    c.AccountID,
//		Creator:      creator,
//		Reviser:      creator,
//		ModelManager: &table.ModelManager{},
//	}
//}
//
//type UpdateAccountBizRelsReq struct {
//	BkBizID    uint64            `json:"bk_biz_id" validate:"required"`
//	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
//}
//
//func (u *UpdateAccountBizRelsReq) Validate() error {
//	return validator.Validate.Struct(u)
//}
//
//func (u *UpdateAccountBizRelsReq) ToModel(reviser string) *tablecloud.AccountBizRelModel {
//	return &tablecloud.AccountBizRelModel{
//		BkBizID:      u.BkBizID,
//		Reviser:      reviser,
//		ModelManager: &table.ModelManager{UpdateFields: validator.ExtractValidFields(u)},
//	}
//}
//
//type ListAccountBizRelsReq struct {
//	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
//}
//
//func (l *ListAccountBizRelsReq) Validate() error {
//	return validator.Validate.Struct(l)
//}
//
//func (l *ListAccountBizRelsReq) ToListOption() *types.ListOption {
//	return &types.ListOption{
//		FilterExpr: &l.FilterExpr,
//		Fields:     table.ListModelFields(new(AccountBizRelData)),
//	}
//}
//
//type AccountBizRelData struct {
//	ID        uint64     `json:"id" db:"id"`
//	BkBizID   uint64     `json:"bk_biz_id" db:"bk_biz_id"`
//	Creator   string     `json:"creator" db:"creator"`
//	Reviser   string     `json:"reviser" db:"reviser"`
//	CreatedAt *time.Time `json:"created_at" db:"created_at"`
//	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
//}
//
//// NewAccountBizRelData ...
//func NewAccountBizRelData(m *tablecloud.AccountBizRelModel) *AccountBizRelData {
//	// TODO 反射机制让创建过程更加"动态"?
//	return &AccountBizRelData{
//		ID:        m.ID,
//		BkBizID:   m.BkBizID,
//		Creator:   m.Creator,
//		Reviser:   m.Reviser,
//		CreatedAt: m.CreatedAt,
//		UpdatedAt: m.UpdatedAt,
//	}
//}
//
//// ListAccountBizRelsResult ...
//type ListAccountBizRelsResult struct {
//	Details []AccountBizRelData `json:"details"`
//}
//
//type DeleteAccountBizRelsReq struct {
//	FilterExpr filter.Expression `json:"filter_expr" validate:"required"`
//}
//
//func (d *DeleteAccountBizRelsReq) Validate() error {
//	return validator.Validate.Struct(d)
//}
