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

// Package tenant ...
package tenant

import (
	"fmt"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/tenant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// -------------------------- Create --------------------------

// CreateTenantReq define create tenant request.
type CreateTenantReq struct {
	Items []CreateTenantField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate CreateTenantReq.
func (req CreateTenantReq) Validate() error {
	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// CreateTenantField define tenant create field.
type CreateTenantField struct {
	TenantID string              `json:"tenant_id" validate:"required"`
	Status   enumor.TenantStatus `json:"status"`
}

// Validate CreateTenantField.
func (req CreateTenantField) Validate() error {
	if req.Status != enumor.TenantEnable && req.Status != enumor.TenantDisable {
		logs.Errorf("status must be 'enable' or 'disable', tenant_id: %s, status: %s", req.TenantID, req.Status)
		return fmt.Errorf("status must be 'enable' or 'disable'")
	}
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// UpdateTenantReq define update tenant request.
type UpdateTenantReq struct {
	Items []UpdateTenantField `json:"items" validate:"required,min=1,dive,required"`
}

// Validate UpdateTenantReq.
func (req UpdateTenantReq) Validate() error {
	for _, item := range req.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(req)
}

// UpdateTenantField define tenant update field.
type UpdateTenantField struct {
	ID string `json:"id" validate:"required"`

	TenantID string              `json:"tenant_id"`
	Status   enumor.TenantStatus `json:"status"`
}

// Validate UpdateTenantField.
func (req UpdateTenantField) Validate() error {
	if req.Status != enumor.TenantEnable && req.Status != enumor.TenantDisable {
		logs.Errorf("status must be 'enable' or 'disable', id: %s, status: %s", req.ID, req.Status)
		return fmt.Errorf("status must be 'enable' or 'disable'")
	}
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ListTenantResult defines list result.
type ListTenantResult = core.ListResultT[tenant.Tenant]
