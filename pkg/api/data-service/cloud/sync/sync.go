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

package dssync

import (
	coresync "hcm/pkg/api/core/cloud/sync"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
)

// CreateReq define create account sync detail request.
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

// CreateField define account sync detail create field.
type CreateField struct {
	Vendor          enumor.Vendor   `json:"vendor" validate:"required"`
	AccountID       string          `json:"account_id" validate:"required"`
	ResName         string          `json:"res_name" validate:"required"`
	ResStatus       string          `json:"res_status" validate:"required"`
	ResEndTime      string          `json:"res_end_time" validate:"required"`
	ResFailedReason types.JsonField `json:"res_failed_reason" validate:"omitempty"`
}

// Validate CreateField.
func (req CreateField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateReq define update account sync detail request.
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

// UpdateField define account sync detail update field.
type UpdateField struct {
	ID              string          `json:"id" validate:"required"`
	ResStatus       string          `json:"res_status" validate:"required"`
	ResEndTime      string          `json:"res_end_time" validate:"required"`
	ResFailedReason types.JsonField `json:"res_failed_reason" validate:"omitempty"`
}

// Validate UpdateField.
func (req UpdateField) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListResult defines list result.
type ListResult struct {
	Count   uint64                            `json:"count"`
	Details []coresync.AccountSyncDetailTable `json:"details"`
}
