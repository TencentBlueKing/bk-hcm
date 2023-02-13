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

package tcloud

import (
	"strconv"

	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

func convertRoute(data *vpc.Route, cloudRouteTableID string) *routetable.TCloudRoute {
	if data == nil {
		return nil
	}

	r := &routetable.TCloudRoute{
		CloudID:                  strconv.FormatUint(converter.PtrToVal(data.RouteId), 10),
		CloudRouteTableID:        cloudRouteTableID,
		DestinationCidrBlock:     converter.PtrToVal(data.DestinationCidrBlock),
		DestinationIpv6CidrBlock: data.DestinationIpv6CidrBlock,
		GatewayType:              converter.PtrToVal(data.GatewayType),
		CloudGatewayID:           converter.PtrToVal(data.GatewayId),
		Enabled:                  converter.PtrToVal(data.Enabled),
		RouteType:                converter.PtrToVal(data.RouteType),
		PublishedToVbc:           converter.PtrToVal(data.PublishedToVbc),
		Memo:                     data.RouteDescription,
	}

	return r
}
