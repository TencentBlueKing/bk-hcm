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

package routetable

import (
	"hcm/pkg/api/core"
	routetable "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// HuaWeiRouteBatchCreateReq defines batch create huawei route request.
type HuaWeiRouteBatchCreateReq struct {
	HuaWeiRoutes []HuaWeiRouteCreateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate HuaWeiRouteBatchCreateReq.
func (c *HuaWeiRouteBatchCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// HuaWeiRouteCreateReq defines create huawei route request.
type HuaWeiRouteCreateReq struct {
	CloudRouteTableID string  `json:"cloud_route_table_id" validate:"required"`
	Type              string  `json:"type" validate:"required"`
	Destination       string  `json:"destination" validate:"required"`
	NextHop           string  `json:"nexthop" validate:"required"`
	Memo              *string `json:"memo,omitempty" validate:"omitempty"`
}

// -------------------------- Update --------------------------

// HuaWeiRouteBatchUpdateReq defines batch update huawei route request.
type HuaWeiRouteBatchUpdateReq struct {
	HuaWeiRoutes []HuaWeiRouteUpdateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate HuaWeiRouteBatchUpdateReq.
func (u *HuaWeiRouteBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// HuaWeiRouteUpdateReq defines update huawei route request.
type HuaWeiRouteUpdateReq struct {
	ID                    string `json:"id" validate:"required"`
	HuaWeiRouteUpdateInfo `json:",inline" validate:"omitempty"`
}

// HuaWeiRouteUpdateInfo defines update huawei route request base info.
type HuaWeiRouteUpdateInfo struct {
	Type        string  `json:"type" validate:"omitempty"`
	Destination string  `json:"destination" validate:"omitempty"`
	NextHop     string  `json:"nexthop" validate:"omitempty"`
	Memo        *string `json:"memo,omitempty" validate:"omitempty"`
}

// -------------------------- List --------------------------

// HuaWeiRouteListReq defines list huawei route request.
type HuaWeiRouteListReq struct {
	*core.ListReq `json:",inline"`
	RouteTableID  string `json:"route_table_id"`
}

// Validate ...
func (r HuaWeiRouteListReq) Validate() error {
	if r.ListReq == nil {
		return errf.New(errf.InvalidParameter, "list request is required")
	}

	if err := r.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// HuaWeiRouteListResp defines list huawei route response.
type HuaWeiRouteListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *HuaWeiRouteListResult `json:"data"`
}

// HuaWeiRouteListResult defines list huawei route result.
type HuaWeiRouteListResult struct {
	Count   uint64                   `json:"count"`
	Details []routetable.HuaWeiRoute `json:"details"`
}
