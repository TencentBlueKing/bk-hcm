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
	"errors"
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typessubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"google.golang.org/api/compute/v1"
)

// CreateSubnet create subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/insert
func (g *Gcp) CreateSubnet(kt *kit.Kit, opt *typessubnet.GcpSubnetCreateOption) (uint64, error) {
	if err := opt.Validate(); err != nil {
		return 0, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return 0, err
	}

	req := &compute.Subnetwork{
		Description:             converter.PtrToVal(opt.Memo),
		EnableFlowLogs:          opt.Extension.EnableFlowLogs,
		ExternalIpv6Prefix:      "",
		IpCidrRange:             opt.Extension.IPv4Cidr,
		Ipv6AccessType:          "",
		LogConfig:               nil,
		Name:                    opt.Name,
		Network:                 opt.CloudVpcID,
		PrivateIpGoogleAccess:   opt.Extension.PrivateIpGoogleAccess,
		PrivateIpv6GoogleAccess: "",
		Purpose:                 "",
		Region:                  opt.Extension.Region,
		Role:                    "",
		SecondaryIpRanges:       nil,
		StackType:               "",
		ForceSendFields:         nil,
		NullFields:              nil,
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
	resp, err := client.Subnetworks.Insert(cloudProjectID, opt.Extension.Region, req).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("create subnet failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	handler := &createSubnetPollingHandler{
		parseSelfLinkToName(resp.Zone),
	}
	respPoller := poller.Poller[*Gcp, []*compute.Operation, []uint64]{Handler: handler}
	results, err := respPoller.PollUntilDone(g, kt, []*string{to.Ptr(resp.OperationGroupId)},
		types.NewBatchCreateSubnetPollerOption())
	if err != nil {
		return 0, err
	}

	if len(converter.PtrToVal(results)) <= 0 {
		return 0, fmt.Errorf("create subnet failed")
	}

	return (converter.PtrToVal(results))[0], nil
}

// UpdateSubnet update subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/patch
// TODO right now only memo is supported to update, but gcp description can not be updated.
func (g *Gcp) UpdateSubnet(_ *kit.Kit, _ *typessubnet.GcpSubnetUpdateOption) error {
	return nil
}

// DeleteSubnet delete subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/delete
func (g *Gcp) DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
	region := parseSelfLinkToName(opt.Region)
	_, err = client.Subnetworks.Delete(cloudProjectID, region, opt.ResourceID).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("delete subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/list
func (g *Gcp) ListSubnet(kt *kit.Kit, opt *typessubnet.GcpSubnetListOption) (*typessubnet.GcpSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID

	listCall := client.Subnetworks.List(cloudProjectID, opt.Region).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		listCall.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if len(opt.SelfLinks) > 0 {
		listCall.Filter(generateResourceFilter("selfLink", opt.SelfLinks))
	}

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]typessubnet.GcpSubnet, 0, len(resp.Items))
	for _, item := range resp.Items {
		details = append(details, converter.PtrToVal(convertSubnet(item)))
	}

	return &typessubnet.GcpSubnetListResult{NextPageToken: resp.NextPageToken, Details: details}, nil
}

