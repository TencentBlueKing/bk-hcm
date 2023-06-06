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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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

// AssociateEip associate eip.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/addAccessConfig
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

	handler := &associateEipPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Gcp, []*eip.GcpEip, []string]{Handler: handler}
	_, err = respPoller.PollUntilDone(g, kt, []*string{to.Ptr(opt.CloudID)}, nil)
	if err != nil {
		return err
	}

	return nil
}

// DisassociateEip disassociate eip.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/deleteAccessConfig
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
		opt.AccessConfigName,
		opt.NetworkInterfaceName,
	).Context(kt.Ctx).RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("disassociate gcp address failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	handler := &disassociateEipPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Gcp, []*eip.GcpEip, []string]{Handler: handler}
	_, err = respPoller.PollUntilDone(g, kt, []*string{to.Ptr(opt.CloudID)}, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateEip ...
// reference: regional https://cloud.google.com/compute/docs/reference/rest/v1/addresses/insert
// reference: global https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses/insert
func (g *Gcp) CreateEip(kt *kit.Kit, opt *eip.GcpEipCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "gcp eip create option is required")
	}

	req, err := opt.ToAddress()
	if err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	if opt.Region == eip.GcpGlobalRegion {
		_, err := client.GlobalAddresses.Insert(g.CloudProjectID(), req).Context(kt.Ctx).Do()
		if err != nil {
			return nil, err
		}
	} else {
		_, err = client.Addresses.Insert(g.CloudProjectID(), opt.Region, req).Context(kt.Ctx).Do()
		if err != nil {
			return nil, err
		}
	}

	var cloudID string
	for {
		address, err := g.getEip(kt, opt.Region, opt.EipName)
		if err != nil {
			return nil, err
		}

		if address != nil && address.Name == opt.EipName {
			cloudID = strconv.FormatUint(address.Id, 10)
			break
		}
	}

	respPoller := poller.Poller[*Gcp, []*eip.GcpEip,
		poller.BaseDoneResult]{Handler: &createEipPollingHandler{region: opt.Region}}
	return respPoller.PollUntilDone(g, kt, []*string{&cloudID}, nil)
}

func (g *Gcp) getEip(kt *kit.Kit, region string, eipName string) (*compute.Address, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	if region == eip.GcpGlobalRegion {
		return client.GlobalAddresses.Get(g.CloudProjectID(), eipName).Context(kt.Ctx).Do()
	} else {
		return client.Addresses.Get(g.CloudProjectID(), region, eipName).Context(kt.Ctx).Do()
	}
}

func convert(resp *compute.AddressList, region string) []*eip.GcpEip {
	eips := make([]*eip.GcpEip, 0)

	for _, item := range resp.Items {
		if item.AddressType == "INTERNAL" {
			continue
		}

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
			Users:        item.Users,
		}
		switch item.AddressType {
		case "EXTERNAL":
			eIp.PublicIp = &item.Address
		case "INTERNAL":
			eIp.PrivateIp = &item.Address
		}

		eips = append(eips, eIp)
	}
	return eips
}

type associateEipPollingHandler struct {
	region string
}

// Done ...
func (h *associateEipPollingHandler) Done(eips []*eip.GcpEip) (bool, *[]string) {
	return eipDone(eips, "IN_USE")
}

// Poll ...
func (h *associateEipPollingHandler) Poll(client *Gcp, kt *kit.Kit, cloudIDs []*string) ([]*eip.GcpEip, error) {
	return eipPoll(client, kt, h.region, cloudIDs)
}

type disassociateEipPollingHandler struct {
	region string
}

// Done ...
func (h *disassociateEipPollingHandler) Done(eips []*eip.GcpEip) (bool, *[]string) {
	return eipDone(eips, "RESERVED")
}

// Poll ...
func (h *disassociateEipPollingHandler) Poll(client *Gcp, kt *kit.Kit, cloudIDs []*string) ([]*eip.GcpEip, error) {
	return eipPoll(client, kt, h.region, cloudIDs)
}

func eipDone(eips []*eip.GcpEip, succeed string) (bool, *[]string) {
	results := make([]string, 0)
	flag := true
	for _, eip := range eips {
		if converter.PtrToVal(eip.Status) != succeed {
			flag = false
			continue
		}

		results = append(results, eip.CloudID)
	}

	return flag, converter.ValToPtr(results)
}

func eipPoll(client *Gcp, kt *kit.Kit, region string, cloudIDs []*string) ([]*eip.GcpEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.GcpEipListOption{Region: region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

type createEipPollingHandler struct {
	region string
}

// Done ...
func (h *createEipPollingHandler) Done(pollResult []*eip.GcpEip) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if converter.PtrToVal(r.Status) == "RESERVING" {
			unknownCloudIDs = append(unknownCloudIDs, r.CloudID)
		} else {
			successCloudIDs = append(successCloudIDs, r.CloudID)
		}
	}

	isDone := false
	if len(successCloudIDs) != 0 && len(successCloudIDs) == len(pollResult) {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
}

// Poll ...
func (h *createEipPollingHandler) Poll(client *Gcp, kt *kit.Kit, cloudIDs []*string) ([]*eip.GcpEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.GcpEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

var _ poller.PollingHandler[*Gcp, []*eip.GcpEip, poller.BaseDoneResult] = new(createEipPollingHandler)
