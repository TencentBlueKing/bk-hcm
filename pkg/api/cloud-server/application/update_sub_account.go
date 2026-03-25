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
	"fmt"

	"hcm/pkg/criteria/validator"
)

// SubAccountBatchUpdateReq define sub account batch update request.
type SubAccountBatchUpdateReq struct {
	SubAccountBaseReq `json:",inline"`
	SubAccounts       []SubAccountUpdateReq `json:"sub_accounts" validate:"required,min=1,max=100"`
}

// Validate sub account batch update request.
func (req *SubAccountBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ValidateBase(); err != nil {
		return err
	}

	for i, item := range req.SubAccounts {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("sub_accounts[%d] validate failed, err: %w", i, err)
		}
	}

	return nil
}

// SubAccountUpdateReq define sub account update request for a single sub-account.
// Pointer fields use nil to indicate "no change"; non-nil means update to that value.
type SubAccountUpdateReq struct {
	ID          string   `json:"id" validate:"required"`
	BkBizID     *int64   `json:"bk_biz_id" validate:"omitempty,gt=0"`
	Email       *string  `json:"email" validate:"omitempty,email"`
	PhoneNum    *string  `json:"phone_num" validate:"omitempty"`
	CountryCode *string  `json:"country_code" validate:"omitempty"`
	Managers    []string `json:"managers" validate:"omitempty"`
	Memo        *string  `json:"memo" validate:"omitempty"`
}

// Validate sub account update request.
func (item *SubAccountUpdateReq) Validate() error {
	return validator.Validate.Struct(item)
}
