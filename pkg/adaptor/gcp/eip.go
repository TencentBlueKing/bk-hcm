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

	"google.golang.org/api/compute/v1"

	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListEip ...
// reference: global address reference: https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses/list
// reference: regional address reference: https://cloud.google.com/compute/docs/reference/rest/v1/addresses/list
func (g *Gcp) ListEip(kt *kit.Kit, opt *eip.GcpEipListOption) (*eip.GcpEipListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	if opt.Region == "global" {
		request := client.GlobalAddresses.List(g.CloudProjectID()).Context(kt.Ctx)

		if len(opt.CloudIDs) > 0 {
			request.Filter(generateResourceIDsFilter(opt.CloudIDs))
		}

		if opt.Page != nil {
			request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
		}

		resp, err := request.Do()
		if err != nil {
			logs.Errorf("list global address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		return &eip.GcpEipListResult{Details: convert(resp, opt), NextPageToken: resp.NextPageToken}, nil
	}

	// 地域Eip
	request := client.Addresses.List(g.CloudProjectID(), opt.Region).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		request.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if opt.Page != nil {
		request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := request.Do()
	if err != nil {
		logs.Errorf("list address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return &eip.GcpEipListResult{Details: convert(resp, opt), NextPageToken: resp.NextPageToken}, nil
}

func convert(resp *compute.AddressList, opt *eip.GcpEipListOption) []*eip.GcpEip {
	eips := make([]*eip.GcpEip, len(resp.Items))

	for idx, item := range resp.Items {
		eIp := &eip.GcpEip{
			CloudID:      strconv.FormatUint(item.Id, 10),
			Name:         &item.Name,
			Region:       opt.Region,
			Status:       &item.Status,
			AddressType:  item.AddressType,
			Description:  item.Description,
			IpVersion:    item.IpVersion,
			NetworkTier:  item.NetworkTier,
			PrefixLength: item.PrefixLength,
			Purpose:      item.Purpose,
			Network:      item.Network,
			Subnetwork:   item.Subnetwork,
			SelfLink:     item.SelfLink,
		}
		switch item.AddressType {
		case "EXTERNAL":
			eIp.PublicIp = &item.Address
		case "INTERNAL":
			eIp.PrivateIp = &item.Address
		}

		eips[idx] = eIp
	}
	return eips
}
