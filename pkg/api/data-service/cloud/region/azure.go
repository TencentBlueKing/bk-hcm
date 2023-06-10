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

// Package region 包提供各类云资源的请求与返回序列化器
package region

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Update --------------------------

// AzureRegionBatchUpdateReq define azure region batch update request.
type AzureRegionBatchUpdateReq struct {
	Regions []AzureRegionBatchUpdate `json:"rules" validate:"required"`
}

// AzureRegionBatchUpdate azure region batch update option.
type AzureRegionBatchUpdate struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type"`
}

// Validate azure resource group batch update request.
func (req *AzureRegionBatchUpdateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("region is required")
	}

	return nil
}

// -------------------------- Create --------------------------

// AzureRegionBatchCreateReq define azure region create request.
type AzureRegionBatchCreateReq struct {
	Regions []AzureRegionBatchCreate `json:"regions" validate:"required"`
}

// AzureRegionBatchCreate define azure region when create.
type AzureRegionBatchCreate struct {
	CloudID           string `json:"cloud_id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	DisplayName       string `json:"display_name"`
	RegionDisplayName string `json:"region_display_name"`
	GeographyGroup    string `json:"geography_group"`
	Latitude          string `json:"latitude"`
	Longitude         string `json:"longitude"`
	PhysicalLocation  string `json:"physical_location"`
	RegionType        string `json:"region_type"`
	PairedRegionName  string `json:"paired_region_name"`
	PairedRegionId    string `json:"paired_region_id"`
}

// Validate azure region create request.
func (req *AzureRegionBatchCreateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("region is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("region count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Delete --------------------------

// AzureRegionBatchDeleteReq azure region delete request.
type AzureRegionBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate azure region delete request.
func (req *AzureRegionBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AzureRegionListReq ...
type AzureRegionListReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (l *AzureRegionListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// AzureRegionListResult define azure region list result.
type AzureRegionListResult struct {
	Count   uint64               `json:"count,omitempty"`
	Details []region.AzureRegion `json:"details,omitempty"`
}

// AzureRegionListResp define azure region list resp.
type AzureRegionListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AzureRegionListResult `json:"data"`
}
