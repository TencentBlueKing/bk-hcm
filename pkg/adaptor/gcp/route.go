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

package gcp

import (
	"strconv"

	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// ListRoute list route.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/routes/list
func (g *Gcp) ListRoute(kt *kit.Kit, opt *core.GcpListOption) (*routetable.GcpRouteListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	listCall := client.Routes.List(g.clientSet.credential.CloudProjectID).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		listCall.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list route failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]routetable.GcpRoute, 0, len(resp.Items))
	for _, item := range resp.Items {
		details = append(details, converter.PtrToVal(convertRoute(item)))
	}

	return &routetable.GcpRouteListResult{NextPageToken: resp.NextPageToken, Details: details}, nil
}

func convertRoute(data *compute.Route) *routetable.GcpRoute {
	if data == nil {
		return nil
	}

	route := &routetable.GcpRoute{
		CloudID:          strconv.FormatUint(data.Id, 10),
		SelfLink:         data.SelfLink,
		Network:          data.Network,
		Name:             data.Name,
		DestRange:        data.DestRange,
		NextHopGateway:   &data.NextHopGateway,
		NextHopIlb:       &data.NextHopIlb,
		NextHopInstance:  &data.NextHopInstance,
		NextHopIp:        &data.NextHopIp,
		NextHopNetwork:   &data.NextHopNetwork,
		NextHopPeering:   &data.NextHopPeering,
		NextHopVpnTunnel: &data.NextHopVpnTunnel,
		Priority:         data.Priority,
		RouteStatus:      data.RouteStatus,
		RouteType:        data.RouteType,
		Tags:             data.Tags,
		Memo:             &data.Description,
	}

	return route
}