// ListSubnetWithIPNumber 查询子网列表和子网的IP计数.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/list
func (g *Gcp) ListSubnetWithIPNumber(kt *kit.Kit, opt *typessubnet.GcpSubnetListOption) (
	*typessubnet.GcpSubnetListResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	subnetResp, err := g.ListSubnet(kt, opt)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subnetIPMap := make(map[string]uint64)
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	request := client.Addresses.List(g.CloudProjectID(), opt.Region).Context(kt.Ctx).MaxResults(core.GcpQueryLimit)
	for {
		resp, err := request.Do()
		if err != nil {
			logs.Errorf("list address failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, item := range resp.Items {
			subnetIPMap[item.Subnetwork] += 1
		}

		if len(resp.NextPageToken) == 0 {
			break
		}

		request.PageToken(opt.Page.PageToken)
	}

	for index, one := range subnetResp.Details {
		total := uint64(0)
		usedIPCount := subnetIPMap[one.Extension.SelfLink]
		for _, ipv4 := range one.Ipv4Cidr {
			counts, err := cidr.CidrIPCounts(ipv4)
			if err != nil {
				logs.Errorf("get cidr ip count failed, err: %v, cidr: %s, rid: %s", err, ipv4, kt.Rid)
				return nil, err
			}

			total += uint64(counts)
		}

		for _, ipv6 := range one.Ipv6Cidr {
			counts, err := cidr.CidrIPCounts(ipv6)
			if err != nil {
				logs.Errorf("get cidr ip count failed, err: %v, cidr: %s, rid: %s", err, ipv6, kt.Rid)
				return nil, err
			}

			total += uint64(counts)
		}

		subnetResp.Details[index].Extension.TotalIpAddressCount = total
		subnetResp.Details[index].Extension.AvailableIPAddressCount = total - usedIPCount
		subnetResp.Details[index].Extension.UsedIpAddressCount = usedIPCount
	}

	return &typessubnet.GcpSubnetListResult{NextPageToken: subnetResp.NextPageToken, Details: subnetResp.Details}, nil
}

func convertSubnet(data *compute.Subnetwork) *typessubnet.GcpSubnet {
	if data == nil {
		return nil
	}

	// @see https://www.googleapis.com/compute/v1/projects/xxxx/regions/us-centrall
	region := ""
	if len(data.Region) > 0 {
		region = parseSelfLinkToName(data.Region)
	}

	subnet := &typessubnet.GcpSubnet{
		CloudVpcID: data.Network,
		CloudID:    strconv.FormatUint(data.Id, 10),
		Name:       data.Name,
		Memo:       &data.Description,
		Extension: &typessubnet.GcpSubnetExtension{
			SelfLink:              data.SelfLink,
			Region:                region,
			StackType:             data.StackType,
			Ipv6AccessType:        data.Ipv6AccessType,
			GatewayAddress:        data.GatewayAddress,
			PrivateIpGoogleAccess: data.PrivateIpGoogleAccess,
			EnableFlowLogs:        data.EnableFlowLogs,
		},
	}

	if len(data.IpCidrRange) != 0 {
		subnet.Ipv4Cidr = []string{data.IpCidrRange}
	}

	if len(data.Ipv6CidrRange) != 0 {
		subnet.Ipv6Cidr = []string{data.Ipv6CidrRange}
	}

	return subnet
}

type createSubnetPollingHandler struct {
	zone string
}

// Done ...
func (h *createSubnetPollingHandler) Done(items []*compute.Operation) (bool, *[]uint64) {
	results := make([]uint64, 0)

	flag := true
	for _, item := range items {
		if item.OperationType == "insert" && item.Status == "DONE" {
			results = append(results, item.TargetId)
			continue
		}

		if item.OperationType == "insert" && item.Status == "PENDING" {
			flag = false
			continue
		}

		if item.OperationType == "insert" && item.Status == "RUNNING" {
			flag = false
			continue
		}
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createSubnetPollingHandler) Poll(client *Gcp, kt *kit.Kit, operGroupIDs []*string) ([]*compute.Operation, error) {
	if len(operGroupIDs) == 0 {
		return nil, errors.New("operation group id is required")
	}

	computeClient, err := client.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	operResp, err := computeClient.ZoneOperations.List(client.CloudProjectID(), h.zone).Context(kt.Ctx).
		Filter(fmt.Sprintf(`operationGroupId="%s"`, *operGroupIDs[0])).Do()
	if err != nil {
		logs.Errorf("list zone operations failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(operResp.Items) <= 1 {
		return nil, errors.New("operation has not been created yet, need to wait")
	}

	return operResp.Items, nil
}
