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

// -------------------------- Create --------------------------

// AwsRegionCreateReq define aws region create request.
type AwsRegionCreateReq struct {
	Regions []AwsRegionBatchCreate `json:"regions" validate:"required"`
}

// AwsRegionBatchCreate define aws region rule when create.
type AwsRegionBatchCreate struct {
	Vendor     enumor.Vendor `json:"vendor" validate:"required"`
	RegionID   string        `json:"region_id" validate:"required"`
	RegionName string        `json:"region_name" validate:"required"`
	Status     string        `json:"status"`
	Endpoint   string        `json:"endpoint"`
}

// Validate aws region create request.
func (req *AwsRegionCreateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Update --------------------------

// AwsRegionBatchUpdateReq define aws region batch update request.
type AwsRegionBatchUpdateReq struct {
	Regions []AwsRegionBatchUpdate `json:"regions" validate:"required"`
}

// AwsRegionBatchUpdate aws region batch update option.
type AwsRegionBatchUpdate struct {
	ID         string        `json:"id" validate:"required"`
	Vendor     enumor.Vendor `json:"vendor" validate:"required"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Endpoint   string        `json:"endpoint"`
}

// Validate aws region batch update request.
func (req *AwsRegionBatchUpdateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("regions is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("regions count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- List --------------------------

// AwsRegionListReq aws region list req.
type AwsRegionListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate aws region list request.
func (req *AwsRegionListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsRegionListResult define aws region list result.
type AwsRegionListResult struct {
	Count   uint64            `json:"count"`
	Details []cloud.AwsRegion `json:"details"`
}

// AwsRegionListResp define aws region list resp.
type AwsRegionListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AwsRegionListResult `json:"data"`
}
