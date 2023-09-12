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

package dssubaccount

import (
	"hcm/pkg/api/core"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
)

// -------------------------- Create --------------------------

// CreateReq define create sub account request.
type CreateReq struct {
	Items []CreateField `json:"items" validate:"required,min=1"`
}

// Validate CreateReq.
func (req CreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CreateField define sub account create field.
type CreateField struct {
	CloudID     string                 `json:"cloud_id" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Vendor      enumor.Vendor          `json:"vendor" validate:"required"`
	Site        enumor.AccountSiteType `json:"site" validate:"required"`
	AccountID   string                 `json:"account_id" validate:"required"`
	AccountType string                 `json:"account_type" validate:"omitempty"`
	Extension   core.ExtMessage        `json:"extension" validate:"required"`
	Managers    types.StringArray      `json:"managers" validate:"omitempty"`
	BkBizIDs    types.Int64Array       `json:"bk_biz_ids" validate:"omitempty"`
	Memo        *string                `json:"memo" validate:"omitempty"`
}

// Validate CreateField.
func (req CreateField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateReq define update sub account request.
type UpdateReq struct {
	Items []UpdateField `json:"items" validate:"required,min=1"`
}

// Validate UpdateReq.
func (req UpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// UpdateField define sub account update field.
type UpdateField struct {
	ID string `json:"id" validate:"required"`

	Name        string                 `json:"name" validate:"omitempty"`
	Vendor      enumor.Vendor          `json:"vendor" validate:"omitempty"`
	Site        enumor.AccountSiteType `json:"site" validate:"omitempty"`
	AccountID   string                 `json:"account_id" validate:"omitempty"`
	AccountType string                 `json:"account_type" validate:"omitempty"`
	Managers    types.StringArray      `json:"managers" validate:"omitempty"`
	BkBizIDs    types.Int64Array       `json:"bk_biz_ids" validate:"omitempty"`
	Extension   core.ExtMessage        `json:"extension" validate:"omitempty"`
	Memo        *string                `json:"memo" validate:"omitempty"`
}

// Validate UpdateField.
func (req UpdateField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListResult defines list result.
type ListResult struct {
	Count uint64 `json:"count"`
	// 对于List接口，只会返回公共数据，不会返回Extension
	Details []coresubaccount.BaseSubAccount `json:"details"`
}

// ListExtResult define list extension result.
type ListExtResult[T coresubaccount.Extension] struct {
	Count   uint64                         `json:"count,omitempty"`
	Details []coresubaccount.SubAccount[T] `json:"details,omitempty"`
}
