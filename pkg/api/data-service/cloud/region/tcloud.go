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
	"hcm/pkg/api/core/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// TCloudRegionCreateReq define region create request.
type TCloudRegionCreateReq struct {
	Regions   []TCloudRegionBatchCreate `json:"regions" validate:"required"`
	AccountID string                    `json:"account_id" validate:"required"`
}

// TCloudRegionBatchCreate define region rule when create.
type TCloudRegionBatchCreate struct {
	Vendor     enumor.Vendor       `json:"vendor" validate:"required"`
	RegionID   string              `json:"region_id" validate:"required"`
	RegionName string              `json:"region_name" validate:"required"`
	AreaName   string              `json:"area_name"`
	CityName   string              `json:"city_name"`
	Status     string              `json:"status"`
	Source     enumor.RegionSource `json:"source"`
}

// Validate validate TCloudRegionCreateReq.
func (req *TCloudRegionCreateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, r := range req.Regions {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validate TCloudRegionBatchCreate.
func (req *TCloudRegionBatchCreate) Validate() error {
	if err := req.Vendor.Validate(); err != nil {
		return err
	}
	if err := req.Source.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// TCloudRegionBatchUpdateReq define tcloud region batch update request.
type TCloudRegionBatchUpdateReq struct {
	Regions []TCloudRegionBatchUpdate `json:"regions" validate:"required"`
}

// TCloudRegionBatchUpdate tcloud region batch update option.
type TCloudRegionBatchUpdate struct {
	ID         string              `json:"id" validate:"required"`
	RegionID   string              `json:"region_id"`
	RegionName string              `json:"region_name"`
	AreaName   string              `json:"area_name"`
	CityName   string              `json:"city_name"`
	Status     string              `json:"status"`
	Source     enumor.RegionSource `json:"source"`
}

// Validate validate TCloudRegionBatchUpdateReq.
func (req *TCloudRegionBatchUpdateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, r := range req.Regions {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validate TCloudRegionBatchUpdate.
func (req *TCloudRegionBatchUpdate) Validate() error {
	if len(req.Source) > 0 {
		if err := req.Source.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(req)
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
	Count   uint64                `json:"count,omitempty"`
	Details []region.TCloudRegion `json:"details,omitempty"`
}
