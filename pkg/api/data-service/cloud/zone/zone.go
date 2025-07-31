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

// Package zone ...
package zone

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Update --------------------------

// ZoneBatchUpdateReq zone batch update request.
type ZoneBatchUpdateReq[Extension zone.ZoneExtension] struct {
	Zones []ZoneBatchUpdate[Extension] `json:"zones" validate:"required"`
}

// ZoneBatchUpdate define zone batch update.
type ZoneBatchUpdate[Extension zone.ZoneExtension] struct {
	ID        string     `json:"id" validate:"required"`
	State     string     `json:"state" validate:"omitempty"`
	Extension *Extension `json:"extension" validate:"omitempty"`
}

// Validate security group update request.
func (req *ZoneBatchUpdateReq[T]) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Zones) == 0 {
		return errors.New("security group is required")
	}

	if len(req.Zones) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Create --------------------------

// ZoneBatchCreateReq zone create request.
type ZoneBatchCreateReq[Extension zone.ZoneExtension] struct {
	Zones []ZoneBatchCreate[Extension] `json:"zones" validate:"required"`
}

// ZoneBatchCreate define zone batch create.
type ZoneBatchCreate[Extension zone.ZoneExtension] struct {
	CloudID   string     `json:"cloud_id" validate:"required"`
	Name      string     `json:"name" validate:"required"`
	State     string     `json:"state" validate:"required"`
	Region    string     `json:"region" validate:"required"`
	NameCn    string     `json:"name_cn" validate:"omitempty"`
	Extension *Extension `json:"extension" validate:"required"`
}

// Validate zone create request.
func (req *ZoneBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Delete --------------------------

// ZoneBatchDeleteReq zone delete request.
type ZoneBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate zone delete request.
func (req *ZoneBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ZoneListReq zone list req.
type ZoneListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate zone list request.
func (req *ZoneListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ZoneListResult define zone list result.
type ZoneListResult struct {
	Count   uint64          `json:"count,omitempty"`
	Details []zone.BaseZone `json:"details,omitempty"`
}

// ZoneListResp define zone list resp.
type ZoneListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ZoneListResult `json:"data"`
}
