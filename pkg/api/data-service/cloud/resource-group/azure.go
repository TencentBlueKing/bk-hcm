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

package resourcegroup

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	resourcegroup "hcm/pkg/api/core/cloud/resource-group"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Update --------------------------

// AzureRGBatchUpdateReq define azure resource group batch update request.
type AzureRGBatchUpdateReq struct {
	ResourceGroups []AzureRGBatchUpdate `json:"rules" validate:"required"`
}

// AzureRGBatchUpdate azure resource group batch update option.
type AzureRGBatchUpdate struct {
	ID       string `json:"id" validate:"required"`
	Location string `json:"location"`
}

// Validate azure resource group batch update request.
func (req *AzureRGBatchUpdateReq) Validate() error {
	if len(req.ResourceGroups) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.ResourceGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Create --------------------------

// AzureRGBatchCreateReq define azure resource group create request.
type AzureRGBatchCreateReq struct {
	ResourceGroups []AzureRGBatchCreate `json:"regions" validate:"required"`
}

// AzureRGBatchCreate define azure resource group when create.
type AzureRGBatchCreate struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Location  string `json:"location"`
	AccountID string `json:"account_id"`
}

// Validate azure resource group create request.
func (req *AzureRGBatchCreateReq) Validate() error {
	if len(req.ResourceGroups) == 0 {
		return errors.New("resource group is required")
	}

	if len(req.ResourceGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("resource group count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Delete --------------------------

// AzureRGBatchDeleteReq azure resource group delete request.
type AzureRGBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate azure resource group delete request.
func (req *AzureRGBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AzureRGListReq ...
type AzureRGListReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (l *AzureRGListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// AzureRGListResult define azure resource group list result.
type AzureRGListResult struct {
	Count   uint64                  `json:"count,omitempty"`
	Details []resourcegroup.AzureRG `json:"details,omitempty"`
}

// AzureRGListResp define azure resource group list resp.
type AzureRGListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AzureRGListResult `json:"data"`
}
