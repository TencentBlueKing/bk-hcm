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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hash"
)

// AwsRoute defines aws route struct.
type AwsRoute struct {
	CloudRouteTableID                string  `json:"cloud_route_table_id"`
	DestinationCidrBlock             *string `json:"destination_cidr_block,omitempty"`
	DestinationIpv6CidrBlock         *string `json:"destination_ipv6_cidr_block,omitempty"`
	CloudCarrierGatewayID            *string `json:"cloud_carrier_gateway_id,omitempty"`
	CoreNetworkArn                   *string `json:"core_network_arn,omitempty"`
	CloudDestinationPrefixListID     *string `json:"cloud_destination_prefix_list_id,omitempty"`
	CloudEgressOnlyInternetGatewayID *string `json:"cloud_egress_only_internet_gateway_id,omitempty"`
	CloudGatewayID                   *string `json:"cloud_gateway_id,omitempty"`
	CloudInstanceID                  *string `json:"cloud_instance_id,omitempty"`
	CloudInstanceOwnerID             *string `json:"cloud_instance_owner_id,omitempty"`
	CloudLocalGatewayID              *string `json:"cloud_local_gateway_id,omitempty"`
	CloudNatGatewayID                *string `json:"cloud_nat_gateway_id,omitempty"`
	CloudNetworkInterfaceID          *string `json:"cloud_network_interface_id,omitempty"`
	CloudTransitGatewayID            *string `json:"cloud_transit_gateway_id,omitempty"`
	CloudVpcPeeringConnectionID      *string `json:"cloud_vpc_peering_connection_id,omitempty"`
	State                            string  `json:"state"`
	Propagated                       bool    `json:"propagated"`
}

// GetCloudID ...
func (route AwsRoute) GetCloudID() string {
	return hash.HashString(route.CloudRouteTableID + converter.PtrToVal(route.DestinationCidrBlock) +
	converter.PtrToVal(route.DestinationIpv6CidrBlock) + converter.PtrToVal(route.CloudCarrierGatewayID))
}
