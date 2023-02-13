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

package cloudserver

import (
	"hcm/pkg/api/core"
	routetable "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Update --------------------------

// RouteTableUpdateReq defines update route table request.
type RouteTableUpdateReq struct {
	Memo *string `json:"memo" validate:"required"`
}

// Validate RouteTableUpdateReq.
func (u RouteTableUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- List --------------------------

// RouteTableListResult defines list route table result.
type RouteTableListResult struct {
	Count   uint64                      `json:"count"`
	Details []routetable.BaseRouteTable `json:"details"`
}

// -------------------------- Relation ------------------------

// AssignRouteTableToBizReq assign route tables to biz request.
type AssignRouteTableToBizReq struct {
	RouteTableIDs []string `json:"route_table_ids"`
	BkBizID       int64    `json:"bk_biz_id"`
}

// Validate AssignRouteTableToBizReq.
func (a AssignRouteTableToBizReq) Validate() error {
	if len(a.RouteTableIDs) == 0 {
		return errf.New(errf.InvalidParameter, "route table ids are required")
	}

	if a.BkBizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is required")
	}

	return nil
}

// -------------------------- Subnet ------------------------

// CountRouteTableSubnetsReq count subnets in route tables request.
type CountRouteTableSubnetsReq struct {
	IDs []string `json:"ids"`
}

// Validate CountRouteTableSubnetsReq.
func (a CountRouteTableSubnetsReq) Validate() error {
	if len(a.IDs) == 0 {
		return errf.New(errf.InvalidParameter, "route table ids are required")
	}

	if uint(len(a.IDs)) > core.DefaultMaxPageLimit {
		return errf.Newf(errf.InvalidParameter, "route table ids exceeds maximum limit: %d", core.DefaultMaxPageLimit)
	}

	return nil
}
