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

package application

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// MainAccountAcommonCreateReq ...
type MainAccountCommonCreateReq struct {
	Vendor       enumor.Vendor                  `json:"vendor" validate:"required"`
	Email        string                         `json:"email" validate:"required"`
	Managers     []string                       `json:"managers" validate:"required,max=5"`
	BakManagers  []string                       `json:"bak_managers" validate:"required,max=5"`
	Site         enumor.MainAccountSiteType     `json:"site" validate:"required"`
	BusinessType enumor.MainAccountBusinessType `json:"business_type" validate:"required"`
	DeptID       int64                          `json:"dept_id" validate:"required"`
	BkBizID      int64                          `json:"bk_biz_id" validate:"required"`
	OpProductID  int64                          `json:"op_product_id" validate:"required"`
	Memo         *string                        `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *MainAccountCommonCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	if err := req.Site.Validate(); err != nil {
		return err
	}

	return nil
}

// MainAccountCreateReq ...
type MainAccountCreateReq struct {
	MainAccountCommonCreateReq `json:",inline"`
	// Extension 格式与AccountCreateReq.Extension，重新定义并覆盖
	Extension map[string]string `json:"extension" validate:"required"`
}

// Validate ...
func (req *MainAccountCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.MainAccountCommonCreateReq.Validate(); err != nil {
		return err
	}

	return nil
}

// MainAccountCompleteReq ...
type MainAccountCompleteReq struct {
	SN            string            `json:"sn" validate:"required,min=3,max=64"`
	ID            string            `json:"id" validate:"required,min=3,max=64"`
	Vendor        enumor.Vendor     `json:"vendor" validate:"required"`
	RootAccountID string            `json:"root_account_id" validate:"required"`
	Extension     map[string]string `json:"extension" validate:"required"`
}

// Validate ...
func (req *MainAccountCompleteReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	return nil
}

// RootAccountCommonAddReq ...
type MainAccountCommonUpdateReq struct {
	ID          string        `json:"id" validate:"required,min=3,max=64"`
	Vendor      enumor.Vendor `json:"vendor" validate:"required"`
	Managers    []string      `json:"managers" validate:"required,max=5"`
	BakManagers []string      `json:"bak_managers" validate:"required,max=5"`
	DeptID      int64         `json:"dept_id" validate:"omitempty"`
	OpProductID int64         `json:"op_product_id" validate:"omitempty"`
	BkBizID     int64         `json:"bk_biz_id" validate:"omitempty"`
}

// Validate ...
func (req *MainAccountCommonUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	return nil
}

// MainAccountUpdateReq ...
type MainAccountUpdateReq struct {
	MainAccountCommonUpdateReq `json:",inline"`
}

// Validate ...
func (req *MainAccountUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.MainAccountCommonUpdateReq.Validate(); err != nil {
		return err
	}

	return nil
}
