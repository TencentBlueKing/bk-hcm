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

// GcpRouteBatchCreateReq defines batch create gcp route request.
type GcpRouteBatchCreateReq struct {
	GcpRoutes []GcpRouteCreateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate GcpRouteBatchCreateReq.
func (c *GcpRouteBatchCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// GcpRouteCreateReq defines create gcp route request.
type GcpRouteCreateReq struct {
	CloudID          string   `json:"cloud_id" validate:"required"`
	SelfLink         string   `json:"self_link" validate:"required"`
	Network          string   `json:"network" validate:"required"`
	Name             string   `json:"name" validate:"required"`
	DestRange        string   `json:"dest_range" validate:"required"`
	NextHopGateway   *string  `json:"next_hop_gateway,omitempty" validate:"omitempty"`
	NextHopIlb       *string  `json:"next_hop_ilb,omitempty" validate:"omitempty"`
	NextHopInstance  *string  `json:"next_hop_instance,omitempty" validate:"omitempty"`
	NextHopIp        *string  `json:"next_hop_ip,omitempty" validate:"omitempty"`
	NextHopNetwork   *string  `json:"next_hop_network,omitempty" validate:"omitempty"`
	NextHopPeering   *string  `json:"next_hop_peering,omitempty" validate:"omitempty"`
	NextHopVpnTunnel *string  `json:"next_hop_vpn_tunnel,omitempty" validate:"omitempty"`
	Priority         int64    `json:"priority" validate:"required"`
	RouteStatus      string   `json:"route_status" validate:"required"`
	RouteType        string   `json:"route_type" validate:"required"`
	Tags             []string `json:"tags,omitempty" validate:"omitempty"`
	Memo             *string  `json:"memo,omitempty" validate:"omitempty"`
}

// -------------------------- List --------------------------

// GcpRouteListReq defines list gcp route request.
type GcpRouteListReq struct {
	*core.ListReq `json:",inline"`
	RouteTableID  string `json:"route_table_id"`
}

func (r GcpRouteListReq) Validate() error {
	if r.ListReq == nil {
		return errf.New(errf.InvalidParameter, "list request is required")
	}

	if err := r.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// GcpRouteListResp defines list gcp route response.
type GcpRouteListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *GcpRouteListResult `json:"data"`
}

// GcpRouteListResult defines list gcp route result.
type GcpRouteListResult struct {
	Count   uint64                `json:"count"`
	Details []routetable.GcpRoute `json:"details"`
}
