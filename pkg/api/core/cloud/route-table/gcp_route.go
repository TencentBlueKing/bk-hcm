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
)

// GcpRoute defines gcp route info.
type GcpRoute struct {
	ID               string   `json:"id"`
	CloudID          string   `json:"cloud_id"`
	RouteTableID     string   `json:"route_table_id"`
	VpcID            string   `json:"vpc_id"`
	CloudVpcID       string   `json:"cloud_vpc_id"`
	SelfLink         string   `json:"self_link"`
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
	*core.Revision   `json:",inline"`
}

// GetID ...
func (route GcpRoute) GetID() string {
	return route.ID
}

// GetCloudID ...
func (route GcpRoute) GetCloudID() string {
	return route.CloudID
}
