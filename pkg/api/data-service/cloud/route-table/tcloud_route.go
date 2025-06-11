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

// TCloudRouteBatchCreateReq defines batch create tcloud route request.
type TCloudRouteBatchCreateReq struct {
	TCloudRoutes []TCloudRouteCreateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate TCloudRouteBatchCreateReq.
func (c *TCloudRouteBatchCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudRouteCreateReq defines create tcloud route request.
type TCloudRouteCreateReq struct {
	CloudID                  string  `json:"cloud_id" validate:"required"`
	CloudRouteTableID        string  `json:"cloud_route_table_id" validate:"required"`
	DestinationCidrBlock     string  `json:"destination_cidr_block" validate:"required"`
	DestinationIpv6CidrBlock *string `json:"destination_ipv6_cidr_block,omitempty" validate:"omitempty"`
	GatewayType              string  `json:"gateway_type" validate:"required"`
	CloudGatewayID           string  `json:"cloud_gateway_id" validate:"required"`
	Enabled                  bool    `json:"enabled" validate:"-"`
	RouteType                string  `json:"route_type" validate:"required"`
	PublishedToVbc           bool    `json:"published_to_vbc" validate:"-"`
	Memo                     *string `json:"memo,omitempty" validate:"omitempty"`
}

// -------------------------- Update --------------------------

// TCloudRouteBatchUpdateReq defines batch update tcloud route request.
type TCloudRouteBatchUpdateReq struct {
	TCloudRoutes []TCloudRouteUpdateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate TCloudRouteBatchUpdateReq.
func (u *TCloudRouteBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// TCloudRouteUpdateReq defines update tcloud route request.
type TCloudRouteUpdateReq struct {
	ID                    string `json:"id" validate:"required"`
	TCloudRouteUpdateInfo `json:",inline" validate:"omitempty"`
}

// TCloudRouteUpdateInfo defines update tcloud route request base info.
type TCloudRouteUpdateInfo struct {
	DestinationCidrBlock     string  `json:"destination_cidr_block"`
	DestinationIpv6CidrBlock *string `json:"destination_ipv6_cidr_block,omitempty"`
	GatewayType              string  `json:"gateway_type"`
	CloudGatewayID           string  `json:"cloud_gateway_id"`
	Enabled                  *bool   `json:"enabled"`
	RouteType                string  `json:"route_type"`
	PublishedToVbc           *bool   `json:"published_to_vbc"`
	Memo                     *string `json:"memo,omitempty"`
}

// -------------------------- List --------------------------

// TCloudRouteListReq defines list tcloud route request.
type TCloudRouteListReq struct {
	*core.ListReq `json:",inline"`
	RouteTableID  string `json:"route_table_id"`
}

// Validate ...
func (r TCloudRouteListReq) Validate() error {
	if r.ListReq == nil {
		return errf.New(errf.InvalidParameter, "list request is required")
	}

	if err := r.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// TCloudRouteListResp defines list tcloud route response.
type TCloudRouteListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudRouteListResult `json:"data"`
}

// TCloudRouteListResult defines list tcloud route result.
type TCloudRouteListResult struct {
	Count   uint64                   `json:"count"`
	Details []routetable.TCloudRoute `json:"details"`
}
