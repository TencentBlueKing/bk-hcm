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

// HuaWeiRegionBatchUpdateReq define huawei region batch update request.
type HuaWeiRegionBatchUpdateReq struct {
	Regions []HuaWeiRegionBatchUpdate `json:"rules" validate:"required"`
}

// HuaWeiRegionBatchUpdate azure resource group batch update option.
type HuaWeiRegionBatchUpdate struct {
	ID          string `json:"id" validate:"required"`
	Type        string `json:"type"`
	LocalesZhCn string `json:"locals_zh_cn"`
}

// Validate azure resource group batch update request.
func (req *HuaWeiRegionBatchUpdateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("security group rule is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Create --------------------------

// HuaWeiRegionBatchCreateReq define huawei region create request.
type HuaWeiRegionBatchCreateReq struct {
	Regions []HuaWeiRegionBatchCreate `json:"regions" validate:"required"`
}

// HuaWeiRegionBatchCreate define huawei region when create.
type HuaWeiRegionBatchCreate struct {
	Service     string `json:service`
	RegionID    string `json:"region_id"`
	Type        string `json:"type"`
	LocalesPtBr string `json:"locales_pt_br"`
	LocalesZhCn string `json:"locales_zh_cn"`
	LocalesEnUs string `json:"locales_en_us"`
	LocalesEsUs string `json:"locales_es_us"`
	LocalesEsEs string `json:"locales_es_es"`
}

// Validate huawei region create request.
func (req *HuaWeiRegionBatchCreateReq) Validate() error {
	if len(req.Regions) == 0 {
		return errors.New("region is required")
	}

	if len(req.Regions) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("region rule count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Delete --------------------------

// HuaWeiRegionBatchDeleteReq huawei region delete request.
type HuaWeiRegionBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate huawei region delete request.
func (req *HuaWeiRegionBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// HuaWeiRegionListReq ...
type HuaWeiRegionListReq struct {
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (l *HuaWeiRegionListReq) Validate() error {
	return validator.Validate.Struct(l)
}

// HuaWeiRegionListResult define huawei region list result.
type HuaWeiRegionListResult struct {
	Count   uint64                `json:"count,omitempty"`
	Details []region.HuaWeiRegion `json:"details,omitempty"`
}

// HuaWeiRegionListResp define huawei region list resp.
type HuaWeiRegionListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *HuaWeiRegionListResult `json:"data"`
}
