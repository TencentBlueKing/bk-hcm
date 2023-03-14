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
	"strings"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// CreateSubnet create subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/insert
func (g *Gcp) CreateSubnet(kt *kit.Kit, opt *types.GcpSubnetCreateOption) (uint64, error) {
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

	return resp.TargetId, nil
}

// UpdateSubnet update subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/patch
// TODO right now only memo is supported to update, but gcp description can not be updated.
func (g *Gcp) UpdateSubnet(_ *kit.Kit, _ *types.GcpSubnetUpdateOption) error {
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
func (g *Gcp) ListSubnet(kt *kit.Kit, opt *types.GcpSubnetListOption) (*types.GcpSubnetListResult, error) {
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

	details := make([]types.GcpSubnet, 0, len(resp.Items))
	for _, item := range resp.Items {
		details = append(details, converter.PtrToVal(convertSubnet(item)))
	}

	return &types.GcpSubnetListResult{NextPageToken: resp.NextPageToken, Details: details}, nil
}

func convertSubnet(data *compute.Subnetwork) *types.GcpSubnet {
	if data == nil {
		return nil
	}

	// @see https://www.googleapis.com/compute/v1/projects/xxxx/regions/us-centrall
	region := ""
	if len(data.Region) > 0 {
		regionArr := strings.Split(data.Region, "/")
		if len(regionArr) >= 9 && regionArr[7] == "regions" {
			region = regionArr[8]
		}
	}

	subnet := &types.GcpSubnet{
		CloudVpcID: data.Network,
		CloudID:    strconv.FormatUint(data.Id, 10),
		Name:       data.Name,
		Memo:       &data.Description,
		Extension: &types.GcpSubnetExtension{
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
