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
	"hcm/pkg/rest"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// TCloudRegionCreateReq define tcloud region create request.
type TCloudRegionCreateReq struct {
	Regions []TCloudRegionBatchCreate `json:"rules" validate:"required"`
}

// Validate tcloud region create request.
func (req *TCloudRegionCreateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// TCloudRegionBatchCreate define tcloud region rule when create.
type TCloudRegionBatchCreate struct {
	Vendor      enumor.Vendor `json:"vendor" validate:"required"`
	RegionID    string        `json:"region_id" validate:"required"`
	RegionName  string        `json:"region_name" validate:"required"`
	IsAvailable int64         `json:"is_available"`
	Creator     string        `json:"creator"`
}

// -------------------------- Update --------------------------

// TCloudRegionBatchUpdateReq define tcloud region batch update request.
type TCloudRegionBatchUpdateReq struct {
	Regions []TCloudRegionBatchUpdate `json:"rules" validate:"required"`
}

// TCloudRegionBatchUpdate tcloud region batch update option.
type TCloudRegionBatchUpdate struct {
	ID          string        `json:"id" validate:"required"`
	Vendor      enumor.Vendor `json:"vendor" validate:"required"`
	RegionID    string        `json:"region_id" validate:"required"`
	RegionName  string        `json:"region_name" validate:"required"`
	IsAvailable int64         `json:"is_available"`
	Creator     string        `json:"creator"`
	Reviser     string        `json:"reviser"`
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

// -------------------------- Get --------------------------

// TCloudRegionBaseInfoBatchUpdateReq defines batch update region base info request.
type TCloudRegionBaseInfoBatchUpdateReq struct {
	Regions []TcloudRegionBaseInfoUpdateReq `json:"regions" validate:"required"`
}

// TcloudRegionBaseInfoUpdateReq defines update region base info request.
type TcloudRegionBaseInfoUpdateReq struct {
	IDs  []string                 `json:"id" validate:"required"`
	Data *TCloudRegionBatchUpdate `json:"data" validate:"required"`
}

// Validate VpcBaseInfoBatchUpdateReq.
func (u *TCloudRegionBaseInfoBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
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

// -------------------------- Delete --------------------------

// TCloudRegionBatchDeleteReq tcloud region delete request.
type TCloudRegionBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate tcloud region delete request.
func (req *TCloudRegionBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
