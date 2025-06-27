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

// Package bill ...
package bill

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// AccountBillConfigBatchCreateReq defines batch create account bill config request.
type AccountBillConfigBatchCreateReq[T AccountBillConfigExtension] struct {
	Bills []AccountBillConfigReq[T] `json:"bills" validate:"required,max=100"`
}

// AccountBillConfigReq defines create account bill config request.
type AccountBillConfigReq[T AccountBillConfigExtension] struct {
	Vendor            enumor.Vendor `json:"vendor" validate:"required"`
	AccountID         string        `json:"account_id" validate:"required"`
	CloudDatabaseName string        `json:"cloud_database_name" validate:"omitempty"`
	CloudTableName    string        `json:"cloud_table_name" validate:"omitempty"`
	Status            int64         `json:"status" validate:"omitempty"`
	ErrMsg            []string      `json:"err_msg" validate:"omitempty"`
	Extension         *T            `json:"extension" validate:"omitempty"`
}

// AccountBillConfigExtension defines create account bill config extensional info.
type AccountBillConfigExtension interface {
	cloud.AwsBillConfigExtension | cloud.GcpBillConfigExtension
}

// Validate AccountBillConfigBatchCreateReq.
func (c *AccountBillConfigBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// AccountBillConfigBatchUpdateReq define batch update account bill config request.
type AccountBillConfigBatchUpdateReq[T AccountBillConfigExtension] struct {
	Bills []AccountBillConfigUpdateReq[T] `json:"bills" validate:"required,max=100"`
}

// AccountBillConfigUpdateReq define batch update account bill config update option.
type AccountBillConfigUpdateReq[T AccountBillConfigExtension] struct {
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
func (req *AccountBillConfigBatchUpdateReq[T]) Validate() error {
	if len(req.Bills) == 0 {
		return errors.New("account bill config is required")
	}

	if len(req.Bills) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("account bill config count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// AccountBillConfigListReq account bill config list req.
type AccountBillConfigListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate account bill config list request.
func (req *AccountBillConfigListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AccountBillConfigListResp defines list account bill config response.
type AccountBillConfigListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountBillConfigListResult `json:"data"`
}

// AccountBillConfigListResult defines list account bill config result.
type AccountBillConfigListResult struct {
	Count   uint64                        `json:"count"`
	Details []cloud.BaseAccountBillConfig `json:"details"`
}

// AccountBillConfigExtListResult define account bill config with extension list result.
type AccountBillConfigExtListResult[T cloud.AccountBillConfigExtension] struct {
	Count   uint64                       `json:"count,omitempty"`
	Details []cloud.AccountBillConfig[T] `json:"details,omitempty"`
}

// AccountBillConfigExtListResp define account bill config with extension list response.
type AccountBillConfigExtListResp[T cloud.AccountBillConfigExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountBillConfigExtListResult[T] `json:"data"`
}
