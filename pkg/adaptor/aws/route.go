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

package aws

import (
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func convertRoute(data *ec2.Route, cloudRouteTableID string) *routetable.AwsRoute {
	if data == nil {
		return nil
	}

	r := &routetable.AwsRoute{
		CloudRouteTableID:                cloudRouteTableID,
		DestinationCidrBlock:             data.DestinationCidrBlock,
		DestinationIpv6CidrBlock:         data.DestinationIpv6CidrBlock,
		CloudCarrierGatewayID:            data.CarrierGatewayId,
		CoreNetworkArn:                   data.CoreNetworkArn,
		CloudDestinationPrefixListID:     data.DestinationPrefixListId,
		CloudEgressOnlyInternetGatewayID: data.EgressOnlyInternetGatewayId,
		CloudGatewayID:                   data.GatewayId,
		CloudInstanceID:                  data.InstanceId,
		CloudInstanceOwnerID:             data.InstanceOwnerId,
		CloudLocalGatewayID:              data.LocalGatewayId,
		CloudNatGatewayID:                data.NatGatewayId,
		CloudNetworkInterfaceID:          data.NetworkInterfaceId,
		CloudTransitGatewayID:            data.TransitGatewayId,
		CloudVpcPeeringConnectionID:      data.VpcPeeringConnectionId,
		State:                            converter.PtrToVal(data.State),
	}

	if data.Origin != nil && *data.Origin == "EnableVgwRoutePropagation" {
		r.Propagated = true
	}

	return r
}
