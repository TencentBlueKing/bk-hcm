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

package region

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

// -------------------------- Sync --------------------------

// TCloudRegionSyncReq define tcloud region sync request.
type TCloudRegionSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate tcloud region sync request.
func (req *TCloudRegionSyncReq) Validate() error {
	if len(req.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	return nil
}

// -------------------------- Update --------------------------

// TCloudRegionBatchUpdateReq define tcloud region batch update request.
type TCloudRegionBatchUpdateReq struct {
	Regions []TCloudRegionBatchUpdate `json:"regions" validate:"required"`
}

// TCloudRegionBatchUpdate tcloud region batch update option.
type TCloudRegionBatchUpdate struct {
	ID          string        `json:"id" validate:"required"`
	Vendor      enumor.Vendor `json:"vendor" validate:"required"`
	RegionID    string        `json:"region_id" validate:"required"`
	RegionName  string        `json:"region_name" validate:"required"`
	IsAvailable int64         `json:"is_available"`
}

// Validate tcloud region batch update request.
func (req *TCloudRegionBatchUpdateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// TCloudRegionListReq tcloud region list req.
type TCloudRegionListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate tcloud region list request.
func (req *TCloudRegionListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudRegionListResp define tcloud region list resp.
type TCloudRegionListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudRegionListResult `json:"data"`
}

// TCloudRegionListResult define tcloud region list result.
type TCloudRegionListResult struct {
	Count   uint64               `json:"count,omitempty"`
	Details []cloud.TCloudRegion `json:"details,omitempty"`
}
