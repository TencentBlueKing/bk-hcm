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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// GcpListOption defines basic gcp list options.
type GcpListOption struct {
	Page      *core.GcpPage `json:"page" validate:"required"`
	CloudIDs  []string      `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string      `json:"self_links" validate:"omitempty"`
	// network link
	Network []string `json:"networks"`
}

// Validate gcp list option.
func (a GcpListOption) Validate() error {

	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.CloudIDs) > core.GcpQueryLimit {
		return errf.Newf(errf.InvalidParameter, "gcp resource ids length should <= %d", core.GcpQueryLimit)
	}

	if len(a.SelfLinks) > core.GcpQueryLimit {
		return errf.Newf(errf.InvalidParameter, "gcp resource self link length should <= %d", core.GcpQueryLimit)
	}

	if err := a.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// GcpRouteListResult defines gcp list route result.
type GcpRouteListResult struct {
	NextPageToken string     `json:"next_page_token,omitempty"`
	Details       []GcpRoute `json:"details"`
}

// GcpRoute defines gcp route struct.
type GcpRoute struct {
	CloudID  string `json:"cloud_id"`
	SelfLink string `json:"self_link"`
	// Network self link
	Network          string   `json:"network"`
	Name             string   `json:"name"`
	DestRange        string   `json:"dest_range"`
	NextHopGateway   *string  `json:"next_hop_gateway,omitempty"`
	NextHopIlb       *string  `json:"next_hop_ilb,omitempty"`
	NextHopInstance  *string  `json:"next_hop_instance,omitempty"`
	NextHopIp        *string  `json:"next_hop_ip,omitempty"`
	NextHopNetwork   *string  `json:"next_hop_network,omitempty"`
	NextHopPeering   *string  `json:"next_hop_peering,omitempty"`
	NextHopVpnTunnel *string  `json:"next_hop_vpn_tunnel,omitempty"`
	Priority         int64    `json:"priority"`
	RouteStatus      string   `json:"route_status"`
	RouteType        string   `json:"route_type"`
	Tags             []string `json:"tags,omitempty"`
	Memo             *string  `json:"memo,omitempty"`
}

// GetCloudID ...
func (route GcpRoute) GetCloudID() string {
	return route.CloudID
}
