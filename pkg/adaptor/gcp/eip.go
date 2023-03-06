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

	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/compute/v1"
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

	if opt.Region == eip.GcpGlobalRegion {
		request := client.GlobalAddresses.List(g.CloudProjectID()).Context(kt.Ctx)

		if len(opt.CloudIDs) > 0 {
			request.Filter(generateResourceIDsFilter(opt.CloudIDs))
		}

		if len(opt.SelfLinks) > 0 {
			request.Filter(generateResourceFilter("selfLink", opt.SelfLinks))
		}

		if opt.Page != nil {
			request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
		}

		resp, err := request.Do()
		if err != nil {
			logs.Errorf("list global address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		return &eip.GcpEipListResult{Details: convert(resp, opt.Region), NextPageToken: resp.NextPageToken}, nil
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

	return &eip.GcpEipListResult{Details: convert(resp, opt.Region), NextPageToken: resp.NextPageToken}, nil
}

// ListAggregatedEip ...
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/addresses/aggregatedList
func (g *Gcp) ListAggregatedEip(kt *kit.Kit, opt *eip.GcpEipAggregatedListOption) ([]*compute.Address, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	// 地域Eip
	resp, err := client.Addresses.AggregatedList(g.CloudProjectID()).Context(kt.Ctx).
		Filter(generateResourceFilter("address", opt.IPAddresses)).Do()
	if err != nil {
		logs.Errorf("list aggregated address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	result := make([]*compute.Address, 0, len(opt.IPAddresses))
	for _, one := range resp.Items {
		for _, address := range one.Addresses {
			result = append(result, address)
		}
	}

	return result, nil
}

// DeleteEip ...
// reference: global address reference: https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses/delete
// reference: regional address reference: https://cloud.google.com/compute/docs/reference/rest/v1/addresses/delete
func (g *Gcp) DeleteEip(kt *kit.Kit, opt *eip.GcpEipDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp eip delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	if opt.Region == eip.GcpGlobalRegion {
		_, err = client.GlobalAddresses.Delete(g.CloudProjectID(), opt.EipName).
			Context(kt.Ctx).
			RequestId(kt.Rid).
			Do()
	} else {
		_, err = client.Addresses.Delete(g.CloudProjectID(), opt.Region, opt.EipName).Context(kt.Ctx).RequestId(kt.Rid).Do()
	}

	if err != nil {
		logs.Errorf("delete address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// AssociateEip ...
// reference:
// https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address?hl=zh-cn#assign_new_instance
func (g *Gcp) AssociateEip(kt *kit.Kit, opt *eip.GcpEipAssociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp eip associate option is required")
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.AddAccessConfig(
		g.CloudProjectID(),
		opt.Zone,
		opt.CvmName,
		opt.NetworkInterfaceName,
		&compute.AccessConfig{Name: eip.DefaultExternalNatName, NatIP: opt.PublicIp},
	).Context(kt.Ctx).RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("associate gcp address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// DisassociateEip ...
// reference: https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address?hl=zh-cn#api_6
func (g *Gcp) DisassociateEip(kt *kit.Kit, opt *eip.GcpEipDisassociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp eip disassociate option is required")
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.DeleteAccessConfig(
		g.CloudProjectID(),
		opt.Zone,
		opt.CvmName,
		eip.DefaultExternalNatName,
		opt.NetworkInterfaceName,
	).Context(kt.Ctx).RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("disassociate gcp address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

func convert(resp *compute.AddressList, region string) []*eip.GcpEip {
	eips := make([]*eip.GcpEip, len(resp.Items))

	for idx, item := range resp.Items {
		eIp := &eip.GcpEip{
			CloudID:      strconv.FormatUint(item.Id, 10),
			Name:         &item.Name,
			Region:       region,
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
