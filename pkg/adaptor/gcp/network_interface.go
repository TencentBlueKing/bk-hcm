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
	"fmt"
	"strconv"
	"strings"

	"hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// ListNetworkInterface list network interface.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
// Note：该接口分页是针对于主机，而不是网络接口，有可能出现查询出来的网络接口数量超过分页数量的情况，使用注意！！！
func (g *Gcp) ListNetworkInterface(kt *kit.Kit, opt *core.GcpListOption) (*typesniproto.GcpInterfaceListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.Zone) == 0 {
		return nil, errf.New(errf.InvalidParameter, "zone is not empty")
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	listCall := client.Instances.List(g.CloudProjectID(), opt.Zone).Context(kt.Ctx)

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	details := make([]typesniproto.GcpNI, 0, len(resp.Items))
	if err := listCall.Pages(kt.Ctx, func(page *compute.InstanceList) error {
		for _, item := range page.Items {
			for _, niItem := range item.NetworkInterfaces {
				details = append(details, converter.PtrToVal(g.ConvertNetworkInterface(item, niItem)))
			}
		}
		return nil
	}); err != nil {
		logs.Errorf("cloudapi failed to list network interface, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &typesniproto.GcpInterfaceListResult{Details: details}, nil
}

// ListNetworkInterfacePage list network interface page.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListNetworkInterfacePage(kt *kit.Kit, opt *core.GcpListOption) (*compute.InstancesListCall, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.Zone) == 0 {
		return nil, errf.New(errf.InvalidParameter, "zone is not empty")
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	listCall := client.Instances.List(g.CloudProjectID(), opt.Zone).Context(kt.Ctx)
	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}
	return listCall, nil
}

// ListNetworkInterfaceByCvmID list network interface by cvm id.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListNetworkInterfaceByCvmID(kt *kit.Kit, opt *typesniproto.GcpListByCvmIDOption) (
	map[string] /*CloudCvmID*/ []typesniproto.GcpNI, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	resp, err := client.Instances.List(g.CloudProjectID(), opt.Zone).Context(kt.Ctx).
		Filter(generateResourceIDsFilter(opt.CloudCvmIDs)).Do()
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make(map[string][]typesniproto.GcpNI, len(resp.Items))
	for _, item := range resp.Items {
		cvmID := strconv.FormatUint(item.Id, 10)
		result[cvmID] = make([]typesniproto.GcpNI, 0, len(item.NetworkInterfaces))

		for _, ni := range item.NetworkInterfaces {
			result[cvmID] = append(result[cvmID], converter.PtrToVal(g.ConvertNetworkInterface(item, ni)))
		}
	}

	return result, nil
}

// ConvertNetworkInterface ...
func (g *Gcp) ConvertNetworkInterface(data *compute.Instance, niItem *compute.NetworkInterface) *typesniproto.GcpNI {
	// @see https://www.googleapis.com/compute/v1/projects/xxxx/zones/us-central1-a
	zone := data.Zone[(strings.LastIndex(data.Zone, "/") + 1):]
	region := zone[:strings.LastIndex(zone, "-")]
	v := &typesniproto.GcpNI{
		Name:       converter.ValToPtr(niItem.Name),
		Region:     converter.ValToPtr(region),
		Zone:       converter.ValToPtr(zone),
		CloudID:    converter.ValToPtr(fmt.Sprintf("%d_%s", data.Id, niItem.Name)),
		InstanceID: converter.ValToPtr(strconv.FormatUint(data.Id, 10)),
	}
	if len(niItem.NetworkIP) > 0 {
		v.PrivateIPv4 = append(v.PrivateIPv4, niItem.NetworkIP)
	}
	if len(niItem.Ipv6Address) > 0 {
		v.PrivateIPv6 = append(v.PrivateIPv6, niItem.Ipv6Address)
	}

	if niItem == nil {
		return v
	}

	v.Extension = &coreni.GcpNIExtension{
		CanIpForward:   data.CanIpForward,
		Status:         data.Status,
		StackType:      niItem.StackType,
		AccessConfigs:  []*coreni.AccessConfig{},
		VpcSelfLink:    niItem.Network,
		SubnetSelfLink: niItem.Subnetwork,
	}

	if len(niItem.AccessConfigs) != 0 {
		for _, tmpAc := range niItem.AccessConfigs {
			v.Extension.AccessConfigs = append(v.Extension.AccessConfigs, &coreni.AccessConfig{
				Type:        tmpAc.Type,
				Name:        tmpAc.Name,
				NatIP:       tmpAc.NatIP,
				NetworkTier: tmpAc.NetworkTier,
			})
			if len(tmpAc.NatIP) > 0 {
				v.PublicIPv4 = append(v.PublicIPv4, tmpAc.NatIP)
			}
			if len(tmpAc.ExternalIpv6) > 0 {
				v.PublicIPv6 = append(v.PublicIPv6, tmpAc.ExternalIpv6)
			}
		}
	}

	if len(niItem.Ipv6AccessConfigs) != 0 {
		for _, tmpAc := range niItem.Ipv6AccessConfigs {
			v.Extension.Ipv6AccessConfigs = append(v.Extension.Ipv6AccessConfigs, &coreni.AccessConfig{
				Type:        tmpAc.Type,
				Name:        tmpAc.Name,
				NatIP:       tmpAc.NatIP,
				NetworkTier: tmpAc.NetworkTier,
			})

			if len(tmpAc.NatIP) > 0 {
				v.PublicIPv4 = append(v.PrivateIPv4, tmpAc.NatIP)
			}
			if len(tmpAc.ExternalIpv6) > 0 {
				v.PublicIPv6 = append(v.PublicIPv6, tmpAc.ExternalIpv6)
			}
		}
	}

	return v
}
