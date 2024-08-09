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

package bill

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// RootAccountBillConfigBatchCreateReq defines batch create account bill config request.
type RootAccountBillConfigBatchCreateReq[T RootAccountBillConfigExtension] struct {
	Bills []RootAccountBillConfigReq[T] `json:"bills" validate:"required,max=100"`
}

// RootAccountBillConfigReq defines create account bill config request.
type RootAccountBillConfigReq[T RootAccountBillConfigExtension] struct {
	Vendor            enumor.Vendor `json:"vendor" validate:"required"`
	RootAccountID     string        `json:"root_account_id" validate:"required"`
	CloudDatabaseName string        `json:"cloud_database_name" validate:"omitempty"`
	CloudTableName    string        `json:"cloud_table_name" validate:"omitempty"`
	Status            int64         `json:"status" validate:"omitempty"`
	ErrMsg            []string      `json:"err_msg" validate:"omitempty"`
	Extension         *T            `json:"extension" validate:"omitempty"`
}

// RootAccountBillConfigExtension defines create account bill config extensional info.
type RootAccountBillConfigExtension interface {
	bill.AwsBillConfigExtension | bill.GcpBillConfigExtension
}

// Validate RootAccountBillConfigBatchCreateReq.
func (c *RootAccountBillConfigBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// RootAccountBillConfigBatchUpdateReq define batch update account bill config request.
type RootAccountBillConfigBatchUpdateReq[T RootAccountBillConfigExtension] struct {
	Bills []RootAccountBillConfigUpdateReq[T] `json:"bills" validate:"required,max=100"`
}

// RootAccountBillConfigUpdateReq define batch update account bill config update option.
type RootAccountBillConfigUpdateReq[T RootAccountBillConfigExtension] struct {
	ID string `json:"id" validate:"required"`

	Vendor            enumor.Vendor `json:"vendor" validate:"omitempty"`
	AccountID         string        `json:"account_id" validate:"omitempty"`
	CloudDatabaseName string        `json:"cloud_database_name" validate:"omitempty"`
	CloudTableName    string        `json:"cloud_table_name" validate:"omitempty"`
	Status            int64         `json:"status" validate:"omitempty"`
	ErrMsg            []string      `json:"err_msg" validate:"omitempty"`
	Extension         *T            `json:"extension" validate:"omitempty"`
}

// Validate account bill config batch update request.
func (req *RootAccountBillConfigBatchUpdateReq[T]) Validate() error {
	if len(req.Bills) == 0 {
		return errors.New("root account bill config is required")
	}

	if len(req.Bills) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("root account bill config count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// RootAccountBillConfigListReq account bill config list req.
type RootAccountBillConfigListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate account bill config list request.
func (req *RootAccountBillConfigListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RootAccountBillConfigListResp defines list account bill config response.
type RootAccountBillConfigListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *RootAccountBillConfigListResult `json:"data"`
}

// RootAccountBillConfigListResult defines list account bill config result.
type RootAccountBillConfigListResult struct {
	Count   uint64                           `json:"count"`
	Details []bill.BaseRootAccountBillConfig `json:"details"`
}

// RootAccountBillConfigExtListResult define account bill config with extension list result.
type RootAccountBillConfigExtListResult[T bill.RootAccountBillConfigExtension] struct {
	Count   uint64                          `json:"count,omitempty"`
	Details []bill.RootAccountBillConfig[T] `json:"details,omitempty"`
}

// RootAccountBillConfigExtListResp define account bill config with extension list response.
type RootAccountBillConfigExtListResp[T bill.RootAccountBillConfigExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *RootAccountBillConfigExtListResult[T] `json:"data"`
}
